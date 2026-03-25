package ports

import "context"

type LearningRepository interface {
	SaveLearningContext(ctx context.Context, agentID string, context interface{}) error
	LoadLearningContext(ctx context.Context, agentID string) (interface{}, error)
	UpdateLearningContext(ctx context.Context, agentID string, context interface{}) error
	DeleteLearningContext(ctx context.Context, agentID string) error
	GetLearningHistory(ctx context.Context, agentID string, limit int) ([]interface{}, error)
}
