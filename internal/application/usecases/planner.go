package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/djalben/istok-agent-core/internal/domain"
	"github.com/djalben/istok-agent-core/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Planner Agent (Layer 2)
//  DAG engine + Context Injection + Smart FSM Gate
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ────────────────────────────────────────────────────
//  Plan & DAG types (canonical в usecases)
// ────────────────────────────────────────────────────

// PlanTask — узел DAG-плана разработки.
type PlanTask struct {
	ID                   string   `json:"id"`
	Title                string   `json:"title"`
	Description          string   `json:"description"`
	DependsOn            []string `json:"depends_on"`
	ImpactedFiles        []string `json:"impacted_files"`
	RequiredDependencies []string `json:"required_dependencies"`
}

// Plan — итоговый план разработки от Planner Agent.
type Plan struct {
	Architecture   string     `json:"architecture"`
	Components     []string   `json:"components"`
	Technologies   []string   `json:"technologies"`
	Timeline       string     `json:"timeline"`
	Steps          []string   `json:"steps"`           // backward-compat flat steps
	Tasks          []PlanTask `json:"dag"`             // DAG представление
	ExecutionOrder []string   `json:"execution_order"` // топологически отсортированные ID
}

// ────────────────────────────────────────────────────
//  Project Context (для Context Injection)
// ────────────────────────────────────────────────────

// ProjectContext — контекст проекта, считанный Планировщиком из package.json и tsconfig.json.
type ProjectContext struct {
	PackageName    string            `json:"package_name,omitempty"`
	Dependencies   map[string]string `json:"dependencies,omitempty"`
	DevDeps        map[string]string `json:"dev_dependencies,omitempty"`
	Scripts        map[string]string `json:"scripts,omitempty"`
	PackageManager string            `json:"package_manager,omitempty"`

	TSTarget  string            `json:"ts_target,omitempty"`
	TSModule  string            `json:"ts_module,omitempty"`
	TSPaths   map[string]string `json:"ts_paths,omitempty"`
	TSStrict  bool              `json:"ts_strict,omitempty"`
	TSBaseURL string            `json:"ts_base_url,omitempty"`

	Loaded bool `json:"loaded"` // true если package.json или tsconfig.json удалось распарсить
}

// ForPrompt — компактный текстовый блок для вставки в промпт LLM.
func (pc *ProjectContext) ForPrompt() string {
	if pc == nil || !pc.Loaded {
		return "\n## PROJECT CONTEXT\n(no project context loaded)\n"
	}
	var b strings.Builder
	b.WriteString("\n## PROJECT CONTEXT (scanned by Planner)\n")
	if pc.PackageName != "" {
		fmt.Fprintf(&b, "Package: %s (manager: %s)\n", pc.PackageName, pc.PackageManager)
	}
	if len(pc.Dependencies) > 0 {
		b.WriteString("Dependencies (use these EXACT versions):\n")
		for k, v := range pc.Dependencies {
			fmt.Fprintf(&b, "  %s@%s\n", k, v)
		}
	}
	if len(pc.DevDeps) > 0 {
		b.WriteString("DevDependencies:\n")
		for k, v := range pc.DevDeps {
			fmt.Fprintf(&b, "  %s@%s\n", k, v)
		}
	}
	if pc.TSTarget != "" || pc.TSModule != "" {
		fmt.Fprintf(&b, "TypeScript: target=%s module=%s strict=%v\n", pc.TSTarget, pc.TSModule, pc.TSStrict)
	}
	if len(pc.TSPaths) > 0 {
		b.WriteString("Path Aliases (USE THESE, no relative imports):\n")
		for alias, target := range pc.TSPaths {
			fmt.Fprintf(&b, "  %s → %s\n", alias, target)
		}
	}

	return b.String()
}

// ────────────────────────────────────────────────────
//  PlannerAgent
// ────────────────────────────────────────────────────

