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

const (
	// perGroupLLMTimeout ограничивает ОДИН вызов LLM на группу. Без него зависший
	// reasoning-запрос держит слот semaphore на весь 13-мин бюджет тира, истощая
	// пул параллелизма и вешая весь тир. Таймаут освобождает слот и даёт ретрай.
	perGroupLLMTimeout = 4 * time.Minute

	// tierStallTimeout — backstop: если за это время в тире не появилось НИ ОДНОГО
	// нового файла (все группы возвращают ошибки/пустоту), watchdog отменяет тир и
	// генерация финализируется с тем, что уже есть, вместо ожидания полного бюджета.
	tierStallTimeout = 4 * time.Minute
)

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

		// После Tier 0 (infra + config): валидируем экспорты api-client.ts и
		// инжектируем их детерминированно ДО старта Tier 1+.
		// Это гарантирует, что все компоненты/хуки/сервисы, которые импортируют
		// queryClient/apiclient, получат корректный файл при построении prevCtx.
		if tier.Level == 0 {
			run.mu.Lock()
			injected := ensureApiClientExports(run.allFiles)
			run.mu.Unlock()
			if injected {
				applog(run.ctx).InfoContext(run.ctx,
					"infra gate: api-client exports injected deterministically")
				run.o.sendThought(run.ctx, RoleCoder, "VALIDATION",
					"api-client.ts: queryClient/apiclient не найдены — детерминированная инъекция выполнена")
			}
		}

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

	// Context minification: к поздним тирам generatedNames раздувается (96+ имён
	// с дублями) и умножается на параллелизм групп — payload-спайки вешают
	// reasoning-вызовы по таймауту. Сжимаем список до ядра + свежих файлов.
	prevCtx := ""
	if minified := minifyFileContext(prevSnapshot); len(minified) > 0 {
		prevCtx = "\nALREADY GENERATED FILES (you can import from them):\n" + strings.Join(minified, "\n")
	}

	// Контекст тира: watchdog отменяет его при зависании (нет новых файлов
	// дольше tierStallTimeout), что прерывает in-flight LLM-вызовы и даёт
	// генерации финализироваться, не дожидаясь полного 13-мин бюджета.
	tierCtx, cancelTier := context.WithCancel(run.ctx)
	defer cancelTier()
	stopWatch := make(chan struct{})
	go run.watchTierStall(tierCtx, cancelTier, ti, stopWatch)

	var wg sync.WaitGroup
	for _, group := range tier.Groups {
		wg.Add(1)
		go func(g fileGroup) {
			defer wg.Done()
			run.processGroup(tierCtx, g, ti, prevCtx)
		}(group)
	}
	wg.Wait()
	close(stopWatch)

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

// watchTierStall — backstop-watchdog тира. Если за tierStallTimeout в allFiles не
// прибавилось НИ ОДНОГО файла (все группы возвращают ошибки/пустоту или зависли),
// отменяет тир через cancel(), прерывая in-flight вызовы. Любой прогресс сбрасывает
// таймер, поэтому здоровая (пусть и медленная) генерация не прерывается.
func (run *chunkedCoderRun) watchTierStall(ctx context.Context, cancel context.CancelFunc, ti int, stop <-chan struct{}) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	run.mu.Lock()
	lastCount := len(run.allFiles)
	run.mu.Unlock()
	lastProgress := time.Now()

	for {
		select {
		case <-stop:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			run.mu.Lock()
			n := len(run.allFiles)
			run.mu.Unlock()
			if n > lastCount {
				lastCount = n
				lastProgress = time.Now()

				continue
			}
			if time.Since(lastProgress) >= tierStallTimeout {
				applog(run.ctx).WarnContext(run.ctx, "chunked coder tier stalled — finalizing early",
					"tier", ti+1, "files", n, "stall", tierStallTimeout)
				run.o.sendStatus(run.ctx, RoleCoder, "running",
					fmt.Sprintf("⏱️ Tier %d завис без прогресса — финализация (%d файлов)", ti+1, n), 0)
				cancel()

				return
			}
		}
	}
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

