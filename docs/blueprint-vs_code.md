# ИСТОК Core — Blueprint vs Реальный код (аудит)

Дата: 2026-05-29
Ветка: sandbox (после merge dev → sandbox)
Сверка: `ИСТОК.md` (идеализированный blueprint) против фактического кода.

## TL;DR

Blueprint описывает идеализированную систему (10 агентов, 11-state FSM, Event Bus
channels+Redis, gRPC, Postgres/S3). Реальный код — вторая итерация, расходится по
инфраструктуре и числу агентов/состояний. Прошлый summary держится на ~85–90%,
с уточнениями ниже.

## Подтверждено в коде

- **FSM = 13 состояний**: `Created → Researching → Planning → ArchitectureApproved
  → StrategySynthesized → Designing → Coding → QualityCheck → SecurityCheck →
  RetryCoding → Verified → Completed/Failed`
  (`internal/domain/task_state_machine.go:13-27`).
- **Жёсткий gate перед Coding**: переход в `StateCoding` блокируется без валидного
  `ApprovedPlan` (`task_state_machine.go:186-193`).
- **Конвейер последовательный + 2 human-in-the-loop паузы**: утверждение бизнес-плана
  (`WaitForApproval`, `orchestrator.go:525`) и медиа (`WaitForMediaApproval`,
  `orchestrator.go:672`).
- **Параллельный блок только Coder ‖ Videographer**: два `go func()` + `wg.Wait()`
  (`orchestrator.go:794-842`).
- **Транспорт — SSE + `globalFileStore` partial delivery**: `Append/Store/MarkComplete`
  (`generate_handler_sse.go`), вывод — статический бандл, без песочницы.
- **DualRouter**: Anthropic Direct + Replicate (`internal/infrastructure/llm/router.go:19`).

## Уточнения к прошлому summary (был неточен)

1. **Состав агентов.** В домене ровно 12 ролей (`agent_event.go:8-21`):
   researcher, brain, director, **architect**, planner, coder, designer,
   videographer, validator, security, tester, ui_reviewer.
   Brain и Architect — РАЗДЕЛЬНЫЕ роли. `Editor` в пайплайне НЕТ.

2. **Editor — surgical-фича, не агент.** `ComponentEditor`
   (`usecases/component_editor.go`) + `POST /api/v1/generate/edit` +
   `InspectorProvider` / point-and-click. Точечная правка одного файла через
   XML-артифакт, ВНЕ оркестратора и FSM. Прилетело из merge с dev.

3. **RetryCoding-петля в agent mode — мёртвый код.**
   `const maxRetries = 0` (`orchestrator.go:866`) → auto-fix
   `RetryCoding → Coding` НИКОГДА не выполняется. Verification = informational
   (комментарий: Railway режет запрос на 6 мин). Путь живёт только в FSM/теории
   и в `generateCodeMode`.

4. **Триада Security/Tester/UIReviewer — НЕ параллельна.** Один синхронный вызов
   `gate.Verify()` (`orchestrator.go:894`), дедлайн 30s. Параллелизм только
   Coder ‖ Videographer.

5. **Провайдер white-label.** anthropic-модели маппятся как `"Istok Core"`, не
   "Anthropic Direct" (`orchestrator.go:287-288`). Модели:
   `anthropic/claude-sonnet-4-6[-thinking]`, `google/nano-banana`, `google/veo-3`.

## Баги-рассинхроны

| # | Где | Проблема | Статус |
|---|-----|----------|--------|
| 1 | `agents_status_handler.go:50` | `FSMStates: 12` захардкожено, а состояний 13 | ИСПРАВЛЕНО → 13 |
| 2 | `task_state_machine.go:57` | комментарий «11 состояний» (стейл) | ИСПРАВЛЕНО → 13 |
| 3 | `CanonicalPipeline` (`orchestrator.go:260-273`) vs фронт `AGENT_PIPELINE` (`useGeneration.ts:62-75`) | оба содержат 12 записей без `editor` — СИНХРОННЫ | OK, правка НЕ нужна (editor вне пайплайна) |

## Что нового с момента прошлого summary (merge dev → sandbox)

- **Surgical Component Editing**: `ComponentEditor`, `EditComponentHandler`,
  `InspectorProvider`, `InspectorEditPanel.tsx` (point-and-click правка компонента).
- **Миграция протокола Coder с JSON на XML-артифакты** (`coder_chunked.go`,
  `agent_helpers.go` — `<file path="...">`).
- **E2E-тесты с mock LLM** (`mock_llm_test.go`, `orchestrator_test.go`).
- Инфра: `.golangci.yml` (depguard по слоям Clean Arch), чистка `_aether_result.txt`,
  миграция правил `.windsurfrules` → `.cursor/rules/`.

## Blueprint vs Реальность (сводка)

| Blueprint (`ИСТОК.md`) | Реальный код |
|---|---|
| Event Bus channels+Redis, gRPC | In-process EventBus + SSE + `globalFileStore` поллинг |
| Postgres / S3 | In-memory store, файлы на диск сервера |
| 10 агентов, 11-state FSM | 12 ролей, 13-state FSM |
| Самовалидация с авто-фиксом | Verification = informational, `maxRetries=0` |
| Реальный запуск/сборка | Статический бандл, без песочницы |