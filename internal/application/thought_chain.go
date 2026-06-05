package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/djalben/istok-agent-core/internal/domain"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Thought Chain (Reflective Reasoning)
//  Паттерн: [Goal] → [Hypothesis] → [Verification] → [Action]
//  Вдохновлено принципами рефлективного рассуждения.
//
//  LAYER MAPPING (5 Super-Agents → 10 ИСТОК Agents):
//  ┌─────────────────────────────────────────────────────────┐
//  │ Layer (external)     │ ИСТОК Agent(s)                   │
//  ├─────────────────────────────────────────────────────────┤
//  │ 1. Thinker/Planner   │ Director + Planner (DAG engine)  │
//  │ 2. Researcher        │ Researcher (3-iteration deep)    │
//  │ 3. Architect         │ Architect (Brain) + Designer     │
//  │ 4. Executor          │ Coder + Videographer             │
//  │ 5. Verifier          │ Validator + Security + Tester +  │
//  │                      │ UI Reviewer (VerificationGate)   │
//  └─────────────────────────────────────────────────────────┘
//
//  Key protocol differences:
//  - External uses single "Thinker" → ИСТОК splits into Director (strategy)
//    + Planner (DAG execution order) for better parallelism control.
//  - External "Verifier" is 1 agent → ИСТОК uses 4-agent VerificationGate
//    (Security + Tester + UI/UX + Integrity) for defense-in-depth.
//  - Reflective Reasoning injected at Director & Architect layers via
//    ThoughtChain() before artifact generation.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ThoughtChainResult содержит скрытый лог рассуждений агента.
type ThoughtChainResult struct {
	Goal         string `json:"goal"`
	Hypothesis   string `json:"hypothesis"`
	Verification string `json:"verification"`
	Action       string `json:"action"`
	RawChain     string `json:"raw_chain"`
}

// reflectiveReasoningPrompt — системный промпт для этапа Thought Chain.
// Имитирует логику глубокого рефлективного рассуждения перед генерацией артефактов.
const reflectiveReasoningPrompt = `You are ИСТОК Reflective Reasoning Engine.

Before producing ANY artifact, you MUST perform a structured Thought Chain:

## PROTOCOL (mandatory, execute silently):
1. [GOAL] — State the concrete objective in 1-2 sentences.
2. [HYPOTHESIS] — Formulate 2-3 alternative approaches. Evaluate trade-offs.
3. [VERIFICATION] — Challenge each hypothesis: What could go wrong? What edge cases exist? Which constraints from the specification are violated?
4. [ACTION] — Select the optimal path. Justify with evidence from verification.

## OUTPUT FORMAT:
<thought_chain>
[GOAL]: <concise goal statement>
[HYPOTHESIS_1]: <approach + trade-offs>
[HYPOTHESIS_2]: <approach + trade-offs>
[VERIFICATION]: <critical analysis of hypotheses>
[ACTION]: <chosen approach + justification>
</thought_chain>

Then produce the requested artifact AFTER the thought chain.

## RULES:
- Never skip verification. Always challenge your own assumptions.
- Prefer composable, modular solutions over monolithic ones.
- Consider scalability, maintainability, and developer experience.
- If the specification is ambiguous, resolve ambiguity in VERIFICATION step.`

// ThoughtChain выполняет этап рефлективного рассуждения перед генерацией артефактов.
// Публикует скрытый лог в EventBus и возвращает результат для включения в контекст.
func (o *Orchestrator) ThoughtChain(ctx context.Context, agent domain.AgentRole, task string) (*ThoughtChainResult, error) {
	o.sendStatus(ctx, agent, "reflecting", fmt.Sprintf("🧠 %s: рефлективное рассуждение...", agent), 5)
	o.busFromCtx(ctx).PublishReflection(agent, "Starting Thought Chain for: "+task)

	agentCfg := o.agents[agent]
	model := agentCfg.Model

	prompt := fmt.Sprintf(`Perform a Thought Chain analysis for this task:

TASK: %s

Output ONLY the <thought_chain> block. Be concise but thorough.`, task)

	result, err := o.callLLMWithReasoning(ctx, model, reflectiveReasoningPrompt, prompt, 2048)
	if err != nil {
		applog(ctx).WarnContext(ctx, "thought chain failed, proceeding without reflection",
			"agent", agent,
			"error", err,
		)

		return nil, err
	}

	chain := parseThoughtChain(result)
	chain.RawChain = result
	applog(ctx).InfoContext(ctx, "thought chain complete",
		"agent", agent,
		"goal", truncateChain(chain.Goal, 80),
		"action", truncateChain(chain.Action, 80),
	)
	o.busFromCtx(ctx).PublishReflection(agent, result)

	return chain, nil
}

// parseThoughtChain извлекает структурированные поля из raw thought chain.
func parseThoughtChain(raw string) *ThoughtChainResult {
	result := &ThoughtChainResult{}

	// Extract between <thought_chain> tags if present
	if idx := strings.Index(raw, "<thought_chain>"); idx != -1 {
		end := strings.Index(raw, "</thought_chain>")
		if end > idx {
			raw = raw[idx+len("<thought_chain>") : end]
		}
	}

	lines := strings.SplitSeq(raw, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "[GOAL]:"):
			result.Goal = strings.TrimSpace(strings.TrimPrefix(line, "[GOAL]:"))
		case strings.HasPrefix(line, "[ACTION]:"):
			result.Action = strings.TrimSpace(strings.TrimPrefix(line, "[ACTION]:"))
		case strings.HasPrefix(line, "[VERIFICATION]:"):
			result.Verification = strings.TrimSpace(strings.TrimPrefix(line, "[VERIFICATION]:"))
		case strings.HasPrefix(line, "[HYPOTHESIS_1]:"):
			result.Hypothesis = strings.TrimSpace(strings.TrimPrefix(line, "[HYPOTHESIS_1]:"))
		}
	}

	return result
}

// ThoughtChainContext формирует контекст из результата рассуждения для инъекции в основной промпт.
func ThoughtChainContext(tc *ThoughtChainResult) string {
	if tc == nil {
		return ""
	}
	var parts []string
	if tc.Goal != "" {
		parts = append(parts, "GOAL: "+tc.Goal)
	}
	if tc.Action != "" {
		parts = append(parts, "CHOSEN APPROACH: "+tc.Action)
	}
	if tc.Verification != "" {
		parts = append(parts, "VERIFIED: "+tc.Verification)
	}
	if len(parts) == 0 {
		return ""
	}

	return "\n\nREFLECTIVE CONTEXT (from Thought Chain):\n" + strings.Join(parts, "\n")
}

// truncateChain обрезает строку до указанной длины (для thought chain логов).
func truncateChain(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen] + "..."
}
