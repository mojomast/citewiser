package ragnode

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/mojomast/citewiser/pkg/citewise"
)

// ToItem returns the embedded Citewise item, using Text as Notes when the
// embedded item does not already carry notes.
func (n RAGNode) ToItem() citewise.Item {
	it := n.Item
	if it.Notes == "" {
		it.Notes = n.Text
	}
	return it
}

// RAGNodeFromItem converts an existing Citewise item using the migration
// defaults from spec section 12.3.
func RAGNodeFromItem(it citewise.Item) RAGNode {
	text := it.Notes
	return RAGNode{
		Item:        it,
		Text:        text,
		ChunkType:   ChunkDocument,
		TokenCount:  EstimateTokenCount(text),
		Sensitivity: SensitivityInternal,
	}
}

// CandidateSetFromBacklog adapts an existing Citewise backlog into a RAG
// candidate set without changing old CLI parsing behavior.
func CandidateSetFromBacklog(b citewise.Backlog) CandidateSet {
	nodes := make([]RAGNode, 0, len(b.Items))
	for _, it := range b.Items {
		nodes = append(nodes, RAGNodeFromItem(it))
	}
	edges := make([]Edge, 0, len(b.Edges))
	for _, e := range b.Edges {
		edges = append(edges, EdgeFromCitewiseEdge(e))
	}
	return CandidateSet{Nodes: nodes, Edges: edges}
}

// ParseCandidateSet reads new RAG JSON candidate-set input. Callers that need
// old backlog compatibility should parse with pkg/citewise first, then call
// CandidateSetFromBacklog.
func ParseCandidateSet(r io.Reader) (CandidateSet, error) {
	var set CandidateSet
	dec := json.NewDecoder(r)
	if err := dec.Decode(&set); err != nil {
		return CandidateSet{}, err
	}
	return set, nil
}

// ToCitewiseEdge normalizes the edge type and preserves confidence and note.
func (e Edge) ToCitewiseEdge() citewise.Edge {
	n := NormalizeEdgeValue(e)
	return citewise.Edge{
		SourceID:   n.SourceID,
		TargetID:   n.TargetID,
		Type:       n.Type,
		Confidence: n.Confidence,
		Note:       n.Note,
	}
}

// EdgeFromCitewiseEdge copies an existing Citewise edge into the RAG edge type.
func EdgeFromCitewiseEdge(e citewise.Edge) Edge {
	return Edge{
		SourceID:   e.SourceID,
		TargetID:   e.TargetID,
		Type:       e.Type,
		Confidence: e.Confidence,
		Note:       e.Note,
	}
}

// BuildAnalysis validates a candidate set and constructs deterministic maps and
// edge indexes for downstream ranking and packing.
func BuildAnalysis(set CandidateSet) (RAGAnalysis, error) {
	nodes := map[string]RAGNode{}
	for _, node := range set.Nodes {
		if strings.TrimSpace(node.ID) == "" {
			return RAGAnalysis{}, fmt.Errorf("ragnode: node is missing id")
		}
		if _, ok := nodes[node.ID]; ok {
			return RAGAnalysis{}, fmt.Errorf("ragnode: duplicate node id %q", node.ID)
		}
		nodes[node.ID] = node
	}

	candidates := map[string]Candidate{}
	for _, candidate := range set.Candidates {
		if strings.TrimSpace(candidate.NodeID) == "" {
			return RAGAnalysis{}, fmt.Errorf("ragnode: candidate is missing node_id")
		}
		if _, ok := nodes[candidate.NodeID]; !ok {
			return RAGAnalysis{}, fmt.Errorf("ragnode: candidate node_id %q does not match any node", candidate.NodeID)
		}
		candidates[candidate.NodeID] = candidate
	}

	edges := make([]Edge, 0, len(set.Edges))
	edgesOut := map[string][]Edge{}
	edgesIn := map[string][]Edge{}
	for _, edge := range set.Edges {
		if _, ok := nodes[edge.SourceID]; !ok {
			return RAGAnalysis{}, fmt.Errorf("ragnode: edge source_id %q does not match any node", edge.SourceID)
		}
		if _, ok := nodes[edge.TargetID]; !ok {
			return RAGAnalysis{}, fmt.Errorf("ragnode: edge target_id %q does not match any node", edge.TargetID)
		}
		normalized := NormalizeEdgeValue(edge)
		edges = append(edges, normalized)
		edgesOut[normalized.SourceID] = append(edgesOut[normalized.SourceID], normalized)
		edgesIn[normalized.TargetID] = append(edgesIn[normalized.TargetID], normalized)
	}
	sortEdges(edges)
	for id := range edgesOut {
		sortEdges(edgesOut[id])
	}
	for id := range edgesIn {
		sortEdges(edgesIn[id])
	}

	items := make([]citewise.Item, 0, len(set.Nodes))
	for _, id := range sortedNodeIDs(nodes) {
		items = append(items, nodes[id].ToItem())
	}
	citewiseEdges := make([]citewise.Edge, 0, len(edges))
	for _, edge := range edges {
		citewiseEdges = append(citewiseEdges, edge.ToCitewiseEdge())
	}
	base := citewise.Analyze(citewise.Backlog{Items: items, Edges: citewiseEdges}, citewise.Goal{Prompt: set.Query})

	return RAGAnalysis{
		QueryID:      set.QueryID,
		Query:        set.Query,
		Nodes:        nodes,
		Edges:        edges,
		EdgesOut:     edgesOut,
		EdgesIn:      edgesIn,
		Candidates:   candidates,
		Roles:        base.Roles,
		Centrality:   base.Centrality,
		PPR:          map[string]float64{},
		Duplicates:   base.Duplicates,
		BaseAnalysis: base,
		Now:          analysisTime(set.GeneratedAt),
	}, nil
}

func analysisTime(generatedAt time.Time) time.Time {
	if generatedAt.IsZero() {
		return time.Now().UTC()
	}
	return generatedAt
}

func sortedNodeIDs(nodes map[string]RAGNode) []string {
	ids := make([]string, 0, len(nodes))
	for id := range nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func sortEdges(edges []Edge) {
	sort.SliceStable(edges, func(i, j int) bool {
		if edges[i].SourceID != edges[j].SourceID {
			return edges[i].SourceID < edges[j].SourceID
		}
		if edges[i].TargetID != edges[j].TargetID {
			return edges[i].TargetID < edges[j].TargetID
		}
		if edges[i].Type != edges[j].Type {
			return edges[i].Type < edges[j].Type
		}
		return edges[i].Note < edges[j].Note
	})
}
