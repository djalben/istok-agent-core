package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/djalben/istok-agent-core/internal/infrastructure/media"
	"github.com/djalben/istok-agent-core/internal/ports"
)

// MediaPreviewHandler handles live image preview generation for the Media Studio modal.
type MediaPreviewHandler struct {
	llm ports.LLMProvider
}

// NewMediaPreviewHandler creates handler for POST /api/v1/generate/media/preview.
func NewMediaPreviewHandler(llm ports.LLMProvider) *MediaPreviewHandler {
	return &MediaPreviewHandler{llm: llm}
}

type mediaPreviewRequest struct {
	AssetID string `json:"asset_id"`
	Prompt  string `json:"prompt"`
	Width   int    `json:"width,omitempty"`
	Height  int    `json:"height,omitempty"`
}

type mediaPreviewResponse struct {
	URL    string `json:"url"`
	Source string `json:"source"` // "ai" | "stock"
	Error  string `json:"error,omitempty"`
}

// Handle processes POST /api/v1/generate/media/preview — generates a single image preview.
// Tries Replicate AI first; falls back to Unsplash stock photo if unavailable.
func (h *MediaPreviewHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req mediaPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	if req.Prompt == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "prompt is required"})
		return
	}

	width := req.Width
	height := req.Height
	if width <= 0 {
		width = 1344
	}
	if height <= 0 {
		height = 768
	}

	// Try AI generation (Replicate) first
	apiKey := os.Getenv("REPLICATE_API_TOKEN")
	if apiKey != "" {
		svc := media.NewMediaServiceWithLLM(apiKey, h.llm)
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
		defer cancel()

		promptSnippet := req.Prompt
		if len(promptSnippet) > 80 {
			promptSnippet = promptSnippet[:80]
		}
		log.Printf("🖼️ MediaPreview: AI generation (%dx%d) prompt=%q", width, height, promptSnippet)
		aiURL, err := svc.GenerateImage(ctx, req.Prompt, width, height)
		if err == nil && aiURL != "" {
			writeJSON(w, http.StatusOK, mediaPreviewResponse{URL: aiURL, Source: "ai"})
			return
		}
		log.Printf("⚠️ MediaPreview: AI failed, falling back to stock: %v", err)
	}

	// Fallback: Unsplash stock photo
	stockURL := unsplashURL(req.Prompt, width, height)
	log.Printf("📷 MediaPreview: stock fallback → %s", stockURL)
	writeJSON(w, http.StatusOK, mediaPreviewResponse{URL: stockURL, Source: "stock"})
}

// unsplashURL builds an Unsplash Source URL from a prompt.
// Extracts 2-3 keywords for best results.
func unsplashURL(prompt string, width, height int) string {
	keywords := extractKeywords(prompt)
	return fmt.Sprintf("https://source.unsplash.com/%dx%d/?%s", width, height, url.QueryEscape(keywords))
}

// extractKeywords pulls short search terms from a prompt.
func extractKeywords(prompt string) string {
	// Remove common AI prompt modifiers
	lower := strings.ToLower(prompt)
	stopwords := []string{
		"photorealistic", "cinematic", "8k", "4k", "ultra", "detailed",
		"professional", "studio lighting", "high quality", "beautiful",
		"modern", "futuristic", "dark", "gradient", "mesh",
	}
	for _, sw := range stopwords {
		lower = strings.ReplaceAll(lower, sw, "")
	}
	// Take first 3 meaningful words
	words := strings.Fields(lower)
	var kw []string
	for _, w := range words {
		w = strings.Trim(w, ".,!?;:\"'()-")
		if len(w) > 2 && len(kw) < 3 {
			kw = append(kw, w)
		}
	}
	if len(kw) == 0 {
		return "abstract,technology"
	}
	return strings.Join(kw, ",")
}
