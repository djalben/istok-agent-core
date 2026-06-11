package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"
	"sync"
	"time"

	"gitlab.com/libs-artifex/wrapper/v2"
)

var errChunkedPartialSuccess = errors.New("chunked coder partial success on cancel")

// MediaContext — контракт состояния медиа между Media-агентом (Designer/Videographer)
// и Кодером. Pending=true означает, что видео/изображения генерируются ПАРАЛЛЕЛЬНО и
// реальные URL ещё не готовы — Кодер ОБЯЗАН использовать stock-плейсхолдеры в
// семантических тегах (<img>/<video>), не дожидаясь URL и не выводя текст промптов.
type MediaContext struct {
	Videos         map[string]string // ключ → URL готового видео (если есть)
	Images         map[string]string // ключ → URL готового изображения (если есть)
	Pending        bool              // true: медиа ещё генерируется (плейсхолдеры)
	VideoRequested bool              // true: пользователь запросил промо-видео
}

// buildMediaContextPrompt рендерит MediaContext в инструкцию для user-промпта Кодера.
// Три состояния: (1) видео не запрошено, (2) видео готово (реальный URL),
// (3) видео ещё не готово / деградация — плейсхолдеры.
func buildMediaContextPrompt(m MediaContext) string {
	var b strings.Builder
	b.WriteString("\nMEDIA CONTRACT (state-aware rendering):\n")

	switch {
	case !m.VideoRequested:
		b.WriteString("- No video is requested for this project. Focus purely on code and layout. Do NOT add any <video> element or video placeholder.\n")
	case !m.Pending && len(m.Videos) > 0:
		for key, url := range m.Videos {
			fmt.Fprintf(&b, "- VIDEO %s: %s\n", key, url)
		}
		b.WriteString("- The promotional video is fully generated and ready. You MUST embed the real URL inside the <video> tag on your very first pass. No placeholders needed: <video autoPlay loop muted playsInline className=\"object-cover w-full h-full rounded-xl\"><source src=\"REAL_URL\" type=\"video/mp4\" /></video>.\n")
	default:
		b.WriteString("- STATUS Pending=true: the promo video is not ready yet (generation failed or in progress). ")
		b.WriteString("You MUST render a generic stock placeholder NOW inside a proper tag: <video autoPlay loop muted playsInline className=\"object-cover w-full h-full rounded-xl\"><source src=\"https://www.w3schools.com/html/mov_bbb.mp4\" type=\"video/mp4\" /></video>. ")
		b.WriteString("DO NOT wait for the real URL.\n")
	}

	if m.Pending {
		b.WriteString("- For any missing image, use an Unsplash/Pexels stock URL inside an <img> tag. NEVER render prompt prose, scripts, or reasoning as visible UI text.\n")
	}

	return b.String()
}

type chunkedCoderRun struct {
	o               *Orchestrator
	ctx             context.Context
	specification   string
	manifestCtx     string
	featureCtx      string
	imgCtx          string
	mediaCtx        string
	tiers           []generationTier
	totalGroups     int
	sessionID       string
	resumeFromTier  int
	semaphore       chan struct{}
	mu              sync.Mutex
	allFiles        map[string]string
	generatedNames  []string
	completedGroups int
}

func (o *Orchestrator) newChunkedCoderRun(
	ctx context.Context,
	specification string,
	manifest *SystemManifest,
	features []CompetitorFeature,
	imageURLs map[string]string,
	media MediaContext,
	groups []fileGroup,
	tiers []generationTier,
) *chunkedCoderRun {
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		manifestJSON = []byte("{}")
	}
	manifestCtx := string(manifestJSON)
	if len(manifestCtx) > 6000 {
		manifestCtx = manifestCtx[:6000] + "..."
	}

	featureCtx := ""
	if len(features) > 0 {
		lines := make([]string, 0, len(features))
		for _, f := range features {
			lines = append(lines, fmt.Sprintf("- [%s] %s: %s", f.Priority, f.Name, f.Description))
		}
		featureCtx = "\nCOMPETITOR FEATURES:\n" + strings.Join(lines, "\n")
	}

	imgCtx := ""
	if len(imageURLs) > 0 {
		imgLines := make([]string, 0, len(imageURLs))
		for key, url := range imageURLs {
			imgLines = append(imgLines, fmt.Sprintf("- %s: %s", key, url))
		}
		imgCtx = "\nGENERATED IMAGES (use real URLs, NOT placeholders):\n" + strings.Join(imgLines, "\n")
	}

	run := &chunkedCoderRun{
		o:              o,
		ctx:            ctx,
		specification:  specification,
		manifestCtx:    manifestCtx,
		featureCtx:     featureCtx,
		imgCtx:         imgCtx,
		mediaCtx:       buildMediaContextPrompt(media),
		tiers:          tiers,
		totalGroups:    len(groups),
		semaphore:      make(chan struct{}, maxParallelLLM),
		allFiles:       make(map[string]string),
		resumeFromTier: -1,
	}
	run.sessionID, _ = ctx.Value(sessionIDKey{}).(string)
	// Сидируем гарантированный каркас ДО восстановления чекпоинта и генерации.
	// Порядок важен: scaffold → checkpoint → LLM-файлы. Каждый следующий слой
	// перезаписывает предыдущий (resume-файлы и сгенерированный App.tsx имеют
	// приоритет над placeholder'ами каркаса).
	run.seedScaffold()
	if run.sessionID != "" {
		run.restoreCheckpoint()
	}

	return run
}

