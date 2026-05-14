package rag

import (
	"errors"
	"testing"

	"github.com/mojomast/citewiseussy/pkg/access"
	"github.com/mojomast/citewiseussy/pkg/citewise"
	"github.com/mojomast/citewiseussy/pkg/packer"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

func TestAnalyzeWrapsInvalidCandidates(t *testing.T) {
	_, err := Analyze(ragnode.CandidateSet{Nodes: []ragnode.RAGNode{{Item: citewise.Item{ID: "n1"}}}, Candidates: []ragnode.Candidate{{NodeID: "missing"}}})
	if !errors.Is(err, ErrInvalidCandidates) {
		t.Fatalf("Analyze() error = %v, want ErrInvalidCandidates", err)
	}
}

func TestPipelineRunEndToEnd(t *testing.T) {
	resp, err := NewPipeline().Run(Request{CandidateSet: pipelineSet("q1"), Access: access.Context{Clearance: access.ClearanceInternal}, AllowDegradedPlan: true})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if resp.Recommendation.QueryType != packer.QueryFactual {
		t.Fatalf("query type = %s, want Factual", resp.Recommendation.QueryType)
	}
	if len(resp.Ranked.Ranked) == 0 || len(resp.Plan.Slots) == 0 {
		t.Fatalf("ranked=%d slots=%d, want non-empty", len(resp.Ranked.Ranked), len(resp.Plan.Slots))
	}
	if resp.Plan.QueryID != "q1" {
		t.Fatalf("plan = %#v", resp.Plan)
	}
}

func TestPipelineAccessDeniedOnly(t *testing.T) {
	set := pipelineSet("q2")
	set.Nodes[0].Sensitivity = ragnode.SensitivityRestricted
	_, err := NewPipeline().Run(Request{CandidateSet: set, Access: access.Context{Clearance: access.ClearancePublic}})
	if !errors.Is(err, ErrAccessDeniedOnly) {
		t.Fatalf("Run() error = %v, want ErrAccessDeniedOnly", err)
	}
}

func TestPipelineRedCorrectiveSignal(t *testing.T) {
	set := pipelineSet("q3")
	set.Query = "Can the agent approve the refund?"
	set.Nodes = append(set.Nodes, ragnode.RAGNode{Item: citewise.Item{ID: "perm", Title: "Permission", Trust: 1}, Text: "permission", ChunkType: ragnode.ChunkPermissionRecord, Sensitivity: ragnode.SensitivityInternal, ApprovedBy: []string{"ops"}})
	set.Candidates = append(set.Candidates, ragnode.Candidate{NodeID: "perm", QueryRelevance: 0.9})
	set.Edges = append(set.Edges, ragnode.Edge{SourceID: "perm", TargetID: "n1", Type: ragnode.EdgeDecisionBasis, Confidence: 1})
	_, err := NewPipeline().Run(Request{CandidateSet: set, Access: access.Context{Clearance: access.ClearanceInternal, TrustedApprovers: []string{"ops"}}})
	if !errors.Is(err, ErrRedCorrectiveSignal) {
		t.Fatalf("Run() error = %v, want ErrRedCorrectiveSignal", err)
	}
}

func TestMetadata(t *testing.T) {
	analysis, err := Analyze(pipelineSet("q4"))
	if err != nil {
		t.Fatal(err)
	}
	metadata := Metadata(analysis)
	if metadata.MaxTopicSpan == 0 || metadata.ChunkTypeCounts[string(ragnode.ChunkDocument)] == 0 {
		t.Fatalf("metadata = %#v", metadata)
	}
}

func pipelineSet(queryID string) ragnode.CandidateSet {
	return ragnode.CandidateSet{
		QueryID: queryID,
		Query:   "refund policy",
		Nodes: []ragnode.RAGNode{
			{Item: citewise.Item{ID: "n1", Title: "Refund policy", Notes: "Policy text", Trust: 1, GoalFit: 1}, Text: "Refund policy text", ChunkType: ragnode.ChunkDocument, Sensitivity: ragnode.SensitivityInternal, Version: "v1", CommunityID: "refunds"},
		},
		Candidates: []ragnode.Candidate{{NodeID: "n1", QueryRelevance: 1}},
	}
}
