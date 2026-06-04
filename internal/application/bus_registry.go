package application

import (
	"sync"

	"github.com/djalben/istok-agent-core/internal/domain"
)

// busRegistry хранит отдельную шину событий на каждую активную сессию генерации.
// Изолирует параллельные генерации: события сессии A не утекают подписчику сессии B.
// Сессии без sessionID (curl, тесты) используют общий defaultBus.
type busRegistry struct {
	mu         sync.RWMutex
	buses      map[string]*domain.EventBus
	bufferSize int
}

func newBusRegistry(bufferSize int) *busRegistry {
	if bufferSize < 1 {
		bufferSize = 128
	}

	return &busRegistry{
		buses:      make(map[string]*domain.EventBus),
		bufferSize: bufferSize,
	}
}

// acquire создаёт (или возвращает существующую) шину для сессии.
func (r *busRegistry) acquire(sessionID string) *domain.EventBus {
	r.mu.Lock()
	defer r.mu.Unlock()
	bus, ok := r.buses[sessionID]
	if !ok {
		bus = domain.NewEventBus(r.bufferSize)
		r.buses[sessionID] = bus
	}

	return bus
}

// get возвращает шину сессии или nil, если она не зарегистрирована.
func (r *busRegistry) get(sessionID string) *domain.EventBus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.buses[sessionID]
}

// release удаляет шину сессии из реестра.
// НЕ закрывает канал: фоновая генерация (при SSE-дисконнекте) может ещё публиковать,
// а запись в закрытый канал паникует. Канал собирается GC после потери ссылок.
func (r *busRegistry) release(sessionID string) {
	r.mu.Lock()
	delete(r.buses, sessionID)
	r.mu.Unlock()
}
