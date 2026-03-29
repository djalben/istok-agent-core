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

// callReplicate вызывает модель через Replicate Predictions API
// Используется для Anthropic моделей (Claude Opus 4.6)
func callReplicate(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens int, temperature float64) (string, error) {
	token := os.Getenv("REPLICATE_API_TOKEN")
	if token == "" {
		return "", fmt.Errorf("REPLICATE_API_TOKEN not set")
	}

	// Replicate model format: "anthropic/claude-opus-4.6" → same as OpenRouter
	endpoint := fmt.Sprintf("%s/models/%s/predictions", replicateBaseURL, model)

	// Build prompt: combine system + user for Replicate's format
	input := map[string]interface{}{
		"prompt":       userPrompt,
		"max_tokens":   maxTokens,
		"temperature":  temperature,
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

	log.Printf("🔗 Replicate: POST %s (%d bytes)", endpoint, len(body))

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "wait") // Synchronous — wait for result

	client := &http.Client{Timeout: 8 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		maxLog := len(respBody)
		if maxLog > 500 {
			maxLog = 500
		}
		log.Printf("🚨 Replicate error | model=%s status=%d | %s", model, resp.StatusCode, string(respBody[:maxLog]))
		return "", fmt.Errorf("Replicate API error (HTTP %d): %s", resp.StatusCode, string(respBody[:maxLog]))
	}

	// Parse Replicate response
	var result struct {
		ID     string      `json:"id"`
		Status string      `json:"status"`
		Output interface{} `json:"output"`
		Error  interface{} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response failed: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("Replicate prediction error: %v", result.Error)
	}

	if result.Status != "succeeded" {
		return "", fmt.Errorf("Replicate prediction status: %s (expected succeeded)", result.Status)
	}

	// Output can be string or []string
	output := extractReplicateOutput(result.Output)
	if output == "" {
		return "", fmt.Errorf("empty output from Replicate")
	}

	log.Printf("✅ Replicate: %s → %d chars, status=%s", model, len(output), result.Status)
	return output, nil
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
