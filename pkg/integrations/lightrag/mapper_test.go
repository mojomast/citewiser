package lightrag

import (
	"testing"

	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

func TestMapHandoffPreservesModesAndRelationships(t *testing.T) {
	set := MapHandoff(Handoff{QueryID: "q", Query: "query", LocalResults: []Result{{NodeID: "local", Title: "Local", Score: 0.8, Rank: 2}}, GlobalResults: []Result{{NodeID: "global", Title: "Global", Score: 0.9, Rank: 1}}, Relationships: []Relationship{{SourceID: "local", TargetID: "global", Type: "in-community", Confidence: 0.7}}})
	if len(set.Nodes) != 2 || len(set.Candidates) != 2 || len(set.Edges) != 1 {
		t.Fatalf("bad set: %+v", set)
	}
	if set.Nodes[0].ChunkType != ragnode.ChunkEntity || set.Candidates[0].RetrievalMode != ModeLocal {
		t.Fatalf("bad local mapping: node=%+v candidate=%+v", set.Nodes[0], set.Candidates[0])
	}
	if set.Nodes[1].ChunkType != ragnode.ChunkCommunitySummary || set.Candidates[1].RetrievalMode != ModeGlobal {
		t.Fatalf("bad global mapping: node=%+v candidate=%+v", set.Nodes[1], set.Candidates[1])
	}
	if set.Edges[0].Type != ragnode.EdgeCommunityMemberOf {
		t.Fatalf("edge type got %q", set.Edges[0].Type)
	}
}
