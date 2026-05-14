package ranker

import (
	"math"
	"reflect"
	"testing"

	"github.com/mojomast/citewiseussy/pkg/citewise"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

func TestPersonalizedPageRankFallbackForSmallGraph(t *testing.T) {
	analysis := pprAnalysis([]ragnode.Edge{{SourceID: "a", TargetID: "b", Type: ragnode.EdgeCites}}, map[string]float64{"a": 1})
	got := PersonalizedPageRank(analysis)
	for id, score := range got {
		if score != 0 {
			t.Fatalf("%s score got %.6f want 0", id, score)
		}
	}
}

func TestPersonalizedPageRankConvergesAndUsesSeeds(t *testing.T) {
	analysis := pprAnalysis([]ragnode.Edge{
		{SourceID: "a", TargetID: "b", Type: ragnode.EdgeCites, Confidence: 1},
		{SourceID: "b", TargetID: "c", Type: ragnode.EdgeCites, Confidence: 1},
		{SourceID: "c", TargetID: "a", Type: ragnode.EdgeCites, Confidence: 1},
	}, map[string]float64{"a": 1})
	got := PersonalizedPageRank(analysis)
	if !close(sumScores(got), 1) {
		t.Fatalf("scores should sum to 1, got %.12f: %+v", sumScores(got), got)
	}
	if got["a"] <= got["b"] || got["b"] <= got["c"] {
		t.Fatalf("seeded cycle scores should decay from seed, got %+v", got)
	}
}

func TestPersonalizedPageRankDisconnectedAndDanglingNodes(t *testing.T) {
	analysis := pprAnalysis([]ragnode.Edge{
		{SourceID: "a", TargetID: "b", Type: ragnode.EdgeCites},
		{SourceID: "b", TargetID: "a", Type: ragnode.EdgeCites},
	}, map[string]float64{"a": 1, "c": 1})
	got := PersonalizedPageRank(analysis)
	if got["c"] == 0 {
		t.Fatalf("dangling seeded node should retain restart mass: %+v", got)
	}
	if !close(sumScores(got), 1) {
		t.Fatalf("scores should sum to 1, got %.12f", sumScores(got))
	}
}

func TestPersonalizedPageRankDeterministicOutput(t *testing.T) {
	analysis := pprAnalysis([]ragnode.Edge{
		{SourceID: "b", TargetID: "c", Type: ragnode.EdgePrerequisite},
		{SourceID: "a", TargetID: "b", Type: ragnode.EdgeCites},
		{SourceID: "c", TargetID: "a", Type: ragnode.EdgeRelatedTo},
	}, map[string]float64{"b": 0.7, "a": 0.3})
	first := PersonalizedPageRank(analysis)
	second := PersonalizedPageRank(analysis)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("PPR output changed between runs: first %+v second %+v", first, second)
	}
}

func pprAnalysis(edges []ragnode.Edge, seeds map[string]float64) ragnode.RAGAnalysis {
	nodes := map[string]ragnode.RAGNode{}
	for _, id := range []string{"a", "b", "c"} {
		nodes[id] = ragnode.RAGNode{Item: citewise.Item{ID: id, Title: id}}
	}
	candidates := map[string]ragnode.Candidate{}
	for id, relevance := range seeds {
		candidates[id] = ragnode.Candidate{NodeID: id, QueryRelevance: relevance}
	}
	edgesOut := map[string][]ragnode.Edge{}
	for _, edge := range edges {
		edgesOut[edge.SourceID] = append(edgesOut[edge.SourceID], edge)
	}
	return ragnode.RAGAnalysis{Nodes: nodes, Edges: edges, EdgesOut: edgesOut, Candidates: candidates}
}

func sumScores(scores map[string]float64) float64 {
	total := 0.0
	for _, score := range scores {
		total += score
	}
	return math.Round(total*1e12) / 1e12
}
