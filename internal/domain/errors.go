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
)
