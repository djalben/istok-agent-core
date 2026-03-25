package ports

import "context"

type GovernancePort interface {
	CheckPolicy(ctx context.Context, action string, context map[string]interface{}) (*PolicyCheckResult, error)
	RecordDecision(ctx context.Context, decision DecisionRecord) error
	AuditAction(ctx context.Context, action AuditAction) error
	GetComplianceStatus(ctx context.Context, agentID string) (*ComplianceStatus, error)
}

type PolicyCheckResult struct {
	Allowed     bool
	Reason      string
	Constraints []string
	RiskLevel   float64
}

type DecisionRecord struct {
	AgentID     string
	Decision    string
	Reasoning   string
	Confidence  float64
	Timestamp   string
	Context     map[string]interface{}
}

type AuditAction struct {
	AgentID   string
	Action    string
	Result    string
	Timestamp string
	Metadata  map[string]interface{}
}

type ComplianceStatus struct {
	IsCompliant     bool
	Violations      []string
	LastAudit       string
	ComplianceScore float64
}
