package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/istok/agent-core/internal/ports"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Anthropic Direct Adapter
//  Прямая интеграция с Anthropic Messages API.
//  Claude Opus 4.7 (+adaptive thinking для планирования/архитектуры).
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const (
	anthropicBaseURL = "https://api.anthropic.com/v1"
	anthropicVersion = "2023-06-01"
	anthropicBeta    = "output-128k-2025-02-19"

	// ModelClaudeOpus47 — единственная production модель (verified via Workbench).
	// 3.7 Sonnet снят с обслуживания → 404 на API. Все агенты идут на Opus 4.7.
	ModelClaudeOpus47         = "claude-opus-4-7"
	ModelClaudeOpus47Thinking = "claude-opus-4-7-thinking" // логический alias → adaptive thinking

	// DefaultMaxTokens — лимит как в проверенном Workbench-запросе.
	// 20000 даёт thinking-режиму запас на reasoning + финальный output.
	DefaultMaxTokens = 20000
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
// Поддерживает Extended Thinking mode: активируется если req.Reasoning=true
// либо модель содержит суффикс "-thinking".
func (a *AnthropicAdapter) Complete(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	if a.apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not configured")
	}

	model, thinking := resolveAnthropicModel(req.Model, req.Reasoning)

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens // 20000 — verified в Workbench
	}

	// ⚠️ Opus 4.7: temperature is DEPRECATED → API возвращает 400 если поле передано.
	// Не включаем в payload ни в thinking, ни в standard режиме.
	payload := map[string]interface{}{
		"model":      model,
		"max_tokens": maxTokens,
	}
	_ = req.Temperature // explicitly ignored for Opus 4.7

	if req.SystemPrompt != "" {
		payload["system"] = req.SystemPrompt
	}

	payload["messages"] = []map[string]interface{}{
		{
			"role":    "user",
			"content": req.UserPrompt,
		},
	}

	if thinking {
		// Adaptive thinking — модель сама управляет reasoning budget.
		// budget_tokens не передаём: adaptive mode не принимает его.
		payload["thinking"] = map[string]interface{}{
			"type": "adaptive",
		}
	}

	// output_config.effort — управляет глубиной reasoning (low/medium/high).
	// Default "medium" экономит токены; Coder/Architect передают "high".
	effort := req.Effort
	if effort == "" {
		effort = "medium"
	}
	payload["output_config"] = map[string]interface{}{
		"effort": effort,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("anthropic marshal failed: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", anthropicBaseURL+"/messages", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("anthropic request build failed: %w", err)
	}
	httpReq.Header.Set("x-api-key", a.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)
	httpReq.Header.Set("anthropic-beta", anthropicBeta)
	httpReq.Header.Set("content-type", "application/json")

	start := time.Now()
	log.Printf("🔗 Anthropic: %s (thinking=%v, max_tokens=%d, %d bytes)",
		model, thinking, maxTokens, len(body))

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("anthropic read failed: %w", err)
	}

	if resp.StatusCode != 200 {
		maxLog := len(raw)
		if maxLog > 400 {
			maxLog = 400
		}
		log.Printf("🚨 Anthropic error | model=%s status=%d | %s",
			model, resp.StatusCode, string(raw[:maxLog]))
		return nil, fmt.Errorf("anthropic API error (HTTP %d): %s",
			resp.StatusCode, string(raw[:maxLog]))
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("anthropic parse failed: %w", err)
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("anthropic API: %s (%s)",
			parsed.Error.Message, parsed.Error.Type)
	}

	var out strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" && block.Text != "" {
			out.WriteString(block.Text)
		}
	}
	if out.Len() == 0 {
		return nil, fmt.Errorf("anthropic returned empty content (stop=%s)", parsed.StopReason)
	}

	log.Printf("✅ Anthropic: %s → %d chars, tokens=%d/%d, stop=%s (%v)",
		model, out.Len(),
		parsed.Usage.InputTokens, parsed.Usage.OutputTokens,
		parsed.StopReason,
		time.Since(start).Round(time.Millisecond))

	return &ports.LLMResponse{
		Content:    out.String(),
		TokensUsed: parsed.Usage.InputTokens + parsed.Usage.OutputTokens,
		Model:      parsed.Model,
	}, nil
}

// resolveAnthropicModel нормализует идентификатор модели и определяет режим
// adaptive thinking. ВСЕ варианты identifier'а резолвятся в claude-opus-4-7,
// потому что 3.7 Sonnet снят с обслуживания (404). Thinking-флаг сохраняется
// для совместимости с конфигами агентов (orchestrator уже размечает роли).
//
// Принимает форматы (любой → claude-opus-4-7):
//   - "anthropic/claude-opus-4-7"
//   - "anthropic/claude-opus-4-7-thinking" → thinking enabled
//   - "anthropic/claude-3-7-sonnet[-thinking]" → legacy alias, маппится на opus-4-7
//   - reqReasoning=true → форсит thinking независимо от модели
func resolveAnthropicModel(raw string, reqReasoning bool) (model string, thinking bool) {
	lower := strings.ToLower(strings.TrimSpace(raw))
	if strings.Contains(lower, "thinking") {
		thinking = true
	}
	if reqReasoning {
		thinking = true
	}
	return ModelClaudeOpus47, thinking
}

// IsAnthropicModel проверяет, нужно ли маршрутизировать модель в Anthropic адаптер.
func IsAnthropicModel(model string) bool {
	lower := strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(lower, "anthropic/") ||
		strings.HasPrefix(lower, "claude-")
}
