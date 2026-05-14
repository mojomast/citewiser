package packer

import (
	"testing"

	"github.com/mojomast/citewiser/pkg/citewise"
	"github.com/mojomast/citewiser/pkg/ragnode"
)

func TestPackQueryTypeRequiredSlots(t *testing.T) {
	cases := []struct {
		queryType QueryType
		required  []SlotType
	}{
		{QueryFactual, []SlotType{SlotFoundation}},
		{QueryComparative, []SlotType{SlotFoundation, SlotBridge, SlotCounterpoint}},
		{QueryProcedural, []SlotType{SlotFoundation, SlotProcedure}},
		{QueryExploratory, []SlotType{SlotOverview, SlotBridge}},
		{QueryTemporal, []SlotType{SlotFoundation}},
		{QueryAdversarial, []SlotType{SlotFoundation, SlotCounterpoint}},
		{QueryAgentic, []SlotType{SlotFoundation, SlotProcedure, SlotPermission, SlotDecision}},
	}
	for _, tc := range cases {
		t.Run(string(tc.queryType), func(t *testing.T) {
			plan := NewPacker().Pack(packAnalysis(allSlotNodes()), tc.queryType, 1000, string(ragnode.SensitivityInternal))
			if plan.HygieneSignal != HygieneGreen {
				t.Fatalf("hygiene got %s want green: %+v", plan.HygieneSignal, plan.Suppressed)
			}
			for _, required := range tc.required {
				if countPlanSlot(plan, required) == 0 {
					t.Fatalf("missing required slot %s in %+v", required, plan.Slots)
				}
			}
		})
	}
}

func TestPackAgenticHardFailures(t *testing.T) {
	cases := []struct {
		name  string
		nodes []ragnode.RAGNode
	}{
		{"no permission", withoutSlot(allSlotNodes(), ragnode.ChunkPermissionRecord)},
		{"unapproved permission", mutateNode(allSlotNodes(), "permission", func(n *ragnode.RAGNode) { n.ApprovedBy = nil })},
		{"missing procedure", withoutSlot(allSlotNodes(), ragnode.ChunkProcedure)},
		{"missing decision", withoutSlot(allSlotNodes(), ragnode.ChunkDecision)},
		{"stale only permission", mutateNode(onlyAgenticRequiredNodes(), "permission", func(n *ragnode.RAGNode) { n.SupersededBy = "permission-v2" })},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan := NewPacker().Pack(packAnalysis(tc.nodes), QueryAgentic, 1000, string(ragnode.SensitivityInternal))
			if plan.HygieneSignal != HygieneRed {
				t.Fatalf("hygiene got %s want red: %+v", plan.HygieneSignal, plan)
			}
		})
	}
}

func TestPackSuppressesDuplicateExceptComparativeDifferentChunk(t *testing.T) {
	nodes := []ragnode.RAGNode{slotNode("foundation-a", ragnode.ChunkSection, citewise.RoleFoundation), slotNode("foundation-b", ragnode.ChunkSection, citewise.RoleFoundation), slotNode("bridge", ragnode.ChunkChunk, citewise.RoleBridge), slotNode("counter", ragnode.ChunkChunk, citewise.RoleCounterpoint)}
	analysis := packAnalysis(nodes)
	analysis.Duplicates = [][]string{{"foundation-a", "foundation-b"}}
	plan := NewPacker().Pack(analysis, QueryFactual, 1000, string(ragnode.SensitivityInternal))
	if countSuppressed(plan, SuppressDuplicate) != 1 {
		t.Fatalf("expected duplicate suppression, got %+v", plan.Suppressed)
	}

	nodes[1].ChunkType = ragnode.ChunkDocument
	analysis = packAnalysis(nodes)
	analysis.Duplicates = [][]string{{"foundation-a", "foundation-b"}}
	plan = NewPacker().Pack(analysis, QueryComparative, 1000, string(ragnode.SensitivityInternal))
	if countSuppressed(plan, SuppressDuplicate) != 0 {
		t.Fatalf("comparative different chunk duplicate should be allowed: %+v", plan.Suppressed)
	}
}

func TestPackSuppressesAccessDeniedNodes(t *testing.T) {
	node := slotNode("secret", ragnode.ChunkSection, citewise.RoleFoundation)
	node.Sensitivity = ragnode.SensitivityConfidential
	plan := NewPacker().Pack(packAnalysis([]ragnode.RAGNode{node}), QueryFactual, 1000, string(ragnode.SensitivityPublic))
	if len(plan.Slots) != 0 || countSuppressed(plan, SuppressAccessControl) != 1 || plan.HygieneSignal != HygieneRed {
		t.Fatalf("expected access suppression only, got %+v", plan)
	}
	if plan.SuppressedByReason[SuppressAccessControl] != 1 {
		t.Fatalf("suppressed breakdown got %+v", plan.SuppressedByReason)
	}
}

