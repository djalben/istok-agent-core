package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
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

// maxFilesPerGroup — groups exceeding this are auto-split into sub-batches.
const maxFilesPerGroup = 8

// maxParallelLLM — semaphore size for concurrent LLM calls within a tier.
// Protects against Anthropic rate limits (RPM).
const maxParallelLLM = 3

// groupFileMap splits FileMap entries into ordered generation groups.
// Components are sub-classified into layout/sections/ui/domain to avoid
// monolithic groups that hit max_tokens limits.
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
				if end > len(files) {
					end = len(files)
				}
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
	plan *MasterPlan,
	audit *ReverseEngineeringResult,
	features []CompetitorFeature,
	imageURLs map[string]string,
) (map[string]string, error) {

	// Hard overall timeout for the entire chunked generation — circuit breaker.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	if manifest == nil || len(manifest.FileMap) < 3 {
		return nil, fmt.Errorf("manifest FileMap too small for chunked generation")
	}

	groups := groupFileMap(manifest.FileMap)
	if len(groups) == 0 {
		return nil, fmt.Errorf("no file groups after classification")
	}

	// ── Build DAG tiers from flat groups ──
	tiers := buildGenerationTiers(groups)
	totalGroups := len(groups)
	totalFiles := len(manifest.FileMap)

	agent := o.agents[RoleCoder]

	// ── Resume support: check checkpoint for this session ──
	sessionID, _ := ctx.Value(sessionIDKey{}).(string)
	var resumeFromTier int = -1
	var mu sync.Mutex
	allFiles := make(map[string]string)
	var generatedFileNames []string

	if sessionID != "" {
		if cp := o.sessionCache.Get(sessionID); cp != nil && len(cp.Files) > 0 {
			log.Printf("🔄 Resume: session %s has checkpoint at tier %d with %d files",
				sessionID, cp.CompletedTier, len(cp.Files))
			resumeFromTier = cp.CompletedTier
			for k, v := range cp.Files {
				allFiles[k] = v
				generatedFileNames = append(generatedFileNames, k)
			}
			// Re-publish cached files so frontend gets them
			for filename, code := range cp.Files {
				o.events.PublishFile(RoleCoder, filename, code)
			}
			o.sendStatus(RoleCoder, "running",
				fmt.Sprintf("🔄 Возобновление: %d файлов из кэша, продолжаю с tier %d...", len(cp.Files), cp.CompletedTier+1), 30)
		}
	}

	// Build manifest context (compact)
	manifestJSON, _ := json.Marshal(manifest)
	manifestCtx := string(manifestJSON)
	if len(manifestCtx) > 6000 {
		manifestCtx = manifestCtx[:6000] + "..."
	}

	// Build feature context
	featureCtx := ""
	if len(features) > 0 {
		var lines []string
		for _, f := range features {
			lines = append(lines, fmt.Sprintf("- [%s] %s: %s", f.Priority, f.Name, f.Description))
		}
		featureCtx = "\nCOMPETITOR FEATURES:\n" + strings.Join(lines, "\n")
	}

	// Image context
	imgCtx := ""
	if len(imageURLs) > 0 {
		var imgLines []string
		for key, url := range imageURLs {
			imgLines = append(imgLines, fmt.Sprintf("- %s: %s", key, url))
		}
		imgCtx = "\nGENERATED IMAGES (use real URLs, NOT placeholders):\n" + strings.Join(imgLines, "\n")
	}

	log.Printf("📦 Parallel Chunked Coder: %d tiers, %d groups, %d total files (max %d concurrent LLM)",
		len(tiers), totalGroups, totalFiles, maxParallelLLM)

	// Semaphore channel — bounds concurrent LLM calls across all goroutines
	semaphore := make(chan struct{}, maxParallelLLM)

	completedGroups := 0

	// ── Execute tiers sequentially; groups within a tier in parallel ──
	for ti, tier := range tiers {
		// Skip already-completed tiers on resume
		if resumeFromTier >= 0 && tier.Level <= resumeFromTier {
			log.Printf("⏩ Skipping tier %d (already in checkpoint)", tier.Level)
			continue
		}

		select {
		case <-ctx.Done():
			mu.Lock()
			n := len(allFiles)
			mu.Unlock()
			log.Printf("ERROR: Chunked Coder circuit breaker — context cancelled at tier %d/%d (%d files so far): %v",
				ti+1, len(tiers), n, ctx.Err())
			if n > 0 {
				return allFiles, nil
			}
			return nil, ctx.Err()
		default:
		}

		tierStart := time.Now()
		log.Printf("🔷 Tier %d/%d: %d parallel groups", ti+1, len(tiers), len(tier.Groups))
		o.sendStatus(RoleCoder, "running",
			fmt.Sprintf("� Tier %d/%d: %d групп параллельно...", ti+1, len(tiers), len(tier.Groups)),
			40+(ti*50/len(tiers)))

		// Snapshot generatedFileNames BEFORE tier starts (immutable for this tier's prompts)
		mu.Lock()
		prevSnapshot := make([]string, len(generatedFileNames))
		copy(prevSnapshot, generatedFileNames)
		mu.Unlock()

		prevCtx := ""
		if len(prevSnapshot) > 0 {
			prevCtx = "\nALREADY GENERATED FILES (you can import from them):\n" + strings.Join(prevSnapshot, "\n")
		}

		// WaitGroup for all groups in this tier
		var wg sync.WaitGroup

		for _, group := range tier.Groups {
			wg.Add(1)
			go func(g fileGroup) {
				defer wg.Done()

				// Acquire semaphore slot (blocks if maxParallelLLM reached)
				select {
				case semaphore <- struct{}{}:
				case <-ctx.Done():
					return
				}
				defer func() { <-semaphore }()

				o.sendStatus(RoleCoder, "running",
					fmt.Sprintf("💻 [T%d] %s (%d файлов)...", g.Tier, g.Label, len(g.Files)),
					40+(ti*50/len(tiers)))

				fileList := strings.Join(g.Files, "\n")

				userPrompt := fmt.Sprintf(`Generate the following files for project: %s

ARCHITECTURE MANIFEST:
%s
%s%s%s
FILES TO GENERATE IN THIS BATCH:
%s

RULES:
1. Output ONLY a JSON object: {"filepath": "file content", ...}
2. Each key is the exact file path from the list above.
3. Write PRODUCTION-READY TypeScript/React code.
4. Use @/* import aliases (e.g., @/components/ui/button, @/hooks/useAuth).
5. Use shadcn/ui components from @/components/ui/*.
6. Include real business logic — forms with validation, data fetching, state management.
7. Use addEventListener pattern, NOT inline event handlers (no onclick/onchange attributes).
8. Import types from @/types/*, services from @/services/*, hooks from @/hooks/*.
9. Every component must be properly typed with TypeScript interfaces.
10. NO Lorem Ipsum — use real content appropriate for "%s".

OUTPUT: {"filepath1": "content1", "filepath2": "content2", ...}`,
					specification, manifestCtx, featureCtx, imgCtx, prevCtx, fileList, specification)

				start := time.Now()
				maxTokens := 4096 + len(g.Files)*1024
				if maxTokens > 12288 {
					maxTokens = 12288
				}

				content, err := o.callLLMWithReasoning(ctx, agent.Model,
					`You are an elite TypeScript/React developer. Generate production-ready code files.
STACK: Vite 5, React 18, TypeScript, TanStack Router+Query, shadcn/ui, TailwindCSS, Zustand.
RULES:
- Every file must be complete and immediately usable.
- Use @/* import aliases. Never use relative paths like ../
- All event handlers via addEventListener or React synthetic events. NO inline handlers.
- Respond with valid JSON only. No markdown, no explanation.`,
					userPrompt, maxTokens)

				elapsed := time.Since(start)

				if err != nil {
					log.Printf("⚠️ Chunked Coder [T%d] %s FAILED after %v: %v",
						g.Tier, g.Name, elapsed, err)
					o.sendStatus(RoleCoder, "running",
						fmt.Sprintf("⚠️ %s: ошибка — пропуск", g.Label), 0)
					return
				}

				files := o.parseCodeFiles(content)
				if len(files) == 0 {
					log.Printf("⚠️ Chunked Coder [T%d] %s: parseCodeFiles returned 0 files", g.Tier, g.Name)
					return
				}

				// Merge results under lock
				mu.Lock()
				for filename, code := range files {
					allFiles[filename] = code
					generatedFileNames = append(generatedFileNames, filename)
				}
				completedGroups++
				mu.Unlock()

				// Publish each file via SSE (EventBus is non-blocking)
				for filename, code := range files {
					o.events.PublishFile(RoleCoder, filename, code)
				}

				log.Printf("✅ Chunked Coder [T%d] %s: %d files in %v",
					g.Tier, g.Name, len(files), elapsed)
				o.sendStatus(RoleCoder, "running",
					fmt.Sprintf("✅ %s: %d файлов (%v)", g.Label, len(files), elapsed.Round(time.Second)),
					40+(ti*50/len(tiers))+10)
			}(group)
		}

		// Wait for ALL groups in this tier before advancing to next tier
		wg.Wait()

		mu.Lock()
		tierFiles := len(allFiles)
		mu.Unlock()
		log.Printf("🔷 Tier %d/%d complete: %d total files so far (%v)",
			ti+1, len(tiers), tierFiles, time.Since(tierStart).Round(time.Millisecond))

		// ── Checkpoint: save progress after each tier ──
		if sessionID != "" {
			mu.Lock()
			snapshot := make(map[string]string, len(allFiles))
			for k, v := range allFiles {
				snapshot[k] = v
			}
			mu.Unlock()
			o.sessionCache.Save(&SessionCheckpoint{
				SessionID:     sessionID,
				Specification: specification,
				Mode:          ModeAgent,
				Files:         snapshot,
				CompletedTier: tier.Level,
				TotalTiers:    len(tiers),
				CreatedAt:     time.Now(),
			})
			log.Printf("💾 Checkpoint saved: session=%s tier=%d files=%d", sessionID, tier.Level, len(snapshot))
		}
	}

	mu.Lock()
	finalCount := len(allFiles)
	result := make(map[string]string, finalCount)
	for k, v := range allFiles {
		result[k] = v
	}
	mu.Unlock()

	if finalCount == 0 {
		return nil, fmt.Errorf("chunked generation produced 0 files across all tiers")
	}

	log.Printf("📦 Parallel Chunked Coder TOTAL: %d files from %d groups across %d tiers",
		finalCount, completedGroups, len(tiers))
	return result, nil
}
