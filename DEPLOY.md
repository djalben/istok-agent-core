# ИСТОК Agent Core v2.0 — Deploy Guide

## Architecture

```
Frontend (Vercel)          Backend (Railway)
web/ → dist/               cmd/server/main.go → bin/server
React 18 + Vite 5          Go 1.24 + DualRouter LLM
shadcn/ui + TanStack       Replicate + OpenRouter
─────────────────          ──────────────────────
VITE_API_BASE_URL ───────→ :8080/api/v1/*
                           SSE: /api/v1/generate/stream
```

## 1. Backend Deploy (Railway)

### Build
- **Builder**: Nixpacks (or Dockerfile)
- **Build command**: `CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server cmd/server/main.go`
- **Start command**: `./bin/server`
- **Config files**: `railway.json`, `nixpacks.toml`, `Dockerfile`

### Required Env Vars (Railway Dashboard → Variables)

| Variable | Required | Example |
|---|---|---|
| `OPENROUTER_API_KEY` | **Yes** | `sk-or-v1-...` |
| `REPLICATE_API_TOKEN` | **Yes** | `r8_...` |
| `OPENROUTER_PROXY_URL` | No | `https://your-cf-worker.workers.dev` |
| `CORS_ALLOWED_ORIGINS` | No | `https://istok.vercel.app,https://custom.domain.com` |
| `JWT_SECRET` | No | `your-jwt-secret-32chars` |
| `PORT` | Auto | Set by Railway automatically |
| `RAILWAY_ENVIRONMENT` | Auto | `production` (set by Railway) |

### Agents (7 total)

| # | Agent | Model | Provider |
|---|---|---|---|
| 1 | Researcher | deepseek/deepseek-v3.2-speciale | OpenRouter |
| 2 | Brain (Architect) | google/gemini-3-pro | Replicate |
| 3 | Director (Planner) | google/gemini-3-pro | Replicate |
| 4 | Coder | google/gemini-3-pro + qwen fallback | Replicate+OpenRouter |
| 5 | Designer | FLUX 1.1 Pro | Replicate |
| 6 | Videographer | google/gemini-3.1-pro | Replicate |
| 7 | Validator | Verification Layer v3 | Local |

## 2. Frontend Deploy (Vercel)

### Build
- **Framework**: Vite (detected automatically)
- **Build command**: `npm run build`
- **Output directory**: `dist`
- **Install command**: `npm install`
- **Root directory**: `web/`
- **Config**: `vercel.json`

### Required Env Vars (Vercel Dashboard → Environment Variables)

| Variable | Required | Example |
|---|---|---|
| `VITE_API_BASE_URL` | **Yes** | `https://your-backend.railway.app/api/v1` |
| `VITE_SUPABASE_URL` | No | `https://your-project.supabase.co` |
| `VITE_SUPABASE_PUBLISHABLE_KEY` | No | `eyJ...` |

## 3. Smoke Tests

After deploy, verify:

```bash
# 1. Health check
curl https://YOUR-BACKEND.railway.app/api/v1/health | jq .

# Expected: {"status":"healthy","agent_count":7,"version":"2.0.0",...}

# 2. SSE stream test (5 second timeout)
curl -N -H "Content-Type: application/json" \
  -d '{"specification":"Test ping","mode":"code"}' \
  https://YOUR-BACKEND.railway.app/api/v1/generate/stream \
  --max-time 10

# Expected: event:status, event:file, event:done SSE events

# 3. Frontend loads
curl -s -o /dev/null -w "%{http_code}" https://YOUR-FRONTEND.vercel.app/
# Expected: 200

# 4. No 404 on assets
curl -s -o /dev/null -w "%{http_code}" https://YOUR-FRONTEND.vercel.app/assets/index-*.js
# Expected: 200
```

## 4. FSM States (12)

```
Created → Researching → Planning → ArchitectureApproved → StrategySynthesized
→ Designing → Coding → QualityCheck → SecurityCheck → Verified → Completed
                              ↓              ↓
                         RetryCoding ←───────┘ (auto-fix, max 2 retries)
```

## 5. Production Checklist

- [ ] `OPENROUTER_API_KEY` set in Railway
- [ ] `REPLICATE_API_TOKEN` set in Railway
- [ ] `VITE_API_BASE_URL` set in Vercel (points to Railway URL)
- [ ] `CORS_ALLOWED_ORIGINS` set (Vercel frontend URL)
- [ ] Backend `/api/v1/health` returns `{"status":"healthy"}`
- [ ] SSE stream `/api/v1/generate/stream` connects
- [ ] Frontend loads without 404 errors
- [ ] Agent count = 7 in health response
