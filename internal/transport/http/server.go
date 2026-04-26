package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/istok/agent-core/internal/application"
	"github.com/istok/agent-core/internal/application/usecases"
	"github.com/istok/agent-core/internal/ports"
)

// Server - HTTP сервер
type Server struct {
	addr             string
	projectGenerator *usecases.ProjectGeneratorService
	orchestrator     *application.Orchestrator
	watcher          *application.Watcher
	server           *http.Server
}

// NewServer создает HTTP сервер с LLM-провайдером (через порт)
func NewServer(addr string, projectGenerator *usecases.ProjectGeneratorService, llm ports.LLMProvider) *Server {
	orch := application.NewOrchestrator(llm)
	return &Server{
		addr:             addr,
		projectGenerator: projectGenerator,
		orchestrator:     orch,
		watcher:          application.NewWatcher(orch, "http://localhost"+addr),
	}
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Регистрация handlers
	generateHandler := NewGenerateHandler(s.projectGenerator)
	statsHandler := NewStatsHandler(s.projectGenerator)
	healthHandler := NewHealthHandler()
	authHandler := NewAuthHandler()

	// ── SSE СТРИМ — регистрируем ПЕРВЫМ (более специфичный путь) ──
	sseHandler := NewGenerateHandlerSSE(s.orchestrator)
	mux.HandleFunc("/api/v1/generate/stream", s.corsMiddleware(sseHandler.HandleStream))
	log.Println("✅ Route registered: /api/v1/generate/stream → SSE HandleStream")

	// API endpoints
	mux.HandleFunc("/api/v1/generate", s.corsMiddleware(generateHandler.Handle))
	mux.HandleFunc("/api/v1/stats", s.corsMiddleware(statsHandler.Handle))
	mux.HandleFunc("/api/v1/health", s.corsMiddleware(healthHandler.Handle))

	// Agents status — каноничный пайплайн для фронта (Zod-контракт)
	agentsStatusHandler := NewAgentsStatusHandler(s.orchestrator)
	mux.HandleFunc("/api/v1/agents/status", s.corsMiddleware(agentsStatusHandler.Handle))

	// Railway deploy integration
	deployHandler := NewDeployHandler()
	mux.HandleFunc("/api/v1/deploy/railway", s.corsMiddleware(deployHandler.HandleRailway))

	// Auth endpoints
	mux.HandleFunc("/api/v1/auth/signup", s.corsMiddleware(authHandler.HandleSignup))
	mux.HandleFunc("/api/v1/auth/login", s.corsMiddleware(authHandler.HandleLogin))
	mux.HandleFunc("/api/v1/auth/me", s.corsMiddleware(authHandler.HandleMe))

	// Diagnostic endpoints
	diagHandler := NewDiagHandler()
	mux.HandleFunc("/api/v1/diag/models", s.corsMiddleware(diagHandler.Handle))
	mux.HandleFunc("/api/v1/diag/env", s.corsMiddleware(diagHandler.HandleEnv))

	// Watcher V1 — error webhook + reports
	watcherHandler := NewWatcherHandler(s.watcher)
	mux.HandleFunc("/api/v1/internal/error-webhook", s.corsMiddleware(watcherHandler.HandleErrorWebhook))
	mux.HandleFunc("/api/v1/internal/watcher/reports", s.corsMiddleware(watcherHandler.HandleReports))

	log.Println("✅ All routes registered: /generate, /generate/stream, /stats, /health, /auth/*, /diag/*, /internal/error-webhook, /internal/watcher/reports")

	// Wire log output into Watcher ring buffer for 5xx log analysis
	log.SetOutput(&application.WatcherLogWriter{Original: log.Writer(), Watcher: s.watcher})

	// Catch-all 404 trap — ОБЯЗАТЕЛЬНО обёрнут в corsMiddleware,
	// иначе браузер блокирует ответ → фронт видит opaque ошибку вместо JSON.
	mux.HandleFunc("/", s.corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			writeJSON(w, http.StatusOK, map[string]string{
				"service": "istok-agent-core",
				"status":  "running",
				"version": "3.0.0",
			})
			return
		}
		log.Printf("⚠️ 404 TRAP: %s %s (Origin: %s, UA: %s)", r.Method, r.URL.Path, r.Header.Get("Origin"), r.Header.Get("User-Agent"))
		writeError(w, http.StatusNotFound, fmt.Sprintf("Route not found: %s %s", r.Method, r.URL.Path))
	}))

	// Middleware chain: Recovery → SecurityHeaders → Logging → Router
	handler := s.recoveryMiddleware(s.securityHeadersMiddleware(s.loggingMiddleware(mux)))

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Minute, // AI generation takes time
		WriteTimeout: 6 * time.Minute, // Must be > Anthropic/Replicate generation timeout (5min)
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("🚀 HTTP сервер запущен на %s\n", s.addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully останавливает сервер
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("⏳ Остановка HTTP сервера...")
	return s.server.Shutdown(ctx)
}

