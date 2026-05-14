package ragnode

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mojomast/citewiser/pkg/citewise"
)

func TestBacklogDefaultsFromSpec123(t *testing.T) {
	backlog := citewise.Backlog{
		Items: []citewise.Item{{ID: "old", Title: "Old Item", Notes: "legacy notes", Source: "kb", URL: "https://example.com", Topics: []string{"x"}}},
		Edges: []citewise.Edge{{SourceID: "old", TargetID: "old", Type: "mentions", Confidence: 0.4}},
	}
	set := CandidateSetFromBacklog(backlog)
	if len(set.Nodes) != 1 {
		t.Fatalf("nodes got %d", len(set.Nodes))
	}
	node := set.Nodes[0]
	if node.Text != "legacy notes" {
		t.Fatalf("Text got %q", node.Text)
	}
	if node.ChunkType != ChunkDocument {
		t.Fatalf("ChunkType got %q", node.ChunkType)
	}
	if node.TokenCount != EstimateTokenCount("legacy notes") {
		t.Fatalf("TokenCount got %d", node.TokenCount)
	}
	if node.Version != "" || node.Sensitivity != SensitivityInternal || node.ApprovedBy != nil || node.CommunityID != "" || node.ContextPrefix != "" {
		t.Fatalf("bad defaults: %+v", node)
	}
	if got := node.ToItem(); got.ID != "old" || got.Notes != "legacy notes" || got.Source != "kb" || got.URL != "https://example.com" {
		t.Fatalf("bad ToItem: %+v", got)
	}
	if got := set.Edges[0].ToCitewiseEdge(); got.Type != EdgeCites || got.Confidence != 0.4 {
		t.Fatalf("bad edge conversion: %+v", got)
	}
}

func TestRAGJSONRoundTripPreservesRAGFields(t *testing.T) {
	generatedAt := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	jsonText := `{
		"query_id":"q1",
		"query":"refund policy",
		"generated_at":"2026-05-14T12:00:00Z",
		"nodes":[{
			"id":"n1",
			"title":"Refund Policy",
			"notes":"old notes",
			"text":"rag text",
			"chunk_type":"permission-record",
			"token_count":33,
			"version":"v3",
			"sensitivity":"confidential",
			"approved_by":["finance"],
			"context_prefix":"policy context",
			"community_id":"c1",
			"origin":"graphrag",
			"locator":{"document_id":"doc1","section_path":"A > B"},
			"attributes":{"human_readable_id":"42"}
		}],
		"edges":[{"source_id":"n1","target_id":"n1","type":"reviewed_by","confidence":0.9}],
		"candidates":[{"node_id":"n1","query_relevance":0.8,"retrieval_mode":"hybrid"}]
	}`
	set, err := ParseCandidateSet(strings.NewReader(jsonText))
	if err != nil {
		t.Fatal(err)
	}
	if !set.GeneratedAt.Equal(generatedAt) || set.Nodes[0].ChunkType != ChunkPermissionRecord || set.Nodes[0].Sensitivity != SensitivityConfidential {
		t.Fatalf("bad parse: %+v", set)
	}
	if set.Nodes[0].ApprovedBy[0] != "finance" || set.Nodes[0].ContextPrefix != "policy context" || set.Nodes[0].Attributes["human_readable_id"] != "42" {
		t.Fatalf("RAG fields not preserved: %+v", set.Nodes[0])
	}
	encoded, err := json.Marshal(set)
	if err != nil {
		t.Fatal(err)
	}
	var roundTripped CandidateSet
	if err := json.Unmarshal(encoded, &roundTripped); err != nil {
		t.Fatal(err)
	}
	if roundTripped.Nodes[0].Text != "rag text" || roundTripped.Nodes[0].Locator.DocumentID != "doc1" || roundTripped.Candidates[0].RetrievalMode != "hybrid" {
		t.Fatalf("round trip lost fields: %+v", roundTripped)
	}
}

func TestOldBacklogParserStillHandlesUnknownRAGFields(t *testing.T) {
	oldJSONWithRAGFields := `{"items":[{"id":"old","title":"Old","notes":"legacy","text":"rag text","chunk_type":"document","sensitivity":"internal"}]}`
	backlog, err := citewise.ParseBacklog(strings.NewReader(oldJSONWithRAGFields), ".json")
	if err != nil {
		t.Fatal(err)
	}
	if len(backlog.Items) != 1 || backlog.Items[0].ID != "old" || backlog.Items[0].Notes != "legacy" {
		t.Fatalf("old parser behavior changed: %+v", backlog)
	}
}

func TestBuildAnalysisValidatesMissingCandidateNodeID(t *testing.T) {
	_, err := BuildAnalysis(CandidateSet{
		Nodes:      []RAGNode{{Item: citewise.Item{ID: "n1", Title: "Node"}, Text: "node", ChunkType: ChunkDocument}},
		Candidates: []Candidate{{NodeID: "missing", QueryRelevance: 0.5}},
	})
	if err == nil || !strings.Contains(err.Error(), `candidate node_id "missing" does not match any node`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildAnalysisDeterministicMapsAndEdges(t *testing.T) {
	set := CandidateSet{
		QueryID:     "q1",
		Query:       "alpha",
		GeneratedAt: time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC),
		Nodes: []RAGNode{
			{Item: citewise.Item{ID: "b", Title: "B", Notes: "notes b"}, Text: "text b", ChunkType: ChunkDocument},
			{Item: citewise.Item{ID: "a", Title: "A", Notes: "notes a"}, Text: "text a", ChunkType: ChunkDocument},
		},
		Edges: []Edge{
			{SourceID: "b", TargetID: "a", Type: "requires", Confidence: 0.7},
			{SourceID: "a", TargetID: "b", Type: "mentions", Confidence: 0.8},
		},
		Candidates: []Candidate{{NodeID: "b", QueryRelevance: 0.9}, {NodeID: "a", QueryRelevance: 0.5}},
	}
	analysis, err := BuildAnalysis(set)
	if err != nil {
		t.Fatal(err)
	}
	if analysis.QueryID != "q1" || analysis.Query != "alpha" || !analysis.Now.Equal(set.GeneratedAt) {
		t.Fatalf("bad analysis metadata: %+v", analysis)
	}
	if analysis.Edges[0].SourceID != "a" || analysis.Edges[0].Type != EdgeCites || analysis.Edges[1].Type != EdgePrerequisite {
		t.Fatalf("edges not normalized/sorted: %+v", analysis.Edges)
	}
	if analysis.BaseAnalysis.Backlog.Items[0].ID != "a" || analysis.BaseAnalysis.Backlog.Items[1].ID != "b" {
		t.Fatalf("base items not sorted: %+v", analysis.BaseAnalysis.Backlog.Items)
	}
	if analysis.Nodes["a"].Text != "text a" || analysis.Candidates["b"].QueryRelevance != 0.9 {
		t.Fatalf("maps not populated: %+v", analysis)
	}
	if _, ok := analysis.Roles["a"]; !ok {
		t.Fatalf("base roles not propagated: %+v", analysis.Roles)
	}
}

func TestToItemUsesTextWhenNotesMissing(t *testing.T) {
	node := RAGNode{Item: citewise.Item{ID: "n1", Title: "Node"}, Text: "body text"}
	item := node.ToItem()
	if item.Notes != "body text" {
		t.Fatalf("Notes got %q", item.Notes)
	}
}
