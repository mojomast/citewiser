package ranker

import (
	"sort"

	"github.com/mojomast/citewiseussy/pkg/access"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

// Score is the inspectable ranking breakdown for one candidate or suppression.
type Score struct {
	NodeID            string   `json:"node_id"`
	Total             float64  `json:"total"`
	QueryRelevance    float64  `json:"query_relevance"`
	GraphImportance   float64  `json:"graph_importance"`
	AuthorityScore    float64  `json:"authority_score"`
	Readiness         float64  `json:"readiness"`
	Freshness         float64  `json:"freshness"`
	TokenBudgetFit    float64  `json:"token_budget_fit"`
	BridgeBonus       float64  `json:"bridge_bonus"`
	CounterpointBonus float64  `json:"counterpoint_bonus"`
	DiversityBonus    float64  `json:"diversity_bonus"`
	RedundancyPenalty float64  `json:"redundancy_penalty"`
	StalenessPenalty  float64  `json:"staleness_penalty"`
	LowTrustPenalty   float64  `json:"low_trust_penalty"`
	Role              string   `json:"role"`
	AccessAllowed     bool     `json:"access_allowed"`
	SuppressionReason string   `json:"suppression_reason,omitempty"`
	Rationale         []string `json:"rationale"`
}

// RankedNode pairs an allowed node with the upstream candidate and score.
type RankedNode struct {
	Node      ragnode.RAGNode   `json:"node"`
	Candidate ragnode.Candidate `json:"candidate"`
	Score     Score             `json:"score"`
}

// RankedSet is the deterministic output of ranking after access gating.
type RankedSet struct {
	QueryID    string       `json:"query_id"`
	Query      string       `json:"query"`
	Ranked     []RankedNode `json:"ranked"`
	Suppressed []Score      `json:"suppressed"`
}

// Ranker ranks a prepared analysis for a caller and token budget.
type Ranker interface {
	Rank(ctx access.Context, analysis ragnode.RAGAnalysis, tokenBudget int) (RankedSet, error)
}

// AuthorityScorer scores node authority within an analysis.
type AuthorityScorer interface {
	Authority(node ragnode.RAGNode, analysis ragnode.RAGAnalysis) float64
}

// GraphScorer provides graph-importance inputs for ranking.
type GraphScorer interface {
	GlobalCentrality(analysis ragnode.RAGAnalysis) map[string]float64
	PersonalizedPageRank(analysis ragnode.RAGAnalysis, seeds map[string]float64) map[string]float64
}

// DiversityScorer scores source/community diversity against selected context.
type DiversityScorer interface {
	Diversity(node ragnode.RAGNode, selected []ragnode.RAGNode) float64
}

// TokenBudgetScorer scores fit against a token budget.
type TokenBudgetScorer interface {
	TokenBudgetFit(node ragnode.RAGNode, tokenBudget int) float64
}

// DefaultRanker composes the deterministic ranker helpers with access gates.
type DefaultRanker struct {
	Access access.Controller
}

// NewRanker returns a ranker using the default access controller.
func NewRanker() DefaultRanker {
	return DefaultRanker{Access: access.NewController()}
}

// Rank applies access gating first, scores allowed candidates, and returns a
// stable order by total descending then node ID ascending.
func (r DefaultRanker) Rank(ctx access.Context, analysis ragnode.RAGAnalysis, tokenBudget int) (RankedSet, error) {
	controller := r.Access
	if controller == nil {
		controller = access.NewController()
	}
	analysis.PPR = PersonalizedPageRank(analysis)
	set := RankedSet{QueryID: analysis.QueryID, Query: analysis.Query}
	selected := []ragnode.RAGNode{}

	for _, id := range sortedRankNodeIDs(analysis) {
		node := analysis.Nodes[id]
		candidate := analysis.Candidates[id]
		if candidate.NodeID == "" {
			candidate.NodeID = id
		}
		decision := controller.CanSeeNode(ctx, node)
		if !decision.Allowed {
			set.Suppressed = append(set.Suppressed, accessSuppression(id, decision))
			continue
		}
		score := scoreNode(ctx, node, candidate, analysis, tokenBudget, selected)
		set.Ranked = append(set.Ranked, RankedNode{Node: node, Candidate: candidate, Score: score})
		selected = append(selected, node)
	}

	sort.SliceStable(set.Ranked, func(i, j int) bool {
		if set.Ranked[i].Score.Total != set.Ranked[j].Score.Total {
			return set.Ranked[i].Score.Total > set.Ranked[j].Score.Total
		}
		return set.Ranked[i].Score.NodeID < set.Ranked[j].Score.NodeID
	})
	sort.SliceStable(set.Suppressed, func(i, j int) bool {
		return set.Suppressed[i].NodeID < set.Suppressed[j].NodeID
	})
	return set, nil
}

func sortedRankNodeIDs(analysis ragnode.RAGAnalysis) []string {
	if len(analysis.Candidates) == 0 {
		return sortedAnalysisNodeIDs(analysis)
	}
	ids := make([]string, 0, len(analysis.Candidates))
	for id := range analysis.Candidates {
		if _, ok := analysis.Nodes[id]; ok {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

func accessSuppression(nodeID string, decision access.Decision) Score {
	rationale := []string{decision.Detail}
	if decision.Detail == "" {
		rationale = nil
	}
	return Score{NodeID: nodeID, AccessAllowed: false, SuppressionReason: decision.Reason, Rationale: rationale}
}