func TestPackPreservesTableLocatorInSlotSource(t *testing.T) {
	node := slotNode("metric", ragnode.ChunkSection, citewise.RoleFoundation)
	node.Locator = ragnode.Locator{DocumentID: "doc-metrics", TableID: "table-revenue", RowStart: 7, RowEnd: 9}
	plan := NewPacker().Pack(packAnalysis([]ragnode.RAGNode{node}), QueryFactual, 1000, string(ragnode.SensitivityInternal))
	if len(plan.Slots) != 1 {
		t.Fatalf("slot count got %d: %+v", len(plan.Slots), plan)
	}
	locator := plan.Slots[0].Source.Locator
	if locator.TableID != "table-revenue" || locator.RowStart != 7 || locator.RowEnd != 9 {
		t.Fatalf("table locator not preserved in slot source: %+v", locator)
	}
}

func allSlotNodes() []ragnode.RAGNode {
	return []ragnode.RAGNode{
		slotNode("overview", ragnode.ChunkCommunitySummary, citewise.RoleOverview),
		slotNode("foundation-a", ragnode.ChunkSection, citewise.RoleFoundation),
		slotNode("foundation-b", ragnode.ChunkDocument, citewise.RoleFoundation),
		slotNode("bridge", ragnode.ChunkChunk, citewise.RoleBridge),
		slotNode("counter", ragnode.ChunkClaim, citewise.RoleCounterpoint),
		slotNode("procedure", ragnode.ChunkProcedure, citewise.RoleFoundation),
		slotNode("permission", ragnode.ChunkPermissionRecord, citewise.RoleFoundation),
		slotNode("decision", ragnode.ChunkDecision, citewise.RoleFoundation),
	}
}

func onlyAgenticRequiredNodes() []ragnode.RAGNode {
	return []ragnode.RAGNode{slotNode("foundation-a", ragnode.ChunkSection, citewise.RoleFoundation), slotNode("procedure", ragnode.ChunkProcedure, citewise.RoleFoundation), slotNode("permission", ragnode.ChunkPermissionRecord, citewise.RoleFoundation), slotNode("decision", ragnode.ChunkDecision, citewise.RoleFoundation)}
}

func slotNode(id string, chunkType ragnode.ChunkType, role string) ragnode.RAGNode {
	return ragnode.RAGNode{Item: citewise.Item{ID: id, Title: id, Source: "kb", Trust: 1}, Text: id + " text", ChunkType: chunkType, TokenCount: 20, Version: "v1", Sensitivity: ragnode.SensitivityInternal, ApprovedBy: []string{"ops"}, SourceTrail: []ragnode.SourceHop{{NodeID: id, EdgeType: "retrieved", Confidence: 1}}, Attributes: map[string]string{"role": role}}
}

func packAnalysis(nodes []ragnode.RAGNode) ragnode.RAGAnalysis {
	analysis := ragnode.RAGAnalysis{QueryID: "q1", Query: "current action", Nodes: map[string]ragnode.RAGNode{}, Candidates: map[string]ragnode.Candidate{}, Roles: map[string]string{}, Centrality: map[string]float64{}, EdgesIn: map[string][]ragnode.Edge{}, EdgesOut: map[string][]ragnode.Edge{}}
	for _, node := range nodes {
		analysis.Nodes[node.ID] = node
		analysis.Candidates[node.ID] = ragnode.Candidate{NodeID: node.ID, QueryRelevance: 1}
		analysis.Roles[node.ID] = node.Attributes["role"]
		analysis.Centrality[node.ID] = 1
	}
	return analysis
}

func countPlanSlot(plan ContextPlan, slotType SlotType) int {
	count := 0
	for _, slot := range plan.Slots {
		if slot.SlotType == slotType {
			count++
		}
	}
	return count
}

func countSuppressed(plan ContextPlan, reason string) int {
	count := 0
	for _, suppressed := range plan.Suppressed {
		if suppressed.Reason == reason {
			count++
		}
	}
	return count
}

func withoutSlot(nodes []ragnode.RAGNode, chunkType ragnode.ChunkType) []ragnode.RAGNode {
	out := []ragnode.RAGNode{}
	for _, node := range nodes {
		if node.ChunkType != chunkType {
			out = append(out, node)
		}
	}
	return out
}

func mutateNode(nodes []ragnode.RAGNode, id string, mutate func(*ragnode.RAGNode)) []ragnode.RAGNode {
	out := append([]ragnode.RAGNode(nil), nodes...)
	for i := range out {
		if out[i].ID == id {
			mutate(&out[i])
		}
	}
	return out
}
