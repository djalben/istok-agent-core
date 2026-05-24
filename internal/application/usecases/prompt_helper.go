package usecases

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/istok/agent-core/internal/ports"
)

const promptHelperSystemInstruction = `Ты — топовый Product Manager с 15-летним опытом запуска цифровых продуктов.

Твоя задача — превратить короткую идею пользователя (и опциональный URL конкурента) в красивый, деловой бизнес-бриф для веб-приложения.

## СТРОГИЕ ЗАПРЕТЫ:
- ЗАПРЕЩЕНО писать структуру БД, таблицы, SQL, миграции
- ЗАПРЕЩЕНО писать API-контракты, эндпоинты, HTTP-методы
- ЗАПРЕЩЕНО писать код, теги, JSON, XML
- ЗАПРЕЩЕНО использовать технический жаргон (FK, indexes, CRUD, REST, WCAG)
- ЗАПРЕЩЕНО выводить теги размышлений, секции [THOUGHT], [SELF-CORRECTION] и подобные

## ФОРМАТ ОТВЕТА (строго на русском языке):

# 🎯 [Название проекта]

## 1. Бизнес-концепция
Что это за продукт, какую проблему решает, для кого предназначен (целевая аудитория, их боли и потребности). 2-3 абзаца живым деловым языком.

## 2. Визуальный стиль и тон
Цветовая палитра (конкретные цвета), типографика, общее настроение интерфейса (минимализм / яркость / корпоративность), референсы стиля. Если пользователь дал URL конкурента — проанализируй его визуальный стиль и предложи улучшения.

## 3. Ключевые экраны и функционал
Перечисли 5-10 основных экранов приложения с кратким описанием того, что пользователь видит и делает на каждом. Используй понятный язык, а не технические термины.

## ПРАВИЛА:
- Пиши на русском языке
- Используй деловой, понятный неспециалисту язык
- Будь конкретным: называй цвета, описывай элементы интерфейса
- Ответ должен быть готов к утверждению заказчиком — это бизнес-документ, не техническое ТЗ`

// PromptHelper enhances user prompts into structured specifications.
type PromptHelper struct {
	llm ports.LLMProvider
}

// NewPromptHelper creates a new PromptHelper instance.
func NewPromptHelper(llm ports.LLMProvider) *PromptHelper {
	return &PromptHelper{llm: llm}
}

// Enhance takes a brief user idea and optional competitor URL, returns a structured specification.
func (ph *PromptHelper) Enhance(ctx context.Context, userPrompt string, referenceURL string) (string, error) {
	if userPrompt == "" {
		return "", fmt.Errorf("empty prompt")
	}

	finalPrompt := userPrompt
	if referenceURL != "" {
		finalPrompt = fmt.Sprintf("%s\n\nURL-референс конкурента для анализа визуального стиля и структуры: %s", userPrompt, referenceURL)
	}

	start := time.Now()
	log.Printf("🪄 PromptHelper: enhancing prompt (%d chars, ref=%q)", len(userPrompt), referenceURL)

	resp, err := ph.llm.Complete(ctx, ports.LLMRequest{
		Model:          "anthropic/claude-sonnet-4-6-thinking",
		SystemPrompt:   promptHelperSystemInstruction,
		UserPrompt:     finalPrompt,
		MaxTokens:      4096,
		Temperature:    0.7,
		Reasoning:      true,
	})
	if err != nil {
		return "", fmt.Errorf("prompt enhance LLM call failed: %w", err)
	}

	log.Printf("✅ PromptHelper: enhanced in %s (%d chars → %d chars)",
		time.Since(start).Round(time.Millisecond), len(userPrompt), len(resp.Content))

	return resp.Content, nil
}
