package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
)

// EditComponentHandler обрабатывает POST /api/v1/generate/edit — точечное редактирование одного файла.
type EditComponentHandler struct {
	editor *usecases.ComponentEditor
}

// NewEditComponentHandler создаёт обработчик.
func NewEditComponentHandler(editor *usecases.ComponentEditor) *EditComponentHandler {
	return &EditComponentHandler{editor: editor}
}

// Handle processes POST /api/v1/generate/edit requests.
func (h *EditComponentHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req usecases.ComponentEditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	if req.FilePath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "filePath is required"})
		return
	}
	if req.CurrentCode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "currentCode is required"})
		return
	}
	if req.Prompt == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "prompt is required"})
		return
	}

	log.Printf("🔧 EditComponent: file=%s, prompt=%q", req.FilePath, req.Prompt)

	result, err := h.editor.Edit(r.Context(), req)
	if err != nil {
		log.Printf("❌ EditComponent error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}
