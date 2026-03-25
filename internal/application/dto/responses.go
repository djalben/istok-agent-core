package dto

// GenerateProjectResponse - ответ с сгенерированным проектом
type GenerateProjectResponse struct {
	Code         string   `json:"code"`
	Explanation  string   `json:"explanation"`
	TokensUsed   int64    `json:"tokens_used"`
	Dependencies []string `json:"dependencies"`
	Model        string   `json:"model"`
}

// AnalyzeWebsiteResponse - ответ с анализом сайта
type AnalyzeWebsiteResponse struct {
	URL          string                 `json:"url"`
	Technologies []string               `json:"technologies"`
	Patterns     []PatternDTO           `json:"patterns"`
	Insights     []InsightDTO           `json:"insights"`
	Summary      string                 `json:"summary"`
	Confidence   float64                `json:"confidence"`
}

// AgentStatsResponse - статистика агента
type AgentStatsResponse struct {
	AgentID              string  `json:"agent_id"`
	Name                 string  `json:"name"`
	Status               string  `json:"status"`
	TokenBalance         int64   `json:"token_balance"`
	TotalTasks           int64   `json:"total_tasks"`
	SuccessRate          float64 `json:"success_rate"`
	KnowledgeNodes       int     `json:"knowledge_nodes"`
	LearningConfidence   float64 `json:"learning_confidence"`
	AverageTokensPerTask float64 `json:"average_tokens_per_task"`
}

// PatternDTO - паттерн для передачи
type PatternDTO struct {
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// InsightDTO - инсайт для передачи
type InsightDTO struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Confidence  float64 `json:"confidence"`
	Priority    int     `json:"priority"`
}

// StreamChunk - чанк для streaming ответа
type StreamChunk struct {
	Type     string                 `json:"type"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
