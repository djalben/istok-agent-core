package main

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

func main() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("Set OPENROUTER_API_KEY")
	}

	proxyURL := os.Getenv("OPENROUTER_PROXY_URL")
	if proxyURL == "" {
		proxyURL = "https://openrouter.ai/api/v1"
	}
	endpoint := proxyURL + "/chat/completions"

	// Test 1: Simple call (no reasoning)
	fmt.Println("=== TEST 1: Claude Opus 4.6 — simple call (no reasoning) ===")
	test(endpoint, apiKey, map[string]interface{}{
		"model": "anthropic/claude-opus-4.6",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a system architect. Output pure JSON only."},
			{"role": "user", "content": "Return a JSON object: {\"status\": \"ok\", \"model\": \"opus-4.6\"}"},
		},
		"max_tokens":  256,
		"temperature": 0.7,
	})

	// Test 2: With reasoning params (what defineArchitecture sends)
	fmt.Println("\n=== TEST 2: Claude Opus 4.6 — with reasoning params ===")
	test(endpoint, apiKey, map[string]interface{}{
		"model": "anthropic/claude-opus-4.6",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a system architect. Output pure JSON only."},
			{"role": "user", "content": "Return a JSON object: {\"status\": \"ok\", \"model\": \"opus-4.6\"}"},
		},
		"max_tokens":  256,
		"temperature": 1,
		"reasoning": map[string]interface{}{
			"effort": "high",
		},
		"thinking": map[string]interface{}{
			"type":          "enabled",
			"budget_tokens": 2048,
		},
	})

	// Test 3: Fallback model (qwen)
	fmt.Println("\n=== TEST 3: Qwen 2.5 72B — fallback test ===")
	test(endpoint, apiKey, map[string]interface{}{
		"model": "qwen/qwen-2.5-72b-instruct",
		"messages": []map[string]string{
			{"role": "user", "content": "Return JSON: {\"status\": \"ok\"}"},
		},
		"max_tokens":  128,
		"temperature": 0.7,
	})

	// Test 4: DeepSeek (researcher model)
	fmt.Println("\n=== TEST 4: DeepSeek V3.2 Speciale — researcher test ===")
	test(endpoint, apiKey, map[string]interface{}{
		"model": "deepseek/deepseek-v3.2-speciale",
		"messages": []map[string]string{
			{"role": "user", "content": "Return JSON: {\"status\": \"ok\"}"},
		},
		"max_tokens":  128,
		"temperature": 0.3,
	})
}

func test(endpoint, apiKey string, payload map[string]interface{}) {
	body, _ := json.Marshal(payload)
	fmt.Printf("URL: %s\nModel: %s\nPayload size: %d bytes\n", endpoint, payload["model"], len(body))

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://istok-agent-core.vercel.app")
	req.Header.Set("X-Title", "Istok Agent Core")
	req.Header.Set("User-Agent", "IstokAgent/2.0")

	client := &http.Client{Timeout: 3 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("HTTP %d\n", resp.StatusCode)

	if resp.StatusCode != 200 {
		fmt.Printf("❌ ERROR: %s\n", string(respBody))
		return
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(respBody, &result)
	if len(result.Choices) > 0 {
		content := result.Choices[0].Message.Content
		if len(content) > 300 {
			content = content[:300] + "..."
		}
		fmt.Printf("✅ Response: %s\n", content)
	} else {
		fmt.Printf("⚠️ No choices in response: %s\n", string(respBody[:min(len(respBody), 500)]))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