// seedScaffold заливает жёстко зашитый React-каркас в allFiles и публикует его в
// EventBus, чтобы фундаментальная оболочка (main.tsx, index.html, configs) была
// доступна фронту/превью с самого начала и пережила даже полный провал генерации.
func (run *chunkedCoderRun) seedScaffold() {
	scaffold := ScaffoldFiles()
	for name, content := range scaffold {
		run.allFiles[name] = content
		run.generatedNames = append(run.generatedNames, name)
		run.o.busFromCtx(run.ctx).PublishFile(RoleCoder, name, content)
	}
	applog(run.ctx).InfoContext(run.ctx, "project scaffold seeded", "files", len(scaffold))
}

func (run *chunkedCoderRun) restoreCheckpoint() {
	cp := run.o.sessionCache.Get(run.sessionID)
	if cp == nil || len(cp.Files) == 0 {
		return
	}
	applog(run.ctx).InfoContext(
		run.ctx, "chunked coder resume from checkpoint",
		"sessionId", run.sessionID,
		"tier", cp.CompletedTier,
		"files", len(cp.Files),
	)
	run.resumeFromTier = cp.CompletedTier
	for k, v := range cp.Files {
		run.allFiles[k] = v
		run.generatedNames = append(run.generatedNames, k)
	}
	for filename, code := range cp.Files {
		run.o.busFromCtx(run.ctx).PublishFile(RoleCoder, filename, code)
	}
	run.o.sendStatus(run.ctx, RoleCoder, "running",
		fmt.Sprintf("🔄 Возобновление: %d файлов из кэша, продолжаю с tier %d...", len(cp.Files), cp.CompletedTier+1), 30)
}

func (run *chunkedCoderRun) execute() (map[string]string, error) {
	applog(run.ctx).InfoContext(
		run.ctx, "chunked coder start",
		"tiers", len(run.tiers),
		"groups", run.totalGroups,
		"maxParallel", maxParallelLLM,
	)

	for ti, tier := range run.tiers {
		if run.resumeFromTier >= 0 && tier.Level <= run.resumeFromTier {
			applog(run.ctx).InfoContext(run.ctx, "chunked coder skip tier", "tier", tier.Level)

			continue
		}
		err := run.runTier(ti, tier)
		if err != nil {
			if errors.Is(err, errChunkedPartialSuccess) {
				break
			}

			return nil, err
		}
	}

	run.mu.Lock()
	finalCount := len(run.allFiles)
	result := make(map[string]string, finalCount)
	maps.Copy(result, run.allFiles)
	run.mu.Unlock()

	if finalCount == 0 {
		return nil, ErrChunkedGenerationEmpty
	}
	applog(run.ctx).InfoContext(
		run.ctx, "chunked coder complete",
		"files", finalCount,
		"groups", run.completedGroups,
		"tiers", len(run.tiers),
	)

	return result, nil
}