// PlannerAgent — модернизированный Агент-Планировщик с DAG-движком,
// инъекцией контекста и smart-связкой с FSM.
type PlannerAgent struct {
	LLM             ports.LLMProvider
	Model           string   // например "anthropic/claude-sonnet-4-6-thinking"
	RequiredEnvKeys []string // env vars обязательные для перехода в StrategySynthesized
}

// NewPlannerAgent создаёт планировщика с дефолтным набором обязательных env-ключей.
func NewPlannerAgent(llm ports.LLMProvider, model string) *PlannerAgent {
	return &PlannerAgent{
		LLM:   llm,
		Model: model,
		RequiredEnvKeys: []string{
			"ANTHROPIC_API_KEY",
			"REPLICATE_API_TOKEN",
		},
	}
}

// ────────────────────────────────────────────────────
//  Context Injection — Step 2
// ────────────────────────────────────────────────────

// ScanProject читает package.json и tsconfig.json по абсолютным путям.
// Возвращает ProjectContext с флагом Loaded=true если хотя бы один файл удалось распарсить.
func (p *PlannerAgent) ScanProject(ctx context.Context, packageJSONPath, tsconfigPath string) (*ProjectContext, error) {
	pc := &ProjectContext{}
	loadedAny := false
	l := ports.LoggerFromContext(ctx)

	if scanPlannerJSONFile(ctx, packageJSONPath, "package.json", func(data []byte) error {
		return parsePackageJSONInto(data, pc)
	}) {
		loadedAny = true
		l.InfoContext(ctx, "planner scanned package.json",
			"path", packageJSONPath,
			"deps", len(pc.Dependencies),
			"devDeps", len(pc.DevDeps),
		)
	}
	if scanPlannerJSONFile(ctx, tsconfigPath, "tsconfig.json", func(data []byte) error {
		return parseTSConfigInto(data, pc)
	}) {
		loadedAny = true
		l.InfoContext(ctx, "planner scanned tsconfig.json",
			"path", tsconfigPath,
			"target", pc.TSTarget,
			"paths", len(pc.TSPaths),
		)
	}

	pc.Loaded = loadedAny

	return pc, nil
}

// ScanProjectFromBytes альтернатива — принимает уже прочитанные байты.
// Удобно для тестов и для случая когда контент уже в памяти.
func (p *PlannerAgent) ScanProjectFromBytes(packageJSON, tsconfigJSON []byte) *ProjectContext {
	pc := &ProjectContext{}
	loadedAny := false
	if len(packageJSON) > 0 {
		err := parsePackageJSONInto(packageJSON, pc)
		if err == nil {
			loadedAny = true
		}
	}
	if len(tsconfigJSON) > 0 {
		err := parseTSConfigInto(tsconfigJSON, pc)
		if err == nil {
			loadedAny = true
		}
	}
	pc.Loaded = loadedAny

	return pc
}

func scanPlannerJSONFile(ctx context.Context, path, label string, parse func([]byte) error) bool {
	l := ports.LoggerFromContext(ctx)
	if path == "" {
		return false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		l.WarnContext(ctx, "planner read project file failed", "label", label, "error", wrapper.Wrap(err))

		return false
	}
	if len(data) == 0 {
		return false
	}
	err = parse(data)
	if err != nil {
		l.WarnContext(ctx, "planner parse project file failed", "label", label, "error", wrapper.Wrap(err))

		return false
	}

	return true
}

func parsePackageJSONInto(data []byte, pc *ProjectContext) error {
	var pkg struct {
		Name            string            `json:"name"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"` //nolint:tagliatelle // package.json schema
		Scripts         map[string]string `json:"scripts"`
		PackageManager  string            `json:"packageManager"` //nolint:tagliatelle // package.json schema
	}
	err := json.Unmarshal(data, &pkg)
	if err != nil {
		return wrapper.Wrap(err)
	}
	pc.PackageName = pkg.Name
	pc.Dependencies = pkg.Dependencies
	pc.DevDeps = pkg.DevDependencies
	pc.Scripts = pkg.Scripts
	switch {
	case strings.HasPrefix(pkg.PackageManager, "bun"):
		pc.PackageManager = "bun"
	case strings.HasPrefix(pkg.PackageManager, "pnpm"):
		pc.PackageManager = "pnpm"
	case strings.HasPrefix(pkg.PackageManager, "npm"):
		pc.PackageManager = "npm"
	default:
		pc.PackageManager = "npm"
	}

	return nil
}

