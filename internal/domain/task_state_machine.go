package domain

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// TaskState определяет состояние задачи в конечном автомате
type TaskState string

const (
	StateCreated              TaskState = "created"
	StateResearching          TaskState = "researching"
	StatePlanning             TaskState = "planning"
	StateArchitectureApproved TaskState = "architecture_approved"
	StateStrategySynthesized  TaskState = "strategy_synthesized"
	StateDesigning            TaskState = "designing"
	StateCoding               TaskState = "coding"
	StateQualityCheck         TaskState = "quality_check"
	StateSecurityCheck        TaskState = "security_check"
	StateRetryCoding          TaskState = "retry_coding"
	StateVerified             TaskState = "verified"
	StateCompleted            TaskState = "completed"
	StateFailed               TaskState = "failed"
)

// ApprovedPlan — утверждённый JSON-план от Агента-Планировщика (Director).
// Должен быть заполнен перед переходом в StateCoding.
type ApprovedPlan struct {
	Architecture string    `json:"architecture"`
	Steps        []string  `json:"steps"`
	Components   []string  `json:"components"`
	Technologies []string  `json:"technologies"`
	ApprovedAt   time.Time `json:"approved_at"`
	ApprovedBy   string    `json:"approved_by"`
}

// Valid проверяет, содержит ли план минимально необходимую информацию.
func (p *ApprovedPlan) Valid() bool {
	return p != nil &&
		p.Architecture != "" &&
		len(p.Steps) > 0 &&
		!p.ApprovedAt.IsZero()
}

// StateTransition описывает конкретный переход между состояниями
type StateTransition struct {
	From      TaskState
	To        TaskState
	Timestamp time.Time
	Reason    string
}

// TaskStateMachine — конечный автомат задачи.
// Управляет переходами между 11 состояниями, валидирует допустимость
// каждого перехода и хранит утверждённый план для gate-проверки перед Coding.
type TaskStateMachine struct {
	mu          sync.RWMutex
	current     TaskState
	plan        *ApprovedPlan
	transitions []StateTransition
	createdAt   time.Time
}

// allowedTransitions — таблица допустимых переходов FSM.
// Ключ — текущее состояние, значение — множество допустимых целевых состояний.
var allowedTransitions = map[TaskState]map[TaskState]bool{
	StateCreated: {
		StateResearching: true,
		StatePlanning:    true, // code mode может пропустить Research
		StateFailed:      true,
	},
	StateResearching: {
		StatePlanning: true,
		StateFailed:   true,
	},
	StatePlanning: {
		StateArchitectureApproved: true,
		StateFailed:               true,
	},
	StateArchitectureApproved: {
		StateStrategySynthesized: true,
		StateDesigning:           true, // может пропустить Strategy
		StateFailed:              true,
	},
	StateStrategySynthesized: {
		StateDesigning: true,
		StateCoding:    true, // code mode без дизайна
		StateFailed:    true,
	},
	StateDesigning: {
		StateCoding: true,
		StateFailed: true,
	},
	StateCoding: {
		StateQualityCheck: true,
		StateFailed:       true,
	},
	StateQualityCheck: {
		StateSecurityCheck: true,
		StateRetryCoding:   true, // auto-fix: return to coder with error log
		StateFailed:        true,
	},
	StateSecurityCheck: {
		StateVerified:    true,
		StateRetryCoding: true, // auto-fix: return to coder with error log
		StateFailed:      true,
	},
	StateRetryCoding: {
		StateCoding: true, // план уже утверждён, gate пройдёт
		StateFailed: true,
	},
	StateVerified: {
		StateCompleted: true,
		StateFailed:    true,
	},
	StateCompleted: {
		// терминальное состояние
	},
	StateFailed: {
		StateCreated: true, // retry с начала
	},
}

