package ranker

import "github.com/mojomast/citewiseussy/pkg/ragnode"

// Diversity returns a bounded bonus for selecting a node after already-selected
// context. Reusing a source costs more than reusing a graph community because
// source monoculture is the stronger assembly risk.
func Diversity(node ragnode.RAGNode, selected []ragnode.RAGNode) float64 {
	score := 1.0
	if node.Source != "" && sourceAlreadySelected(node.Source, selected) {
		score -= 0.35
	}
	if node.CommunityID != "" && communityAlreadySelected(node.CommunityID, selected) {
		score -= 0.25
	}
	return clamp01(score)
}

func sourceAlreadySelected(source string, selected []ragnode.RAGNode) bool {
	for _, node := range selected {
		if node.Source == source {
			return true
		}
	}
	return false
}

func communityAlreadySelected(communityID string, selected []ragnode.RAGNode) bool {
	for _, node := range selected {
		if node.CommunityID == communityID {
			return true
		}
	}
	return false
}
