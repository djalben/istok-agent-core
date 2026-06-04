package http

import (
	"encoding/json"

	"net/http"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
	"log/slog"

	// EditComponentHandler обрабатывает POST /api/v1/generate/edit — точечное редактирование одного файла.
	"fmt"
)

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
		_ = writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})

		return
	}

	var req usecases.ComponentEditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})

		return
	}

	if req.FilePath == "" {
		_ = writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file_path is required"})

		return
	}
	if req.CurrentCode == "" {
		_ = writeJSON(w, http.StatusBadRequest, map[string]string{"error": "current_code is required"})

		return
	}
	if req.Prompt == "" {
		_ = writeJSON(w, http.StatusBadRequest, map[string]string{"error": "prompt is required"})

		return
	}
	slog.Info(fmt.Sprintf("🔧 EditComponent: file=%s, prompt=%q", req.FilePath, req.Prompt))

	result, err := h.editor.Edit(r.Context(), req)
	if err != nil {
		slog.Info(fmt.Sprintf("❌ EditComponent error: %v", err))
		_ = writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})

		return
	}

	_ = writeJSON(w, http.StatusOK, result)
}
