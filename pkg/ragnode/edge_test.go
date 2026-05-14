package ragnode

import "testing"

func TestNormalizeEdgeTypeCanonicalAndAliases(t *testing.T) {
	cases := map[string]string{
		"prerequisite":        EdgePrerequisite,
		"prereq":              EdgePrerequisite,
		"requires":            EdgePrerequisite,
		"supersedes":          EdgeSupersedes,
		"updates":             EdgeSupersedes,
		"replaces":            EdgeSupersedes,
		"supersedes-version":  EdgeSupersedes,
		"decision-basis":      EdgeDecisionBasis,
		"basis":               EdgeDecisionBasis,
		"decided-from":        EdgeDecisionBasis,
		"approved-by":         EdgeApprovedBy,
		"approved":            EdgeApprovedBy,
		"reviewed-by":         EdgeApprovedBy,
		"overview":            EdgeOverview,
		"review":              EdgeOverview,
		"surveys":             EdgeOverview,
		"primer-for":          EdgeOverview,
		"evidence-for":        EdgeEvidenceFor,
		"supports":            EdgeEvidenceFor,
		"proves":              EdgeEvidenceFor,
		"evidence-against":    EdgeEvidenceAgainst,
		"refutes":             EdgeEvidenceAgainst,
		"opposes":             EdgeEvidenceAgainst,
		"implements":          EdgeImplements,
		"implements-policy":   EdgeImplements,
		"satisfies":           EdgeImplements,
		"exception-to":        EdgeExceptionTo,
		"exempts":             EdgeExceptionTo,
		"waives":              EdgeExceptionTo,
		"cites":               EdgeCites,
		"mentions":            EdgeCites,
		"mention":             EdgeCites,
		"applies-to":          EdgeAppliesTo,
		"governs":             EdgeAppliesTo,
		"covers":              EdgeAppliesTo,
		"version-of":          EdgeVersionOf,
		"revision-of":         EdgeVersionOf,
		"variant-of":          EdgeVersionOf,
		"contradicts":         EdgeContradicts,
		"counterpoint":        EdgeContradicts,
		"responds-to":         EdgeContradicts,
		"community-member-of": EdgeCommunityMemberOf,
		"in-community":        EdgeCommunityMemberOf,
		"cluster-member":      EdgeCommunityMemberOf,
		"same-question":       EdgeSameQuestion,
		"same-topic":          EdgeSameQuestion,
		"similar":             EdgeSameQuestion,
		"owned-by":            EdgeOwnedBy,
		"owner":               EdgeOwnedBy,
		"maintained-by":       EdgeOwnedBy,
		"related-to":          EdgeRelatedTo,
		"relates":             EdgeRelatedTo,
		"associated-with":     EdgeRelatedTo,
		"duplicate":           EdgeDuplicate,
		"duplicates":          EdgeDuplicate,
		"duplicate-of":        EdgeDuplicate,
		"dupe":                EdgeDuplicate,
	}
	for raw, want := range cases {
		if got := NormalizeEdgeType(raw); got != want {
			t.Fatalf("NormalizeEdgeType(%q) got %q want %q", raw, got, want)
		}
		if !EdgeTypeKnown(raw) {
			t.Fatalf("EdgeTypeKnown(%q) got false", raw)
		}
	}
}

func TestNormalizeEdgeTypeKebabCase(t *testing.T) {
	cases := map[string]string{
		" Decision Basis ": EdgeDecisionBasis,
		"decision_basis":   EdgeDecisionBasis,
		"SAME TOPIC":       EdgeSameQuestion,
		"reviewed_by":      EdgeApprovedBy,
	}
	for raw, want := range cases {
		if got := NormalizeEdgeType(raw); got != want {
			t.Fatalf("NormalizeEdgeType(%q) got %q want %q", raw, got, want)
		}
	}
}

func TestUnknownEdgeTypeFallsBackToRelatedTo(t *testing.T) {
	n := NormalizeEdge("totally-custom")
	if n.Type != EdgeRelatedTo || n.Known || n.RawType != "totally-custom" {
		t.Fatalf("unexpected normalization: %+v", n)
	}
	if got := EdgeWeight("totally-custom"); got != 0.60 {
		t.Fatalf("unknown weight got %.2f", got)
	}
}

func TestNormalizeEdgeValuePreservesRawTypeInNote(t *testing.T) {
	e := NormalizeEdgeValue(Edge{SourceID: "a", TargetID: "b", Type: "decided_from"})
	if e.Type != EdgeDecisionBasis {
		t.Fatalf("type got %q", e.Type)
	}
	if e.Note != "raw_type=decided_from" {
		t.Fatalf("note got %q", e.Note)
	}

	e = NormalizeEdgeValue(Edge{SourceID: "a", TargetID: "b", Type: "custom", Note: "keep"})
	if e.Type != EdgeRelatedTo || e.Note != "keep" {
		t.Fatalf("existing note not preserved: %+v", e)
	}
}

func TestEdgeWeights(t *testing.T) {
	cases := map[string]float64{
		EdgePrerequisite:      1.50,
		EdgeSupersedes:        1.40,
		EdgeDecisionBasis:     1.35,
		EdgeApprovedBy:        1.30,
		EdgeOverview:          1.25,
		EdgeEvidenceFor:       1.20,
		EdgeEvidenceAgainst:   1.20,
		EdgeImplements:        1.20,
		EdgeExceptionTo:       1.15,
		EdgeCites:             1.10,
		EdgeAppliesTo:         1.10,
		EdgeVersionOf:         1.00,
		EdgeContradicts:       1.00,
		EdgeCommunityMemberOf: 0.90,
		EdgeSameQuestion:      0.80,
		EdgeOwnedBy:           0.70,
		EdgeRelatedTo:         0.60,
		EdgeDuplicate:         0.20,
	}
	for raw, want := range cases {
		if got := EdgeWeight(raw); got != want {
			t.Fatalf("EdgeWeight(%q) got %.2f want %.2f", raw, got, want)
		}
	}
}
