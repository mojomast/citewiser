package router

import "github.com/mojomast/citewiser/pkg/packer"

type RetrievalMode string

const (
	ModeGlobalGraph       RetrievalMode = "GlobalGraph"
	ModeLocalNeighborhood RetrievalMode = "LocalNeighborhood"
	ModeDRIFTChain        RetrievalMode = "DRIFTChain"
	ModeHybridBM25Dense   RetrievalMode = "HybridBM25Dense"
	ModeBasicVector       RetrievalMode = "BasicVector"
)

type GraphMetadata struct {
	EntityIDs         []string       `json:"entity_ids"`
	Topics            []string       `json:"topics"`
	CommunityIDs      []string       `json:"community_ids"`
	RoleCounts        map[string]int `json:"role_counts"`
	ChunkTypeCounts   map[string]int `json:"chunk_type_counts"`
	HasPermissionNode bool           `json:"has_permission_node"`
	HasDecisionBasis  bool           `json:"has_decision_basis"`
	MaxTopicSpan      int            `json:"max_topic_span"`
}

type Recommendation struct {
	QueryType            packer.QueryType `json:"query_type"`
	RetrievalMode        RetrievalMode    `json:"retrieval_mode"`
	ContextBudgetHint    int              `json:"context_budget_hint"`
	Reasons              []string         `json:"reasons"`
	CounterpointRequired bool             `json:"counterpoint_required"`
}

type QueryRouter interface {
	Route(query string, metadata GraphMetadata) Recommendation
}
