package hybrid

import (
	"encoding/json"
	"io"

	"github.com/mojomast/citewiser/pkg/access"
	"github.com/mojomast/citewiser/pkg/ragnode"
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

// RedactForReranker returns a handoff containing only nodes, candidates, and
// edges the caller may expose to an upstream reranker. It preserves safe
// relevance metadata for allowed candidates and drops unauthorized candidates
// rather than returning redacted text placeholders.
func RedactForReranker(ctx access.Context, controller access.Controller, h Handoff) Handoff {
	if controller == nil {
		controller = access.NewController()
	}
	allowed := map[string]bool{}
	redacted := Handoff{QueryID: h.QueryID, Query: h.Query}
	for _, node := range h.Nodes {
		if controller.CanSeeNode(ctx, node).Allowed {
			allowed[node.ID] = true
			redacted.Nodes = append(redacted.Nodes, node)
		}
	}
	for _, edge := range h.Edges {
		if !allowed[edge.SourceID] || !allowed[edge.TargetID] {
			continue
		}
		if controller.CanUseEdge(ctx, edge).Allowed {
			redacted.Edges = append(redacted.Edges, edge)
		}
	}
	for _, candidate := range h.Candidates {
		if allowed[candidate.NodeID] {
			redacted.Candidates = append(redacted.Candidates, candidate)
		}
	}
	return redacted
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
