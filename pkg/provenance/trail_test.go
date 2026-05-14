package provenance

import (
	"testing"
	"time"

	"github.com/mojomast/citewiser/pkg/access"
	"github.com/mojomast/citewiser/pkg/citewise"
	"github.com/mojomast/citewiser/pkg/ragnode"
)

func TestBuildSourceRefFromNodeMetadata(t *testing.T) {
	observed := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 5, 14, 13, 0, 0, 0, time.UTC)
	node := ragnode.RAGNode{
		Item:        citewise.Item{ID: "n1", Source: "KB", URL: "https://example.com"},
		Origin:      "graphrag",
		Version:     "v2",
		ObservedAt:  observed,
		UpdatedAt:   updated,
		Locator:     ragnode.Locator{DocumentID: "doc", SectionPath: "A > B"},
		CommunityID: "c1",
	}
	ref := BuildSourceRef(node)
	if ref.NodeID != "n1" || ref.Origin != "graphrag" || ref.Source != "KB" || ref.URL != "https://example.com" || ref.Version != "v2" || ref.CommunityID != "c1" {
		t.Fatalf("bad ref: %+v", ref)
	}
	if !ref.ObservedAt.Equal(observed) || !ref.UpdatedAt.Equal(updated) || ref.Locator.DocumentID != "doc" {
		t.Fatalf("bad ref metadata: %+v", ref)
	}
}

func TestBuildSourceRefPreservesTableLocatorWhenSupplied(t *testing.T) {
	node := ragnode.RAGNode{
		Item: citewise.Item{ID: "metric", Source: "warehouse"},
		Locator: ragnode.Locator{
			DocumentID: "doc-metrics",
			TableID:    "table-revenue",
			RowStart:   7,
			RowEnd:     9,
		},
	}
	ref := BuildSourceRef(node)
	if ref.Locator.TableID != "table-revenue" || ref.Locator.RowStart != 7 || ref.Locator.RowEnd != 9 {
		t.Fatalf("table locator not preserved: %+v", ref.Locator)
	}
}

func TestBuildSourceTrailDirectRetrievedAndRequiredSlot(t *testing.T) {
	node := ragnode.RAGNode{Item: citewise.Item{ID: "n1"}, ChunkType: ragnode.ChunkDocument}
	candidate := ragnode.Candidate{NodeID: "n1", QueryRelevance: 0.87}
	trail := BuildSourceTrail(ragnode.RAGAnalysis{}, node, TrailOptions{Candidate: &candidate, RequiredSlot: true, ScoreTotal: 82})
	if len(trail) != 2 {
		t.Fatalf("trail got %+v", trail)
	}
	if trail[0] != (ragnode.SourceHop{NodeID: "n1", EdgeType: EdgeRetrieved, Confidence: 0.87}) {
		t.Fatalf("direct hop got %+v", trail[0])
	}
	if trail[1] != (ragnode.SourceHop{NodeID: "n1", EdgeType: EdgeSlotRequired, Confidence: 0.82}) {
		t.Fatalf("slot hop got %+v", trail[1])
	}
}

func TestBuildSourceTrailCommunitySummaryPath(t *testing.T) {
	analysis := mustAnalysis(t, ragnode.CandidateSet{
		Nodes: []ragnode.RAGNode{
			{Item: citewise.Item{ID: "member", Title: "Member"}, ChunkType: ragnode.ChunkEntity},
			{Item: citewise.Item{ID: "summary", Title: "Summary"}, ChunkType: ragnode.ChunkCommunitySummary},
		},
		Edges: []ragnode.Edge{{SourceID: "member", TargetID: "summary", Type: ragnode.EdgeCommunityMemberOf, Confidence: 0.8}},
	})
	trail := BuildSourceTrail(analysis, analysis.Nodes["summary"], TrailOptions{})
	if len(trail) != 1 || trail[0] != (ragnode.SourceHop{NodeID: "member", EdgeType: ragnode.EdgeCommunityMemberOf, Confidence: 0.8}) {
		t.Fatalf("bad community trail: %+v", trail)
	}
}

