package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/istok/agent-core/internal/application/dto"
	"github.com/istok/agent-core/internal/domain"
	"github.com/istok/agent-core/internal/ports"
)

// ProjectGeneratorService - сервис генерации проектов
type ProjectGeneratorService struct {
	agent              *domain.Agent
	codeGenerator      ports.CodeGenerator
	webCrawler         ports.WebCrawler
	intelligenceService *domain.AgentIntelligenceService
}

// NewProjectGeneratorService создает новый сервис генерации
func NewProjectGeneratorService(
	agent *domain.Agent,
	codeGenerator ports.CodeGenerator,
	webCrawler ports.WebCrawler,
) *ProjectGeneratorService {
	return &ProjectGeneratorService{
		agent:              agent,
		codeGenerator:      codeGenerator,
		webCrawler:         webCrawler,
		intelligenceService: domain.NewAgentIntelligenceService(),
	}
}

// GenerateProject генерирует проект на основе спецификации
func (s *ProjectGeneratorService) GenerateProject(ctx context.Context, req dto.GenerateProjectRequest) (*dto.GenerateProjectResponse, error) {
	startTime := time.Now()

	// Если указан URL для анализа, сначала анализируем его
	if req.AnalyzeURL != "" {
		if err := s.analyzeCompetitor(ctx, req.AnalyzeURL); err != nil {
			// Логируем ошибку, но продолжаем генерацию
			fmt.Printf("Предупреждение: не удалось проанализировать URL %s: %v\n", req.AnalyzeURL, err)
		}
	}

	// Создаем задачу генерации
	task := domain.NewTask(
		s.agent.ID,
		"code_generation",
		fmt.Sprintf("Генерация проекта: %s", req.Specification),
		8, // высокий приоритет
		5000, // оценка токенов
	)

	s.agent.EnqueueTask(task)
	s.agent.UpdateStatus(domain.StatusCoding)

	// Оцениваем риск
	riskScore, riskReason := s.intelligenceService.EvaluateRisk(s.agent, task)
	if riskScore > 0.8 {
		return nil, fmt.Errorf("высокий риск выполнения задачи: %s", riskReason)
	}

	// Получаем рекомендацию стратегии
	strategy, confidence := s.intelligenceService.RecommendStrategy(s.agent, "code_generation")
	
	// Записываем решение
	s.agent.RecordDecision(
		task.ID,
		fmt.Sprintf("Использовать стратегию: %s", strategy),
		fmt.Sprintf("Рекомендация основана на анализе контекста обучения. Уверенность: %.2f", confidence),
		confidence,
	)

	// Подготавливаем запрос к генератору кода
	genReq := ports.GenerateCodeRequest{
		Specification: req.Specification,
		Language:      req.Language,
		Framework:     req.Framework,
		Context:       s.buildContextFromLearning(),
	}

	// Оцениваем стоимость
	costEstimate, err := s.codeGenerator.EstimateCost(ctx, genReq)
	if err != nil {
		return nil, fmt.Errorf("ошибка оценки стоимости: %w", err)
	}

	// Проверяем баланс токенов
	if !s.agent.CanExecuteTask(costEstimate.EstimatedTokens) {
		return nil, fmt.Errorf("недостаточно токенов: требуется %d, доступно %d", 
			costEstimate.EstimatedTokens, s.agent.TokenBalance)
	}

	// Генерируем код с использованием контекста обучения
	var response *ports.GenerateCodeResponse
	if s.agent.LearningContext.TotalNodes > 0 {
		// Используем контекст обучения для улучшенной генерации
		response, err = s.codeGenerator.GenerateWithContext(ctx, genReq, s.agent.LearningContext)
	} else {
		// Стандартная генерация
		response, err = s.codeGenerator.GenerateCode(ctx, genReq)
	}

	if err != nil {
		s.agent.RecordTaskFailure()
		s.agent.UpdateStatus(domain.StatusError)
		task.Fail(err.Error())
		return nil, fmt.Errorf("ошибка генерации кода: %w", err)
	}

	// Списываем токены
	if err := s.agent.DeductTokens(response.TokensUsed); err != nil {
		return nil, fmt.Errorf("ошибка списания токенов: %w", err)
	}

	// Записываем успех
	duration := time.Since(startTime)
	s.agent.RecordTaskSuccess(response.TokensUsed, duration)
	s.agent.UpdateStatus(domain.StatusIdle)
	task.Complete(map[string]interface{}{
		"code":         response.Code,
		"tokens_used":  response.TokensUsed,
		"dependencies": response.Dependencies,
	})

	return &dto.GenerateProjectResponse{
		Code:         response.Code,
		Explanation:  response.Explanation,
		TokensUsed:   response.TokensUsed,
		Dependencies: response.Dependencies,
		Model:        "claude-3.5-sonnet", // TODO: получать из response
	}, nil
}

