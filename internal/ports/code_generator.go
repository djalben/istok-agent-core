package ports

import "context"

type GenerateCodeRequest struct {
	Specification string
	Language      string
	Framework     string
	Context       map[string]interface{}
}

type GenerateCodeResponse struct {
	Code         string
	Explanation  string
	TokensUsed   int64
	Dependencies []string
}

type AnalyzeWebsiteRequest struct {
	URL          string
	AnalysisType string
	Depth        int
}

type AnalyzeWebsiteResponse struct {
	Structure    map[string]interface{}
	Technologies []string
	Summary      string
	TokensUsed   int64
}

type RefactorCodeRequest struct {
	Code          string
	Language      string
	Instructions  string
	TargetPattern string
}

type RefactorCodeResponse struct {
	RefactoredCode string
	Changes        []string
	TokensUsed     int64
}

type CodeGenerator interface {
	GenerateCode(ctx context.Context, req GenerateCodeRequest) (*GenerateCodeResponse, error)
	GenerateWithContext(ctx context.Context, req GenerateCodeRequest, learningContext interface{}) (*GenerateCodeResponse, error)
	AnalyzeWebsite(ctx context.Context, req AnalyzeWebsiteRequest) (*AnalyzeWebsiteResponse, error)
	RefactorCode(ctx context.Context, req RefactorCodeRequest) (*RefactorCodeResponse, error)
	EstimateCost(ctx context.Context, req GenerateCodeRequest) (*CostEstimateResponse, error)
	ExplainDecision(ctx context.Context, decision string) (*ExplanationResponse, error)
	ValidateOutput(ctx context.Context, code string, language string) (*ValidationResponse, error)
}

type CostEstimateResponse struct {
	EstimatedTokens int64
	EstimatedCost   float64
	Confidence      float64
}

type ExplanationResponse struct {
	Reasoning      string
	Confidence     float64
	Alternatives   []string
	Considerations []string
}

type ValidationResponse struct {
	IsValid      bool
	Issues       []string
	Suggestions  []string
	QualityScore float64
}
