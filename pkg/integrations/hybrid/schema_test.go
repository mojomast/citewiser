package hybrid

import (
	"strings"
	"testing"

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
