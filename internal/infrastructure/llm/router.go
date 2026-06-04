package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/djalben/istok-agent-core/internal/ports"
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
	l := ports.LoggerFromContext(ctx)

	select {
	case <-ctx.Done():
		l.WarnContext(ctx, "cancelled before LLM call", "model", req.Model)

		return nil, fmt.Errorf("cancelled before LLM call: %w", ctx.Err())
	default:
	}

	if IsAnthropicModel(req.Model) {
		l.InfoContext(ctx, "routing to anthropic", "model", req.Model)

		return r.anthropic.Complete(ctx, req)
	}

	if isReplicateMediaOrText(req.Model) {
		l.InfoContext(ctx, "routing to replicate", "model", req.Model)

		return r.replicate.Complete(ctx, req)
	}

	l.WarnContext(ctx, "unknown model prefix, defaulting to anthropic", "model", req.Model)

	return r.anthropic.Complete(ctx, req)
}

// isReplicateMediaOrText определяет, является ли модель Replicate-моделью
// (медиа: google/nano-banana, google/veo-3, black-forest-labs/*, ideogram-ai/*,
//
//	reasoning: skywork/skywork-o1-open-llama-3.1-8b).
func isReplicateMediaOrText(model string) bool {
	lower := strings.ToLower(strings.TrimSpace(model))

	return strings.HasPrefix(lower, "google/") ||
		strings.HasPrefix(lower, "black-forest-labs/") ||
		strings.HasPrefix(lower, "ideogram-ai/") ||
		strings.HasPrefix(lower, "stability-ai/") ||
		strings.HasPrefix(lower, "meta/") ||
		strings.HasPrefix(lower, "deepseek-ai/") ||
		strings.HasPrefix(lower, "skywork/")
}

// IsReplicateModel — публичная проверка для совместимости с предыдущим API.
func IsReplicateModel(model string) bool {
	return isReplicateMediaOrText(model) && !IsAnthropicModel(model)
}