func parseTSConfigInto(data []byte, pc *ProjectContext) error {
	var ts struct {
		CompilerOptions struct {
			Target  string              `json:"target"`
			Module  string              `json:"module"`
			BaseURL string              `json:"baseUrl"` //nolint:tagliatelle // tsconfig.json schema
			Strict  bool                `json:"strict"`
			Paths   map[string][]string `json:"paths"`
		} `json:"compilerOptions"` //nolint:tagliatelle // tsconfig.json schema
	}
	err := json.Unmarshal(data, &ts)
	if err != nil {
		return wrapper.Wrap(err)
	}
	pc.TSTarget = ts.CompilerOptions.Target
	pc.TSModule = ts.CompilerOptions.Module
	pc.TSBaseURL = ts.CompilerOptions.BaseURL
	pc.TSStrict = ts.CompilerOptions.Strict
	if len(ts.CompilerOptions.Paths) > 0 {
		pc.TSPaths = make(map[string]string)
		for alias, targets := range ts.CompilerOptions.Paths {
			if len(targets) > 0 {
				pc.TSPaths[alias] = targets[0]
			}
		}
	}

	return nil
}

// ────────────────────────────────────────────────────
//  Smart State Transition — Step 3 (FSM Gate)
// ────────────────────────────────────────────────────

// ReadinessReport — результат проверки готовности перед StrategySynthesized.
type ReadinessReport struct {
	Ready          bool     `json:"ready"`
	MissingEnvKeys []string `json:"missing_env_keys"`
	ContextLoaded  bool     `json:"context_loaded"`
	Reason         string   `json:"reason,omitempty"`
}

// ValidateReadiness проверяет наличие обязательных API-ключей и контекста проекта.
// Missing package.json не блокирует: Planner использует дефолтный шаблон.
func (p *PlannerAgent) ValidateReadiness(ctx context.Context, pc *ProjectContext) *ReadinessReport {
	r := &ReadinessReport{}
	l := ports.LoggerFromContext(ctx)
	for _, key := range p.RequiredEnvKeys {
		if os.Getenv(key) == "" {
			r.MissingEnvKeys = append(r.MissingEnvKeys, key)
		}
	}
	r.ContextLoaded = pc != nil && pc.Loaded

	switch {
	case len(r.MissingEnvKeys) > 0:
		r.Reason = "missing env keys: " + strings.Join(r.MissingEnvKeys, ", ")
	case !r.ContextLoaded:
		l.InfoContext(ctx, "planner no project context loaded, using default template")
		r.Ready = true
		r.ContextLoaded = true
		r.Reason = "no package.json found — using default Vite+React+TailwindCSS template"
	default:
		r.Ready = true
		r.Reason = "all readiness checks passed"
	}

	return r
}

// AdvanceToStrategySynthesized — smart FSM gate.
// Переход в StrategySynthesized разрешён только после успешного ValidateReadiness.
// При неудаче FSM остаётся в текущем состоянии и возвращается ошибка с причиной.
func (p *PlannerAgent) AdvanceToStrategySynthesized(ctx context.Context, fsm *domain.TaskStateMachine, pc *ProjectContext) error {
	if fsm == nil {
		return ErrPlannerNilFSM
	}
	l := ports.LoggerFromContext(ctx)
	report := p.ValidateReadiness(ctx, pc)
	if !report.Ready {
		l.WarnContext(ctx, "planner fsm gate blocked", "reason", report.Reason)

		return wrapper.Wrapf(ErrPlannerReadinessCheckFailed, "%s", report.Reason)
	}
	err := fsm.TransitionTo(domain.StateStrategySynthesized, "planner: readiness verified")
	if err != nil {
		return wrapper.Wrapf(ErrPlannerFSMTransitionFailed, "%v", err)
	}
	l.InfoContext(ctx, "planner fsm gate passed")

	return nil
}

