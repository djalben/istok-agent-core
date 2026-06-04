package http

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"sync"
	"time"

	"github.com/djalben/istok-agent-core/internal/application"
	"github.com/djalben/istok-agent-core/internal/application/dto"
	"github.com/djalben/istok-agent-core/internal/application/usecases"
	"gitlab.com/libs-artifex/wrapper"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — SSE Generate Handler
//  Server-Sent Events для real-time статусов
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// GenerateHandlerSSE обрабатывает запросы с SSE стримингом.
// Включает Output Guardrails: каждое исходящее SSE-событие проходит через
// PromptMaskFilter, который ищет утечки системных промптов и jailbreak-маркеров.
type GenerateHandlerSSE struct {
	orchestrator    *application.Orchestrator
	projectService  *usecases.ProjectService // Layer 2: авто-сохранение результата в БД
	guardrails      *PromptMaskFilter
	limiter         *genLimiter                   // защита от cost-DoS: семафор конкурентности + per-IP rate-limit
	activeSessions  map[string]context.CancelFunc // session_id → cancel of previous stream
	activeSessionMu sync.Mutex
}

// NewGenerateHandlerSSE создает новый SSE handler с включёнными Output Guardrails.
func NewGenerateHandlerSSE(orchestrator *application.Orchestrator, projectService *usecases.ProjectService) *GenerateHandlerSSE {
	return &GenerateHandlerSSE{
		orchestrator:   orchestrator,
		projectService: projectService,
		guardrails:     NewPromptMaskFilter(),
		limiter:        newGenLimiter(),
		activeSessions: make(map[string]context.CancelFunc),
	}
}

// HandleStream обрабатывает POST /api/v1/generate/stream.
func (h *GenerateHandlerSSE) HandleStream(w http.ResponseWriter, r *http.Request) {
	logSSERequestMeta(r.Context())

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")

		return
	}

	// Парсим запрос
	var req dto.GenerateProjectRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Ошибка парсинга запроса: %v", err))

		return
	}
	logFrom(r.Context()).InfoContext(
		r.Context(), "sse generation started",
		"mode", req.Mode,
		"spec_len", len(req.Specification),
	)

	// Валидация
	if req.Specification == "" {
		writeError(w, http.StatusBadRequest, "Спецификация обязательна")

		return
	}

	// ── Layer 2: owner_id из JWT (route за AuthMiddleware — всегда присутствует) ──
	ownerID, _ := userIDFromContext(r.Context())

	// ── Cost-DoS guard: per-IP rate-limit + concurrency semaphore ──
	// Проверяем ДО SSE-заголовков, чтобы отдать честный JSON 429.
	ip := clientIP(r)
	if !h.limiter.allowIP(ip) {
		logRateLimitExceeded(r.Context())
		writeError(w, http.StatusTooManyRequests, "Слишком много запросов. Повторите через минуту.")

		return
	}
	releaseSlot, ok := h.limiter.acquire()
	if !ok {
		logConcurrencyLimitReached(r.Context())
		writeError(w, http.StatusTooManyRequests, "Сервер занят: достигнут лимит одновременных генераций. Повторите позже.")

		return
	}
	slotReleased := false
	defer func() {
		// Если генерация так и не стартовала в фоне — освобождаем слот при выходе хендлера.
		if !slotReleased {
			releaseSlot()
		}
	}()

	sess, ok := h.beginSSEStream(w, r, req, ownerID, releaseSlot)
	if !ok {
		return
	}
	slotReleased = true
	sess.runEventLoop()
}

