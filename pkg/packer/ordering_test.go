package packer

import "testing"

func TestOrderBandsAndAgenticPriority(t *testing.T) {
	plan := ContextPlan{Slots: []ContextSlot{
		{NodeID: "procedure", SlotType: SlotProcedure, Score: 99, TokenCount: 10},
		{NodeID: "permission", SlotType: SlotPermission, Score: 80, TokenCount: 10},
		{NodeID: "foundation", SlotType: SlotFoundation, Score: 90, TokenCount: 10},
		{NodeID: "counter", SlotType: SlotCounterpoint, Score: 100, TokenCount: 10},
		{NodeID: "overview", SlotType: SlotOverview, Score: 100, TokenCount: 10},
	}}
	got := orderAndTrim(plan, QueryAgentic, 1000)
	want := []string{"permission", "foundation", "procedure", "overview", "counter"}
	assertNodeOrder(t, got, want)
	if got.Slots[0].Position != PositionFront || got.Slots[4].Position != PositionBack {
		t.Fatalf("unexpected positions: %+v", got.Slots)
	}
}

func TestOrderTieBreaksByScoreThenNodeID(t *testing.T) {
	plan := ContextPlan{Slots: []ContextSlot{
		{NodeID: "b", SlotType: SlotOverview, Score: 50, TokenCount: 10},
		{NodeID: "a", SlotType: SlotOverview, Score: 50, TokenCount: 10},
		{NodeID: "c", SlotType: SlotOverview, Score: 60, TokenCount: 10},
	}}
	got := orderAndTrim(plan, QueryFactual, 1000)
	assertNodeOrder(t, got, []string{"c", "a", "b"})
}

func TestTrimBudgetOrder(t *testing.T) {
	plan := ContextPlan{Slots: []ContextSlot{
		{NodeID: "foundation", SlotType: SlotFoundation, Score: 90, TokenCount: 50},
		{NodeID: "support", SlotType: SlotSupport, Score: 10, TokenCount: 30},
		{NodeID: "overview", SlotType: SlotOverview, Score: 20, TokenCount: 30},
		{NodeID: "bridge", SlotType: SlotBridge, Score: 30, TokenCount: 30},
	}}
	got := orderAndTrim(plan, QueryFactual, 80)
	assertNodeOrder(t, got, []string{"foundation", "bridge"})
	if countSuppressed(got, SuppressBudget) != 2 {
		t.Fatalf("expected two budget suppressions, got %+v", got.Suppressed)
	}
	if got.TokensUsed != 80 {
		t.Fatalf("tokens used got %d want 80", got.TokensUsed)
	}
}

func TestRequiredSlotOverBudgetReturnsRed(t *testing.T) {
	plan := ContextPlan{Slots: []ContextSlot{{NodeID: "foundation", SlotType: SlotFoundation, Score: 90, TokenCount: 120}}}
	got := orderAndTrim(plan, QueryFactual, 100)
	if got.HygieneSignal != HygieneRed || len(got.Slots) != 1 {
		t.Fatalf("required over-budget slot should remain with red signal: %+v", got)
	}
}

func assertNodeOrder(t *testing.T, plan ContextPlan, want []string) {
	t.Helper()
	if len(plan.Slots) != len(want) {
		t.Fatalf("slot count got %d want %d: %+v", len(plan.Slots), len(want), plan.Slots)
	}
	for i, wantID := range want {
		if plan.Slots[i].NodeID != wantID || plan.EvidencePath[i] != wantID || plan.Slots[i].Index != i {
			t.Fatalf("at %d got slot=%s evidence=%s index=%d want %s", i, plan.Slots[i].NodeID, plan.EvidencePath[i], plan.Slots[i].Index, wantID)
		}
	}
}
