package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/djalben/istok-agent-core/internal/application"
	"github.com/djalben/istok-agent-core/internal/application/usecases"
	"github.com/djalben/istok-agent-core/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// Server - HTTP сервер.
type Server struct {
	addr             string
	projectGenerator *usecases.ProjectGeneratorService
	orchestrator     *application.Orchestrator
	watcher          *application.Watcher
	authService      *usecases.AuthService
	projectService   *usecases.ProjectService
	server           *http.Server
}

// NewServer создает HTTP сервер с LLM-провайдером (через порт) и сервисами Layer 1.
func NewServer(
	addr string,
	projectGenerator *usecases.ProjectGeneratorService,
	llm ports.LLMProvider,
	authService *usecases.AuthService,
	projectService *usecases.ProjectService,
	uiMedia ports.UIMediaService,
	orchOpts ...application.Option,
) *Server {
	orch := application.NewOrchestrator(llm, uiMedia, orchOpts...)
	watcher := application.NewWatcher(orch, "http://localhost"+addr)
	application.LogWatcherInitialized(context.Background(), watcher.MaxCreditsConfigured(), watcher.AutoHealEnabled())

	return &Server{
		addr:             addr,
		projectGenerator: projectGenerator,
		orchestrator:     orch,
		watcher:          watcher,
		authService:      authService,
		projectService:   projectService,
	}
}

// Watcher exposes the error webhook ring-buffer consumer (slog tee wiring in cmd/server).
func (s *Server) Watcher() *application.Watcher {
	return s.watcher
}

