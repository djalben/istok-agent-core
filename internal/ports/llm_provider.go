package ports

import (
	"context"
	"errors"
)

// ErrInsufficientFunds — sentinel error returned by LLM adapters when the
// provider reports exhausted credits (HTTP 402, 429 with quota text, etc.).
// The orchestrator catches this to pause generation until the user tops up.
var ErrInsufficientFunds = errors.New("insufficient funds: credit balance exhausted")

// LLMRequest запрос к LLM-провайдеру
type LLMRequest struct {
	Model        string
	SystemPrompt string
	UserPrompt   string
	MaxTokens    int
	Temperature  float64
	Reasoning    bool
}

// LLMResponse ответ от LLM-провайдера
type LLMResponse struct {
	Content    string
	TokensUsed int
	Model      string
}

// LLMProvider — порт для любого LLM-провайдера (Anthropic Direct, Replicate, и т.д.).
// Application-слой вызывает только этот интерфейс; конкретная реализация
// и HTTP-транспорт скрыты в infrastructure/.
type LLMProvider interface {
	// Complete отправляет prompt и возвращает текстовый ответ модели.
	Complete(ctx context.Context, req LLMRequest) (*LLMResponse, error)
}
