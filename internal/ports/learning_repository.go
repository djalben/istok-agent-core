package ports

import "context"

type LearningRepository interface {
	SaveLearningContext(ctx context.Context, agentID string, context any) error
	LoadLearningContext(ctx context.Context, agentID string) (any, error)
	UpdateLearningContext(ctx context.Context, agentID string, context any) error
	DeleteLearningContext(ctx context.Context, agentID string) error
	GetLearningHistory(ctx context.Context, agentID string, limit int) ([]any, error)
}
