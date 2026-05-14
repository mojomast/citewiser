package ranker

import "github.com/mojomast/citewiser/pkg/ragnode"

// Authority returns the bounded source authority score for a node within an
// analysis, using incoming cites normalized by the analysis maximum.
func Authority(node ragnode.RAGNode, analysis ragnode.RAGAnalysis) float64 {
	return clamp01(
		0.35*clamp01(node.Trust) +
			0.20*normalizedIncomingCites(node.ID, analysis) +
			0.20*approvedByScore(node) +
			0.15*ChunkTypeAuthority(node.ChunkType) +
			0.10*versionCurrentness(node),
	)
}

// ChunkTypeAuthority returns the spec-defined authority prior for a chunk type.
func ChunkTypeAuthority(chunkType ragnode.ChunkType) float64 {
	switch chunkType {
	case ragnode.ChunkPermissionRecord:
		return 1.00
	case ragnode.ChunkDecision:
		return 0.95
	case ragnode.ChunkProcedure:
		return 0.90
	case ragnode.ChunkDefinition:
		return 0.85
	case ragnode.ChunkDocument:
		return 0.80
	case ragnode.ChunkSection:
		return 0.75
	case ragnode.ChunkCommunitySummary:
		return 0.70
	case ragnode.ChunkClaim:
		return 0.65
	case ragnode.ChunkChunk:
		return 0.60
	case ragnode.ChunkEntity:
		return 0.55
	default:
		return 0
	}
}

func normalizedIncomingCites(nodeID string, analysis ragnode.RAGAnalysis) float64 {
	maxCites := 0
	counts := map[string]int{}
	for _, edge := range analysis.Edges {
		if edge.Type != ragnode.EdgeCites {
			continue
		}
		counts[edge.TargetID]++
		if counts[edge.TargetID] > maxCites {
			maxCites = counts[edge.TargetID]
		}
	}
	if maxCites == 0 {
		return 0
	}
	return float64(counts[nodeID]) / float64(maxCites)
}

func approvedByScore(node ragnode.RAGNode) float64 {
	if len(node.ApprovedBy) == 0 {
		return 0
	}
	return 1
}

func versionCurrentness(node ragnode.RAGNode) float64 {
	if node.SupersededBy != "" {
		return 0
	}
	if node.Version == "" {
		return 0.5
	}
	return 1
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
