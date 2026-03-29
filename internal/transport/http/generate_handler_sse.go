package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	log.Printf("DEBUG: SSE /generate/stream вызван method=%s origin=%s", r.Method, r.Header.Get("Origin"))

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

	log.Printf("DEBUG: Запуск генерации для режима %s, spec_len=%d", req.Mode, len(req.Specification))

	// Валидация
	if req.Specification == "" {
		writeError(w, http.StatusBadRequest, "Спецификация обязательна")
		return
	}

	// ── Проверяем Flusher ДО всего остального ──────────────────────────
	// КРИТИЧНО: проверка ДОЛЖНА быть до горутины. Если упадёт здесь —
	// горутина не запустится и контекст не отменится раньше времени.
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Printf("ERROR: ResponseWriter не поддерживает http.Flusher (%T)", w)
		writeError(w, http.StatusInternalServerError, "SSE не поддерживается")
		return
	}

	// ── Устанавливаем SSE-заголовки и немедленно флашим ───────────────
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // отключает буферизацию nginx/Railway
	w.WriteHeader(http.StatusOK)              // явно фиксируем 200 до первого Flush
	flusher.Flush()                           // отправляем заголовки клиенту

	// ── Создаем контекст с отменой ────────────────────────────────────
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	// ── Запускаем генерацию в горутине ПОСЛЕ проверки Flusher ─────────
	resultChan := make(chan *application.GenerationResult, 1)
	errorChan := make(chan error, 1)

	go func() {
		mode := application.ModeCode
		if req.Mode == "agent" {
			mode = application.ModeAgent
		} else if req.Mode == "synthesis" {
			mode = application.ModeSynthesis
		}
		log.Printf("DEBUG: горутина запущена mode=%s", mode)
		result, err := h.orchestrator.GenerateWithMode(ctx, req.Specification, req.URL, mode)
		if err != nil {
			log.Printf("ERROR: GenerateWithMode вернул ошибку: %v", err)
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// ── Получаем поток статусов ───────────────────────────────────────
	statusStream := h.orchestrator.GetStatusStream()

	// Отправляем начальное событие
	h.sendSSE(w, flusher, "status", map[string]interface{}{
		"agent":    "system",
		"status":   "started",
		"message":  "🚀 Запуск S-Tier AI Orchestrator...",
		"progress": 0,
	})

	// Keepalive ticker: sends SSE comment every 20s to prevent Railway/LB from closing idle connections
	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	// Слушаем статусы и результат
	for {
		select {
		case <-heartbeat.C:
			// SSE comment line — invisible to frontend, keeps TCP connection alive
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()

		case status := <-statusStream:
			// Отправляем статус агента — все string-поля явно приводим к string
			h.sendSSE(w, flusher, "status", map[string]interface{}{
				"agent":     fmt.Sprintf("%s", status.Agent),
				"status":    fmt.Sprintf("%s", status.Status),
				"message":   fmt.Sprintf("%s", status.Message),
				"progress":  status.Progress,
				"timestamp": status.Timestamp.Format(time.RFC3339),
			})

		case result := <-resultChan:
			// Генерация завершена успешно
			// NOTE: отправляем как "files" (map[string]string) — фронтенд проверяет result.files первым
			h.sendSSE(w, flusher, "result", map[string]interface{}{
				"files":    result.Code,
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
				"message": "⏱️ Превышено время ожидания (30 мин)",
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
