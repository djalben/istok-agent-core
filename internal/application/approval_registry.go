package application

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/istok/agent-core/internal/domain"
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
func (r *ApprovalRegistry) Register(sessionID string) {
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
	log.Printf("🔒 ApprovalRegistry: registered wait channel for session %s", sessionID)
}

// WaitForApproval блокирует горутину до получения решения, таймаута или отмены контекста.
// Возвращает решение пользователя или ошибку (timeout/cancel).
// ГАРАНТИЯ: канал удаляется из реестра при любом исходе (no goroutine leak).
func (r *ApprovalRegistry) WaitForApproval(ctx context.Context, sessionID string) (ApprovalDecision, error) {
	r.mu.Lock()
	ch, exists := r.channels[sessionID]
	r.mu.Unlock()

	if !exists {
		return ApprovalDecision{}, fmt.Errorf("no approval channel for session %s", sessionID)
	}

	defer r.Cleanup(sessionID)

	timer := time.NewTimer(r.timeout)
	defer timer.Stop()

	select {
	case decision := <-ch:
		log.Printf("✅ ApprovalRegistry: received decision for session %s: approved=%v", sessionID, decision.Approved)
		return decision, nil
	case <-timer.C:
		log.Printf("⏱️ ApprovalRegistry: timeout (%v) for session %s", r.timeout, sessionID)
		return ApprovalDecision{}, fmt.Errorf("approval timeout (%v) for session %s", r.timeout, sessionID)
	case <-ctx.Done():
		log.Printf("🚫 ApprovalRegistry: context cancelled for session %s: %v", sessionID, ctx.Err())
		return ApprovalDecision{}, fmt.Errorf("approval cancelled: %w", ctx.Err())
	}
}

// Submit отправляет решение пользователя в ожидающую горутину.
// Возвращает ошибку если сессия не найдена (уже завершилась/таймаут).
func (r *ApprovalRegistry) Submit(sessionID string, decision ApprovalDecision) error {
	r.mu.Lock()
	ch, exists := r.channels[sessionID]
	r.mu.Unlock()

	if !exists {
		return fmt.Errorf("session %s not found or already resolved", sessionID)
	}

	select {
	case ch <- decision:
		log.Printf("📨 ApprovalRegistry: submitted decision for session %s", sessionID)
		return nil
	default:
		return fmt.Errorf("session %s channel full or closed", sessionID)
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
func (r *ApprovalRegistry) RegisterMedia(sessionID string) {
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
	log.Printf("🎨 ApprovalRegistry: registered media wait channel for session %s", sessionID)
}

// WaitForMediaApproval блокирует до решения пользователя по медиа-промптам.
func (r *ApprovalRegistry) WaitForMediaApproval(ctx context.Context, sessionID string) (MediaApprovalDecision, error) {
	r.mu.Lock()
	ch, exists := r.mediaChannels[sessionID]
	r.mu.Unlock()

	if !exists {
		return MediaApprovalDecision{}, fmt.Errorf("no media approval channel for session %s", sessionID)
	}

	defer r.CleanupMedia(sessionID)

	timer := time.NewTimer(r.timeout)
	defer timer.Stop()

	select {
	case decision := <-ch:
		log.Printf("✅ ApprovalRegistry: media decision for session %s: approved=%v, assets=%d", sessionID, decision.Approved, len(decision.Assets))
		return decision, nil
	case <-timer.C:
		log.Printf("⏱️ ApprovalRegistry: media timeout (%v) for session %s", r.timeout, sessionID)
		return MediaApprovalDecision{}, fmt.Errorf("media approval timeout (%v) for session %s", r.timeout, sessionID)
	case <-ctx.Done():
		log.Printf("🚫 ApprovalRegistry: media context cancelled for session %s: %v", sessionID, ctx.Err())
		return MediaApprovalDecision{}, fmt.Errorf("media approval cancelled: %w", ctx.Err())
	}
}

// SubmitMedia отправляет решение по медиа-промптам.
func (r *ApprovalRegistry) SubmitMedia(sessionID string, decision MediaApprovalDecision) error {
	r.mu.Lock()
	ch, exists := r.mediaChannels[sessionID]
	r.mu.Unlock()

	if !exists {
		return fmt.Errorf("media session %s not found or already resolved", sessionID)
	}

	select {
	case ch <- decision:
		log.Printf("📨 ApprovalRegistry: submitted media decision for session %s", sessionID)
		return nil
	default:
		return fmt.Errorf("media session %s channel full or closed", sessionID)
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
