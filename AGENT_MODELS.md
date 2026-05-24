# Матрица ИИ-компетенций ИСТОКА

> Техническая карта распределения моделей по агентам пайплайна Istok Core.
> Последнее обновление: 2025-05-24

## Инфраструктура маршрутизации

Все LLM-запросы проходят через `DualRouter` (`internal/infrastructure/llm/router.go`):

| Адаптер | Endpoint | Модели | Назначение |
|---------|----------|--------|------------|
| **AnthropicAdapter** | `https://api.anthropic.com/v1` | `anthropic/*`, `claude-*` | Reasoning, code generation, planning |
| **ReplicateAdapter** | `https://api.replicate.com/v1` | `google/*`, `black-forest-labs/*`, `ideogram-ai/*`, `stability-ai/*` | Media generation (images, video) |

Маршрутизация по префиксу модели: `anthropic/` → Anthropic Direct API, остальное → Replicate.

---

## Матрица агентов

| # | Агент | Роль в пайплайне | Модель | Адаптер | Тип задачи | Timeout | Критичность |
|---|-------|------------------|--------|---------|-------------|---------|-------------|
| 1 | **Director** | Стратегическое планирование | `claude-sonnet-4-6-thinking` | Anthropic | Формирование DAG-плана генерации, координация пайплайна | 5 мин | 🔴 Критичен — без плана ни один агент не стартует |
| 2 | **Researcher** | Визуальный и технический аудит | `claude-sonnet-4-6-thinking` | Anthropic | Анализ URL конкурента: цвета, шрифты, компоненты, layout, CSS-переменные | 5 мин | 🟡 Важен для Synthesis-режима, пропускается в Agent-режиме без URL |
| 3 | **Brain** | Архитектура системы | `claude-sonnet-4-6-thinking` | Anthropic | SystemManifest: endpoints, DB-схема, FileMap, стек технологий | 10 мин | 🔴 Критичен — определяет всю структуру проекта |
| 4 | **Planner** | DAG-планирование с контекстом | `claude-sonnet-4-6-thinking` | Anthropic | Инъекция package.json/tsconfig.json, FSM gate validation, tier-разбивка | 5 мин* | 🔴 Критичен — DAG-тиры управляют параллельной генерацией |
| 5 | **Coder** | Генерация кода | `claude-sonnet-4-6` | Anthropic | Chunked multi-file code generation (до 112 файлов, 6 тиров) | 10 мин | 🔴 Критичен — основной генератор итогового ZIP |
| 6 | **Designer** | Визуальные ассеты | `google/nano-banana` | Replicate | Генерация UI-ассетов, иконок, фоновых изображений | 5 мин | 🟡 Опционален — проект работает без ассетов |
| 7 | **Validator** | Синтаксическая верификация | `claude-sonnet-4-6` | Anthropic | Валидация HTML/CSS/JS синтаксиса, проверка runtime-ошибок | 3 мин | 🟡 Важен — ловит критические ошибки до поставки |
| 8 | **Security** | Аудит безопасности | Детерминированный (rule-based) | — | XSS, injection, unsafe patterns, CSP-проверки | 30 сек | 🟢 Информационный — не блокирует поставку |
| 9 | **Tester** | Функциональное тестирование | Детерминированный (rule-based) | — | Структурная валидация, integrity-check файлов | 30 сек | 🟢 Информационный — не блокирует поставку |
| 10 | **UI Reviewer** | UX/Accessibility аудит | Детерминированный (rule-based) | — | Проверка a11y, контрастности, отзывчивости макета | 30 сек | 🟢 Информационный — не блокирует поставку |
| 11 | **Videographer** | Промо-видео | `google/veo-3` | Replicate | Генерация промо-ролика проекта | 15 мин | 🟢 Опционален — отдельный медиа-ассет |

> \* Planner использует модель Director'а через `PlannerAgent` — конфигурируется при создании: `NewPlannerAgent(llm, "anthropic/claude-sonnet-4-6-thinking")`.

---

## Модели и версии

| Модель | API-идентификатор | Особенности |
|--------|-------------------|-------------|
| **Claude Sonnet 4.6 (Adaptive Thinking)** | `anthropic/claude-sonnet-4-6-thinking` | Adaptive Thinking API `{type: "adaptive", effort: "high"}`, max 128K output tokens |
| **Claude Sonnet 4.6** | `anthropic/claude-sonnet-4-6` | Стандартный режим без цепочки рассуждений, max 128K output tokens |
| **Nano Banana** | `google/nano-banana` | Replicate image generation (2048×2048, $0.01/img) |
| **Veo 3** | `google/veo-3` | Replicate video generation |

---

## Медиа-генерация (ImageGeneratorAdapter)

Отдельный адаптер `internal/infrastructure/media/image_generator.go`:

| Модель | Провайдер | Endpoint | Стоимость | Max Resolution |
|--------|-----------|----------|-----------|----------------|
| `nano-banana-2` | Replicate | `api.replicate.com/v1/predictions` | $0.01/img | 2048×2048 |
| `gemini-flash-image` | Google | `generativelanguage.googleapis.com` | $0.005/img | 1024×1024 |

---

## Порядок выполнения пайплайна

```
Director → Researcher → Brain (Architect) → Planner → Coder (6 тиров) → Designer → Validator → [Security + Tester + UI Reviewer] → Videographer
```

**Partial Delivery:** файлы стримятся клиенту через EventBus сразу после Coder, не дожидаясь Verification Layer.

**Verification Gate:** Security, Tester, UI Reviewer работают параллельно (rule-based, без LLM). Deadline: 30 секунд. При таймауте — авто-approve.

---

## Диаграмма зависимостей

```
┌─────────────┐     ┌──────────────┐     ┌───────────────┐
│  Director   │────▶│  Researcher  │────▶│  Brain        │
│  (thinking) │     │  (thinking)  │     │  (thinking)   │
└─────────────┘     └──────────────┘     └───────┬───────┘
                                                 │
                                         ┌───────▼───────┐
                                         │   Planner     │
                                         │  (thinking)   │
                                         └───────┬───────┘
                                                 │
                                         ┌───────▼───────┐
                                         │    Coder      │ ──── Partial Delivery ──▶ Client
                                         │  (standard)   │
                                         └───────┬───────┘
                                                 │
                                    ┌────────────┼────────────┐
                                    ▼            ▼            ▼
                              ┌──────────┐ ┌──────────┐ ┌──────────┐
                              │ Security │ │  Tester  │ │ UI Review│
                              │ (rules)  │ │ (rules)  │ │ (rules)  │
                              └──────────┘ └──────────┘ └──────────┘
                                                 │
                                         ┌───────▼───────┐     ┌───────────────┐
                                         │   Designer    │     │ Videographer  │
                                         │  (Replicate)  │     │  (Replicate)  │
                                         └───────────────┘     └───────────────┘
```
