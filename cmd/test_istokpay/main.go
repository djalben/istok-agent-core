// cmd/test_istokpay/main.go — E2E тест генерации IstokPay через Railway + Cloudflare Proxy
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

	spec := map[string]string{
		"specification": `IstokPay — full-featured payment system (Revolut clone).
Dark premium UI with glassmorphism.
Features:
- P2P transfers by phone number
- Virtual cards with spending limits
- Transaction history with filters
- Dashboard with balance and analytics charts
- Multi-currency wallet (USD, EUR, RUB)
- User profile and settings
- JWT authentication (login/register)
Stack: React + TailwindCSS + shadcn/ui frontend, Go Fiber backend, PostgreSQL.
Generate a COMPLETE working prototype with all pages.`,
		"mode": "agent",
	}

	payload, _ := json.Marshal(spec)
	fmt.Printf("🚀 IstokPay Full-Stack Generation\n")
	fmt.Printf("🔗 POST %s\n", url)
	fmt.Printf("📦 Payload: %d bytes\n\n", len(payload))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("❌ HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("✅ HTTP %d   Content-Type: %s\n\n", resp.StatusCode, resp.Header.Get("Content-Type"))

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("❌ Non-200 response: %s", string(body))
	}

	var eventType string
	buf := make([]byte, 16384)
	lineBuffer := ""
	eventCount := 0
	var resultData map[string]interface{}

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
					eventCount++

					var parsed map[string]interface{}
					if jsonErr := json.Unmarshal([]byte(raw), &parsed); jsonErr != nil {
						fmt.Printf("  ⚠️  JSON parse error: %v\n", jsonErr)
						continue
					}

					switch eventType {
					case "status":
						agent, _ := parsed["agent"].(string)
						msg, _ := parsed["message"].(string)
						progress, _ := parsed["progress"].(float64)
						fmt.Printf("[%s] %s (%.0f%%)\n", agent, msg, progress)
					case "result":
						resultData = parsed
						fmt.Printf("\n🎉 [RESULT] received!\n")
						if files, ok := parsed["files"].(map[string]interface{}); ok {
							fmt.Printf("   📁 Files generated: %d\n", len(files))
							for name := range files {
								fmt.Printf("      → %s\n", name)
							}
						}
						if assets, ok := parsed["assets"].(map[string]interface{}); ok {
							fmt.Printf("   🎨 Assets: %d\n", len(assets))
						}
						if dur, ok := parsed["duration"].(string); ok {
							fmt.Printf("   ⏱️  Duration: %s\n", dur)
						}
					case "done":
						msg, _ := parsed["message"].(string)
						fmt.Printf("\n✅ [DONE] %s\n", msg)
					case "error":
						msg, _ := parsed["message"].(string)
						fmt.Printf("\n❌ [ERROR] %s\n", msg)
					}
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("\n⚠️ Stream ended: %v\n", err)
			break
		}
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📊 Total SSE events: %d\n", eventCount)

	if resultData != nil {
		if files, ok := resultData["files"].(map[string]interface{}); ok {
			fmt.Printf("📁 Generated files: %d\n", len(files))
			totalChars := 0
			for _, content := range files {
				if s, ok := content.(string); ok {
					totalChars += len(s)
				}
			}
			fmt.Printf("📝 Total code size: %d chars\n", totalChars)
		}
		fmt.Println("\n✅ IstokPay generation SUCCESSFUL!")
	} else {
		fmt.Println("\n❌ No result event received — generation may have failed")
	}
}
