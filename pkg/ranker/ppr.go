package ranker

import (
	"math"
	"sort"

	"github.com/mojomast/citewiser/pkg/ragnode"
)

const (
	defaultPPRAlpha         = 0.15
	defaultPPRMaxIterations = 100
	defaultPPRTolerance     = 1e-8
)

// PersonalizedPageRank computes query-personalized graph importance using
// candidate QueryRelevance as restart seeds and ontology edge weights as
// transition weights.
func PersonalizedPageRank(analysis ragnode.RAGAnalysis) map[string]float64 {
	return personalizedPageRank(analysis, defaultPPRAlpha, defaultPPRMaxIterations, defaultPPRTolerance)
}

func personalizedPageRank(analysis ragnode.RAGAnalysis, alpha float64, maxIterations int, tolerance float64) map[string]float64 {
	ids := sortedAnalysisNodeIDs(analysis)
	scores := zeroScores(ids)
	if len(ids) == 0 || usableEdgeCount(analysis) < 2 {
		return scores
	}
	seeds := normalizedSeeds(ids, analysis)
	if seedTotal(seeds) == 0 {
		uniform := 1 / float64(len(ids))
		for _, id := range ids {
			seeds[id] = uniform
		}
	}
	for _, id := range ids {
		scores[id] = seeds[id]
	}

	outWeights := outgoingWeights(analysis)
	for iteration := 0; iteration < maxIterations; iteration++ {
		next := map[string]float64{}
		for _, id := range ids {
			next[id] = alpha * seeds[id]
		}

		danglingMass := 0.0
		for _, id := range ids {
			if outWeights[id] == 0 {
				danglingMass += scores[id]
				continue
			}
			for _, edge := range analysis.EdgesOut[id] {
				if !usablePPRNode(edge.TargetID, analysis) {
					continue
				}
				weight := pprEdgeWeight(edge)
				if weight == 0 {
					continue
				}
				next[edge.TargetID] += (1 - alpha) * scores[id] * weight / outWeights[id]
			}
		}

		for _, id := range ids {
			next[id] += (1 - alpha) * danglingMass * seeds[id]
		}
		if pprDelta(ids, scores, next) < tolerance {
			return next
		}
		scores = next
	}
	return scores
}

func sortedAnalysisNodeIDs(analysis ragnode.RAGAnalysis) []string {
	ids := make([]string, 0, len(analysis.Nodes))
	for id := range analysis.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func zeroScores(ids []string) map[string]float64 {
	scores := make(map[string]float64, len(ids))
	for _, id := range ids {
		scores[id] = 0
	}
	return scores
}

func normalizedSeeds(ids []string, analysis ragnode.RAGAnalysis) map[string]float64 {
	seeds := make(map[string]float64, len(ids))
	total := 0.0
	for _, id := range ids {
		candidate := analysis.Candidates[id]
		seed := clamp01(candidate.QueryRelevance)
		seeds[id] = seed
		total += seed
	}
	if total == 0 {
		return seeds
	}
	for _, id := range ids {
		seeds[id] /= total
	}
	return seeds
}

func seedTotal(seeds map[string]float64) float64 {
	total := 0.0
	for _, seed := range seeds {
		total += seed
	}
	return total
}

func outgoingWeights(analysis ragnode.RAGAnalysis) map[string]float64 {
	weights := map[string]float64{}
	for _, edge := range analysis.Edges {
		if !usablePPRNode(edge.SourceID, analysis) || !usablePPRNode(edge.TargetID, analysis) {
			continue
		}
		weights[edge.SourceID] += pprEdgeWeight(edge)
	}
	return weights
}

func usableEdgeCount(analysis ragnode.RAGAnalysis) int {
	count := 0
	for _, edge := range analysis.Edges {
		if usablePPRNode(edge.SourceID, analysis) && usablePPRNode(edge.TargetID, analysis) && pprEdgeWeight(edge) > 0 {
			count++
		}
	}
	return count
}

func usablePPRNode(id string, analysis ragnode.RAGAnalysis) bool {
	_, ok := analysis.Nodes[id]
	return ok
}

func pprEdgeWeight(edge ragnode.Edge) float64 {
	weight := ragnode.EdgeWeight(edge.Type)
	if edge.Confidence > 0 {
		weight *= edge.Confidence
	}
	return weight
}

func pprDelta(ids []string, a, b map[string]float64) float64 {
	delta := 0.0
	for _, id := range ids {
		delta += math.Abs(a[id] - b[id])
	}
	return delta
}
