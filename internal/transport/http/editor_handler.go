package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
)

// EditorHandler обрабатывает POST /api/v1/editor/chat — интерактивное редактирование через чат.
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
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req editorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	if req.Message == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "message is required"})
		return
	}
	if len(req.Files) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "files map is required"})
		return
	}

	log.Printf("🖊️ Editor request: session=%s, message=%q, files=%d", req.SessionID, req.Message, len(req.Files))

	patches, err := h.editor.Edit(r.Context(), req.Message, req.Files)
	if err != nil {
		log.Printf("❌ Editor error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, editorResponse{Patches: patches})
}
