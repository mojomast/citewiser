package packer

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/mojomast/citewiser/pkg/ragnode"
)

// ContextSlot is one selected context item with provenance and placement data.
type ContextSlot struct {
	Index       int                 `json:"index"`
	SlotType    SlotType            `json:"slot_type"`
	NodeID      string              `json:"node_id"`
	Role        string              `json:"role"`
	Title       string              `json:"title"`
	Text        string              `json:"text"`
	TokenCount  int                 `json:"token_count"`
	Score       float64             `json:"score"`
	Position    SlotPosition        `json:"position"`
	Source      ragnode.SourceRef   `json:"source"`
	SourceTrail []ragnode.SourceHop `json:"source_trail"`
	MustCite    bool                `json:"must_cite"`
	Rationale   string              `json:"rationale"`
}

// SuppressedEntry records a node omitted from a context plan without leaking
// unauthorized text.
type SuppressedEntry struct {
	NodeID string  `json:"node_id"`
	Reason string  `json:"reason"`
	Detail string  `json:"detail,omitempty"`
	Score  float64 `json:"score,omitempty"`
}

// ContextPlan is the packed context payload for downstream agents.
type ContextPlan struct {
	QueryID            string              `json:"query_id"`
	QueryType          QueryType           `json:"query_type"`
	Slots              []ContextSlot       `json:"slots"`
	Suppressed         []SuppressedEntry   `json:"suppressed"`
	SuppressedByReason map[string]int      `json:"suppressed_by_reason,omitempty"`
	EvidencePath       []string            `json:"evidence_path"`
	SourceTrail        []ragnode.SourceHop `json:"source_trail"`
	CritiqueSummary    string              `json:"critique_summary"`
	TokensUsed         int                 `json:"tokens_used"`
	HygieneSignal      HygieneSignal       `json:"hygiene_signal"`
	WriteBackPayload   any                 `json:"write_back_payload,omitempty"`
}

// Packer converts an analysis into an ordered context plan.
type Packer interface {
	Pack(analysis ragnode.RAGAnalysis, queryType QueryType, tokenBudget int, callerClearance string) ContextPlan
}

// PlanHash returns a stable SHA-256 hex digest of the plan's structural
// identity: QueryID, slot NodeIDs in order, suppressed node IDs in order, and
// HygieneSignal. Text content is intentionally excluded.
func (p ContextPlan) PlanHash() string {
	parts := []string{p.QueryID}
	for _, slot := range p.Slots {
		parts = append(parts, slot.NodeID)
	}
	parts = append(parts, "--suppressed--")
	for _, suppressed := range p.Suppressed {
		parts = append(parts, suppressed.NodeID)
	}
	parts = append(parts, string(p.HygieneSignal))
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}
