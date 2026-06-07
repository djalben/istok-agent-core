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
//  Claude Sonnet 4.6 (+Adaptive Thinking API для планирования/архитектуры).
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const (
	anthropicBaseURL = "https://api.anthropic.com/v1"
	anthropicVersion = "2023-06-01"
	anthropicBeta    = "output-128k-2025-02-19"

	// ModelClaudeSonnet46 — production модель Sonnet 4.6 (быстрее Opus 4.7, Adaptive Thinking API).
	ModelClaudeSonnet46         = "claude-sonnet-4-6"
	ModelClaudeSonnet46Thinking = "claude-sonnet-4-6-thinking" // логический alias → adaptive thinking

	// DefaultMaxTokens — Sonnet 4.6 поддерживает до 128k output tokens.
	DefaultMaxTokens = 128000
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
		httpClient: &http.Client{
			Timeout: 8 * time.Minute,
		},
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

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens // 128000 — Sonnet 4.6 max output
	}

	// Sonnet 4.6: temperature не передаём в thinking-режиме.
	payload := map[string]any{
		"model":      model,
		"max_tokens": maxTokens,
	}
	_ = req.Temperature // explicitly ignored in thinking mode

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
		// Adaptive Thinking API (Sonnet 4.6) — strictly {type: "adaptive"}.
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
	httpReq.Header.Set("Anthropic-Beta", anthropicBeta)
	httpReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	l := ports.LoggerFromContext(ctx)
	l.InfoContext(
		ctx, "anthropic request",
		"model", model,
		"thinking", thinking,
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
// Adaptive Thinking. ВСЕ варианты identifier'а резолвятся в claude-sonnet-4-6.
// Thinking-флаг сохраняется для совместимости с конфигами агентов.
//
// Принимает форматы (любой → claude-sonnet-4-6):
//   - "anthropic/claude-sonnet-4-6"
//   - "anthropic/claude-sonnet-4-6-thinking" → thinking enabled
//   - "anthropic/claude-opus-4-7[-thinking]" → legacy, маппится на sonnet-4-6
//   - reqReasoning=true → форсит thinking независимо от модели
func resolveAnthropicModel(raw string, reqReasoning bool) (model string, thinking bool) {
	lower := strings.ToLower(strings.TrimSpace(raw))
	if strings.Contains(lower, "thinking") {
		thinking = true
	}
	if reqReasoning {
		thinking = true
	}

	return ModelClaudeSonnet46, thinking
}

// IsAnthropicModel проверяет, нужно ли маршрутизировать модель в Anthropic адаптер.
func IsAnthropicModel(model string) bool {
	lower := strings.ToLower(strings.TrimSpace(model))

	return strings.HasPrefix(lower, "anthropic/") ||
		strings.HasPrefix(lower, "claude-")
}
