package rag

import (
	"errors"

	"github.com/mojomast/citewiser/pkg/access"
	"github.com/mojomast/citewiser/pkg/hygiene"
	"github.com/mojomast/citewiser/pkg/memory"
	"github.com/mojomast/citewiser/pkg/packer"
	"github.com/mojomast/citewiser/pkg/ragnode"
	"github.com/mojomast/citewiser/pkg/ranker"
	"github.com/mojomast/citewiser/pkg/router"
)

var (
	ErrInvalidCandidates   = errors.New("rag: invalid candidates")
	ErrAccessDeniedOnly    = errors.New("rag: all candidates suppressed by access control")
	ErrRedCorrectiveSignal = errors.New("rag: red corrective signal")
)

// Pipeline contains the default CitewiseRAG components.
type Pipeline struct {
	Ranker  ranker.Ranker
	Packer  packer.Packer
	Router  router.QueryRouter
	Hygiene interface {
		Analyze(analysis ragnode.RAGAnalysis, allowDegradedPlan bool) hygiene.HygieneReport
	}
	Memory memory.MemoryWriteBack
}

// Request is the library orchestration input.
type Request struct {
	CandidateSet      ragnode.CandidateSet `json:"candidate_set"`
	Access            access.Context       `json:"access"`
	QueryType         packer.QueryType     `json:"query_type,omitempty"`
	TokenBudget       int                  `json:"token_budget,omitempty"`
	AllowDegradedPlan bool                 `json:"allow_degraded_plan,omitempty"`
}

// Response is the library orchestration output.
type Response struct {
	Recommendation router.Recommendation `json:"recommendation"`
	Ranked         ranker.RankedSet      `json:"ranked"`
	Hygiene        hygiene.HygieneReport `json:"hygiene"`
	Plan           packer.ContextPlan    `json:"plan"`
}
