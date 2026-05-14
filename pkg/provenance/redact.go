package provenance

import (
	"github.com/mojomast/citewiser/pkg/access"
	"github.com/mojomast/citewiser/pkg/ragnode"
)

func RedactSourceTrail(ctx access.Context, controller access.Controller, analysis ragnode.RAGAnalysis, trail []ragnode.SourceHop) []ragnode.SourceHop {
	if controller == nil {
		controller = access.NewController()
	}
	redacted := make([]ragnode.SourceHop, 0, len(trail))
	for _, hop := range trail {
		node, ok := analysis.Nodes[hop.NodeID]
		if !ok {
			continue
		}
		if controller.CanSeeNode(ctx, node).Allowed {
			redacted = append(redacted, hop)
		}
	}
	return redacted
}

func RedactSourceRef(ctx access.Context, controller access.Controller, node ragnode.RAGNode) ragnode.SourceRef {
	if controller == nil {
		controller = access.NewController()
	}
	if !controller.CanSeeNode(ctx, node).Allowed {
		return ragnode.SourceRef{NodeID: node.ID}
	}
	return BuildSourceRef(node)
}
