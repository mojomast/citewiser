package hygiene

import (
	"sort"

	"github.com/mojomast/citewiser/pkg/citewise"
	"github.com/mojomast/citewiser/pkg/packer"
	"github.com/mojomast/citewiser/pkg/ragnode"
)

const DefaultThreshold = 0.55

type EdgeSuggestion struct {
	SourceID   string  `json:"source_id"`
	TargetID   string  `json:"target_id"`
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

type HygieneReport struct {
	DuplicateClusters [][]string           `json:"duplicate_clusters"`
	OrphanNodes       []string             `json:"orphan_nodes"`
	StaleNodes        []string             `json:"stale_nodes"`
	MissingBridges    []string             `json:"missing_bridges"`
	MissingEdges      []EdgeSuggestion     `json:"missing_edges"`
	Score             float64              `json:"score"`
	Signal            packer.HygieneSignal `json:"signal"`
	RetrievalTargets  []string             `json:"retrieval_targets,omitempty"`
}

type Analyzer interface {
	SuggestMissingEdges(analysis ragnode.RAGAnalysis) []EdgeSuggestion
	HygieneScore(analysis ragnode.RAGAnalysis) float64
	CorrectiveSignal(analysis ragnode.RAGAnalysis, threshold float64) packer.HygieneSignal
}

type DefaultAnalyzer struct{}

func NewAnalyzer() DefaultAnalyzer { return DefaultAnalyzer{} }

func (a DefaultAnalyzer) Analyze(analysis ragnode.RAGAnalysis, allowDegradedPlan bool) HygieneReport {
	report := HygieneReport{
		DuplicateClusters: cloneClusters(analysis.Duplicates),
		OrphanNodes:       orphanNodes(analysis),
		StaleNodes:        staleNodes(analysis),
		MissingBridges:    missingBridges(analysis),
		MissingEdges:      a.SuggestMissingEdges(analysis),
	}
	report.Score = a.HygieneScore(analysis)
	report.Signal = signalForScore(report.Score, DefaultThreshold)
	if report.Signal == packer.HygieneRed && !allowDegradedPlan {
		report.RetrievalTargets = retrievalTargets(report)
	}
	return report
}

func (DefaultAnalyzer) HygieneScore(analysis ragnode.RAGAnalysis) float64 {
	n := float64(len(analysis.Nodes))
	if n == 0 {
		return 1
	}
	return clamp01(1.00 -
		0.18*ratio(float64(len(orphanNodes(analysis))), n) -
		0.18*ratio(float64(duplicateNodeCount(analysis.Duplicates)), n) -
		0.20*ratio(float64(len(staleNodes(analysis))), n) -
		0.20*ratio(float64(len(missingBridges(analysis))), n) -
		0.14*ratio(float64(lowTrustCount(analysis)), n) -
		0.10*ratio(float64(unapprovedAgenticCount(analysis)), n))
}

func (a DefaultAnalyzer) CorrectiveSignal(analysis ragnode.RAGAnalysis, threshold float64) packer.HygieneSignal {
	if threshold <= 0 {
		threshold = DefaultThreshold
	}
	return signalForScore(a.HygieneScore(analysis), threshold)
}

func orphanNodes(analysis ragnode.RAGAnalysis) []string {
	if len(analysis.BaseAnalysis.Items) > 0 {
		h := citewise.Hygiene(analysis.BaseAnalysis)
		out := append([]string(nil), h.OrphanItems...)
		sort.Strings(out)
		return out
	}
	out := []string{}
	for id := range analysis.Nodes {
		if len(analysis.EdgesIn[id]) == 0 && len(analysis.EdgesOut[id]) == 0 {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

func staleNodes(analysis ragnode.RAGAnalysis) []string {
	seen := map[string]bool{}
	if len(analysis.BaseAnalysis.Items) > 0 {
		for _, id := range citewise.Hygiene(analysis.BaseAnalysis).StaleItems {
			seen[id] = true
		}
	}
	for id, node := range analysis.Nodes {
		if node.SupersededBy != "" {
			seen[id] = true
		}
	}
	return sortedKeys(seen)
}

func missingBridges(analysis ragnode.RAGAnalysis) []string {
	if len(analysis.BaseAnalysis.Items) > 0 {
		out := append([]string(nil), citewise.Hygiene(analysis.BaseAnalysis).MissingBridges...)
		sort.Strings(out)
		return out
	}
	seen := map[string]bool{}
	for id := range analysis.Nodes {
		hasPrereq := false
		hasBridge := false
		for _, edge := range append(analysis.EdgesIn[id], analysis.EdgesOut[id]...) {
			if edge.Type == ragnode.EdgePrerequisite {
				hasPrereq = true
			}
			other := edge.SourceID
			if other == id {
				other = edge.TargetID
			}
			if analysis.Roles[other] == citewise.RoleOverview || analysis.Roles[other] == citewise.RoleBridge {
				hasBridge = true
			}
		}
		if hasPrereq && !hasBridge {
			seen[id] = true
		}
	}
	return sortedKeys(seen)
}

func duplicateNodeCount(clusters [][]string) int {
	seen := map[string]bool{}
	for _, cluster := range clusters {
		for _, id := range cluster {
			seen[id] = true
		}
	}
	return len(seen)
}

func lowTrustCount(analysis ragnode.RAGAnalysis) int {
	count := 0
	for _, node := range analysis.Nodes {
		if node.Trust < 0.35 {
			count++
		}
	}
	return count
}

func unapprovedAgenticCount(analysis ragnode.RAGAnalysis) int {
	count := 0
	for _, node := range analysis.Nodes {
		if len(node.ApprovedBy) == 0 && (node.ChunkType == ragnode.ChunkPermissionRecord || node.ChunkType == ragnode.ChunkDecision || node.ChunkType == ragnode.ChunkProcedure) {
			count++
		}
	}
	return count
}

func retrievalTargets(report HygieneReport) []string {
	seen := map[string]bool{}
	for _, id := range report.MissingBridges {
		seen[id] = true
	}
	for _, suggestion := range report.MissingEdges {
		seen[suggestion.SourceID] = true
		seen[suggestion.TargetID] = true
	}
	return sortedKeys(seen)
}

func cloneClusters(clusters [][]string) [][]string {
	out := make([][]string, len(clusters))
	for i, cluster := range clusters {
		out[i] = append([]string(nil), cluster...)
		sort.Strings(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool {
		if len(out[i]) == 0 || len(out[j]) == 0 {
			return len(out[i]) < len(out[j])
		}
		return out[i][0] < out[j][0]
	})
	return out
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func ratio(v, total float64) float64 {
	if total == 0 {
		return 0
	}
	return v / total
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
