package packer

import "sort"

func orderAndTrim(plan ContextPlan, queryType QueryType, tokenBudget int) ContextPlan {
	for i := range plan.Slots {
		plan.Slots[i].Position = positionFor(plan.Slots[i], queryType)
	}
	plan.Slots = orderedSlots(plan.Slots, queryType)
	plan = rebuildPlanIndexes(plan)
	if tokenBudget > 0 && plan.TokensUsed > tokenBudget {
		plan = trimToBudget(plan, tokenBudget)
	}
	sortSuppressions(plan.Suppressed)
	return rebuildPlanIndexes(plan)
}

func positionFor(slot ContextSlot, queryType QueryType) SlotPosition {
	if queryType == QueryAgentic && slot.SlotType == SlotPermission {
		return PositionFront
	}
	if slot.SlotType == SlotFoundation || slot.SlotType == SlotProcedure {
		return PositionFront
	}
	if slot.SlotType == SlotDecision {
		return PositionBack
	}
	if slot.SlotType == SlotCounterpoint && (queryType == QueryAdversarial || queryType == QueryComparative || queryType == QueryAgentic) {
		return PositionBack
	}
	if slot.SlotType == SlotBridge && queryType == QueryComparative {
		return PositionFront
	}
	return PositionMiddle
}

func orderedSlots(slots []ContextSlot, queryType QueryType) []ContextSlot {
	front, middle, back := []ContextSlot{}, []ContextSlot{}, []ContextSlot{}
	for _, slot := range slots {
		switch slot.Position {
		case PositionFront:
			front = append(front, slot)
		case PositionBack:
			back = append(back, slot)
		default:
			middle = append(middle, slot)
		}
	}
	sortBand(front, queryType)
	sortBand(middle, queryType)
	sortBand(back, queryType)
	out := append(front, middle...)
	out = append(out, back...)
	return out
}

func sortBand(slots []ContextSlot, queryType QueryType) {
	sort.SliceStable(slots, func(i, j int) bool {
		if slotPriority(slots[i], queryType) != slotPriority(slots[j], queryType) {
			return slotPriority(slots[i], queryType) < slotPriority(slots[j], queryType)
		}
		if slots[i].Score != slots[j].Score {
			return slots[i].Score > slots[j].Score
		}
		return slots[i].NodeID < slots[j].NodeID
	})
}

func slotPriority(slot ContextSlot, queryType QueryType) int {
	if queryType == QueryAgentic {
		switch slot.SlotType {
		case SlotPermission:
			return 0
		case SlotFoundation:
			return 1
		case SlotProcedure:
			return 2
		}
	}
	if queryType == QueryExploratory && slot.SlotType == SlotOverview {
		return 0
	}
	if queryType == QueryComparative && slot.SlotType == SlotBridge {
		return 1
	}
	return 3
}

func trimToBudget(plan ContextPlan, tokenBudget int) ContextPlan {
	for plan.TokensUsed > tokenBudget {
		idx := trimCandidate(plan.Slots, SlotSupport, PositionMiddle, true)
		if idx < 0 {
			idx = trimCandidate(plan.Slots, SlotOverview, PositionMiddle, false)
		}
		if idx < 0 {
			idx = trimCandidate(plan.Slots, SlotBridge, PositionMiddle, false)
		}
		if idx < 0 {
			plan.HygieneSignal = HygieneRed
			return plan
		}
		removed := plan.Slots[idx]
		plan.Suppressed = append(plan.Suppressed, SuppressedEntry{NodeID: removed.NodeID, Reason: SuppressBudget, Detail: "trimmed to token budget", Score: removed.Score})
		plan.Slots = append(plan.Slots[:idx], plan.Slots[idx+1:]...)
		plan = rebuildPlanIndexes(plan)
	}
	return plan
}

func trimCandidate(slots []ContextSlot, slotType SlotType, position SlotPosition, supportOnly bool) int {
	idx := -1
	for i, slot := range slots {
		if slot.Position != position {
			continue
		}
		if supportOnly && slot.SlotType != SlotSupport {
			continue
		}
		if !supportOnly && slot.SlotType != slotType {
			continue
		}
		if idx < 0 || slot.Score < slots[idx].Score || slot.Score == slots[idx].Score && slot.NodeID > slots[idx].NodeID {
			idx = i
		}
	}
	return idx
}
