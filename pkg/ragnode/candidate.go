package ragnode

import "time"

type MethodScore struct {
	Method string  `json:"method"`
	Rank   int     `json:"rank"`
	Score  float64 `json:"score"`
	Weight float64 `json:"weight,omitempty"`
}

type Candidate struct {
	NodeID         string        `json:"node_id"`
	QueryRelevance float64       `json:"query_relevance"`
	MethodScores   []MethodScore `json:"method_scores,omitempty"`
	RerankerScore  float64       `json:"reranker_score,omitempty"`
	Rank           int           `json:"rank,omitempty"`
	RetrievalMode  string        `json:"retrieval_mode,omitempty"`
	SourceTrail    []SourceHop   `json:"source_trail,omitempty"`
}

type CandidateSet struct {
	QueryID     string      `json:"query_id"`
	Query       string      `json:"query"`
	Nodes       []RAGNode   `json:"nodes"`
	Edges       []Edge      `json:"edges"`
	Candidates  []Candidate `json:"candidates"`
	GeneratedAt time.Time   `json:"generated_at"`
}
