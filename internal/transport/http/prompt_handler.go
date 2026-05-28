package http

import (
	"encoding/json"
	"net/http"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
)

// PromptHandler handles prompt enhancement requests.
type PromptHandler struct {
	helper *usecases.PromptHelper
}

// NewPromptHandler creates a new PromptHandler.
func NewPromptHandler(helper *usecases.PromptHelper) *PromptHandler {
	return &PromptHandler{helper: helper}
}

type enhanceRequest struct {
	Prompt       string `json:"prompt"`
	ReferenceURL string `json:"reference_url,omitempty"`
}

type enhanceResponse struct {
	Enhanced string `json:"enhanced"`
}

// HandleEnhance processes POST /api/v1/prompt/enhance requests.
func (h *PromptHandler) HandleEnhance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req enhanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	if req.Prompt == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "prompt is required"})
		return
	}

	enhanced, err := h.helper.Enhance(r.Context(), req.Prompt, req.ReferenceURL)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, enhanceResponse{Enhanced: enhanced})
}
