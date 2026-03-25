package http

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler обрабатывает health check запросы
type HealthHandler struct {
	startTime time.Time
}

// NewHealthHandler создает новый handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
	}
}

// Handle обрабатывает GET /api/v1/health
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	uptime := time.Since(h.startTime)

	response := map[string]interface{}{
		"status":  "healthy",
		"uptime":  uptime.String(),
		"service": "istok-agent-core",
		"version": "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
