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

### 2. Frontend (Vite/React)

```bash
cd web

# Установите зависимости (если еще не установлены)
npm install --legacy-peer-deps

# Запустите dev сервер
npm run dev
```

Frontend запустится на `http://localhost:5173`

**⚠️ ВАЖНО: Режимы работы**

Фронтенд может работать в двух режимах:

1. **MOCKS (по умолчанию)** - работает без бэкенда, использует заглушки
   - Текущий режим: `USE_MOCKS = true` в `web/src/lib/api.ts`
   - Не требует запущенного Go сервера
   - Идеально для разработки UI

2. **REAL BACKEND** - подключается к реальному API
   - Измените `USE_MOCKS = false` в `web/src/lib/api.ts`
   - Требует запущенный Go сервер на localhost:8080
   - Или установите `VITE_API_BASE_URL` для production backend

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

**Создайте `.env.local` в папке `/web` (не коммитится в Git):**

```bash
# URL бэкенда (для production или локального бэкенда)
VITE_API_BASE_URL=http://localhost:8080/api/v1

# Использовать моки (true/false)
VITE_USE_MOCKS=false
```

**Для Vercel (production):**
```bash
VITE_API_BASE_URL=https://your-backend.railway.app/api/v1
VITE_USE_MOCKS=false
```

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

**ERR_CONNECTION_REFUSED на localhost:8080:**
- ✅ **Решение 1:** Используйте моки (по умолчанию включены)
- ✅ **Решение 2:** Запустите Go backend: `go run cmd/server/main.go`
- ✅ **Решение 3:** Задеплойте backend на Railway (см. `ДЕПЛОЙ_BACKEND.md`)

**Backend не запускается:**
- Проверьте, что порт 8080 свободен: `netstat -ano | findstr :8080`
- Убедитесь, что OPENROUTER_API_KEY установлен

**Frontend показывает моки вместо реальных данных:**
- Измените `USE_MOCKS = false` в `web/src/lib/api.ts`
- Или создайте `.env.local` с `VITE_USE_MOCKS=false`
- Убедитесь, что backend запущен

**CORS ошибки:**
- Backend уже настроен для Vercel и localhost
- Проверьте, что origin разрешен в `internal/transport/http/server.go`

**Ошибка генерации:**
- Проверьте баланс токенов в статистике
- Убедитесь, что API ключ валиден
- Проверьте логи backend сервера

## Быстрый старт для Vercel

1. **Фронтенд уже задеплоен на Vercel** ✅
2. **Локально работает с моками** ✅
3. **Для production:**
   - Задеплойте backend на Railway (см. `ДЕПЛОЙ_BACKEND.md`)
   - Добавьте `VITE_API_BASE_URL` в Vercel
   - Измените `USE_MOCKS = false` в коде
