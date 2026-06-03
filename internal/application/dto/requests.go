package dto

// GenerateProjectRequest - запрос на генерацию проекта
type GenerateProjectRequest struct {
	Specification string   `json:"specification"`
	URL           string   `json:"url"`
	Language      string   `json:"language"`
	Framework     string   `json:"framework"`
	AnalyzeURL    string   `json:"analyze_url,omitempty"`
	Mode          string   `json:"mode,omitempty"`           // "agent" (deep reasoning) | "code" (fast UI)
	SessionID     string   `json:"session_id,omitempty"`     // unique session for checkpoint/resume
	Resume        bool     `json:"resume,omitempty"`         // true = resume from last checkpoint
	ExistingFiles []string `json:"existing_files,omitempty"` // files already received by client
	ProjectID     string   `json:"project_id,omitempty"`     // Layer 2: если задан — обновляем существующий проект, иначе создаём новый
	Name          string   `json:"name,omitempty"`           // Layer 2: имя проекта для авто-сохранения в БД
}

// AnalyzeWebsiteRequest - запрос на анализ сайта
type AnalyzeWebsiteRequest struct {
	URL          string `json:"url"`
	AnalysisType string `json:"analysis_type"`
	Depth        int    `json:"depth"`
}
