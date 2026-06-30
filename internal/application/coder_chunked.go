package application

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Chunked Multi-File Coder (Parallel DAG)
//  Генерация файлов группами по FileMap от Архитектора.
//  Tier-based parallelism: independent groups within a tier
//  run concurrently via semaphore (rate limit protection).
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// sessionIDKey is a context key for passing session ID through generation pipeline.
type sessionIDKey struct{}

// ContextWithSessionID attaches a session ID to context for checkpoint/resume support.
func ContextWithSessionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, sessionIDKey{}, id)
}

// generateVideoKey is a context key for the "generate promo video" feature gate.
type generateVideoKey struct{}

// ContextWithGenerateVideo attaches the user's promo-video preference to context.
// true → Videographer runs sequentially BEFORE the Coder (real URL on first pass);
// false → Videographer is skipped entirely (fast prototype, token economy).
func ContextWithGenerateVideo(ctx context.Context, generate bool) context.Context {
	return context.WithValue(ctx, generateVideoKey{}, generate)
}

// generateVideoFromContext reads the promo-video preference (defaults to false).
func generateVideoFromContext(ctx context.Context) bool {
	v, _ := ctx.Value(generateVideoKey{}).(bool)

	return v
}

// fileGroup — группа файлов для одного LLM-вызова.
type fileGroup struct {
	Name  string   // "types", "lib", "services", "components", "routes", "config"
	Label string   // human-readable label for status
	Files []string // file paths from FileMap
	Tier  int      // DAG tier (0 = no deps, higher = depends on previous tiers)
}

// generationTier — один слой DAG: все группы внутри могут выполняться параллельно.
type generationTier struct {
	Level  int         // 0, 1, 2, ...
	Groups []fileGroup // groups to generate in parallel
}

// maxFilesPerGroup — hard limit on files per LLM call.
// Kept at 2 to prevent token-limit truncation on heavy domain components.
// XML artifact protocol recovers partial output, but smaller chunks = higher success rate.
const maxFilesPerGroup = 2

// maxParallelLLM — semaphore size for concurrent LLM calls within a tier.
// Protects against Anthropic rate limits (RPM) AND payload-spike timeouts: at 6
// the simultaneous reasoning requests (each with the large static system prompt +
// manifest) overwhelmed the provider's HTTP/upload window, causing 'LLM reasoning
// call timed out' that stalled late tiers. Lowered to 3 to cap concurrent load.
const maxParallelLLM = 3