func (run *chunkedCoderRun) runTier(ti int, tier generationTier) error {
	select {
	case <-run.ctx.Done():
		run.mu.Lock()
		n := len(run.allFiles)
		run.mu.Unlock()
		applog(run.ctx).ErrorContext(
			run.ctx, "chunked coder context cancelled",
			"tier", ti+1,
			"tiersTotal", len(run.tiers),
			"filesSoFar", n,
			"error", wrapper.Wrap(run.ctx.Err()),
		)
		if n > 0 {
			return errChunkedPartialSuccess
		}

		return wrapper.Wrap(run.ctx.Err())
	default:
	}

	tierStart := time.Now()
	applog(run.ctx).InfoContext(
		run.ctx, "chunked coder tier start",
		"tier", ti+1,
		"tiersTotal", len(run.tiers),
		"groups", len(tier.Groups),
	)
	run.o.sendStatus(run.ctx, RoleCoder, "running",
		fmt.Sprintf(" Tier %d/%d: %d групп параллельно...", ti+1, len(run.tiers), len(tier.Groups)),
		40+(ti*50/len(run.tiers)))

	run.mu.Lock()
	prevSnapshot := append([]string(nil), run.generatedNames...)
	run.mu.Unlock()

	prevCtx := ""
	if len(prevSnapshot) > 0 {
		prevCtx = "\nALREADY GENERATED FILES (you can import from them):\n" + strings.Join(prevSnapshot, "\n")
	}

	var wg sync.WaitGroup
	for _, group := range tier.Groups {
		wg.Add(1)
		go func(g fileGroup) {
			defer wg.Done()
			run.processGroup(g, ti, prevCtx)
		}(group)
	}
	wg.Wait()

	run.mu.Lock()
	tierFiles := len(run.allFiles)
	run.mu.Unlock()
	applog(run.ctx).InfoContext(
		run.ctx, "chunked coder tier complete",
		"tier", ti+1,
		"tiersTotal", len(run.tiers),
		"filesTotal", tierFiles,
		"duration", time.Since(tierStart).Round(time.Millisecond),
	)

	if run.sessionID != "" {
		run.saveCheckpoint(tier)
	}

	return nil
}

func (run *chunkedCoderRun) saveCheckpoint(tier generationTier) {
	run.mu.Lock()
	snapshot := make(map[string]string, len(run.allFiles))
	maps.Copy(snapshot, run.allFiles)
	run.mu.Unlock()
	run.o.sessionCache.Save(&SessionCheckpoint{
		SessionID:     run.sessionID,
		Specification: run.specification,
		Mode:          ModeAgent,
		Files:         snapshot,
		CompletedTier: tier.Level,
		TotalTiers:    len(run.tiers),
		CreatedAt:     time.Now(),
	})
	applog(run.ctx).InfoContext(
		run.ctx, "chunked coder checkpoint saved",
		"sessionId", run.sessionID,
		"tier", tier.Level,
		"files", len(snapshot),
	)
}

func (run *chunkedCoderRun) processGroup(g fileGroup, ti int, prevCtx string) {
	select {
	case run.semaphore <- struct{}{}:
	case <-run.ctx.Done():
		return
	}
	defer func() { <-run.semaphore }()

	run.o.sendStatus(run.ctx, RoleCoder, "running",
		fmt.Sprintf("💻 [T%d] %s (%d файлов)...", g.Tier, g.Label, len(g.Files)),
		40+(ti*50/len(run.tiers)))

	userPrompt := buildChunkedCoderUserPrompt(run.specification, run.manifestCtx, run.featureCtx, run.imgCtx, run.mediaCtx, prevCtx, g.Files)
	systemPrompt := chunkedCoderSystemPrompt
	maxTokens := 4096 + len(g.Files)*3072
	maxTokens = min(maxTokens, 16384)

	start := time.Now()
	agent := run.o.agents[RoleCoder]
	var content string
	var err error
	for attempt := range 2 {
		content, err = run.o.callLLMWithReasoning(run.ctx, agent.Model, systemPrompt, userPrompt, maxTokens)
		if err == nil {
			break
		}
		if run.ctx.Err() != nil {
			break
		}
		if attempt == 0 {
			applog(run.ctx).WarnContext(
				run.ctx, "chunked coder group retry",
				"tier", g.Tier,
				"group", g.Name,
				"error", wrapper.Wrap(err),
			)
			time.Sleep(3 * time.Second)
		}
	}
	elapsed := time.Since(start)

	if err != nil {
		applog(run.ctx).WarnContext(
			run.ctx, "chunked coder group failed",
			"tier", g.Tier,
			"group", g.Name,
			"duration", elapsed,
			"error", wrapper.Wrap(err),
		)
		run.o.sendStatus(run.ctx, RoleCoder, "running", fmt.Sprintf("⚠️ %s: ошибка — пропуск", g.Label), 0)

		return
	}

	files := run.o.parseCodeFiles(run.ctx, content)
	if len(files) == 0 {
		applog(run.ctx).WarnContext(
			run.ctx, "chunked coder parse returned no files",
			"tier", g.Tier,
			"group", g.Name,
			"requested", len(g.Files),
		)

		return
	}
	if len(files) < len(g.Files) {
		applog(run.ctx).WarnContext(
			run.ctx, "chunked coder partial files",
			"tier", g.Tier,
			"group", g.Name,
			"got", len(files),
			"requested", len(g.Files),
		)
	}

	run.mu.Lock()
	for filename, code := range files {
		run.allFiles[filename] = code
		run.generatedNames = append(run.generatedNames, filename)
	}
	run.completedGroups++
	run.mu.Unlock()

	for filename, code := range files {
		run.o.busFromCtx(run.ctx).PublishFile(RoleCoder, filename, code)
	}
	applog(run.ctx).InfoContext(
		run.ctx, "chunked coder group done",
		"tier", g.Tier,
		"group", g.Name,
		"files", len(files),
		"duration", elapsed,
	)
	run.o.sendStatus(run.ctx, RoleCoder, "running",
		fmt.Sprintf("✅ %s: %d файлов (%v)", g.Label, len(files), elapsed.Round(time.Second)),
		40+(ti*50/len(run.tiers))+10)
}

