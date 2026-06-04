package application

import "errors"

// Sentinel-ошибки application (err113).
var (
	ErrNoApprovalChannel       = errors.New("no approval channel for session")
	ErrApprovalTimeout         = errors.New("approval timeout for session")
	ErrApprovalCancelled       = errors.New("approval cancelled")
	ErrApprovalSessionNotFound = errors.New("session not found or already resolved")
	ErrApprovalChannelClosed   = errors.New("session channel full or closed")
	ErrNoMediaApprovalChannel  = errors.New("no media approval channel for session")
	ErrMediaApprovalTimeout    = errors.New("media approval timeout for session")
	ErrMediaApprovalCancelled  = errors.New("media approval cancelled")
	ErrMediaSessionNotFound    = errors.New("media session not found or already resolved")
	ErrMediaChannelClosed      = errors.New("media session channel full or closed")
	ErrNoFundsWaitChannel      = errors.New("no funds wait channel for session")
	ErrFundsWaitTimeout        = errors.New("funds wait timeout for session")
	ErrFundsWaitCancelled      = errors.New("funds wait cancelled")
	ErrFundsSessionNotFound    = errors.New("session not found or not paused for funds")
	ErrFundsChannelClosed      = errors.New("session channel full or already resumed")
	ErrPlannerNotInitialized   = errors.New("planner not initialized")
	ErrFeaturesRejected        = errors.New("features rejected by user")
	ErrCodingPhaseDone         = errors.New("agent coding phase completed")
	ErrCoderPanic              = errors.New("coder panic")
	ErrManifestFileMapTooSmall = errors.New("manifest FileMap too small for chunked generation")
	ErrNoFileGroups            = errors.New("no file groups after classification")
	ErrChunkedGenerationEmpty  = errors.New("chunked generation produced 0 files across all tiers")
	ErrEmptyPackageJSON        = errors.New("package.json is empty")
	ErrEmptyTSConfigJSON       = errors.New("tsconfig.json is empty")
)
