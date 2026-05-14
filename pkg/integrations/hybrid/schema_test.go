package hybrid

import (
	"strings"
	"testing"

	"github.com/mojomast/citewiseussy/pkg/access"
	"github.com/mojomast/citewiseussy/pkg/citewise"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

func TestMapHandoffPreservesRRFAndCrossEncoder(t *testing.T) {
	set := MapHandoff(Handoff{QueryID: "q", Candidates: []ragnode.Candidate{{NodeID: "policy", QueryRelevance: 1.2, RerankerScore: 0.97, RetrievalMode: "hybrid", MethodScores: []ragnode.MethodScore{{Method: "rrf", Rank: 1, Score: 0.046}, {Method: "cross-encoder", Rank: 1, Score: 0.97}}}}})
	if set.Candidates[0].QueryRelevance != 1 {
		t.Fatalf("query relevance should clamp to 1, got %.2f", set.Candidates[0].QueryRelevance)
	}
	if set.Candidates[0].RerankerScore != 0.97 || len(set.Candidates[0].MethodScores) != 2 {
		t.Fatalf("reranker/method scores not preserved: %+v", set.Candidates[0])
	}
}

func TestParseJSON(t *testing.T) {
	set, err := ParseJSON(strings.NewReader(`{"query_id":"q","candidates":[{"node_id":"n","query_relevance":0.93,"reranker_score":0.97,"method_scores":[{"method":"cross-encoder","rank":1,"score":0.97}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if set.QueryID != "q" || set.Candidates[0].MethodScores[0].Method != "cross-encoder" {
		t.Fatalf("bad parse: %+v", set)
	}
}

func TestRedactForRerankerDropsUnauthorizedNodesCandidatesAndEdges(t *testing.T) {
	handoff := Handoff{
		QueryID: "q",
		Query:   "rank these",
		Nodes: []ragnode.RAGNode{
			{Item: citewise.Item{ID: "allowed", Title: "Allowed", Source: "KB"}, Text: "safe", Sensitivity: ragnode.SensitivityInternal},
			{Item: citewise.Item{ID: "secret", Title: "Secret", Source: "Secret KB"}, Text: "hidden", Sensitivity: ragnode.SensitivityRestricted},
		},
		Edges: []ragnode.Edge{
			{SourceID: "allowed", TargetID: "secret", Type: ragnode.EdgeRelatedTo, Confidence: 1},
			{SourceID: "allowed", TargetID: "allowed", Type: ragnode.EdgeSameQuestion, Confidence: 0.8},
		},
		Candidates: []ragnode.Candidate{
			{NodeID: "allowed", QueryRelevance: 0.9, RerankerScore: 0.7, MethodScores: []ragnode.MethodScore{{Method: "rrf", Rank: 1, Score: 0.1}}},
			{NodeID: "secret", QueryRelevance: 1, RerankerScore: 1},
		},
	}

	redacted := RedactForReranker(access.Context{Clearance: access.ClearanceInternal}, nil, handoff)
	if len(redacted.Nodes) != 1 || redacted.Nodes[0].ID != "allowed" || redacted.Nodes[0].Text != "safe" {
		t.Fatalf("bad nodes after redaction: %+v", redacted.Nodes)
	}
	if len(redacted.Candidates) != 1 || redacted.Candidates[0].NodeID != "allowed" || redacted.Candidates[0].RerankerScore != 0.7 || len(redacted.Candidates[0].MethodScores) != 1 {
		t.Fatalf("bad candidates after redaction: %+v", redacted.Candidates)
	}
	if len(redacted.Edges) != 1 || redacted.Edges[0].SourceID != "allowed" || redacted.Edges[0].TargetID != "allowed" {
		t.Fatalf("bad edges after redaction: %+v", redacted.Edges)
	}
}
