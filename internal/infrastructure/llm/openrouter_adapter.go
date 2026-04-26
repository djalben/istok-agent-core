package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/istok/agent-core/internal/ports"
)

// OpenRouterAdapter реализует ports.LLMProvider через OpenRouter API.
// Используется для DeepSeek, Qwen и других моделей, доступных через OpenRouter.
type OpenRouterAdapter struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewOpenRouterAdapter создаёт адаптер для OpenRouter API.
// baseURL — URL прокси (Cloudflare) или https://openrouter.ai/api/v1.
func NewOpenRouterAdapter(apiKey, baseURL string) *OpenRouterAdapter {
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	return &OpenRouterAdapter{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// Complete реализует ports.LLMProvider — отправляет запрос через OpenRouter chat/completions.
func (a *OpenRouterAdapter) Complete(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	if a.apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY not configured")
	}

	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	var messages []msg
	if req.SystemPrompt != "" {
		messages = append(messages, msg{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, msg{Role: "user", Content: req.UserPrompt})

	temp := req.Temperature
	if temp == 0 {
		temp = 0.7
	}

	payload := map[string]interface{}{
		"model":       req.Model,
		"messages":    messages,
		"max_tokens":  req.MaxTokens,
		"temperature": temp,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/chat/completions", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("HTTP-Referer", "https://istok-agent-core.vercel.app")
	httpReq.Header.Set("X-Title", "Istok Agent Core")
	httpReq.Header.Set("User-Agent", "IstokAgent/2.0")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != 200 {
		maxLog := len(body)
		if maxLog > 400 {
			maxLog = 400
		}
		log.Printf("🚨 LLM error | model=%s status=%d | %s", req.Model, resp.StatusCode, string(body[:maxLog]))
		return nil, fmt.Errorf("LLM API error (HTTP %d): %s", resp.StatusCode, string(body[:maxLog]))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("empty response from model")
	}

	return &ports.LLMResponse{
		Content:    parsed.Choices[0].Message.Content,
		TokensUsed: parsed.Usage.TotalTokens,
		Model:      req.Model,
	}, nil
}
