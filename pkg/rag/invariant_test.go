package rag

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"testing/quick"

	"github.com/mojomast/citewiseussy/pkg/access"
	"github.com/mojomast/citewiseussy/pkg/citewise"
	"github.com/mojomast/citewiseussy/pkg/packer"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

func TestAccessInvariantNoUnauthorizedTextInOutputs(t *testing.T) {
	check := func(secret string) bool {
		secret = strings.TrimSpace(secret)
		if secret == "" || len(secret) > 80 || strings.Contains(secret, "public") {
			return true
		}
		set := invariantSet()
		set.Nodes = append(set.Nodes, ragnode.RAGNode{Item: citewise.Item{ID: "secret", Title: "Hidden", Trust: 1}, Text: secret, ChunkType: ragnode.ChunkDocument, Sensitivity: ragnode.SensitivityRestricted})
		set.Candidates = append(set.Candidates, ragnode.Candidate{NodeID: "secret", QueryRelevance: 1})
		resp, err := NewPipeline().Run(Request{CandidateSet: set, Access: access.Context{Clearance: access.ClearancePublic}, AllowDegradedPlan: true})
		if err != nil {
			return false
		}
		out, err := json.Marshal(struct {
			Ranked any `json:"ranked"`
			Plan   any `json:"plan"`
		}{resp.Ranked, resp.Plan})
		return err == nil && !bytes.Contains(out, []byte(secret))
	}
	if err := quick.Check(check, &quick.Config{MaxCount: 50}); err != nil {
		t.Fatal(err)
	}
}

func TestBudgetDeterminismAndProvenanceInvariants(t *testing.T) {
	check := func(budget uint16) bool {
		limit := int(budget%200) + 1
		req := Request{CandidateSet: invariantSet(), Access: access.Context{Clearance: access.ClearanceInternal}, TokenBudget: limit, AllowDegradedPlan: true}
		first, err := NewPipeline().Run(req)
		if err != nil {
			return false
		}
		second, err := NewPipeline().Run(req)
		if err != nil {
			return false
		}
		firstJSON, err := json.Marshal(first.Plan)
		if err != nil {
			return false
		}
		secondJSON, err := json.Marshal(second.Plan)
		if err != nil || !bytes.Equal(firstJSON, secondJSON) {
			return false
		}
		if first.Plan.TokensUsed > limit && first.Plan.HygieneSignal == packer.HygieneGreen {
			return false
		}
		path := map[string]bool{}
		for _, id := range first.Plan.EvidencePath {
			path[id] = true
		}
		for _, slot := range first.Plan.Slots {
			if !path[slot.NodeID] || slot.Source.NodeID == "" {
				return false
			}
		}
		return true
	}
	if err := quick.Check(check, &quick.Config{MaxCount: 50}); err != nil {
		t.Fatal(err)
	}
}

func TestDuplicateInvariantForFactualPlans(t *testing.T) {
	set := invariantSet()
	set.Nodes = append(set.Nodes, ragnode.RAGNode{Item: citewise.Item{ID: "dup", Title: "Foundation duplicate", Notes: "duplicate policy", Trust: 1, GoalFit: 1}, Text: "duplicate policy", ChunkType: ragnode.ChunkDocument, Sensitivity: ragnode.SensitivityInternal, Version: "v1"})
	set.Edges = append(set.Edges, ragnode.Edge{SourceID: "dup", TargetID: "public", Type: ragnode.EdgeDuplicate, Confidence: 1})
	set.Candidates = append(set.Candidates, ragnode.Candidate{NodeID: "dup", QueryRelevance: 0.99})
	resp, err := NewPipeline().Run(Request{CandidateSet: set, Access: access.Context{Clearance: access.ClearanceInternal}, QueryType: packer.QueryFactual, AllowDegradedPlan: true})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, slot := range resp.Plan.Slots {
		if slot.NodeID == "public" || slot.NodeID == "dup" {
			count++
		}
	}
	if count > 1 {
		t.Fatalf("duplicate cluster contributed %d slots, want at most 1: %#v", count, resp.Plan.Slots)
	}
}

func invariantSet() ragnode.CandidateSet {
	return ragnode.CandidateSet{
		QueryID: "q-invariant",
		Query:   "refund policy",
		Nodes: []ragnode.RAGNode{
			{Item: citewise.Item{ID: "public", Title: "Foundation refund policy", Notes: "public policy", Trust: 1, GoalFit: 1}, Text: "public policy", ChunkType: ragnode.ChunkDocument, Sensitivity: ragnode.SensitivityPublic, Version: "v1", SourceTrail: []ragnode.SourceHop{{NodeID: "public", EdgeType: "retrieved", Confidence: 1}}},
		},
		Candidates: []ragnode.Candidate{{NodeID: "public", QueryRelevance: 1}},
	}
}
