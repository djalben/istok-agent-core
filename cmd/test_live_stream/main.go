package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// test_live_stream — connects to Railway SSE endpoint and prints everything.
// Usage: go run cmd/test_live_stream/main.go [optional-spec]

func main() {
	railwayURL := os.Getenv("RAILWAY_URL")
	if railwayURL == "" {
		railwayURL = "https://web-production-18f7f.up.railway.app"
	}
	endpoint := railwayURL + "/api/v1/generate/stream"

	spec := "Landing page for AI startup NeoMind"
	if len(os.Args) > 1 {
		spec = strings.Join(os.Args[1:], " ")
	}

	payload, _ := json.Marshal(map[string]string{
		"specification": spec,
		"mode":          "code",
	})

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  ИСТОК SSE LIVE STREAM TEST")
	fmt.Printf("  URL:  %s\n", endpoint)
	fmt.Printf("  Spec: %s\n", spec)
	fmt.Printf("  Time: %s\n", time.Now().Format(time.RFC3339))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	client := &http.Client{Timeout: 10 * time.Minute}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Request create failed: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Connection failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("\n✅ Connected! HTTP %d | Content-Type: %s | Latency: %v\n\n",
		resp.StatusCode, resp.Header.Get("Content-Type"), time.Since(start))

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "❌ HTTP %d: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	// Read SSE stream line by line
	buf := make([]byte, 4096)
	var buffer string
	eventCount := 0
	var resultSize int

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			buffer += string(buf[:n])

			// Process complete SSE events (separated by \n\n)
			for {
				idx := strings.Index(buffer, "\n\n")
				if idx == -1 {
					break
				}
				event := buffer[:idx]
				buffer = buffer[idx+2:]

				if strings.TrimSpace(event) == "" {
					continue
				}

				elapsed := time.Since(start).Round(time.Millisecond)

				// Heartbeat
				if strings.HasPrefix(strings.TrimSpace(event), ":") {
					fmt.Printf("[%v] 💓 %s\n", elapsed, strings.TrimSpace(event))
					continue
				}

				eventCount++

				// Parse event type and data
				var eventType, eventData string
				for _, line := range strings.Split(event, "\n") {
					if strings.HasPrefix(line, "event: ") {
						eventType = strings.TrimPrefix(line, "event: ")
					} else if strings.HasPrefix(line, "data: ") {
						eventData = strings.TrimPrefix(line, "data: ")
					}
				}

				switch eventType {
				case "status":
					var status struct {
						Agent    string `json:"agent"`
						Status   string `json:"status"`
						Message  string `json:"message"`
						Progress int    `json:"progress"`
					}
					json.Unmarshal([]byte(eventData), &status)
					fmt.Printf("[%v] 📡 #%d STATUS | agent=%s status=%s progress=%d | %s\n",
						elapsed, eventCount, status.Agent, status.Status, status.Progress, status.Message)

				case "result":
					resultSize = len(eventData)
					var result struct {
						Files    map[string]string `json:"files"`
						Duration string            `json:"duration"`
					}
					if err := json.Unmarshal([]byte(eventData), &result); err != nil {
						fmt.Printf("[%v] 🎯 #%d RESULT | ⚠️ JSON PARSE ERROR: %v | raw_size=%d bytes\n",
							elapsed, eventCount, err, resultSize)
						// Try to show what we got
						if resultSize > 200 {
							fmt.Printf("    first200: %s\n", eventData[:200])
							fmt.Printf("    last200:  %s\n", eventData[resultSize-200:])
						} else {
							fmt.Printf("    raw: %s\n", eventData)
						}
					} else {
						fmt.Printf("[%v] 🎯 #%d RESULT | %d files | duration=%s | json_size=%d bytes\n",
							elapsed, eventCount, len(result.Files), result.Duration, resultSize)
						for fname, content := range result.Files {
							fmt.Printf("    📄 %s — %d chars | starts: %s\n",
								fname, len(content), content[:min(80, len(content))])
						}
					}

				case "error":
					fmt.Printf("[%v] ❌ #%d ERROR | %s\n", elapsed, eventCount, eventData)

				case "done":
					fmt.Printf("[%v] ✅ #%d DONE | %s\n", elapsed, eventCount, eventData)

				default:
					fmt.Printf("[%v] ❓ #%d UNKNOWN event=%q | data_len=%d\n",
						elapsed, eventCount, eventType, len(eventData))
				}
			}
		}

		if err != nil {
			if err == io.EOF {
				fmt.Printf("\n━━━ STREAM ENDED ━━━\n")
				fmt.Printf("Total events: %d | Result size: %d bytes | Duration: %v\n",
					eventCount, resultSize, time.Since(start).Round(time.Millisecond))
			} else {
				fmt.Fprintf(os.Stderr, "\n❌ Read error: %v\n", err)
			}
			break
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
