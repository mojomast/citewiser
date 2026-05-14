package ragnode

import (
	"time"

	"github.com/mojomast/citewiser/pkg/citewise"
)

type RAGAnalysis struct {
	QueryID      string
	Query        string
	Nodes        map[string]RAGNode
	Edges        []Edge
	EdgesOut     map[string][]Edge
	EdgesIn      map[string][]Edge
	Candidates   map[string]Candidate
	Roles        map[string]string
	Centrality   map[string]float64
	PPR          map[string]float64
	Duplicates   [][]string
	BaseAnalysis citewise.Analysis
	Now          time.Time
}
