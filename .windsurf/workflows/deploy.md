---
description: Deploy frontend (Vercel) + backend (Railway) after successful code changes
---

# Deploy Workflow — Vercel + Railway

Codified rule: invoke this workflow after every successful build to ship changes
to production. Cascade must verify builds first, then ask for explicit
confirmation before pushing — production deploys are destructive and must not
auto-run silently.

## Pre-flight (mandatory, auto-runnable)

1. Verify backend builds.
// turbo
```powershell
go build ./...
```

2. Verify backend tests + vet.
// turbo
```powershell
go vet ./...
```

3. Verify frontend strict typecheck.
// turbo
```powershell
cd web; npx tsc --noEmit -p tsconfig.app.json
```

4. Verify frontend production build.
// turbo
```powershell
cd web; npx vite build
```

If any step above fails — STOP. Do not deploy. Report errors and fix root cause.

## Confirmation gate

5. Pause and ask the user:
   `Builds зелёные. Деплоить frontend (Vercel) + backend (Railway)? [yes/no]`

   Do NOT proceed without explicit `yes`. This step is non-negotiable —
   production deploys mutate live state.

## Deploy (only after confirmation)

6. Backend → Railway. Requires `railway` CLI authenticated + `railway link` already run.
```powershell
railway up
```

7. Frontend → Vercel. Requires `vercel` CLI authenticated + project linked.
```powershell
cd web; vercel --prod
```

## Smoke test (mandatory post-deploy)

8. Verify backend health.
```powershell
curl https://YOUR-BACKEND.railway.app/api/v1/health
# expect: {"status":"healthy","agent_count":10}
```

9. Verify agents endpoint.
```powershell
curl https://YOUR-BACKEND.railway.app/api/v1/agents/status
# expect: 10 agents in canonical pipeline order
```

10. Verify frontend.
```powershell
curl -s -o /dev/null -w "%{http_code}" https://YOUR-FRONTEND.vercel.app/
# expect: 200
```

If any smoke test fails — roll back immediately:
```powershell
railway rollback           # backend
cd web; vercel rollback    # frontend
```

## Notes

- Env vars (`ANTHROPIC_API_KEY`, `REPLICATE_API_TOKEN`, `VITE_API_BASE_URL`,
  `CORS_ALLOWED_ORIGINS`) must already be set in Railway/Vercel dashboards.
- Self-deploy endpoint `/api/v1/deploy/railway` is for **user-generated apps**,
  NOT for deploying Istok itself. Istok deploys exclusively via this workflow.
- The "deploy after every code change" rule applies only to commits that pass
  the pre-flight. Failing builds never deploy.
