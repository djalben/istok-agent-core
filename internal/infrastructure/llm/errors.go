package llm

import (
	"strings"

	"github.com/djalben/istok-agent-core/internal/ports"
)

// ErrInsufficientFunds re-exports the sentinel for convenience within adapters.
var ErrInsufficientFunds = ports.ErrInsufficientFunds

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
