package domain

import "errors"

var (
	ErrInsufficientTokens  = errors.New("insufficient token balance")
	ErrInvalidAgent        = errors.New("invalid agent")
	ErrAgentNotFound       = errors.New("agent not found")
	ErrCapabilityNotFound  = errors.New("capability not found")
	ErrInvalidTask         = errors.New("invalid task")
	ErrTaskNotFound        = errors.New("task not found")
	ErrLearningContextFull = errors.New("learning context capacity exceeded")

	// ErrNotFound — ресурс не найден (Layer 1: auth & DB).
	ErrNotFound     = errors.New("not found")
	ErrEmailExists  = errors.New("email already registered")
	ErrInvalidCreds = errors.New("invalid email or password")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")

	// ErrPlanRejected — FSM: план отклонён (нет архитектуры или шагов).
	ErrPlanRejected            = errors.New("plan rejected: architecture and steps are required")
	ErrFSMNoTransitions        = errors.New("FSM: no transitions defined from state")
	ErrFSMTransitionNotAllowed = errors.New("FSM: transition is not allowed")
	ErrFSMCodingBlocked        = errors.New("FSM: transition to coding blocked — approved plan required")
)
