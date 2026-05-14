package ranker

import (
	"math"

	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

// TokenBudgetFit returns the bounded fit score for a node under tokenBudget,
// normalizing token density across the supplied analysis.
func TokenBudgetFit(node ragnode.RAGNode, analysis ragnode.RAGAnalysis, tokenBudget int) float64 {
	return tokenBudgetFitWithDensity(node, tokenBudget, normalizedTokenDensity(node, analysis))
}

// TokenBudgetFitSingle returns token fit for callers without an analysis. The
// node's density is clamped directly, which is equivalent to a safe bounded
// single-node normalization fallback.
func TokenBudgetFitSingle(node ragnode.RAGNode, tokenBudget int) float64 {
	return tokenBudgetFitWithDensity(node, tokenBudget, clamp01(tokenDensity(node, ragnode.RAGAnalysis{})))
}

func tokenBudgetFitWithDensity(node ragnode.RAGNode, tokenBudget int, normalizedDensity float64) float64 {
	return clamp01(0.60*lengthFit(node.EffectiveTokenCount(), tokenBudget) + 0.40*normalizedDensity)
}

func normalizedTokenDensity(node ragnode.RAGNode, analysis ragnode.RAGAnalysis) float64 {
	maxDensity := 0.0
	for _, candidate := range analysis.Nodes {
		density := tokenDensity(candidate, analysis)
		if density > maxDensity {
			maxDensity = density
		}
	}
	if maxDensity == 0 {
		return 0
	}
	return tokenDensity(node, analysis) / maxDensity
}

func tokenDensity(node ragnode.RAGNode, analysis ragnode.RAGAnalysis) float64 {
	tokens := node.EffectiveTokenCount()
	denominator := math.Max(1, math.Log2(float64(tokens+2)))
	return Authority(node, analysis) / denominator
}

func lengthFit(tokens, tokenBudget int) float64 {
	if tokenBudget <= 0 {
		return 0.25
	}
	fraction := float64(tokens) / float64(tokenBudget)
	switch {
	case fraction <= 0.08:
		return 1.00
	case fraction <= 0.15:
		return 0.80
	case fraction <= 0.25:
		return 0.55
	default:
		return 0.25
	}
}