// ────────────────────────────────────────────────────
//  DAG Engine — Step 1
// ────────────────────────────────────────────────────

// ValidateDAG проверяет граф задач на корректность:
// - все depends_on ссылаются на существующие ID
// - нет циклов (DAG-инвариант)
// - нет self-deps.
func (p *PlannerAgent) ValidateDAG(plan *Plan) error {
	if plan == nil || len(plan.Tasks) == 0 {
		return ErrEmptyPlan
	}
	err := validatePlanTaskRefs(plan)
	if err != nil {
		return err
	}

	return detectPlanDAGCycles(plan)
}

func validatePlanTaskRefs(plan *Plan) error {
	ids := make(map[string]bool, len(plan.Tasks))
	for _, t := range plan.Tasks {
		if t.ID == "" {
			return wrapper.Wrapf(ErrTaskEmptyID, "%q", t.Title)
		}
		if ids[t.ID] {
			return wrapper.Wrapf(ErrDuplicateTaskID, "%s", t.ID)
		}
		ids[t.ID] = true
	}
	for _, t := range plan.Tasks {
		for _, dep := range t.DependsOn {
			if dep == t.ID {
				return wrapper.Wrapf(ErrTaskSelfDependency, "%s", t.ID)
			}
			if !ids[dep] {
				return wrapper.Wrapf(ErrTaskMissingDependency, "task %s depends on missing task %s", t.ID, dep)
			}
		}
	}

	return nil
}

