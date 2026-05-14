package ranker

import (
	"fmt"

	"github.com/mojomast/citewiseussy/pkg/access"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

func rationaleForScore(ctx access.Context, node ragnode.RAGNode, score Score, analysis ragnode.RAGAnalysis) []string {
	rationale := []string{
		fmt.Sprintf("query relevance %.2f", score.QueryRelevance),
		fmt.Sprintf("authority %.2f", score.AuthorityScore),
		fmt.Sprintf("graph importance %.2f", score.GraphImportance),
		fmt.Sprintf("token budget fit %.2f", score.TokenBudgetFit),
	}
	if node.UsesEstimatedTokenCount() {
		rationale = append(rationale, fmt.Sprintf("token count estimated as %d", node.EffectiveTokenCount()))
	}
	if hasUnknownEdgeFallback(node.ID, analysis) {
		rationale = append(rationale, "unknown edge type weighted as related-to")
	}
	if score.RedundancyPenalty > 0 {
		rationale = append(rationale, fmt.Sprintf("redundancy penalty %.2f", score.RedundancyPenalty))
	}
	if score.StalenessPenalty > 0 {
		rationale = append(rationale, fmt.Sprintf("staleness penalty %.2f", score.StalenessPenalty))
	}
	if score.LowTrustPenalty > 0 {
		rationale = append(rationale, fmt.Sprintf("low trust penalty %.2f", score.LowTrustPenalty))
	}
	if queryType(ctx) != "" {
		rationale = append(rationale, fmt.Sprintf("query type %s modifier applied", queryType(ctx)))
	}
	return rationale
}
