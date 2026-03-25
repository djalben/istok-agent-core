package domain

import (
	"time"
)

type AgentStatus string

const (
	StatusIdle      AgentStatus = "idle"
	StatusAnalyzing AgentStatus = "analyzing"
	StatusCoding    AgentStatus = "coding"
	StatusDeploying AgentStatus = "deploying"
	StatusError     AgentStatus = "error"
)

type Agent struct {
	ID                 string
	Name               string
	Status             AgentStatus
	TokenBalance       int64
	LearningContext    *LearningContext
	TaskQueue          []*Task
	PerformanceMetrics *PerformanceMetrics
	Capabilities       map[string]*Capability
	DecisionHistory    []*DecisionRecord
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewAgent(id, name string, initialTokens int64) *Agent {
	now := time.Now()
	return &Agent{
		ID:              id,
		Name:            name,
		Status:          StatusIdle,
		TokenBalance:    initialTokens,
		LearningContext: NewLearningContext(id),
		TaskQueue:       make([]*Task, 0),
		PerformanceMetrics: &PerformanceMetrics{
			AgentID:     id,
			LastUpdated: now,
		},
		Capabilities:    make(map[string]*Capability),
		DecisionHistory: make([]*DecisionRecord, 0),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (a *Agent) CanExecuteTask(requiredTokens int64) bool {
	return a.TokenBalance >= requiredTokens
}

func (a *Agent) DeductTokens(amount int64) error {
	if !a.CanExecuteTask(amount) {
		return ErrInsufficientTokens
	}
	a.TokenBalance -= amount
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Agent) AddTokens(amount int64) {
	a.TokenBalance += amount
	a.UpdatedAt = time.Now()
}

func (a *Agent) UpdateStatus(status AgentStatus) {
	a.Status = status
	a.UpdatedAt = time.Now()
}

func (a *Agent) IsActive() bool {
	return a.Status != StatusIdle && a.Status != StatusError
}

func (a *Agent) EnqueueTask(task *Task) {
	a.TaskQueue = append(a.TaskQueue, task)
	a.UpdatedAt = time.Now()
}

func (a *Agent) GetNextTask() *Task {
	if len(a.TaskQueue) == 0 {
		return nil
	}

	highestPriority := -1
	var selectedTask *Task
	var selectedIndex int

	for i, task := range a.TaskQueue {
		if task.Status == TaskStatusPending && task.Priority > highestPriority {
			highestPriority = task.Priority
			selectedTask = task
			selectedIndex = i
		}
	}

	if selectedTask != nil {
		a.TaskQueue = append(a.TaskQueue[:selectedIndex], a.TaskQueue[selectedIndex+1:]...)
	}

	return selectedTask
}

func (a *Agent) RecordDecision(taskID, decision, reasoning string, confidence float64) {
	record := NewDecisionRecord(a.ID, taskID, decision, reasoning, confidence)
	a.DecisionHistory = append(a.DecisionHistory, record)
	a.UpdatedAt = time.Now()
}

func (a *Agent) AddCapability(capability *Capability) {
	a.Capabilities[capability.Name] = capability
	a.UpdatedAt = time.Now()
}

func (a *Agent) HasCapability(name string) bool {
	_, exists := a.Capabilities[name]
	return exists
}

func (a *Agent) UseCapability(name string) error {
	capability, exists := a.Capabilities[name]
	if !exists {
		return ErrCapabilityNotFound
	}
	capability.Use()
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Agent) LearnFromWebsite(snapshot *WebsiteSnapshot) {
	node := NewKnowledgeNode(NodeTypeWebsite, snapshot.URL)
	node.Properties["title"] = snapshot.Title
	node.Properties["technologies"] = snapshot.Technologies
	node.Confidence = snapshot.Confidence

	a.LearningContext.AddNode(node)

	for _, tech := range snapshot.Technologies {
		techNode := NewKnowledgeNode(NodeTypeTechnology, tech)
		a.LearningContext.AddNode(techNode)

		edge := NewKnowledgeEdge(node.ID, techNode.ID, "uses", 1.0)
		a.LearningContext.AddEdge(edge)
	}

	a.UpdatedAt = time.Now()
}

func (a *Agent) AddPattern(pattern *Pattern) {
	a.LearningContext.AddPattern(pattern)
	a.UpdatedAt = time.Now()
}

func (a *Agent) AddInsight(insight *Insight) {
	a.LearningContext.AddInsight(insight)
	a.UpdatedAt = time.Now()
}

func (a *Agent) GetLearningConfidence() float64 {
	return a.LearningContext.Confidence
}

func (a *Agent) GetKnowledgeNodeCount() int {
	return a.LearningContext.TotalNodes
}

func (a *Agent) GetSuccessRate() float64 {
	return a.PerformanceMetrics.SuccessRate
}

func (a *Agent) RecordTaskSuccess(tokensUsed int64, executionTime time.Duration) {
	a.PerformanceMetrics.RecordSuccess(tokensUsed, executionTime)
	a.UpdatedAt = time.Now()
}

func (a *Agent) RecordTaskFailure() {
	a.PerformanceMetrics.RecordFailure()
	a.UpdatedAt = time.Now()
}
