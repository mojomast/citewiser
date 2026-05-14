package packer

import "testing"

func TestContextPlanPlanHash(t *testing.T) {
	base := ContextPlan{
		QueryID:       "q1",
		HygieneSignal: HygieneGreen,
		Slots:         []ContextSlot{{NodeID: "a", Text: "first"}, {NodeID: "b", Text: "second"}},
		Suppressed:    []SuppressedEntry{{NodeID: "s1", Reason: SuppressAccessControl}},
	}
	same := base
	same.Slots = append([]ContextSlot(nil), base.Slots...)
	same.Suppressed = append([]SuppressedEntry(nil), base.Suppressed...)
	if base.PlanHash() != same.PlanHash() {
		t.Fatal("identical plans should hash the same")
	}

	changedSlot := same
	changedSlot.Slots = append([]ContextSlot(nil), same.Slots...)
	changedSlot.Slots[1].NodeID = "c"
	if base.PlanHash() == changedSlot.PlanHash() {
		t.Fatal("changing slot node ID should change hash")
	}

	changedText := same
	changedText.Slots = append([]ContextSlot(nil), same.Slots...)
	changedText.Slots[0].Text = "different text"
	if base.PlanHash() != changedText.PlanHash() {
		t.Fatal("changing text content should not change hash")
	}
}