// groupFileMap splits FileMap entries into ordered generation groups.
// Components are sub-classified into layout/sections/ui/domain to avoid
// monolithic groups that hit max_tokens limits.
//
//nolint:gocyclo // classification tree is intentionally explicit
func groupFileMap(fileMap []string) []fileGroup {
	groups := map[string][]string{
		"infra":      {}, // api-client, query-client — фундамент, Tier 0
		"config":     {},
		"types":      {},
		"lib":        {},
		"services":   {},
		"hooks":      {},
		"store":      {},
		"layout":     {},
		"sections":   {},
		"ui":         {},
		"components": {},
		"routes":     {},
		"server":     {},
		"styles":     {},
	}

	for _, f := range fileMap {
		fl := strings.ToLower(f)
		switch {
		case strings.Contains(fl, "vite.config") || strings.Contains(fl, "tailwind.config") ||
			strings.Contains(fl, "tsconfig") || strings.Contains(fl, "package.json"):
			groups["config"] = append(groups["config"], f)
		case strings.Contains(fl, "/types/") || strings.Contains(fl, ".d.ts"):
			groups["types"] = append(groups["types"], f)
		// Инфраструктурные файлы: api-client, query-client — выделяются в Tier 0,
		// чтобы фундамент был готов ДО любых компонентов, сервисов и хуков.
		case strings.Contains(fl, "api-client") || strings.Contains(fl, "apiclient") ||
			strings.Contains(fl, "query-client") || strings.Contains(fl, "queryclient") ||
			(strings.Contains(fl, "/lib/") && strings.Contains(fl, "client")):
			groups["infra"] = append(groups["infra"], f)
		case strings.Contains(fl, "/lib/") || strings.Contains(fl, "/utils"):
			groups["lib"] = append(groups["lib"], f)
		case strings.Contains(fl, "/services/") || strings.Contains(fl, "/api/"):
			groups["services"] = append(groups["services"], f)
		case strings.Contains(fl, "/hooks/"):
			groups["hooks"] = append(groups["hooks"], f)
		case strings.Contains(fl, "/store/") || strings.Contains(fl, "/stores/"):
			groups["store"] = append(groups["store"], f)
		case strings.Contains(fl, "/styles/") || strings.HasSuffix(fl, ".css"):
			groups["styles"] = append(groups["styles"], f)
		case strings.Contains(fl, "/routes/") || strings.Contains(fl, "/pages/"):
			groups["routes"] = append(groups["routes"], f)
		case strings.Contains(fl, "server/"):
			groups["server"] = append(groups["server"], f)
		// ── Component sub-classification ──
		case strings.Contains(fl, "/layout/") || strings.Contains(fl, "navbar") ||
			strings.Contains(fl, "footer") || strings.Contains(fl, "sidebar") ||
			strings.Contains(fl, "header") || strings.Contains(fl, "applayout"):
			groups["layout"] = append(groups["layout"], f)
		case strings.Contains(fl, "/sections/") || strings.Contains(fl, "hero") ||
			strings.Contains(fl, "about") || strings.Contains(fl, "contact") ||
			strings.Contains(fl, "testimonial") || strings.Contains(fl, "gallery"):
			groups["sections"] = append(groups["sections"], f)
		case strings.Contains(fl, "/ui/"):
			groups["ui"] = append(groups["ui"], f)
		default:
			groups["components"] = append(groups["components"], f)
		}
	}

	// Ordered generation sequence — deps first
	order := []struct {
		key   string
		label string
	}{
		{"infra", "🏗️ Инфраструктура (API-клиент, QueryClient)"},
		{"config", "⚙️ Конфигурация проекта"},
		{"types", "📝 Типы и интерфейсы"},
		{"lib", "🔧 Утилиты и хелперы"},
		{"styles", "🎨 Стили и CSS"},
		{"store", "💾 Стейт-менеджмент"},
		{"services", "🔌 Сервисы и API"},
		{"hooks", "🪝 React хуки"},
		{"ui", "🧱 UI-примитивы (shadcn)"},
		{"layout", "📐 Лейаут (Navbar, Footer, Sidebar)"},
		{"sections", "🖼️ Секции (Hero, About, Gallery)"},
		{"components", "🧩 Компоненты (domain-specific)"},
		{"routes", "🗺️ Страницы и маршруты"},
		{"server", "🖥️ Серверная часть"},
	}

	var result []fileGroup
	for _, o := range order {
		files := groups[o.key]
		if len(files) == 0 {
			continue
		}
		// Auto-split groups exceeding maxFilesPerGroup
		if len(files) > maxFilesPerGroup {
			for i := 0; i < len(files); i += maxFilesPerGroup {
				end := i + maxFilesPerGroup
				end = min(end, len(files))
				batchNum := i/maxFilesPerGroup + 1
				result = append(result, fileGroup{
					Name:  fmt.Sprintf("%s_%d", o.key, batchNum),
					Label: fmt.Sprintf("%s (часть %d)", o.label, batchNum),
					Files: files[i:end],
				})
			}
		} else {
			result = append(result, fileGroup{
				Name:  o.key,
				Label: o.label,
				Files: files,
			})
		}
	}

	return result
}

