package openrouter

import "time"

type ModelCapability string

const (
	CapabilityCodeGeneration ModelCapability = "code_generation"
	CapabilityAnalysis       ModelCapability = "analysis"
	CapabilityReasoning      ModelCapability = "reasoning"
	CapabilityVision         ModelCapability = "vision"
	CapabilityFastResponse   ModelCapability = "fast_response"
)

type ModelTier string

const (
	TierPremium  ModelTier = "premium"
	TierStandard ModelTier = "standard"
	TierEconomy  ModelTier = "economy"
)

type ModelConfig struct {
	ID               string
	Name             string
	Provider         string
	Tier             ModelTier
	Capabilities     []ModelCapability
	MaxTokens        int
	CostPer1KTokens  float64
	AverageLatencyMs int
	ReliabilityScore float64
	ContextWindow    int
}

type ModelHealth struct {
	ModelID          string
	IsAvailable      bool
	LastChecked      time.Time
	FailureCount     int
	SuccessCount     int
	AverageLatency   time.Duration
	LastError        string
	ConsecutiveFails int
}

type FallbackStrategy struct {
	Models           []ModelConfig
	MaxRetries       int
	TimeoutPerModel  time.Duration
	CostThreshold    float64
	QualityThreshold float64
	PreferFast       bool
	PreferCheap      bool
}

var ModelRegistry = map[string]ModelConfig{
	"anthropic/claude-3.5-sonnet": {
		ID:               "anthropic/claude-3.5-sonnet",
		Name:             "Claude 3.5 Sonnet",
		Provider:         "Anthropic",
		Tier:             TierPremium,
		Capabilities:     []ModelCapability{CapabilityCodeGeneration, CapabilityAnalysis, CapabilityReasoning},
		MaxTokens:        8192,
		CostPer1KTokens:  0.015,
		AverageLatencyMs: 2000,
		ReliabilityScore: 0.98,
		ContextWindow:    200000,
	},
	"openai/gpt-4o": {
		ID:               "openai/gpt-4o",
		Name:             "GPT-4o",
		Provider:         "OpenAI",
		Tier:             TierPremium,
		Capabilities:     []ModelCapability{CapabilityCodeGeneration, CapabilityAnalysis, CapabilityReasoning, CapabilityVision},
		MaxTokens:        4096,
		CostPer1KTokens:  0.01,
		AverageLatencyMs: 1500,
		ReliabilityScore: 0.97,
		ContextWindow:    128000,
	},
	"google/gemini-2.0-flash-exp": {
		ID:               "google/gemini-2.0-flash-exp",
		Name:             "Gemini 2.0 Flash",
		Provider:         "Google",
		Tier:             TierStandard,
		Capabilities:     []ModelCapability{CapabilityCodeGeneration, CapabilityAnalysis, CapabilityFastResponse},
		MaxTokens:        8192,
		CostPer1KTokens:  0.005,
		AverageLatencyMs: 800,
		ReliabilityScore: 0.95,
		ContextWindow:    1000000,
	},
	"anthropic/claude-4.6-sonnet": {
		ID:               "anthropic/claude-4.6-sonnet",
		Name:             "Claude 4.6 Sonnet",
		Provider:         "Anthropic",
		Tier:             TierPremium,
		Capabilities:     []ModelCapability{CapabilityCodeGeneration, CapabilityAnalysis, CapabilityReasoning, CapabilityVision},
		MaxTokens:        8192,
		CostPer1KTokens:  0.025,
		AverageLatencyMs: 2500,
		ReliabilityScore: 0.99,
		ContextWindow:    200000,
	},
	"openai/gpt-4o-mini": {
		ID:               "openai/gpt-4o-mini",
		Name:             "GPT-4o Mini",
		Provider:         "OpenAI",
		Tier:             TierEconomy,
		Capabilities:     []ModelCapability{CapabilityCodeGeneration, CapabilityAnalysis, CapabilityFastResponse},
		MaxTokens:        4096,
		CostPer1KTokens:  0.0008,
		AverageLatencyMs: 800,
		ReliabilityScore: 0.94,
		ContextWindow:    128000,
	},
	"meta-llama/llama-3.3-70b-instruct": {
		ID:               "meta-llama/llama-3.3-70b-instruct",
		Name:             "Llama 3.3 70B",
		Provider:         "Meta",
		Tier:             TierEconomy,
		Capabilities:     []ModelCapability{CapabilityCodeGeneration, CapabilityAnalysis},
		MaxTokens:        4096,
		CostPer1KTokens:  0.002,
		AverageLatencyMs: 1200,
		ReliabilityScore: 0.92,
		ContextWindow:    128000,
	},
}

func GetDefaultFallbackStrategy() *FallbackStrategy {
	return &FallbackStrategy{
		Models: []ModelConfig{
			ModelRegistry["anthropic/claude-3.5-sonnet"],
			ModelRegistry["openai/gpt-4o"],
			ModelRegistry["google/gemini-2.0-flash-exp"],
			ModelRegistry["meta-llama/llama-3.3-70b-instruct"],
		},
		MaxRetries:       3,
		TimeoutPerModel:  30 * time.Second,
		CostThreshold:    0.02,
		QualityThreshold: 0.7,
		PreferFast:       false,
		PreferCheap:      false,
	}
}

func GetFastFallbackStrategy() *FallbackStrategy {
	return &FallbackStrategy{
		Models: []ModelConfig{
			ModelRegistry["google/gemini-2.0-flash-exp"],
			ModelRegistry["openai/gpt-4o"],
			ModelRegistry["anthropic/claude-3.5-sonnet"],
		},
		MaxRetries:       2,
		TimeoutPerModel:  15 * time.Second,
		CostThreshold:    0.015,
		QualityThreshold: 0.6,
		PreferFast:       true,
		PreferCheap:      false,
	}
}

func GetEconomyFallbackStrategy() *FallbackStrategy {
	return &FallbackStrategy{
		Models: []ModelConfig{
			ModelRegistry["meta-llama/llama-3.3-70b-instruct"],
			ModelRegistry["google/gemini-2.0-flash-exp"],
			ModelRegistry["openai/gpt-4o"],
		},
		MaxRetries:       3,
		TimeoutPerModel:  30 * time.Second,
		CostThreshold:    0.01,
		QualityThreshold: 0.5,
		PreferFast:       false,
		PreferCheap:      true,
	}
}

func (mc *ModelConfig) HasCapability(capability ModelCapability) bool {
	for _, cap := range mc.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

func (mc *ModelConfig) EstimateCost(inputTokens, outputTokens int) float64 {
	totalTokens := float64(inputTokens + outputTokens)
	return (totalTokens / 1000.0) * mc.CostPer1KTokens
}
