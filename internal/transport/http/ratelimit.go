package http

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// genLimiter защищает дорогой LLM-пайплайн от cost-DoS:
//   - семафор ограничивает число ОДНОВРЕМЕННЫХ генераций (env MAX_CONCURRENT_GENERATIONS);
//   - per-IP окно ограничивает частоту запусков (env GEN_RATE_PER_MIN).
type genLimiter struct {
	sem chan struct{}

	mu       sync.Mutex
	hits     map[string][]time.Time
	perIP    int
	window   time.Duration
	lastSeen time.Time
}

func newGenLimiter() *genLimiter {
	maxConcurrent := max(envInt("MAX_CONCURRENT_GENERATIONS", 3), 1)
	perIP := envInt("GEN_RATE_PER_MIN", 10)

	return &genLimiter{
		sem:    make(chan struct{}, maxConcurrent),
		hits:   make(map[string][]time.Time),
		perIP:  perIP,
		window: time.Minute,
	}
}

// allowIP возвращает false, если IP превысил частотный лимит за окно.
func (l *genLimiter) allowIP(ip string) bool {
	if l.perIP <= 0 {
		return true
	}
	now := time.Now()
	cutoff := now.Add(-l.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	// Периодическая очистка мапы от старых IP (раз в окно).
	if now.Sub(l.lastSeen) > l.window {
		for k, ts := range l.hits {
			if len(ts) == 0 || ts[len(ts)-1].Before(cutoff) {
				delete(l.hits, k)
			}
		}
		l.lastSeen = now
	}

	recent := make([]time.Time, 0, len(l.hits[ip]))
	for _, t := range l.hits[ip] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	if len(recent) >= l.perIP {
		l.hits[ip] = recent

		return false
	}
	l.hits[ip] = append(recent, now)

	return true
}

// acquire пытается занять слот конкурентности без блокировки.
// Возвращает release-функцию и true при успехе; (nil, false) если все слоты заняты.
func (l *genLimiter) acquire() (func(), bool) {
	select {
	case l.sem <- struct{}{}:
		var once sync.Once

		return func() { once.Do(func() { <-l.sem }) }, true
	default:
		return nil, false
	}
}

// clientIP извлекает IP клиента с учётом прокси (Railway/Vercel).
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")

		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

func envInt(key string, def int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return def
	}

	return v
}
