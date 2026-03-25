package ports

import "context"

type OrchestratorPort interface {
	ScheduleTask(ctx context.Context, task Task) error
	CancelTask(ctx context.Context, taskID string) error
	GetTaskStatus(ctx context.Context, taskID string) (*TaskStatus, error)
	CoordinateAgents(ctx context.Context, agents []string, workflow Workflow) error
}

type Task struct {
	ID          string
	Type        string
	Priority    int
	Payload     map[string]interface{}
	Dependencies []string
}

type TaskStatus struct {
	ID          string
	State       string
	Progress    float64
	Result      map[string]interface{}
	Error       string
}

type Workflow struct {
	ID    string
	Steps []WorkflowStep
}

type WorkflowStep struct {
	Name     string
	AgentID  string
	Action   string
	Inputs   map[string]interface{}
	Outputs  []string
}