// analyzeCompetitor анализирует сайт конкурента и добавляет знания в контекст
func (s *ProjectGeneratorService) analyzeCompetitor(ctx context.Context, url string) error {
	s.agent.UpdateStatus(domain.StatusAnalyzing)

	// Используем web crawler для анализа
	crawlReq := ports.CrawlRequest{
		URL:   url,
		Depth: 1,
	}

	crawlResp, err := s.webCrawler.CrawlWebsite(ctx, crawlReq)
	if err != nil {
		return fmt.Errorf("ошибка crawling: %w", err)
	}

	// Создаем snapshot для обучения
	snapshot := &domain.WebsiteSnapshot{
		ID:           domain.GenerateID(),
		URL:          url,
		Title:        crawlResp.Title,
		Technologies: crawlResp.Technologies,
		Structure:    crawlResp.Structure,
		Confidence:   crawlResp.Confidence,
		AnalyzedAt:   time.Now(),
	}

	// Проверяем, можем ли учиться от этого сайта
	canLearn, reason := s.intelligenceService.CanLearnFrom(s.agent, snapshot)
	if !canLearn {
		return fmt.Errorf("невозможно обучиться от сайта: %s", reason)
	}

	// Добавляем знания в контекст обучения
	s.agent.LearnFromWebsite(snapshot)

	// Добавляем паттерны из анализа
	for _, pattern := range crawlResp.Patterns {
		s.agent.AddPattern(pattern)
	}

	// Добавляем инсайты
	for _, insight := range crawlResp.Insights {
		s.agent.AddInsight(insight)
	}

	fmt.Printf("✓ Успешно проанализирован сайт: %s (узлов знаний: %d)\n", url, s.agent.GetKnowledgeNodeCount())

	return nil
}

// buildContextFromLearning создает контекст для генерации из накопленных знаний
func (s *ProjectGeneratorService) buildContextFromLearning() map[string]interface{} {
	context := make(map[string]interface{})

	if s.agent.LearningContext.TotalNodes > 0 {
		context["knowledge_nodes"] = s.agent.LearningContext.TotalNodes
		context["learning_confidence"] = s.agent.LearningContext.Confidence
		
		// Добавляем популярные технологии
		techNodes := s.agent.LearningContext.GetNodesByType(domain.NodeTypeTechnology)
		technologies := make([]string, 0)
		for _, node := range techNodes {
			technologies = append(technologies, node.Label)
		}
		context["learned_technologies"] = technologies

		// Добавляем паттерны
		patterns := make([]string, 0)
		for _, pattern := range s.agent.LearningContext.Patterns {
			patterns = append(patterns, pattern.Name)
		}
		context["learned_patterns"] = patterns

		// Добавляем действенные инсайты
		insights := s.agent.LearningContext.GetActionableInsights()
		insightTitles := make([]string, 0)
		for _, insight := range insights {
			insightTitles = append(insightTitles, insight.Title)
		}
		context["actionable_insights"] = insightTitles
	}

	return context
}

// GetAgent возвращает агента для доступа к статистике
func (s *ProjectGeneratorService) GetAgent() *domain.Agent {
	return s.agent
}
