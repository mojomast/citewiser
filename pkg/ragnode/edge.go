package ragnode

import (
	"strings"
	"time"
)

const (
	EdgePrerequisite      = "prerequisite"
	EdgeSupersedes        = "supersedes"
	EdgeDecisionBasis     = "decision-basis"
	EdgeApprovedBy        = "approved-by"
	EdgeOverview          = "overview"
	EdgeEvidenceFor       = "evidence-for"
	EdgeEvidenceAgainst   = "evidence-against"
	EdgeImplements        = "implements"
	EdgeExceptionTo       = "exception-to"
	EdgeCites             = "cites"
	EdgeAppliesTo         = "applies-to"
	EdgeVersionOf         = "version-of"
	EdgeContradicts       = "contradicts"
	EdgeCommunityMemberOf = "community-member-of"
	EdgeSameQuestion      = "same-question"
	EdgeOwnedBy           = "owned-by"
	EdgeRelatedTo         = "related-to"
	EdgeDuplicate         = "duplicate"
)

type Edge struct {
	SourceID   string    `json:"source_id"`
	TargetID   string    `json:"target_id"`
	Type       string    `json:"type"`
	Confidence float64   `json:"confidence,omitempty"`
	Note       string    `json:"note,omitempty"`
	Origin     string    `json:"origin,omitempty"`
	Version    string    `json:"version,omitempty"`
	ObservedAt time.Time `json:"observed_at,omitempty"`
}

type EdgeNormalization struct {
	Type    string
	Known   bool
	RawType string
}

func NormalizeEdgeType(raw string) string {
	return NormalizeEdge(raw).Type
}

func NormalizeEdge(raw string) EdgeNormalization {
	canonical := normalizeEdgeToken(raw)
	if canonical == "" {
		return EdgeNormalization{Type: EdgeRelatedTo, RawType: raw}
	}
	if typ, ok := edgeAliases[canonical]; ok {
		return EdgeNormalization{Type: typ, Known: true, RawType: raw}
	}
	return EdgeNormalization{Type: EdgeRelatedTo, RawType: raw}
}

func NormalizeEdgeValue(e Edge) Edge {
	n := NormalizeEdge(e.Type)
	if n.RawType != "" && n.RawType != n.Type && e.Note == "" {
		e.Note = "raw_type=" + n.RawType
	}
	e.Type = n.Type
	return e
}

func EdgeTypeKnown(raw string) bool {
	return NormalizeEdge(raw).Known
}

func EdgeWeight(raw string) float64 {
	switch NormalizeEdgeType(raw) {
	case EdgePrerequisite:
		return 1.50
	case EdgeSupersedes:
		return 1.40
	case EdgeDecisionBasis:
		return 1.35
	case EdgeApprovedBy:
		return 1.30
	case EdgeOverview:
		return 1.25
	case EdgeEvidenceFor, EdgeEvidenceAgainst, EdgeImplements:
		return 1.20
	case EdgeExceptionTo:
		return 1.15
	case EdgeCites, EdgeAppliesTo:
		return 1.10
	case EdgeVersionOf, EdgeContradicts:
		return 1.00
	case EdgeCommunityMemberOf:
		return 0.90
	case EdgeSameQuestion:
		return 0.80
	case EdgeOwnedBy:
		return 0.70
	case EdgeDuplicate:
		return 0.20
	default:
		return 0.60
	}
}

func normalizeEdgeToken(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.Join(strings.Fields(s), "-")
	return s
}

var edgeAliases = map[string]string{
	EdgePrerequisite:      EdgePrerequisite,
	"prereq":              EdgePrerequisite,
	"requires":            EdgePrerequisite,
	EdgeSupersedes:        EdgeSupersedes,
	"updates":             EdgeSupersedes,
	"replaces":            EdgeSupersedes,
	"supersedes-version":  EdgeSupersedes,
	EdgeDecisionBasis:     EdgeDecisionBasis,
	"basis":               EdgeDecisionBasis,
	"decided-from":        EdgeDecisionBasis,
	EdgeApprovedBy:        EdgeApprovedBy,
	"approved":            EdgeApprovedBy,
	"reviewed-by":         EdgeApprovedBy,
	EdgeOverview:          EdgeOverview,
	"review":              EdgeOverview,
	"surveys":             EdgeOverview,
	"primer-for":          EdgeOverview,
	EdgeEvidenceFor:       EdgeEvidenceFor,
	"supports":            EdgeEvidenceFor,
	"proves":              EdgeEvidenceFor,
	EdgeEvidenceAgainst:   EdgeEvidenceAgainst,
	"refutes":             EdgeEvidenceAgainst,
	"opposes":             EdgeEvidenceAgainst,
	EdgeImplements:        EdgeImplements,
	"implements-policy":   EdgeImplements,
	"satisfies":           EdgeImplements,
	EdgeExceptionTo:       EdgeExceptionTo,
	"exempts":             EdgeExceptionTo,
	"waives":              EdgeExceptionTo,
	EdgeCites:             EdgeCites,
	"mentions":            EdgeCites,
	"mention":             EdgeCites,
	EdgeAppliesTo:         EdgeAppliesTo,
	"governs":             EdgeAppliesTo,
	"covers":              EdgeAppliesTo,
	EdgeVersionOf:         EdgeVersionOf,
	"revision-of":         EdgeVersionOf,
	"variant-of":          EdgeVersionOf,
	EdgeContradicts:       EdgeContradicts,
	"counterpoint":        EdgeContradicts,
	"responds-to":         EdgeContradicts,
	EdgeCommunityMemberOf: EdgeCommunityMemberOf,
	"in-community":        EdgeCommunityMemberOf,
	"cluster-member":      EdgeCommunityMemberOf,
	EdgeSameQuestion:      EdgeSameQuestion,
	"same-topic":          EdgeSameQuestion,
	"similar":             EdgeSameQuestion,
	EdgeOwnedBy:           EdgeOwnedBy,
	"owner":               EdgeOwnedBy,
	"maintained-by":       EdgeOwnedBy,
	EdgeRelatedTo:         EdgeRelatedTo,
	"relates":             EdgeRelatedTo,
	"associated-with":     EdgeRelatedTo,
	EdgeDuplicate:         EdgeDuplicate,
	"duplicates":          EdgeDuplicate,
	"duplicate-of":        EdgeDuplicate,
	"dupe":                EdgeDuplicate,
}
