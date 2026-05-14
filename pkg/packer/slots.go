package packer

import (
	"strings"

	"github.com/mojomast/citewiseussy/pkg/citewise"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
	"github.com/mojomast/citewiseussy/pkg/ranker"
)

type slotRule struct {
	Slot      SlotType
	Min       int
	Max       int
	Required  bool
	Forbidden bool
}

func slotPolicy(queryType QueryType) []slotRule {
	switch queryType {
	case QueryComparative:
		return []slotRule{{SlotOverview, 0, 1, false, false}, {SlotFoundation, 2, 4, true, false}, {SlotBridge, 1, 2, true, false}, {SlotCounterpoint, 1, 1, true, false}, {SlotProcedure, 0, 0, false, true}, {SlotPermission, 0, 1, false, false}, {SlotDecision, 0, 1, false, false}}
	case QueryProcedural:
		return []slotRule{{SlotOverview, 0, 1, false, false}, {SlotFoundation, 1, 2, true, false}, {SlotBridge, 0, 1, false, false}, {SlotCounterpoint, 0, 1, false, false}, {SlotProcedure, 1, 3, true, false}, {SlotPermission, 0, 1, false, false}, {SlotDecision, 0, 1, false, false}}
	case QueryExploratory:
		return []slotRule{{SlotOverview, 1, 3, true, false}, {SlotFoundation, 0, 2, false, false}, {SlotBridge, 1, 3, true, false}, {SlotCounterpoint, 0, 1, false, false}, {SlotProcedure, 0, 0, false, true}, {SlotPermission, 0, 0, false, true}, {SlotDecision, 0, 1, false, false}}
	case QueryTemporal:
		return []slotRule{{SlotOverview, 0, 1, false, false}, {SlotFoundation, 1, 3, true, false}, {SlotBridge, 0, 1, false, false}, {SlotCounterpoint, 0, 1, false, false}, {SlotProcedure, 0, 0, false, true}, {SlotPermission, 0, 1, false, false}, {SlotDecision, 0, 1, false, false}}
	case QueryAdversarial:
		return []slotRule{{SlotOverview, 0, 1, false, false}, {SlotFoundation, 1, 3, true, false}, {SlotBridge, 0, 1, false, false}, {SlotCounterpoint, 1, 2, true, false}, {SlotProcedure, 0, 1, false, false}, {SlotPermission, 0, 1, false, false}, {SlotDecision, 0, 1, false, false}}
	case QueryAgentic:
		return []slotRule{{SlotOverview, 0, 1, false, false}, {SlotFoundation, 1, 2, true, false}, {SlotBridge, 0, 1, false, false}, {SlotCounterpoint, 0, 1, false, false}, {SlotProcedure, 1, 3, true, false}, {SlotPermission, 1, 2, true, false}, {SlotDecision, 1, 2, true, false}}
	default:
		return []slotRule{{SlotOverview, 0, 1, false, false}, {SlotFoundation, 1, 3, true, false}, {SlotBridge, 0, 1, false, false}, {SlotCounterpoint, 0, 1, false, false}, {SlotProcedure, 0, 0, false, true}, {SlotPermission, 0, 1, false, false}, {SlotDecision, 0, 1, false, false}}
	}
}

func slotTypeFor(node ragnode.RAGNode, score ranker.Score) SlotType {
	switch node.ChunkType {
	case ragnode.ChunkPermissionRecord:
		return SlotPermission
	case ragnode.ChunkDecision:
		return SlotDecision
	case ragnode.ChunkProcedure:
		return SlotProcedure
	case ragnode.ChunkCommunitySummary:
		return SlotOverview
	}
	switch score.Role {
	case citewise.RoleOverview:
		return SlotOverview
	case citewise.RoleFoundation:
		return SlotFoundation
	case citewise.RoleBridge:
		return SlotBridge
	case citewise.RoleCounterpoint:
		return SlotCounterpoint
	default:
		return SlotSupport
	}
}

func requiredSlots(queryType QueryType) map[SlotType]int {
	required := map[SlotType]int{}
	for _, rule := range slotPolicy(queryType) {
		if rule.Required {
			required[rule.Slot] = rule.Min
		}
	}
	return required
}

func slotAllowed(queryType QueryType, slot SlotType) bool {
	for _, rule := range slotPolicy(queryType) {
		if rule.Slot == slot {
			return !rule.Forbidden
		}
	}
	return slot == SlotSupport
}

func slotMax(queryType QueryType, slot SlotType) int {
	for _, rule := range slotPolicy(queryType) {
		if rule.Slot == slot {
			return rule.Max
		}
	}
	return 99
}

func isHistoricalQuery(analysis ragnode.RAGAnalysis) bool {
	q := strings.ToLower(analysis.Query)
	return strings.Contains(q, "historical") || strings.Contains(q, "history") || strings.Contains(q, "old") || strings.Contains(q, "as of")
}
