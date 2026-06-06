package application

import (
	"context"
	"sync"
	"time"

	"gitlab.com/libs-artifex/wrapper/v2"
)

// FundsRegistry — thread-safe registry of wait channels for "insufficient funds" pauses.
// When generation hits ErrInsufficientFunds, the orchestrator blocks on WaitForFunds.
// The user clicks "Resume" on the frontend → POST /generate/resume_funds → Resume() unblocks.
type FundsRegistry struct {
	mu       sync.Mutex
	channels map[string]chan struct{}
	timeout  time.Duration
}

// NewFundsRegistry creates the registry with a maximum wait duration.
func NewFundsRegistry(timeout time.Duration) *FundsRegistry {
	if timeout <= 0 {
		timeout = 2 * time.Hour
	}

	return &FundsRegistry{
		channels: make(map[string]chan struct{}),
		timeout:  timeout,
	}
}

// Register creates a wait channel for the given session.
func (r *FundsRegistry) Register(ctx context.Context, sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if old, exists := r.channels[sessionID]; exists {
		select {
		case <-old:
		default:
			close(old)
		}
	}
	r.channels[sessionID] = make(chan struct{}, 1)
	applog(ctx).InfoContext(ctx, "funds wait channel registered", "sessionId", sessionID)
}

// WaitForFunds blocks the generation goroutine until Resume is called, timeout, or context cancel.
func (r *FundsRegistry) WaitForFunds(ctx context.Context, sessionID string) error {
	r.mu.Lock()
	ch, exists := r.channels[sessionID]
	r.mu.Unlock()

	if !exists {
		return wrapper.Wrapf(ErrNoFundsWaitChannel, "%s", sessionID)
	}

	defer r.cleanup(sessionID)

	timer := time.NewTimer(r.timeout)
	defer timer.Stop()

	select {
	case <-ch:
		applog(ctx).InfoContext(ctx, "funds wait resumed", "sessionId", sessionID)

		return nil
	case <-timer.C:
		applog(ctx).WarnContext(ctx, "funds wait timeout",
			"sessionId", sessionID,
			"timeout", r.timeout,
		)

		return wrapper.Wrapf(ErrFundsWaitTimeout, "(%v) for session %s", r.timeout, sessionID)
	case <-ctx.Done():
		applog(ctx).WarnContext(ctx, "funds wait cancelled",
			"sessionId", sessionID,
			"error", wrapper.Wrap(ctx.Err()),
		)

		return wrapper.Wrapf(ErrFundsWaitCancelled, "%v", ctx.Err())
	}
}

// Resume unblocks the generation goroutine waiting on WaitForFunds.
func (r *FundsRegistry) Resume(ctx context.Context, sessionID string) error {
	r.mu.Lock()
	ch, exists := r.channels[sessionID]
	r.mu.Unlock()

	if !exists {
		return wrapper.Wrapf(ErrFundsSessionNotFound, "%s", sessionID)
	}

	select {
	case ch <- struct{}{}:
		applog(ctx).InfoContext(ctx, "funds resume signal sent", "sessionId", sessionID)

		return nil
	default:
		return wrapper.Wrapf(ErrFundsChannelClosed, "%s", sessionID)
	}
}

func (r *FundsRegistry) cleanup(sessionID string) {
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
}
