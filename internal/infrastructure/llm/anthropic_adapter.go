package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/djalben/istok-agent-core/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Anthropic Direct Adapter
//  Прямая интеграция с Anthropic Messages API.
//  Sonnet 4.6 (дефолт) + Opus 4.8 (Architect/Planner). Adaptive Thinking + effort.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const (
	anthropicBaseURL = "https://api.anthropic.com/v1"
	anthropicVersion = "2023-06-01"

	// ModelClaudeSonnet46 — дефолтная модель (Coder/Researcher/Validator).
	ModelClaudeSonnet46         = "claude-sonnet-4-6"
	ModelClaudeSonnet46Thinking = "claude-sonnet-4-6-thinking" // alias → adaptive thinking
	// ModelClaudeOpus48 — флагман для Architect/Planner (максимум интеллекта).
	ModelClaudeOpus48 = "claude-opus-4-8"

	// Нативные лимиты output-токенов синхронного Messages API (без beta-заголовков):
	// Opus 4.8 → 128k, Sonnet 4.6 → 64k.
	maxOutputOpus48   = 128000
	maxOutputSonnet46 = 64000

	// defaultEffort — API-дефолт Anthropic (Opus 4.8 / Sonnet 4.6) при пустом req.Effort.
	defaultEffort = ports.EffortHigh
)

// AnthropicAdapter реализует ports.LLMProvider через Anthropic Messages API.
type AnthropicAdapter struct {
	apiKey     string
	httpClient *http.Client
}

// NewAnthropicAdapter создаёт адаптер для Anthropic API.
func NewAnthropicAdapter(apiKey string) *AnthropicAdapter {
	return &AnthropicAdapter{
		apiKey: apiKey,
		// Без фиксированного Client.Timeout: дедлайн задаётся per-request через
		// context, масштабируемый по effort (см. effortTimeout).
		httpClient: &http.Client{},
	}
}

type anthropicContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

