package http

import (
	"encoding/json"
	"net/http"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
)

type EditorHandler struct {
	editor *usecases.Editor
}

// NewEditorHandler создаёт обработчик для агента-редактора.
func NewEditorHandler(editor *usecases.Editor) *EditorHandler {
	return &EditorHandler{editor: editor}
}

type editorRequest struct {
	SessionID string            `json:"session_id"`
	Message   string            `json:"message"`
	Files     map[string]string `json:"files"`
}

type editorResponse struct {
	Patches []usecases.FilePatch `json:"patches"`
}

// Handle processes POST /api/v1/editor/chat requests.
func (h *EditorHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})

		return
	}

	var req editorRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		_ = writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})

		return
	}

	if req.Message == "" {
		_ = writeJSON(w, http.StatusBadRequest, map[string]string{"error": "message is required"})

		return
	}
	if len(req.Files) == 0 {
		_ = writeJSON(w, http.StatusBadRequest, map[string]string{"error": "files map is required"})

		return
	}

	ctx := r.Context()
	logFrom(ctx).InfoContext(ctx, "editor request",
		"sessionId", req.SessionID,
		"message", req.Message,
		"files", len(req.Files),
	)

	patches, err := h.editor.Edit(ctx, req.Message, req.Files)
	if err != nil {
		logFrom(ctx).ErrorContext(ctx, "editor failed", "error", err)
		_ = writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})

		return
	}

	_ = writeJSON(w, http.StatusOK, editorResponse{Patches: patches})
}
