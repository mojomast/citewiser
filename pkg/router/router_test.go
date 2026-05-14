package router

import (
	"testing"

	"github.com/mojomast/citewiser/pkg/packer"
)

func TestRouteDecisionBranches(t *testing.T) {
	cases := []struct {
		name      string
		query     string
		metadata  GraphMetadata
		queryType packer.QueryType
		mode      RetrievalMode
		budget    int
		counter   bool
	}{
		{"agentic", "approve refund for account", GraphMetadata{HasPermissionNode: true}, packer.QueryAgentic, ModeDRIFTChain, 6000, false},
		{"temporal", "latest vendor access version", GraphMetadata{}, packer.QueryTemporal, ModeHybridBM25Dense, 4500, false},
		{"procedural", "how do I troubleshoot login", GraphMetadata{}, packer.QueryProcedural, ModeLocalNeighborhood, 5000, false},
		{"comparative", "compare postgres versus mysql", GraphMetadata{}, packer.QueryComparative, ModeDRIFTChain, 5500, false},
		{"adversarial", "is it safe to grant access", GraphMetadata{}, packer.QueryAdversarial, ModeDRIFTChain, 6000, true},
		{"entity", "status for account A", GraphMetadata{EntityIDs: []string{"account A"}}, packer.QueryFactual, ModeLocalNeighborhood, 3500, false},
		{"exploratory", "what are the themes in refunds", GraphMetadata{}, packer.QueryExploratory, ModeGlobalGraph, 6500, false},
		{"topic span", "explain refunds", GraphMetadata{MaxTopicSpan: 4}, packer.QueryExploratory, ModeGlobalGraph, 6500, false},
		{"short", "refund policy", GraphMetadata{}, packer.QueryFactual, ModeHybridBM25Dense, 3000, false},
		{"fallback", "tell me everything relevant about the customer support refund knowledge base", GraphMetadata{}, packer.QueryFactual, ModeHybridBM25Dense, 4000, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewRouter().Route(tc.query, tc.metadata)
			if got.QueryType != tc.queryType || got.RetrievalMode != tc.mode || got.ContextBudgetHint != tc.budget || got.CounterpointRequired != tc.counter {
				t.Fatalf("got %+v", got)
			}
			if len(got.Reasons) == 0 {
				t.Fatalf("missing reasons: %+v", got)
			}
		})
	}
}

func TestRouteRuleOrder(t *testing.T) {
	got := NewRouter().Route("approve latest refund", GraphMetadata{HasDecisionBasis: true})
	if got.QueryType != packer.QueryAgentic {
		t.Fatalf("agentic should win before temporal, got %+v", got)
	}
}
