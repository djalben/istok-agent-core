package openrouter

import (
	"fmt"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — OpenRouter Configuration
//  S-Tier AI Squad 2026
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ModelConfig конфигурация модели
type ModelConfig struct {
	ID              string
	Name            string
	Provider        string
	Description     string
	MaxTokens       int
	Temperature     float64
	TopP            float64
	Timeout         time.Duration
	CostPer1K       float64 // Стоимость за 1000 токенов в рублях
	ThinkingEnabled bool    // Активировать extended thinking (Claude)
	ThinkingBudget  int     // Бюджет токенов для размышлений
}

// EstimateCost оценивает стоимость запроса
func (m *ModelConfig) EstimateCost(inputTokens int, outputTokens int) float64 {
	totalTokens := inputTokens + outputTokens
	return float64(totalTokens) / 1000 * m.CostPer1K
}

// AgentSquad конфигурация команды агентов
type AgentSquad struct {
	Director     ModelConfig
	Researcher   ModelConfig
	Coder        ModelConfig
	Designer     ModelConfig
	Videographer ModelConfig
}

// ThinkingSquad команда с активным режимом мышления
type ThinkingSquad struct {
	Brain ModelConfig // Claude Opus — Глубокое мышление
	Coder ModelConfig // DeepSeek-V3 — Реализация
}

// GetThinkingSquad возвращает команду с активным Claude Thinking режимом
func GetThinkingSquad() *ThinkingSquad {
	return &ThinkingSquad{
		Brain: ModelConfig{
			ID:              "anthropic/claude-opus-4-5",
			Name:            "Claude Opus 4.5 (Thinking)",
			Provider:        "Anthropic",
			Description:     "🧠 Мозг — Глубокий анализ, стратегия, архитектура. Extended Thinking активирован.",
			MaxTokens:       16000,
			Temperature:     1.0,
			TopP:            0.95,
			Timeout:         10 * time.Minute,
			CostPer1K:       15.0,
			ThinkingEnabled: true,
			ThinkingBudget:  10000,
		},
		Coder: ModelConfig{
			ID:          "deepseek/deepseek-v3",
			Name:        "DeepSeek-V3",
			Provider:    "DeepSeek",
			Description: "💻 Кодер — Clean Code по стандартам.",
			MaxTokens:   16384,
			Temperature: 0.3,
			TopP:        0.9,
			Timeout:     10 * time.Minute,
			CostPer1K:   0.5,
		},
	}
}

// GetSquad2026 возвращает конфигурацию S-Tier команды 2026
func GetSquad2026() *AgentSquad {
	return &AgentSquad{
		Director: ModelConfig{
			ID:          "anthropic/claude-3.5-sonnet",
			Name:        "Claude 3.5 Sonnet",
			Provider:    "Anthropic",
			Description: "🧠 Директор — Логика, архитектура, декомпозиция задач. Лучший для стратегического планирования и системного дизайна.",
			MaxTokens:   8192,
			Temperature: 0.7,
			TopP:        0.9,
			Timeout:     5 * time.Minute,
			CostPer1K:   3.0,
		},
		Researcher: ModelConfig{
			ID:          "google/gemini-2.0-pro",
			Name:        "Gemini 2.0 Pro",
			Provider:    "Google",
			Description: "🔍 Исследователь — Анализ URL, реверс-инжиниринг, технический аудит. Мультимодальный анализ с поддержкой изображений.",
			MaxTokens:   32768,
			Temperature: 0.5,
			TopP:        0.95,
			Timeout:     3 * time.Minute,
			CostPer1K:   1.5,
		},
		Coder: ModelConfig{
			ID:          "deepseek/deepseek-v3",
			Name:        "DeepSeek-V3",
			Provider:    "DeepSeek",
			Description: "💻 Кодер — Clean Code по стандартам. Специализируется на типизированном коде и best practices.",
			MaxTokens:   16384,
			Temperature: 0.3,
			TopP:        0.9,
			Timeout:     10 * time.Minute,
			CostPer1K:   0.5,
		},
		Designer: ModelConfig{
			ID:          "google/nano-banana-pro",
			Name:        "Nano Banana Pro",
			Provider:    "Google",
			Description: "🎨 Дизайнер — UI-ассеты и промпты для изображений. Генерация визуального контента.",
			MaxTokens:   4096,
			Temperature: 0.8,
			TopP:        0.95,
			Timeout:     5 * time.Minute,
			CostPer1K:   2.0,
		},
		Videographer: ModelConfig{
			ID:          "google/veo",
			Name:        "Veo",
			Provider:    "Google",
			Description: "🎬 Видеограф — Создание промо-видео. Генерация видеоконтента по текстовому описанию.",
			MaxTokens:   2048,
			Temperature: 0.9,
			TopP:        0.95,
			Timeout:     15 * time.Minute,
			CostPer1K:   10.0,
		},
	}
}

// OpenRouterConfig конфигурация OpenRouter API
type OpenRouterConfig struct {
	BaseURL     string
	APIKey      string
	HTTPReferer string
	AppName     string
	Timeout     time.Duration
}

// GetDefaultConfig возвращает конфигурацию по умолчанию
func GetDefaultConfig(apiKey string) *OpenRouterConfig {
	return &OpenRouterConfig{
		BaseURL:     "https://openrouter.ai/api/v1",
		APIKey:      apiKey,
		HTTPReferer: "https://istok-agent.vercel.app",
		AppName:     "ИСТОК Агент",
		Timeout:     30 * time.Minute,
	}
}

// ModelPricing информация о ценах моделей
type ModelPricing struct {
	Director     float64
	Researcher   float64
	Coder        float64
	Designer     float64
	Videographer float64
	Total        float64
}

// CalculatePricing рассчитывает стоимость генерации
func CalculatePricing(tokensUsed map[string]int) *ModelPricing {
	squad := GetSquad2026()

	pricing := &ModelPricing{
		Director:     float64(tokensUsed["director"]) / 1000 * squad.Director.CostPer1K,
		Researcher:   float64(tokensUsed["researcher"]) / 1000 * squad.Researcher.CostPer1K,
		Coder:        float64(tokensUsed["coder"]) / 1000 * squad.Coder.CostPer1K,
		Designer:     float64(tokensUsed["designer"]) / 1000 * squad.Designer.CostPer1K,
		Videographer: float64(tokensUsed["videographer"]) / 1000 * squad.Videographer.CostPer1K,
	}

	pricing.Total = pricing.Director + pricing.Researcher + pricing.Coder + pricing.Designer + pricing.Videographer

	return pricing
}

// GetModelByRole возвращает конфигурацию модели по роли
func GetModelByRole(role string) *ModelConfig {
	squad := GetSquad2026()

	switch role {
	case "director":
		return &squad.Director
	case "researcher":
		return &squad.Researcher
	case "coder":
		return &squad.Coder
	case "designer":
		return &squad.Designer
	case "videographer":
		return &squad.Videographer
	default:
		return &squad.Director
	}
}

// ValidateConfig проверяет корректность конфигурации
func ValidateConfig(config *OpenRouterConfig) error {
	if config.APIKey == "" {
		return ErrMissingAPIKey
	}
	if config.BaseURL == "" {
		return ErrInvalidBaseURL
	}
	return nil
}

// Errors
var (
	ErrMissingAPIKey  = fmt.Errorf("OpenRouter API key is required")
	ErrInvalidBaseURL = fmt.Errorf("invalid OpenRouter base URL")
)
