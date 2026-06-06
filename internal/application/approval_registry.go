package application

import (
	"context"
	"sync"
	"time"

	"github.com/djalben/istok-agent-core/internal/domain"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// ApprovalDecision — ответ пользователя на запрос утверждения.
type ApprovalDecision struct {
	Approved bool   `json:"approved"`
	Feedback string `json:"feedback,omitempty"` // опциональный комментарий
}

// MediaApprovalDecision — решение пользователя по медиа-ассетам (дизайн-ревью).
type MediaApprovalDecision struct {
	Approved bool                `json:"approved"`
	Assets   []domain.MediaAsset `json:"assets"` // утверждённые/отредактированные ассеты
}

// ApprovalRegistry — потокобезопасный реестр каналов ожидания решений пользователя.
// Каждая сессия генерации может заблокироваться ровно один раз, ожидая ApprovalDecision.
// Защита от утечек горутин: WaitForApproval использует select с ctx.Done() и таймаутом.
type ApprovalRegistry struct {
	mu            sync.Mutex
	channels      map[string]chan ApprovalDecision
	mediaChannels map[string]chan MediaApprovalDecision
	timeout       time.Duration
}

// NewApprovalRegistry создаёт реестр с указанным максимальным временем ожидания.
func NewApprovalRegistry(timeout time.Duration) *ApprovalRegistry {
	if timeout <= 0 {
		timeout = 1 * time.Hour
	}

	return &ApprovalRegistry{
		channels:      make(map[string]chan ApprovalDecision),
		mediaChannels: make(map[string]chan MediaApprovalDecision),
		timeout:       timeout,
	}
}

// Register создаёт канал ожидания для сессии. Если канал уже существует — перезаписывает.
func (r *ApprovalRegistry) Register(ctx context.Context, sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Закрываем старый канал если есть (safety)
	if old, exists := r.channels[sessionID]; exists {
		select {
		case <-old:
		default:
			close(old)
		}
	}
	r.channels[sessionID] = make(chan ApprovalDecision, 1)
	applog(ctx).InfoContext(ctx, "approval wait channel registered", "sessionId", sessionID)
}

// WaitForApproval блокирует горутину до получения решения, таймаута или отмены контекста.
// Возвращает решение пользователя или ошибку (timeout/cancel).
// ГАРАНТИЯ: канал удаляется из реестра при любом исходе (no goroutine leak).
func (r *ApprovalRegistry) WaitForApproval(ctx context.Context, sessionID string) (ApprovalDecision, error) {
	r.mu.Lock()
	ch, exists := r.channels[sessionID]
	r.mu.Unlock()

	if !exists {
		return ApprovalDecision{}, wrapper.Wrapf(ErrNoApprovalChannel, "%s", sessionID)
	}

	defer r.Cleanup(sessionID)

	timer := time.NewTimer(r.timeout)
	defer timer.Stop()

	select {
	case decision := <-ch:
		applog(ctx).InfoContext(ctx, "approval decision received",
			"sessionId", sessionID,
			"approved", decision.Approved,
		)

		return decision, nil
	case <-timer.C:
		applog(ctx).WarnContext(ctx, "approval wait timeout",
			"sessionId", sessionID,
			"timeout", r.timeout,
		)

		return ApprovalDecision{}, wrapper.Wrapf(ErrApprovalTimeout, "(%v) for session %s", r.timeout, sessionID)
	case <-ctx.Done():
		applog(ctx).WarnContext(ctx, "approval wait cancelled",
			"sessionId", sessionID,
			"error", wrapper.Wrap(ctx.Err()),
		)

		return ApprovalDecision{}, wrapper.Wrapf(ErrApprovalCancelled, "%v", ctx.Err())
	}
}

// Submit отправляет решение пользователя в ожидающую горутину.
// Возвращает ошибку если сессия не найдена (уже завершилась/таймаут).
func (r *ApprovalRegistry) Submit(ctx context.Context, sessionID string, decision ApprovalDecision) error {
	r.mu.Lock()
	ch, exists := r.channels[sessionID]
	r.mu.Unlock()

	if !exists {
		return wrapper.Wrapf(ErrApprovalSessionNotFound, "%s", sessionID)
	}

	select {
	case ch <- decision:
		applog(ctx).InfoContext(ctx, "approval decision submitted",
			"sessionId", sessionID,
			"approved", decision.Approved,
		)

		return nil
	default:
		return wrapper.Wrapf(ErrApprovalChannelClosed, "%s", sessionID)
	}
}

// Cleanup удаляет канал из реестра (exported для использования transport layer при отмене сессии).
func (r *ApprovalRegistry) Cleanup(sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ch, exists := r.channels[sessionID]; exists {
		select {
		case <-ch:
		default:
			close(ch)
		}
	}
	delete(r.channels, sessionID)
	if ch, exists := r.mediaChannels[sessionID]; exists {
		select {
		case <-ch:
		default:
			close(ch)
		}
	}
	delete(r.mediaChannels, sessionID)
}

// ── Media Approval (Design Review) ──────────────────────────────────

// RegisterMedia создаёт канал ожидания медиа-решения для сессии.
func (r *ApprovalRegistry) RegisterMedia(ctx context.Context, sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if old, exists := r.mediaChannels[sessionID]; exists {
		select {
		case <-old:
		default:
			close(old)
		}
	}
	r.mediaChannels[sessionID] = make(chan MediaApprovalDecision, 1)
	applog(ctx).InfoContext(ctx, "media approval wait channel registered", "sessionId", sessionID)
}

// WaitForMediaApproval блокирует до решения пользователя по медиа-промптам.
func (r *ApprovalRegistry) WaitForMediaApproval(ctx context.Context, sessionID string) (MediaApprovalDecision, error) {
	r.mu.Lock()
	ch, exists := r.mediaChannels[sessionID]
	r.mu.Unlock()

	if !exists {
		return MediaApprovalDecision{}, wrapper.Wrapf(ErrNoMediaApprovalChannel, "%s", sessionID)
	}

	defer r.CleanupMedia(sessionID)

	timer := time.NewTimer(r.timeout)
	defer timer.Stop()

	select {
	case decision := <-ch:
		applog(ctx).InfoContext(ctx, "media approval decision received",
			"sessionId", sessionID,
			"approved", decision.Approved,
			"assets", len(decision.Assets),
		)

		return decision, nil
	case <-timer.C:
		applog(ctx).WarnContext(ctx, "media approval wait timeout",
			"sessionId", sessionID,
			"timeout", r.timeout,
		)

		return MediaApprovalDecision{}, wrapper.Wrapf(ErrMediaApprovalTimeout, "(%v) for session %s", r.timeout, sessionID)
	case <-ctx.Done():
		applog(ctx).WarnContext(ctx, "media approval wait cancelled",
			"sessionId", sessionID,
			"error", wrapper.Wrap(ctx.Err()),
		)

		return MediaApprovalDecision{}, wrapper.Wrapf(ErrMediaApprovalCancelled, "%v", ctx.Err())
	}
}

// SubmitMedia отправляет решение по медиа-промптам.
func (r *ApprovalRegistry) SubmitMedia(ctx context.Context, sessionID string, decision MediaApprovalDecision) error {
	r.mu.Lock()
	ch, exists := r.mediaChannels[sessionID]
	r.mu.Unlock()

	if !exists {
		return wrapper.Wrapf(ErrMediaSessionNotFound, "%s", sessionID)
	}

	select {
	case ch <- decision:
		applog(ctx).InfoContext(ctx, "media approval decision submitted",
			"sessionId", sessionID,
			"approved", decision.Approved,
			"assets", len(decision.Assets),
		)

		return nil
	default:
		return wrapper.Wrapf(ErrMediaChannelClosed, "%s", sessionID)
	}
}

// CleanupMedia удаляет медиа-канал из реестра.
func (r *ApprovalRegistry) CleanupMedia(sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ch, exists := r.mediaChannels[sessionID]; exists {
		select {
		case <-ch:
		default:
			close(ch)
		}
	}
	delete(r.mediaChannels, sessionID)
}
