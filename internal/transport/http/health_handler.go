package http

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"
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

	// Check critical env vars
	envStatus := map[string]string{}
	for _, key := range []string{"ANTHROPIC_API_KEY", "REPLICATE_API_TOKEN", "JWT_SECRET", "CORS_ALLOWED_ORIGINS"} {
		if os.Getenv(key) != "" {
			envStatus[key] = "set"
		} else {
			envStatus[key] = "missing"
		}
	}

	env := os.Getenv("RAILWAY_ENVIRONMENT")
	if env == "" {
		env = os.Getenv("GO_ENV")
	}
	if env == "" {
		env = "development"
	}

	response := map[string]interface{}{
		"status":  "healthy",
		"uptime":  uptime.String(),
		"service": "istok-agent-core",
		"version": "3.0.0",
		"env":     env,
		"go":      runtime.Version(),
		"agents": []string{
			"director", "researcher", "brain", "architect", "planner",
			"coder", "designer", "security", "tester", "ui_reviewer",
		},
		"agent_count":       10,
		"secrets":           envStatus,
		"fsm_states":        12,
		"verification_gate": []string{"security", "tester", "ui_reviewer"},
		"features": []string{
			"DAG planner", "10-agent pipeline", "Verification Gate",
			"SSE streaming with Agent field", "Auto-Fix retry",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
