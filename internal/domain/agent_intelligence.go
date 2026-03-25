package domain

import (
	"fmt"
	"strings"
)

type AgentIntelligenceService struct{}

func NewAgentIntelligenceService() *AgentIntelligenceService {
	return &AgentIntelligenceService{}
}

func (s *AgentIntelligenceService) CanLearnFrom(agent *Agent, snapshot *WebsiteSnapshot) (bool, string) {
	if snapshot.Confidence < 0.3 {
		return false, "website analysis confidence too low"
	}

	if len(snapshot.Technologies) == 0 {
		return false, "no technologies detected"
	}

	existingNodes := agent.LearningContext.GetNodesByType(NodeTypeWebsite)
	for _, node := range existingNodes {
		if url, ok := node.Properties["url"].(string); ok && url == snapshot.URL {
			return false, "website already analyzed"
		}
	}

	if agent.LearningContext.TotalNodes > 10000 {
		return false, "learning context capacity reached"
	}

	return true, "website suitable for learning"
}

func (s *AgentIntelligenceService) SynthesizeKnowledge(agent *Agent) []*Insight {
	insights := make([]*Insight, 0)

	techFrequency := make(map[string]int)
	websites := agent.LearningContext.GetNodesByType(NodeTypeWebsite)
	
	for _, website := range websites {
		if techs, ok := website.Properties["technologies"].([]string); ok {
			for _, tech := range techs {
				techFrequency[tech]++
			}
		}
	}

	for tech, count := range techFrequency {
		if count >= 3 {
			confidence := float64(count) / float64(len(websites))
			insight := NewInsight(
				fmt.Sprintf("Popular Technology: %s", tech),
				fmt.Sprintf("Technology %s appears in %d of %d analyzed websites", tech, count, len(websites)),
				"technology_trend",
				confidence,
			)
			insight.Priority = 7
			insights = append(insights, insight)
		}
	}

	patterns := agent.LearningContext.GetPatternsByType(PatternTypeUI)
	if len(patterns) > 5 {
		insight := NewInsight(
			"UI Pattern Library Available",
			fmt.Sprintf("Agent has learned %d UI patterns that can be reused", len(patterns)),
			"capability",
			0.9,
		)
		insight.Priority = 8
		insights = append(insights, insight)
	}

	return insights
}

func (s *AgentIntelligenceService) RecommendStrategy(agent *Agent, taskType string) (string, float64) {
	if agent.LearningContext.TotalNodes == 0 {
		return "explore_and_learn", 0.5
	}

	switch taskType {
	case "website_analysis":
		websites := agent.LearningContext.GetNodesByType(NodeTypeWebsite)
		if len(websites) > 10 {
			return "apply_learned_patterns", 0.85
		}
		return "analyze_with_context", 0.7

	case "code_generation":
		patterns := agent.LearningContext.Patterns
		if len(patterns) > 20 {
			return "pattern_based_generation", 0.9
		}
		return "standard_generation", 0.6

	case "domain_registration":
		insights := agent.LearningContext.GetActionableInsights()
		if len(insights) > 0 {
			return "insight_driven_registration", 0.8
		}
		return "standard_registration", 0.5

	default:
		return "adaptive_approach", 0.6
	}
}

func (s *AgentIntelligenceService) EvaluateRisk(agent *Agent, task *Task) (float64, string) {
	riskScore := 0.0
	reasons := make([]string, 0)

	if task.TokenCost > agent.TokenBalance {
		riskScore += 0.5
		reasons = append(reasons, "insufficient tokens")
	}

	if agent.PerformanceMetrics.SuccessRate < 0.5 && agent.PerformanceMetrics.TotalTasks > 10 {
		riskScore += 0.3
		reasons = append(reasons, "low historical success rate")
	}

	if agent.LearningContext.Confidence < 0.3 {
		riskScore += 0.2
		reasons = append(reasons, "low learning confidence")
	}

	requiredCapability := s.getRequiredCapability(task.Type)
	if requiredCapability != "" && !agent.HasCapability(requiredCapability) {
		riskScore += 0.4
		reasons = append(reasons, fmt.Sprintf("missing capability: %s", requiredCapability))
	}

	if riskScore > 1.0 {
		riskScore = 1.0
	}

	reasonStr := strings.Join(reasons, "; ")
	if reasonStr == "" {
		reasonStr = "low risk"
	}

	return riskScore, reasonStr
}

func (s *AgentIntelligenceService) getRequiredCapability(taskType string) string {
	capabilityMap := map[string]string{
		"website_analysis":    "web_crawler",
		"code_generation":     "code_synthesis",
		"domain_registration": "domain_management",
		"deployment":          "infrastructure_management",
	}
	return capabilityMap[taskType]
}

func (s *AgentIntelligenceService) SuggestNextAction(agent *Agent) string {
	if len(agent.TaskQueue) > 0 {
		return "process_task_queue"
	}

	if agent.LearningContext.TotalNodes < 5 {
		return "explore_websites"
	}

	insights := agent.LearningContext.GetActionableInsights()
	if len(insights) > 0 {
		return "apply_insights"
	}

	if agent.PerformanceMetrics.SuccessRate < 0.7 && agent.PerformanceMetrics.TotalTasks > 5 {
		return "improve_capabilities"
	}

	return "await_instructions"
}

func (s *AgentIntelligenceService) OptimizeTokenUsage(agent *Agent, estimatedCost int64) (bool, string) {
	if estimatedCost > agent.TokenBalance {
		return false, "insufficient token balance"
	}

	avgCost := agent.PerformanceMetrics.AverageTokensPerTask
	if avgCost > 0 && float64(estimatedCost) > avgCost*2 {
		return false, "cost significantly higher than average"
	}

	utilizationRate := float64(agent.PerformanceMetrics.TotalTokensUsed) / float64(agent.TokenBalance+agent.PerformanceMetrics.TotalTokensUsed)
	if utilizationRate > 0.9 {
		return false, "approaching token budget limit"
	}

	return true, "token usage within acceptable limits"
}
