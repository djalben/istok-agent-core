package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Diagnostic Handler
//  Тестирует Anthropic Direct + Replicate (nano-banana, Veo 3).
//  OpenRouter удалён.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// DiagHandler диагностика моделей
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

	results := make([]diagResult, 0, 3)

	// ── Anthropic Direct — Claude 3.7 Sonnet (medium) ──
	results = append(results, anthropicProbe("claude-3-7-sonnet-medium", false))

	// ── Anthropic Direct — Claude 3.7 Sonnet (thinking) ──
	results = append(results, anthropicProbe("claude-3-7-sonnet-thinking", true))

	// ── Replicate — nano-banana availability ──
	results = append(results, replicateModelProbe("google/nano-banana", "image"))

	// ── Replicate — Veo 3 availability ──
	results = append(results, replicateModelProbe("google/veo-3", "video"))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
	})
}

// anthropicProbe — минимальный пинг Anthropic Messages API.
func anthropicProbe(name string, thinking bool) diagResult {
	start := time.Now()
	tr := diagResult{Name: name, Provider: "Anthropic Direct"}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		tr.Error = "ANTHROPIC_API_KEY not set"
		tr.Duration = time.Since(start).String()
		return tr
	}

	payload := map[string]interface{}{
		"model":       "claude-3-7-sonnet-20250219",
		"max_tokens":  256,
		"temperature": 0.7,
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Reply with ONLY the JSON: {\"ok\":true}"},
		},
	}
	if thinking {
		payload["temperature"] = 1.0
		payload["max_tokens"] = 8192
		payload["thinking"] = map[string]interface{}{
			"type":          "enabled",
			"budget_tokens": 4096,
		}
	}

	body, _ := json.Marshal(payload)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(body))
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := (&http.Client{Timeout: 2 * time.Minute}).Do(req)
	tr.Duration = time.Since(start).String()
	if err != nil {
		tr.Error = err.Error()
		return tr
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	tr.Status = resp.StatusCode
	if resp.StatusCode != 200 {
		tr.Error = truncate(string(respBody), 500)
		log.Printf("🔍 DIAG %s → HTTP %d", name, resp.StatusCode)
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
	log.Printf("🔍 DIAG %s → HTTP %d (%s)", name, resp.StatusCode, tr.Duration)
	return tr
}

// replicateModelProbe — проверяет что модель существует на Replicate (GET /v1/models/{slug}).
func replicateModelProbe(slug, kind string) diagResult {
	start := time.Now()
	tr := diagResult{Name: fmt.Sprintf("replicate-%s", slug), Provider: "Replicate (" + kind + ")"}

	token := os.Getenv("REPLICATE_API_TOKEN")
	if token == "" {
		tr.Error = "REPLICATE_API_TOKEN not set"
		tr.Duration = time.Since(start).String()
		return tr
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.replicate.com/v1/models/"+slug, nil)
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

	if resp.StatusCode != 200 {
		tr.Error = truncate(string(respBody), 500)
		log.Printf("🔍 DIAG %s → HTTP %d", tr.Name, resp.StatusCode)
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
	log.Printf("🔍 DIAG %s → HTTP %d (%s)", tr.Name, resp.StatusCode, tr.Duration)
	return tr
}

// HandleEnv GET /api/v1/diag/env — показывает routing без секретов.
func (h *DiagHandler) HandleEnv(w http.ResponseWriter, r *http.Request) {
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	replicateKey := os.Getenv("REPLICATE_API_TOKEN")

	out := map[string]interface{}{
		"providers": map[string]interface{}{
			"anthropic": map[string]interface{}{
				"endpoint": "https://api.anthropic.com/v1/messages",
				"has_key":  anthropicKey != "",
				"key_hint": fmt.Sprintf("sk-ant-...%s", lastN(anthropicKey, 4)),
				"models": []string{
					"claude-3-7-sonnet-20250219 (medium)",
					"claude-3-7-sonnet-20250219 (thinking)",
				},
			},
			"replicate": map[string]interface{}{
				"endpoint": "https://api.replicate.com/v1",
				"has_key":  replicateKey != "",
				"key_hint": fmt.Sprintf("r8_...%s", lastN(replicateKey, 4)),
				"models": []string{
					"google/nano-banana (image)",
					"google/veo-3 (video)",
				},
			},
		},
		"routing": "Anthropic: Director/Brain/Researcher/Coder/Validator/Planner. Replicate: Designer (nano-banana), Videographer (veo-3).",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
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
