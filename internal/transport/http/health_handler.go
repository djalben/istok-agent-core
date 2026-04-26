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
		"version": "2.0.0",
		"env":     env,
		"go":      runtime.Version(),
		"agents": map[string]string{
			"researcher":   "claude-3-7-sonnet-thinking (Anthropic Direct)",
			"brain":        "claude-3-7-sonnet-thinking (Anthropic Direct)",
			"director":     "claude-3-7-sonnet-thinking (Anthropic Direct)",
			"coder":        "claude-3-7-sonnet (Anthropic Direct, medium)",
			"designer":     "google/nano-banana (Replicate)",
			"videographer": "google/veo-3 (Replicate)",
			"validator":    "Verification Layer v3 (Quality Gate + Security Agent)",
		},
		"agent_count": 7,
		"secrets":     envStatus,
		"fsm_states":  12,
		"features":    []string{"DAG planner", "Lovable KB", "ProjectScanner", "Auto-Fix retry", "SSE streaming"},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
