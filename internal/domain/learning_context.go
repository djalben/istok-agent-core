package domain

import (
	"time"
)

type PatternType string

const (
	PatternTypeUI           PatternType = "ui"
	PatternTypeArchitecture PatternType = "architecture"
	PatternTypeBusiness     PatternType = "business"
	PatternTypeTechnology   PatternType = "technology"
	PatternTypeDesign       PatternType = "design"
)

type KnowledgeNodeType string

const (
	NodeTypeWebsite    KnowledgeNodeType = "website"
	NodeTypeTechnology KnowledgeNodeType = "technology"
	NodeTypePattern    KnowledgeNodeType = "pattern"
	NodeTypeInsight    KnowledgeNodeType = "insight"
)

type LearningContext struct {
	ID             string
	AgentID        string
	KnowledgeGraph *KnowledgeGraph
	Patterns       []*Pattern
	Insights       []*Insight
	Confidence     float64
	LastUpdated    time.Time
	Version        int64
	TotalNodes     int
	TotalEdges     int
}

type KnowledgeGraph struct {
	Nodes map[string]*KnowledgeNode
	Edges []*KnowledgeEdge
}

type KnowledgeNode struct {
	ID          string
	Type        KnowledgeNodeType
	Label       string
	Properties  map[string]interface{}
	Confidence  float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	SourceCount int
}

type KnowledgeEdge struct {
	ID         string
	FromNodeID string
	ToNodeID   string
	Relation   string
	Weight     float64
	Properties map[string]interface{}
	CreatedAt  time.Time
}

type Pattern struct {
	ID          string
	Type        PatternType
	Name        string
	Description string
	Examples    []string
	Frequency   int
	Confidence  float64
	Tags        []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Insight struct {
	ID          string
	Title       string
	Description string
	Category    string
	Confidence  float64
	Sources     []string
	Actionable  bool
	Priority    int
	CreatedAt   time.Time
	AppliedAt   *time.Time
}

type WebsiteSnapshot struct {
	ID           string
	URL          string
	Title        string
	Description  string
	Technologies []string
	Structure    map[string]interface{}
	Metadata     map[string]interface{}
	AnalyzedAt   time.Time
	Confidence   float64
}

func NewLearningContext(agentID string) *LearningContext {
	return &LearningContext{
		ID:      generateID(),
		AgentID: agentID,
		KnowledgeGraph: &KnowledgeGraph{
			Nodes: make(map[string]*KnowledgeNode),
			Edges: make([]*KnowledgeEdge, 0),
		},
		Patterns:    make([]*Pattern, 0),
		Insights:    make([]*Insight, 0),
		Confidence:  0.0,
		LastUpdated: time.Now(),
		Version:     1,
	}
}

func (lc *LearningContext) AddNode(node *KnowledgeNode) {
	if lc.KnowledgeGraph.Nodes == nil {
		lc.KnowledgeGraph.Nodes = make(map[string]*KnowledgeNode)
	}
	lc.KnowledgeGraph.Nodes[node.ID] = node
	lc.TotalNodes++
	lc.LastUpdated = time.Now()
	lc.Version++
}

func (lc *LearningContext) AddEdge(edge *KnowledgeEdge) {
	lc.KnowledgeGraph.Edges = append(lc.KnowledgeGraph.Edges, edge)
	lc.TotalEdges++
	lc.LastUpdated = time.Now()
	lc.Version++
}

func (lc *LearningContext) AddPattern(pattern *Pattern) {
	lc.Patterns = append(lc.Patterns, pattern)
	lc.recalculateConfidence()
	lc.LastUpdated = time.Now()
	lc.Version++
}

func (lc *LearningContext) AddInsight(insight *Insight) {
	lc.Insights = append(lc.Insights, insight)
	lc.recalculateConfidence()
	lc.LastUpdated = time.Now()
	lc.Version++
}

func (lc *LearningContext) GetNodesByType(nodeType KnowledgeNodeType) []*KnowledgeNode {
	nodes := make([]*KnowledgeNode, 0)
	for _, node := range lc.KnowledgeGraph.Nodes {
		if node.Type == nodeType {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (lc *LearningContext) GetPatternsByType(patternType PatternType) []*Pattern {
	patterns := make([]*Pattern, 0)
	for _, pattern := range lc.Patterns {
		if pattern.Type == patternType {
			patterns = append(patterns, pattern)
		}
	}
	return patterns
}

func (lc *LearningContext) GetActionableInsights() []*Insight {
	insights := make([]*Insight, 0)
	for _, insight := range lc.Insights {
		if insight.Actionable && insight.AppliedAt == nil {
			insights = append(insights, insight)
		}
	}
	return insights
}

func (lc *LearningContext) recalculateConfidence() {
	if len(lc.Patterns) == 0 && len(lc.Insights) == 0 {
		lc.Confidence = 0.0
		return
	}

	totalConfidence := 0.0
	count := 0

	for _, pattern := range lc.Patterns {
		totalConfidence += pattern.Confidence
		count++
	}

	for _, insight := range lc.Insights {
		totalConfidence += insight.Confidence
		count++
	}

	if count > 0 {
		lc.Confidence = totalConfidence / float64(count)
	}
}

func NewKnowledgeNode(nodeType KnowledgeNodeType, label string) *KnowledgeNode {
	return &KnowledgeNode{
		ID:         generateID(),
		Type:       nodeType,
		Label:      label,
		Properties: make(map[string]interface{}),
		Confidence: 1.0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func NewKnowledgeEdge(fromID, toID, relation string, weight float64) *KnowledgeEdge {
	return &KnowledgeEdge{
		ID:         generateID(),
		FromNodeID: fromID,
		ToNodeID:   toID,
		Relation:   relation,
		Weight:     weight,
		Properties: make(map[string]interface{}),
		CreatedAt:  time.Now(),
	}
}

func NewPattern(patternType PatternType, name, description string) *Pattern {
	return &Pattern{
		ID:          generateID(),
		Type:        patternType,
		Name:        name,
		Description: description,
		Examples:    make([]string, 0),
		Frequency:   1,
		Confidence:  0.5,
		Tags:        make([]string, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func NewInsight(title, description, category string, confidence float64) *Insight {
	return &Insight{
		ID:          generateID(),
		Title:       title,
		Description: description,
		Category:    category,
		Confidence:  confidence,
		Sources:     make([]string, 0),
		Actionable:  true,
		Priority:    5,
		CreatedAt:   time.Now(),
	}
}
