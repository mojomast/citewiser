package provenance

import "github.com/mojomast/citewiser/pkg/ragnode"

const (
	EdgeRetrieved    = "retrieved"
	EdgeSlotRequired = "slot-required"
)

type TrailOptions struct {
	Candidate    *ragnode.Candidate
	RequiredSlot bool
	ScoreTotal   float64
}

func BuildSourceRef(node ragnode.RAGNode) ragnode.SourceRef {
	return ragnode.SourceRef{
		NodeID:      node.ID,
		Origin:      node.Origin,
		Source:      node.Source,
		URL:         node.URL,
		Version:     node.Version,
		ObservedAt:  node.ObservedAt,
		UpdatedAt:   node.UpdatedAt,
		Locator:     node.Locator,
		CommunityID: node.CommunityID,
	}
}

func BuildSourceTrail(analysis ragnode.RAGAnalysis, node ragnode.RAGNode, opts TrailOptions) []ragnode.SourceHop {
	var trail []ragnode.SourceHop
	if opts.Candidate != nil {
		if opts.Candidate.NodeID == node.ID {
			trail = appendHop(trail, ragnode.SourceHop{NodeID: node.ID, EdgeType: EdgeRetrieved, Confidence: opts.Candidate.QueryRelevance})
		} else {
			trail = appendCandidateEdgeHops(trail, analysis, opts.Candidate.NodeID, node.ID)
		}
	}
	if opts.RequiredSlot {
		trail = appendHop(trail, ragnode.SourceHop{NodeID: node.ID, EdgeType: EdgeSlotRequired, Confidence: clamp01(opts.ScoreTotal / 100)})
	}
	if node.ChunkType == ragnode.ChunkCommunitySummary {
		trail = appendCommunityMemberHops(trail, analysis, node.ID)
	}
	if node.ChunkType == ragnode.ChunkDecision {
		trail = appendDecisionBasisHops(trail, analysis, node.ID)
	}
	return trail
}

func appendCandidateEdgeHops(trail []ragnode.SourceHop, analysis ragnode.RAGAnalysis, candidateID, nodeID string) []ragnode.SourceHop {
	for _, edge := range analysis.Edges {
		if edge.SourceID == candidateID && edge.TargetID == nodeID || edge.SourceID == nodeID && edge.TargetID == candidateID {
			trail = appendHop(trail, ragnode.SourceHop{NodeID: edge.SourceID, EdgeType: edge.Type, Confidence: edge.Confidence})
		}
	}
	return trail
}

func appendCommunityMemberHops(trail []ragnode.SourceHop, analysis ragnode.RAGAnalysis, nodeID string) []ragnode.SourceHop {
	for _, edge := range analysis.Edges {
		if edge.Type == ragnode.EdgeCommunityMemberOf && edge.TargetID == nodeID {
			trail = appendHop(trail, ragnode.SourceHop{NodeID: edge.SourceID, EdgeType: edge.Type, Confidence: edge.Confidence})
		}
	}
	return trail
}

func appendDecisionBasisHops(trail []ragnode.SourceHop, analysis ragnode.RAGAnalysis, nodeID string) []ragnode.SourceHop {
	for _, edge := range analysis.Edges {
		if edge.Type == ragnode.EdgeDecisionBasis && edge.SourceID == nodeID {
			trail = appendHop(trail, ragnode.SourceHop{NodeID: edge.TargetID, EdgeType: edge.Type, Confidence: edge.Confidence})
		}
	}
	return trail
}

func appendHop(trail []ragnode.SourceHop, hop ragnode.SourceHop) []ragnode.SourceHop {
	for _, existing := range trail {
		if existing.NodeID == hop.NodeID && existing.EdgeType == hop.EdgeType {
			return trail
		}
	}
	return append(trail, hop)
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
