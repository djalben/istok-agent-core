package application

import (
	"context"
	"testing"
	"time"

	"github.com/djalben/istok-agent-core/internal/domain"
)

// TestEventBus_SessionIsolation проверяет, что события двух параллельных
// сессий НЕ пересекаются: подписчик сессии A видит только события A, B — только B.
// Регрессионный тест на P0-баг «единый канал EventBus → cross-session leakage».
func TestEventBus_SessionIsolation(t *testing.T) {
	orch := NewOrchestrator(newMockLLM())

	chA := orch.SubscribeSession("sess-A")
	defer orch.ReleaseSession("sess-A")
	chB := orch.SubscribeSession("sess-B")
	defer orch.ReleaseSession("sess-B")

	ctxA := ContextWithSessionID(context.Background(), "sess-A")
	ctxB := ContextWithSessionID(context.Background(), "sess-B")

	// Публикуем файлы в обе сессии вперемешку.
	o := orch
	o.busFromCtx(ctxA).PublishFile(domain.RoleCoder, "a.tsx", "AAA")
	o.busFromCtx(ctxB).PublishFile(domain.RoleCoder, "b.tsx", "BBB")
	o.busFromCtx(ctxA).PublishStatus(domain.RoleCoder, "", "statusA", 50)
	o.busFromCtx(ctxB).PublishStatus(domain.RoleCoder, "", "statusB", 50)

	gotA := drain(t, chA, 2)
	gotB := drain(t, chB, 2)

	for _, e := range gotA {
		if e.Filename == "b.tsx" || e.Message == "statusB" {
			t.Fatalf("session A received session B event: %+v", e)
		}
	}
	for _, e := range gotB {
		if e.Filename == "a.tsx" || e.Message == "statusA" {
			t.Fatalf("session B received session A event: %+v", e)
		}
	}
	if len(gotA) != 2 {
		t.Fatalf("session A: expected 2 events, got %d", len(gotA))
	}
	if len(gotB) != 2 {
		t.Fatalf("session B: expected 2 events, got %d", len(gotB))
	}
}

// drain собирает до n событий из канала с таймаутом.
func drain(t *testing.T, ch <-chan domain.AgentEvent, n int) []domain.AgentEvent {
	t.Helper()
	out := make([]domain.AgentEvent, 0, n)
	deadline := time.After(2 * time.Second)
	for len(out) < n {
		select {
		case e := <-ch:
			out = append(out, e)
		case <-deadline:
			return out
		}
	}
	return out
}
