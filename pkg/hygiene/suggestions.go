package hygiene

import (
	"sort"
	"strings"

	"github.com/mojomast/citewiser/pkg/ragnode"
)

func (DefaultAnalyzer) SuggestMissingEdges(analysis ragnode.RAGAnalysis) []EdgeSuggestion {
	ids := make([]string, 0, len(analysis.Nodes))
	for id := range analysis.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var suggestions []EdgeSuggestion
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			a := analysis.Nodes[ids[i]]
			b := analysis.Nodes[ids[j]]
			suggestions = append(suggestions, suggestionsForPair(analysis, a, b)...)
		}
	}
	sort.SliceStable(suggestions, func(i, j int) bool {
		if suggestions[i].SourceID != suggestions[j].SourceID {
			return suggestions[i].SourceID < suggestions[j].SourceID
		}
		if suggestions[i].TargetID != suggestions[j].TargetID {
			return suggestions[i].TargetID < suggestions[j].TargetID
		}
		return suggestions[i].Type < suggestions[j].Type
	})
	return suggestions
}

func suggestionsForPair(analysis ragnode.RAGAnalysis, a, b ragnode.RAGNode) []EdgeSuggestion {
	var out []EdgeSuggestion
	if a.SupersededBy == b.ID && !hasEdge(analysis, a.ID, b.ID, ragnode.EdgeSupersedes) {
		out = append(out, EdgeSuggestion{SourceID: b.ID, TargetID: a.ID, Type: ragnode.EdgeSupersedes, Confidence: 1.00, Reason: "superseded_by metadata"})
	}
	if b.SupersededBy == a.ID && !hasEdge(analysis, b.ID, a.ID, ragnode.EdgeSupersedes) {
		out = append(out, EdgeSuggestion{SourceID: a.ID, TargetID: b.ID, Type: ragnode.EdgeSupersedes, Confidence: 1.00, Reason: "superseded_by metadata"})
	}
	if a.CommunityID != "" && a.CommunityID == b.CommunityID && topicOverlap(a.Topics, b.Topics) >= 0.60 && !hasAnyEdgeBetween(analysis, a.ID, b.ID, ragnode.EdgeSameQuestion) {
		out = append(out, EdgeSuggestion{SourceID: a.ID, TargetID: b.ID, Type: ragnode.EdgeSameQuestion, Confidence: 0.70, Reason: "topic overlap in same community"})
	}
	if titleSimilarity(a.Title, b.Title) >= 0.88 && !hasAnyEdgeBetween(analysis, a.ID, b.ID, ragnode.EdgeDuplicate) {
		out = append(out, EdgeSuggestion{SourceID: a.ID, TargetID: b.ID, Type: ragnode.EdgeDuplicate, Confidence: 0.90, Reason: "title similarity"})
	}
	if a.CommunityID != "" && a.CommunityID == b.CommunityID {
		if a.ChunkType == ragnode.ChunkCommunitySummary && !hasEdge(analysis, b.ID, a.ID, ragnode.EdgeCommunityMemberOf) {
			out = append(out, EdgeSuggestion{SourceID: b.ID, TargetID: a.ID, Type: ragnode.EdgeCommunityMemberOf, Confidence: 0.80, Reason: "shared community summary"})
		}
		if b.ChunkType == ragnode.ChunkCommunitySummary && !hasEdge(analysis, a.ID, b.ID, ragnode.EdgeCommunityMemberOf) {
			out = append(out, EdgeSuggestion{SourceID: a.ID, TargetID: b.ID, Type: ragnode.EdgeCommunityMemberOf, Confidence: 0.80, Reason: "shared community summary"})
		}
	}
	if containsFold(a.Title, "policy") && containsFold(b.Title, "exception") && !hasEdge(analysis, b.ID, a.ID, ragnode.EdgeExceptionTo) {
		out = append(out, EdgeSuggestion{SourceID: b.ID, TargetID: a.ID, Type: ragnode.EdgeExceptionTo, Confidence: 0.55, Reason: "policy exception title cue"})
	}
	if containsFold(b.Title, "policy") && containsFold(a.Title, "exception") && !hasEdge(analysis, a.ID, b.ID, ragnode.EdgeExceptionTo) {
		out = append(out, EdgeSuggestion{SourceID: a.ID, TargetID: b.ID, Type: ragnode.EdgeExceptionTo, Confidence: 0.55, Reason: "policy exception title cue"})
	}
	return out
}

func hasAnyEdgeBetween(analysis ragnode.RAGAnalysis, a, b, typ string) bool {
	return hasEdge(analysis, a, b, typ) || hasEdge(analysis, b, a, typ)
}

func hasEdge(analysis ragnode.RAGAnalysis, source, target, typ string) bool {
	for _, edge := range analysis.Edges {
		if edge.SourceID == source && edge.TargetID == target && edge.Type == typ {
			return true
		}
	}
	return false
}

func topicOverlap(a, b []string) float64 {
	setA := stringSet(a)
	setB := stringSet(b)
	if len(setA) == 0 || len(setB) == 0 {
		return 0
	}
	intersection := 0
	for term := range setA {
		if setB[term] {
			intersection++
		}
	}
	denominator := len(setA)
	if len(setB) < denominator {
		denominator = len(setB)
	}
	return float64(intersection) / float64(denominator)
}

func titleSimilarity(a, b string) float64 {
	setA := tokenSet(a)
	setB := tokenSet(b)
	if len(setA) == 0 || len(setB) == 0 {
		return 0
	}
	intersection := 0
	for term := range setA {
		if setB[term] {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	return float64(intersection) / float64(union)
}

func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		if term := strings.ToLower(strings.TrimSpace(value)); term != "" {
			out[term] = true
		}
	}
	return out
}

func tokenSet(value string) map[string]bool {
	return stringSet(strings.FieldsFunc(value, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9')
	}))
}

func containsFold(value, needle string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(needle))
}