type anthropicResponse struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Role       string                  `json:"role"`
	Content    []anthropicContentBlock `json:"content"`
	Model      string                  `json:"model"`
	StopReason string                  `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Complete реализует ports.LLMProvider — прямой вызов Anthropic Messages API.
// Поддерживает Adaptive Thinking API: активируется если req.Reasoning=true
// либо модель содержит суффикс "-thinking".
func (a *AnthropicAdapter) Complete(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	if a.apiKey == "" {
		return nil, ErrAnthropicAPIKeyNotConfigured
	}

	model, thinking := resolveAnthropicModel(req.Model, req.Reasoning)
	effort := resolveEffort(req.Effort)

	// Клампим max_tokens к нативному лимиту модели (Opus 4.8 → 128k, Sonnet 4.6 → 64k).
	modelMax := maxOutputFor(model)
	maxTokens := req.MaxTokens
	if maxTokens <= 0 || maxTokens > modelMax {
		maxTokens = modelMax
	}

	// Per-request дедлайн, масштабируемый по effort (high/xhigh/max думают дольше).
	ctx, cancel := context.WithTimeout(ctx, effortTimeout(effort))
	defer cancel()

	payload := map[string]any{
		"model":      model,
		"max_tokens": maxTokens,
		// effort заменяет deprecated budget_tokens: управляет глубиной thinking и расходом токенов.
		"output_config": map[string]any{"effort": effort},
	}
	_ = req.Temperature // в adaptive thinking temperature не передаём

	if req.SystemPrompt != "" {
		payload["system"] = req.SystemPrompt
	}

	payload["messages"] = []map[string]any{
		{
			"role":    "user",
			"content": req.UserPrompt,
		},
	}

	if thinking {
		// Adaptive Thinking (Opus 4.8 / Sonnet 4.6): {type:"adaptive"}; глубину задаёт effort.
		// budget_tokens НЕ отправляем — на Opus 4.8 это вернёт 400.
		payload["thinking"] = map[string]any{
			"type": "adaptive",
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicBaseURL+"/messages", bytes.NewBuffer(body))
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	httpReq.Header.Set("X-Api-Key", a.apiKey)
	httpReq.Header.Set("Anthropic-Version", anthropicVersion)
	httpReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	l := ports.LoggerFromContext(ctx)
	l.InfoContext(
		ctx, "anthropic request",
		"model", model,
		"thinking", thinking,
		"effort", effort,
		"maxTokens", maxTokens,
		"bodyBytes", len(body),
	)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	if resp.StatusCode != http.StatusOK {
		maxLog := min(len(raw), 400)
		l.ErrorContext(
			ctx, "anthropic api error",
			"model", model,
			"status", resp.StatusCode,
			"body", string(raw[:maxLog]),
		)
		// Detect credit exhaustion → return sentinel so orchestrator can pause
		if isInsufficientFundsError(resp.StatusCode, string(raw[:maxLog])) {
			return nil, ErrInsufficientFunds
		}

		return nil, wrapper.Wrapf(ErrAnthropicAPIError, "(HTTP %d): %s", resp.StatusCode, string(raw[:maxLog]))
	}

	var parsed anthropicResponse
	err = json.Unmarshal(raw, &parsed)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if parsed.Error != nil {
		return nil, wrapper.Wrapf(ErrAnthropicAPIResponseError, "%s (%s)", parsed.Error.Message, parsed.Error.Type)
	}

	// Defensive parse: собираем только финальный текст. Блоки thinking / tool_use /
	// server_tool_use / code_execution_tool_result молча пропускаются (не ломают извлечение).
	var out strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" && block.Text != "" {
			out.WriteString(block.Text)
		}
	}
	if out.Len() == 0 {
		return nil, wrapper.Wrapf(ErrAnthropicEmptyContent, "(stop=%s)", parsed.StopReason)
	}
	l.InfoContext(
		ctx, "anthropic response",
		"model", model,
		"chars", out.Len(),
		"inputTokens", parsed.Usage.InputTokens,
		"outputTokens", parsed.Usage.OutputTokens,
		"stopReason", parsed.StopReason,
		"duration", time.Since(start).Round(time.Millisecond),
	)

	return &ports.LLMResponse{
		Content:    out.String(),
		TokensUsed: parsed.Usage.InputTokens + parsed.Usage.OutputTokens,
		Model:      parsed.Model,
	}, nil
}

// resolveAnthropicModel нормализует идентификатор модели и определяет режим
// Adaptive Thinking.
//
//   - "...opus..."        → claude-opus-4-8 (Architect/Planner)
//   - всё остальное       → claude-sonnet-4-6 (дефолт: Coder/Researcher/Validator)
//   - суффикс "-thinking" → adaptive thinking
//   - reqReasoning=true   → форсит thinking независимо от модели
//
// Legacy-идентификаторы Opus (opus-4-7 и т.п.) поднимаются до Opus 4.8.
func resolveAnthropicModel(raw string, reqReasoning bool) (model string, thinking bool) {
	lower := strings.ToLower(strings.TrimSpace(raw))
	thinking = reqReasoning || strings.Contains(lower, "thinking")

	if strings.Contains(lower, "opus") {
		return ModelClaudeOpus48, thinking
	}

	return ModelClaudeSonnet46, thinking
}

// resolveEffort валидирует уровень effort; пустой/неизвестный → defaultEffort ("high").
func resolveEffort(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case ports.EffortLow, ports.EffortMedium, ports.EffortHigh, ports.EffortXHigh, ports.EffortMax:
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return defaultEffort
	}
}

// effortTimeout масштабирует per-request дедлайн по уровню effort: высокие уровни
// думают дольше, поэтому требуют больший бюджет времени.
func effortTimeout(effort string) time.Duration {
	switch effort {
	case ports.EffortLow, ports.EffortMedium:
		return 5 * time.Minute
	case ports.EffortXHigh:
		return 15 * time.Minute
	case ports.EffortMax:
		return 20 * time.Minute
	default: // high
		return 10 * time.Minute
	}
}

// maxOutputFor возвращает нативный лимит output-токенов синхронного Messages API.
func maxOutputFor(model string) int {
	if strings.Contains(model, "opus") {
		return maxOutputOpus48
	}

	return maxOutputSonnet46
}

// IsAnthropicModel проверяет, нужно ли маршрутизировать модель в Anthropic адаптер.
func IsAnthropicModel(model string) bool {
	lower := strings.ToLower(strings.TrimSpace(model))

	return strings.HasPrefix(lower, "anthropic/") ||
		strings.HasPrefix(lower, "claude-")
}
