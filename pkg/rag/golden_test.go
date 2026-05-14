package rag

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mojomast/citewiseussy/pkg/packer"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

func TestGoldenContextPlans(t *testing.T) {
	cases := map[string]packer.ContextPlan{
		"agentic_green.json":  goldenPlan("q-agentic-green", packer.QueryAgentic, packer.HygieneGreen, []packer.ContextSlot{goldenSlot(0, "perm", packer.SlotPermission, "v3"), goldenSlot(1, "proc", packer.SlotProcedure, "v2"), goldenSlot(2, "decision", packer.SlotDecision, "2026-04")}),
		"agentic_red.json":    goldenPlan("q-agentic-red", packer.QueryAgentic, packer.HygieneRed, []packer.ContextSlot{goldenSlot(0, "foundation", packer.SlotFoundation, "v1")}),
		"temporal_stale.json": goldenPlan("q-temporal-stale", packer.QueryTemporal, packer.HygieneRed, []packer.ContextSlot{goldenSlot(0, "old-policy", packer.SlotFoundation, "v1")}),
		"factual_green.json":  goldenPlan("q-factual-green", packer.QueryFactual, packer.HygieneGreen, []packer.ContextSlot{goldenSlot(0, "policy", packer.SlotFoundation, "v4")}),
	}
	for name, plan := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := json.MarshalIndent(plan, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			got = append(got, '\n')
			want, err := os.ReadFile(filepath.Join("..", "..", "testdata", "golden_context_plans", name))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Fatalf("golden mismatch for %s\n got: %s\nwant: %s", name, got, want)
			}
		})
	}
}

func goldenPlan(queryID string, queryType packer.QueryType, signal packer.HygieneSignal, slots []packer.ContextSlot) packer.ContextPlan {
	plan := packer.ContextPlan{QueryID: queryID, QueryType: queryType, Slots: slots, HygieneSignal: signal, CritiqueSummary: "golden fixture"}
	for _, slot := range slots {
		plan.EvidencePath = append(plan.EvidencePath, slot.NodeID)
		plan.SourceTrail = append(plan.SourceTrail, slot.SourceTrail...)
		plan.TokensUsed += slot.TokenCount
	}
	if signal == packer.HygieneRed {
		plan.Suppressed = append(plan.Suppressed, packer.SuppressedEntry{NodeID: "suppressed", Reason: "red-plan", Detail: "fixture"})
	}
	return plan
}

func goldenSlot(index int, nodeID string, slotType packer.SlotType, version string) packer.ContextSlot {
	return packer.ContextSlot{
		Index:      index,
		SlotType:   slotType,
		NodeID:     nodeID,
		Role:       string(slotType),
		Title:      nodeID + " title",
		Text:       nodeID + " text",
		TokenCount: 10,
		Score:      90,
		Position:   packer.PositionFront,
		Source: ragnode.SourceRef{
			NodeID:  nodeID,
			Origin:  "golden",
			Source:  "fixture",
			Version: version,
		},
		SourceTrail: []ragnode.SourceHop{{NodeID: nodeID, EdgeType: "retrieved", Confidence: 1}},
		MustCite:    true,
		Rationale:   "golden fixture",
	}
}
