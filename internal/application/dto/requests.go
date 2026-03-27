package dto

// GenerateProjectRequest - запрос на генерацию проекта
type GenerateProjectRequest struct {
	Specification string `json:"specification"`
	URL           string `json:"url"`
	Language      string `json:"language"`
	Framework     string `json:"framework"`
	AnalyzeURL    string `json:"analyze_url,omitempty"`
}

// AnalyzeWebsiteRequest - запрос на анализ сайта
type AnalyzeWebsiteRequest struct {
	URL          string `json:"url"`
	AnalysisType string `json:"analysis_type"`
	Depth        int    `json:"depth"`
}
