package openrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/istok/agent-core/internal/ports"
)

type CodeGeneratorAdapter struct {
	client   *Client
	strategy *FallbackStrategy
}

func NewCodeGeneratorAdapter(apiKey string) *CodeGeneratorAdapter {
	return &CodeGeneratorAdapter{
		client:   NewClient(apiKey),
		strategy: GetDefaultFallbackStrategy(),
	}
}

func (a *CodeGeneratorAdapter) GenerateCode(ctx context.Context, req ports.GenerateCodeRequest) (*ports.GenerateCodeResponse, error) {
	prompt := a.buildCodeGenerationPrompt(req)
	
	completion, err := a.client.CompleteWithFallback(ctx, CompletionRequest{
		Model: "anthropic/claude-3.5-sonnet",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   4096,
		Temperature: 0.7,
	}, a.strategy)

	if err != nil {
		return nil, fmt.Errorf("code generation failed: %w", err)
	}

	return &ports.GenerateCodeResponse{
		Code:         completion.Choices[0].Message.Content,
		Explanation:  "Generated using AI with fallback strategy",
		TokensUsed:   int64(completion.Usage.TotalTokens),
		Dependencies: a.extractDependencies(completion.Choices[0].Message.Content, req.Language),
	}, nil
}

func (a *CodeGeneratorAdapter) GenerateWithContext(ctx context.Context, req ports.GenerateCodeRequest, learningContext interface{}) (*ports.GenerateCodeResponse, error) {
	prompt := a.buildCodeGenerationPromptWithContext(req, learningContext)
	
	completion, err := a.client.CompleteWithFallback(ctx, CompletionRequest{
		Model: "anthropic/claude-3.5-sonnet",
		Messages: []Message{
			{Role: "system", Content: "You are an expert code generator that learns from analyzed websites and applies patterns."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   4096,
		Temperature: 0.7,
	}, a.strategy)

	if err != nil {
		return nil, fmt.Errorf("context-aware code generation failed: %w", err)
	}

	return &ports.GenerateCodeResponse{
		Code:         completion.Choices[0].Message.Content,
		Explanation:  "Generated using learning context and AI reasoning",
		TokensUsed:   int64(completion.Usage.TotalTokens),
		Dependencies: a.extractDependencies(completion.Choices[0].Message.Content, req.Language),
	}, nil
}

func (a *CodeGeneratorAdapter) AnalyzeWebsite(ctx context.Context, req ports.AnalyzeWebsiteRequest) (*ports.AnalyzeWebsiteResponse, error) {
	prompt := fmt.Sprintf(`Analyze the website at %s. 
Analysis type: %s
Depth: %d

Provide a detailed analysis including:
1. Technology stack
2. Architecture patterns
3. UI/UX patterns
4. Business model insights

Return the analysis in JSON format.`, req.URL, req.AnalysisType, req.Depth)

	completion, err := a.client.CompleteWithFallback(ctx, CompletionRequest{
		Model: "anthropic/claude-3.5-sonnet",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   2048,
		Temperature: 0.5,
	}, a.strategy)

	if err != nil {
		return nil, fmt.Errorf("website analysis failed: %w", err)
	}

	structure := make(map[string]interface{})
	content := completion.Choices[0].Message.Content
	
	if err := json.Unmarshal([]byte(content), &structure); err != nil {
		structure["raw_analysis"] = content
	}

	return &ports.AnalyzeWebsiteResponse{
		Structure:    structure,
		Technologies: a.extractTechnologies(content),
		Summary:      a.extractSummary(content),
		TokensUsed:   int64(completion.Usage.TotalTokens),
	}, nil
}

func (a *CodeGeneratorAdapter) RefactorCode(ctx context.Context, req ports.RefactorCodeRequest) (*ports.RefactorCodeResponse, error) {
	prompt := fmt.Sprintf(`Refactor the following %s code according to these instructions: %s

Target pattern: %s

Code:
%s

Provide the refactored code and list of changes.`, req.Language, req.Instructions, req.TargetPattern, req.Code)

	completion, err := a.client.CompleteWithFallback(ctx, CompletionRequest{
		Model: "anthropic/claude-3.5-sonnet",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   4096,
		Temperature: 0.6,
	}, a.strategy)

	if err != nil {
		return nil, fmt.Errorf("code refactoring failed: %w", err)
	}

	return &ports.RefactorCodeResponse{
		RefactoredCode: completion.Choices[0].Message.Content,
		Changes:        a.extractChanges(completion.Choices[0].Message.Content),
		TokensUsed:     int64(completion.Usage.TotalTokens),
	}, nil
}

func (a *CodeGeneratorAdapter) EstimateCost(ctx context.Context, req ports.GenerateCodeRequest) (*ports.CostEstimateResponse, error) {
	prompt := a.buildCodeGenerationPrompt(req)
	
	estimate, err := a.client.EstimateCost(CompletionRequest{
		Model: "anthropic/claude-3.5-sonnet",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens: 4096,
	})

	if err != nil {
		return nil, fmt.Errorf("cost estimation failed: %w", err)
	}

	return &ports.CostEstimateResponse{
		EstimatedTokens: int64(estimate.InputTokens + estimate.OutputTokens),
		EstimatedCost:   estimate.EstimatedCost,
		Confidence:      estimate.ConfidenceLevel,
	}, nil
}

func (a *CodeGeneratorAdapter) ExplainDecision(ctx context.Context, decision string) (*ports.ExplanationResponse, error) {
	prompt := fmt.Sprintf("Explain the reasoning behind this decision: %s\n\nProvide alternatives and key considerations.", decision)

	completion, err := a.client.Complete(ctx, CompletionRequest{
		Model: "anthropic/claude-3.5-sonnet",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   1024,
		Temperature: 0.7,
	})

	if err != nil {
		return nil, fmt.Errorf("decision explanation failed: %w", err)
	}

	content := completion.Choices[0].Message.Content
	
	return &ports.ExplanationResponse{
		Reasoning:      content,
		Confidence:     0.85,
		Alternatives:   a.extractAlternatives(content),
		Considerations: a.extractConsiderations(content),
	}, nil
}

func (a *CodeGeneratorAdapter) ValidateOutput(ctx context.Context, code string, language string) (*ports.ValidationResponse, error) {
	prompt := fmt.Sprintf("Validate this %s code for correctness, best practices, and potential issues:\n\n%s", language, code)

	completion, err := a.client.Complete(ctx, CompletionRequest{
		Model: "anthropic/claude-3.5-sonnet",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   1024,
		Temperature: 0.3,
	})

	if err != nil {
		return nil, fmt.Errorf("output validation failed: %w", err)
	}

	content := completion.Choices[0].Message.Content
	isValid := !strings.Contains(strings.ToLower(content), "error") && !strings.Contains(strings.ToLower(content), "invalid")

	return &ports.ValidationResponse{
		IsValid:      isValid,
		Issues:       a.extractIssues(content),
		Suggestions:  a.extractSuggestions(content),
		QualityScore: a.calculateQualityScore(content),
	}, nil
}

func (a *CodeGeneratorAdapter) buildCodeGenerationPrompt(req ports.GenerateCodeRequest) string {
	return fmt.Sprintf(`Generate %s code for the following specification:

%s

Framework: %s
Additional context: %v

Provide clean, production-ready code with proper error handling.`,
		req.Language, req.Specification, req.Framework, req.Context)
}

func (a *CodeGeneratorAdapter) buildCodeGenerationPromptWithContext(req ports.GenerateCodeRequest, learningContext interface{}) string {
	contextStr := fmt.Sprintf("%v", learningContext)
	return fmt.Sprintf(`Generate %s code using learned patterns and insights:

Specification: %s
Framework: %s
Learning Context: %s

Apply relevant patterns from the learning context to generate optimal code.`,
		req.Language, req.Specification, req.Framework, contextStr)
}

func (a *CodeGeneratorAdapter) extractDependencies(code, language string) []string {
	deps := make([]string, 0)
	lines := strings.Split(code, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if language == "Go" && strings.HasPrefix(line, "import") {
			deps = append(deps, line)
		} else if language == "JavaScript" && (strings.HasPrefix(line, "import") || strings.HasPrefix(line, "require")) {
			deps = append(deps, line)
		}
	}
	
	return deps
}

func (a *CodeGeneratorAdapter) extractTechnologies(content string) []string {
	techs := make([]string, 0)
	commonTechs := []string{"React", "Vue", "Angular", "Node.js", "Go", "Python", "Docker", "Kubernetes", "PostgreSQL", "MongoDB"}
	
	for _, tech := range commonTechs {
		if strings.Contains(content, tech) {
			techs = append(techs, tech)
		}
	}
	
	return techs
}

func (a *CodeGeneratorAdapter) extractSummary(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return "Analysis completed"
}

func (a *CodeGeneratorAdapter) extractChanges(content string) []string {
	changes := make([]string, 0)
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		if strings.Contains(line, "Changed:") || strings.Contains(line, "Modified:") || strings.Contains(line, "Refactored:") {
			changes = append(changes, strings.TrimSpace(line))
		}
	}
	
	if len(changes) == 0 {
		changes = append(changes, "Code refactored successfully")
	}
	
	return changes
}

func (a *CodeGeneratorAdapter) extractAlternatives(content string) []string {
	alternatives := make([]string, 0)
	if strings.Contains(content, "alternative") || strings.Contains(content, "Alternative") {
		alternatives = append(alternatives, "Alternative approaches mentioned in reasoning")
	}
	return alternatives
}

func (a *CodeGeneratorAdapter) extractConsiderations(content string) []string {
	considerations := make([]string, 0)
	if strings.Contains(content, "consider") || strings.Contains(content, "Consider") {
		considerations = append(considerations, "Key considerations outlined in explanation")
	}
	return considerations
}

func (a *CodeGeneratorAdapter) extractIssues(content string) []string {
	issues := make([]string, 0)
	keywords := []string{"error", "issue", "problem", "bug", "warning"}
	
	for _, keyword := range keywords {
		if strings.Contains(strings.ToLower(content), keyword) {
			issues = append(issues, fmt.Sprintf("Potential %s detected", keyword))
		}
	}
	
	return issues
}

func (a *CodeGeneratorAdapter) extractSuggestions(content string) []string {
	suggestions := make([]string, 0)
	if strings.Contains(content, "suggest") || strings.Contains(content, "recommend") {
		suggestions = append(suggestions, "Improvements suggested in validation")
	}
	return suggestions
}

func (a *CodeGeneratorAdapter) calculateQualityScore(content string) float64 {
	score := 0.8
	
	if strings.Contains(strings.ToLower(content), "excellent") {
		score = 0.95
	} else if strings.Contains(strings.ToLower(content), "good") {
		score = 0.85
	} else if strings.Contains(strings.ToLower(content), "error") {
		score = 0.4
	}
	
	return score
}

func (a *CodeGeneratorAdapter) SetFallbackStrategy(strategy *FallbackStrategy) {
	a.strategy = strategy
}

func (a *CodeGeneratorAdapter) GetClient() *Client {
	return a.client
}
