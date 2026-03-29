package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// DiagHandler диагностика моделей
type DiagHandler struct{}

func NewDiagHandler() *DiagHandler { return &DiagHandler{} }

// Handle GET /api/v1/diag/models — тестирует доступность моделей через прокси
func (h *DiagHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	proxyURL := os.Getenv("OPENROUTER_PROXY_URL")
	if proxyURL == "" {
		proxyURL = "https://openrouter.ai/api/v1"
	}
	endpoint := proxyURL + "/chat/completions"

	type testCase struct {
		Name    string
		Payload map[string]interface{}
	}

	tests := []testCase{
		{
			Name: "claude-opus-4.6-simple",
			Payload: map[string]interface{}{
				"model": "anthropic/claude-opus-4.6",
				"messages": []map[string]string{
					{"role": "user", "content": "Return ONLY: {\"ok\":true}"},
				},
				"max_tokens":  64,
				"temperature": 0.7,
			},
		},
		{
			Name: "claude-opus-4.6-reasoning",
			Payload: map[string]interface{}{
				"model": "anthropic/claude-opus-4.6",
				"messages": []map[string]string{
					{"role": "user", "content": "Return ONLY: {\"ok\":true}"},
				},
				"max_tokens":  64,
				"temperature": 1,
				"reasoning": map[string]interface{}{
					"effort": "high",
				},
				"thinking": map[string]interface{}{
					"type":          "enabled",
					"budget_tokens": 1024,
				},
			},
		},
		{
			Name: "deepseek-v3.2-speciale",
			Payload: map[string]interface{}{
				"model": "deepseek/deepseek-v3.2-speciale",
				"messages": []map[string]string{
					{"role": "user", "content": "Return ONLY: {\"ok\":true}"},
				},
				"max_tokens":  64,
				"temperature": 0.3,
			},
		},
	}

	type testResult struct {
		Name     string `json:"name"`
		Status   int    `json:"http_status"`
		OK       bool   `json:"ok"`
		Response string `json:"response,omitempty"`
		Error    string `json:"error,omitempty"`
		Duration string `json:"duration"`
	}

	results := make([]testResult, 0, len(tests))

	for _, tc := range tests {
		start := time.Now()
		body, _ := json.Marshal(tc.Payload)

		req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("HTTP-Referer", "https://istok-agent-core.vercel.app")
		req.Header.Set("X-Title", "Istok Agent Core")
		req.Header.Set("User-Agent", "IstokAgent/2.0")

		client := &http.Client{Timeout: 2 * time.Minute}
		resp, err := client.Do(req)
		dur := time.Since(start)

		tr := testResult{Name: tc.Name, Duration: dur.String()}

		if err != nil {
			tr.Error = err.Error()
			results = append(results, tr)
			continue
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		tr.Status = resp.StatusCode

		if resp.StatusCode != 200 {
			tr.Error = string(respBody)
			if len(tr.Error) > 500 {
				tr.Error = tr.Error[:500]
			}
		} else {
			tr.OK = true
			var parsed struct {
				Choices []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				} `json:"choices"`
			}
			json.Unmarshal(respBody, &parsed)
			if len(parsed.Choices) > 0 {
				tr.Response = parsed.Choices[0].Message.Content
				if len(tr.Response) > 200 {
					tr.Response = tr.Response[:200]
				}
			}
		}

		log.Printf("🔍 DIAG %s → HTTP %d (%s)", tc.Name, tr.Status, dur)
		results = append(results, tr)
	}

	// ── Replicate test for Claude Opus ──
	replicateToken := os.Getenv("REPLICATE_API_TOKEN")
	if replicateToken != "" {
		start := time.Now()
		replicateEndpoint := "https://api.replicate.com/v1/models/anthropic/claude-opus-4.6/predictions"
		replicatePayload, _ := json.Marshal(map[string]interface{}{
			"input": map[string]interface{}{
				"prompt":      "Return ONLY: {\"ok\":true}",
				"max_tokens":  1024,
				"temperature": 0.7,
			},
		})

		req, _ := http.NewRequest("POST", replicateEndpoint, bytes.NewBuffer(replicatePayload))
		req.Header.Set("Authorization", "Bearer "+replicateToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Prefer", "wait")

		client := &http.Client{Timeout: 2 * time.Minute}
		resp, err := client.Do(req)
		dur := time.Since(start)

		tr := testResult{Name: "replicate-claude-opus-4.6", Duration: dur.String()}

		if err != nil {
			tr.Error = err.Error()
		} else {
			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)
			tr.Status = resp.StatusCode

			if resp.StatusCode != 200 && resp.StatusCode != 201 {
				tr.Error = string(respBody)
				if len(tr.Error) > 500 {
					tr.Error = tr.Error[:500]
				}
			} else {
				tr.OK = true
				var rr struct {
					Status string      `json:"status"`
					Output interface{} `json:"output"`
				}
				json.Unmarshal(respBody, &rr)
				tr.Response = fmt.Sprintf("status=%s output=%v", rr.Status, rr.Output)
				if len(tr.Response) > 200 {
					tr.Response = tr.Response[:200]
				}
			}
		}
		log.Printf("\U0001f50d DIAG %s → HTTP %d (%s)", tr.Name, tr.Status, dur)
		results = append(results, tr)
	}

	out := map[string]interface{}{
		"proxy_url": proxyURL,
		"results":   results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// HandleEnv GET /api/v1/diag/env — показывает какой proxy URL используется (без секретов)
func (h *DiagHandler) HandleEnv(w http.ResponseWriter, r *http.Request) {
	proxyURL := os.Getenv("OPENROUTER_PROXY_URL")
	if proxyURL == "" {
		proxyURL = "(not set, using default)"
	}
	hasKey := os.Getenv("OPENROUTER_API_KEY") != ""

	hasReplicate := os.Getenv("REPLICATE_API_TOKEN") != ""

	out := map[string]interface{}{
		"proxy_url":         proxyURL,
		"has_api_key":       hasKey,
		"api_key_hint":      fmt.Sprintf("sk-...%s", lastN(os.Getenv("OPENROUTER_API_KEY"), 4)),
		"has_replicate_key": hasReplicate,
		"replicate_hint":    fmt.Sprintf("r8-...%s", lastN(os.Getenv("REPLICATE_API_TOKEN"), 4)),
		"routing":           "Anthropic→Replicate, DeepSeek/Gemini→OpenRouter",
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
