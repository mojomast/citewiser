package packer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mojomast/citewiseussy/pkg/access"
	"github.com/mojomast/citewiseussy/pkg/provenance"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
	"github.com/mojomast/citewiseussy/pkg/ranker"
)

const (
	SuppressAccessControl = "access-control"
	SuppressDuplicate     = "duplicate"
	SuppressStale         = "stale"
	SuppressLowTrust      = "low-trust"
	SuppressBudget        = "budget"
	SuppressSlotPolicy    = "slot-policy"
)

// DefaultPacker creates context plans from ranked nodes.
type DefaultPacker struct {
	Ranker ranker.Ranker
}

// NewPacker returns a packer backed by the default ranker.
func NewPacker() DefaultPacker {
	return DefaultPacker{Ranker: ranker.NewRanker()}
}

// Pack ranks the analysis, applies slot policy, and orders/trims the result.
func (p DefaultPacker) Pack(analysis ragnode.RAGAnalysis, queryType QueryType, tokenBudget int, callerClearance string) ContextPlan {
	r := p.Ranker
	if r == nil {
		r = ranker.NewRanker()
	}
	ranked, _ := r.Rank(access.Context{Clearance: access.Clearance(callerClearance), Attributes: map[string]string{"query_type": string(queryType), access.AttrAllowUnapprovedAgenticNodes: "true"}}, analysis, tokenBudget)
	plan := ContextPlan{QueryID: analysis.QueryID, QueryType: queryType, HygieneSignal: HygieneGreen}
	for _, suppressed := range ranked.Suppressed {
		plan.Suppressed = append(plan.Suppressed, SuppressedEntry{NodeID: suppressed.NodeID, Reason: suppressed.SuppressionReason, Score: suppressed.Total})
	}

	counts := map[SlotType]int{}
	seenDuplicateClusters := map[int]ragnode.ChunkType{}
	for _, rankedNode := range ranked.Ranked {
		node := rankedNode.Node
		score := rankedNode.Score
		slot := slotTypeFor(node, score)
		if !slotAllowed(queryType, slot) {
			plan.Suppressed = append(plan.Suppressed, SuppressedEntry{NodeID: node.ID, Reason: SuppressSlotPolicy, Detail: fmt.Sprintf("%s forbidden for %s", slot, queryType), Score: score.Total})
			continue
		}
		if node.SupersededBy != "" && !(queryType == QueryTemporal && isHistoricalQuery(analysis)) {
			plan.Suppressed = append(plan.Suppressed, SuppressedEntry{NodeID: node.ID, Reason: SuppressStale, Detail: "superseded by " + node.SupersededBy, Score: score.Total})
			continue
		}
		if queryType == QueryAgentic && isControllingSlot(slot) && len(node.ApprovedBy) == 0 {
			plan.Suppressed = append(plan.Suppressed, SuppressedEntry{NodeID: node.ID, Reason: SuppressSlotPolicy, Detail: "agentic controlling node lacks approval", Score: score.Total})
			continue
		}
		if duplicateSuppressed(node, queryType, analysis, seenDuplicateClusters) {
			plan.Suppressed = append(plan.Suppressed, SuppressedEntry{NodeID: node.ID, Reason: SuppressDuplicate, Score: score.Total})
			continue
		}
		if counts[slot] >= slotMax(queryType, slot) {
			plan.Suppressed = append(plan.Suppressed, SuppressedEntry{NodeID: node.ID, Reason: SuppressSlotPolicy, Detail: fmt.Sprintf("%s slot full", slot), Score: score.Total})
			continue
		}
		plan.Slots = append(plan.Slots, buildSlot(analysis, rankedNode, slot))
		counts[slot]++
	}

	plan = orderAndTrim(plan, queryType, tokenBudget)
	plan.HygieneSignal = hygieneSignal(plan, queryType, countSlots(plan.Slots))
	plan.CritiqueSummary = critiqueSummary(plan)
	return plan
}

func buildSlot(analysis ragnode.RAGAnalysis, rankedNode ranker.RankedNode, slotType SlotType) ContextSlot {
	node := rankedNode.Node
	score := rankedNode.Score
	candidate := rankedNode.Candidate
	trail := provenance.BuildSourceTrail(analysis, node, provenance.TrailOptions{Candidate: &candidate, RequiredSlot: true, ScoreTotal: score.Total})
	if len(trail) == 0 {
		trail = node.SourceTrail
	}
	return ContextSlot{
		SlotType:    slotType,
		NodeID:      node.ID,
		Role:        score.Role,
		Title:       node.Title,
		Text:        node.Text,
		TokenCount:  node.EffectiveTokenCount(),
		Score:       score.Total,
		Position:    PositionMiddle,
		Source:      provenance.BuildSourceRef(node),
		SourceTrail: trail,
		MustCite:    slotType != SlotSupport,
		Rationale:   strings.Join(score.Rationale, "; "),
	}
}

func duplicateSuppressed(node ragnode.RAGNode, queryType QueryType, analysis ragnode.RAGAnalysis, seen map[int]ragnode.ChunkType) bool {
	for i, cluster := range analysis.Duplicates {
		for _, id := range cluster {
			if id != node.ID {
				continue
			}
			if firstType, ok := seen[i]; ok {
				return !((queryType == QueryComparative || queryType == QueryAdversarial) && firstType != node.ChunkType)
			}
			seen[i] = node.ChunkType
			return false
		}
	}
	return false
}

func hygieneSignal(plan ContextPlan, queryType QueryType, counts map[SlotType]int) HygieneSignal {
	for slot, min := range requiredSlots(queryType) {
		if counts[slot] < min {
			return HygieneRed
		}
	}
	if queryType == QueryAgentic && agenticHardFailure(plan) {
		return HygieneRed
	}
	for _, slot := range plan.Slots {
		if slot.Score < 55 {
			return HygieneYellow
		}
	}
	return HygieneGreen
}

func agenticHardFailure(plan ContextPlan) bool {
	hasPermission := false
	hasProcedure := false
	for _, slot := range plan.Slots {
		if slot.SlotType == SlotPermission {
			hasPermission = true
		}
		if slot.SlotType == SlotProcedure {
			hasProcedure = true
		}
		if (slot.SlotType == SlotPermission || slot.SlotType == SlotProcedure || slot.SlotType == SlotDecision) && len(slot.SourceTrail) == 0 {
			return true
		}
	}
	return !hasPermission || !hasProcedure
}

func isControllingSlot(slot SlotType) bool {
	return slot == SlotPermission || slot == SlotProcedure || slot == SlotDecision
}

func countSlots(slots []ContextSlot) map[SlotType]int {
	counts := map[SlotType]int{}
	for _, slot := range slots {
		counts[slot.SlotType]++
	}
	return counts
}

func rebuildPlanIndexes(plan ContextPlan) ContextPlan {
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

func critiqueSummary(plan ContextPlan) string {
	return fmt.Sprintf("Packed %d slots with %s hygiene and %d suppressions.", len(plan.Slots), plan.HygieneSignal, len(plan.Suppressed))
}

func sortSuppressions(suppressed []SuppressedEntry) {
	sort.SliceStable(suppressed, func(i, j int) bool {
		if suppressed[i].NodeID != suppressed[j].NodeID {
			return suppressed[i].NodeID < suppressed[j].NodeID
		}
		return suppressed[i].Reason < suppressed[j].Reason
	})
}
