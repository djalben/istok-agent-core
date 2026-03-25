package usecases

import (
	"context"
	"fmt"

	"github.com/istok/agent-core/internal/domain"
)

// ReasoningService сервис для размышлений агента перед генерацией
type ReasoningService struct {
	reasoningEngine *domain.ReasoningEngine
}

// NewReasoningService создает новый сервис размышлений
func NewReasoningService(agent *domain.Agent) *ReasoningService {
	return &ReasoningService{
		reasoningEngine: domain.NewReasoningEngine(agent),
	}
}

// ReasonAboutTask выполняет размышление над задачей
func (rs *ReasoningService) ReasonAboutTask(ctx context.Context, task *domain.Task) (*domain.ReasoningChain, error) {
	fmt.Printf("🧠 Запуск размышления над задачей: %s\n", task.Description)
	
	chain, err := rs.reasoningEngine.Reason(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("ошибка размышления: %w", err)
	}
	
	fmt.Printf("✅ Размышление завершено. Уверенность: %.2f%%\n", chain.TotalConfidence*100)
	fmt.Printf("📋 Заключение: %s\n", chain.FinalConclusion)
	
	return chain, nil
}

// GetReasoningSummary возвращает краткое резюме размышлений
func (rs *ReasoningService) GetReasoningSummary(chain *domain.ReasoningChain) string {
	return chain.GetSummary()
}
