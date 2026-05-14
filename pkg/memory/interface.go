package memory

import "github.com/mojomast/citewiseussy/pkg/packer"

// MemoryWriteBack stores accepted context plans and retrieves prior plans.
type MemoryWriteBack interface {
	StoreContextPlan(queryID string, plan packer.ContextPlan) error
	LoadPriorPlan(queryID string) (packer.ContextPlan, bool, error)
	SimilarPriorPlans(topics []string, limit int) ([]packer.ContextPlan, error)
}

// WriteBackPayload is the opaque memory payload callers can persist after plan
// acceptance or successful agent completion.
type WriteBackPayload struct {
	PlanHash      string               `json:"plan_hash"`
	QueryID       string               `json:"query_id"`
	QueryType     packer.QueryType     `json:"query_type"`
	Topics        []string             `json:"topics"`
	EvidencePath  []string             `json:"evidence_path"`
	NodeVersions  map[string]string    `json:"node_versions"`
	HygieneSignal packer.HygieneSignal `json:"hygiene_signal"`
	CreatedAt     string               `json:"created_at"`
}

// StoredPlan is one JSONL record in the file-backed memory store.
type StoredPlan struct {
	QueryID   string             `json:"query_id"`
	Topics    []string           `json:"topics"`
	Plan      packer.ContextPlan `json:"plan"`
	CreatedAt string             `json:"created_at"`
	PlanHash  string             `json:"plan_hash"`
}

// ReuseRejection explains why a prior plan cannot be reused.
type ReuseRejection struct {
	QueryID string `json:"query_id"`
	Reason  string `json:"reason"`
	Detail  string `json:"detail,omitempty"`
}

const (
	RejectVersionMismatch = "version-mismatch"
	RejectSuperseded      = "superseded"
	RejectAccessControl   = "access-control"
	RejectTopicMismatch   = "topic-mismatch"
	RejectQueryType       = "query-type"
)
