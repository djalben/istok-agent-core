package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/djalben/istok-agent-core/internal/application"
	"github.com/djalben/istok-agent-core/internal/application/usecases"
	"github.com/djalben/istok-agent-core/internal/config"
	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/infrastructure/crawler"
	"github.com/djalben/istok-agent-core/internal/infrastructure/llm"
	logHandler "github.com/djalben/istok-agent-core/internal/infrastructure/logger/handler"
	"github.com/djalben/istok-agent-core/internal/infrastructure/media"
	"github.com/djalben/istok-agent-core/internal/infrastructure/persistence"
	"github.com/djalben/istok-agent-core/internal/ports"
	httpTransport "github.com/djalben/istok-agent-core/internal/transport/http"
	"gitlab.com/libs-artifex/wrapper"
)

func isNumericPort(p string) bool {
	if p == "" {
		return false
	}
	for _, c := range p {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

func main() {
	cfg, err := config.Parse()
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to parse config: " + err.Error() + "\n")
		os.Exit(1)
	}

	h := logHandler.Create(cfg.LogPlain, cfg.LogLevel)
	logger := slog.New(h)
	slog.SetDefault(logger)

	logger.Info("🚀 Запуск Исток Agent Core...")

	if cfg.IsProduction() {
		logger.Info("🏭 Mode: PRODUCTION")
	} else {
		logger.Info("🔧 Mode: DEVELOPMENT")
	}

	type envCheck struct {
		name     string
		value    string
		required bool
	}
	checks := []envCheck{
		{"ANTHROPIC_API_KEY", cfg.AnthropicKey, true},
		{"REPLICATE_API_TOKEN", cfg.ReplicateKey, true},
		{"CORS_ALLOWED_ORIGINS", cfg.CORSAllowedOrigins, false},
		{"JWT_SECRET", cfg.JWTSecret, false},
	}

	missing := 0
	for _, c := range checks {
		if c.value == "" {
			if c.required {
				logger.Warn("🚨 MISSING (required)", "env", c.name)
				missing++
			} else {
				logger.Warn("⚠️  MISSING (optional) — using default", "env", c.name)
			}
		} else {
			logger.Info("✅ configured", "env", c.name)
		}
	}
	if missing > 0 && cfg.IsProduction() {
		logger.Warn("🚨 required env vars missing — AI requests will fail", "count", missing)
	}

	anthropicKey := cfg.AnthropicKey
	if anthropicKey == "" {
		anthropicKey = "MISSING_KEY_CHECK_RAILWAY_ENV"
	}

	port := cfg.Port
	if port == "" || !isNumericPort(port) {
		port = "8080"
	}

	logger.Info("📦 Инициализация зависимостей...")

	agent := domain.NewAgent("agent-001", "Исток", 100000)
	logger.Info("✓ Агент создан", "name", agent.Name, "balance", agent.TokenBalance)

	agent.AddCapability(domain.NewCapability(
		"web_crawler",
		"Анализ сайтов и извлечение паттернов",
		domain.CapabilityAdvanced,
	))
	agent.AddCapability(domain.NewCapability(
		"code_synthesis",
		"Генерация production-ready кода",
		domain.CapabilityExpert,
	))
	logger.Info("✓ Способности добавлены", "count", len(agent.Capabilities))

	anthropicAdapter := llm.NewAnthropicAdapter(anthropicKey)
	replicateAdapter := llm.NewReplicateAdapter(cfg.ReplicateKey)
	llmProvider := llm.NewDualRouter(anthropicAdapter, replicateAdapter)
	logger.Info("✓ LLM инфраструктура создана (DualRouter: Anthropic Direct + Replicate)")

	codeGeneratorAdapter := llm.NewCodeGeneratorAdapter(llmProvider, "anthropic/claude-opus-4-7")
	webCrawler := crawler.NewSimpleCrawler()
	logger.Info("✓ Инфраструктурные компоненты созданы")

	projectGenerator := usecases.NewProjectGeneratorService(
		agent,
		codeGeneratorAdapter,
		webCrawler,
	)
	logger.Info("✓ Use Cases инициализированы")

	var userRepo ports.UserRepository
	var projectRepo ports.ProjectRepository
	if cfg.DatabaseURL != "" {
		pg, pgErr := persistence.NewPostgres(context.Background(), cfg.DatabaseURL)
		if pgErr != nil {
			logger.Warn("⚠️ Postgres init failed — откат на in-memory", "error", wrapper.Wrap(pgErr))
			userRepo = persistence.NewUserRepoMemory()
			projectRepo = persistence.NewProjectRepoMemory()
		} else {
			userRepo = persistence.NewUserRepoPostgres(pg.Pool)
			projectRepo = persistence.NewProjectRepoPostgres(pg.Pool)
			logger.Info("✓ Persistence: PostgreSQL (DATABASE_URL)")
		}
	} else {
		userRepo = persistence.NewUserRepoMemory()
		projectRepo = persistence.NewProjectRepoMemory()
		logger.Warn("⚠️ Persistence: in-memory fallback (DATABASE_URL не задан)")
	}
	authService := usecases.NewAuthService(userRepo, cfg.JWTSecret)
	projectService := usecases.NewProjectService(projectRepo)
	logger.Info("✓ Layer 1 сервисы инициализированы (Auth + Projects)")

	uiMedia := media.NewUIMediaService(llmProvider)
	server := httpTransport.NewServer(":"+port, projectGenerator, llmProvider, authService, projectService, uiMedia)
	tee := &application.WatcherLogWriter{Original: os.Stdout, Watcher: server.Watcher()}
	slog.SetDefault(slog.New(logHandler.CreateWithWriter(cfg.LogPlain, cfg.LogLevel, tee)))

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Info("⏳ Получен сигнал остановки...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownErr := server.Shutdown(ctx)
		if shutdownErr != nil {
			logger.Error("❌ Ошибка при остановке сервера", "error", wrapper.Wrap(shutdownErr))
		}

		logger.Info("✓ Сервер остановлен")
		os.Exit(0)
	}()

	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Info("  ИСТОК AGENT CORE v3.0.0 — Startup Banner")
	logger.Info("  BUILD: 10-agent pipeline + Verification Gate")
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	agents := []struct{ role, model, provider string }{
		{"Director", "claude-opus-4-7 (adaptive thinking)", "Anthropic Direct"},
		{"Researcher", "claude-opus-4-7 (adaptive thinking)", "Anthropic Direct"},
		{"Brain", "claude-opus-4-7 (adaptive thinking)", "Anthropic Direct"},
		{"Architect", "claude-opus-4-7 (adaptive thinking)", "Anthropic Direct"},
		{"Planner", "claude-opus-4-7 (adaptive thinking)", "Anthropic Direct"},
		{"Coder", "claude-opus-4-7", "Anthropic Direct"},
		{"Designer", "google/nano-banana", "Replicate"},
		{"Security", "claude-opus-4-7", "Anthropic Direct"},
		{"Tester", "local + claude-opus-4-7", "Anthropic Direct"},
		{"UI Reviewer", "claude-opus-4-7", "Anthropic Direct"},
	}
	for i, a := range agents {
		logger.Info(
			"agent ready",
			"index", i+1,
			"total", 10,
			"role", a.role,
			"model", a.model,
			"provider", a.provider,
		)
	}
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Info(
		"pipeline ready",
		"fsm_states", 12,
		"verification_gate", "Security ∧ Tester ∧ UI Reviewer",
		"sse_agent_field", true,
		"auto_fix_max_retries", 2,
	)
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	logger.Info("🌐 Сервер доступен на http://localhost (см. PORT)", "port", port)
	logger.Info(
		"📡 API endpoints",
		"health", "GET /api/v1/health",
		"generate", "POST /api/v1/generate",
		"generate_stream", "POST /api/v1/generate/stream",
		"stats", "GET /api/v1/stats",
		"diag_models", "GET /api/v1/diag/models",
		"diag_env", "GET /api/v1/diag/env",
	)
	if cfg.IsProduction() {
		logger.Info("🏭 Production mode", "log_level", cfg.LogLevel)
	}
	logger.Info("✨ Исток Agent v3.0.0 — все 10 агентов инициализированы и готовы к работе!")

	startErr := server.Start()
	if startErr != nil {
		logger.Error("❌ Ошибка запуска сервера", "error", wrapper.Wrap(startErr))
		os.Exit(1)
	}
}