func detectPlanDAGCycles(plan *Plan) error {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	colour := make(map[string]int, len(plan.Tasks))
	deps := make(map[string][]string, len(plan.Tasks))
	for _, t := range plan.Tasks {
		deps[t.ID] = t.DependsOn
	}

	var dfs func(id string, path []string) error
	dfs = func(id string, path []string) error {
		colour[id] = gray
		path = append(path, id)
		for _, d := range deps[id] {
			switch colour[d] {
			case gray:
				return wrapper.Wrapf(ErrDAGCycleDetected, "%s", strings.Join(append(path, d), " → "))
			case white:
				err := dfs(d, path)
				if err != nil {
					return err
				}
			}
		}
		colour[id] = black

		return nil
	}

	for _, t := range plan.Tasks {
		if colour[t.ID] == white {
			err := dfs(t.ID, nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// TopologicalOrder возвращает task IDs в порядке исполнения (Kahn's algorithm).
// Задачи без зависимостей идут первыми.
func (p *PlannerAgent) TopologicalOrder(plan *Plan) ([]string, error) {
	err := p.ValidateDAG(plan)
	if err != nil {
		return nil, err
	}

	// Build in-degree map and adjacency
	inDeg := make(map[string]int, len(plan.Tasks))
	revAdj := make(map[string][]string, len(plan.Tasks)) // dep → [tasks that depend on it]
	for _, t := range plan.Tasks {
		if _, ok := inDeg[t.ID]; !ok {
			inDeg[t.ID] = 0
		}
		for _, d := range t.DependsOn {
			inDeg[t.ID]++
			revAdj[d] = append(revAdj[d], t.ID)
		}
	}

	// Initial queue: tasks with no deps
	var queue []string
	for _, t := range plan.Tasks {
		if inDeg[t.ID] == 0 {
			queue = append(queue, t.ID)
		}
	}

	order := make([]string, 0, len(plan.Tasks))
	for len(queue) > 0 {
		head := queue[0]
		queue = queue[1:]
		order = append(order, head)
		for _, next := range revAdj[head] {
			inDeg[next]--
			if inDeg[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(order) != len(plan.Tasks) {
		return nil, wrapper.Wrapf(ErrTopoSortIncomplete, "%d/%d (cycle?)", len(order), len(plan.Tasks))
	}

	return order, nil
}

// ────────────────────────────────────────────────────
//  BuildPlan — главный entry point
// ────────────────────────────────────────────────────

// BuildPlan вызывает LLM с инъекцией контекста проекта, парсит ответ,
// валидирует DAG, выполняет топологическую сортировку.
// Возвращает готовый Plan с заполненным ExecutionOrder.
func (p *PlannerAgent) BuildPlan(ctx context.Context, specification, auditSummary string, pc *ProjectContext) (*Plan, error) {
	if p.LLM == nil {
		return nil, ErrPlannerLLMNotConfigured
	}

	systemPrompt := `You are a senior software architect. Output a development plan as a Directed Acyclic Graph (DAG).
Each task must have a unique ID, depends_on (DAG edges), impacted_files, required_dependencies.
Architecture rules:
- Vite 5 + Bun + React 18 + TypeScript + TanStack Router/Query + shadcn/ui + TailwindCSS
- All imports MUST use @/* aliases (no relative paths)
- Mandatory directories: components/ui, components/layout, hooks, services, routes, lib, types
- Forms: react-hook-form + zod
- State: zustand or TanStack Query cache

## REFLECTION BLOCK (execute BEFORE outputting JSON):
Before producing your final answer, verify against these 3 critical errors:
1. [ORPHAN CHECK] — Are there any tasks with depends_on referencing non-existent IDs? Fix them.
2. [COMPLETENESS CHECK] — Does the DAG cover ALL features from the specification? If a feature is missing, add a task.
3. [PARALLELISM CHECK] — Are independent tasks correctly marked with no shared dependencies? Maximize parallel execution paths.
If any check fails, silently correct the DAG before output.

Output ONLY valid JSON. No markdown, no commentary.`

	userPrompt := fmt.Sprintf(`Build a DAG plan for this project.

SPECIFICATION:
%s

DESIGN AUDIT:
%s
%s

Output JSON shape:
{
  "architecture": "...",
  "components": ["..."],
  "technologies": ["vite","react","tailwindcss","..."],
  "timeline": "...",
  "steps": ["Step 1: ...", "Step 2: ..."],
  "dag": [
    {"id":"T1","title":"Scaffold","description":"Init Vite project","depends_on":[],"impacted_files":["package.json","vite.config.ts"],"required_dependencies":["vite","react","react-dom"]},
    {"id":"T2","title":"Layout","description":"Build AppLayout","depends_on":["T1"],"impacted_files":["src/components/layout/AppLayout.tsx"],"required_dependencies":["@radix-ui/react-slot"]}
  ]
}

CRITICAL:
- Every task ID must be unique (T1, T2, T3, ...)
- depends_on may only reference existing IDs
- No self-dependencies, no cycles
- Coding/UI tasks must depend on scaffold/architecture tasks
- Use EXACT package names from PROJECT CONTEXT if available`,
		specification, auditSummary, pc.ForPrompt())

	resp, err := p.LLM.Complete(ctx, ports.LLMRequest{
		Model:        p.Model,
		SystemPrompt: "Strict Rule: Minimise reasoning. No conversational fillers. Be concise. Use White Label (Istok Core only).\n\n" + systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    8192,
		Reasoning:    true,
	})
	if err != nil {
		return nil, wrapper.Wrapf(ErrPlannerLLMCallFailed, "%v", err)
	}

	l := ports.LoggerFromContext(ctx)
	plan, err := parsePlanJSON(resp.Content)
	if err != nil {
		l.WarnContext(ctx, "planner parse error",
			"error", wrapper.Wrap(err),
			"preview", resp.Content[:min(len(resp.Content), 200)],
		)

		return nil, wrapper.Wrapf(ErrPlannerParseFailed, "%v", err)
	}

	// Synthesize DAG from steps if LLM didn't produce one
	if len(plan.Tasks) == 0 && len(plan.Steps) > 0 {
		l.InfoContext(ctx, "planner synthesizing dag from steps", "steps", len(plan.Steps))
		for i, step := range plan.Steps {
			var deps []string
			if i > 0 {
				deps = []string{fmt.Sprintf("T%d", i)}
			}
			plan.Tasks = append(plan.Tasks, PlanTask{
				ID:          fmt.Sprintf("T%d", i+1),
				Title:       step,
				Description: step,
				DependsOn:   deps,
			})
		}
	}

	// Hard floor: если LLM вернул пустоту (Opus 4.7 иногда выдаёт пустые массивы
	// на очень сложных спецах) — явная ошибка, не прячем симптом.
	if len(plan.Tasks) == 0 {
		l.WarnContext(ctx, "planner empty llm plan")

		return nil, ErrPlannerEmptyLLMResponse
	}

	// Floor: minimum 3 задачи для осмысленного DAG (scaffold + UI + features).
	// Дополняем базовыми задачами, не бросая людей в broken pipeline.
	if len(plan.Tasks) < 3 {
		l.InfoContext(ctx, "planner padding tasks with default scaffold", "tasks", len(plan.Tasks))
		plan.Tasks = padWithDefaultTasks(plan.Tasks)
	}

	// Validate DAG
	err = p.ValidateDAG(plan)
	if err != nil {
		l.WarnContext(ctx, "planner dag validation failed, flattening", "error", wrapper.Wrap(err))
		// Recovery: drop deps and produce a linear chain
		flat := make([]PlanTask, 0, len(plan.Tasks))
		for i, t := range plan.Tasks {
			t.ID = fmt.Sprintf("T%d", i+1)
			if i == 0 {
				t.DependsOn = nil
			} else {
				t.DependsOn = []string{fmt.Sprintf("T%d", i)}
			}
			flat = append(flat, t)
		}
		plan.Tasks = flat
	}

	// Topological execution order
	order, err := p.TopologicalOrder(plan)
	if err != nil {
		l.WarnContext(ctx, "planner topo sort failed", "error", wrapper.Wrap(err))
		order = nil
		for _, t := range plan.Tasks {
			order = append(order, t.ID)
		}
	}
	plan.ExecutionOrder = order
	l.InfoContext(ctx, "planner plan ready",
		"tasks", len(plan.Tasks),
		"execOrder", order,
	)

	return plan, nil
}

// padWithDefaultTasks добивает план до минимума 4 базовыми задачами, если LLM вернул
// вырожденный DAG. Сохраняет существующие задачи в начале цепочки.
func padWithDefaultTasks(existing []PlanTask) []PlanTask {
	defaults := []PlanTask{
		{ID: "T1", Title: "Project scaffold", Description: "Initialize Vite + React 18 + TypeScript project with TailwindCSS, shadcn/ui and @/* aliases", DependsOn: nil, ImpactedFiles: []string{"package.json", "vite.config.ts", "tsconfig.json"}, RequiredDependencies: []string{"vite", "react", "react-dom", "tailwindcss"}},
		{ID: "T2", Title: "UI shell", Description: "Build AppLayout with Sidebar, Header and route container", DependsOn: []string{"T1"}, ImpactedFiles: []string{"src/components/layout/AppLayout.tsx"}, RequiredDependencies: []string{"@radix-ui/react-slot", "lucide-react"}},
		{ID: "T3", Title: "Data layer", Description: "Create TanStack Query hooks and API services", DependsOn: []string{"T1"}, ImpactedFiles: []string{"src/hooks/useApi.ts", "src/services/api.ts"}, RequiredDependencies: []string{"@tanstack/react-query"}},
		{ID: "T4", Title: "Feature pages", Description: "Build all route pages consuming hooks from T3 and layout from T2", DependsOn: []string{"T2", "T3"}, ImpactedFiles: []string{"src/pages/Home.tsx"}, RequiredDependencies: nil},
	}

	// Привязываем default-задачи к последней существующей, чтобы сохранить DAG-целостность.
	out := append([]PlanTask(nil), existing...)
	need := 4 - len(out)
	if need <= 0 {
		return out
	}
	lastID := ""
	if len(out) > 0 {
		lastID = out[len(out)-1].ID
	}
	nextNum := len(out) + 1
	for i := range need {
		t := defaults[i%len(defaults)]
		t.ID = fmt.Sprintf("T%d", nextNum)
		nextNum++
		if lastID != "" {
			t.DependsOn = []string{lastID}
		}
		lastID = t.ID
		out = append(out, t)
	}

	return out
}

// parsePlanJSON извлекает JSON-блок из ответа LLM (стрипает thinking-блоки и ```fences).
// Использует bracket-counting (ExtractFirstJSONObject) для устойчивости к длинным ответам
// Opus 4.7, где модель добавляет prose ДО или ПОСЛЕ JSON.
func parsePlanJSON(content string) (*Plan, error) {
	jsonBlock, ok := ExtractFirstJSONObject(content)
	if !ok {
		return nil, wrapper.Wrapf(ErrPlannerNoJSONObject, "len=%d", len(content))
	}

	var raw struct {
		Architecture string     `json:"architecture"`
		Components   []string   `json:"components"`
		Technologies []string   `json:"technologies"`
		Timeline     string     `json:"timeline"`
		Steps        []string   `json:"steps"`
		DAG          []PlanTask `json:"dag"`
	}
	err := json.Unmarshal([]byte(jsonBlock), &raw)
	if err != nil {
		return nil, wrapper.Wrapf(ErrPlannerJSONUnmarshalFailed, "block_len=%d: %v | first 300: %.300s",
			len(jsonBlock), err, jsonBlock)
	}

	return &Plan{
		Architecture: raw.Architecture,
		Components:   raw.Components,
		Technologies: raw.Technologies,
		Timeline:     raw.Timeline,
		Steps:        raw.Steps,
		Tasks:        raw.DAG,
	}, nil
}

// ExtractFirstJSONObject находит первый завершённый JSON-объект в строке методом
// bracket-counting (учитывая строки и escape-последовательности). Стрипает thinking-блоки
// и ```fences. Используется парсерами Planner/Director для устойчивости к Opus 4.7,
// который может выдавать prose ДО и ПОСЛЕ JSON-объекта.
//
// Возвращает (json_string, true) при успехе, либо ("", false) если объект не найден.
func ExtractFirstJSONObject(content string) (string, bool) {
	// 1. Strip <thinking>...</thinking> blocks
	for strings.Contains(content, "<thinking>") {
		start := strings.Index(content, "<thinking>")
		end := strings.Index(content, "</thinking>")
		if end == -1 {
			break
		}
		content = content[:start] + content[end+len("</thinking>"):]
	}

	// 2. Strip markdown fences
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```JSON")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")

	// 3. Bracket-counting: найти первый '{' и сматчить парный '}'.
	first := strings.Index(content, "{")
	if first == -1 {
		return "", false
	}
	depth := 0
	inStr := false
	escape := false
	for i := first; i < len(content); i++ {
		c := content[i]
		if escape {
			escape = false

			continue
		}
		if c == '\\' && inStr {
			escape = true

			continue
		}
		if c == '"' {
			inStr = !inStr

			continue
		}
		if inStr {
			continue
		}
		switch c {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return content[first : i+1], true
			}
		}
	}
	// Не нашли парный '}' — fallback на старый LastIndex (могут быть оборванные ответы).
	if last := strings.LastIndex(content, "}"); last > first {
		return content[first : last+1], true
	}

	return "", false
}
