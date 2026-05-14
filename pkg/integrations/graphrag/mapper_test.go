package graphrag

import (
	"os"
	"testing"
	"time"

	"github.com/mojomast/citewiser/pkg/ragnode"
)

func TestParseJSONFixture(t *testing.T) {
	f, err := os.Open("../../../testdata/graphrag_minimal/export.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	set, err := ParseJSON(f)
	if err != nil {
		t.Fatal(err)
	}
	analysis, err := ragnode.BuildAnalysis(set)
	if err != nil {
		t.Fatal(err)
	}
	report := analysis.Nodes["report-c1"]
	if report.ChunkType != ragnode.ChunkCommunitySummary || report.Text != "Vendor access overview" || report.CommunityID != "c1" {
		t.Fatalf("bad community report node: %+v", report)
	}
	if report.Type != "overview" {
		t.Fatalf("community report should be overview-compatible, got type %q", report.Type)
	}
	if report.Attributes["human_readable_id"] != "CR1" || report.Attributes["community_report_rank"] != "0.91" {
		t.Fatalf("missing report attrs: %+v", report.Attributes)
	}
	if !report.ObservedAt.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("period not parsed: %s", report.ObservedAt)
	}
	if got := analysis.Nodes["tu-1"].TokenCount; got != 12 {
		t.Fatalf("text unit token count got %d", got)
	}
	assertEdge(t, analysis.Edges, "entity-1", "tu-1", ragnode.EdgeRelatedTo)
	assertEdge(t, analysis.Edges, "entity-1", "report-c1", ragnode.EdgeCommunityMemberOf)
	assertEdge(t, analysis.Edges, "tu-1", "report-c1", ragnode.EdgeCommunityMemberOf)
	assertEdge(t, analysis.Edges, "tu-1", "claim-1", ragnode.EdgeEvidenceFor)
}

func assertEdge(t *testing.T, edges []ragnode.Edge, source, target, typ string) {
	t.Helper()
	for _, edge := range edges {
		if edge.SourceID == source && edge.TargetID == target && edge.Type == typ {
			return
		}
	}
	t.Fatalf("missing edge %s -> %s %s in %+v", source, target, typ, edges)
}
