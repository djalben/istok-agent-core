package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/istok/agent-core/internal/application/dto"
	"github.com/istok/agent-core/internal/application/usecases"
)

// GenerateHandler обрабатывает запросы на генерацию проектов
type GenerateHandler struct {
	projectGenerator *usecases.ProjectGeneratorService
}

// NewGenerateHandler создает новый handler
func NewGenerateHandler(projectGenerator *usecases.ProjectGeneratorService) *GenerateHandler {
	return &GenerateHandler{
		projectGenerator: projectGenerator,
	}
}

// Handle обрабатывает POST /api/v1/generate
func (h *GenerateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Парсим запрос
	var req dto.GenerateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Ошибка парсинга запроса: %v", err))
		return
	}

	// Валидация
	if req.Specification == "" {
		writeError(w, http.StatusBadRequest, "Спецификация обязательна")
		return
	}

	// Устанавливаем дефолтные значения
	if req.Language == "" {
		req.Language = "JavaScript"
	}
	if req.Framework == "" {
		req.Framework = "React"
	}

	// Генерируем проект
	response, err := h.projectGenerator.GenerateProject(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Ошибка генерации: %v", err))
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Ошибка кодирования ответа: %v", err))
		return
	}
}
