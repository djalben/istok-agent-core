package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/djalben/istok-agent-core/internal/application"
)

// WatcherHandler обрабатывает webhook-сигналы об ошибках.
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
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())

		return
	}

	if payload.StatusCode == 0 {
		writeError(w, http.StatusBadRequest, "status_code is required")

		return
	}
	slog.Info(fmt.Sprintf("🔭 Webhook received: %d %s %s from %s", payload.StatusCode, payload.Method, payload.Path, payload.Source))

	report := h.watcher.HandleError(r.Context(), payload)

	_ = writeJSON(w, http.StatusOK, report)
}

// HandleReports GET /api/v1/internal/watcher/reports
// Возвращает все отчёты о диагностике.
func (h *WatcherHandler) HandleReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET only")

		return
	}

	reports := h.watcher.GetReports()
	_ = writeJSON(w, http.StatusOK, map[string]any{
		"count":   len(reports),
		"reports": reports,
	})
}
