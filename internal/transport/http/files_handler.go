package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// FilesHandler — GET /api/v1/generate/files?session_id=xxx
// Клиент вызывает после получения SSE event "done" для загрузки сгенерированных файлов.
// Обычный HTTP response с Content-Length — прокси обрабатывает корректно.
type FilesHandler struct{}

func NewFilesHandler() *FilesHandler {
	return &FilesHandler{}
}

func (h *FilesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		// Try path-based: /api/v1/generate/files/{sessionId}
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 0 {
			sessionID = parts[len(parts)-1]
		}
	}
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "session_id required")
		return
	}

	files := globalFileStore.Get(sessionID)
	complete := globalFileStore.IsComplete(sessionID)

	lastStatus := globalFileStore.GetStatus(sessionID)

	if files == nil {
		// Return 200 with empty files + complete=false so polling works
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"session_id":  sessionID,
			"files":       map[string]string{},
			"file_count":  0,
			"complete":    false,
			"last_status": lastStatus,
		})
		return
	}

	log.Printf("📦 FilesHandler: delivering %d files (complete=%v) for session %s", len(files), complete, sessionID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id":  sessionID,
		"files":       files,
		"file_count":  len(files),
		"complete":    complete,
		"last_status": lastStatus,
	})
}
