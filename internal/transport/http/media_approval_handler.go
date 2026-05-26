package http

import (
	"encoding/json"
	"net/http"

	"github.com/istok/agent-core/internal/application"
)

// MediaApprovalHandler обрабатывает решения пользователя по медиа-промптам (дизайн-ревью).
type MediaApprovalHandler struct {
	registry *application.ApprovalRegistry
}

// NewMediaApprovalHandler создаёт обработчик для POST /api/v1/generate/approve_media.
func NewMediaApprovalHandler(registry *application.ApprovalRegistry) *MediaApprovalHandler {
	return &MediaApprovalHandler{registry: registry}
}

type mediaApprovalRequest struct {
	SessionID string   `json:"session_id"`
	Approved  bool     `json:"approved"`
	Prompts   []string `json:"prompts"`
}

// Handle processes POST /api/v1/generate/approve_media requests.
func (h *MediaApprovalHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req mediaApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	if req.SessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session_id is required"})
		return
	}

	decision := application.MediaApprovalDecision{
		Approved: req.Approved,
		Prompts:  req.Prompts,
	}

	if err := h.registry.SubmitMedia(req.SessionID, decision); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "submitted"})
}