// tierMap assigns each group key to a DAG tier level.
// Groups within the same tier have no mutual dependencies and can run in parallel.
// UI-FIRST PRIORITY: the presentation layer (layout, sections, routes/pages) is
// promoted to tier 4 so it generates BEFORE the long tail. Previously routes lived
// in the last tier (5) — exactly where timeouts struck — leaving the app with a
// data layer but no rendered UI (blank screen). 'server' (backend, irrelevant to
// the preview) is now the only late tier, so a stall there never blanks the UI.
//
//	Tier 0: infra, config                           (no deps — foundation first)
//	Tier 1: types                                   (depends on config)
//	Tier 2: lib, styles, store                      (depends on types)
//	Tier 3: services, hooks, ui                     (depends on lib)
//	Tier 4: layout, sections, components, routes    (the full presentation layer)
//	Tier 5: server                                  (backend — least critical, last)
var tierMap = map[string]int{
	"infra":      0, // api-client, query-client — фундамент раньше всего
	"config":     0,
	"types":      1,
	"lib":        2,
	"styles":     2,
	"store":      2,
	"services":   3,
	"hooks":      3,
	"ui":         3,
	"layout":     4,
	"sections":   4,
	"components": 4,
	"routes":     4,
	"server":     5,
}

// tierForGroup resolves the tier level for a fileGroup by its base name
// (strips _N suffixes from auto-split groups like "components_2").
func tierForGroup(groupName string) int {
	base := groupName
	// Strip "_N" suffix from auto-split groups
	if idx := strings.LastIndex(groupName, "_"); idx > 0 {
		candidate := groupName[:idx]
		if _, ok := tierMap[candidate]; ok {
			base = candidate
		}
	}
	if tier, ok := tierMap[base]; ok {
		return tier
	}

	return 5 // unknown → last tier (safest)
}

// buildGenerationTiers organizes flat fileGroups into DAG tiers.
// Each tier completes fully before the next tier starts.
// Within a tier, groups run in parallel (bounded by maxParallelLLM).
func buildGenerationTiers(groups []fileGroup) []generationTier {
	tierBuckets := make(map[int][]fileGroup)
	maxTier := 0

	for _, g := range groups {
		t := tierForGroup(g.Name)
		g.Tier = t
		tierBuckets[t] = append(tierBuckets[t], g)
		if t > maxTier {
			maxTier = t
		}
	}

	tiers := make([]generationTier, 0, maxTier+1)
	for level := 0; level <= maxTier; level++ {
		if bucket, ok := tierBuckets[level]; ok && len(bucket) > 0 {
			tiers = append(tiers, generationTier{
				Level:  level,
				Groups: bucket,
			})
		}
	}

	return tiers
}

// ── Mirror Protocol: Director Pre-flight Check ─────────────────────────────
//
// preflightManifestCheck validates the Architect's FileMap against structural
// heuristics before the Coder starts — acting as an Architectural Mentor.
// It does NOT block generation; it emits VALIDATION thoughts so the operator
// can see gaps before they cause build failures downstream.
//
// Heuristics mirror the "Decision Logic" used by elite AI agents:
//   - Re-evaluate the plan before executing (don't just patch after the fact).
//   - Surface missing architectural layers (types/services/hooks/routes) early.
//   - Warn when the FileMap is too sparse to produce a functional application.
func (o *Orchestrator) preflightManifestCheck(ctx context.Context, manifest *SystemManifest) {
	if manifest == nil || len(manifest.FileMap) == 0 {
		return
	}

	var hasTypes, hasServices, hasHooks, hasRoutes bool
	for _, f := range manifest.FileMap {
		fl := strings.ToLower(f)
		if strings.Contains(fl, "/types/") || strings.HasSuffix(fl, ".d.ts") {
			hasTypes = true
		}
		if strings.Contains(fl, "/services/") || strings.Contains(fl, "/api/") {
			hasServices = true
		}
		if strings.Contains(fl, "/hooks/") {
			hasHooks = true
		}
		if strings.Contains(fl, "/routes/") || strings.Contains(fl, "/pages/") {
			hasRoutes = true
		}
	}

	var warnings []string
	if !hasTypes {
		warnings = append(warnings, "no types/ layer — shared interfaces will be inline-duplicated, causing type drift")
	}
	if hasServices && !hasTypes {
		warnings = append(warnings, "services exist but no shared types — API responses will be typed as 'any'")
	}
	if !hasRoutes {
		warnings = append(warnings, "no routes/pages — app will have no navigable screens")
	}
	if hasServices && !hasHooks {
		warnings = append(warnings, "services exist but no hooks — components will call services directly, bypassing React lifecycle")
	}
	if len(manifest.FileMap) < 6 {
		warnings = append(warnings, fmt.Sprintf("FileMap has only %d files — likely too sparse for a functional application", len(manifest.FileMap)))
	}

	if len(warnings) == 0 {
		o.sendThought(ctx, RoleArchitect, "VALIDATION",
			fmt.Sprintf("Pre-flight OK: %d files, all architectural layers present", len(manifest.FileMap)))
		return
	}

	o.sendThought(ctx, RoleArchitect, "VALIDATION",
		fmt.Sprintf("Pre-flight WARNINGS (%d):\n• %s",
			len(warnings), strings.Join(warnings, "\n• ")))
}

