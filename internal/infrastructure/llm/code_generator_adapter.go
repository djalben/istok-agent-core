package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/istok/agent-core/internal/ports"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — CodeGeneratorAdapter (LLM-backed)
//  Реализует ports.CodeGenerator поверх ports.LLMProvider (Anthropic Direct).
//  Полностью заменяет legacy openrouter-адаптер — пакет openrouter удалён.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// CodeGeneratorAdapter генерирует/рефакторит код через любой LLMProvider.
type CodeGeneratorAdapter struct {
	llm   ports.LLMProvider
	model string // модель по умолчанию для всех вызовов
}

// NewCodeGeneratorAdapter создаёт адаптер.
// model — каноничный идентификатор Anthropic-модели (например "anthropic/claude-3-7-sonnet").
func NewCodeGeneratorAdapter(llm ports.LLMProvider, model string) *CodeGeneratorAdapter {
	if model == "" {
		model = "anthropic/claude-3-7-sonnet"
	}
	return &CodeGeneratorAdapter{llm: llm, model: model}
}

// GenerateCode генерирует код по спецификации.
func (a *CodeGeneratorAdapter) GenerateCode(ctx context.Context, req ports.GenerateCodeRequest) (*ports.GenerateCodeResponse, error) {
	prompt := a.buildPrompt(req, nil)
	resp, err := a.llm.Complete(ctx, ports.LLMRequest{
		Model:        a.model,
		SystemPrompt: "You are a senior engineer. Return only production-ready code, no commentary.",
		UserPrompt:   prompt,
		MaxTokens:    8000,
		Temperature:  0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	return &ports.GenerateCodeResponse{
		Code:         stripCodeFences(resp.Content),
		Explanation:  "",
		TokensUsed:   int64(resp.TokensUsed),
		Dependencies: extractDeps(resp.Content, req.Language),
	}, nil
}

// GenerateWithContext генерирует код с обогащённым контекстом обучения.
func (a *CodeGeneratorAdapter) GenerateWithContext(ctx context.Context, req ports.GenerateCodeRequest, learningContext interface{}) (*ports.GenerateCodeResponse, error) {
	prompt := a.buildPrompt(req, learningContext)
	resp, err := a.llm.Complete(ctx, ports.LLMRequest{
		Model:        a.model,
		SystemPrompt: "You are a senior engineer. Apply learned patterns. Return only code.",
		UserPrompt:   prompt,
		MaxTokens:    8000,
		Temperature:  0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}
	return &ports.GenerateCodeResponse{
		Code:         stripCodeFences(resp.Content),
		Explanation:  "",
		TokensUsed:   int64(resp.TokensUsed),
		Dependencies: extractDeps(resp.Content, req.Language),
	}, nil
}

// AnalyzeWebsite — анализ сайта по URL (текстовый).
func (a *CodeGeneratorAdapter) AnalyzeWebsite(ctx context.Context, req ports.AnalyzeWebsiteRequest) (*ports.AnalyzeWebsiteResponse, error) {
	prompt := fmt.Sprintf(
		"Analyze the website at %s. AnalysisType=%s, Depth=%d. Return JSON with structure, technologies, summary.",
		req.URL, req.AnalysisType, req.Depth)
	resp, err := a.llm.Complete(ctx, ports.LLMRequest{
		Model:       a.model,
		UserPrompt:  prompt,
		MaxTokens:   2048,
		Temperature: 0.2,
	})
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Structure    map[string]interface{} `json:"structure"`
		Technologies []string               `json:"technologies"`
		Summary      string                 `json:"summary"`
	}
	body := stripCodeFences(resp.Content)
	_ = json.Unmarshal([]byte(body), &parsed)
	if parsed.Summary == "" {
		parsed.Summary = body
	}

	return &ports.AnalyzeWebsiteResponse{
		Structure:    parsed.Structure,
		Technologies: parsed.Technologies,
		Summary:      parsed.Summary,
		TokensUsed:   int64(resp.TokensUsed),
	}, nil
}

// RefactorCode рефакторит код по инструкциям.
func (a *CodeGeneratorAdapter) RefactorCode(ctx context.Context, req ports.RefactorCodeRequest) (*ports.RefactorCodeResponse, error) {
	prompt := fmt.Sprintf(
		"Refactor this %s code per instructions: %s\nTarget pattern: %s\n\nCode:\n%s\n\nReturn refactored code only.",
		req.Language, req.Instructions, req.TargetPattern, req.Code)
	resp, err := a.llm.Complete(ctx, ports.LLMRequest{
		Model:       a.model,
		UserPrompt:  prompt,
		MaxTokens:   8000,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, err
	}
	return &ports.RefactorCodeResponse{
		RefactoredCode: stripCodeFences(resp.Content),
		Changes:        nil,
		TokensUsed:     int64(resp.TokensUsed),
	}, nil
}

// EstimateCost — грубая оценка по длине промпта (без отдельного LLM-вызова).
func (a *CodeGeneratorAdapter) EstimateCost(_ context.Context, req ports.GenerateCodeRequest) (*ports.CostEstimateResponse, error) {
	tokens := int64(len(req.Specification)/4) + 2000
	return &ports.CostEstimateResponse{
		EstimatedTokens: tokens,
		EstimatedCost:   float64(tokens) / 1000.0 * 3.0, // Anthropic Sonnet pricing approx
		Confidence:      0.5,
	}, nil
}

// ExplainDecision объясняет решение.
func (a *CodeGeneratorAdapter) ExplainDecision(ctx context.Context, decision string) (*ports.ExplanationResponse, error) {
	resp, err := a.llm.Complete(ctx, ports.LLMRequest{
		Model:      a.model,
		UserPrompt: fmt.Sprintf("Explain reasoning behind: %s. List alternatives + considerations.", decision),
		MaxTokens:  1024,
	})
	if err != nil {
		return nil, err
	}
	return &ports.ExplanationResponse{
		Reasoning:      resp.Content,
		Confidence:     0.8,
		Alternatives:   nil,
		Considerations: nil,
	}, nil
}

// ValidateOutput валидирует код через LLM.
func (a *CodeGeneratorAdapter) ValidateOutput(ctx context.Context, code, language string) (*ports.ValidationResponse, error) {
	resp, err := a.llm.Complete(ctx, ports.LLMRequest{
		Model:       a.model,
		UserPrompt:  fmt.Sprintf("Validate this %s code for correctness, best practices, security:\n\n%s\n\nReturn JSON: {\"is_valid\":bool,\"issues\":[],\"suggestions\":[],\"quality_score\":0..1}", language, code),
		MaxTokens:   2048,
		Temperature: 0.1,
	})
	if err != nil {
		return nil, err
	}

	body := stripCodeFences(resp.Content)
	var parsed struct {
		IsValid      bool     `json:"is_valid"`
		Issues       []string `json:"issues"`
		Suggestions  []string `json:"suggestions"`
		QualityScore float64  `json:"quality_score"`
	}
	_ = json.Unmarshal([]byte(body), &parsed)

	return &ports.ValidationResponse{
		IsValid:      parsed.IsValid,
		Issues:       parsed.Issues,
		Suggestions:  parsed.Suggestions,
		QualityScore: parsed.QualityScore,
	}, nil
}

// buildPrompt собирает промпт генерации.
func (a *CodeGeneratorAdapter) buildPrompt(req ports.GenerateCodeRequest, learning interface{}) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Generate %s code for the following specification:\n\n", req.Language))
	sb.WriteString(req.Specification)
	if req.Framework != "" {
		sb.WriteString(fmt.Sprintf("\n\nFramework: %s", req.Framework))
	}
	if len(req.Context) > 0 {
		ctxJSON, _ := json.Marshal(req.Context)
		sb.WriteString(fmt.Sprintf("\n\nContext: %s", string(ctxJSON)))
	}
	if learning != nil {
		sb.WriteString(fmt.Sprintf("\n\nLearned patterns: %v", learning))
	}
	sb.WriteString("\n\nReturn ONLY the code, no markdown fences, no commentary.")
	return sb.String()
}

func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// Drop opening fence line
		if idx := strings.Index(s, "\n"); idx > 0 {
			s = s[idx+1:]
		}
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return s
}

func extractDeps(code, language string) []string {
	deps := make([]string, 0)
	for _, line := range strings.Split(code, "\n") {
		line = strings.TrimSpace(line)
		switch language {
		case "Go", "go":
			if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "\"") {
				deps = append(deps, line)
			}
		case "JavaScript", "TypeScript", "javascript", "typescript":
			if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "require(") {
				deps = append(deps, line)
			}
		case "Python", "python":
			if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "from ") {
				deps = append(deps, line)
			}
		}
	}
	return deps
}
