package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Chunked Multi-File Coder
//  Генерация файлов группами по FileMap от Архитектора.
//  Types → Lib → Services → Components → Routes → Config
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// fileGroup — группа файлов для одного LLM-вызова.
type fileGroup struct {
	Name  string   // "types", "lib", "services", "components", "routes", "config"
	Label string   // human-readable label for status
	Files []string // file paths from FileMap
}

// maxFilesPerGroup — groups exceeding this are auto-split into sub-batches.
const maxFilesPerGroup = 8

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

// generateCodeChunked generates project files in groups from the Architect's FileMap.
// Each group is a separate LLM call. Files are published to EventBus as they're generated.
// Falls back to single-file generation on failure.
func (o *Orchestrator) generateCodeChunked(
	ctx context.Context,
	specification string,
	manifest *SystemManifest,
	plan *MasterPlan,
	audit *ReverseEngineeringResult,
	features []CompetitorFeature,
	imageURLs map[string]string,
) (map[string]string, error) {

	if manifest == nil || len(manifest.FileMap) < 3 {
		return nil, fmt.Errorf("manifest FileMap too small for chunked generation")
	}

	groups := groupFileMap(manifest.FileMap)
	if len(groups) == 0 {
		return nil, fmt.Errorf("no file groups after classification")
	}

	agent := o.agents[RoleCoder]
	allFiles := make(map[string]string)
	var generatedFileNames []string
	totalFiles := len(manifest.FileMap)

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

	log.Printf("📦 Chunked Coder: %d groups, %d total files from FileMap", len(groups), totalFiles)

	for gi, group := range groups {
		select {
		case <-ctx.Done():
			log.Printf("⚠️ Chunked Coder: context cancelled at group %d/%d", gi+1, len(groups))
			if len(allFiles) > 0 {
				return allFiles, nil
			}
			return nil, ctx.Err()
		default:
		}

		progressBase := 40 + (gi * 50 / len(groups))
		o.sendStatus(RoleCoder, "running",
			fmt.Sprintf("💻 Группа %d/%d: %s (%d файлов)...", gi+1, len(groups), group.Label, len(group.Files)),
			progressBase)

		// Build context of already-generated files (names only, not content — saves tokens)
		prevCtx := ""
		if len(generatedFileNames) > 0 {
			prevCtx = "\nALREADY GENERATED FILES (you can import from them):\n" + strings.Join(generatedFileNames, "\n")
		}

		fileList := strings.Join(group.Files, "\n")

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
		// Token budget scales with file count (capped at maxFilesPerGroup=8)
		maxTokens := 4096 + len(group.Files)*1024
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
			userPrompt, maxTokens, agent.ThinkingBudget)

		elapsed := time.Since(start)

		if err != nil {
			log.Printf("⚠️ Chunked Coder group %d/%d (%s) FAILED after %v: %v",
				gi+1, len(groups), group.Name, elapsed, err)
			o.sendStatus(RoleCoder, "running",
				fmt.Sprintf("⚠️ Группа %s: ошибка — пропуск", group.Label), progressBase)
			continue
		}

		// Parse the response as map[string]string
		files := o.parseCodeFiles(content)
		if len(files) == 0 {
			log.Printf("⚠️ Chunked Coder group %d/%d (%s): parseCodeFiles returned 0 files",
				gi+1, len(groups), group.Name)
			continue
		}

		// Merge and publish
		for filename, code := range files {
			allFiles[filename] = code
			generatedFileNames = append(generatedFileNames, filename)
			// Publish each file immediately via SSE
			o.events.PublishFile(RoleCoder, filename, code)
		}

		log.Printf("✅ Chunked Coder group %d/%d (%s): %d files in %v",
			gi+1, len(groups), group.Name, len(files), elapsed)
		o.sendStatus(RoleCoder, "running",
			fmt.Sprintf("✅ %s: %d файлов (%v)", group.Label, len(files), elapsed.Round(time.Second)),
			progressBase+10)
	}

	if len(allFiles) == 0 {
		return nil, fmt.Errorf("chunked generation produced 0 files across all groups")
	}

	log.Printf("📦 Chunked Coder TOTAL: %d files generated from %d groups", len(allFiles), len(groups))
	return allFiles, nil
}