// generateCodeChunked generates project files in DAG tiers from the Architect's FileMap.
// Groups within the same tier run in parallel (semaphore-bounded).
// Tiers execute sequentially: tier N completes before tier N+1 starts.
// Files are published to EventBus as they're generated.
// Supports resume: if sessionID is set and checkpoint exists, skips completed tiers.
func (o *Orchestrator) generateCodeChunked(
	ctx context.Context,
	specification string,
	manifest *SystemManifest,
	_ *MasterPlan,
	_ *ReverseEngineeringResult,
	features []CompetitorFeature,
	imageURLs map[string]string,
	media MediaContext,
) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	if manifest == nil || len(manifest.FileMap) < 3 {
		return nil, ErrManifestFileMapTooSmall
	}

	groups := groupFileMap(manifest.FileMap)
	if len(groups) == 0 {
		return nil, ErrNoFileGroups
	}

	tiers := buildGenerationTiers(groups)
	applog(ctx).InfoContext(ctx, "parallel chunked coder start",
		"tiers", len(tiers),
		"groups", len(groups),
		"totalFiles", len(manifest.FileMap),
		"maxConcurrentLLM", maxParallelLLM,
	)

	// Mirror Protocol: validate FileMap structure before Coder starts.
	// Emits VALIDATION thoughts; does not block generation.
	o.preflightManifestCheck(ctx, manifest)

	run := o.newChunkedCoderRun(ctx, specification, manifest, features, imageURLs, media, groups, tiers)

	files, err := run.execute()
	if len(files) > 0 {
		// Гейт + досборка: гарантируем, что App.tsx подключает реальный UI, а не
		// остаётся инертной scaffold-заглушкой (иначе — белый экран).
		o.finalizeAppShell(ctx, specification, manifest, files)
		// Детерминированный guard: LLM-паттерн «лишняя ) внутри ${...}% шаблона»
		// — вызывает esbuild-ошибку «Expected } but found )». Исправляем до self-heal.
		if n := fixTemplateLiteralOverClose(files); n > 0 {
			o.sendTelemetry(ctx, RoleCoder, fmt.Sprintf(
				"[AST GUARD] fixTemplateLiteralOverClose: fixed %d file(s) — removed double-close-paren in ${...}%% template literals", n))
		}
		// Inner self-healing with circuit breaker: точечное исправление критических
		// ошибок до возврата файлов. Мутирует files на месте, не перезапускает pipeline.
		// Если circuit breaker сработал — файлы возвращаются как есть; UI уже получил
		// EventCircuitBreaker и показывает "Critical Failure".
		if circuitTripped := o.selfHealFiles(ctx, specification, manifest, files); circuitTripped {
			applog(ctx).WarnContext(ctx, "self-heal circuit breaker tripped — returning partial output as-is")
		}
		// Детерминированный guard: если Кодер выдал named export вместо default
		// — main.tsx упадёт с "No matching export". Дешевле одной строки строки.
		if ensureAppTsxDefaultExport(files) {
			o.sendTelemetry(ctx, RoleCoder,
				"[AST GUARD] ensureAppTsxDefaultExport: injected 'export default App;' into src/App.tsx")
		}
		// Инфра-guard: если LLM не добавил queryClient/apiclient в api-client.ts —
		// все импортеры в компонентах упадут с "has no exported member".
		if ensureApiClientExports(files) {
			o.sendTelemetry(ctx, RoleCoder,
				"[AST GUARD] ensureApiClientExports: injected missing queryClient/apiclient stubs into src/lib/api-client.ts")
		}
	}

	return files, err
}