// recoveryMiddleware перехватывает panic и логирует полный стектрейс в Railway
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("🔥 PANIC recovered [%s %s]: %v", r.Method, r.URL.Path, rec)
				writeError(w, http.StatusInternalServerError, fmt.Sprintf("panic: %v", rec))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// securityHeadersMiddleware устанавливает строгие security headers на каждом ответе.
// HSTS, CSP, X-Frame-Options=DENY → блокирует embed в iframe без явного разрешения.
// Список Frame-разрешённых доменов читается из FRAME_ALLOWED_ORIGINS env (опционально).
func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ── Strict Transport Security: 1 год + subdomains + preload ──
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// ── X-Frame-Options: запретить iframe-embed по умолчанию ──
		// Если задан FRAME_ALLOWED_ORIGINS — используем CSP frame-ancestors вместо DENY.
		frameAllowed := os.Getenv("FRAME_ALLOWED_ORIGINS")
		if frameAllowed == "" {
			w.Header().Set("X-Frame-Options", "DENY")
		} else {
			// Современные браузеры используют CSP frame-ancestors (см. ниже),
			// X-Frame-Options оставляем для совместимости со старыми.
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		}

		// ── X-Content-Type-Options: отключаем MIME-sniffing ──
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// ── Referrer-Policy: не утекать URL во внешние домены ──
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// ── Permissions-Policy: отключаем потенциально опасные API ──
		w.Header().Set("Permissions-Policy",
			"geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()")

		// ── Cross-Origin policies ──
		w.Header().Set("X-XSS-Protection", "0") // современные браузеры доверяют CSP, X-XSS-Protection отключаем
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin") // фронт на vercel должен иметь доступ

		// ── Content-Security-Policy ──
		// SSE-эндпоинт пропускаем, т.к. строгая CSP может ломать стриминг proxy'ами.
		if !strings.HasPrefix(r.URL.Path, "/api/v1/generate/stream") {
			frameAncestors := "'none'"
			if frameAllowed != "" {
				// frame-ancestors допускает space-separated origin list
				frameAncestors = strings.ReplaceAll(frameAllowed, ",", " ")
			}
			csp := strings.Join([]string{
				"default-src 'self'",
				"script-src 'self'",
				"style-src 'self' 'unsafe-inline'", // Tailwind inline styles
				"img-src 'self' data: https: blob:",
				"font-src 'self' data:",
				"connect-src 'self' https://*.replicate.com https://api.anthropic.com https://*.vercel.app",
				"frame-ancestors " + frameAncestors,
				"form-action 'self'",
				"base-uri 'self'",
				"object-src 'none'",
			}, "; ")
			w.Header().Set("Content-Security-Policy", csp)
		}

		next.ServeHTTP(w, r)
	})
}

// corsMiddleware добавляет CORS headers.
// Читает CORS_ALLOWED_ORIGINS из env (comma-separated) и мерджит с defaults.
func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Defaults: localhost dev + Vercel production
		allowedOrigins := map[string]bool{
			"http://localhost:3000":               true,
			"http://localhost:5173":               true,
			"http://localhost:8080":               true,
			"https://istok-agent-core.vercel.app": true,
		}

		// Merge from CORS_ALLOWED_ORIGINS env (comma-separated)
		if extra := os.Getenv("CORS_ALLOWED_ORIGINS"); extra != "" {
			for _, o := range strings.Split(extra, ",") {
				o = strings.TrimSpace(o)
				if o != "" {
					allowedOrigins[o] = true
				}
			}
		}

		// Allow all *.vercel.app subdomains
		if origin != "" && len(origin) > 11 && origin[len(origin)-11:] == ".vercel.app" {
			allowedOrigins[origin] = true
		}

		// Устанавливаем CORS headers
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin == "" {
			// Для запросов без Origin (например, curl)
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Cache-Control, Connection")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type, Cache-Control, Connection, X-Accel-Buffering")
		w.Header().Set("X-Accel-Buffering", "no") // запретить буферизацию на ВСЕХ ответах (Railway/nginx)

		// Обработка preflight запросов
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// loggingMiddleware логирует все запросы
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем wrapper для ResponseWriter чтобы захватить status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf(
			"📝 %s %s - %d (%v)",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	})
}

// responseWriter оборачивает http.ResponseWriter для захвата status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Flush проксирует вызов к оригинальному ResponseWriter если он поддерживает http.Flusher.
// БЕЗ ЭТОГО flusher-проверка в SSE хендлере всегда падала → 500 за 81µs.
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// writeJSON сериализует data в JSON и отправляет ответ с application/json.
// НИКОГДА не возвращает HTML — это ломало парсер на фронте ("Unexpected token 'T'").
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// writeError отправляет правильно экранированный JSON с ошибкой.
// Использует encoding/json → безопасно для message с кавычками/переводами строк.
func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  message,
		"status": statusCode,
	})
}
