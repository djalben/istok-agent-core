# 🚀 ДЕПЛОЙ БЭКЕНДА ИСТОК НА RAILWAY

## 🎯 Проблема

Vercel (фронтенд) не может достучаться до `localhost:8080` (бэкенд), потому что:
- **localhost** - это адрес **вашего компьютера**
- Vercel работает на **серверах в облаке**
- Облачные серверы не имеют доступа к вашему локальному компьютеру

**Решение:** Задеплоить Go бэкенд в облако (Railway, Render, Fly.io)

---

## ✅ РЕШЕНИЕ 1: Railway (Рекомендуется)

### Почему Railway?
- ✅ Бесплатный tier ($5 кредитов в месяц)
- ✅ Автоматический деплой из GitHub
- ✅ Поддержка Go из коробки
- ✅ HTTPS сертификаты автоматически
- ✅ Простая настройка переменных окружения

### Шаг 1: Подготовка проекта

**Создайте `Procfile` в корне проекта:**
```
web: ./bin/server
```

**Создайте `railway.json`:**
```json
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "NIXPACKS",
    "buildCommand": "go build -o bin/server cmd/server/main.go"
  },
  "deploy": {
    "startCommand": "./bin/server",
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
```

### Шаг 2: Деплой на Railway

1. **Зарегистрируйтесь на Railway:**
   - Перейдите на https://railway.app
   - Войдите через GitHub

2. **Создайте новый проект:**
   - Нажмите "New Project"
   - Выберите "Deploy from GitHub repo"
   - Выберите `istok-agent-core`

3. **Настройте переменные окружения:**
   ```
   PORT=8080
   OPENROUTER_API_KEY=your_openrouter_key
   GO_ENV=production
   ```

4. **Railway автоматически:**
   - Обнаружит Go проект
   - Запустит `go build`
   - Задеплоит приложение
   - Выдаст публичный URL: `https://istok-agent-core-production.up.railway.app`

### Шаг 3: Обновите фронтенд

**В Vercel добавьте переменную окружения:**
```
VITE_API_BASE_URL=https://istok-agent-core-production.up.railway.app/api/v1
```

**Или создайте `.env.production` в `/web`:**
```bash
VITE_API_BASE_URL=https://istok-agent-core-production.up.railway.app/api/v1
VITE_USE_MOCKS=false
```

### Шаг 4: Редеплой фронтенда

```bash
git add .
git commit -m "🚀 Подключение к Railway backend"
git push origin main
```

Vercel автоматически задеплоит с новыми переменными.

---

## ✅ РЕШЕНИЕ 2: Render

### Шаг 1: Подготовка

**Создайте `render.yaml`:**
```yaml
services:
  - type: web
    name: istok-backend
    env: go
    buildCommand: go build -o bin/server cmd/server/main.go
    startCommand: ./bin/server
    envVars:
      - key: PORT
        value: 8080
      - key: OPENROUTER_API_KEY
        sync: false
```

### Шаг 2: Деплой

1. Зарегистрируйтесь на https://render.com
2. Подключите GitHub репозиторий
3. Render автоматически обнаружит `render.yaml`
4. Получите URL: `https://istok-backend.onrender.com`

---

## ✅ РЕШЕНИЕ 3: Fly.io

### Шаг 1: Установка CLI

```bash
# Windows (PowerShell)
iwr https://fly.io/install.ps1 -useb | iex

# macOS/Linux
curl -L https://fly.io/install.sh | sh
```

### Шаг 2: Деплой

```bash
# Логин
fly auth login

# Инициализация
fly launch --name istok-backend

# Деплой
fly deploy
```

Получите URL: `https://istok-backend.fly.dev`

---

## 🔧 РЕШЕНИЕ 4: Ngrok (Временное решение для тестов)

**Только для разработки! НЕ для production!**

### Установка

```bash
# Windows (Chocolatey)
choco install ngrok

# macOS
brew install ngrok
```

### Использование