// NewTaskStateMachine создаёт FSM в начальном состоянии Created.
func NewTaskStateMachine() *TaskStateMachine {
	now := time.Now()
	return &TaskStateMachine{
		current:   StateCreated,
		createdAt: now,
		transitions: []StateTransition{
			{From: "", To: StateCreated, Timestamp: now, Reason: "initialized"},
		},
	}
}

// Current возвращает текущее состояние FSM.
func (fsm *TaskStateMachine) Current() TaskState {
	fsm.mu.RLock()
	defer fsm.mu.RUnlock()
	return fsm.current
}

// Plan возвращает утверждённый план (может быть nil).
func (fsm *TaskStateMachine) Plan() *ApprovedPlan {
	fsm.mu.RLock()
	defer fsm.mu.RUnlock()
	return fsm.plan
}

// ApprovePlan сохраняет утверждённый план в FSM.
// Вызывается после того, как Director/Planner сгенерировал план.
func (fsm *TaskStateMachine) ApprovePlan(plan ApprovedPlan) error {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	if plan.Architecture == "" || len(plan.Steps) == 0 {
		return fmt.Errorf("plan rejected: architecture and steps are required")
	}

	plan.ApprovedAt = time.Now()
	fsm.plan = &plan
	return nil
}

// TransitionTo выполняет переход в целевое состояние.
// Возвращает ошибку если:
//   - переход недопустим по таблице allowedTransitions
//   - переход в StateCoding, но plan не утверждён (gate)
func (fsm *TaskStateMachine) TransitionTo(target TaskState, reason string) error {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	// Проверка допустимости перехода
	allowed, ok := allowedTransitions[fsm.current]
	if !ok {
		return fmt.Errorf("FSM: no transitions defined from state %q", fsm.current)
	}
	if !allowed[target] {
		return fmt.Errorf("FSM: transition %q → %q is not allowed", fsm.current, target)
	}

	// Gate: переход в Coding требует утверждённого плана
	if target == StateCoding {
		if fsm.plan == nil || !fsm.plan.Valid() {
			return fmt.Errorf(
				"FSM: transition to %q blocked — approved plan from Planner agent is required (plan=%v)",
				StateCoding, fsm.plan != nil,
			)
		}
	}

	transition := StateTransition{
		From:      fsm.current,
		To:        target,
		Timestamp: time.Now(),
		Reason:    reason,
	}
	fsm.transitions = append(fsm.transitions, transition)
	fsm.current = target

	return nil
}

// Transitions возвращает копию истории переходов.
func (fsm *TaskStateMachine) Transitions() []StateTransition {
	fsm.mu.RLock()
	defer fsm.mu.RUnlock()
	cp := make([]StateTransition, len(fsm.transitions))
	copy(cp, fsm.transitions)
	return cp
}

// Duration возвращает время с момента создания FSM.
func (fsm *TaskStateMachine) Duration() time.Duration {
	fsm.mu.RLock()
	defer fsm.mu.RUnlock()
	return time.Since(fsm.createdAt)
}

// IsTerminal возвращает true если FSM в терминальном состоянии.
func (fsm *TaskStateMachine) IsTerminal() bool {
	fsm.mu.RLock()
	defer fsm.mu.RUnlock()
	return fsm.current == StateCompleted || fsm.current == StateFailed
}

// MarshalJSON сериализует FSM для отладки / SSE.
func (fsm *TaskStateMachine) MarshalJSON() ([]byte, error) {
	fsm.mu.RLock()
	defer fsm.mu.RUnlock()
	return json.Marshal(struct {
		State       TaskState `json:"state"`
		HasPlan     bool      `json:"has_plan"`
		Transitions int       `json:"transitions_count"`
		Duration    string    `json:"duration"`
	}{
		State:       fsm.current,
		HasPlan:     fsm.plan != nil && fsm.plan.Valid(),
		Transitions: len(fsm.transitions),
		Duration:    time.Since(fsm.createdAt).String(),
	})
}
