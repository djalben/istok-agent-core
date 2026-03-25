package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/istok/agent-core/internal/domain"
	"github.com/istok/agent-core/internal/infrastructure/openrouter"
	"github.com/istok/agent-core/internal/ports"
)

func main() {
	fmt.Println("=== Исток Agent - Enterprise AI System ===\n")

	agent := domain.NewAgent("agent-001", "Исток", 100000)
	fmt.Printf("✓ Created agent: %s (Balance: %d tokens)\n", agent.Name, agent.TokenBalance)

	agent.AddCapability(domain.NewCapability(
		"web_crawler",
		"Analyze websites and extract patterns",
		domain.CapabilityAdvanced,
	))
	agent.AddCapability(domain.NewCapability(
		"code_synthesis",
		"Generate production-ready code",
		domain.CapabilityExpert,
	))
	fmt.Printf("✓ Added %d capabilities\n\n", len(agent.Capabilities))

	snapshot1 := &domain.WebsiteSnapshot{
		ID:           "site-1",
		URL:          "https://react.dev",
		Title:        "React Documentation",
		Technologies: []string{"React", "JavaScript", "Webpack"},
		Confidence:   0.95,
		AnalyzedAt:   time.Now(),
	}

	snapshot2 := &domain.WebsiteSnapshot{
		ID:           "site-2",
		URL:          "https://go.dev",
		Title:        "Go Programming Language",
		Technologies: []string{"Go", "Docker", "Kubernetes"},
		Confidence:   0.92,
		AnalyzedAt:   time.Now(),
	}

	fmt.Println("=== Learning Phase ===")
	agent.LearnFromWebsite(snapshot1)
	fmt.Printf("✓ Learned from: %s\n", snapshot1.URL)
	agent.LearnFromWebsite(snapshot2)
	fmt.Printf("✓ Learned from: %s\n", snapshot2.URL)

	fmt.Printf("\nKnowledge Graph:\n")
	fmt.Printf("  - Nodes: %d\n", agent.GetKnowledgeNodeCount())
	fmt.Printf("  - Confidence: %.2f\n", agent.GetLearningConfidence())

	pattern := domain.NewPattern(
		domain.PatternTypeUI,
		"Component-Based Architecture",
		"Modern web apps use reusable components",
	)
	pattern.Examples = []string{"React Components", "Vue Components"}
	pattern.Frequency = 2
	pattern.Confidence = 0.9
	agent.AddPattern(pattern)

	insight := domain.NewInsight(
		"React + Go Stack Popular",
		"Both React and Go appear frequently in modern web applications",
		"technology_trend",
		0.88,
	)
	insight.Priority = 9
	agent.AddInsight(insight)

	fmt.Printf("\nLearning Context:\n")
	fmt.Printf("  - Patterns: %d\n", len(agent.LearningContext.Patterns))
	fmt.Printf("  - Insights: %d\n", len(agent.LearningContext.Insights))

	fmt.Println("\n=== Intelligence Analysis ===")
	intelligenceService := domain.NewAgentIntelligenceService()

	insights := intelligenceService.SynthesizeKnowledge(agent)
	fmt.Printf("✓ Synthesized %d insights\n", len(insights))
	for _, ins := range insights {
		fmt.Printf("  - %s (confidence: %.2f)\n", ins.Title, ins.Confidence)
	}

	strategy, confidence := intelligenceService.RecommendStrategy(agent, "code_generation")
	fmt.Printf("\n✓ Recommended strategy: %s (%.2f confidence)\n", strategy, confidence)

	task := domain.NewTask(
		agent.ID,
		"code_generation",
		"Generate React component with Go backend",
		8,
		5000,
	)

	riskScore, riskReason := intelligenceService.EvaluateRisk(agent, task)
	fmt.Printf("✓ Risk evaluation: %.2f - %s\n", riskScore, riskReason)

	fmt.Println("\n=== Task Execution ===")
	agent.EnqueueTask(task)
	fmt.Printf("✓ Enqueued task: %s\n", task.Description)

	nextTask := agent.GetNextTask()
	if nextTask != nil {
		fmt.Printf("✓ Processing task: %s (priority: %d)\n", nextTask.Description, nextTask.Priority)
		
		nextTask.Start()
		agent.UpdateStatus(domain.StatusCoding)
		
		agent.RecordDecision(
			nextTask.ID,
			"Use React + Go stack based on learning context",
			"Both technologies appear in analyzed websites with high confidence",
			0.88,
		)
		
		time.Sleep(100 * time.Millisecond)
		
		result := map[string]interface{}{
			"frontend": "React component generated",
			"backend":  "Go API handler generated",
			"tests":    "Unit tests included",
		}
		nextTask.Complete(result)
		
		agent.RecordTaskSuccess(4800, 2*time.Second)
		agent.UpdateStatus(domain.StatusIdle)
		
		fmt.Printf("✓ Task completed successfully\n")
		fmt.Printf("  - Tokens used: 4800\n")
		fmt.Printf("  - Duration: 2s\n")
	}

	fmt.Println("\n=== Performance Metrics ===")
	fmt.Printf("Total tasks: %d\n", agent.PerformanceMetrics.TotalTasks)
	fmt.Printf("Success rate: %.2f%%\n", agent.GetSuccessRate()*100)
	fmt.Printf("Avg tokens/task: %.0f\n", agent.PerformanceMetrics.AverageTokensPerTask)
	fmt.Printf("Avg execution time: %v\n", agent.PerformanceMetrics.AverageExecutionTime)

	fmt.Println("\n=== Decision Audit Trail ===")
	for i, decision := range agent.DecisionHistory {
		fmt.Printf("%d. %s\n", i+1, decision.Decision)
		fmt.Printf("   Reasoning: %s\n", decision.Reasoning)
		fmt.Printf("   Confidence: %.2f\n", decision.Confidence)
	}

	fmt.Println("\n=== AI Integration Demo ===")
	apiKey := "your-openrouter-api-key-here"
	if apiKey == "your-openrouter-api-key-here" {
		fmt.Println("⚠ Set OPENROUTER_API_KEY to test AI integration")
		fmt.Println("  Example: export OPENROUTER_API_KEY='sk-or-...'")
	} else {
		demoAIIntegration(agent, apiKey)
	}

	fmt.Println("\n=== Summary ===")
	fmt.Printf("✓ Agent '%s' operational\n", agent.Name)
	fmt.Printf("✓ Knowledge: %d nodes, %d patterns, %d insights\n",
		agent.GetKnowledgeNodeCount(),
		len(agent.LearningContext.Patterns),
		len(agent.LearningContext.Insights))
	fmt.Printf("✓ Capabilities: %d acquired\n", len(agent.Capabilities))
	fmt.Printf("✓ Performance: %.0f%% success rate\n", agent.GetSuccessRate()*100)
	fmt.Printf("✓ Token balance: %d remaining\n", agent.TokenBalance)
	
	fmt.Println("\n🚀 Исток Agent ready for autonomous operations!")
}

