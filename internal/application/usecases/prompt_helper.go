package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/djalben/istok-agent-core/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
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
		return "", ErrEmptyPrompt
	}

	finalPrompt := userPrompt
	if referenceURL != "" {
		finalPrompt = fmt.Sprintf("%s\n\nURL-референс конкурента для анализа визуального стиля и структуры: %s", userPrompt, referenceURL)
	}

	start := time.Now()
	l := ports.LoggerFromContext(ctx)
	l.InfoContext(ctx, "prompt helper enhance started",
		"promptChars", len(userPrompt),
		"referenceUrl", referenceURL,
	)

	resp, err := ph.llm.Complete(ctx, ports.LLMRequest{
		Model:        "anthropic/claude-sonnet-4-5",
		SystemPrompt: promptHelperSystemInstruction,
		UserPrompt:   finalPrompt,
		MaxTokens:    4096,
		Temperature:  0.7,
		Reasoning:    true,
		Effort:       ports.EffortMedium,
	})
	if err != nil {
		return "", wrapper.Wrap(err)
	}
	l.InfoContext(ctx, "prompt helper enhance complete",
		"duration", time.Since(start).Round(time.Millisecond),
		"inputChars", len(userPrompt),
		"outputChars", len(resp.Content),
	)

	return resp.Content, nil
}
