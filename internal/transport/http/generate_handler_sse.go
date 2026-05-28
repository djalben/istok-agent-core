package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/istok/agent-core/internal/application"
	"github.com/istok/agent-core/internal/application/dto"
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
	guardrails      *PromptMaskFilter
	activeSessions  map[string]context.CancelFunc // session_id → cancel of previous stream
	activeSessionMu sync.Mutex
}

// NewGenerateHandlerSSE создает новый SSE handler с включёнными Output Guardrails.
func NewGenerateHandlerSSE(orchestrator *application.Orchestrator) *GenerateHandlerSSE {
	return &GenerateHandlerSSE{
		orchestrator:   orchestrator,
		guardrails:     NewPromptMaskFilter(),
		activeSessions: make(map[string]context.CancelFunc),
	}
}

// cancelPreviousSession kills the old generation goroutine for a session and clears approval registry.
func (h *GenerateHandlerSSE) cancelPreviousSession(sessionID string) {
	if sessionID == "" {
		return
	}
	h.activeSessionMu.Lock()
	if prev, ok := h.activeSessions[sessionID]; ok {
		log.Printf("⚡ Cancelling previous stream for session %s", sessionID)
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

// sseChunkSize — hard limit for RAW content in a single SSE chunk (500 bytes).
// After JSON encoding (escaping \n, \", \t, <, > etc.) 500 bytes can expand to ~1KB.
// Railway/Envoy proxy drops HTTP/2 connections when data: line exceeds buffer.
const sseChunkSize = 500

// fileStreamDelay — pause between complete files to let proxy drain buffer.
const fileStreamDelay = 100 * time.Millisecond

// ChunkString splits s into slices of at most chunkSize bytes.
func ChunkString(s string, chunkSize int) []string {
	if len(s) <= chunkSize {
		return []string{s}
	}
	var chunks []string
	for len(s) > 0 {
		end := chunkSize
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[:end])
		s = s[end:]
	}
	return chunks
}

// sendFileEvent sends a file via chunked SSE events to prevent HTTP/2 frame overflow.
// Small files (<= sseChunkSize) go as a single "file" event (backward-compat).
// Large files: file_start (no content) → file_chunk* → file_end.
func (h *GenerateHandlerSSE) sendFileEvent(w http.ResponseWriter, flusher http.Flusher, name, content string) {
	if len(content) <= sseChunkSize {
		h.sendSSE(w, flusher, "file", map[string]interface{}{
			"name":    name,
			"content": content,
		})
		time.Sleep(fileStreamDelay)
		return
	}

	// Large file — chunked delivery
	chunks := ChunkString(content, sseChunkSize)

	// file_start: metadata only, no content
	h.sendSSE(w, flusher, "file_start", map[string]interface{}{
		"name":         name,
		"total_chunks": len(chunks),
	})

	// file_chunk: each chunk individually
	for i, chunk := range chunks {
		time.Sleep(10 * time.Millisecond)
		h.sendSSE(w, flusher, "file_chunk", map[string]interface{}{
			"name":    name,
			"content": chunk,
			"index":   i,
		})
	}

	// file_end: signal assembly complete
	h.sendSSE(w, flusher, "file_end", map[string]interface{}{
		"name": name,
	})

	// Inter-file delay: let proxy drain buffer before next file
	time.Sleep(fileStreamDelay)
}

// HandleStream обрабатывает POST /api/v1/generate/stream
func (h *GenerateHandlerSSE) HandleStream(w http.ResponseWriter, r *http.Request) {
	log.Printf("━━━ SSE REQUEST ARRIVED ━━━ method=%s origin=%s remote=%s content-length=%s time=%s",
		r.Method, r.Header.Get("Origin"), r.RemoteAddr, r.Header.Get("Content-Length"), time.Now().Format(time.RFC3339))

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
	w.Header().Set("X-Accel-Buffering", "no")           // отключает буферизацию nginx/Railway
	w.Header().Set("X-Content-Type-Options", "nosniff") // предотвращает MIME-sniffing
	w.Header().Set("Content-Encoding", "identity")      // явно отключаем gzip (ломает SSE на больших объёмах)
	w.WriteHeader(http.StatusOK)                        // явно фиксируем 200 до первого Flush
	flusher.Flush()                                     // отправляем заголовки клиенту

	// ── Kill previous stream for same session (prevents ghost goroutines) ──
	h.cancelPreviousSession(req.SessionID)

	// ── DETACHED generation context — survives SSE disconnect ──────────
	// The HTTP request context (r.Context()) gets cancelled when the proxy
	// kills the SSE stream. We must NOT derive the generation context from it,
	// otherwise the generation goroutine dies and no files are stored.
	genCtx, genCancel := context.WithTimeout(context.Background(), 25*time.Minute)
	backgroundDrainerActive := false
	defer func() {
		if !backgroundDrainerActive {
			genCancel() // cleanup if handler exits normally (result/error delivered via SSE)
		}
	}()

	// ── Session ID for checkpoint/resume ──
	if req.SessionID != "" {
		genCtx = application.ContextWithSessionID(genCtx, req.SessionID)
		h.registerSession(req.SessionID, genCancel)
		defer h.unregisterSession(req.SessionID)
		log.Printf("🔑 Session ID attached: %s (resume=%v)", req.SessionID, req.Resume)
	}

	// ── Запускаем генерацию в горутине с DETACHED context ─────────
	resultChan := make(chan *application.GenerationResult, 1)
	errorChan := make(chan error, 1)
	sessionID := req.SessionID // capture for goroutine

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("🔥 PANIC in generation goroutine: %v", rec)
				errorChan <- fmt.Errorf("internal panic: %v", rec)
			}
		}()
		mode := application.ModeCode
		if req.Mode == "agent" {
			mode = application.ModeAgent
		} else if req.Mode == "synthesis" {
			mode = application.ModeSynthesis
		}
		log.Printf("DEBUG: горутина запущена mode=%s", mode)
		result, err := h.orchestrator.GenerateWithMode(genCtx, req.Specification, req.URL, mode)
		if err != nil {
			log.Printf("ERROR: GenerateWithMode вернул ошибку: %v", err)
			if result != nil && len(result.Code) > 0 {
				log.Printf("⚠️ Partial result: %d files — storing before error", len(result.Code))
				// Store files even on error — polling client can recover them
				if sessionID != "" {
					globalFileStore.Store(sessionID, result.Code)
					globalFileStore.MarkComplete(sessionID)
				}
				resultChan <- result
				return
			}
			// Mark complete on error too (so polling doesn't wait forever)
			if sessionID != "" {
				globalFileStore.MarkComplete(sessionID)
			}
			errorChan <- err
			return
		}
		log.Printf("✅ GenerateWithMode completed: %d files, duration=%v", len(result.Code), result.Duration)
		// Store files BEFORE sending to channel (goroutine-safe even if handler exited)
		if sessionID != "" && len(result.Code) > 0 {
			globalFileStore.Store(sessionID, result.Code)
			globalFileStore.MarkComplete(sessionID)
			log.Printf("💾 Files stored + marked complete in goroutine for session %s", sessionID)
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

	// Keepalive ticker: sends SSE comment every 10s to prevent Railway/Vercel from closing idle connections
	heartbeat := time.NewTicker(10 * time.Second)
	defer heartbeat.Stop()

	// Слушаем статусы и результат
	for {
		select {
		case <-heartbeat.C:
			// SSE comment line — invisible to frontend, keeps TCP connection alive
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()

		case event := <-statusStream:
			// ── Partial Delivery: store file server-side SILENTLY (no SSE event) ──
			// Sending per-file events caused HTTP/2 stream death on Railway proxy.
			// Client fetches all files via GET /generate/files after done/disconnect.
			if event.Kind == "file" && event.Filename != "" && event.Content != "" {
				if req.SessionID != "" {
					globalFileStore.Append(req.SessionID, event.Filename, event.Content)
				}
				log.Printf("💾 Stored file '%s' (%d bytes) for session %s (silent, no SSE)",
					event.Filename, len(event.Content), req.SessionID)
				continue
			}

			// Drop reflection events entirely — they're huge and the client doesn't need them
			if event.Kind == "reflection" {
				continue
			}

			// General message truncation — enforce 500-byte SSE payload safety
			const maxSSEMessageBytes = 500
			msg := event.Message
			if event.Kind != "user_action" && len(msg) > maxSSEMessageBytes {
				msg = msg[:maxSSEMessageBytes] + "..."
			}

			// Отправляем событие агента
			payload := map[string]interface{}{
				"agent":     fmt.Sprintf("%s", event.Agent),
				"status":    string(event.Kind),
				"state":     string(event.State),
				"message":   msg,
				"progress":  event.Progress,
				"timestamp": event.Timestamp.Format(time.RFC3339),
			}
			// user_action: include draft_plan and session_id for approval UI
			if event.Kind == "user_action" {
				payload["draft_plan"] = event.DraftPlan
				payload["session_id"] = event.SessionID
			}
			// media_approval: include media_assets and session_id for design review modal
			if event.Kind == "media_approval" {
				payload["media_assets"] = event.MediaAssets
				payload["session_id"] = event.SessionID
				h.sendSSE(w, flusher, "media_approval", payload)
				log.Printf("🎨 SSE: media_approval event sent, %d assets for session %s", len(event.MediaAssets), event.SessionID)
				continue
			}
			// replan: include feedback and session_id, then close stream
			if event.Kind == "replan" {
				payload["feedback"] = event.Message
				payload["session_id"] = event.SessionID
				h.sendSSE(w, flusher, "replan", payload)
				log.Printf("🔄 SSE: replan event sent, closing stream for session %s", event.SessionID)
				return
			}
			h.sendSSE(w, flusher, string(event.Kind), payload)

		case result := <-resultChan:
			// Генерация завершена — store all files + mark complete
			log.Printf("📤 SSE: result ready, files=%d, duration=%v", len(result.Code), result.Duration)

			if req.SessionID != "" && len(result.Code) > 0 {
				globalFileStore.Store(req.SessionID, result.Code)
				globalFileStore.MarkComplete(req.SessionID)
				log.Printf("💾 Files stored + marked complete for session %s (%d files)", req.SessionID, len(result.Code))
			}

			// Tiny done event — no file content, no file list, just counts
			h.sendSSE(w, flusher, "done", map[string]interface{}{
				"message":    "✅ Проект успешно сгенерирован",
				"session_id": req.SessionID,
				"file_count": len(result.Code),
				"duration":   result.Duration.String(),
			})
			log.Printf("📤 SSE: done sent (files stored server-side), closing handler")
			return

		case err := <-errorChan:
			// Intercept replan_requested — send as replan event, not error
			if strings.Contains(err.Error(), "replan_requested") {
				log.Printf("🔄 SSE: replan_requested detected, sending replan event")
				h.sendSSE(w, flusher, "replan", map[string]interface{}{
					"message": "🔄 Перепланирование с учётом правок...",
				})
				return
			}
			// Ошибка генерации
			log.Printf("📤 SSE: sending error event: %v", err)
			h.sendSSE(w, flusher, "error", map[string]interface{}{
				"message": fmt.Sprintf("❌ Ошибка: %v", err),
			})
			return

		case <-r.Context().Done():
			// SSE disconnect (proxy killed connection) — generation continues in background!
			log.Printf("� SSE disconnected (proxy/client) for session %s — generation continues in background", sessionID)

			// Spawn background drainer: keeps consuming statusStream so orchestrator doesn't block
			backgroundDrainerActive = true
			go func() {
				defer genCancel()
				for {
					select {
					case event, ok := <-statusStream:
						if !ok {
							return
						}
						if event.Kind == "file" && event.Filename != "" && event.Content != "" {
							if sessionID != "" {
								globalFileStore.Append(sessionID, event.Filename, event.Content)
							}
						}
					case result, ok := <-resultChan:
						if !ok {
							return
						}
						if sessionID != "" && len(result.Code) > 0 {
							globalFileStore.Store(sessionID, result.Code)
							globalFileStore.MarkComplete(sessionID)
							log.Printf("💾 Background: stored %d files for session %s", len(result.Code), sessionID)
						}
						return
					case <-errorChan:
						if sessionID != "" {
							globalFileStore.MarkComplete(sessionID)
						}
						log.Printf("💾 Background: generation error for session %s, marked complete", sessionID)
						return
					case <-genCtx.Done():
						return
					}
				}
			}()
			return
		}
	}
}

