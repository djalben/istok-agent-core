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

const replicateBaseURL = "https://api.replicate.com/v1"

// replicatePrediction — структура ответа Replicate API
type replicatePrediction struct {
	ID     string      `json:"id"`
	Status string      `json:"status"`
	Output interface{} `json:"output"`
	Error  interface{} `json:"error"`
	URLs   struct {
		Get string `json:"get"`
	} `json:"urls"`
}

// ReplicateAdapter реализует ports.LLMProvider через Replicate Predictions API.
// Используется для Google/Anthropic моделей (Gemini 3 Pro, Claude Opus).
type ReplicateAdapter struct {
	token      string
	httpClient *http.Client
}

// NewReplicateAdapter создаёт адаптер для Replicate API.
func NewReplicateAdapter(token string) *ReplicateAdapter {
	return &ReplicateAdapter{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Complete реализует ports.LLMProvider — отправляет запрос через Replicate Predictions API
// с асинхронным поллингом результата.
func (a *ReplicateAdapter) Complete(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	if a.token == "" {
		return nil, fmt.Errorf("REPLICATE_API_TOKEN not set")
	}

	maxTokens := req.MaxTokens
	if maxTokens < 1024 {
		maxTokens = 1024
	}

	temp := req.Temperature
	if temp == 0 {
		temp = 0.7
	}
	if req.Reasoning {
		temp = 1.0
	}

	input := map[string]interface{}{
		"prompt":      req.UserPrompt,
		"max_tokens":  maxTokens,
		"temperature": temp,
	}
	if req.SystemPrompt != "" {
		input["system_prompt"] = req.SystemPrompt
	}

	payload := map[string]interface{}{
		"input": input,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}

	endpoint := fmt.Sprintf("%s/models/%s/predictions", replicateBaseURL, req.Model)
	log.Printf("🔗 Replicate: creating prediction for %s (%d bytes)", req.Model, len(body))

	pred, err := a.post(ctx, endpoint, body)
	if err != nil {
		return nil, err
	}

	if pred.Status == "succeeded" {
		output := extractReplicateOutput(pred.Output)
		log.Printf("✅ Replicate: %s → %d chars (instant)", req.Model, len(output))
		return &ports.LLMResponse{Content: output, Model: req.Model}, nil
	}

	if pred.Error != nil {
		return nil, fmt.Errorf("Replicate prediction error: %v", pred.Error)
	}

	// Poll for completion
	pollURL := pred.URLs.Get
	if pollURL == "" {
		pollURL = fmt.Sprintf("%s/predictions/%s", replicateBaseURL, pred.ID)
	}

	log.Printf("⏳ Replicate: polling %s (id=%s)", req.Model, pred.ID)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := time.After(8 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("Replicate prediction timed out after 8min (id=%s)", pred.ID)
		case <-ticker.C:
			poll, err := a.get(ctx, pollURL)
			if err != nil {
				log.Printf("⚠️ Replicate poll error: %v", err)
				continue
			}

			switch poll.Status {
			case "succeeded":
				output := extractReplicateOutput(poll.Output)
				if output == "" {
					return nil, fmt.Errorf("empty output from Replicate (id=%s)", pred.ID)
				}
				log.Printf("✅ Replicate: %s → %d chars (id=%s)", req.Model, len(output), pred.ID)
				return &ports.LLMResponse{Content: output, Model: req.Model}, nil
			case "failed", "canceled":
				return nil, fmt.Errorf("Replicate prediction %s: %v", poll.Status, poll.Error)
			default:
				// "starting", "processing" — keep polling
			}
		}
	}
}

// post creates a new prediction
func (a *ReplicateAdapter) post(ctx context.Context, url string, body []byte) (*replicatePrediction, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		maxLog := len(respBody)
		if maxLog > 500 {
			maxLog = 500
		}
		return nil, fmt.Errorf("Replicate API error (HTTP %d): %s", resp.StatusCode, string(respBody[:maxLog]))
	}

	var pred replicatePrediction
	if err := json.Unmarshal(respBody, &pred); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}
	return &pred, nil
}

// get polls a prediction status
func (a *ReplicateAdapter) get(ctx context.Context, url string) (*replicatePrediction, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		maxLog := min(len(respBody), 300)
		return nil, fmt.Errorf("poll HTTP %d: %s", resp.StatusCode, string(respBody[:maxLog]))
	}

	var pred replicatePrediction
	if err := json.Unmarshal(respBody, &pred); err != nil {
		return nil, err
	}
	return &pred, nil
}

// extractReplicateOutput handles different output formats from Replicate
func extractReplicateOutput(output interface{}) string {
	if output == nil {
		return ""
	}

	if s, ok := output.(string); ok {
		return s
	}

	if arr, ok := output.([]interface{}); ok {
		var sb strings.Builder
		for _, chunk := range arr {
			if s, ok := chunk.(string); ok {
				sb.WriteString(s)
			}
		}
		return sb.String()
	}

	b, _ := json.Marshal(output)
	return string(b)
}
