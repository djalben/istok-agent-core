package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/istok/agent-core/internal/infrastructure/media"
	"github.com/istok/agent-core/internal/ports"
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
	Prompt string `json:"prompt"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type mediaPreviewResponse struct {
	URL   string `json:"url"`
	Error string `json:"error,omitempty"`
}

// Handle processes POST /api/v1/generate/media/preview — generates a single image preview.
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

	// Defaults
	width := req.Width
	height := req.Height
	if width <= 0 {
		width = 1344
	}
	if height <= 0 {
		height = 768
	}

	apiKey := os.Getenv("REPLICATE_API_TOKEN")
	if apiKey == "" {
		writeJSON(w, http.StatusServiceUnavailable, mediaPreviewResponse{Error: "media service unavailable"})
		return
	}

	svc := media.NewMediaServiceWithLLM(apiKey, h.llm)

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
	defer cancel()

	log.Printf("🖼️ MediaPreview: generating preview (%dx%d) prompt=%q", width, height, req.Prompt[:min(len(req.Prompt), 80)])
	url, err := svc.GenerateImage(ctx, req.Prompt, width, height)
	if err != nil {
		log.Printf("⚠️ MediaPreview: generation failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, mediaPreviewResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, mediaPreviewResponse{URL: url})
}
