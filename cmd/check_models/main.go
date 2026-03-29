package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	resp, err := http.Get("https://openrouter.ai/api/v1/models")
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("parse failed: %v", err)
	}

	fmt.Println("=== CLAUDE MODELS ===")
	for _, m := range result.Data {
		if strings.Contains(strings.ToLower(m.ID), "claude") {
			fmt.Println(m.ID)
		}
	}

	fmt.Println("\n=== DEEPSEEK MODELS ===")
	for _, m := range result.Data {
		if strings.Contains(strings.ToLower(m.ID), "deepseek") {
			fmt.Println(m.ID)
		}
	}
}
