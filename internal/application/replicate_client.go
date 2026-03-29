package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Replicate Client
//  Dual routing: Replicate для Anthropic моделей
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

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

// callReplicate вызывает модель через Replicate Predictions API с async polling.
// Используется для Anthropic моделей (Claude Opus 4.6)
func callReplicate(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens int, temperature float64) (string, error) {
	token := os.Getenv("REPLICATE_API_TOKEN")
	if token == "" {
		return "", fmt.Errorf("REPLICATE_API_TOKEN not set")
	}

	// Replicate requires min 1024 max_tokens for Claude
	if maxTokens < 1024 {
		maxTokens = 1024
	}

	input := map[string]interface{}{
		"prompt":      userPrompt,
		"max_tokens":  maxTokens,
		"temperature": temperature,
	}
	if systemPrompt != "" {
		input["system_prompt"] = systemPrompt
	}

	payload := map[string]interface{}{
		"input": input,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal failed: %w", err)
	}

	// Step 1: Create prediction (async)
	endpoint := fmt.Sprintf("%s/models/%s/predictions", replicateBaseURL, model)
	log.Printf("🔗 Replicate: creating prediction for %s (%d bytes)", model, len(body))

	pred, err := replicatePost(ctx, token, endpoint, body)
	if err != nil {
		return "", err
	}

	// If already succeeded (small requests with Prefer: wait fallback)
	if pred.Status == "succeeded" {
		output := extractReplicateOutput(pred.Output)
		log.Printf("✅ Replicate: %s → %d chars (instant)", model, len(output))
		return output, nil
	}

	if pred.Error != nil {
		return "", fmt.Errorf("Replicate prediction error: %v", pred.Error)
	}

	// Step 2: Poll for completion
	pollURL := pred.URLs.Get
	if pollURL == "" {
		pollURL = fmt.Sprintf("%s/predictions/%s", replicateBaseURL, pred.ID)
	}

	log.Printf("⏳ Replicate: polling %s (id=%s)", model, pred.ID)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := time.After(8 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-timeout:
			return "", fmt.Errorf("Replicate prediction timed out after 8min (id=%s)", pred.ID)
		case <-ticker.C:
			poll, err := replicateGet(ctx, token, pollURL)
			if err != nil {
				log.Printf("⚠️ Replicate poll error: %v", err)
				continue
			}

			switch poll.Status {
			case "succeeded":
				output := extractReplicateOutput(poll.Output)
				if output == "" {
					return "", fmt.Errorf("empty output from Replicate (id=%s)", pred.ID)
				}
				log.Printf("✅ Replicate: %s → %d chars (id=%s)", model, len(output), pred.ID)
				return output, nil
			case "failed", "canceled":
				return "", fmt.Errorf("Replicate prediction %s: %v", poll.Status, poll.Error)
			default:
				// "starting", "processing" — keep polling
			}
		}
	}
}

// replicatePost creates a new prediction
func replicatePost(ctx context.Context, token, url string, body []byte) (*replicatePrediction, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
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

// replicateGet polls a prediction status
func replicateGet(ctx context.Context, token, url string) (*replicatePrediction, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("poll HTTP %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 300)]))
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

	// String output
	if s, ok := output.(string); ok {
		return s
	}

	// Array of strings (streaming chunks concatenated)
	if arr, ok := output.([]interface{}); ok {
		var sb strings.Builder
		for _, chunk := range arr {
			if s, ok := chunk.(string); ok {
				sb.WriteString(s)
			}
		}
		return sb.String()
	}

	// Fallback: try JSON marshal
	b, _ := json.Marshal(output)
	return string(b)
}

// isAnthropicModel проверяет, является ли модель Anthropic (роутим на Replicate)
func isAnthropicModel(model string) bool {
	return strings.HasPrefix(model, "anthropic/")
}
