package memory

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/mojomast/citewiseussy/pkg/access"
	"github.com/mojomast/citewiseussy/pkg/packer"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

func TestStoreLoadPayloadAndHash(t *testing.T) {
	path := filepath.Join(t.TempDir(), "memory.jsonl")
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	store := &FileStore{Path: path, Caller: access.Context{Clearance: access.ClearanceInternal}, Now: func() time.Time { return now }}
	plan := testPlan("q1", packer.QueryFactual, packer.HygieneGreen)

	if err := store.StoreContextPlan("q1", plan); err != nil {
		t.Fatalf("StoreContextPlan() error = %v", err)
	}
	loaded, ok, err := store.LoadPriorPlan("q1")
	if err != nil || !ok {
		t.Fatalf("LoadPriorPlan() ok=%v err=%v", ok, err)
	}
	payload, ok := loaded.WriteBackPayload.(map[string]any)
	if !ok {
		t.Fatalf("WriteBackPayload type = %T, want decoded object", loaded.WriteBackPayload)
	}
	if payload["query_id"] != "q1" || payload["query_type"] != string(packer.QueryFactual) || payload["hygiene_signal"] != string(packer.HygieneGreen) {
		t.Fatalf("payload = %#v", payload)
	}
	versions := map[string]string{"n1": "v1", "n2": "v2"}
	wantHash := PlanHash(packer.QueryFactual, []string{"n1", "n2"}, versions)
	if loadedHash, _ := payload["plan_hash"].(string); loadedHash != wantHash {
		t.Fatalf("plan_hash = %q, want %q", loadedHash, wantHash)
	}
	if got := PlanHash(packer.QueryFactual, []string{"n1", "n2"}, map[string]string{"n2": "v2", "n1": "v1"}); got != wantHash {
		t.Fatalf("PlanHash unstable = %q, want %q", got, wantHash)
	}
}

func TestStoreRejectsRedPlan(t *testing.T) {
	store := &FileStore{Path: filepath.Join(t.TempDir(), "memory.jsonl")}
	if err := store.StoreContextPlan("q1", testPlan("q1", packer.QueryFactual, packer.HygieneRed)); !errors.Is(err, ErrRedPlan) {
		t.Fatalf("StoreContextPlan(red) error = %v, want ErrRedPlan", err)
	}
}

func TestSimilarPriorPlansAndReuseRejection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "memory.jsonl")
	store := &FileStore{Path: path, Caller: access.Context{Clearance: access.ClearanceRestricted}}
	if err := store.StoreContextPlan("q1", testPlan("q1", packer.QueryFactual, packer.HygieneGreen)); err != nil {
		t.Fatal(err)
	}
	other := testPlan("q2", packer.QueryProcedural, packer.HygieneGreen)
	other.Slots[0].Source.CommunityID = "unrelated"
	other.Slots[1].Source.CommunityID = "other"
	if err := store.StoreContextPlan("q2", other); err != nil {
		t.Fatal(err)
	}
	plans, err := store.SimilarPriorPlans([]string{"topic-a", "topic-b"}, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(plans) != 1 || plans[0].QueryID != "q1" {
		t.Fatalf("SimilarPriorPlans() = %#v, want q1 only", plans)
	}

	store.CurrentNodes = map[string]ragnode.RAGNode{"n1": {Version: "v-new"}}
	if rejection, ok := store.ReuseRejection(testPlan("q1", packer.QueryFactual, packer.HygieneGreen), []string{"topic-a", "topic-b"}, packer.QueryFactual); !ok || rejection.Reason != RejectVersionMismatch {
		t.Fatalf("version rejection = %#v ok=%v", rejection, ok)
	}
	store.CurrentNodes = map[string]ragnode.RAGNode{"n1": {Version: "v1", SupersededBy: "n3"}}
	if rejection, ok := store.ReuseRejection(testPlan("q1", packer.QueryFactual, packer.HygieneGreen), []string{"topic-a", "topic-b"}, packer.QueryFactual); !ok || rejection.Reason != RejectSuperseded {
		t.Fatalf("superseded rejection = %#v ok=%v", rejection, ok)
	}
	if rejection, ok := store.ReuseRejection(testPlan("q1", packer.QueryProcedural, packer.HygieneGreen), []string{"topic-a", "topic-b"}, packer.QueryFactual); !ok || rejection.Reason != RejectQueryType {
		t.Fatalf("query type rejection = %#v ok=%v", rejection, ok)
	}
}

func TestLoadRegatesAccess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "memory.jsonl")
	writer := &FileStore{Path: path, Caller: access.Context{Clearance: access.ClearanceRestricted}}
	if err := writer.StoreContextPlan("q1", testPlan("q1", packer.QueryFactual, packer.HygieneGreen)); err != nil {
		t.Fatal(err)
	}
	reader := &FileStore{
		Path:   path,
		Caller: access.Context{Clearance: access.ClearanceInternal},
		CurrentNodes: map[string]ragnode.RAGNode{
			"n1": {Sensitivity: ragnode.SensitivityRestricted},
			"n2": {Sensitivity: ragnode.SensitivityInternal},
		},
	}
	loaded, ok, err := reader.LoadPriorPlan("q1")
	if err != nil || !ok {
		t.Fatalf("LoadPriorPlan() ok=%v err=%v", ok, err)
	}
	if len(loaded.Slots) != 1 || loaded.Slots[0].NodeID != "n2" {
		t.Fatalf("regated slots = %#v, want only n2", loaded.Slots)
	}
	if len(loaded.Suppressed) == 0 || loaded.Suppressed[len(loaded.Suppressed)-1].Reason != access.ReasonAccessControl {
		t.Fatalf("suppressed = %#v, want access-control", loaded.Suppressed)
	}
}

func TestConcurrentAppend(t *testing.T) {
	path := filepath.Join(t.TempDir(), "memory.jsonl")
	store := &FileStore{Path: path, Caller: access.Context{Clearance: access.ClearanceInternal}}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			plan := testPlan(string(rune('a'+i)), packer.QueryFactual, packer.HygieneGreen)
			if err := store.StoreContextPlan(plan.QueryID, plan); err != nil {
				t.Errorf("StoreContextPlan(%d) error = %v", i, err)
			}
		}(i)
	}
	wg.Wait()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	if lines != 8 {
		t.Fatalf("stored lines = %d, want 8; data=%s", lines, data)
	}
}

func testPlan(queryID string, queryType packer.QueryType, signal packer.HygieneSignal) packer.ContextPlan {
	return packer.ContextPlan{
		QueryID:       queryID,
		QueryType:     queryType,
		HygieneSignal: signal,
		EvidencePath:  []string{"n1", "n2"},
		Slots: []packer.ContextSlot{
			{Index: 0, NodeID: "n1", SlotType: packer.SlotFoundation, Title: "A", Text: "secret", TokenCount: 4, Source: ragnode.SourceRef{NodeID: "n1", Version: "v1", CommunityID: "topic-a"}},
			{Index: 1, NodeID: "n2", SlotType: packer.SlotFoundation, Title: "B", Text: "public", TokenCount: 5, Source: ragnode.SourceRef{NodeID: "n2", Version: "v2", CommunityID: "topic-b"}},
		},
	}
}
