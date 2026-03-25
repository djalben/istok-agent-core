# Исток Agent - Быстрый Старт

## Запуск Полной Системы

### 1. Backend (Go сервер)

```bash
# Установите переменную окружения с API ключом OpenRouter
export OPENROUTER_API_KEY="sk-or-your-key-here"

# Запустите сервер
cd d:\ПРОЕКТЫ\istok-agent-core
go run cmd/server/main.go

# Или используйте скомпилированный бинарник
./istok-server.exe
```

Сервер запустится на `http://localhost:8080`

**API Endpoints:**
- `POST http://localhost:8080/api/v1/generate` - Генерация проекта
- `GET http://localhost:8080/api/v1/stats` - Статистика агента
- `GET http://localhost:8080/api/v1/health` - Health check

### 2. Frontend (Next.js)

```bash
cd web

# Создайте файл .env.local с содержимым:
# NEXT_PUBLIC_API_URL=http://localhost:8080

# Запустите dev сервер
npm run dev
```

Frontend запустится на `http://localhost:3000`

## Использование

1. Откройте браузер на `http://localhost:3000`
2. В терминале агента введите описание проекта
3. (Опционально) Укажите URL сайта для анализа
4. Нажмите "Сгенерировать"
5. Результат появится в окне предпросмотра справа

## Киллер-Фича: Анализ Конкурентов

В поле "URL для анализа" введите адрес сайта конкурента:
```
https://example.com
```

Агент проанализирует сайт, извлечет технологии и паттерны, и использует эти знания для генерации вашего проекта.

## Переменные Окружения

### Backend
- `OPENROUTER_API_KEY` - API ключ OpenRouter (обязательно)
- `PORT` - Порт сервера (по умолчанию: 8080)

### Frontend
- `NEXT_PUBLIC_API_URL` - URL бэкенда (по умолчанию: http://localhost:8080)

## Архитектура

```
Backend (Go)                    Frontend (Next.js)
├── Domain Layer                ├── Components
│   ├── Agent                   │   ├── AgentTerminal
│   ├── LearningContext         │   ├── SandboxPreview
│   └── Intelligence            │   └── AgentStats
├── Application Layer           ├── API Client
│   └── UseCases                └── Hooks
├── Infrastructure                  ├── useAgentGenerate
│   ├── OpenRouter                  └── useAgentStats
│   └── WebCrawler
└── Transport (HTTP)
    └── Handlers
```

## Troubleshooting

**Backend не запускается:**
- Проверьте, что порт 8080 свободен
- Убедитесь, что OPENROUTER_API_KEY установлен

**Frontend не подключается к Backend:**
- Убедитесь, что backend запущен
- Проверьте CORS настройки
- Проверьте NEXT_PUBLIC_API_URL в .env.local

**Ошибка генерации:**
- Проверьте баланс токенов в статистике
- Убедитесь, что API ключ валиден
- Проверьте логи backend сервера
