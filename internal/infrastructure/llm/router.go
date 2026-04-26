package llm

import (
	"context"
	"fmt"
	"log"

	"github.com/istok/agent-core/internal/ports"
)

// DualRouter реализует ports.LLMProvider, маршрутизируя запросы
// к Replicate (Anthropic/Google модели) или OpenRouter (DeepSeek/Qwen и др.)
// в зависимости от префикса модели.
type DualRouter struct {
	replicate  *ReplicateAdapter
	openrouter *OpenRouterAdapter
}

// NewDualRouter создаёт маршрутизатор с двумя бэкендами.
func NewDualRouter(replicate *ReplicateAdapter, openrouter *OpenRouterAdapter) *DualRouter {
	return &DualRouter{
		replicate:  replicate,
		openrouter: openrouter,
	}
}

// Complete маршрутизирует запрос к нужному провайдеру и возвращает ответ.
func (r *DualRouter) Complete(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	// Проверка: если клиент уже отключился — не тратим кредиты
	select {
	case <-ctx.Done():
		log.Printf("⛔ ОТМЕНА: клиент отключился до вызова LLM model=%s", req.Model)
		return nil, fmt.Errorf("cancelled before LLM call: %w", ctx.Err())
	default:
	}

	if IsReplicateModel(req.Model) {
		log.Printf("🔀 Routing %s → Replicate", req.Model)
		return r.replicate.Complete(ctx, req)
	}

	log.Printf("🔀 Routing %s → OpenRouter", req.Model)
	return r.openrouter.Complete(ctx, req)
}
