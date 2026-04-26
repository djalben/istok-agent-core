package ports

import "context"

// LLMRequest запрос к LLM-провайдеру
type LLMRequest struct {
	Model          string
	SystemPrompt   string
	UserPrompt     string
	MaxTokens      int
	Temperature    float64
	Reasoning      bool
	ThinkingBudget int
}

// LLMResponse ответ от LLM-провайдера
type LLMResponse struct {
	Content    string
	TokensUsed int
	Model      string
}

// LLMProvider — порт для любого LLM-провайдера (OpenRouter, Replicate, и т.д.).
// Application-слой вызывает только этот интерфейс; конкретная реализация
// и HTTP-транспорт скрыты в infrastructure/.
type LLMProvider interface {
	// Complete отправляет prompt и возвращает текстовый ответ модели.
	Complete(ctx context.Context, req LLMRequest) (*LLMResponse, error)
}
