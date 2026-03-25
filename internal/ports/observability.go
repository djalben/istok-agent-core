package ports

import (
	"context"
	"time"
)

type ObservabilityPort interface {
	RecordMetric(ctx context.Context, name string, value float64, tags map[string]string) error
	RecordTrace(ctx context.Context, operation string, duration time.Duration, metadata map[string]interface{}) error
	RecordLog(ctx context.Context, level LogLevel, message string, fields map[string]interface{}) error
	RecordEvent(ctx context.Context, eventType string, payload map[string]interface{}) error
}

type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
)
