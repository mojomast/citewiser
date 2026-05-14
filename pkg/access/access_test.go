package access

import (
	"testing"

	"github.com/mojomast/citewiser/pkg/citewise"
	"github.com/mojomast/citewiser/pkg/ragnode"
)

func TestClearanceCombinations(t *testing.T) {
	controller := NewController()
	cases := []struct {
		name        string
		clearance   Clearance
		sensitivity ragnode.Sensitivity
		allowed     bool
	}{
		{"public sees public", ClearancePublic, ragnode.SensitivityPublic, true},
		{"public denied internal", ClearancePublic, ragnode.SensitivityInternal, false},
		{"internal sees internal", ClearanceInternal, ragnode.SensitivityInternal, true},
		{"internal denied confidential", ClearanceInternal, ragnode.SensitivityConfidential, false},
		{"confidential sees confidential", ClearanceConfidential, ragnode.SensitivityConfidential, true},
		{"confidential denied restricted", ClearanceConfidential, ragnode.SensitivityRestricted, false},
		{"restricted sees restricted", ClearanceRestricted, ragnode.SensitivityRestricted, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decision := controller.CanSeeNode(Context{Clearance: tc.clearance}, ragnode.RAGNode{Sensitivity: tc.sensitivity})
			if decision.Allowed != tc.allowed {
				t.Fatalf("Allowed got %v want %v: %+v", decision.Allowed, tc.allowed, decision)
			}
			if !tc.allowed && decision.Reason != ReasonAccessControl {
				t.Fatalf("Reason got %q", decision.Reason)
			}
		})
	}
}

func TestTrustedApproversForConfidentialNodes(t *testing.T) {
	controller := NewController()
	node := ragnode.RAGNode{Sensitivity: ragnode.SensitivityConfidential, ApprovedBy: []string{"legal"}}
	denied := controller.CanSeeNode(Context{Clearance: ClearanceConfidential, TrustedApprovers: []string{"finance"}}, node)
	if denied.Allowed || denied.Reason != ReasonAccessControl {
		t.Fatalf("expected denial, got %+v", denied)
	}
	allowed := controller.CanSeeNode(Context{Clearance: ClearanceConfidential, TrustedApprovers: []string{"legal"}}, node)
	if !allowed.Allowed || allowed.Reason != ReasonAllowed {
		t.Fatalf("expected allow, got %+v", allowed)
	}
}

func TestAgenticNodeApprovalRule(t *testing.T) {
	controller := NewController()
	for _, chunkType := range []ragnode.ChunkType{ragnode.ChunkPermissionRecord, ragnode.ChunkDecision, ragnode.ChunkProcedure} {
		node := ragnode.RAGNode{ChunkType: chunkType, Sensitivity: ragnode.SensitivityInternal}
		decision := controller.CanSeeNode(Context{Clearance: ClearanceInternal}, node)
		if decision.Allowed {
			t.Fatalf("%s without approval should be denied", chunkType)
		}
		allowed := controller.CanSeeNode(Context{Clearance: ClearanceInternal, Attributes: map[string]string{AttrAllowUnapprovedAgenticNodes: "true"}}, node)
		if !allowed.Allowed {
			t.Fatalf("%s with escape hatch should be allowed: %+v", chunkType, allowed)
		}
		node.ApprovedBy = []string{"ops"}
		approved := controller.CanSeeNode(Context{Clearance: ClearanceInternal}, node)
		if !approved.Allowed {
			t.Fatalf("%s with approval should be allowed: %+v", chunkType, approved)
		}
	}
}

func TestAccessPolicyConstantsDocumentDefaults(t *testing.T) {
	if AttrAllowUnapprovedAgenticNodes != "allow_unapproved_agentic_nodes" {
		t.Fatalf("unexpected attribute key %q", AttrAllowUnapprovedAgenticNodes)
	}
	if SuppressionAuditRedacted != "redacted" {
		t.Fatalf("unexpected suppression audit level %q", SuppressionAuditRedacted)
	}
}

func TestCanUseEdgeAllowsEdges(t *testing.T) {
	decision := NewController().CanUseEdge(Context{Clearance: ClearancePublic}, ragnode.Edge{SourceID: "a", TargetID: "b", Type: ragnode.EdgeRelatedTo})
	if !decision.Allowed || decision.Reason != ReasonAllowed {
		t.Fatalf("edge decision got %+v", decision)
	}
}

func TestRedactNodeRemovesSensitiveFields(t *testing.T) {
	node := ragnode.RAGNode{
		Item: citewise.Item{
			ID:     "secret",
			Title:  "Secret Title",
			Source: "Secret KB",
			URL:    "https://secret.example",
			Notes:  "hidden notes",
		},
		Text:           "hidden text",
		ChunkType:      ragnode.ChunkPermissionRecord,
		Sensitivity:    ragnode.SensitivityRestricted,
		EmbeddingModel: "embedder",
		ApprovedBy:     []string{"legal"},
		ContextPrefix:  "hidden context",
		CommunityID:    "c1",
		SourceTrail:    []ragnode.SourceHop{{NodeID: "secret", EdgeType: "retrieved", Confidence: 1}},
		Origin:         "graphrag",
		Locator:        ragnode.Locator{DocumentID: "doc"},
		SemanticType:   "policy",
		Attributes:     map[string]string{"raw": "secret"},
	}
	redacted := NewController().RedactNode(Context{}, node)
	if redacted.ID != "secret" {
		t.Fatalf("redacted ID got %q", redacted.ID)
	}
	if redacted.Title != "" || redacted.Source != "" || redacted.URL != "" || redacted.Notes != "" || redacted.Text != "" || len(redacted.SourceTrail) != 0 || redacted.Attributes != nil {
		t.Fatalf("redacted node leaked fields: %+v", redacted)
	}
	if redacted.ChunkType != "" || redacted.Sensitivity != "" || len(redacted.ApprovedBy) != 0 || redacted.Origin != "" || redacted.Locator.DocumentID != "" {
		t.Fatalf("redacted node retained non-audit fields: %+v", redacted)
	}
}
