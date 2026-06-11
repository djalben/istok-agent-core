package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/djalben/istok-agent-core/internal/application"
	"github.com/djalben/istok-agent-core/internal/application/dto"
	"github.com/djalben/istok-agent-core/internal/domain"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type sseStreamSession struct {
	h                       *GenerateHandlerSSE
	w                       http.ResponseWriter
	flusher                 http.Flusher
	r                       *http.Request
	req                     dto.GenerateProjectRequest
	ownerID                 string
	genCtx                  context.Context
	genCancel               context.CancelFunc
	releaseSlot             func()
	sessionID               string
	statusStream            <-chan domain.AgentEvent
	resultChan              chan *application.GenerationResult
	errorChan               chan error
	backgroundDrainerActive bool
}

func (h *GenerateHandlerSSE) beginSSEStream(w http.ResponseWriter, r *http.Request, req dto.GenerateProjectRequest, ownerID string, releaseSlot func()) (*sseStreamSession, bool) {
	ctx := r.Context()
	flusher, ok := w.(http.Flusher)
	if !ok {
		sseLog(ctx).ErrorContext(ctx, "response writer does not support flusher", "type", fmt.Sprintf("%T", w))
		writeError(w, http.StatusInternalServerError, "SSE не поддерживается")

		return nil, false
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Encoding", "identity")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	h.cancelPreviousSession(ctx, req.SessionID)

	// КРИТИЧНО: контекст генерации ОТВЯЗАН от r.Context(). Railway Envoy убивает
	// долгоживущие SSE-стримы через ~120с, а навигация/закрытие вкладки отменяет
	// r.Context() — без отвязки это каскадно отменяло genCtx и роняло LLM-вызов
	// Кодера ("context canceled"). context.WithoutCancel сохраняет значения
	// (logger, requestID), но НЕ наследует отмену. Единственные источники отмены —
	// 30-минутный таймаут и явный genCancel (вызывается по завершении генерации
	// или background-дренера). Отключение клиента ловится отдельно в runEventLoop
	// через r.Context().Done() и запускает фоновый дренер.
	genCtx, genCancel := context.WithTimeout(context.WithoutCancel(r.Context()), 30*time.Minute)
	s := &sseStreamSession{
		h: h, w: w, flusher: flusher, r: r, req: req, ownerID: ownerID,
		genCtx: genCtx, genCancel: genCancel, releaseSlot: releaseSlot, sessionID: req.SessionID,
		resultChan: make(chan *application.GenerationResult, 1),
		errorChan:  make(chan error, 1),
	}

	if req.SessionID != "" {
		genCtx = application.ContextWithSessionID(genCtx, req.SessionID)
		s.genCtx = genCtx
		h.registerSession(req.SessionID, genCancel)
		sseLog(genCtx).InfoContext(
			genCtx, "session attached",
			"sessionId", req.SessionID,
			"resume", req.Resume,
		)
	}

	// Feature gate: промо-видео генерируется только по явному запросу. true → Videographer
	// выполняется до Кодера (последовательно); false → видео пропускается (быстрый прототип).
	genCtx = application.ContextWithGenerateVideo(genCtx, req.GenerateVideo)
	s.genCtx = genCtx

	s.statusStream = h.orchestrator.SubscribeSession(req.SessionID)
	s.startGenerationGoroutine(genCtx, true)

	h.sendSSE(genCtx, w, flusher, "status", map[string]any{
		"agent": "system", "status": "started",
		"message": "🚀 Запуск S-Tier AI Orchestrator...", "progress": 0,
	})

	return s, true
}

func (s *sseStreamSession) startGenerationGoroutine(ctx context.Context, holdSlot bool) {
	if holdSlot {
		go func(runCtx context.Context) {
			defer s.releaseSlot()
			defer s.h.orchestrator.ReleaseSession(s.sessionID)
			s.runGenerationWithRecovery(runCtx)
		}(ctx)

		return
	}
	go s.runGenerationWithRecovery(ctx)
}

func (s *sseStreamSession) runGenerationWithRecovery(ctx context.Context) {
	defer func() {
		if s.sessionID != "" {
			globalFileStore.MarkComplete(s.sessionID)
			sseLog(ctx).InfoContext(ctx, "generation goroutine exited", "sessionId", s.sessionID)
		}
	}()
	defer func() {
		if rec := recover(); rec != nil {
			sseLog(ctx).ErrorContext(ctx, "panic in generation goroutine", "panic", rec)
			s.errorChan <- wrapper.Wrapf(ErrInternalPanic, "%v", rec)
		}
	}()

	mode := application.ModeCode
	switch s.req.Mode {
	case "agent":
		mode = application.ModeAgent
	case "synthesis":
		mode = application.ModeSynthesis
	}
	sseLog(ctx).InfoContext(ctx, "generation goroutine started", "mode", mode)
	result, err := s.h.orchestrator.GenerateWithMode(ctx, s.req.Specification, s.req.URL, mode)
	if err != nil {
		sseLog(ctx).ErrorContext(ctx, "GenerateWithMode failed", "error", wrapper.Wrap(err))
		if result != nil && len(result.Code) > 0 {
			sseLog(ctx).WarnContext(ctx, "partial result before error", "files", len(result.Code))
			if s.sessionID != "" {
				globalFileStore.Store(s.sessionID, result.Code)
				globalFileStore.MarkComplete(s.sessionID)
			}
			s.resultChan <- result

			return
		}
		if s.sessionID != "" {
			globalFileStore.MarkComplete(s.sessionID)
		}
		s.errorChan <- err

		return
	}
	sseLog(ctx).InfoContext(
		ctx, "GenerateWithMode completed",
		"files", len(result.Code),
		"duration", result.Duration,
	)
	if s.sessionID != "" && len(result.Code) > 0 {
		globalFileStore.Store(s.sessionID, result.Code)
		globalFileStore.MarkComplete(s.sessionID)
		sseLog(ctx).InfoContext(ctx, "files stored in goroutine", "sessionId", s.sessionID, "files", len(result.Code))
	}
	s.resultChan <- result
}

func (s *sseStreamSession) runEventLoop() {
	heartbeat := time.NewTicker(10 * time.Second)
	defer heartbeat.Stop()
	defer func() {
		if !s.backgroundDrainerActive {
			s.genCancel()
		}
	}()
	if s.sessionID != "" {
		defer s.h.unregisterSession(s.sessionID)
	}

	for {
		select {
		case <-heartbeat.C:
			fmt.Fprintf(s.w, ": heartbeat\n\n")
			s.flusher.Flush()
		case event := <-s.statusStream:
			if s.handleStatusEvent(event) {
				return
			}
		case result := <-s.resultChan:
			s.handleResult(result)

			return
		case err := <-s.errorChan:
			s.handleError(err)

			return
		case <-s.r.Context().Done():
			s.onClientDisconnect()

			return
		}
	}
}

func (s *sseStreamSession) handleStatusEvent(event domain.AgentEvent) bool {
	ctx := s.genCtx
	if event.Kind == domain.EventFile && event.Filename != "" && event.Content != "" {
		if s.req.SessionID != "" {
			globalFileStore.Append(s.req.SessionID, event.Filename, event.Content)
		}
		sseLog(ctx).DebugContext(ctx, "file stored silently", "file", event.Filename)

		return false
	}
	if event.Kind == "reflection" {
		return false
	}

	const maxSSEMessageBytes = 500
	msg := event.Message
	if event.Kind != domain.EventUserAction && len(msg) > maxSSEMessageBytes {
		msg = msg[:maxSSEMessageBytes] + "..."
	}

	payload := map[string]any{
		"agent": string(event.Agent), "status": string(event.Kind), "state": string(event.State),
		"message": msg, "progress": event.Progress, "timestamp": event.Timestamp.Format(time.RFC3339),
	}
	if event.Kind == domain.EventUserAction {
		payload["draft_plan"] = event.DraftPlan
		payload["session_id"] = event.SessionID
	}
	if event.Kind == "media_approval" {
		payload["media_assets"] = event.MediaAssets
		payload["session_id"] = event.SessionID
		s.h.sendSSE(ctx, s.w, s.flusher, "media_approval", payload)
		sseLog(ctx).InfoContext(
			ctx, "sse media_approval sent",
			"assets", len(event.MediaAssets),
			"sessionIdLen", len(event.SessionID),
		)

		return false
	}
	if event.Kind == "insufficient_funds" {
		payload["session_id"] = event.SessionID
		s.h.sendSSE(ctx, s.w, s.flusher, "insufficient_funds", payload)
		sseLog(ctx).InfoContext(ctx, "sse insufficient_funds sent", "sessionIdLen", len(event.SessionID))

		return false
	}
	if event.Kind == "replan" {
		payload["feedback"] = event.Message
		payload["session_id"] = event.SessionID
		s.h.sendSSE(ctx, s.w, s.flusher, "replan", payload)
		sseLog(ctx).InfoContext(ctx, "sse replan sent, closing stream", "sessionIdLen", len(event.SessionID))

		return true
	}

	s.h.sendSSE(ctx, s.w, s.flusher, string(event.Kind), payload)

	return false
}

func (s *sseStreamSession) handleResult(result *application.GenerationResult) {
	ctx := s.genCtx
	sseLog(ctx).InfoContext(
		ctx, "sse result ready",
		"files", len(result.Code),
		"duration", result.Duration,
	)
	if s.req.SessionID != "" && len(result.Code) > 0 {
		globalFileStore.Store(s.req.SessionID, result.Code)
		globalFileStore.MarkComplete(s.req.SessionID)
		sseLog(ctx).InfoContext(ctx, "files stored before done event", "sessionId", s.req.SessionID)
	}
	savedID, saveErr := s.h.persistGenerated(ctx, s.ownerID, s.req, result.Code)
	if saveErr != nil {
		sseLog(ctx).WarnContext(ctx, "db auto-save failed", "error", wrapper.Wrap(saveErr))
	} else if savedID != "" {
		sseLog(ctx).InfoContext(ctx, "project persisted to db", "projectId", savedID)
	}
	s.h.sendSSE(ctx, s.w, s.flusher, "done", map[string]any{
		"message": "✅ Проект успешно сгенерирован", "session_id": s.req.SessionID,
		"project_id": savedID, "file_count": len(result.Code), "duration": result.Duration.String(),
	})
	sseLog(ctx).InfoContext(ctx, "sse done sent, closing handler")
}

func (s *sseStreamSession) handleError(err error) {
	ctx := s.genCtx
	if strings.Contains(err.Error(), "replan_requested") {
		sseLog(ctx).InfoContext(ctx, "replan_requested, sending replan event")
		s.h.sendSSE(ctx, s.w, s.flusher, "replan", map[string]any{"message": "🔄 Перепланирование с учётом правок..."})

		return
	}
	sseLog(ctx).ErrorContext(ctx, "sse sending error event", "error", wrapper.Wrap(err))
	s.h.sendSSE(ctx, s.w, s.flusher, "error", map[string]any{"message": fmt.Sprintf("❌ Ошибка: %v", err)})
}

func (s *sseStreamSession) onClientDisconnect() {
	sseLog(s.genCtx).InfoContext(s.genCtx, "sse client disconnected, generation continues in background")
	s.h.spawnAutoApproveOnDisconnect(s.genCtx, s.sessionID)
	s.backgroundDrainerActive = true
	go s.runBackgroundDrainer()
}

func (s *sseStreamSession) runBackgroundDrainer() {
	ctx := s.genCtx
	defer s.genCancel()
	for {
		select {
		case event, ok := <-s.statusStream:
			if !ok {
				return
			}
			s.drainStatusEvent(event)
		case result, ok := <-s.resultChan:
			if !ok {
				return
			}
			if s.sessionID != "" && len(result.Code) > 0 {
				globalFileStore.Store(s.sessionID, result.Code)
				globalFileStore.MarkComplete(s.sessionID)
				sseLog(ctx).InfoContext(ctx, "background stored files", "files", len(result.Code))
			}
			savedID, saveErr := s.h.persistGenerated(ctx, s.ownerID, s.req, result.Code)
			if saveErr != nil {
				sseLog(ctx).WarnContext(ctx, "background db auto-save failed", "error", wrapper.Wrap(saveErr))
			} else if savedID != "" {
				sseLog(ctx).InfoContext(ctx, "background project persisted", "projectId", savedID)
			}

			return
		case <-s.errorChan:
			if s.sessionID != "" {
				globalFileStore.MarkComplete(s.sessionID)
			}
			sseLog(ctx).WarnContext(ctx, "background generation error, marked complete")

			return
		case <-ctx.Done():
			if s.sessionID != "" {
				globalFileStore.MarkComplete(s.sessionID)
			}

			return
		}
	}
}

func (s *sseStreamSession) drainStatusEvent(event domain.AgentEvent) {
	if s.sessionID == "" {
		return
	}
	switch event.Kind {
	case domain.EventFile:
		if event.Filename != "" && event.Content != "" {
			globalFileStore.Append(s.sessionID, event.Filename, event.Content)
		}
	case domain.EventUserAction:
		globalFileStore.SetPendingAction(s.sessionID, &PendingAction{
			Kind: string(domain.EventUserAction), DraftPlan: event.DraftPlan, SessionID: event.SessionID,
		})
		if event.Message != "" {
			globalFileStore.UpdateStatus(s.sessionID, event.Message)
		}
	case domain.EventMediaApproval:
		globalFileStore.SetPendingAction(s.sessionID, &PendingAction{
			Kind: string(domain.EventMediaApproval), Assets: event.MediaAssets, SessionID: event.SessionID,
		})
		if event.Message != "" {
			globalFileStore.UpdateStatus(s.sessionID, event.Message)
		}
	case domain.EventInsufficientFunds:
		globalFileStore.SetPendingAction(s.sessionID, &PendingAction{
			Kind: string(domain.EventInsufficientFunds), SessionID: event.SessionID,
		})
		if event.Message != "" {
			globalFileStore.UpdateStatus(s.sessionID, event.Message)
		}
	case domain.EventStatus, domain.EventError:
		if event.Message != "" {
			globalFileStore.UpdateStatus(s.sessionID, event.Message)
		}
		globalFileStore.ClearPendingAction(s.sessionID)
	case domain.EventFSM, domain.EventPlan, domain.EventDone, domain.EventReflection, domain.EventReplan:
		// handled elsewhere or no file-store side effects
	default:
		// reflection and unknown kinds — no-op
	}
}
