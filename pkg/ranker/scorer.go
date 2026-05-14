package ranker

import (
	"math"
	"strings"

	"github.com/mojomast/citewiser/pkg/access"
	"github.com/mojomast/citewiser/pkg/citewise"
	"github.com/mojomast/citewiser/pkg/ragnode"
)

func scoreNode(ctx access.Context, node ragnode.RAGNode, candidate ragnode.Candidate, analysis ragnode.RAGAnalysis, tokenBudget int, selected []ragnode.RAGNode) Score {
	role := analysis.Roles[node.ID]
	base := analysis.BaseAnalysis.Scores[node.ID]
	score := Score{
		NodeID:            node.ID,
		QueryRelevance:    clamp01(candidate.QueryRelevance),
		GraphImportance:   graphImportance(node.ID, analysis),
		AuthorityScore:    Authority(node, analysis),
		Readiness:         readiness(base),
		Freshness:         freshness(base, node),
		TokenBudgetFit:    TokenBudgetFit(node, analysis, tokenBudget),
		BridgeBonus:       bridgeBonus(role, node, analysis),
		CounterpointBonus: counterpointBonus(role, node, analysis),
		DiversityBonus:    Diversity(node, selected),
		RedundancyPenalty: redundancyPenalty(node.ID, role, analysis),
		StalenessPenalty:  stalenessPenalty(node),
		LowTrustPenalty:   lowTrustPenalty(node),
		Role:              role,
		AccessAllowed:     true,
	}
	score.Total = totalScore(ctx, node, score)
	score.Rationale = rationaleForScore(ctx, node, score, analysis)
	return score
}

func graphImportance(nodeID string, analysis ragnode.RAGAnalysis) float64 {
	return clamp01(0.55*analysis.Centrality[nodeID] + 0.45*analysis.PPR[nodeID])
}

func readiness(base citewise.Score) float64 {
	if base.ItemID == "" {
		return 1
	}
	return clamp01(base.Readiness)
}

func freshness(base citewise.Score, node ragnode.RAGNode) float64 {
	if node.SupersededBy != "" {
		return 0
	}
	if base.ItemID != "" {
		return clamp01(base.Freshness)
	}
	return versionCurrentness(node)
}

func bridgeBonus(role string, node ragnode.RAGNode, analysis ragnode.RAGAnalysis) float64 {
	if role == citewise.RoleBridge {
		return 1
	}
	for _, edge := range append(analysis.EdgesIn[node.ID], analysis.EdgesOut[node.ID]...) {
		if edge.Type == ragnode.EdgeCommunityMemberOf || edge.Type == ragnode.EdgePrerequisite || edge.Type == ragnode.EdgeAppliesTo {
			return 0.5
		}
	}
	return 0
}

func counterpointBonus(role string, node ragnode.RAGNode, analysis ragnode.RAGAnalysis) float64 {
	if role == citewise.RoleCounterpoint {
		return 1
	}
	for _, edge := range append(analysis.EdgesIn[node.ID], analysis.EdgesOut[node.ID]...) {
		if edge.Type == ragnode.EdgeEvidenceAgainst || edge.Type == ragnode.EdgeContradicts {
			return 0.5
		}
	}
	return 0
}

func redundancyPenalty(nodeID, role string, analysis ragnode.RAGAnalysis) float64 {
	if role == citewise.RoleDuplicate {
		return 1
	}
	for _, cluster := range analysis.Duplicates {
		for _, id := range cluster {
			if id == nodeID {
				return 1
			}
		}
	}
	return 0
}

func stalenessPenalty(node ragnode.RAGNode) float64 {
	if node.SupersededBy != "" {
		return 1
	}
	return 0
}

func lowTrustPenalty(node ragnode.RAGNode) float64 {
	if node.Trust >= 0.35 {
		return 0
	}
	return clamp01(1 - node.Trust/0.35)
}

func totalScore(ctx access.Context, node ragnode.RAGNode, score Score) float64 {
	positive := 0.28*score.QueryRelevance +
		0.18*score.AuthorityScore +
		0.14*score.GraphImportance +
		0.10*score.Readiness +
		0.10*score.Freshness +
		0.10*score.TokenBudgetFit +
		0.05*score.DiversityBonus +
		0.03*score.BridgeBonus +
		0.02*score.CounterpointBonus
	negative := 0.14*score.RedundancyPenalty +
		0.10*score.StalenessPenalty +
		0.08*score.LowTrustPenalty
	if queryType(ctx) == "Temporal" {
		positive += 0.08 * score.Freshness
	}
	if queryType(ctx) == "Adversarial" {
		negative += 0.08 * score.LowTrustPenalty
	}
	points := clamp01(positive-negative)*100 + queryTypeModifier(ctx, node, score)
	return round1(clamp01(points/100) * 100)
}

func queryType(ctx access.Context) string {
	if ctx.Attributes == nil {
		return ""
	}
	return ctx.Attributes["query_type"]
}

func queryTypeModifier(ctx access.Context, node ragnode.RAGNode, score Score) float64 {
	switch queryType(ctx) {
	case "Factual":
		if score.Role == citewise.RoleFoundation {
			return 8
		}
	case "Comparative":
		if score.Role == citewise.RoleBridge || score.Role == citewise.RoleCounterpoint {
			return 8
		}
	case "Procedural":
		return chunkTypeModifier(node, ragnode.ChunkProcedure, 12, ragnode.ChunkPermissionRecord, 6)
	case "Exploratory":
		if score.Role == citewise.RoleOverview {
			return 12
		}
		if score.Role == citewise.RoleBridge {
			return 8
		}
	case "Adversarial":
		if score.Role == citewise.RoleCounterpoint {
			return 15
		}
	case "Agentic":
		return chunkTypeModifier(node, ragnode.ChunkPermissionRecord, 20, ragnode.ChunkDecision, 15, ragnode.ChunkProcedure, 12)
	}
	return 0
}

func chunkTypeModifier(node ragnode.RAGNode, pairs ...any) float64 {
	for i := 0; i+1 < len(pairs); i += 2 {
		chunkType := pairs[i].(ragnode.ChunkType)
		if node.ChunkType == chunkType {
			switch modifier := pairs[i+1].(type) {
			case int:
				return float64(modifier)
			case float64:
				return modifier
			}
		}
	}
	return 0
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func hasUnknownEdgeFallback(nodeID string, analysis ragnode.RAGAnalysis) bool {
	for _, edge := range append(analysis.EdgesIn[nodeID], analysis.EdgesOut[nodeID]...) {
		if strings.Contains(edge.Note, "raw_type=") && edge.Type == ragnode.EdgeRelatedTo {
			return true
		}
	}
	return false
}
