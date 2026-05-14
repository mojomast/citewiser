package memory

import (
	"github.com/mojomast/citewiser/pkg/access"
	"github.com/mojomast/citewiser/pkg/packer"
)

func (s *FileStore) regate(plan packer.ContextPlan) packer.ContextPlan {
	controller := s.Access
	if controller == nil {
		controller = access.NewController()
	}
	allowed := plan.Slots[:0]
	for _, slot := range plan.Slots {
		node, ok := s.CurrentNodes[slot.NodeID]
		if ok {
			decision := controller.CanSeeNode(s.Caller, node)
			if !decision.Allowed {
				plan.Suppressed = append(plan.Suppressed, packer.SuppressedEntry{NodeID: slot.NodeID, Reason: decision.Reason, Detail: decision.Detail})
				continue
			}
		}
		allowed = append(allowed, slot)
	}
	plan.Slots = allowed
	plan.EvidencePath = nil
	plan.SourceTrail = nil
	plan.TokensUsed = 0
	for i := range plan.Slots {
		plan.Slots[i].Index = i
		plan.EvidencePath = append(plan.EvidencePath, plan.Slots[i].NodeID)
		plan.SourceTrail = append(plan.SourceTrail, plan.Slots[i].SourceTrail...)
		plan.TokensUsed += plan.Slots[i].TokenCount
	}
	return plan
}
