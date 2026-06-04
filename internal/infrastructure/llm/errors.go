package llm

import (
	"errors"
	"strings"

	"github.com/djalben/istok-agent-core/internal/ports"
)

// ErrInsufficientFunds re-exports the sentinel for convenience within adapters.
var ErrInsufficientFunds = ports.ErrInsufficientFunds

// Sentinel-ошибки LLM-адаптеров (err113).
var (
	ErrAnthropicAPIKeyNotConfigured = errors.New("anthropic API key not configured")
	ErrAnthropicAPIError            = errors.New("anthropic API error")
	ErrAnthropicAPIResponseError    = errors.New("anthropic API response error")
	ErrAnthropicEmptyContent        = errors.New("anthropic returned empty content")
	ErrReplicateTokenNotSet         = errors.New("replicate API token not set")
	ErrReplicatePredictionError     = errors.New("replicate prediction error")
	ErrReplicatePredictionTimeout   = errors.New("replicate prediction timed out")
	ErrReplicateEmptyOutput         = errors.New("empty output from replicate")
	ErrReplicatePredictionFailed    = errors.New("replicate prediction failed")
	ErrReplicateAPIError            = errors.New("replicate API error")
	ErrReplicatePollHTTPError       = errors.New("replicate poll HTTP error")
)

// isInsufficientFundsError detects credit-exhaustion responses from LLM providers.
// Matches HTTP 402 (Payment Required), or 400/429 with known quota/credit keywords.
func isInsufficientFundsError(statusCode int, body string) bool {
	if statusCode == 402 {
		return true
	}
	if statusCode != 400 && statusCode != 429 {
		return false
	}
	lower := strings.ToLower(body)
	keywords := []string{
		"insufficient_quota",
		"credit balance is too low",
		"out of credits",
		"billing",
		"exceeded your current quota",
		"rate_limit_error",
		"spending limit",
	}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}

	return false
}
