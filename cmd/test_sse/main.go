// cmd/test_sse/main.go — диагностический тест SSE стрима
// Запуск: go run ./cmd/test_sse/main.go
// Убедись, что бэкенд запущен локально на :8080 (или передай VITE_API_BASE_URL)
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	apiBase := os.Getenv("VITE_API_BASE_URL")
	if apiBase == "" {
		apiBase = "https://web-production-18f7f.up.railway.app/api/v1"
	}

	url := apiBase + "/generate/stream"
	payload := `{"specification":"CyberBeans coffee landing page — dark neon theme","mode":"agent"}`

	fmt.Printf("🔍 POST %s\n\n", url)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(payload))
	if err != nil {
		log.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 8 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("❌ HTTP request failed: %v\n\nПроверь, что бэкенд запущен и VITE_API_BASE_URL указан корректно.", err)
	}
	defer resp.Body.Close()

	fmt.Printf("✅ HTTP %d   Content-Type: %s\n\n", resp.StatusCode, resp.Header.Get("Content-Type"))

	var eventType string
	buf := make([]byte, 8192)
	lineBuffer := ""

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			lineBuffer += string(buf[:n])
			for {
				idx := strings.Index(lineBuffer, "\n")
				if idx == -1 {
					break
				}
				line := strings.TrimRight(lineBuffer[:idx], "\r")
				lineBuffer = lineBuffer[idx+1:]

				if strings.HasPrefix(line, "event: ") {
					eventType = strings.TrimPrefix(line, "event: ")
				} else if strings.HasPrefix(line, "data: ") {
					raw := strings.TrimPrefix(line, "data: ")
					fmt.Printf("[event=%s]\n", eventType)

					var parsed map[string]interface{}
					if jsonErr := json.Unmarshal([]byte(raw), &parsed); jsonErr != nil {
						fmt.Printf("  ⚠️  JSON parse error: %v\n  raw: %s\n", jsonErr, raw)
					} else {
						for k, v := range parsed {
							fmt.Printf("  %-12s  GO_TYPE=%-10T  JS_WOULD_BE=%s\n",
								k, v, jsType(v))
						}
						// Check specifically for message field
						if msg, ok := parsed["message"]; ok {
							if _, isStr := msg.(string); !isStr {
								fmt.Printf("\n  ❌❌❌  ПРОБЛЕМА: поле 'message' НЕ строка! Тип=%T Значение=%v\n\n", msg, msg)
							} else {
								fmt.Printf("  ✅  message — строка: %q\n", msg)
							}
						}
					}
					fmt.Println("---")
				}
			}
		}
		if err == io.EOF {
			fmt.Println("\n✅ Стрим завершён.")
			break
		}
		if err != nil {
			log.Fatalf("❌ Read error: %v", err)
		}
	}
}

func jsType(v interface{}) string {
	switch v.(type) {
	case string:
		return "string ✅"
	case float64:
		return "number ✅"
	case bool:
		return "boolean ✅"
	case nil:
		return "null"
	case map[string]interface{}:
		return "object ❌ (ПРОБЛЕМА — фронтенд получит Object!)"
	case []interface{}:
		return "array ❌"
	default:
		return fmt.Sprintf("unknown(%T)", v)
	}
}
