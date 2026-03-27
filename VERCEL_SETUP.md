# 🚀 НАСТРОЙКА VERCEL ДЛЯ ИСТОК

## 🔧 Environment Variables

Для корректной работы приложения на Vercel необходимо настроить следующие переменные окружения:

### Обязательные переменные

#### 1. Go Backend API

```bash
VITE_API_BASE_URL=https://your-backend.railway.app/api/v1
```

**Где взять:**
- Задеплойте Go backend на Railway (см. `ДЕПЛОЙ_BACKEND.md`)
- Скопируйте публичный URL вашего backend
- Добавьте `/api/v1` в конец

### Опциональные переменные (Supabase)

#### 2. Supabase URL

```bash
VITE_SUPABASE_URL=https://your-project.supabase.co
```

#### 3. Supabase Anon Key

```bash
VITE_SUPABASE_PUBLISHABLE_KEY=your-anon-key
```

**Где взять:**
1. Зарегистрируйтесь на https://supabase.com
2. Создайте новый проект
3. Перейдите в Settings → API
4. Скопируйте:
   - Project URL → `VITE_SUPABASE_URL`
   - Project API keys → anon/public → `VITE_SUPABASE_PUBLISHABLE_KEY`

**Примечание:** Если Supabase не настроен, приложение будет работать с ограниченным функционалом:
- ❌ Нет аутентификации пользователей
- ❌ Нет сохранения проектов в облаке
- ✅ Генерация кода через Go backend работает
- ✅ Локальное сохранение проектов работает

---

## 📝 Как добавить переменные в Vercel

### Через Dashboard:

1. Откройте ваш проект на https://vercel.com
2. Перейдите в **Settings** → **Environment Variables**
3. Добавьте каждую переменную:
   - **Key:** `VITE_API_BASE_URL`
   - **Value:** `https://your-backend.railway.app/api/v1`
   - **Environment:** Production, Preview, Development (выберите все)
4. Нажмите **Save**
5. Повторите для остальных переменных

### Через CLI:

```bash
# Установите Vercel CLI
npm i -g vercel

# Логин
vercel login

# Добавьте переменные
vercel env add VITE_API_BASE_URL production
# Введите значение: https://your-backend.railway.app/api/v1

vercel env add VITE_SUPABASE_URL production
# Введите значение: https://your-project.supabase.co

vercel env add VITE_SUPABASE_PUBLISHABLE_KEY production
# Введите значение: your-anon-key
```

---

## 🔄 Redeploy после настройки

После добавления переменных окружения необходимо передеплоить приложение:

### Автоматический redeploy:

```bash
git commit --allow-empty -m "Trigger Vercel redeploy"
git push origin main
```

### Через Dashboard:

1. Откройте проект на Vercel
2. Перейдите в **Deployments**
3. Найдите последний деплой
4. Нажмите **...** → **Redeploy**

---

## ✅ Проверка

После redeploy откройте ваше приложение и проверьте консоль браузера (F12):

**Должно быть:**
```
🔌 API Configuration: {
  API_BASE: "https://your-backend.railway.app/api/v1",
  mode: "production"
}
```

**Не должно быть:**
```
❌ supabaseUrl is required
❌ Failed to load manifest.json (401)
```

---

## 🐛 Troubleshooting

### Ошибка: "supabaseUrl is required"

**Причина:** Не установлены Supabase переменные окружения

**Решение 1 (рекомендуется):** Настройте Supabase
- Создайте проект на https://supabase.com
- Добавьте переменные в Vercel
- Redeploy

**Решение 2:** Используйте без Supabase
- Приложение уже имеет fallback значения
- Функционал аутентификации будет недоступен
- Генерация кода через Go backend будет работать

### Ошибка: "Failed to load manifest.json (401)"

**Причина:** Vercel пытается загрузить manifest.json, но получает 401

**Решение:** Проверьте, что файл `public/manifest.json` существует и доступен

### Ошибка: "ERR_CONNECTION_REFUSED" к API

**Причина:** Backend не доступен или неправильный URL

**Решение:**
1. Проверьте, что backend задеплоен на Railway
2. Проверьте `VITE_API_BASE_URL` в Vercel
3. Убедитесь, что URL правильный (без trailing slash)

---

## 📋 Чеклист настройки

- [ ] Go backend задеплоен на Railway
- [ ] Получен публичный URL backend
- [ ] `VITE_API_BASE_URL` добавлен в Vercel
- [ ] (Опционально) Supabase проект создан
- [ ] (Опционально) `VITE_SUPABASE_URL` добавлен в Vercel
- [ ] (Опционально) `VITE_SUPABASE_PUBLISHABLE_KEY` добавлен в Vercel
- [ ] Выполнен redeploy на Vercel
- [ ] Приложение открывается без ошибок
- [ ] Консоль браузера не показывает ошибок

---

**Готово! Ваше приложение настроено и работает на Vercel!** 🎉
