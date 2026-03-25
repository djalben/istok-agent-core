package http

import (
	"encoding/json"
	"net/http"

	"github.com/istok/agent-core/internal/application/dto"
	"github.com/istok/agent-core/internal/application/usecases"
)

// StatsHandler обрабатывает запросы статистики
type StatsHandler struct {
	projectGenerator *usecases.ProjectGeneratorService
}

// NewStatsHandler создает новый handler
func NewStatsHandler(projectGenerator *usecases.ProjectGeneratorService) *StatsHandler {
	return &StatsHandler{
		projectGenerator: projectGenerator,
	}
}

// Handle обрабатывает GET /api/v1/stats
func (h *StatsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Получаем агента из сервиса
	agent := h.projectGenerator.GetAgent()

	// Формируем ответ
	response := dto.AgentStatsResponse{
		AgentID:              agent.ID,
		Name:                 agent.Name,
		Status:               string(agent.Status),
		TokenBalance:         agent.TokenBalance,
		TotalTasks:           agent.PerformanceMetrics.TotalTasks,
		SuccessRate:          agent.GetSuccessRate(),
		KnowledgeNodes:       agent.GetKnowledgeNodeCount(),
		LearningConfidence:   agent.GetLearningConfidence(),
		AverageTokensPerTask: agent.PerformanceMetrics.AverageTokensPerTask,
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка кодирования ответа")
		return
	}
}
