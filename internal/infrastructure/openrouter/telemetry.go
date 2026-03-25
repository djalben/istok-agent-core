package openrouter

import (
	"sync"
	"time"
)

type RequestMetrics struct {
	ModelID      string
	TotalCount   int64
	SuccessCount int64
	FailureCount int64
	TotalLatency time.Duration
	AvgLatency   time.Duration
	MinLatency   time.Duration
	MaxLatency   time.Duration
	LastRequest  time.Time
}

type FallbackMetrics struct {
	ModelID       string
	AttemptNumber int
	Success       bool
	Error         error
	Timestamp     time.Time
}

type Telemetry struct {
	requestMetrics  map[string]*RequestMetrics
	fallbackHistory []FallbackMetrics
	mutex           sync.RWMutex
	startTime       time.Time
}

func NewTelemetry() *Telemetry {
	return &Telemetry{
		requestMetrics:  make(map[string]*RequestMetrics),
		fallbackHistory: make([]FallbackMetrics, 0),
		startTime:       time.Now(),
	}
}

func (t *Telemetry) RecordRequest(modelID string, duration time.Duration, success bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	metrics, exists := t.requestMetrics[modelID]
	if !exists {
		metrics = &RequestMetrics{
			ModelID:    modelID,
			MinLatency: duration,
			MaxLatency: duration,
		}
		t.requestMetrics[modelID] = metrics
	}

	metrics.TotalCount++
	if success {
		metrics.SuccessCount++
	} else {
		metrics.FailureCount++
	}

	metrics.TotalLatency += duration
	metrics.AvgLatency = time.Duration(int64(metrics.TotalLatency) / metrics.TotalCount)
	metrics.LastRequest = time.Now()

	if duration < metrics.MinLatency {
		metrics.MinLatency = duration
	}
	if duration > metrics.MaxLatency {
		metrics.MaxLatency = duration
	}
}

func (t *Telemetry) RecordFallbackAttempt(modelID string, attemptNumber int, err error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.fallbackHistory = append(t.fallbackHistory, FallbackMetrics{
		ModelID:       modelID,
		AttemptNumber: attemptNumber,
		Success:       false,
		Error:         err,
		Timestamp:     time.Now(),
	})

	if len(t.fallbackHistory) > 1000 {
		t.fallbackHistory = t.fallbackHistory[len(t.fallbackHistory)-1000:]
	}
}

func (t *Telemetry) RecordFallbackSuccess(modelID string, attemptNumber int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.fallbackHistory = append(t.fallbackHistory, FallbackMetrics{
		ModelID:       modelID,
		AttemptNumber: attemptNumber,
		Success:       true,
		Timestamp:     time.Now(),
	})
}

func (t *Telemetry) GetModelMetrics(modelID string) *RequestMetrics {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if metrics, exists := t.requestMetrics[modelID]; exists {
		metricsCopy := *metrics
		return &metricsCopy
	}
	return nil
}

func (t *Telemetry) GetAllMetrics() map[string]*RequestMetrics {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	result := make(map[string]*RequestMetrics)
	for k, v := range t.requestMetrics {
		metricsCopy := *v
		result[k] = &metricsCopy
	}
	return result
}

func (t *Telemetry) GetFallbackHistory(limit int) []FallbackMetrics {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if limit <= 0 || limit > len(t.fallbackHistory) {
		limit = len(t.fallbackHistory)
	}

	start := len(t.fallbackHistory) - limit
	if start < 0 {
		start = 0
	}

	history := make([]FallbackMetrics, limit)
	copy(history, t.fallbackHistory[start:])
	return history
}

func (t *Telemetry) GetOverallStats() map[string]interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	totalRequests := int64(0)
	totalSuccess := int64(0)
	totalFailures := int64(0)

	for _, metrics := range t.requestMetrics {
		totalRequests += metrics.TotalCount
		totalSuccess += metrics.SuccessCount
		totalFailures += metrics.FailureCount
	}

	successRate := 0.0
	if totalRequests > 0 {
		successRate = float64(totalSuccess) / float64(totalRequests)
	}

	return map[string]interface{}{
		"total_requests":  totalRequests,
		"total_success":   totalSuccess,
		"total_failures":  totalFailures,
		"success_rate":    successRate,
		"uptime_seconds":  time.Since(t.startTime).Seconds(),
		"models_tracked":  len(t.requestMetrics),
		"fallback_events": len(t.fallbackHistory),
	}
}

func (t *Telemetry) Reset() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.requestMetrics = make(map[string]*RequestMetrics)
	t.fallbackHistory = make([]FallbackMetrics, 0)
	t.startTime = time.Now()
}
