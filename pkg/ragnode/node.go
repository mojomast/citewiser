package ragnode

import (
	"time"

	"github.com/mojomast/citewiseussy/pkg/citewise"
)

type ChunkType string

const (
	ChunkDocument         ChunkType = "document"
	ChunkSection          ChunkType = "section"
	ChunkChunk            ChunkType = "chunk"
	ChunkEntity           ChunkType = "entity"
	ChunkClaim            ChunkType = "claim"
	ChunkCommunitySummary ChunkType = "community-summary"
	ChunkProcedure        ChunkType = "procedure"
	ChunkDefinition       ChunkType = "definition"
	ChunkDecision         ChunkType = "decision"
	ChunkPermissionRecord ChunkType = "permission-record"
)

type Sensitivity string

const (
	SensitivityPublic       Sensitivity = "public"
	SensitivityInternal     Sensitivity = "internal"
	SensitivityConfidential Sensitivity = "confidential"
	SensitivityRestricted   Sensitivity = "restricted"
)

type SourceHop struct {
	NodeID     string  `json:"node_id"`
	EdgeType   string  `json:"edge_type"`
	Confidence float64 `json:"confidence"`
}

type SourceRef struct {
	NodeID      string    `json:"node_id"`
	Origin      string    `json:"origin"`
	Source      string    `json:"source"`
	URL         string    `json:"url,omitempty"`
	Version     string    `json:"version"`
	ObservedAt  time.Time `json:"observed_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Locator     Locator   `json:"locator"`
	CommunityID string    `json:"community_id,omitempty"`
}

type Locator struct {
	DocumentID  string `json:"document_id,omitempty"`
	SectionID   string `json:"section_id,omitempty"`
	SectionPath string `json:"section_path,omitempty"`
	PageStart   int    `json:"page_start,omitempty"`
	PageEnd     int    `json:"page_end,omitempty"`
	TableID     string `json:"table_id,omitempty"`
	RowStart    int    `json:"row_start,omitempty"`
	RowEnd      int    `json:"row_end,omitempty"`
	CharStart   int    `json:"char_start,omitempty"`
	CharEnd     int    `json:"char_end,omitempty"`
}

type RAGNode struct {
	citewise.Item

	Text           string            `json:"text"`
	ChunkType      ChunkType         `json:"chunk_type"`
	TokenCount     int               `json:"token_count"`
	Version        string            `json:"version"`
	Sensitivity    Sensitivity       `json:"sensitivity"`
	EmbeddingModel string            `json:"embedding_model,omitempty"`
	ApprovedBy     []string          `json:"approved_by,omitempty"`
	SupersededBy   string            `json:"superseded_by,omitempty"`
	ContextPrefix  string            `json:"context_prefix,omitempty"`
	CommunityID    string            `json:"community_id,omitempty"`
	SourceTrail    []SourceHop       `json:"source_trail,omitempty"`
	Origin         string            `json:"origin,omitempty"`
	ObservedAt     time.Time         `json:"observed_at,omitempty"`
	UpdatedAt      time.Time         `json:"updated_at,omitempty"`
	Locator        Locator           `json:"locator,omitempty"`
	SemanticType   string            `json:"semantic_type,omitempty"`
	Attributes     map[string]string `json:"attributes,omitempty"`
}

func EstimateTokenCount(text string) int {
	if text == "" {
		return 0
	}
	return (len(text) + 3) / 4
}

func (n RAGNode) EffectiveTokenCount() int {
	if n.TokenCount > 0 {
		return n.TokenCount
	}
	return EstimateTokenCount(n.Text)
}
