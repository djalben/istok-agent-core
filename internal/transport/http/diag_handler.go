package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Diagnostic Handler
//  Тестирует Anthropic Direct + Replicate (nano-banana, Veo 3).
//  OpenRouter удалён.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// DiagHandler диагностика моделей.
type DiagHandler struct{}

func NewDiagHandler() *DiagHandler { return &DiagHandler{} }

type diagResult struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Status   int    `json:"http_status"`
	OK       bool   `json:"ok"`
	Response string `json:"response,omitempty"`
	Error    string `json:"error,omitempty"`
	Duration string `json:"duration"`
}

// Handle GET /api/v1/diag/models — тестирует Anthropic + Replicate.
func (h *DiagHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET only")

		return
	}

	ctx := r.Context()
	results := make([]diagResult, 0, 3)

	// ── Anthropic Direct — Claude Sonnet 4.6 (standard) ──
	results = append(results, anthropicProbe(ctx, "claude-sonnet-4-6", false))

	// ── Anthropic Direct — Claude Sonnet 4.6 (adaptive thinking) ──
	results = append(results, anthropicProbe(ctx, "claude-sonnet-4-6-thinking", true))

	// ── Replicate — nano-banana availability ──
	results = append(results, replicateModelProbe(ctx, "google/nano-banana", "image"))

	// ── Replicate — Veo 3 availability ──
	results = append(results, replicateModelProbe(ctx, "google/veo-3", "video"))

	_ = writeJSON(w, http.StatusOK, map[string]any{
		"results": results,
	})
}

// anthropicProbe — минимальный пинг Anthropic Messages API.
func anthropicProbe(parentCtx context.Context, name string, thinking bool) diagResult {
	start := time.Now()
	tr := diagResult{Name: name, Provider: "Anthropic Direct"}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		tr.Error = "ANTHROPIC_API_KEY not set"
		tr.Duration = time.Since(start).String()

		return tr
	}

	payload := map[string]any{
		"model":      "claude-sonnet-4-6",
		"max_tokens": 1024,
		"messages": []map[string]any{
			{"role": "user", "content": "Reply with ONLY the JSON: {\"ok\":true}"},
		},
	}
	if thinking {
		payload["max_tokens"] = 16384
		payload["thinking"] = map[string]any{
			"type": "adaptive",
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		tr.Error = err.Error()
		tr.Duration = time.Since(start).String()

		return tr
	}
	ctx, cancel := context.WithTimeout(parentCtx, 2*time.Minute)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewBuffer(body))
	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("Anthropic-Version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 2 * time.Minute}).Do(req)
	tr.Duration = time.Since(start).String()
	if err != nil {
		tr.Error = err.Error()

		return tr
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	tr.Status = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		tr.Error = truncate(string(respBody), 500)
		logFrom(parentCtx).WarnContext(parentCtx, "diag anthropic probe failed",
			"name", name,
			"status", resp.StatusCode,
		)

		return tr
	}

	var parsed struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	_ = json.Unmarshal(respBody, &parsed)
	for _, c := range parsed.Content {
		if c.Type == "text" {
			tr.Response = truncate(c.Text, 200)

			break
		}
	}
	tr.OK = true
	logFrom(parentCtx).InfoContext(parentCtx, "diag anthropic probe ok",
		"name", name,
		"status", resp.StatusCode,
		"duration", tr.Duration,
	)

	return tr
}

// replicateModelProbe — проверяет что модель существует на Replicate (GET /v1/models/{slug}).
func replicateModelProbe(parentCtx context.Context, slug, kind string) diagResult {
	start := time.Now()
	tr := diagResult{Name: "replicate-" + slug, Provider: "Replicate (" + kind + ")"}

	token := os.Getenv("REPLICATE_API_TOKEN")
	if token == "" {
		tr.Error = "REPLICATE_API_TOKEN not set"
		tr.Duration = time.Since(start).String()

		return tr
	}

	ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.replicate.com/v1/models/"+slug, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	tr.Duration = time.Since(start).String()
	if err != nil {
		tr.Error = err.Error()

		return tr
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	tr.Status = resp.StatusCode

	if resp.StatusCode != http.StatusOK {
		tr.Error = truncate(string(respBody), 500)
		logFrom(parentCtx).WarnContext(parentCtx, "diag replicate probe failed",
			"name", tr.Name,
			"status", resp.StatusCode,
		)

		return tr
	}

	var model struct {
		Name        string `json:"name"`
		Owner       string `json:"owner"`
		Description string `json:"description"`
	}
	_ = json.Unmarshal(respBody, &model)
	tr.OK = true
	tr.Response = truncate(fmt.Sprintf("%s/%s: %s", model.Owner, model.Name, model.Description), 200)
	logFrom(parentCtx).InfoContext(parentCtx, "diag replicate probe ok",
		"name", tr.Name,
		"status", resp.StatusCode,
		"duration", tr.Duration,
	)

	return tr
}

// HandleEnv GET /api/v1/diag/env — показывает routing без секретов.
func (h *DiagHandler) HandleEnv(w http.ResponseWriter, _ *http.Request) {
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	replicateKey := os.Getenv("REPLICATE_API_TOKEN")

	out := map[string]any{
		"providers": map[string]any{
			"anthropic": map[string]any{
				"endpoint": "https://api.anthropic.com/v1/messages",
				"has_key":  anthropicKey != "",
				"key_hint": "sk-ant-..." + lastN(anthropicKey, 4),
				"models": []string{
					"claude-sonnet-4-6",
					"claude-sonnet-4-6 (adaptive thinking)",
				},
			},
			"replicate": map[string]any{
				"endpoint": "https://api.replicate.com/v1",
				"has_key":  replicateKey != "",
				"key_hint": "r8_..." + lastN(replicateKey, 4),
				"models": []string{
					"google/nano-banana (image)",
					"google/veo-3 (video)",
				},
			},
		},
		"routing": "Anthropic: Director/Brain/Researcher/Coder/Validator/Planner. Replicate: Designer (nano-banana), Videographer (veo-3).",
	}

	_ = writeJSON(w, http.StatusOK, out)
}

func lastN(s string, n int) string {
	if len(s) <= n {
		return "****"
	}

	return s[len(s)-n:]
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}

	return s[:n]
}