// Start запускает HTTP сервер.
func (s *Server) Start() error {
	startupCtx := context.Background()
	mux := http.NewServeMux()

	// Регистрация handlers
	generateHandler := NewGenerateHandler(s.projectGenerator)
	statsHandler := NewStatsHandler(s.projectGenerator)
	healthHandler := NewHealthHandler()
	authHandler := NewAuthHandler(s.authService)

	// ── SSE СТРИМ — регистрируем ПЕРВЫМ (более специфичный путь) ──
	sseHandler := NewGenerateHandlerSSE(s.orchestrator, s.projectService)
	// Layer 2: генерация требует аутентификации — owner_id гарантирован для авто-сохранения в БД.
	mux.HandleFunc("POST /api/v1/generate/stream", s.corsMiddleware(AuthMiddleware(s.authService, sseHandler.HandleStream)))
	mux.HandleFunc("OPTIONS /api/v1/generate/stream", s.corsMiddleware(sseHandler.HandleStream))
	logFrom(startupCtx).InfoContext(startupCtx, "route registered", "method", "POST", "path", "/api/v1/generate/stream", "handler", "GenerateHandlerSSE")

	// API endpoints
	mux.HandleFunc("POST /api/v1/generate", s.corsMiddleware(generateHandler.Handle))
	mux.HandleFunc("OPTIONS /api/v1/generate", s.corsMiddleware(generateHandler.Handle))
	mux.HandleFunc("/api/v1/stats", s.corsMiddleware(statsHandler.Handle))
	mux.HandleFunc("/api/v1/health", s.corsMiddleware(healthHandler.Handle))

	// Agents status — каноничный пайплайн для фронта (Zod-контракт)
	agentsStatusHandler := NewAgentsStatusHandler(s.orchestrator)
	mux.HandleFunc("/api/v1/agents/status", s.corsMiddleware(agentsStatusHandler.Handle))

	// Railway deploy integration
	deployHandler := NewDeployHandler()
	mux.HandleFunc("/api/v1/deploy/railway", s.corsMiddleware(deployHandler.HandleRailway))

	// ── Auth endpoints (signup/login открыты; me — за JWT-middleware) ──
	mux.HandleFunc("/api/v1/auth/signup", s.corsMiddleware(authHandler.HandleSignup))
	mux.HandleFunc("/api/v1/auth/login", s.corsMiddleware(authHandler.HandleLogin))
	mux.HandleFunc("/api/v1/auth/me", s.corsMiddleware(AuthMiddleware(s.authService, authHandler.HandleMe)))

	// ── Layer 1: Projects CRUD + remix (все за JWT-middleware) ──
	projectsHandler := NewProjectsHandler(s.projectService)
	protected := func(f http.HandlerFunc) http.HandlerFunc {
		return s.corsMiddleware(AuthMiddleware(s.authService, f))
	}
	corsOnly := func(_ http.ResponseWriter, _ *http.Request) {} // OPTIONS short-circuits в corsMiddleware

	mux.HandleFunc("GET /api/v1/projects", protected(projectsHandler.HandleList))
	mux.HandleFunc("POST /api/v1/projects", protected(projectsHandler.HandleCreate))
	mux.HandleFunc("OPTIONS /api/v1/projects", s.corsMiddleware(corsOnly))

	mux.HandleFunc("GET /api/v1/projects/{id}", protected(projectsHandler.HandleGet))
	mux.HandleFunc("PATCH /api/v1/projects/{id}", protected(projectsHandler.HandleUpdate))
	mux.HandleFunc("DELETE /api/v1/projects/{id}", protected(projectsHandler.HandleDelete))
	mux.HandleFunc("OPTIONS /api/v1/projects/{id}", s.corsMiddleware(corsOnly))

	mux.HandleFunc("POST /api/v1/projects/{id}/remix", protected(projectsHandler.HandleRemix))
	mux.HandleFunc("OPTIONS /api/v1/projects/{id}/remix", s.corsMiddleware(corsOnly))
	logFrom(startupCtx).InfoContext(startupCtx, "routes registered", "group", "projects")

	// ── Layer 1: User profile + Folders/Workspaces (stubs) ──
	profileHandler := NewProfileHandler(s.authService, s.projectService)
	mux.HandleFunc("/api/v1/user/profile", s.corsMiddleware(AuthMiddleware(s.authService, profileHandler.HandleProfile)))

	workspaceHandler := NewWorkspaceHandler()
	mux.HandleFunc("/api/v1/folders", s.corsMiddleware(AuthMiddleware(s.authService, workspaceHandler.HandleFolders)))
	mux.HandleFunc("/api/v1/workspaces", s.corsMiddleware(AuthMiddleware(s.authService, workspaceHandler.HandleWorkspaces)))

	// Diagnostic endpoints
	diagHandler := NewDiagHandler()
	mux.HandleFunc("/api/v1/diag/models", s.corsMiddleware(diagHandler.Handle))
	mux.HandleFunc("/api/v1/diag/env", s.corsMiddleware(diagHandler.HandleEnv))

	// Project export (ZIP download)
	exportHandler := NewExportHandler()
	mux.HandleFunc("/api/v1/project/export", s.corsMiddleware(exportHandler.HandleExport))

	// Human-in-the-Loop: architecture approval
	approvalHandler := NewApprovalHandler(s.orchestrator.GetApprovalRegistry())
	mux.HandleFunc("POST /api/v1/generate/approve", s.corsMiddleware(approvalHandler.Handle))
	mux.HandleFunc("OPTIONS /api/v1/generate/approve", s.corsMiddleware(approvalHandler.Handle))
	logFrom(startupCtx).InfoContext(startupCtx, "route registered", "method", "POST", "path", "/api/v1/generate/approve", "handler", "ApprovalHandler")

	// Human-in-the-Loop: media prompt approval (design review)
	mediaApprovalHandler := NewMediaApprovalHandler(s.orchestrator.GetApprovalRegistry())
	mux.HandleFunc("POST /api/v1/generate/approve_media", s.corsMiddleware(mediaApprovalHandler.Handle))
	mux.HandleFunc("OPTIONS /api/v1/generate/approve_media", s.corsMiddleware(mediaApprovalHandler.Handle))
	logFrom(startupCtx).InfoContext(startupCtx, "route registered", "method", "POST", "path", "/api/v1/generate/approve_media", "handler", "MediaApprovalHandler")

	// Media Studio: live image preview generation
	mediaPreviewHandler := NewMediaPreviewHandler(s.orchestrator.GetUIMedia())
	mux.HandleFunc("POST /api/v1/generate/media/preview", s.corsMiddleware(mediaPreviewHandler.Handle))
	mux.HandleFunc("OPTIONS /api/v1/generate/media/preview", s.corsMiddleware(mediaPreviewHandler.Handle))
	logFrom(startupCtx).InfoContext(startupCtx, "route registered", "method", "POST", "path", "/api/v1/generate/media/preview", "handler", "MediaPreviewHandler")

	// Pause & Resume: insufficient funds
	resumeFundsHandler := NewResumeFundsHandler(s.orchestrator.GetFundsRegistry())
	mux.HandleFunc("POST /api/v1/generate/resume_funds", s.corsMiddleware(resumeFundsHandler.Handle))
	mux.HandleFunc("OPTIONS /api/v1/generate/resume_funds", s.corsMiddleware(resumeFundsHandler.Handle))
	logFrom(startupCtx).InfoContext(startupCtx, "route registered", "method", "POST", "path", "/api/v1/generate/resume_funds", "handler", "ResumeFundsHandler")

	// File download endpoint (client fetches after SSE "done" event)
	filesHandler := NewFilesHandler()
	mux.HandleFunc("GET /api/v1/generate/files", s.corsMiddleware(filesHandler.Handle))
	mux.HandleFunc("OPTIONS /api/v1/generate/files", s.corsMiddleware(filesHandler.Handle))
	logFrom(startupCtx).InfoContext(startupCtx, "route registered", "method", "GET", "path", "/api/v1/generate/files", "handler", "FilesHandler")

	// Server-side preview build (region-proof): bundles the generated app on the server
	// (esm.sh + proxied Tailwind) so the user's browser loads everything same-origin.
	previewHandler := NewPreviewHandler()
	mux.HandleFunc("GET /api/v1/preview/tailwind.js", s.corsMiddleware(previewHandler.HandleTailwind))
	mux.HandleFunc("GET /api/v1/preview/{session_id}", s.corsMiddleware(previewHandler.Handle))
	mux.HandleFunc("OPTIONS /api/v1/preview/{session_id}", s.corsMiddleware(previewHandler.Handle))
	logFrom(startupCtx).InfoContext(startupCtx, "route registered", "method", "GET", "path", "/api/v1/preview/{session_id}", "handler", "PreviewHandler")

	// Prompt enhancer (Magic Wand)
	promptHelper := usecases.NewPromptHelper(s.orchestrator.GetLLM())
	promptHandler := NewPromptHandler(promptHelper)
	mux.HandleFunc("/api/v1/prompt/enhance", s.corsMiddleware(promptHandler.HandleEnhance))

	// Watcher V1 — error webhook + reports
	watcherHandler := NewWatcherHandler(s.watcher)
	mux.HandleFunc("/api/v1/internal/error-webhook", s.corsMiddleware(watcherHandler.HandleErrorWebhook))
	mux.HandleFunc("/api/v1/internal/watcher/reports", s.corsMiddleware(watcherHandler.HandleReports))

	// Interactive Editor Agent (Chat-to-Modify)
	editorUsecase := usecases.NewEditor(s.orchestrator.GetLLM())
	editorHandler := NewEditorHandler(editorUsecase)
	mux.HandleFunc("POST /api/v1/editor/chat", s.corsMiddleware(editorHandler.Handle))
	mux.HandleFunc("OPTIONS /api/v1/editor/chat", s.corsMiddleware(editorHandler.Handle))
	logFrom(startupCtx).InfoContext(startupCtx, "route registered", "method", "POST", "path", "/api/v1/editor/chat", "handler", "EditorHandler")

	// Surgical Component Edit (Inspector point-and-click)
	componentEditor := usecases.NewComponentEditor(s.orchestrator.GetLLM())
	editComponentHandler := NewEditComponentHandler(componentEditor)
	mux.HandleFunc("POST /api/v1/generate/edit", s.corsMiddleware(editComponentHandler.Handle))
	mux.HandleFunc("OPTIONS /api/v1/generate/edit", s.corsMiddleware(editComponentHandler.Handle))
	logFrom(startupCtx).InfoContext(startupCtx, "route registered", "method", "POST", "path", "/api/v1/generate/edit", "handler", "EditComponentHandler")
	logFrom(startupCtx).InfoContext(startupCtx, "http routes registered")

	// Catch-all 404 trap — ОБЯЗАТЕЛЬНО обёрнут в corsMiddleware,
	// иначе браузер блокирует ответ → фронт видит opaque ошибку вместо JSON.
	mux.HandleFunc("/", s.corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_ = writeJSON(w, http.StatusOK, map[string]string{
				"service": "istok-agent-core",
				"status":  "running",
				"version": "3.0.0",
			})

			return
		}
		logFrom(r.Context()).WarnContext(r.Context(), "route not found")
		writeError(w, http.StatusNotFound, fmt.Sprintf("Route not found: %s %s", r.Method, r.URL.Path))
	}))

	// Middleware chain: Recovery → SecurityHeaders → RequestLogger → Router
	handler := s.recoveryMiddleware(s.securityHeadersMiddleware(s.requestLoggerMiddleware(mux)))

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Minute,  // AI generation takes time
		WriteTimeout: 30 * time.Minute, // SSE chunked generation (112 files) needs ~22min; must exceed SSE ctx (25min)
		IdleTimeout:  120 * time.Second,
	}
	logFrom(startupCtx).InfoContext(startupCtx, "http server started", "addr", s.addr)

	err := s.server.ListenAndServe()
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// Shutdown gracefully останавливает сервер.
func (s *Server) Shutdown(ctx context.Context) error {
	logFrom(ctx).InfoContext(ctx, "http server shutting down")

	err := s.server.Shutdown(ctx)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// recoveryMiddleware перехватывает panic и логирует полный стектрейс в Railway.
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if rec := recover(); rec != nil {
				logFrom(ctx).ErrorContext(ctx, "panic recovered", "panic", rec)
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

		// ── Preview endpoints: self-contained sandboxed app, должен встраиваться в iframe ──
		// Отдаётся в <iframe sandbox> на фронте (Vercel) и грузит inline-бандл + Tailwind
		// Play CDN (eval). Строгие framing/CSP заголовки сломали бы предпросмотр полностью,
		// поэтому пропускаем их для /api/v1/preview (изоляция — через sandbox-атрибут iframe).
		isPreview := strings.HasPrefix(r.URL.Path, "/api/v1/preview")

		// ── X-Frame-Options: запретить iframe-embed по умолчанию ──
		// Если задан FRAME_ALLOWED_ORIGINS — используем CSP frame-ancestors вместо DENY.
		frameAllowed := os.Getenv("FRAME_ALLOWED_ORIGINS")
		switch {
		case isPreview:
			// no framing restriction — iframe-embed разрешён
		case frameAllowed == "":
			w.Header().Set("X-Frame-Options", "DENY")
		default:
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
		// Preview пропускаем: inline-бандл + Tailwind eval несовместимы с script-src 'self'.
		if !isPreview && !strings.HasPrefix(r.URL.Path, "/api/v1/generate/stream") {
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
// Читает CORS_ALLOWED_ORIGINS и ALLOWED_ORIGINS из env (comma-separated) и мерджит с defaults.
func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Defaults: localhost dev + Vercel production (включая текущий деплой-домен)
		allowedOrigins := map[string]bool{
			"http://localhost:3000":               true,
			"http://localhost:5173":               true,
			"http://localhost:8080":               true,
			"https://istok-agent-core.vercel.app": true,
			"https://istok-agent-core-7fvsc2jbd-djalbens-projects.vercel.app": true,
		}

		// Merge from CORS_ALLOWED_ORIGINS и ALLOWED_ORIGINS env (comma-separated)
		for _, envName := range []string{"CORS_ALLOWED_ORIGINS", "ALLOWED_ORIGINS"} {
			extra := os.Getenv(envName)
			if extra == "" {
				continue
			}
			for o := range strings.SplitSeq(extra, ",") {
				o = strings.TrimSpace(o)
				if o != "" {
					allowedOrigins[o] = true
				}
			}
		}

		// Allow arbitrary *.vercel.app preview subdomains ТОЛЬКО при явном opt-in.
		// С Allow-Credentials:true рефлексия любого *.vercel.app (включая чужие)
		// — риск; продакшн-origin остаётся в allowedOrigins по умолчанию.
		if os.Getenv("ALLOW_VERCEL_PREVIEWS") == "true" &&
			strings.HasPrefix(origin, "https://") && strings.HasSuffix(origin, ".vercel.app") {
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-Requested-With, Cache-Control, Connection")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type, Cache-Control, Connection, X-Accel-Buffering")
		w.Header().Set("X-Accel-Buffering", "no") // запретить буферизацию на ВСЕХ ответах (Railway/nginx)

		// Обработка preflight запросов
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)

			return
		}

		next(w, r)
	}
}

// responseWriter оборачивает http.ResponseWriter для захвата status code.
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
func writeJSON(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// writeError отправляет правильно экранированный JSON с ошибкой.
// Использует encoding/json → безопасно для message с кавычками/переводами строк.
func writeError(w http.ResponseWriter, statusCode int, message string) {
	_ = writeJSON(w, statusCode, map[string]any{
		"error":  message,
		"status": statusCode,
	})
}
