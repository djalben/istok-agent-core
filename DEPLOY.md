# ИСТОК Agent Core — Deploy Guide

## Architecture

```
Frontend (Vercel)          Backend (Railway)
web/ → dist/               cmd/server/main.go → bin/server
React 18 + Vite 5          Go 1.24 + DualRouter LLM
shadcn/ui + TanStack       Anthropic Direct + Replicate
─────────────────          ──────────────────────
VITE_API_BASE_URL ───────→ :8080/api/v1/*
                           SSE: /api/v1/generate/stream
                           Deploy: /api/v1/deploy/railway
                           Agents: /api/v1/agents/status
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
| `ANTHROPIC_API_KEY` | **Yes** | `sk-ant-api03-...` |
| `REPLICATE_API_TOKEN` | **Yes** | `r8_...` |
| `RAILWAY_API_TOKEN` | No | `xxxxx` (enables `/api/v1/deploy/railway`) |
| `CORS_ALLOWED_ORIGINS` | No | `https://istok.vercel.app,https://custom.domain.com` |
| `JWT_SECRET` | No | `your-jwt-secret-32chars` |
| `PORT` | Auto | Set by Railway automatically |
| `RAILWAY_ENVIRONMENT` | Auto | `production` (set by Railway) |

### Agents (10 canonical pipeline)

| # | Agent | Model | Provider |
|---|---|---|---|
| 1 | Director | claude-3-7-sonnet | Anthropic Direct |
| 2 | Researcher | claude-3-7-sonnet-thinking | Anthropic Direct |
| 3 | Brain | claude-3-7-sonnet-thinking | Anthropic Direct |
| 4 | Architect | claude-3-7-sonnet-thinking | Anthropic Direct |
| 5 | Planner | claude-3-7-sonnet-thinking | Anthropic Direct |
| 6 | Coder | claude-3-7-sonnet | Anthropic Direct |
| 7 | Designer | google/nano-banana | Replicate |
| 8 | Security | claude-3-7-sonnet | Anthropic Direct |
| 9 | Tester | local + claude-3-7-sonnet | Anthropic Direct |
| 10 | UI Reviewer | claude-3-7-sonnet | Anthropic Direct |

Verification Gate aggregates Security + Tester + UI Reviewer; FSM blocks `Completed`
until all three approve.

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
# Expected: {"status":"healthy","agent_count":10,"version":"3.0.0",...}

# 2. Agents pipeline metadata
curl https://YOUR-BACKEND.railway.app/api/v1/agents/status | jq .
# Expected: { "agents": [ { "id":"director", ... }, ... 10 entries ] }

# 3. SSE stream test (10 second timeout)
curl -N -H "Content-Type: application/json" \
  -d '{"specification":"Test ping","mode":"code"}' \
  https://YOUR-BACKEND.railway.app/api/v1/generate/stream \
  --max-time 10
# Expected: event:status (with agent field), event:file, event:done

# 4. Frontend loads
curl -s -o /dev/null -w "%{http_code}" https://YOUR-FRONTEND.vercel.app/
# Expected: 200
```

## 4. FSM States

```
Created → Researching → Planning → ArchitectureApproved → StrategySynthesized
→ Designing → Coding → QualityCheck → SecurityCheck → Verified → Completed
                              ↓              ↓
                         RetryCoding ←───────┘ (auto-fix, max 2 retries)
```

`Completed` requires all three Verification Gate flags
(`security_approved`, `tester_approved`, `ui_reviewer_approved`).

## 5. Local Deploy Commands

```powershell
# Backend (Railway via CLI)
railway up                                    # from repo root, after `railway link`

# Frontend (Vercel via CLI)
cd web
vercel --prod                                 # uses vercel.json + env from dashboard
```

Or trigger the codified workflow: `/deploy` in Cascade chat — runs both with confirmation.

## 6. Production Checklist

- [ ] `ANTHROPIC_API_KEY` set in Railway
- [ ] `REPLICATE_API_TOKEN` set in Railway
- [ ] `RAILWAY_API_TOKEN` set in Railway (optional, for self-deploy endpoint)
- [ ] `VITE_API_BASE_URL` set in Vercel (points to Railway URL)
- [ ] `CORS_ALLOWED_ORIGINS` set (Vercel frontend URL)
- [ ] Backend `/api/v1/health` returns `{"status":"healthy","agent_count":10}`
- [ ] `/api/v1/agents/status` returns 10 agents
- [ ] SSE `/api/v1/generate/stream` emits `event.Agent` field
- [ ] Frontend loads, AgentPulseTimeline renders 10 rows
