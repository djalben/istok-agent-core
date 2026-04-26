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
//  Claude 3.7 Sonnet (+thinking для планирования).
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const (
	anthropicBaseURL = "https://api.anthropic.com/v1"
	anthropicVersion = "2023-06-01"
	anthropicBeta    = "output-128k-2025-02-19"

	// ModelClaude37Sonnet — базовая модель для coding/validation/analysis.
	ModelClaude37Sonnet = "claude-3-7-sonnet-20250219"
	// ModelClaude37SonnetThinking — с extended thinking для planning/architecture.
	ModelClaude37SonnetThinking = "claude-3-7-sonnet-20250219-thinking"
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
		maxTokens = 4096
	}
	if thinking && maxTokens < 8192 {
		// Thinking needs headroom for internal reasoning tokens + final output.
		maxTokens = 16384
	}

	temperature := req.Temperature
	if thinking {
		// Thinking mode requires temperature=1.0 per Anthropic docs.
		temperature = 1.0
	} else if temperature == 0 {
		temperature = 0.7
	}

	payload := map[string]interface{}{
		"model":       model,
		"max_tokens":  maxTokens,
		"temperature": temperature,
	}

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
		budget := req.ThinkingBudget
		if budget <= 0 {
			budget = 8000
		}
		if budget > maxTokens-1024 {
			budget = maxTokens - 1024
		}
		payload["thinking"] = map[string]interface{}{
			"type":          "enabled",
			"budget_tokens": budget,
		}
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

	log.Printf("✅ Anthropic: %s → %d chars, tokens=%d/%d (%v)",
		model, out.Len(),
		parsed.Usage.InputTokens, parsed.Usage.OutputTokens,
		time.Since(start).Round(time.Millisecond))

	return &ports.LLMResponse{
		Content:    out.String(),
		TokensUsed: parsed.Usage.InputTokens + parsed.Usage.OutputTokens,
		Model:      parsed.Model,
	}, nil
}

// resolveAnthropicModel нормализует идентификатор модели и определяет режим
// extended thinking. Принимает форматы:
//   - "anthropic/claude-3-7-sonnet" → claude-3-7-sonnet-20250219
//   - "anthropic/claude-3-7-sonnet-thinking" → thinking enabled
//   - "claude-3-7-sonnet-20250219" → as-is
//   - reqReasoning=true → форсит thinking независимо от модели
func resolveAnthropicModel(raw string, reqReasoning bool) (model string, thinking bool) {
	id := strings.TrimSpace(raw)
	id = strings.TrimPrefix(id, "anthropic/")
	id = strings.TrimPrefix(id, "claude-3.7-sonnet")
	id = strings.TrimPrefix(id, "claude-3-7-sonnet")

	// Whatever remains may include "-thinking" suffix or a version id.
	if strings.Contains(strings.ToLower(id), "thinking") || strings.Contains(strings.ToLower(raw), "thinking") {
		thinking = true
	}
	if reqReasoning {
		thinking = true
	}

	return ModelClaude37Sonnet, thinking
}

// IsAnthropicModel проверяет, нужно ли маршрутизировать модель в Anthropic адаптер.
func IsAnthropicModel(model string) bool {
	lower := strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(lower, "anthropic/") ||
		strings.HasPrefix(lower, "claude-") ||
		strings.Contains(lower, "claude-3-7-sonnet")
}
