package ranker

import (
	"strings"
	"testing"

	"github.com/mojomast/citewiser/pkg/access"
	"github.com/mojomast/citewiser/pkg/citewise"
	"github.com/mojomast/citewiser/pkg/ragnode"
)

func TestRankerComputesDefaultFormula(t *testing.T) {
	analysis := rankAnalysis([]ragnode.RAGNode{rankNode("a", 1)}, map[string]float64{"a": 1})
	analysis.Centrality["a"] = 1
	set, err := NewRanker().Rank(access.Context{Clearance: access.ClearanceInternal}, analysis, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Ranked) != 1 {
		t.Fatalf("ranked count got %d", len(set.Ranked))
	}
	score := set.Ranked[0].Score
	if got, want := score.Total, 83.2; !close(got, want) {
		t.Fatalf("total got %.6f want %.6f: %+v", got, want, score)
	}
	if got, want := score.GraphImportance, 0.55; !close(got, want) {
		t.Fatalf("graph importance got %.6f want %.6f", got, want)
	}
}

func TestRankerAccessSuppressionDoesNotExposeNode(t *testing.T) {
	node := rankNode("secret", 1)
	node.Text = "hidden"
	node.Source = "secret source"
	node.Sensitivity = ragnode.SensitivityConfidential
	analysis := rankAnalysis([]ragnode.RAGNode{node}, map[string]float64{"secret": 1})
	set, err := NewRanker().Rank(access.Context{Clearance: access.ClearancePublic}, analysis, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Ranked) != 0 || len(set.Suppressed) != 1 {
		t.Fatalf("unexpected rank/suppress counts: %+v", set)
	}
	suppressed := set.Suppressed[0]
	if suppressed.NodeID != "secret" || suppressed.SuppressionReason != access.ReasonAccessControl || suppressed.AccessAllowed {
		t.Fatalf("bad suppression: %+v", suppressed)
	}
	if strings.Contains(strings.Join(suppressed.Rationale, " "), "hidden") || strings.Contains(strings.Join(suppressed.Rationale, " "), "secret source") {
		t.Fatalf("suppression leaked node content: %+v", suppressed)
	}
}

func TestRankerAppliesQueryTypeModifier(t *testing.T) {
	node := rankNode("perm", 1)
	node.ChunkType = ragnode.ChunkPermissionRecord
	analysis := rankAnalysis([]ragnode.RAGNode{node}, map[string]float64{"perm": 1})
	plain, err := NewRanker().Rank(access.Context{Clearance: access.ClearanceInternal}, analysis, 100)
	if err != nil {
		t.Fatal(err)
	}
	agentic, err := NewRanker().Rank(access.Context{Clearance: access.ClearanceInternal, Attributes: map[string]string{"query_type": "Agentic"}}, analysis, 100)
	if err != nil {
		t.Fatal(err)
	}
	if agentic.Ranked[0].Score.Total-plain.Ranked[0].Score.Total != 20 {
		t.Fatalf("agentic modifier got plain %.1f agentic %.1f", plain.Ranked[0].Score.Total, agentic.Ranked[0].Score.Total)
	}
}

func TestRankerPenaltiesAndRationale(t *testing.T) {
	node := rankNode("old", 0.1)
	node.SupersededBy = "new"
	node.TokenCount = 0
	node.Text = "abcdefghijklmnopqrstuvwxyzabcdefghi"
	analysis := rankAnalysis([]ragnode.RAGNode{node}, map[string]float64{"old": 1})
	analysis.Duplicates = [][]string{{"old", "copy"}}
	analysis.EdgesIn["old"] = []ragnode.Edge{{SourceID: "x", TargetID: "old", Type: ragnode.EdgeRelatedTo, Note: "raw_type=weird"}}
	set, err := NewRanker().Rank(access.Context{Clearance: access.ClearanceInternal}, analysis, 100)
	if err != nil {
		t.Fatal(err)
	}
	score := set.Ranked[0].Score
	if score.RedundancyPenalty != 1 || score.StalenessPenalty != 1 || score.LowTrustPenalty == 0 {
		t.Fatalf("expected penalties, got %+v", score)
	}
	rationale := strings.Join(score.Rationale, " ")
	for _, phrase := range []string{"token count estimated", "unknown edge type", "redundancy penalty", "staleness penalty", "low trust penalty"} {
		if !strings.Contains(rationale, phrase) {
			t.Fatalf("missing rationale %q in %q", phrase, rationale)
		}
	}
}

func TestRankerOmitsEstimateRationaleForExplicitTokenCount(t *testing.T) {
	node := rankNode("explicit", 1)
	node.Text = "abcdefghijklmnopqrstuvwxyzabcdefghi"
	node.TokenCount = 9
	analysis := rankAnalysis([]ragnode.RAGNode{node}, map[string]float64{"explicit": 1})
	set, err := NewRanker().Rank(access.Context{Clearance: access.ClearanceInternal}, analysis, 100)
	if err != nil {
		t.Fatal(err)
	}
	if rationale := strings.Join(set.Ranked[0].Score.Rationale, " "); strings.Contains(rationale, "token count estimated") {
		t.Fatalf("explicit token count should not be marked estimated: %q", rationale)
	}
}

func TestRankerStableSortsByTotalThenNodeID(t *testing.T) {
	analysis := rankAnalysis([]ragnode.RAGNode{rankNode("b", 1), rankNode("a", 1)}, map[string]float64{"a": 1, "b": 1})
	set, err := NewRanker().Rank(access.Context{Clearance: access.ClearanceInternal}, analysis, 100)
	if err != nil {
		t.Fatal(err)
	}
	if got := []string{set.Ranked[0].Score.NodeID, set.Ranked[1].Score.NodeID}; got[0] != "a" || got[1] != "b" {
		t.Fatalf("stable tie order got %v", got)
	}
}

func rankNode(id string, trust float64) ragnode.RAGNode {
	return ragnode.RAGNode{
		Item: citewise.Item{
			ID:     id,
			Title:  id,
			Source: "kb-" + id,
			Trust:  trust,
		},
		ChunkType:   ragnode.ChunkSection,
		TokenCount:  10,
		Version:     "v1",
		Sensitivity: ragnode.SensitivityInternal,
		ApprovedBy:  []string{"ops"},
	}
}

func rankAnalysis(nodes []ragnode.RAGNode, relevance map[string]float64) ragnode.RAGAnalysis {
	analysis := ragnode.RAGAnalysis{
		Nodes:      map[string]ragnode.RAGNode{},
		Candidates: map[string]ragnode.Candidate{},
		Centrality: map[string]float64{},
		Roles:      map[string]string{},
		EdgesIn:    map[string][]ragnode.Edge{},
		EdgesOut:   map[string][]ragnode.Edge{},
	}
	for _, node := range nodes {
		analysis.Nodes[node.ID] = node
		analysis.Candidates[node.ID] = ragnode.Candidate{NodeID: node.ID, QueryRelevance: relevance[node.ID]}
		analysis.Roles[node.ID] = citewise.RoleFoundation
	}
	return analysis
}