```bash
# 1. Запустите Go backend локально
go run cmd/server/main.go

# 2. В другом терминале запустите ngrok
ngrok http 8080
```

**Вы получите публичный URL:**
```
Forwarding: https://abc123.ngrok.io -> http://localhost:8080
```

**Добавьте в Vercel:**
```
VITE_API_BASE_URL=https://abc123.ngrok.io/api/v1
```

**⚠️ Проблемы Ngrok:**
- URL меняется при каждом перезапуске
- Бесплатный tier имеет лимиты
- Требует постоянно запущенный компьютер
- Не подходит для production

---

## 📊 Сравнение решений

| Решение | Цена | Сложность | Production Ready | Рекомендация |
|---------|------|-----------|------------------|--------------|
| **Railway** | $5/мес бесплатно | ⭐ Легко | ✅ Да | ⭐⭐⭐⭐⭐ |
| **Render** | Бесплатный tier | ⭐⭐ Средне | ✅ Да | ⭐⭐⭐⭐ |
| **Fly.io** | Бесплатный tier | ⭐⭐⭐ Сложно | ✅ Да | ⭐⭐⭐ |
| **Ngrok** | Бесплатно | ⭐ Легко | ❌ Нет | ⭐ (только dev) |

---

## 🔐 Безопасность

### CORS уже настроен!

Бэкенд автоматически разрешает запросы от:
- `http://localhost:3000` (dev)
- `http://localhost:5173` (Vite dev)
- `*.vercel.app` (production)

### Переменные окружения

**НЕ коммитьте в Git:**
- `OPENROUTER_API_KEY`
- `BETTER_AUTH_SECRET`
- Другие секретные ключи

**Добавьте в Railway/Render через UI!**

---

## 🧪 Проверка деплоя

### 1. Проверьте health endpoint

```bash
curl https://your-backend-url.railway.app/api/v1/health
```

**Ожидаемый ответ:**
```json
{
  "status": "ok",
  "uptime": "2h34m12s",
  "version": "1.0.0"
}
```

### 2. Проверьте CORS

```bash
curl -H "Origin: https://your-app.vercel.app" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS \
     https://your-backend-url.railway.app/api/v1/generate
```

**Должны увидеть:**
```
Access-Control-Allow-Origin: https://your-app.vercel.app
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
```

### 3. Проверьте фронтенд

Откройте консоль браузера на `https://your-app.vercel.app`:

```
🔌 API Configuration: {
  API_BASE: "https://your-backend-url.railway.app/api/v1",
  USE_MOCKS: false,
  env: "production"
}
```

---

## 🚨 Troubleshooting

### Ошибка: "ERR_CONNECTION_REFUSED"

**Причина:** Фронтенд пытается подключиться к `localhost:8080`

**Решение:**
1. Проверьте переменную окружения в Vercel: `VITE_API_BASE_URL`
2. Убедитесь, что она начинается с `https://`
3. Редеплойте фронтенд

### Ошибка: "CORS policy"

**Причина:** Бэкенд не разрешает запросы от вашего домена

**Решение:**
1. Проверьте, что CORS middleware обновлен (см. `server.go`)
2. Убедитесь, что ваш Vercel домен заканчивается на `.vercel.app`
3. Перезапустите бэкенд

### Ошибка: "502 Bad Gateway"

**Причина:** Бэкенд не запустился или упал

**Решение:**
1. Проверьте логи в Railway/Render
2. Убедитесь, что `PORT` переменная установлена
3. Проверьте, что `OPENROUTER_API_KEY` установлен

---

## 📚 Дополнительные ресурсы

- [Railway Documentation](https://docs.railway.app)
- [Render Go Guide](https://render.com/docs/deploy-go)
- [Fly.io Go Guide](https://fly.io/docs/languages-and-frameworks/golang/)
- [Vercel Environment Variables](https://vercel.com/docs/concepts/projects/environment-variables)

---

**ИСТОК теперь полностью в облаке!** ☁️🚀
