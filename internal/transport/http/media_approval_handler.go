package http

import (
	"encoding/json"
	"net/http"

	"github.com/djalben/istok-agent-core/internal/application"
	"github.com/djalben/istok-agent-core/internal/domain"
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
	SessionID string              `json:"session_id"`
	Approved  bool                `json:"approved"`
	Assets    []domain.MediaAsset `json:"assets"`
}

// Handle processes POST /api/v1/generate/approve_media requests.
func (h *MediaApprovalHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})

		return
	}

	var req mediaApprovalRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		_ = writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})

		return
	}

	if req.SessionID == "" {
		_ = writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session_id is required"})

		return
	}

	decision := application.MediaApprovalDecision{
		Approved: req.Approved,
		Assets:   req.Assets,
	}

	err = h.registry.SubmitMedia(r.Context(), req.SessionID, decision)
	if err != nil {
		_ = writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})

		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]string{"status": "submitted"})
}
