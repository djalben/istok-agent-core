package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/istok/agent-core/internal/application"
	"github.com/istok/agent-core/internal/application/usecases"
)

// Server - HTTP сервер
type Server struct {
	addr             string
	projectGenerator *usecases.ProjectGeneratorService
	orchestrator     *application.Orchestrator
	watcher          *application.Watcher
	server           *http.Server
}

// NewServer создает новый HTTP сервер
func NewServer(addr string, projectGenerator *usecases.ProjectGeneratorService) *Server {
	orch := application.NewOrchestrator()
	return &Server{
		addr:             addr,
		projectGenerator: projectGenerator,
		orchestrator:     orch,
		watcher:          application.NewWatcher(orch, "http://localhost"+addr),
	}
}

// NewServerWithKey создает HTTP сервер с API ключом для мультимодального оркестратора
func NewServerWithKey(addr string, projectGenerator *usecases.ProjectGeneratorService, apiKey string) *Server {
	orch := application.NewOrchestratorWithKey(apiKey)
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

	// Catch-all 404 trap — логирует ВСЕ неизвестные пути
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"service":"istok-agent-core","status":"running"}`)
			return
		}
		log.Printf("⚠️ 404 TRAP: %s %s (Origin: %s, UA: %s)", r.Method, r.URL.Path, r.Header.Get("Origin"), r.Header.Get("User-Agent"))
		writeError(w, http.StatusNotFound, fmt.Sprintf("Route not found: %s %s", r.Method, r.URL.Path))
	})

	// Middleware chain: Recovery → Logging → Router
	handler := s.recoveryMiddleware(s.loggingMiddleware(mux))

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Minute, // AI generation takes time
		WriteTimeout: 6 * time.Minute, // Must be > OpenRouter timeout (5min)
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

// corsMiddleware добавляет CORS headers
func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Разрешаем запросы от localhost (dev) и Vercel (production)
		allowedOrigins := map[string]bool{
			"http://localhost:3000":               true,
			"http://localhost:5173":               true,
			"https://vercel.app":                  true,
			"https://istok-agent-core.vercel.app": true,
			"https://istok-agent-core-dtoqkzr8x-djalbens-projects.vercel.app": true,
		}

		// Разрешаем все поддомены vercel.app + наш проект
		if origin != "" {
			if len(origin) > 11 && origin[len(origin)-11:] == ".vercel.app" {
				allowedOrigins[origin] = true
			}
			if origin == "https://istok-agent-core.vercel.app" {
				allowedOrigins[origin] = true
			}
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

// writeJSON отправляет JSON ответ
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Для простоты используем fmt.Fprintf, в production лучше encoding/json
	_, err := fmt.Fprintf(w, "%v", data)
	return err
}

// writeError отправляет JSON ошибку
func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"error": "%s"}`, message)
}
