package application

import (
	"context"
	"fmt"
	"log/slog"
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
// Protects against Anthropic rate limits (RPM). If you see 429s, lower to 5.
const maxParallelLLM = 6

// groupFileMap splits FileMap entries into ordered generation groups.
// Components are sub-classified into layout/sections/ui/domain to avoid
// monolithic groups that hit max_tokens limits.
//
//nolint:gocyclo // classification tree is intentionally explicit
func groupFileMap(fileMap []string) []fileGroup {
	groups := map[string][]string{
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
//
//	Tier 0: config                          (no deps)
//	Tier 1: types                           (depends on config)
//	Tier 2: lib, styles, store              (depends on types)
//	Tier 3: services, hooks, ui             (depends on lib)
//	Tier 4: layout, sections, components    (depends on services/hooks/ui)
//	Tier 5: routes, server                  (depends on all above)
var tierMap = map[string]int{
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
	"routes":     5,
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
) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 13*time.Minute)
	defer cancel()

	if manifest == nil || len(manifest.FileMap) < 3 {
		return nil, ErrManifestFileMapTooSmall
	}

	groups := groupFileMap(manifest.FileMap)
	if len(groups) == 0 {
		return nil, ErrNoFileGroups
	}

	tiers := buildGenerationTiers(groups)
	slog.Info(fmt.Sprintf("📦 Parallel Chunked Coder: %d tiers, %d groups, %d total files (max %d concurrent LLM)",
		len(tiers), len(groups), len(manifest.FileMap), maxParallelLLM))

	run := o.newChunkedCoderRun(ctx, specification, manifest, features, imageURLs, groups, tiers)

	return run.execute()
}
