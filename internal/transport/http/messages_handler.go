package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/istok/agent-core/internal/application/usecases"
)

// MessagesHandler обрабатывает запросы к сообщениям проекта
type MessagesHandler struct {
	projectGenerator *usecases.ProjectGeneratorService
}

// NewMessagesHandler создает новый handler
func NewMessagesHandler(projectGenerator *usecases.ProjectGeneratorService) *MessagesHandler {
	return &MessagesHandler{
		projectGenerator: projectGenerator,
	}
}

// AgentMessage структура сообщения
type AgentMessage struct {
	ID        string                 `json:"id"`
	ProjectID string                 `json:"projectId"`
	Role      string                 `json:"role"` // "user", "agent", "system"
	Content   string                 `json:"content"`
	Timestamp string                 `json:"timestamp"`
	Status    string                 `json:"status"` // "pending", "streaming", "complete", "error"
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SendMessageRequest запрос на отправку сообщения
type SendMessageRequest struct {
	Content string                 `json:"content"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// HandleGetMessages обрабатывает GET /api/v1/projects/:id/messages
func (h *MessagesHandler) HandleGetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Извлекаем projectId из URL
	projectID := extractProjectID(r.URL.Path)
	if projectID == "" {
		writeError(w, http.StatusBadRequest, "Не указан ID проекта")
		return
	}

	// Возвращаем моковые сообщения (в будущем будет из БД)
	messages := []AgentMessage{
		{
			ID:        "msg_1",
			ProjectID: projectID,
			Role:      "system",
			Content:   "Сессия инициализирована. Модель: Claude 3.5 Sonnet. Готов к анализу.",
			Timestamp: time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			Status:    "complete",
		},
		{
			ID:        "msg_2",
			ProjectID: projectID,
			Role:      "user",
			Content:   "Проанализируй сайт и создай улучшенную версию",
			Timestamp: time.Now().Add(-3 * time.Minute).Format(time.RFC3339),
			Status:    "complete",
		},
		{
			ID:        "msg_3",
			ProjectID: projectID,
			Role:      "agent",
			Content:   "Запускаю глубокий анализ... Обнаружено 847 узлов. Генерирую оптимизированную архитектуру.",
			Timestamp: time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
			Status:    "complete",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(messages)
}

// HandleSendMessage обрабатывает POST /api/v1/projects/:id/messages
func (h *MessagesHandler) HandleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	projectID := extractProjectID(r.URL.Path)
	if projectID == "" {
		writeError(w, http.StatusBadRequest, "Не указан ID проекта")
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Ошибка парсинга запроса: %v", err))
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "Содержимое сообщения не может быть пустым")
		return
	}

	// Создаем ответное сообщение от агента
	response := AgentMessage{
		ID:        fmt.Sprintf("msg_%d", time.Now().Unix()),
		ProjectID: projectID,
		Role:      "agent",
		Content:   "Принято. Запускаю анализ и генерацию. Ожидаемое время: ~15 секунд.",
		Timestamp: time.Now().Format(time.RFC3339),
		Status:    "complete",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// extractProjectID извлекает ID проекта из URL
func extractProjectID(path string) string {
	// Ожидаем путь вида: /api/v1/projects/{id}/messages
	parts := strings.Split(path, "/")
	if len(parts) >= 5 && parts[3] == "projects" {
		return parts[4]
	}
	return ""
}
