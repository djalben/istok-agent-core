package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type Task struct {
	ID          string
	AgentID     string
	Type        string
	Description string
	Status      TaskStatus
	Priority    int
	TokenCost   int64
	Result      map[string]interface{}
	Error       string
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}

type PerformanceMetrics struct {
	AgentID              string
	TotalTasks           int64
	SuccessfulTasks      int64
	FailedTasks          int64
	TotalTokensUsed      int64
	AverageTokensPerTask float64
	AverageExecutionTime time.Duration
	SuccessRate          float64
	LastUpdated          time.Time
}

type DecisionRecord struct {
	ID           string
	AgentID      string
	TaskID       string
	Decision     string
	Reasoning    string
	Confidence   float64
	Alternatives []string
	Context      map[string]interface{}
	Outcome      string
	Timestamp    time.Time
}

type CapabilityLevel string

const (
	CapabilityNovice       CapabilityLevel = "novice"
	CapabilityIntermediate CapabilityLevel = "intermediate"
	CapabilityAdvanced     CapabilityLevel = "advanced"
	CapabilityExpert       CapabilityLevel = "expert"
)

type Capability struct {
	Name        string
	Description string
	Level       CapabilityLevel
	Confidence  float64
	UsageCount  int64
	LastUsed    time.Time
	AcquiredAt  time.Time
}

func NewTask(agentID, taskType, description string, priority int, tokenCost int64) *Task {
	return &Task{
		ID:          generateID(),
		AgentID:     agentID,
		Type:        taskType,
		Description: description,
		Status:      TaskStatusPending,
		Priority:    priority,
		TokenCost:   tokenCost,
		Result:      make(map[string]interface{}),
		CreatedAt:   time.Now(),
	}
}

func (t *Task) Start() {
	now := time.Now()
	t.Status = TaskStatusInProgress
	t.StartedAt = &now
}

func (t *Task) Complete(result map[string]interface{}) {
	now := time.Now()
	t.Status = TaskStatusCompleted
	t.CompletedAt = &now
	t.Result = result
}

func (t *Task) Fail(err string) {
	now := time.Now()
	t.Status = TaskStatusFailed
	t.CompletedAt = &now
	t.Error = err
}

func (pm *PerformanceMetrics) RecordSuccess(tokensUsed int64, executionTime time.Duration) {
	pm.TotalTasks++
	pm.SuccessfulTasks++
	pm.TotalTokensUsed += tokensUsed
	pm.AverageTokensPerTask = float64(pm.TotalTokensUsed) / float64(pm.TotalTasks)
	pm.SuccessRate = float64(pm.SuccessfulTasks) / float64(pm.TotalTasks)
	
	if pm.AverageExecutionTime == 0 {
		pm.AverageExecutionTime = executionTime
	} else {
		pm.AverageExecutionTime = (pm.AverageExecutionTime + executionTime) / 2
	}
	
	pm.LastUpdated = time.Now()
}

func (pm *PerformanceMetrics) RecordFailure() {
	pm.TotalTasks++
	pm.FailedTasks++
	pm.SuccessRate = float64(pm.SuccessfulTasks) / float64(pm.TotalTasks)
	pm.LastUpdated = time.Now()
}

func NewDecisionRecord(agentID, taskID, decision, reasoning string, confidence float64) *DecisionRecord {
	return &DecisionRecord{
		ID:           generateID(),
		AgentID:      agentID,
		TaskID:       taskID,
		Decision:     decision,
		Reasoning:    reasoning,
		Confidence:   confidence,
		Alternatives: make([]string, 0),
		Context:      make(map[string]interface{}),
		Timestamp:    time.Now(),
	}
}

func NewCapability(name, description string, level CapabilityLevel) *Capability {
	return &Capability{
		Name:        name,
		Description: description,
		Level:       level,
		Confidence:  0.5,
		UsageCount:  0,
		AcquiredAt:  time.Now(),
	}
}

func (c *Capability) Use() {
	c.UsageCount++
	c.LastUsed = time.Now()
	
	if c.UsageCount > 100 && c.Level == CapabilityNovice {
		c.Level = CapabilityIntermediate
	} else if c.UsageCount > 500 && c.Level == CapabilityIntermediate {
		c.Level = CapabilityAdvanced
	} else if c.UsageCount > 1000 && c.Level == CapabilityAdvanced {
		c.Level = CapabilityExpert
	}
	
	c.Confidence = min(1.0, c.Confidence+0.001)
}

func generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
