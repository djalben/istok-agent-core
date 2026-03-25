package domain

import (
	"context"
	"fmt"
	"time"
)

// ReasoningStep представляет шаг размышления агента
type ReasoningStep struct {
	ID         string
	StepNumber int
	Type       ReasoningStepType
	Question   string
	Answer     string
	Confidence float64
	Duration   time.Duration
	Timestamp  time.Time
}

// ReasoningStepType тип шага размышления
type ReasoningStepType string

const (
	ReasoningTypeAnalysis      ReasoningStepType = "analysis"      // Анализ задачи
	ReasoningTypeDecomposition ReasoningStepType = "decomposition" // Декомпозиция
	ReasoningTypePlanning      ReasoningStepType = "planning"      // Планирование
	ReasoningTypeValidation    ReasoningStepType = "validation"    // Валидация решения
	ReasoningTypeOptimization  ReasoningStepType = "optimization"  // Оптимизация
)

// ReasoningChain цепочка размышлений агента
type ReasoningChain struct {
	ID              string
	TaskID          string
	Steps           []*ReasoningStep
	FinalConclusion string
	TotalConfidence float64
	TotalDuration   time.Duration
	StartedAt       time.Time
	CompletedAt     time.Time
}

// NewReasoningChain создает новую цепочку размышлений
func NewReasoningChain(taskID string) *ReasoningChain {
	return &ReasoningChain{
		ID:        GenerateID(),
		TaskID:    taskID,
		Steps:     make([]*ReasoningStep, 0),
		StartedAt: time.Now(),
	}
}

// AddStep добавляет шаг размышления
func (rc *ReasoningChain) AddStep(stepType ReasoningStepType, question, answer string, confidence float64, duration time.Duration) {
	step := &ReasoningStep{
		ID:         GenerateID(),
		StepNumber: len(rc.Steps) + 1,
		Type:       stepType,
		Question:   question,
		Answer:     answer,
		Confidence: confidence,
		Duration:   duration,
		Timestamp:  time.Now(),
	}
	rc.Steps = append(rc.Steps, step)
}

// Complete завершает цепочку размышлений
func (rc *ReasoningChain) Complete(conclusion string) {
	rc.FinalConclusion = conclusion
	rc.CompletedAt = time.Now()
	rc.TotalDuration = rc.CompletedAt.Sub(rc.StartedAt)

	// Вычисляем общую уверенность как среднее
	if len(rc.Steps) > 0 {
		sum := 0.0
		for _, step := range rc.Steps {
			sum += step.Confidence
		}
		rc.TotalConfidence = sum / float64(len(rc.Steps))
	}
}

// GetSummary возвращает краткое резюме размышлений
func (rc *ReasoningChain) GetSummary() string {
	summary := fmt.Sprintf("Размышление завершено за %v. Шагов: %d. Уверенность: %.2f%%\n\n",
		rc.TotalDuration, len(rc.Steps), rc.TotalConfidence*100)

	for _, step := range rc.Steps {
		summary += fmt.Sprintf("Шаг %d (%s):\nВопрос: %s\nОтвет: %s\nУверенность: %.2f%%\n\n",
			step.StepNumber, step.Type, step.Question, step.Answer, step.Confidence*100)
	}

	summary += fmt.Sprintf("Заключение: %s", rc.FinalConclusion)
	return summary
}

// ReasoningEngine движок размышлений агента
type ReasoningEngine struct {
	agent *Agent
}

// NewReasoningEngine создает новый движок размышлений
func NewReasoningEngine(agent *Agent) *ReasoningEngine {
	return &ReasoningEngine{
		agent: agent,
	}
}

