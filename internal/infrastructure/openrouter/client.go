package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Client struct {
	apiKey         string
	baseURL        string
	httpClient     *http.Client
	healthMap      map[string]*ModelHealth
	healthMutex    sync.RWMutex
	circuitBreaker *CircuitBreaker
	rateLimiter    *RateLimiter
	telemetry      *Telemetry
}

// ThinkingConfig параметры extended thinking для Claude
type ThinkingConfig struct {
	Type         string `json:"type"`          // "enabled"
	BudgetTokens int    `json:"budget_tokens"` // количество токенов на размышление
}

type CompletionRequest struct {
	Model       string                 `json:"model"`
	Messages    []Message              `json:"messages"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Thinking    *ThinkingConfig        `json:"thinking,omitempty"` // Extended thinking для Claude
	Metadata    map[string]interface{} `json:"-"`
}

// NewThinkingRequest создает запрос с активным extended thinking
func NewThinkingRequest(model string, messages []Message, budgetTokens int) CompletionRequest {
	return CompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   16000,
		Temperature: 1.0, // Обязательно 1.0 для thinking mode
		Thinking: &ThinkingConfig{
			Type:         "enabled",
			BudgetTokens: budgetTokens,
		},
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type CostEstimate struct {
	EstimatedCost   float64
	InputTokens     int
	OutputTokens    int
	Model           string
	ConfidenceLevel float64
}

func NewClient(apiKey string) *Client {
	baseURL := os.Getenv("OPENROUTER_PROXY_URL")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	log.Printf("🔗 OpenRouter base URL: %s", baseURL)
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // AI models могут долго думать
		},
		healthMap:      make(map[string]*ModelHealth),
		circuitBreaker: NewCircuitBreaker(10, 2*time.Minute), // Увеличили порог: 10 ошибок, 2мин ресет
		rateLimiter:    NewRateLimiter(100, time.Minute),
		telemetry:      NewTelemetry(),
	}
}

func (c *Client) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	startTime := time.Now()

	if !c.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	if !c.circuitBreaker.Allow() {
		return nil, fmt.Errorf("circuit breaker open")
	}

	resp, err := c.doRequest(ctx, req)

	duration := time.Since(startTime)
	c.telemetry.RecordRequest(req.Model, duration, err == nil)

	if err != nil {
		c.circuitBreaker.RecordFailure()
		c.updateModelHealth(req.Model, false, err.Error())
		return nil, err
	}

	c.circuitBreaker.RecordSuccess()
	c.updateModelHealth(req.Model, true, "")

	return resp, nil
}

func (c *Client) CompleteWithFallback(ctx context.Context, req CompletionRequest, strategy *FallbackStrategy) (*CompletionResponse, error) {
	if strategy == nil {
		strategy = GetDefaultFallbackStrategy()
	}

	var lastErr error

	for attempt := 0; attempt < strategy.MaxRetries; attempt++ {
		for _, model := range strategy.Models {
			health := c.GetModelHealth(model.ID)

			if health.ConsecutiveFails >= 3 {
				continue
			}

			reqCopy := req
			reqCopy.Model = model.ID

			ctxWithTimeout, cancel := context.WithTimeout(ctx, strategy.TimeoutPerModel)

			resp, err := c.Complete(ctxWithTimeout, reqCopy)
			cancel()

			if err == nil {
				c.telemetry.RecordFallbackSuccess(model.ID, attempt)
				return resp, nil
			}

			lastErr = err
			c.telemetry.RecordFallbackAttempt(model.ID, attempt, err)
		}

		if attempt < strategy.MaxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	return nil, fmt.Errorf("all fallback attempts failed: %w", lastErr)
}

func (c *Client) EstimateCost(req CompletionRequest) (*CostEstimate, error) {
	model, exists := ModelRegistry[req.Model]
	if !exists {
		return nil, fmt.Errorf("unknown model: %s", req.Model)
	}

	inputTokens := c.estimateTokenCount(req.Messages)
	outputTokens := req.MaxTokens
	if outputTokens == 0 {
		outputTokens = 1000
	}

	cost := model.EstimateCost(inputTokens, outputTokens)

	return &CostEstimate{
		EstimatedCost:   cost,
		InputTokens:     inputTokens,
		OutputTokens:    outputTokens,
		Model:           req.Model,
		ConfidenceLevel: 0.8,
	}, nil
}

func (c *Client) GetModelHealth(modelID string) ModelHealth {
	c.healthMutex.RLock()
	defer c.healthMutex.RUnlock()

	if health, exists := c.healthMap[modelID]; exists {
		return *health
	}

	return ModelHealth{
		ModelID:     modelID,
		IsAvailable: true,
		LastChecked: time.Now(),
	}
}

func (c *Client) doRequest(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("HTTP-Referer", "https://istok-agent-core.vercel.app")
	httpReq.Header.Set("X-Title", "Istok Agent Core")
	httpReq.Header.Set("User-Agent", "IstokAgent/2.0")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		// Подробное логирование ошибок OpenRouter
		var apiErr struct {
			Error struct {
				Message string `json:"message"`
				Code    int    `json:"code"`
				Type    string `json:"type"`
			} `json:"error"`
		}
		errMsg := string(body)
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr == nil && apiErr.Error.Message != "" {
			errMsg = fmt.Sprintf("[%d] %s (type: %s)", apiErr.Error.Code, apiErr.Error.Message, apiErr.Error.Type)
		}
		log.Printf("🚨 OpenRouter Error | model=%s status=%d | %s", req.Model, httpResp.StatusCode, errMsg)
		return nil, fmt.Errorf("OpenRouter API error (HTTP %d): %s", httpResp.StatusCode, errMsg)
	}

	var resp CompletionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

func (c *Client) updateModelHealth(modelID string, success bool, errorMsg string) {
	c.healthMutex.Lock()
	defer c.healthMutex.Unlock()

	health, exists := c.healthMap[modelID]
	if !exists {
		health = &ModelHealth{
			ModelID:     modelID,
			IsAvailable: true,
			LastChecked: time.Now(),
		}
		c.healthMap[modelID] = health
	}

	health.LastChecked = time.Now()

	if success {
		health.SuccessCount++
		health.ConsecutiveFails = 0
		health.IsAvailable = true
	} else {
		health.FailureCount++
		health.ConsecutiveFails++
		health.LastError = errorMsg

		if health.ConsecutiveFails >= 5 {
			health.IsAvailable = false
		}
	}
}

func (c *Client) estimateTokenCount(messages []Message) int {
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Content)
	}
	return totalChars / 4
}

func (c *Client) GetTelemetry() *Telemetry {
	return c.telemetry
}
