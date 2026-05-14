package lightrag

import (
	"encoding/json"
	"io"

	"github.com/mojomast/citewiseussy/pkg/citewise"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

const (
	ModeLocal  = "lightrag-local"
	ModeGlobal = "lightrag-global"
)

type Handoff struct {
	QueryID       string         `json:"query_id,omitempty"`
	Query         string         `json:"query,omitempty"`
	LocalResults  []Result       `json:"local_results,omitempty"`
	GlobalResults []Result       `json:"global_results,omitempty"`
	Relationships []Relationship `json:"relationships,omitempty"`
}

type Result struct {
	NodeID      string            `json:"node_id"`
	Title       string            `json:"title,omitempty"`
	Text        string            `json:"text,omitempty"`
	Score       float64           `json:"score,omitempty"`
	Rank        int               `json:"rank,omitempty"`
	CommunityID string            `json:"community_id,omitempty"`
	ChunkType   ragnode.ChunkType `json:"chunk_type,omitempty"`
}

type Relationship struct {
	SourceID   string  `json:"source_id"`
	TargetID   string  `json:"target_id"`
	Type       string  `json:"type,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

func ParseJSON(r io.Reader) (ragnode.CandidateSet, error) {
	var handoff Handoff
	if err := json.NewDecoder(r).Decode(&handoff); err != nil {
		return ragnode.CandidateSet{}, err
	}
	return MapHandoff(handoff), nil
}

func MapHandoff(h Handoff) ragnode.CandidateSet {
	set := ragnode.CandidateSet{QueryID: h.QueryID, Query: h.Query}
	for _, result := range h.LocalResults {
		set.Nodes = append(set.Nodes, resultNode(result, ragnode.ChunkEntity))
		set.Candidates = append(set.Candidates, candidate(result, ModeLocal))
	}
	for _, result := range h.GlobalResults {
		set.Nodes = append(set.Nodes, resultNode(result, ragnode.ChunkCommunitySummary))
		set.Candidates = append(set.Candidates, candidate(result, ModeGlobal))
	}
	for _, rel := range h.Relationships {
		set.Edges = append(set.Edges, ragnode.Edge{SourceID: rel.SourceID, TargetID: rel.TargetID, Type: ragnode.NormalizeEdgeType(rel.Type), Confidence: rel.Confidence})
	}
	return set
}

func resultNode(result Result, fallback ragnode.ChunkType) ragnode.RAGNode {
	chunkType := result.ChunkType
	if chunkType == "" {
		chunkType = fallback
	}
	return ragnode.RAGNode{Item: citewise.Item{ID: result.NodeID, Title: result.Title, Notes: result.Text, Source: "lightrag", Trust: 0.7}, Text: result.Text, ChunkType: chunkType, Sensitivity: ragnode.SensitivityInternal, Origin: "lightrag", CommunityID: result.CommunityID}
}

func candidate(result Result, mode string) ragnode.Candidate {
	return ragnode.Candidate{NodeID: result.NodeID, QueryRelevance: clamp01(result.Score), Rank: result.Rank, RetrievalMode: mode, MethodScores: []ragnode.MethodScore{{Method: mode, Rank: result.Rank, Score: result.Score}}}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