// Reason выполняет размышление над задачей
func (re *ReasoningEngine) Reason(ctx context.Context, task *Task) (*ReasoningChain, error) {
	chain := NewReasoningChain(task.ID)

	// Шаг 1: Анализ задачи
	analysisStart := time.Now()
	analysisAnswer := re.analyzeTask(task)
	chain.AddStep(
		ReasoningTypeAnalysis,
		"Что требуется сделать и какова сложность задачи?",
		analysisAnswer,
		0.9,
		time.Since(analysisStart),
	)

	// Шаг 2: Декомпозиция
	decompositionStart := time.Now()
	decompositionAnswer := re.decomposeTask(task)
	chain.AddStep(
		ReasoningTypeDecomposition,
		"Как разбить задачу на подзадачи?",
		decompositionAnswer,
		0.85,
		time.Since(decompositionStart),
	)

	// Шаг 3: Планирование
	planningStart := time.Now()
	planningAnswer := re.planExecution(task)
	chain.AddStep(
		ReasoningTypePlanning,
		"Какой оптимальный план выполнения?",
		planningAnswer,
		0.88,
		time.Since(planningStart),
	)

	// Шаг 4: Валидация
	validationStart := time.Now()
	validationAnswer := re.validateApproach(task)
	chain.AddStep(
		ReasoningTypeValidation,
		"Корректен ли выбранный подход?",
		validationAnswer,
		0.92,
		time.Since(validationStart),
	)

	// Шаг 5: Оптимизация
	optimizationStart := time.Now()
	optimizationAnswer := re.optimizeStrategy(task)
	chain.AddStep(
		ReasoningTypeOptimization,
		"Как оптимизировать решение?",
		optimizationAnswer,
		0.87,
		time.Since(optimizationStart),
	)

	// Формируем заключение
	conclusion := fmt.Sprintf(
		"Задача '%s' проанализирована. Рекомендуется использовать поэтапный подход с акцентом на %s. "+
			"Ожидаемое время выполнения: ~%d секунд. Риски минимальны.",
		task.Description,
		task.Type,
		task.TokenCost/100, // грубая оценка
	)

	chain.Complete(conclusion)
	return chain, nil
}

// analyzeTask анализирует задачу
func (re *ReasoningEngine) analyzeTask(task *Task) string {
	complexity := "средняя"
	if task.TokenCost > 10000 {
		complexity = "высокая"
	} else if task.TokenCost < 3000 {
		complexity = "низкая"
	}

	return fmt.Sprintf(
		"Задача типа '%s' с описанием: '%s'. Сложность: %s. "+
			"Требуется генерация кода с использованием контекста обучения агента. "+
			"Приоритет: %d/10.",
		task.Type, task.Description, complexity, task.Priority,
	)
}

// decomposeTask разбивает задачу на подзадачи
func (re *ReasoningEngine) decomposeTask(task *Task) string {
	return fmt.Sprintf(
		"Подзадачи:\n" +
			"1. Анализ спецификации и извлечение требований\n" +
			"2. Поиск релевантных паттернов в базе знаний\n" +
			"3. Генерация архитектуры проекта\n" +
			"4. Создание компонентов и логики\n" +
			"5. Валидация и оптимизация кода",
	)
}

// planExecution планирует выполнение
func (re *ReasoningEngine) planExecution(task *Task) string {
	strategy := "стандартная генерация"
	if re.agent.LearningContext.TotalNodes > 0 {
		strategy = "генерация с использованием накопленных знаний"
	}

	return fmt.Sprintf(
		"План:\n"+
			"1. Использовать стратегию: %s\n"+
			"2. Модель: Claude 3.5 Sonnet (или fallback на GPT-4o)\n"+
			"3. Применить паттерны из %d узлов знаний\n"+
			"4. Оценить стоимость перед выполнением\n"+
			"5. Записать результат в контекст обучения",
		strategy, re.agent.LearningContext.TotalNodes,
	)
}

// validateApproach валидирует подход
func (re *ReasoningEngine) validateApproach(task *Task) string {
	hasBalance := re.agent.TokenBalance >= task.TokenCost
	hasCapability := re.agent.HasCapability("code_generation")

	validation := "Подход корректен. "
	if !hasBalance {
		validation += "ВНИМАНИЕ: недостаточно токенов. "
	}
	if !hasCapability {
		validation += "ВНИМАНИЕ: отсутствует capability для генерации кода. "
	}

	if hasBalance && hasCapability {
		validation += "Все проверки пройдены успешно."
	}

	return validation
}

// optimizeStrategy оптимизирует стратегию
func (re *ReasoningEngine) optimizeStrategy(task *Task) string {
	optimizations := []string{
		"Использовать кэш для повторяющихся паттернов",
		"Применить инкрементальную генерацию для больших проектов",
		"Переиспользовать компоненты из базы знаний",
	}

	result := "Рекомендуемые оптимизации:\n"
	for i, opt := range optimizations {
		result += fmt.Sprintf("%d. %s\n", i+1, opt)
	}

	return result
}
