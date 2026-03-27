package openrouter

import "time"

// ModelHealth информация о здоровье модели
type ModelHealth struct {
	ModelID          string
	IsAvailable      bool
	LastChecked      time.Time
	FailureCount     int
	SuccessCount     int
	AverageLatency   time.Duration
	LastError        string
	ConsecutiveFails int
}

// FallbackStrategy стратегия переключения между моделями
type FallbackStrategy struct {
	Models           []ModelConfig
	MaxRetries       int
	TimeoutPerModel  time.Duration
	CostThreshold    float64
	QualityThreshold float64
	PreferFast       bool
	PreferCheap      bool
}

// ModelRegistry реестр доступных моделей
var ModelRegistry = map[string]ModelConfig{
	"anthropic/claude-3.5-sonnet": {
		ID:          "anthropic/claude-3.5-sonnet",
		Name:        "Claude 3.5 Sonnet",
		Provider:    "Anthropic",
		Description: "🧠 Директор — Логика, архитектура, декомпозиция задач",
		MaxTokens:   8192,
		Temperature: 0.7,
		TopP:        0.9,
		Timeout:     5 * time.Minute,
		CostPer1K:   3.0,
	},
	"google/gemini-2.0-pro": {
		ID:          "google/gemini-2.0-pro",
		Name:        "Gemini 2.0 Pro",
		Provider:    "Google",
		Description: "🔍 Исследователь — Анализ URL, реверс-инжиниринг",
		MaxTokens:   32768,
		Temperature: 0.5,
		TopP:        0.95,
		Timeout:     3 * time.Minute,
		CostPer1K:   1.5,
	},
	"deepseek/deepseek-v3": {
		ID:          "deepseek/deepseek-v3",
		Name:        "DeepSeek-V3",
		Provider:    "DeepSeek",
		Description: "💻 Кодер — Clean Code по стандартам",
		MaxTokens:   16384,
		Temperature: 0.3,
		TopP:        0.9,
		Timeout:     10 * time.Minute,
		CostPer1K:   0.5,
	},
}

// GetDefaultFallbackStrategy возвращает стратегию по умолчанию
func GetDefaultFallbackStrategy() *FallbackStrategy {
	squad := GetSquad2026()
	return &FallbackStrategy{
		Models: []ModelConfig{
			squad.Director,
			squad.Coder,
		},
		MaxRetries:       3,
		TimeoutPerModel:  5 * time.Minute,
		CostThreshold:    10.0,
		QualityThreshold: 0.8,
		PreferFast:       false,
		PreferCheap:      false,
	}
}
