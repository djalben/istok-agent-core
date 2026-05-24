package http

import (
	"encoding/json"
	"net/http"

	"github.com/istok/agent-core/internal/application"
)

// ApprovalHandler обрабатывает решения пользователя по утверждению архитектуры.
type ApprovalHandler struct {
	registry *application.ApprovalRegistry
}

// NewApprovalHandler создаёт обработчик для POST /api/v1/generate/approve.
func NewApprovalHandler(registry *application.ApprovalRegistry) *ApprovalHandler {
	return &ApprovalHandler{registry: registry}
}

type approvalRequest struct {
	SessionID string `json:"session_id"`
	Approved  bool   `json:"approved"`
	Feedback  string `json:"feedback,omitempty"`
}

// Handle processes POST /api/v1/generate/approve requests.
func (h *ApprovalHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req approvalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	if req.SessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session_id is required"})
		return
	}

	decision := application.ApprovalDecision{
		Approved: req.Approved,
		Feedback: req.Feedback,
	}

	if err := h.registry.Submit(req.SessionID, decision); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "submitted"})
}
