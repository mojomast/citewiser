package ragnode

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestRAGNodeJSONTags(t *testing.T) {
	n := RAGNode{
		Text:           "body",
		ChunkType:      ChunkProcedure,
		TokenCount:     12,
		Version:        "v1",
		Sensitivity:    SensitivityConfidential,
		EmbeddingModel: "embedder",
		ApprovedBy:     []string{"legal"},
		SupersededBy:   "n2",
		ContextPrefix:  "ctx",
		CommunityID:    "c1",
		SourceTrail:    []SourceHop{{NodeID: "n1", EdgeType: "retrieved", Confidence: 0.9}},
		Origin:         "graphrag",
		ObservedAt:     time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC),
		UpdatedAt:      time.Date(2026, 5, 14, 13, 0, 0, 0, time.UTC),
		Locator:        Locator{DocumentID: "doc", SectionPath: "A > B", TableID: "t", RowStart: 1, RowEnd: 2},
		SemanticType:   "policy",
		Attributes:     map[string]string{"k": "v"},
	}
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, key := range []string{"text", "chunk_type", "token_count", "version", "sensitivity", "embedding_model", "approved_by", "superseded_by", "context_prefix", "community_id", "source_trail", "origin", "observed_at", "updated_at", "locator", "semantic_type", "attributes", "document_id", "section_path", "table_id", "row_start", "row_end"} {
		if !strings.Contains(s, `"`+key+`"`) {
			t.Fatalf("missing JSON key %q in %s", key, s)
		}
	}
}

func TestCandidateJSONTags(t *testing.T) {
	cs := CandidateSet{
		QueryID: "q1",
		Query:   "query",
		Nodes:   []RAGNode{{Text: "node"}},
		Edges:   []Edge{{SourceID: "a", TargetID: "b", Type: "cites", Confidence: 0.8, Note: "raw", Origin: "x", Version: "v1", ObservedAt: time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)}},
		Candidates: []Candidate{{
			NodeID:         "n1",
			QueryRelevance: 0.7,
			MethodScores:   []MethodScore{{Method: "bm25", Rank: 1, Score: 2.5, Weight: 1.2}},
			RerankerScore:  0.9,
			Rank:           1,
			RetrievalMode:  "hybrid",
			SourceTrail:    []SourceHop{{NodeID: "n1", EdgeType: "retrieved", Confidence: 0.7}},
		}},
		GeneratedAt: time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC),
	}
	b, err := json.Marshal(cs)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, key := range []string{"query_id", "query", "nodes", "edges", "candidates", "generated_at", "source_id", "target_id", "type", "confidence", "note", "origin", "method_scores", "reranker_score", "rank", "retrieval_mode", "source_trail", "method", "score", "weight"} {
		if !strings.Contains(s, `"`+key+`"`) {
			t.Fatalf("missing JSON key %q in %s", key, s)
		}
	}
}

func TestSourceRefJSONTags(t *testing.T) {
	ref := SourceRef{
		NodeID:      "n1",
		Origin:      "notion",
		Source:      "kb",
		URL:         "https://example.com",
		Version:     "v1",
		ObservedAt:  time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 14, 13, 0, 0, 0, time.UTC),
		Locator:     Locator{PageStart: 3, PageEnd: 4, CharStart: 5, CharEnd: 9},
		CommunityID: "c1",
	}
	b, err := json.Marshal(ref)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, key := range []string{"node_id", "origin", "source", "url", "version", "observed_at", "updated_at", "locator", "community_id", "page_start", "page_end", "char_start", "char_end"} {
		if !strings.Contains(s, `"`+key+`"`) {
			t.Fatalf("missing JSON key %q in %s", key, s)
		}
	}
}

func TestZeroValuesAreSafe(t *testing.T) {
	var n RAGNode
	if got := n.EffectiveTokenCount(); got != 0 {
		t.Fatalf("zero node token count got %d", got)
	}
	if _, err := json.Marshal(n); err != nil {
		t.Fatal(err)
	}

	var c Candidate
	if _, err := json.Marshal(c); err != nil {
		t.Fatal(err)
	}

	var cs CandidateSet
	if _, err := json.Marshal(cs); err != nil {
		t.Fatal(err)
	}

	var a RAGAnalysis
	if a.Nodes != nil || a.Candidates != nil || a.EdgesOut != nil {
		t.Fatalf("zero analysis maps should be nil: %+v", a)
	}
}

func TestEstimateTokenCount(t *testing.T) {
	cases := map[string]int{
		"":          0,
		"a":         1,
		"abcd":      1,
		"abcde":     2,
		"abcdefgh":  2,
		"abcdefghi": 3,
	}
	for text, want := range cases {
		if got := EstimateTokenCount(text); got != want {
			t.Fatalf("EstimateTokenCount(%q) got %d want %d", text, got, want)
		}
	}
}

func TestEffectiveTokenCountPrefersExplicitCount(t *testing.T) {
	n := RAGNode{Text: "abcdefghi", TokenCount: 42}
	if got := n.EffectiveTokenCount(); got != 42 {
		t.Fatalf("explicit token count got %d", got)
	}
	if n.TokenCount != 42 {
		t.Fatalf("EffectiveTokenCount mutated TokenCount to %d", n.TokenCount)
	}

	n.TokenCount = 0
	if got := n.EffectiveTokenCount(); got != 3 {
		t.Fatalf("estimated token count got %d", got)
	}
	if n.TokenCount != 0 {
		t.Fatalf("EffectiveTokenCount mutated TokenCount to %d", n.TokenCount)
	}
}
