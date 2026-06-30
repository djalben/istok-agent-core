---
name: use-wrapper
description: Оборачивает ошибки и логирование через gitlab.com/libs-artifex/wrapper. Use when handling errors, logging, or when user asks about error handling in XPLR.
---

# Использование wrapper

## Правило

**Всегда** используй `gitlab.com/libs-artifex/wrapper/v2` вместо стандартных механизмов оборачивания.

## Импорт

```go
import "gitlab.com/libs-artifex/wrapper/v2"
```

## Типичные функции (v2 API)

- `wrapper.Wrap(err)` — простая обёртка существующей ошибки.
- `wrapper.Wrapf(err, "format %s", arg)` — обёртка с форматированием (используй СТРОГО вместо `fmt.Errorf`).
- Используй другие функции библиотеки строго согласно её документации.

## Запрещено

- Использование `fmt.Errorf` (заменяй на `wrapper.Wrapf`).
- Использование `errors.New` внутри функций (РАЗРЕШЕНО ТОЛЬКО для объявления глобальных sentinel-ошибок уровня пакета, например: `var ErrNotFound = errors.New("...")`).
- Самописные обёртки ошибок.
- Прямое использование стандартного `log`/`slog` для вывода ошибок (ошибки логируются на верхнем уровне через `ErrorContext(..., "error", wrapper.Wrap(err))`, обычные логи продолжают использовать `applog(ctx).InfoContext`).
