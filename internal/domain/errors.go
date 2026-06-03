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

	// ── Layer 1: Auth & DB ──
	ErrNotFound     = errors.New("not found")
	ErrEmailExists  = errors.New("email already registered")
	ErrInvalidCreds = errors.New("invalid email or password")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)
