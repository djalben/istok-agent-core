package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/istok/agent-core/internal/application"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  Watcher Handler — HTTP endpoints for error webhook
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// WatcherHandler обрабатывает webhook-сигналы об ошибках
type WatcherHandler struct {
	watcher *application.Watcher
}

// NewWatcherHandler creates handler with Watcher reference.
func NewWatcherHandler(watcher *application.Watcher) *WatcherHandler {
	return &WatcherHandler{watcher: watcher}
}

// HandleErrorWebhook POST /api/v1/internal/error-webhook
// Принимает сигнал об ошибке, запускает triage, возвращает отчёт.
func (h *WatcherHandler) HandleErrorWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}

	var payload application.ErrorWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if payload.StatusCode == 0 {
		writeError(w, http.StatusBadRequest, "status_code is required")
		return
	}

	log.Printf("🔭 Webhook received: %d %s %s from %s", payload.StatusCode, payload.Method, payload.Path, payload.Source)

	report := h.watcher.HandleError(r.Context(), payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}

// HandleReports GET /api/v1/internal/watcher/reports
// Возвращает все отчёты о диагностике.
func (h *WatcherHandler) HandleReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}

	reports := h.watcher.GetReports()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":   len(reports),
		"reports": reports,
	})
}