// criticalMediaContract — строгий контракт рендеринга медиа (запрет утечки сценариев/
// скриптов/промптов в UI). Раньше жил внутри ultimatePremiumUIRule; вынесен отдельно при
// переходе на Titan-директивы. Бэктики оригинала → одинарные кавычки (Go raw-string).
const criticalMediaContract = `CRITICAL MEDIA CONTRACT (NO HALLUCINATIONS):
- If the architecture or state provides media (video URLs, image URLs, or pending media requests), NEVER render the internal reasoning, prompt text, or script to the UI.
- ALWAYS render videos using a proper semantic tag: <video autoPlay loop muted playsInline className="object-cover w-full h-full rounded-xl"><source src={url} type="video/mp4" /></video>.
- If the exact URL is missing, use a visually appealing generic stock URL (e.g., Unsplash/Pexels placeholder), but NEVER print raw text/scripts on the screen.

`

const chunkedCoderSystemPrompt = PremiumDesignSystem + ComponentKnowledgeBase + TitanSystemDirectives + criticalMediaContract + `ENGINEERING RULES:
You are an elite TypeScript/React developer. Generate production-ready code files.
STACK: Vite 5, React 18, TypeScript, TanStack Router+Query, shadcn/ui, TailwindCSS, Zustand.
RULES:
- Every file must be complete and immediately usable.
- Use @/* import aliases. Never use relative paths like ../
- All event handlers via addEventListener or React synthetic events. NO inline handlers.
- Add data-component-name="ComponentName" to root element of every component for visual inspector.
- DEFENSIVE LOOKUPS (NO RUNTIME CRASHES): whenever you index a config/lookup map with a dynamic key (e.g. severity/status/variant/type/role), you MUST provide a fallback so the value is never undefined. NEVER write 'const config = severityConfig[severity];'. ALWAYS write 'const config = severityConfig[severity] ?? severityConfig.info ?? Object.values(severityConfig)[0];' (use a sensible default key, then the first entry). This applies to ALL such maps (statusConfig, variantStyles, typeMap, etc.). Then read properties off the resolved value, never off the raw map access. This prevents "Cannot read properties of undefined" errors.
- If generating App.tsx or main entry, wrap content in <InspectorProvider> from @/components/InspectorProvider.
- CRITICAL: Output each file wrapped in <file path="exact/path">...</file> XML tags.
- Write raw code inside tags. NO JSON. NO escaping. NO markdown fences.`

func buildChunkedCoderUserPrompt(specification, manifestCtx, featureCtx, imgCtx, mediaCtx, prevCtx string, files []string) string {
	fileList := strings.Join(files, "\n")

	return fmt.Sprintf(`Generate the following files for project: %s

ARCHITECTURE MANIFEST:
%s
%s%s%s%s
FILES TO GENERATE IN THIS BATCH:
%s

RULES:
1. Write PRODUCTION-READY TypeScript/React code.
2. Use @/* import aliases (e.g., @/components/ui/button, @/hooks/useAuth).
3. Use shadcn/ui components from @/components/ui/*.
4. Include real business logic — forms with validation, data fetching, state management.
5. Use addEventListener pattern, NOT inline event handlers (no onclick/onchange attributes).
6. Import types from @/types/*, services from @/services/*, hooks from @/hooks/*.
7. Every component must be properly typed with TypeScript interfaces.
8. NO Lorem Ipsum — use real content appropriate for "%s".
9. Add data-component-name="ComponentName" attribute to the root element of every React component.
10. If generating App.tsx, wrap the entire app content in <InspectorProvider> from @/components/InspectorProvider.

CRITICAL OUTPUT FORMAT — XML artifact protocol:
Wrap each file in <file path="..."> tags. Write raw unescaped code inside. Example:
<file path="src/components/Button.tsx">
import React from 'react';
export const Button = () => <button>Click</button>;
</file>

Output ONLY <file> blocks. No JSON. No markdown fences. No explanation outside <file> tags.`,
		specification, manifestCtx, featureCtx, imgCtx, mediaCtx, prevCtx, fileList, specification)
}
