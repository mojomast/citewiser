package main

import (
	"time"

	"github.com/mojomast/citewiseussy/pkg/access"
	"github.com/mojomast/citewiseussy/pkg/hygiene"
	"github.com/mojomast/citewiseussy/pkg/packer"
	"github.com/mojomast/citewiseussy/pkg/rag"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
	"github.com/mojomast/citewiseussy/pkg/ranker"
	"github.com/mojomast/citewiseussy/pkg/router"
)

type candidateRequest struct {
	QueryID           string              `json:"query_id"`
	Query             string              `json:"query"`
	TokenBudget       int                 `json:"token_budget,omitempty"`
	Access            access.Context      `json:"access"`
	QueryType         packer.QueryType    `json:"query_type,omitempty"`
	AllowDegradedPlan bool                `json:"allow_degraded_plan,omitempty"`
	Nodes             []ragnode.RAGNode   `json:"nodes"`
	Edges             []ragnode.Edge      `json:"edges,omitempty"`
	Candidates        []ragnode.Candidate `json:"candidates,omitempty"`
	GeneratedAt       time.Time           `json:"generated_at,omitempty"`
}

type routeRequest struct {
	Query    string               `json:"query"`
	Metadata router.GraphMetadata `json:"metadata"`
}

type healthResponse struct {
	Status             string               `json:"status"`
	GraphHygieneSignal packer.HygieneSignal `json:"graph_hygiene_signal"`
}

type rankResponse struct {
	Ranked ranker.RankedSet `json:"ranked"`
}

type packResponse struct {
	Plan packer.ContextPlan `json:"plan"`
}

type hygieneResponse struct {
	Hygiene hygiene.HygieneReport `json:"hygiene"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (r candidateRequest) candidateSet() ragnode.CandidateSet {
	return ragnode.CandidateSet{QueryID: r.QueryID, Query: r.Query, Nodes: r.Nodes, Edges: r.Edges, Candidates: r.Candidates, GeneratedAt: r.GeneratedAt}
}

func (r candidateRequest) ragRequest() rag.Request {
	return rag.Request{CandidateSet: r.candidateSet(), Access: r.Access, QueryType: r.QueryType, TokenBudget: r.TokenBudget, AllowDegradedPlan: r.AllowDegradedPlan}
}