// persistGenerated сохраняет сгенерированные файлы в БД через ProjectService (Layer 2).
// Возвращает id проекта (новый или обновлённый). Без owner/файлов — no-op.
func (h *GenerateHandlerSSE) persistGenerated(ctx context.Context, ownerID string, req dto.GenerateProjectRequest, files map[string]string) (string, error) {
	if h.projectService == nil || ownerID == "" || len(files) == 0 {
		return req.ProjectID, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	p, err := h.projectService.SaveGenerated(ctx, ownerID, req.ProjectID, req.Name, req.Framework, req.Specification, files)
	if err != nil {
		return req.ProjectID, wrapper.Wrap(err)
	}

	return p.ID, nil
}

// cancelPreviousSession kills the old generation goroutine for a session and clears approval registry.
func (h *GenerateHandlerSSE) cancelPreviousSession(ctx context.Context, sessionID string) {
	if sessionID == "" {
		return
	}
	h.activeSessionMu.Lock()
	if prev, ok := h.activeSessions[sessionID]; ok {
		sseLog(ctx).InfoContext(ctx, "cancelling previous stream", "session_id", sessionID)
		prev() // cancel context → goroutine dies via ctx.Done()
		delete(h.activeSessions, sessionID)
	}
	h.activeSessionMu.Unlock()

	// Also clear any stale approval channel
	if reg := h.orchestrator.GetApprovalRegistry(); reg != nil {
		reg.Cleanup(sessionID)
	}
}

// registerSession stores the cancel func for a new stream session.
func (h *GenerateHandlerSSE) registerSession(sessionID string, cancel context.CancelFunc) {
	if sessionID == "" {
		return
	}
	h.activeSessionMu.Lock()
	h.activeSessions[sessionID] = cancel
	h.activeSessionMu.Unlock()
}

// unregisterSession removes a session from the active map on stream completion.
func (h *GenerateHandlerSSE) unregisterSession(sessionID string) {
	if sessionID == "" {
		return
	}
	h.activeSessionMu.Lock()
	delete(h.activeSessions, sessionID)
	h.activeSessionMu.Unlock()
}

// sendSSE отправляет SSE событие через Output Guardrails.
// Все string-поля payload-а проходят через PromptMaskFilter; при детекции
// утечки системного промпта/jailbreak-маркера поле подменяется на refusal.
func (h *GenerateHandlerSSE) sendSSE(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, event string, data any) {
	data = h.guardrailSSEPayload(ctx, event, data)

	jsonData, err := json.Marshal(data)
	if err != nil {
		sseLog(ctx).ErrorContext(ctx, "sendSSE json.Marshal failed", "event", event, "error", err)

		return
	}

	if len(jsonData) > 4096 {
		sseLog(ctx).WarnContext(ctx, "sse oversized payload", "event", event, "bytes", len(jsonData))
	}

	_, err = fmt.Fprintf(w, "event: %s\n", event)
	if err != nil {
		sseLog(ctx).ErrorContext(ctx, "sendSSE write failed", "event", event, "error", err)

		return
	}
	_, err = fmt.Fprintf(w, "data: %s\n\n", jsonData)
	if err != nil {
		sseLog(ctx).ErrorContext(ctx, "sendSSE data write failed", "event", event, "error", err)

		return
	}
	flusher.Flush()
}

func (h *GenerateHandlerSSE) spawnAutoApproveOnDisconnect(genCtx context.Context, sessionID string) {
	if sessionID == "" {
		return
	}
	registry := h.orchestrator.GetApprovalRegistry()
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		timeout := time.After(5 * time.Minute)
		planApproved, mediaApproved := false, false
		for {
			select {
			case <-ticker.C:
				if !planApproved {
					err := registry.Submit(sessionID, application.ApprovalDecision{Approved: true, Feedback: "auto-approved (SSE disconnected)"})
					if err == nil {
						sseLog(genCtx).InfoContext(genCtx, "auto-approved plan", "session_id", sessionID)
						planApproved = true
					}
				}
				if !mediaApproved {
					err := registry.SubmitMedia(sessionID, application.MediaApprovalDecision{Approved: true})
					if err == nil {
						sseLog(genCtx).InfoContext(genCtx, "auto-approved media", "session_id", sessionID)
						mediaApproved = true
					}
				}
				if planApproved && mediaApproved {
					return
				}
			case <-timeout:
				sseLog(genCtx).WarnContext(genCtx, "auto-approve timeout", "session_id", sessionID)

				return
			case <-genCtx.Done():
				return
			}
		}
	}()
}

func (h *GenerateHandlerSSE) guardrailSSEPayload(ctx context.Context, event string, data any) any {
	if h.guardrails == nil {
		return data
	}
	switch payload := data.(type) {
	case map[string]any:
		return h.guardrailMapPayload(ctx, event, payload)
	case string:
		sanitized, leaked, _ := h.guardrails.Sanitize(payload)
		if leaked {
			sseLog(ctx).WarnContext(ctx, "output guardrail triggered", "payload_type", "string")
		}

		return sanitized
	default:
		return data
	}
}

func (h *GenerateHandlerSSE) guardrailMapPayload(ctx context.Context, event string, payload map[string]any) any {
	fileEvents := event == "file" || event == "files_batch" || event == "file_start" || event == "file_chunk" || event == "file_end"
	if !fileEvents {
		sanitized, leakCount, _ := h.guardrails.SanitizeMap(payload)
		if leakCount > 0 {
			sseLog(ctx).WarnContext(ctx, "output guardrail triggered", "payload_type", "map", "leaks", leakCount)
		}

		return sanitized
	}
	name, ok := payload["name"].(string)
	if !ok {
		return payload
	}
	sanitized, leaked, _ := h.guardrails.Sanitize(name)
	if !leaked {
		return payload
	}
	sseLog(ctx).WarnContext(ctx, "output guardrail triggered", "payload_type", "file_name")
	newPayload := make(map[string]any, len(payload))
	maps.Copy(newPayload, payload)
	newPayload["name"] = sanitized

	return newPayload
}