// sendSSE отправляет SSE событие через Output Guardrails.
// Все string-поля payload-а проходят через PromptMaskFilter; при детекции
// утечки системного промпта/jailbreak-маркера поле подменяется на refusal.
func (h *GenerateHandlerSSE) sendSSE(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
	// ── Output Guardrails ──
	if h.guardrails != nil {
		switch payload := data.(type) {
		case map[string]interface{}:
			// "file"/"files_batch" содержат сгенерированный код пользователя — НЕ фильтруем,
			// иначе подменим валидный код на refusal. Фильтруем только статус-сообщения.
			if event != "file" && event != "files_batch" && event != "file_start" && event != "file_chunk" && event != "file_end" {
				sanitized, leakCount, reason := h.guardrails.SanitizeMap(payload)
				if leakCount > 0 {
					log.Printf("🛡️ Output Guardrail TRIGGERED: event=%s leaks=%d reason=%s",
						event, leakCount, reason)
				}
				data = sanitized
			} else {
				// Для file-событий проверяем только filename/имя, не content
				if name, ok := payload["name"].(string); ok {
					sanitized, leaked, reason := h.guardrails.Sanitize(name)
					if leaked {
						log.Printf("🛡️ Output Guardrail TRIGGERED on file name: reason=%s", reason)
						newPayload := make(map[string]interface{}, len(payload))
						for k, v := range payload {
							newPayload[k] = v
						}
						newPayload["name"] = sanitized
						data = newPayload
					}
				}
			}
		case string:
			sanitized, leaked, reason := h.guardrails.Sanitize(payload)
			if leaked {
				log.Printf("🛡️ Output Guardrail TRIGGERED: event=%s reason=%s", event, reason)
			}
			data = sanitized
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("ERROR: sendSSE json.Marshal failed for event '%s': %v", event, err)
		return
	}

	// Safety net: warn if serialized payload exceeds proxy-safe threshold
	if len(jsonData) > 4096 {
		log.Printf("⚠️ SSE OVERSIZED: event=%s payload=%d bytes (proxy limit ~4KB)", event, len(jsonData))
	}

	if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
		log.Printf("ERROR: sendSSE write failed for event '%s': %v", event, err)
		return
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", jsonData); err != nil {
		log.Printf("ERROR: sendSSE data write failed for event '%s': %v", event, err)
		return
	}
	flusher.Flush()
}
