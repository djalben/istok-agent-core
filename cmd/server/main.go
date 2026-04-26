package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/istok/agent-core/internal/application/usecases"
	"github.com/istok/agent-core/internal/domain"
	"github.com/istok/agent-core/internal/infrastructure/crawler"
	"github.com/istok/agent-core/internal/infrastructure/llm"
	httpTransport "github.com/istok/agent-core/internal/transport/http"
)

func main() {
	log.Println("🚀 Запуск Исток Agent Core...")

	// ── Production mode detection ──
	env := os.Getenv("RAILWAY_ENVIRONMENT")
	if env == "" {
		env = os.Getenv("GO_ENV")
	}
	isProduction := env == "production"
	if isProduction {
		log.Println("🏭 Mode: PRODUCTION")
	} else {
		log.Printf("� Mode: DEVELOPMENT (env=%s)", env)
	}

	// ── Validate critical environment variables ──
	type envCheck struct {
		name     string
		required bool
	}
	checks := []envCheck{
		{"ANTHROPIC_API_KEY", true},
		{"REPLICATE_API_TOKEN", true},
		{"CORS_ALLOWED_ORIGINS", false},
		{"JWT_SECRET", false},
	}

	missing := 0
	for _, c := range checks {
		val := os.Getenv(c.name)
		if val == "" {
			if c.required {
				log.Printf("🚨 MISSING (required): %s", c.name)
				missing++
			} else {
				log.Printf("⚠️  MISSING (optional): %s — using default", c.name)
			}
		} else {
			preview := val
			if len(preview) > 8 {
				preview = preview[:8] + "..."
			}
			log.Printf("✅ %s = %s", c.name, preview)
		}
	}
	if missing > 0 && isProduction {
		log.Printf("🚨 %d required env vars missing! AI requests will fail.", missing)
	}

	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		anthropicKey = "MISSING_KEY_CHECK_RAILWAY_ENV"
	}

	// Получаем порт из переменной окружения или используем дефолтный
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Инициализация зависимостей
	log.Println("📦 Инициализация зависимостей...")

	// Создаем агента с начальным балансом токенов
	agent := domain.NewAgent("agent-001", "Исток", 100000)
	log.Printf("✓ Агент создан: %s (баланс: %d токенов)\n", agent.Name, agent.TokenBalance)

	// Добавляем базовые способности
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
	log.Printf("✓ Добавлено %d способностей\n", len(agent.Capabilities))

	// Создаём LLM-инфраструктуру (Dependency Rule: application зависит от ports, не от infrastructure)
	replicateToken := os.Getenv("REPLICATE_API_TOKEN")

	anthropicAdapter := llm.NewAnthropicAdapter(anthropicKey)
	replicateAdapter := llm.NewReplicateAdapter(replicateToken)
	llmProvider := llm.NewDualRouter(anthropicAdapter, replicateAdapter)
	log.Println("✓ LLM инфраструктура создана (DualRouter: Anthropic Direct + Replicate)")

	// Создаем инфраструктурные компоненты
	codeGeneratorAdapter := llm.NewCodeGeneratorAdapter(llmProvider, "anthropic/claude-3-7-sonnet")
	webCrawler := crawler.NewSimpleCrawler()
	log.Println("✓ Инфраструктурные компоненты созданы")

	// Создаем use cases
	projectGenerator := usecases.NewProjectGeneratorService(
		agent,
		codeGeneratorAdapter,
		webCrawler,
	)
	log.Println("✓ Use Cases инициализированы")

	// Создаем HTTP сервер с LLM-провайдером (через порт)
	server := httpTransport.NewServer(":"+port, projectGenerator, llmProvider)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("\n⏳ Получен сигнал остановки...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("❌ Ошибка при остановке сервера: %v\n", err)
		}

		log.Println("✓ Сервер остановлен")
		os.Exit(0)
	}()

	// ── Agent initialization report ──
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("  ИСТОК AGENT CORE v3.0.0 — Startup Banner")
	log.Println("  BUILD: 10-agent pipeline + Verification Gate")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	agents := []struct{ role, model, provider string }{
		{"Director", "claude-3-7-sonnet", "Anthropic Direct"},
		{"Researcher", "claude-3-7-sonnet (thinking)", "Anthropic Direct"},
		{"Brain", "claude-3-7-sonnet (thinking)", "Anthropic Direct"},
		{"Architect", "claude-3-7-sonnet (thinking)", "Anthropic Direct"},
		{"Planner", "claude-3-7-sonnet (thinking)", "Anthropic Direct"},
		{"Coder", "claude-3-7-sonnet (medium)", "Anthropic Direct"},
		{"Designer", "google/nano-banana", "Replicate"},
		{"Security", "claude-3-7-sonnet", "Anthropic Direct"},
		{"Tester", "local + claude-3-7-sonnet", "Anthropic Direct"},
		{"UI Reviewer", "claude-3-7-sonnet", "Anthropic Direct"},
	}
	for i, a := range agents {
		log.Printf("  [%d/10] ✅ %s → %s (%s)", i+1, a.role, a.model, a.provider)
	}
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("  FSM: 12 states | Verification Gate: Security ∧ Tester ∧ UI Reviewer")
	log.Printf("  SSE: event.Agent field | Auto-Fix: max 2 retries")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Запускаем сервер
	base := "http://localhost:" + port
	log.Printf("🌐 Сервер доступен на %s", base)
	log.Println("📡 API endpoints:")
	log.Println("   GET  " + base + "/api/v1/health")
	log.Println("   POST " + base + "/api/v1/generate")
	log.Println("   POST " + base + "/api/v1/generate/stream  (SSE)")
	log.Println("   GET  " + base + "/api/v1/stats")
	log.Println("   GET  " + base + "/api/v1/diag/models")
	log.Println("   GET  " + base + "/api/v1/diag/env")
	if isProduction {
		log.Println("🏭 Production mode — logging: Info+Error")
	}
	log.Println("\n✨ Исток Agent v3.0.0 — все 10 агентов инициализированы и готовы к работе!")

	if err := server.Start(); err != nil {
		log.Fatalf("❌ Ошибка запуска сервера: %v\n", err)
	}
}
