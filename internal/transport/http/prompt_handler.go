package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/istok/agent-core/internal/application/usecases"
	"github.com/istok/agent-core/internal/ports"
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
		if errors.Is(err, ports.ErrInsufficientFunds) {
			writeJSON(w, http.StatusPaymentRequired, map[string]string{
				"error":   "insufficient_funds",
				"message": "Недостаточно средств на балансе AI-провайдера",
			})
			return
		}
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, enhanceResponse{Enhanced: enhanced})
}
