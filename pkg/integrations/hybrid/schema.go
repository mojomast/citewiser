package hybrid

import (
	"encoding/json"
	"io"

	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

type Handoff struct {
	QueryID    string              `json:"query_id,omitempty"`
	Query      string              `json:"query,omitempty"`
	Nodes      []ragnode.RAGNode   `json:"nodes,omitempty"`
	Edges      []ragnode.Edge      `json:"edges,omitempty"`
	Candidates []ragnode.Candidate `json:"candidates,omitempty"`
}

func ParseJSON(r io.Reader) (ragnode.CandidateSet, error) {
	var h Handoff
	if err := json.NewDecoder(r).Decode(&h); err != nil {
		return ragnode.CandidateSet{}, err
	}
	return MapHandoff(h), nil
}

func MapHandoff(h Handoff) ragnode.CandidateSet {
	candidates := make([]ragnode.Candidate, 0, len(h.Candidates))
	for _, candidate := range h.Candidates {
		candidate.QueryRelevance = clamp01(candidate.QueryRelevance)
		candidates = append(candidates, candidate)
	}
	return ragnode.CandidateSet{QueryID: h.QueryID, Query: h.Query, Nodes: h.Nodes, Edges: h.Edges, Candidates: candidates}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
