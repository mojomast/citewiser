package hygiene

import (
	"testing"

	"github.com/mojomast/citewiser/pkg/citewise"
	"github.com/mojomast/citewiser/pkg/packer"
	"github.com/mojomast/citewiser/pkg/ragnode"
)

func TestSuggestMissingEdgesHeuristics(t *testing.T) {
	analysis := hygieneAnalysis([]ragnode.RAGNode{
		hygieneNode("policy", "Refund Policy", []string{"refund", "policy"}, "c1", ragnode.ChunkSection),
		hygieneNode("exception", "Refund Exception", []string{"refund", "policy"}, "c1", ragnode.ChunkSection),
		hygieneNode("summary", "Refund Summary", []string{"refund"}, "c1", ragnode.ChunkCommunitySummary),
		hygieneNode("dupe-a", "Vendor Access Policy", nil, "c2", ragnode.ChunkSection),
		hygieneNode("dupe-b", "Vendor Access Policy", nil, "c3", ragnode.ChunkSection),
		hygieneNode("old", "Old Runbook", nil, "c4", ragnode.ChunkProcedure),
		hygieneNode("new", "New Runbook", nil, "c4", ragnode.ChunkProcedure),
	})
	old := analysis.Nodes["old"]
	old.SupersededBy = "new"
	analysis.Nodes["old"] = old

	suggestions := NewAnalyzer().SuggestMissingEdges(analysis)
	assertSuggestion(t, suggestions, "exception", "policy", ragnode.EdgeSameQuestion, 0.70)
	assertSuggestion(t, suggestions, "exception", "policy", ragnode.EdgeExceptionTo, 0.55)
	assertSuggestion(t, suggestions, "policy", "summary", ragnode.EdgeCommunityMemberOf, 0.80)
	assertSuggestion(t, suggestions, "dupe-a", "dupe-b", ragnode.EdgeDuplicate, 0.90)
	assertSuggestion(t, suggestions, "new", "old", ragnode.EdgeSupersedes, 1.00)
}

func TestHygieneReportAndScore(t *testing.T) {
	nodes := []ragnode.RAGNode{
		hygieneNode("orphan", "Orphan", nil, "", ragnode.ChunkSection),
		hygieneNode("stale", "Stale", nil, "", ragnode.ChunkSection),
		hygieneNode("low", "Low", nil, "", ragnode.ChunkSection),
		hygieneNode("perm", "Permission", nil, "", ragnode.ChunkPermissionRecord),
		hygieneNode("needs-bridge", "Needs Bridge", nil, "", ragnode.ChunkSection),
	}
	for i := range nodes {
		nodes[i].SupersededBy = "new"
		nodes[i].Trust = 0.1
	}
	nodes[3].ApprovedBy = nil
	analysis := hygieneAnalysis(nodes)
	analysis.Duplicates = [][]string{{"low", "stale", "orphan", "perm", "needs-bridge"}}
	analysis.Edges = []ragnode.Edge{{SourceID: "perm", TargetID: "needs-bridge", Type: ragnode.EdgePrerequisite}}
	analysis.EdgesOut["perm"] = analysis.Edges
	analysis.EdgesIn["needs-bridge"] = analysis.Edges

	report := NewAnalyzer().Analyze(analysis, false)
	if report.Score >= 0.55 || report.Signal != packer.HygieneRed {
		t.Fatalf("expected red low hygiene, got score %.3f signal %s", report.Score, report.Signal)
	}
	if len(report.RetrievalTargets) == 0 {
		t.Fatalf("red report should include retrieval targets: %+v", report)
	}
	if len(report.OrphanNodes) == 0 || len(report.StaleNodes) == 0 || len(report.MissingBridges) == 0 {
		t.Fatalf("expected orphan/stale/missing bridge accounting: %+v", report)
	}
}

func TestCorrectiveSignalThresholds(t *testing.T) {
	a := NewAnalyzer()
	green := hygieneAnalysis([]ragnode.RAGNode{hygieneNode("a", "A", nil, "", ragnode.ChunkSection)})
	green.Edges = []ragnode.Edge{{SourceID: "a", TargetID: "a", Type: ragnode.EdgeRelatedTo}}
	green.EdgesOut["a"] = green.Edges
	green.EdgesIn["a"] = green.Edges
	if got := a.CorrectiveSignal(green, DefaultThreshold); got != packer.HygieneGreen {
		t.Fatalf("green signal got %s", got)
	}

	yellow := hygieneAnalysis([]ragnode.RAGNode{hygieneNode("a", "A", nil, "", ragnode.ChunkSection), hygieneNode("b", "B", nil, "", ragnode.ChunkSection)})
	yellow.Duplicates = [][]string{{"a", "b"}}
	if got := a.CorrectiveSignal(yellow, DefaultThreshold); got != packer.HygieneYellow {
		t.Fatalf("yellow signal got %s score %.3f", got, a.HygieneScore(yellow))
	}

	red := hygieneAnalysis([]ragnode.RAGNode{hygieneNode("a", "A", nil, "", ragnode.ChunkPermissionRecord)})
	red.Nodes["a"] = ragnode.RAGNode{Item: citewise.Item{ID: "a", Title: "A", Trust: 0.1}, ChunkType: ragnode.ChunkPermissionRecord}
	if got := a.CorrectiveSignal(red, 0.90); got != packer.HygieneRed {
		t.Fatalf("red signal got %s score %.3f", got, a.HygieneScore(red))
	}
}

func hygieneNode(id, title string, topics []string, community string, chunkType ragnode.ChunkType) ragnode.RAGNode {
	return ragnode.RAGNode{Item: citewise.Item{ID: id, Title: title, Topics: topics, Trust: 1}, CommunityID: community, ChunkType: chunkType, ApprovedBy: []string{"ops"}}
}

func hygieneAnalysis(nodes []ragnode.RAGNode) ragnode.RAGAnalysis {
	analysis := ragnode.RAGAnalysis{Nodes: map[string]ragnode.RAGNode{}, EdgesIn: map[string][]ragnode.Edge{}, EdgesOut: map[string][]ragnode.Edge{}, Roles: map[string]string{}}
	for _, node := range nodes {
		analysis.Nodes[node.ID] = node
	}
	return analysis
}

func assertSuggestion(t *testing.T, suggestions []EdgeSuggestion, source, target, typ string, confidence float64) {
	t.Helper()
	for _, suggestion := range suggestions {
		if suggestion.SourceID == source && suggestion.TargetID == target && suggestion.Type == typ && suggestion.Confidence == confidence {
			return
		}
	}
	t.Fatalf("missing suggestion %s -> %s %s %.2f in %+v", source, target, typ, confidence, suggestions)
}
