package http

import (
	"encoding/json"
	"net/http"

	"github.com/istok/agent-core/internal/application"
)

// ResumeFundsHandler обрабатывает POST /api/v1/generate/resume_funds.
// Разблокирует горутину генерации, ожидающую пополнения баланса.
type ResumeFundsHandler struct {
	registry *application.FundsRegistry
}

// NewResumeFundsHandler creates a handler for the resume_funds endpoint.
func NewResumeFundsHandler(registry *application.FundsRegistry) *ResumeFundsHandler {
	return &ResumeFundsHandler{registry: registry}
}

type resumeFundsRequest struct {
	SessionID string `json:"session_id"`
}

// Handle processes POST /api/v1/generate/resume_funds requests.
func (h *ResumeFundsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req resumeFundsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	if req.SessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session_id is required"})
		return
	}

	if err := h.registry.Resume(req.SessionID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	globalFileStore.ClearPendingAction(req.SessionID)
	writeJSON(w, http.StatusOK, map[string]string{"status": "resumed"})
}
