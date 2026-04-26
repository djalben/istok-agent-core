package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/istok/agent-core/internal/domain"
	"github.com/istok/agent-core/internal/ports"
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
		b.WriteString(fmt.Sprintf("Package: %s (manager: %s)\n", pc.PackageName, pc.PackageManager))
	}
	if len(pc.Dependencies) > 0 {
		b.WriteString("Dependencies (use these EXACT versions):\n")
		for k, v := range pc.Dependencies {
			b.WriteString(fmt.Sprintf("  %s@%s\n", k, v))
		}
	}
	if len(pc.DevDeps) > 0 {
		b.WriteString("DevDependencies:\n")
		for k, v := range pc.DevDeps {
			b.WriteString(fmt.Sprintf("  %s@%s\n", k, v))
		}
	}
	if pc.TSTarget != "" || pc.TSModule != "" {
		b.WriteString(fmt.Sprintf("TypeScript: target=%s module=%s strict=%v\n", pc.TSTarget, pc.TSModule, pc.TSStrict))
	}
	if len(pc.TSPaths) > 0 {
		b.WriteString("Path Aliases (USE THESE, no relative imports):\n")
		for alias, target := range pc.TSPaths {
			b.WriteString(fmt.Sprintf("  %s → %s\n", alias, target))
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
	Model           string   // например "anthropic/claude-3-7-sonnet-thinking"
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
func (p *PlannerAgent) ScanProject(packageJSONPath, tsconfigPath string) (*ProjectContext, error) {
	pc := &ProjectContext{}
	loadedAny := false

	// package.json
	if packageJSONPath != "" {
		data, err := os.ReadFile(packageJSONPath)
		if err == nil && len(data) > 0 {
			if err := parsePackageJSONInto(data, pc); err == nil {
				loadedAny = true
				log.Printf("📦 Planner: scanned %s — %d deps, %d devDeps", packageJSONPath, len(pc.Dependencies), len(pc.DevDeps))
			} else {
				log.Printf("⚠️ Planner: parse package.json failed: %v", err)
			}
		} else if err != nil {
			log.Printf("⚠️ Planner: read package.json failed: %v", err)
		}
	}

	// tsconfig.json
	if tsconfigPath != "" {
		data, err := os.ReadFile(tsconfigPath)
		if err == nil && len(data) > 0 {
			if err := parseTSConfigInto(data, pc); err == nil {
				loadedAny = true
				log.Printf("📘 Planner: scanned %s — target=%s, %d paths", tsconfigPath, pc.TSTarget, len(pc.TSPaths))
			} else {
				log.Printf("⚠️ Planner: parse tsconfig.json failed: %v", err)
			}
		} else if err != nil {
			log.Printf("⚠️ Planner: read tsconfig.json failed: %v", err)
		}
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
		if err := parsePackageJSONInto(packageJSON, pc); err == nil {
			loadedAny = true
		}
	}
	if len(tsconfigJSON) > 0 {
		if err := parseTSConfigInto(tsconfigJSON, pc); err == nil {
			loadedAny = true
		}
	}
	pc.Loaded = loadedAny
	return pc
}

func parsePackageJSONInto(data []byte, pc *ProjectContext) error {
	var pkg struct {
		Name            string            `json:"name"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Scripts         map[string]string `json:"scripts"`
		PackageManager  string            `json:"packageManager"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return err
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
			BaseURL string              `json:"baseUrl"`
			Strict  bool                `json:"strict"`
			Paths   map[string][]string `json:"paths"`
		} `json:"compilerOptions"`
	}
	if err := json.Unmarshal(data, &ts); err != nil {
		return err
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
// Возвращает Ready=true только если все условия выполнены.
func (p *PlannerAgent) ValidateReadiness(pc *ProjectContext) *ReadinessReport {
	r := &ReadinessReport{}
	for _, key := range p.RequiredEnvKeys {
		if os.Getenv(key) == "" {
			r.MissingEnvKeys = append(r.MissingEnvKeys, key)
		}
	}
	r.ContextLoaded = pc != nil && pc.Loaded

	switch {
	case len(r.MissingEnvKeys) > 0:
		r.Reason = fmt.Sprintf("missing env keys: %s", strings.Join(r.MissingEnvKeys, ", "))
	case !r.ContextLoaded:
		r.Reason = "project context not loaded (no package.json or tsconfig.json)"
	default:
		r.Ready = true
		r.Reason = "all readiness checks passed"
	}
	return r
}

// AdvanceToStrategySynthesized — smart FSM gate.
// Переход в StrategySynthesized разрешён только после успешного ValidateReadiness.
// При неудаче FSM остаётся в текущем состоянии и возвращается ошибка с причиной.
func (p *PlannerAgent) AdvanceToStrategySynthesized(fsm *domain.TaskStateMachine, pc *ProjectContext) error {
	if fsm == nil {
		return fmt.Errorf("planner: nil FSM")
	}
	report := p.ValidateReadiness(pc)
	if !report.Ready {
		log.Printf("🚫 Planner FSM gate BLOCKED: %s", report.Reason)
		return fmt.Errorf("planner readiness check failed: %s", report.Reason)
	}
	if err := fsm.TransitionTo(domain.StateStrategySynthesized, "planner: readiness verified"); err != nil {
		return fmt.Errorf("planner FSM transition failed: %w", err)
	}
	log.Printf("✅ Planner FSM gate PASSED → StrategySynthesized")
	return nil
}

// ────────────────────────────────────────────────────
//  DAG Engine — Step 1
// ────────────────────────────────────────────────────

// ValidateDAG проверяет граф задач на корректность:
// - все depends_on ссылаются на существующие ID
// - нет циклов (DAG-инвариант)
// - нет self-deps
func (p *PlannerAgent) ValidateDAG(plan *Plan) error {
	if plan == nil || len(plan.Tasks) == 0 {
		return fmt.Errorf("empty plan")
	}

	// Build ID → Task index
	ids := make(map[string]bool, len(plan.Tasks))
	for _, t := range plan.Tasks {
		if t.ID == "" {
			return fmt.Errorf("task with empty ID: %q", t.Title)
		}
		if ids[t.ID] {
			return fmt.Errorf("duplicate task ID: %s", t.ID)
		}
		ids[t.ID] = true
	}

	// Validate references + self-deps
	for _, t := range plan.Tasks {
		for _, dep := range t.DependsOn {
			if dep == t.ID {
				return fmt.Errorf("task %s has self-dependency", t.ID)
			}
			if !ids[dep] {
				return fmt.Errorf("task %s depends on missing task %s", t.ID, dep)
			}
		}
	}

	// Cycle detection via DFS with white/gray/black colouring
	const (
		white = 0 // не посещено
		gray  = 1 // в текущей DFS-ветке
		black = 2 // полностью обработано
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
				return fmt.Errorf("cycle detected: %s", strings.Join(append(path, d), " → "))
			case white:
				if err := dfs(d, path); err != nil {
					return err
				}
			}
		}
		colour[id] = black
		return nil
	}

	for _, t := range plan.Tasks {
		if colour[t.ID] == white {
			if err := dfs(t.ID, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

// TopologicalOrder возвращает task IDs в порядке исполнения (Kahn's algorithm).
// Задачи без зависимостей идут первыми.
func (p *PlannerAgent) TopologicalOrder(plan *Plan) ([]string, error) {
	if err := p.ValidateDAG(plan); err != nil {
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
		return nil, fmt.Errorf("topo sort incomplete: %d/%d (cycle?)", len(order), len(plan.Tasks))
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
		return nil, fmt.Errorf("planner: LLM provider not configured")
	}

	systemPrompt := `You are a senior software architect. Output a development plan as a Directed Acyclic Graph (DAG).
Each task must have a unique ID, depends_on (DAG edges), impacted_files, required_dependencies.
Architecture rules:
- Vite 5 + Bun + React 18 + TypeScript + TanStack Router/Query + shadcn/ui + TailwindCSS
- All imports MUST use @/* aliases (no relative paths)
- Mandatory directories: components/ui, components/layout, hooks, services, routes, lib, types
- Forms: react-hook-form + zod
- State: zustand or TanStack Query cache
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
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
		Temperature:  0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("planner LLM call failed: %w", err)
	}

	plan, err := parsePlanJSON(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("planner parse failed: %w", err)
	}

	// Synthesize DAG from steps if LLM didn't produce one
	if len(plan.Tasks) == 0 && len(plan.Steps) > 0 {
		log.Printf("⚠️ Planner: LLM returned no DAG, synthesizing from %d steps", len(plan.Steps))
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

	// Validate DAG
	if err := p.ValidateDAG(plan); err != nil {
		log.Printf("⚠️ Planner DAG validation failed: %v — flattening", err)
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
		log.Printf("⚠️ Planner topo sort failed: %v", err)
		order = nil
		for _, t := range plan.Tasks {
			order = append(order, t.ID)
		}
	}
	plan.ExecutionOrder = order

	log.Printf("✅ Planner: plan ready — %d tasks, exec order: %v", len(plan.Tasks), order)
	return plan, nil
}

// parsePlanJSON извлекает JSON-блок из ответа LLM (стрипает thinking-блоки и ```fences).
func parsePlanJSON(content string) (*Plan, error) {
	// Strip <thinking> blocks
	for strings.Contains(content, "<thinking>") {
		start := strings.Index(content, "<thinking>")
		end := strings.Index(content, "</thinking>")
		if end == -1 {
			break
		}
		content = content[:start] + content[end+len("</thinking>"):]
	}

	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	if first := strings.Index(content, "{"); first != -1 {
		if last := strings.LastIndex(content, "}"); last > first {
			content = content[first : last+1]
		}
	}

	var raw struct {
		Architecture string     `json:"architecture"`
		Components   []string   `json:"components"`
		Technologies []string   `json:"technologies"`
		Timeline     string     `json:"timeline"`
		Steps        []string   `json:"steps"`
		DAG          []PlanTask `json:"dag"`
	}
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return nil, err
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
