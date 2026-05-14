package ranker

import (
	"math"
	"testing"

	"github.com/mojomast/citewiseussy/pkg/citewise"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

func TestAuthorityFormulaAndClamp(t *testing.T) {
	node := ragnode.RAGNode{
		Item:       citewise.Item{ID: "target", Trust: 2},
		ChunkType:  ragnode.ChunkPermissionRecord,
		ApprovedBy: []string{"legal"},
		Version:    "v3",
	}
	analysis := ragnode.RAGAnalysis{Edges: []ragnode.Edge{
		{SourceID: "a", TargetID: "target", Type: ragnode.EdgeCites},
		{SourceID: "b", TargetID: "target", Type: ragnode.EdgeCites},
		{SourceID: "c", TargetID: "other", Type: ragnode.EdgeCites},
	}}

	if got := Authority(node, analysis); got != 1 {
		t.Fatalf("Authority got %.6f want 1", got)
	}
}

func TestAuthorityChunkPriorsAndVersionCurrentness(t *testing.T) {
	cases := []struct {
		chunkType ragnode.ChunkType
		want      float64
	}{
		{ragnode.ChunkPermissionRecord, 1.00},
		{ragnode.ChunkDecision, 0.95},
		{ragnode.ChunkProcedure, 0.90},
		{ragnode.ChunkDefinition, 0.85},
		{ragnode.ChunkDocument, 0.80},
		{ragnode.ChunkSection, 0.75},
		{ragnode.ChunkCommunitySummary, 0.70},
		{ragnode.ChunkClaim, 0.65},
		{ragnode.ChunkChunk, 0.60},
		{ragnode.ChunkEntity, 0.55},
	}
	for _, tc := range cases {
		if got := ChunkTypeAuthority(tc.chunkType); got != tc.want {
			t.Fatalf("ChunkTypeAuthority(%q) got %.2f want %.2f", tc.chunkType, got, tc.want)
		}
	}

	base := ragnode.RAGNode{Item: citewise.Item{ID: "n", Trust: 0.5}, ChunkType: ragnode.ChunkSection}
	if got, want := Authority(base, ragnode.RAGAnalysis{}), 0.3375; !close(got, want) {
		t.Fatalf("missing version authority got %.6f want %.6f", got, want)
	}
	base.Version = "v1"
	if got, want := Authority(base, ragnode.RAGAnalysis{}), 0.3875; !close(got, want) {
		t.Fatalf("current version authority got %.6f want %.6f", got, want)
	}
	base.SupersededBy = "n2"
	if got, want := Authority(base, ragnode.RAGAnalysis{}), 0.2875; !close(got, want) {
		t.Fatalf("superseded authority got %.6f want %.6f", got, want)
	}
}

func TestTokenBudgetFitThresholdsAndEstimateFallback(t *testing.T) {
	cases := []struct {
		name       string
		node       ragnode.RAGNode
		budget     int
		wantTokens int
		want       float64
	}{
		{"eight percent", authorityNode("n1", 8), 100, 8, 1.00},
		{"estimated nine percent", authorityTextNode("n2", "abcdefghijklmnopqrstuvwxyzabcdefghi"), 100, 9, 0.88},
		{"sixteen percent", authorityNode("n3", 16), 100, 16, 0.73},
		{"twenty six percent", authorityNode("n4", 26), 100, 26, 0.55},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			analysis := ragnode.RAGAnalysis{Nodes: map[string]ragnode.RAGNode{tc.node.ID: tc.node}}
			if got := tc.node.EffectiveTokenCount(); got != tc.wantTokens {
				t.Fatalf("EffectiveTokenCount got %d want %d", got, tc.wantTokens)
			}
			if got := TokenBudgetFit(tc.node, analysis, tc.budget); !close(got, tc.want) {
				t.Fatalf("TokenBudgetFit got %.6f want %.6f", got, tc.want)
			}
		})
	}
}

func TestTokenBudgetFitNormalizesDensityAcrossAnalysis(t *testing.T) {
	short := authorityNode("short", 8)
	long := authorityNode("long", 98)
	long.Trust = 0.2
	analysis := ragnode.RAGAnalysis{Nodes: map[string]ragnode.RAGNode{short.ID: short, long.ID: long}}

	shortFit := TokenBudgetFit(short, analysis, 100)
	longFit := TokenBudgetFit(long, analysis, 100)
	if shortFit <= longFit {
		t.Fatalf("short authoritative node should fit better: short %.6f long %.6f", shortFit, longFit)
	}
	if longFit <= 0 || longFit >= 1 {
		t.Fatalf("long fit should remain bounded, got %.6f", longFit)
	}
}

func TestDiversitySourceAndCommunityPenalties(t *testing.T) {
	selected := []ragnode.RAGNode{{Item: citewise.Item{Source: "kb"}, CommunityID: "c1"}}
	cases := []struct {
		name string
		node ragnode.RAGNode
		want float64
	}{
		{"new source and community", ragnode.RAGNode{Item: citewise.Item{Source: "runbook"}, CommunityID: "c2"}, 1.00},
		{"same source", ragnode.RAGNode{Item: citewise.Item{Source: "kb"}, CommunityID: "c2"}, 0.65},
		{"same community", ragnode.RAGNode{Item: citewise.Item{Source: "runbook"}, CommunityID: "c1"}, 0.75},
		{"same source and community", ragnode.RAGNode{Item: citewise.Item{Source: "kb"}, CommunityID: "c1"}, 0.40},
	}
	for _, tc := range cases {
		if got := Diversity(tc.node, selected); !close(got, tc.want) {
			t.Fatalf("%s got %.6f want %.6f", tc.name, got, tc.want)
		}
	}
}

func authorityNode(id string, tokens int) ragnode.RAGNode {
	return ragnode.RAGNode{
		Item:       citewise.Item{ID: id, Trust: 1},
		ChunkType:  ragnode.ChunkPermissionRecord,
		TokenCount: tokens,
		ApprovedBy: []string{"legal"},
		Version:    "v1",
	}
}

func authorityTextNode(id, text string) ragnode.RAGNode {
	node := authorityNode(id, 0)
	node.Text = text
	return node
}

func close(got, want float64) bool {
	return math.Abs(got-want) < 0.000001
}
