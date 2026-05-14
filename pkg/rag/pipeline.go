package rag

import (
	"fmt"
	"sort"

	"github.com/mojomast/citewiseussy/pkg/access"
	"github.com/mojomast/citewiseussy/pkg/hygiene"
	"github.com/mojomast/citewiseussy/pkg/memory"
	"github.com/mojomast/citewiseussy/pkg/packer"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
	"github.com/mojomast/citewiseussy/pkg/ranker"
	"github.com/mojomast/citewiseussy/pkg/router"
)

// NewPipeline returns the default in-process CitewiseRAG pipeline.
func NewPipeline() Pipeline {
	return Pipeline{
		Ranker:  DefaultRanker(),
		Packer:  DefaultPacker(),
		Router:  DefaultRouter(),
		Hygiene: DefaultHygieneAnalyzer(),
		Memory:  DefaultMemoryStore(memory.DefaultPath),
	}
}

// Analyze validates candidates and builds deterministic graph indexes.
func Analyze(set ragnode.CandidateSet) (ragnode.RAGAnalysis, error) {
	analysis, err := ragnode.BuildAnalysis(set)
	if err != nil {
		return ragnode.RAGAnalysis{}, fmt.Errorf("%w: %v", ErrInvalidCandidates, err)
	}
	return analysis, nil
}

// DefaultRanker returns the default access-gated ranker.
func DefaultRanker() ranker.Ranker { return ranker.NewRanker() }

// DefaultPacker returns the default context packer.
func DefaultPacker() packer.Packer { return packer.NewPacker() }

// DefaultRouter returns the deterministic query router.
func DefaultRouter() router.QueryRouter { return router.NewRouter() }

// DefaultHygieneAnalyzer returns the deterministic graph hygiene analyzer.
func DefaultHygieneAnalyzer() interface {
	Analyze(ragnode.RAGAnalysis, bool) hygiene.HygieneReport
} {
	return hygiene.NewAnalyzer()
}

// DefaultMemoryStore returns a file-backed memory store at path.
func DefaultMemoryStore(path string) memory.MemoryWriteBack { return &memory.FileStore{Path: path} }

// Run executes validate -> analyze -> access/rank -> classify -> hygiene -> pack.
func (p Pipeline) Run(req Request) (Response, error) {
	analysis, err := Analyze(req.CandidateSet)
	if err != nil {
		return Response{}, err
	}
	r := p.Ranker
	if r == nil {
		r = DefaultRanker()
	}
	ranked, err := r.Rank(req.Access, analysis, req.TokenBudget)
	if err != nil {
		return Response{}, err
	}
	if accessDeniedOnly(ranked) {
		return Response{Ranked: ranked}, ErrAccessDeniedOnly
	}

	rec := router.Recommendation{QueryType: req.QueryType, ContextBudgetHint: req.TokenBudget}
	if rec.QueryType == "" {
		routerComponent := p.Router
		if routerComponent == nil {
			routerComponent = DefaultRouter()
		}
		rec = routerComponent.Route(analysis.Query, Metadata(analysis))
	}
	budget := req.TokenBudget
	if budget <= 0 {
		budget = rec.ContextBudgetHint
	}
	if budget <= 0 {
		budget = 4000
	}

	analyzer := p.Hygiene
	if analyzer == nil {
		analyzer = DefaultHygieneAnalyzer()
	}
	hygieneReport := analyzer.Analyze(analysis, req.AllowDegradedPlan)
	packComponent := p.Packer
	if packComponent == nil {
		packComponent = DefaultPacker()
	}
	plan := packComponent.Pack(analysis, rec.QueryType, budget, string(req.Access.Clearance))
	if plan.HygieneSignal == packer.HygieneGreen && hygieneReport.Signal != packer.HygieneGreen {
		plan.HygieneSignal = hygieneReport.Signal
	}
	resp := Response{Recommendation: rec, Ranked: ranked, Hygiene: hygieneReport, Plan: plan}
	if plan.HygieneSignal == packer.HygieneRed && !req.AllowDegradedPlan {
		return resp, ErrRedCorrectiveSignal
	}
	return resp, nil
}

// Metadata summarizes an analysis for query routing.
func Metadata(analysis ragnode.RAGAnalysis) router.GraphMetadata {
	metadata := router.GraphMetadata{RoleCounts: map[string]int{}, ChunkTypeCounts: map[string]int{}}
	topicNodes := map[string]bool{}
	communityIDs := map[string]bool{}
	for id, node := range analysis.Nodes {
		if node.ChunkType == ragnode.ChunkEntity {
			metadata.EntityIDs = append(metadata.EntityIDs, id)
		}
		if node.CommunityID != "" {
			communityIDs[node.CommunityID] = true
			topicNodes[node.CommunityID] = true
		}
		if node.SemanticType != "" {
			topicNodes[node.SemanticType] = true
		}
		if node.ChunkType == ragnode.ChunkPermissionRecord {
			metadata.HasPermissionNode = true
		}
		metadata.ChunkTypeCounts[string(node.ChunkType)]++
	}
	for _, edge := range analysis.Edges {
		if edge.Type == ragnode.EdgeDecisionBasis {
			metadata.HasDecisionBasis = true
		}
	}
	for _, role := range analysis.Roles {
		metadata.RoleCounts[role]++
	}
	metadata.Topics = sortedSet(topicNodes)
	metadata.CommunityIDs = sortedSet(communityIDs)
	metadata.MaxTopicSpan = len(metadata.Topics)
	sort.Strings(metadata.EntityIDs)
	return metadata
}

func accessDeniedOnly(set ranker.RankedSet) bool {
	if len(set.Ranked) > 0 || len(set.Suppressed) == 0 {
		return false
	}
	for _, suppressed := range set.Suppressed {
		if suppressed.SuppressionReason != access.ReasonAccessControl {
			return false
		}
	}
	return true
}

func sortedSet(set map[string]bool) []string {
	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}
