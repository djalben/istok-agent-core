package llm

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/istok/agent-core/internal/ports"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Dual Router (Anthropic + Replicate)
//  Anthropic Direct API: text/code/reasoning models.
//  Replicate: media generation (nano-banana, Veo 3).
//  OpenRouter полностью удалён.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// DualRouter маршрутизирует запросы между Anthropic Direct API
// и Replicate (медиа-генерация) на основе префикса модели.
type DualRouter struct {
	anthropic *AnthropicAdapter
	replicate *ReplicateAdapter
}

// NewDualRouter создаёт маршрутизатор с двумя бэкендами.
func NewDualRouter(anthropic *AnthropicAdapter, replicate *ReplicateAdapter) *DualRouter {
	return &DualRouter{
		anthropic: anthropic,
		replicate: replicate,
	}
}

// Complete маршрутизирует запрос к нужному провайдеру.
//   - anthropic/* | claude-* → Anthropic Direct API
//   - всё остальное (google/, black-forest-labs/, ideogram, …) → Replicate
func (r *DualRouter) Complete(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	select {
	case <-ctx.Done():
		log.Printf("⛔ ОТМЕНА: клиент отключился до вызова LLM model=%s", req.Model)
		return nil, fmt.Errorf("cancelled before LLM call: %w", ctx.Err())
	default:
	}

	if IsAnthropicModel(req.Model) {
		log.Printf("🔀 Routing %s → Anthropic Direct", req.Model)
		return r.anthropic.Complete(ctx, req)
	}

	if isReplicateMediaOrText(req.Model) {
		log.Printf("🔀 Routing %s → Replicate", req.Model)
		return r.replicate.Complete(ctx, req)
	}

	// Неизвестный префикс — по умолчанию Anthropic (text-first контракт).
	log.Printf("⚠️ Unknown model prefix %q — defaulting to Anthropic", req.Model)
	return r.anthropic.Complete(ctx, req)
}

// isReplicateMediaOrText определяет, является ли модель Replicate-моделью
// (медиа: google/nano-banana, google/veo-3, black-forest-labs/*, ideogram-ai/*).
func isReplicateMediaOrText(model string) bool {
	lower := strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(lower, "google/") ||
		strings.HasPrefix(lower, "black-forest-labs/") ||
		strings.HasPrefix(lower, "ideogram-ai/") ||
		strings.HasPrefix(lower, "stability-ai/") ||
		strings.HasPrefix(lower, "meta/") ||
		strings.HasPrefix(lower, "deepseek-ai/")
}

// IsReplicateModel — публичная проверка для совместимости с предыдущим API.
func IsReplicateModel(model string) bool {
	return isReplicateMediaOrText(model) && !IsAnthropicModel(model)
}
