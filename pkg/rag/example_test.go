package rag_test

import (
	"fmt"

	"github.com/mojomast/citewiser/pkg/access"
	"github.com/mojomast/citewiser/pkg/citewise"
	"github.com/mojomast/citewiser/pkg/packer"
	"github.com/mojomast/citewiser/pkg/rag"
	"github.com/mojomast/citewiser/pkg/ragnode"
)

func ExampleAnalyze() {
	analysis, err := rag.Analyze(exampleCandidateSet())
	if err != nil {
		panic(err)
	}
	fmt.Println(analysis.QueryID, len(analysis.Nodes))
	// Output: q-example 1
}

func ExampleDefaultRanker() {
	analysis, err := rag.Analyze(exampleCandidateSet())
	if err != nil {
		panic(err)
	}
	ranked, err := rag.DefaultRanker().Rank(access.Context{Clearance: access.ClearanceInternal}, analysis, 1200)
	if err != nil {
		panic(err)
	}
	fmt.Println(ranked.QueryID, len(ranked.Ranked))
	// Output: q-example 1
}

func ExampleDefaultPacker() {
	analysis, err := rag.Analyze(exampleCandidateSet())
	if err != nil {
		panic(err)
	}
	plan := rag.DefaultPacker().Pack(analysis, packer.QueryFactual, 1200, string(access.ClearanceInternal))
	fmt.Println(plan.QueryID, len(plan.Slots))
	// Output: q-example 1
}

func exampleCandidateSet() ragnode.CandidateSet {
	return ragnode.CandidateSet{
		QueryID: "q-example",
		Query:   "refund policy",
		Nodes: []ragnode.RAGNode{{
			Item:        citewise.Item{ID: "policy", Title: "Refund Policy", Notes: "Policy text", Trust: 1, GoalFit: 1},
			Text:        "Policy text",
			ChunkType:   ragnode.ChunkDocument,
			Sensitivity: ragnode.SensitivityInternal,
			Version:     "v1",
		}},
		Candidates: []ragnode.Candidate{{NodeID: "policy", QueryRelevance: 1}},
	}
}