func TestBuildSourceTrailDecisionBasisPath(t *testing.T) {
	analysis := mustAnalysis(t, ragnode.CandidateSet{
		Nodes: []ragnode.RAGNode{
			{Item: citewise.Item{ID: "decision", Title: "Decision"}, ChunkType: ragnode.ChunkDecision, ApprovedBy: []string{"legal"}},
			{Item: citewise.Item{ID: "policy", Title: "Policy"}, ChunkType: ragnode.ChunkDocument},
		},
		Edges: []ragnode.Edge{{SourceID: "decision", TargetID: "policy", Type: ragnode.EdgeDecisionBasis, Confidence: 0.91}},
	})
	trail := BuildSourceTrail(analysis, analysis.Nodes["decision"], TrailOptions{})
	if len(trail) != 1 || trail[0] != (ragnode.SourceHop{NodeID: "policy", EdgeType: ragnode.EdgeDecisionBasis, Confidence: 0.91}) {
		t.Fatalf("bad decision trail: %+v", trail)
	}
}

func TestBuildSourceTrailCandidateEdgePath(t *testing.T) {
	analysis := mustAnalysis(t, ragnode.CandidateSet{
		Nodes: []ragnode.RAGNode{
			{Item: citewise.Item{ID: "candidate", Title: "Candidate"}, ChunkType: ragnode.ChunkDocument},
			{Item: citewise.Item{ID: "bridge", Title: "Bridge"}, ChunkType: ragnode.ChunkDocument},
		},
		Edges: []ragnode.Edge{{SourceID: "candidate", TargetID: "bridge", Type: ragnode.EdgeEvidenceFor, Confidence: 0.73}},
	})
	candidate := ragnode.Candidate{NodeID: "candidate", QueryRelevance: 0.8}
	trail := BuildSourceTrail(analysis, analysis.Nodes["bridge"], TrailOptions{Candidate: &candidate})
	if len(trail) != 1 || trail[0] != (ragnode.SourceHop{NodeID: "candidate", EdgeType: ragnode.EdgeEvidenceFor, Confidence: 0.73}) {
		t.Fatalf("bad candidate edge trail: %+v", trail)
	}
}

func TestRedactSourceTrailDropsUnauthorizedHops(t *testing.T) {
	analysis := mustAnalysis(t, ragnode.CandidateSet{Nodes: []ragnode.RAGNode{
		{Item: citewise.Item{ID: "allowed", Title: "Allowed"}, Sensitivity: ragnode.SensitivityInternal},
		{Item: citewise.Item{ID: "secret", Title: "Secret"}, Sensitivity: ragnode.SensitivityRestricted},
	}})
	trail := []ragnode.SourceHop{
		{NodeID: "allowed", EdgeType: EdgeRetrieved, Confidence: 1},
		{NodeID: "secret", EdgeType: ragnode.EdgeDecisionBasis, Confidence: 1},
		{NodeID: "unknown", EdgeType: ragnode.EdgeRelatedTo, Confidence: 1},
	}
	redacted := RedactSourceTrail(access.Context{Clearance: access.ClearanceInternal}, access.NewController(), analysis, trail)
	if len(redacted) != 1 || redacted[0].NodeID != "allowed" {
		t.Fatalf("bad redacted trail: %+v", redacted)
	}
}

func TestRedactSourceRefReturnsOnlyNodeIDWhenDenied(t *testing.T) {
	node := ragnode.RAGNode{Item: citewise.Item{ID: "secret", Source: "KB", URL: "https://example.com"}, Sensitivity: ragnode.SensitivityRestricted, Origin: "graphrag", Version: "v1"}
	ref := RedactSourceRef(access.Context{Clearance: access.ClearanceInternal}, access.NewController(), node)
	if ref.NodeID != "secret" || ref.Source != "" || ref.URL != "" || ref.Origin != "" || ref.Version != "" {
		t.Fatalf("source ref leaked data: %+v", ref)
	}
}

func mustAnalysis(t *testing.T, set ragnode.CandidateSet) ragnode.RAGAnalysis {
	t.Helper()
	analysis, err := ragnode.BuildAnalysis(set)
	if err != nil {
		t.Fatal(err)
	}
	return analysis
}
