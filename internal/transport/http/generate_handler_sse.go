package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/istok/agent-core/internal/application"
	"github.com/istok/agent-core/internal/application/dto"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — SSE Generate Handler
//  Server-Sent Events для real-time статусов
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// GenerateHandlerSSE обрабатывает запросы с SSE стримингом
type GenerateHandlerSSE struct {
	orchestrator *application.Orchestrator
}

// NewGenerateHandlerSSE создает новый SSE handler
func NewGenerateHandlerSSE(orchestrator *application.Orchestrator) *GenerateHandlerSSE {
	return &GenerateHandlerSSE{
		orchestrator: orchestrator,
	}
}

// HandleStream обрабатывает POST /api/v1/generate/stream
func (h *GenerateHandlerSSE) HandleStream(w http.ResponseWriter, r *http.Request) {
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

	// Настраиваем SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Создаем контекст с отменой
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	// Запускаем генерацию в горутине
	resultChan := make(chan *application.GenerationResult, 1)
	errorChan := make(chan error, 1)

	go func() {
		mode := application.ModeCode
		if req.Mode == "agent" {
			mode = application.ModeAgent
		}
		result, err := h.orchestrator.GenerateWithMode(ctx, req.Specification, req.URL, mode)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// Получаем поток статусов
	statusStream := h.orchestrator.GetStatusStream()
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "SSE не поддерживается")
		return
	}

	// Отправляем начальное событие
	h.sendSSE(w, flusher, "status", map[string]interface{}{
		"agent":    "system",
		"status":   "started",
		"message":  "🚀 Запуск S-Tier AI Orchestrator...",
		"progress": 0,
	})

	// Слушаем статусы и результат
	for {
		select {
		case status := <-statusStream:
			// Отправляем статус агента
			h.sendSSE(w, flusher, "status", map[string]interface{}{
				"agent":     string(status.Agent),
				"status":    status.Status,
				"message":   status.Message,
				"progress":  status.Progress,
				"timestamp": status.Timestamp.Format(time.RFC3339),
			})

		case result := <-resultChan:
			// Генерация завершена успешно
			h.sendSSE(w, flusher, "result", map[string]interface{}{
				"code":     result.Code,
				"assets":   result.Assets,
				"video":    result.Video,
				"duration": result.Duration.String(),
			})

			h.sendSSE(w, flusher, "done", map[string]interface{}{
				"message": "✅ Проект успешно сгенерирован",
			})
			return

		case err := <-errorChan:
			// Ошибка генерации
			h.sendSSE(w, flusher, "error", map[string]interface{}{
				"message": fmt.Sprintf("❌ Ошибка: %v", err),
			})
			return

		case <-ctx.Done():
			// Таймаут или отмена
			h.sendSSE(w, flusher, "error", map[string]interface{}{
				"message": "⏱️ Превышено время ожидания",
			})
			return
		}
	}
}

// sendSSE отправляет SSE событие
func (h *GenerateHandlerSSE) sendSSE(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()
}