// processGroup генерирует одну группу файлов. ctx — контекст ТИРА (производный от
// run.ctx), который watchdog может отменить при зависании, освобождая слот semaphore.
func (run *chunkedCoderRun) processGroup(ctx context.Context, g fileGroup, ti int, prevCtx string) {
	select {
	case run.semaphore <- struct{}{}:
	case <-ctx.Done():
		return
	}
	defer func() { <-run.semaphore }()

	run.o.sendStatus(ctx, RoleCoder, "running",
		fmt.Sprintf("💻 [T%d] %s (%d файлов)...", g.Tier, g.Label, len(g.Files)),
		40+(ti*50/len(run.tiers)))

	if len(g.Files) > 0 {
		fileList := strings.Join(g.Files, ", ")
		if len(fileList) > 200 {
			fileList = fileList[:200] + "..."
		}
		run.o.sendThought(ctx, RoleCoder, "EXECUTION",
			fmt.Sprintf("[T%d] %s → %s", g.Tier, g.Label, fileList))
		run.o.sendTelemetry(ctx, RoleCoder, fmt.Sprintf(
			"[CODER] group=%s | tier=%d | files=%d | requested=[%s]",
			g.Name, g.Tier, len(g.Files), fileList))
	}

	userPrompt := buildChunkedCoderUserPrompt(run.specification, run.manifestCtx, run.featureCtx, run.imgCtx, run.mediaCtx, prevCtx, g.Files)
	systemPrompt := chunkedCoderSystemPrompt
	// Минимум 8192: гарантирует, что Haiku (лимит 8k) может использовать весь
	// выход даже для однофайловых групп (формула даёт 7168, что меньше 8192).
	maxTokens := max(8192, 4096+len(g.Files)*3072)
	maxTokens = min(maxTokens, 16384)

	start := time.Now()
	agent := run.o.agents[RoleCoder]
	var content string
	var err error
	for attempt := range 2 {
		// Per-attempt deadline: зависший reasoning-вызов не держит слот semaphore
		// дольше perGroupLLMTimeout, иначе один stuck-запрос вешает весь пул.
		attemptCtx, cancelAttempt := context.WithTimeout(ctx, perGroupLLMTimeout)
		content, err = run.o.callLLMWithReasoning(attemptCtx, agent.Model, systemPrompt, userPrompt, maxTokens)
		cancelAttempt()
		if err == nil {
			break
		}
		if ctx.Err() != nil {
			break
		}
		if attempt == 0 {
			applog(ctx).WarnContext(
				ctx, "chunked coder group retry",
				"tier", g.Tier,
				"group", g.Name,
				"error", wrapper.Wrap(err),
			)
			time.Sleep(3 * time.Second)
		}
	}
	elapsed := time.Since(start)

	if err != nil {
		applog(ctx).WarnContext(
			ctx, "chunked coder group failed",
			"tier", g.Tier,
			"group", g.Name,
			"duration", elapsed,
			"error", wrapper.Wrap(err),
		)
		run.o.sendStatus(ctx, RoleCoder, "running", fmt.Sprintf("⚠️ %s: ошибка — пропуск", g.Label), 0)

		return
	}

	files := run.o.parseCodeFiles(ctx, content)
	if len(files) == 0 {
		applog(ctx).WarnContext(
			ctx, "chunked coder parse returned no files",
			"tier", g.Tier,
			"group", g.Name,
			"requested", len(g.Files),
		)

		return
	}
	if len(files) < len(g.Files) {
		applog(ctx).WarnContext(
			ctx, "chunked coder partial files",
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
	completed := run.completedGroups
	total := run.totalGroups
	run.mu.Unlock()

	bus := run.o.busFromCtx(ctx)
	for filename, code := range files {
		bus.PublishFile(RoleCoder, filename, code)
	}

	// Прогресс задач: сколько групп выполнено из всех.
	bus.PublishTaskProgress(RoleCoder, completed, total)

	// Action log: что сгенерировано в группе.
	fileNames := make([]string, 0, len(files))
	for name := range files {
		fileNames = append(fileNames, name)
	}
	bus.PublishActionLog(RoleCoder, "code_gen",
		fmt.Sprintf("Generated %d file(s) in group %q (tier %d)", len(files), g.Name, g.Tier),
		fileNames)

	// Code diff: публикуем для первого значимого .tsx/.ts файла группы.
	for _, name := range fileNames {
		if strings.HasSuffix(name, ".tsx") || strings.HasSuffix(name, ".ts") {
			hunk, adds := buildDiffHunk(files[name])
			bus.PublishCodeDiff(RoleCoder, name, hunk, adds, 0)
			break
		}
	}
	applog(ctx).InfoContext(
		ctx, "chunked coder group done",
		"tier", g.Tier,
		"group", g.Name,
		"files", len(files),
		"duration", elapsed,
	)
	totalBytes := 0
	for _, c := range files {
		totalBytes += len(c)
	}
	run.o.sendTelemetry(ctx, RoleCoder, fmt.Sprintf(
		"[CODER DONE] group=%s | tier=%d | files_written=%d | bytes=%d | latency=%s",
		g.Name, g.Tier, len(files), totalBytes, elapsed.Truncate(time.Millisecond)))
	run.o.sendStatus(ctx, RoleCoder, "running",
		fmt.Sprintf("✅ %s: %d файлов (%v)", g.Label, len(files), elapsed.Round(time.Second)),
		40+(ti*50/len(run.tiers))+10)
}

// buildDiffHunk генерирует unified-дифф для нового файла (все строки — добавления).
// Показывает максимум maxDiffLines строк; остальное обозначает как "@@ +N more @@".
func buildDiffHunk(content string) (hunk string, additions int) {
	const maxDiffLines = 25
	lines := strings.Split(content, "\n")
	additions = len(lines)
	shown := lines
	if len(lines) > maxDiffLines {
		shown = lines[:maxDiffLines]
	}
	var b strings.Builder
	fmt.Fprintf(&b, "@@ -0,0 +1,%d @@\n", additions)
	for _, l := range shown {
		b.WriteString("+" + l + "\n")
	}
	if len(lines) > maxDiffLines {
		fmt.Fprintf(&b, "@@ +%d more lines @@\n", len(lines)-maxDiffLines)
	}
	return b.String(), additions
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

// maxContextFiles — потолок числа имён файлов в prevCtx ("уже сгенерировано").
// Сверх него список агрессивно прунится, чтобы payload к Anthropic не раздувался
// и параллельные reasoning-вызовы не падали по таймауту (self-DDOS).
const maxContextFiles = 45

// isCoreContextFile — файлы, которые ВСЕГДА остаются в контексте: каркас (entry/
// configs) + слои, от которых импортируют почти все компоненты (types/stores/
// services/hooks). Их потеря из контекста ломает связность импортов.
func isCoreContextFile(p string) bool {
	switch p {
	case "src/main.tsx", "src/App.tsx", "index.html", "src/index.css",
		"vite.config.ts", "tailwind.config.ts", "postcss.config.js", "tsconfig.json":
		return true
	}

	return strings.Contains(p, "/types/") || strings.HasPrefix(p, "types/") ||
		strings.Contains(p, "/stores/") || strings.HasPrefix(p, "stores/") ||
		strings.Contains(p, "/services/") || strings.HasPrefix(p, "services/") ||
		strings.Contains(p, "/hooks/") || strings.HasPrefix(p, "hooks/")
}

// minifyFileContext сжимает список имён уже сгенерированных файлов:
//  1. дедуп (generatedNames копит дубли при перезаписи путей),
//  2. ядро (isCoreContextFile) сохраняется всегда,
//  3. генерик-UI (components/ui/*) сбрасывается первым при переполнении,
//  4. из остальных оставляем самые свежие (хвост) в рамках бюджета maxContextFiles.
//
// На выходе — компактный, релевантный список для импортов, а не вся история проекта.
func minifyFileContext(names []string) []string {
	seen := make(map[string]struct{}, len(names))
	uniq := make([]string, 0, len(names))
	for _, n := range names {
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		uniq = append(uniq, n)
	}
	if len(uniq) <= maxContextFiles {
		return uniq
	}

	var priority, normal []string
	for _, n := range uniq {
		switch {
		case isCoreContextFile(n):
			priority = append(priority, n)
		case strings.Contains(n, "components/ui/"):
			// Генерик-UI (shadcn-примитивы) — выбрасываем из контекста: они
			// описаны в манифесте и системном промпте, перечислять их не нужно.
		default:
			normal = append(normal, n)
		}
	}

	budget := maxContextFiles - len(priority)
	if budget < 0 {
		budget = 0
	}
	if len(normal) > budget {
		// Хвост = самые недавно сгенерированные файлы (ближе к текущему тиру).
		normal = normal[len(normal)-budget:]
	}

	out := make([]string, 0, len(priority)+len(normal))
	out = append(out, priority...)
	out = append(out, normal...)

	return out
}

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
