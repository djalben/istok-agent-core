package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/djalben/istok-agent-core/internal/ports"
	"gitlab.com/libs-artifex/wrapper"
)

const replicateBaseURL = "https://api.replicate.com/v1"

// replicatePrediction — структура ответа Replicate API.
type replicatePrediction struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Output any    `json:"output"`
	Error  any    `json:"error"`
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
		return nil, ErrReplicateTokenNotSet
	}

	maxTokens := max(req.MaxTokens, 1024)

	temp := req.Temperature
	if temp == 0 {
		temp = 0.7
	}
	if req.Reasoning {
		temp = 1.0
	}

	input := map[string]any{
		"prompt":      req.UserPrompt,
		"max_tokens":  maxTokens,
		"temperature": temp,
	}
	if req.SystemPrompt != "" {
		input["system_prompt"] = req.SystemPrompt
	}

	payload := map[string]any{
		"input": input,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}

	endpoint := fmt.Sprintf("%s/models/%s/predictions", replicateBaseURL, req.Model)
	l := ports.LoggerFromContext(ctx)
	l.InfoContext(ctx, "replicate prediction create", "model", req.Model, "bodyBytes", len(body))

	pred, err := a.post(ctx, endpoint, body)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	if pred.Status == "succeeded" {
		output := extractReplicateOutput(pred.Output)
		l.InfoContext(ctx, "replicate instant success", "model", req.Model, "chars", len(output))

		return &ports.LLMResponse{Content: output, Model: req.Model}, nil
	}

	if pred.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrReplicatePredictionError, pred.Error)
	}

	// Poll for completion
	pollURL := pred.URLs.Get
	if pollURL == "" {
		pollURL = fmt.Sprintf("%s/predictions/%s", replicateBaseURL, pred.ID)
	}
	l.InfoContext(ctx, "replicate polling", "model", req.Model, "predictionId", pred.ID)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := time.After(8 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return nil, wrapper.Wrap(ctx.Err())
		case <-timeout:
			return nil, fmt.Errorf("%w after 8min (id=%s)", ErrReplicatePredictionTimeout, pred.ID)
		case <-ticker.C:
			poll, err := a.get(ctx, pollURL)
			if err != nil {
				l.WarnContext(ctx, "replicate poll error", "error", err)

				continue
			}

			switch poll.Status {
			case "succeeded":
				output := extractReplicateOutput(poll.Output)
				if output == "" {
					return nil, fmt.Errorf("%w (id=%s)", ErrReplicateEmptyOutput, pred.ID)
				}
				l.InfoContext(ctx, "replicate success", "model", req.Model, "chars", len(output), "predictionId", pred.ID)

				return &ports.LLMResponse{Content: output, Model: req.Model}, nil
			case "failed", "canceled":
				return nil, fmt.Errorf("%w %s: %v", ErrReplicatePredictionFailed, poll.Status, poll.Error)
			default:
				// "starting", "processing" — keep polling
			}
		}
	}
}

// post creates a new prediction.
func (a *ReplicateAdapter) post(ctx context.Context, url string, body []byte) (*replicatePrediction, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		maxLog := min(len(respBody), 500)
		if isInsufficientFundsError(resp.StatusCode, string(respBody[:maxLog])) {
			return nil, ErrInsufficientFunds
		}

		return nil, fmt.Errorf("%w (HTTP %d): %s", ErrReplicateAPIError, resp.StatusCode, string(respBody[:maxLog]))
	}

	var pred replicatePrediction
	err = json.Unmarshal(respBody, &pred)
	if err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &pred, nil
}

// get polls a prediction status.
func (a *ReplicateAdapter) get(ctx context.Context, url string) (*replicatePrediction, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		maxLog := min(len(respBody), 300)

		return nil, fmt.Errorf("%w %d: %s", ErrReplicatePollHTTPError, resp.StatusCode, string(respBody[:maxLog]))
	}

	var pred replicatePrediction
	err = json.Unmarshal(respBody, &pred)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &pred, nil
}

// extractReplicateOutput handles different output formats from Replicate.
func extractReplicateOutput(output any) string {
	if output == nil {
		return ""
	}

	if s, ok := output.(string); ok {
		return s
	}

	if arr, ok := output.([]any); ok {
		var sb strings.Builder
		for _, chunk := range arr {
			if s, ok := chunk.(string); ok {
				sb.WriteString(s)
			}
		}

		return sb.String()
	}

	b, err := json.Marshal(output)
	if err != nil {
		return fmt.Sprintf("%v", output)
	}

	return string(b)
}
