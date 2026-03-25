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
	"github.com/istok/agent-core/internal/infrastructure/openrouter"
	httpTransport "github.com/istok/agent-core/internal/transport/http"
)

func main() {
	log.Println("🚀 Запуск Исток Agent Core...")

	// Получаем API ключ из переменной окружения
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Println("⚠️  OPENROUTER_API_KEY не установлен, используется демо режим")
		apiKey = "demo-key"
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

	// Создаем инфраструктурные компоненты
	codeGeneratorAdapter := openrouter.NewCodeGeneratorAdapter(apiKey)
	webCrawler := crawler.NewSimpleCrawler()
	log.Println("✓ Инфраструктурные компоненты созданы")

	// Создаем use cases
	projectGenerator := usecases.NewProjectGeneratorService(
		agent,
		codeGeneratorAdapter,
		webCrawler,
	)
	log.Println("✓ Use Cases инициализированы")

	// Создаем HTTP сервер
	server := httpTransport.NewServer(":"+port, projectGenerator)

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

	// Запускаем сервер
	log.Printf("🌐 Сервер доступен на http://localhost:%s\n", port)
	log.Println("📡 API endpoints:")
	log.Println("   POST http://localhost:" + port + "/api/v1/generate")
	log.Println("   GET  http://localhost:" + port + "/api/v1/stats")
	log.Println("   GET  http://localhost:" + port + "/api/v1/health")
	log.Println("\n✨ Исток Agent готов к работе!")

	if err := server.Start(); err != nil {
		log.Fatalf("❌ Ошибка запуска сервера: %v\n", err)
	}
}