func demoAIIntegration(agent *domain.Agent, apiKey string) {
	ctx := context.Background()
	adapter := openrouter.NewCodeGeneratorAdapter(apiKey)

	fmt.Println("Testing AI code generation with fallback...")

	req := ports.GenerateCodeRequest{
		Specification: "Create a simple HTTP handler in Go",
		Language:      "Go",
		Framework:     "net/http",
	}

	estimate, err := adapter.EstimateCost(ctx, req)
	if err != nil {
		log.Printf("Cost estimation failed: %v", err)
		return
	}

	fmt.Printf("✓ Estimated cost: $%.4f (%d tokens)\n", estimate.EstimatedCost, estimate.EstimatedTokens)

	if agent.CanExecuteTask(estimate.EstimatedTokens) {
		response, err := adapter.GenerateCode(ctx, req)
		if err != nil {
			log.Printf("Code generation failed: %v", err)
			return
		}

		fmt.Printf("✓ Code generated successfully\n")
		fmt.Printf("  - Tokens used: %d\n", response.TokensUsed)
		fmt.Printf("  - Dependencies: %d\n", len(response.Dependencies))

		agent.DeductTokens(response.TokensUsed)
		fmt.Printf("  - Remaining balance: %d tokens\n", agent.TokenBalance)
	} else {
		fmt.Println("⚠ Insufficient tokens for generation")
	}

	telemetry := adapter.GetClient().GetTelemetry()
	stats := telemetry.GetOverallStats()
	fmt.Printf("\nTelemetry:\n")
	fmt.Printf("  - Total requests: %v\n", stats["total_requests"])
	fmt.Printf("  - Success rate: %.2f%%\n", stats["success_rate"].(float64)*100)
}
