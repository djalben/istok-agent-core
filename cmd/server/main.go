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
	infralogger "github.com/djalben/istok-agent-core/internal/infrastructure/logger"
	logHandler "github.com/djalben/istok-agent-core/internal/infrastructure/logger/handler"
	"github.com/djalben/istok-agent-core/internal/infrastructure/media"
	"github.com/djalben/istok-agent-core/internal/infrastructure/persistence"
	"github.com/djalben/istok-agent-core/internal/ports"
	httpTransport "github.com/djalben/istok-agent-core/internal/transport/http"
	"gitlab.com/libs-artifex/wrapper/v2"
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
	infralogger.SetRoot(logger)
	ports.SetFallbackLogger(infralogger.Root)
	startupCtx := context.Background()

	logger.InfoContext(startupCtx, "istok agent core starting")

	if cfg.IsProduction() {
		logger.InfoContext(startupCtx, "mode production")
	} else {
		logger.InfoContext(startupCtx, "mode development")
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
				logger.WarnContext(startupCtx, "env missing required", "env", c.name)
				missing++
			} else {
				logger.WarnContext(startupCtx, "env missing optional", "env", c.name)
			}
		} else {
			logger.InfoContext(startupCtx, "env configured", "env", c.name)
		}
	}
	if missing > 0 && cfg.IsProduction() {
		logger.WarnContext(startupCtx, "required env vars missing", "count", missing)
	}

	anthropicKey := cfg.AnthropicKey
	if anthropicKey == "" {
		anthropicKey = "MISSING_KEY_CHECK_RAILWAY_ENV"
	}

	port := cfg.Port
	if port == "" || !isNumericPort(port) {
		port = "8080"
	}

	logger.InfoContext(startupCtx, "initializing dependencies")

	agent := domain.NewAgent("agent-001", "Исток", 100000)
	logger.InfoContext(startupCtx, "agent created", "name", agent.Name, "balance", agent.TokenBalance)

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
	logger.InfoContext(startupCtx, "capabilities added", "count", len(agent.Capabilities))

	anthropicAdapter := llm.NewAnthropicAdapter(anthropicKey)
	replicateAdapter := llm.NewReplicateAdapter(cfg.ReplicateKey)
	llmProvider := llm.NewDualRouter(anthropicAdapter, replicateAdapter)
	logger.InfoContext(startupCtx, "llm infrastructure ready", "router", "DualRouter")

	codeGeneratorAdapter := llm.NewCodeGeneratorAdapter(llmProvider, "anthropic/claude-opus-4-7")
	webCrawler := crawler.NewSimpleCrawler()
	logger.InfoContext(startupCtx, "infrastructure components ready")

	projectGenerator := usecases.NewProjectGeneratorService(
		agent,
		codeGeneratorAdapter,
		webCrawler,
	)
	logger.InfoContext(startupCtx, "use cases initialized")

	var userRepo ports.UserRepository
	var projectRepo ports.ProjectRepository
	if cfg.DatabaseURL != "" {
		pg, pgErr := persistence.NewPostgres(context.Background(), cfg.DatabaseURL)
		if pgErr != nil {
			logger.WarnContext(startupCtx, "postgres init failed, using memory", "error", wrapper.Wrap(pgErr))
			userRepo = persistence.NewUserRepoMemory()
			projectRepo = persistence.NewProjectRepoMemory()
		} else {
			userRepo = persistence.NewUserRepoPostgres(pg.Pool)
			projectRepo = persistence.NewProjectRepoPostgres(pg.Pool)
			logger.InfoContext(startupCtx, "persistence postgres", "driver", "DATABASE_URL")
		}
	} else {
		userRepo = persistence.NewUserRepoMemory()
		projectRepo = persistence.NewProjectRepoMemory()
		logger.WarnContext(startupCtx, "persistence in-memory fallback")
	}
	authService := usecases.NewAuthService(userRepo, cfg.JWTSecret)
	projectService := usecases.NewProjectService(projectRepo)
	logger.InfoContext(startupCtx, "layer1 services ready")

	uiMedia := media.NewUIMediaService(llmProvider)
	server := httpTransport.NewServer(":"+port, projectGenerator, llmProvider, authService, projectService, uiMedia)
	tee := &application.WatcherLogWriter{Original: os.Stdout, Watcher: server.Watcher()}
	logger = slog.New(logHandler.CreateWithWriter(cfg.LogPlain, cfg.LogLevel, tee))
	infralogger.SetRoot(logger)
	ports.SetFallbackLogger(infralogger.Root)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		logger.InfoContext(ctx, "shutdown signal received")

		shutdownErr := server.Shutdown(ctx)
		if shutdownErr != nil {
			logger.ErrorContext(ctx, "server shutdown failed", "error", wrapper.Wrap(shutdownErr))
		}

		logger.InfoContext(ctx, "server stopped")
		os.Exit(0)
	}()

	logger.InfoContext(startupCtx, "startup banner begin")
	logger.InfoContext(startupCtx, "istok agent core version", "version", "3.0.0", "build", "10-agent pipeline + Verification Gate")
	logger.InfoContext(startupCtx, "startup banner end")
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
		logger.InfoContext(startupCtx,
			"agent ready",
			"index", i+1,
			"total", 10,
			"role", a.role,
			"model", a.model,
			"provider", a.provider,
		)
	}
	logger.InfoContext(startupCtx, "pipeline ready",
		"fsmStates", 12,
		"verificationGate", "Security ∧ Tester ∧ UI Reviewer",
		"sseAgentField", true,
		"autoFixMaxRetries", 2,
	)

	logger.InfoContext(startupCtx, "server listening", "port", port)
	logger.InfoContext(startupCtx, "api endpoints ready",
		"health", "GET /api/v1/health",
		"generate", "POST /api/v1/generate",
		"generateStream", "POST /api/v1/generate/stream",
		"stats", "GET /api/v1/stats",
		"diagModels", "GET /api/v1/diag/models",
		"diagEnv", "GET /api/v1/diag/env",
	)
	if cfg.IsProduction() {
		logger.InfoContext(startupCtx, "production mode", "logLevel", cfg.LogLevel)
	}
	logger.InfoContext(startupCtx, "all agents initialized")

	startErr := server.Start()
	if startErr != nil {
		logger.ErrorContext(startupCtx, "server start failed", "error", wrapper.Wrap(startErr))
		os.Exit(1)
	}
}
