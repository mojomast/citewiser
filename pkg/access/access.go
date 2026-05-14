package access

import (
	"github.com/mojomast/citewiser/pkg/citewise"
	"github.com/mojomast/citewiser/pkg/ragnode"
)

type Clearance string

const (
	ClearancePublic       Clearance = "public"
	ClearanceInternal     Clearance = "internal"
	ClearanceConfidential Clearance = "confidential"
	ClearanceRestricted   Clearance = "restricted"
)

const (
	ReasonAllowed       = "allowed"
	ReasonAccessControl = "access-control"
)

const (
	// AttrAllowUnapprovedAgenticNodes lets internal validation paths inspect
	// unapproved controlling nodes. Production Agentic packing remains fail-closed.
	AttrAllowUnapprovedAgenticNodes = "allow_unapproved_agentic_nodes"
)

const (
	// SuppressionAuditRedacted is the MVP audit level for ordinary callers:
	// suppressed access-control entries may expose IDs, reasons, and details only.
	SuppressionAuditRedacted = "redacted"
)

type Context struct {
	CallerID         string            `json:"caller_id"`
	Groups           []string          `json:"groups,omitempty"`
	Clearance        Clearance         `json:"clearance"`
	TrustedApprovers []string          `json:"trusted_approvers,omitempty"`
	Attributes       map[string]string `json:"attributes,omitempty"`
}

type Decision struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
	Detail  string `json:"detail,omitempty"`
}

type Controller interface {
	CanSeeNode(ctx Context, node ragnode.RAGNode) Decision
	CanUseEdge(ctx Context, edge ragnode.Edge) Decision
	RedactNode(ctx Context, node ragnode.RAGNode) ragnode.RAGNode
}

type DefaultController struct{}

func NewController() DefaultController {
	return DefaultController{}
}

func (DefaultController) CanSeeNode(ctx Context, node ragnode.RAGNode) Decision {
	if sensitivityOrdinal(node.Sensitivity) > clearanceOrdinal(ctx.Clearance) {
		return deny("node sensitivity %q exceeds caller clearance %q", node.Sensitivity, ctx.Clearance)
	}
	if sensitivityOrdinal(node.Sensitivity) >= sensitivityOrdinal(ragnode.SensitivityConfidential) && len(node.ApprovedBy) > 0 && !HasTrustedApprover(ctx, node.ApprovedBy) {
		return deny("confidential or restricted node lacks a trusted approver for caller")
	}
	if requiresAgenticApproval(node) && len(node.ApprovedBy) == 0 && !allowUnapprovedAgenticNodes(ctx) {
		return deny("%s node lacks required approval", node.ChunkType)
	}
	return Decision{Allowed: true, Reason: ReasonAllowed}
}

func (DefaultController) CanUseEdge(ctx Context, edge ragnode.Edge) Decision {
	return Decision{Allowed: true, Reason: ReasonAllowed}
}

func (DefaultController) RedactNode(ctx Context, node ragnode.RAGNode) ragnode.RAGNode {
	return ragnode.RAGNode{Item: citewise.Item{ID: node.ID}}
}

func sensitivityOrdinal(s ragnode.Sensitivity) int {
	switch s {
	case ragnode.SensitivityPublic:
		return 0
	case "", ragnode.SensitivityInternal:
		return 1
	case ragnode.SensitivityConfidential:
		return 2
	case ragnode.SensitivityRestricted:
		return 3
	default:
		return 3
	}
}

func clearanceOrdinal(c Clearance) int {
	switch c {
	case ClearancePublic:
		return 0
	case "", ClearanceInternal:
		return 1
	case ClearanceConfidential:
		return 2
	case ClearanceRestricted:
		return 3
	default:
		return 0
	}
}

func requiresAgenticApproval(node ragnode.RAGNode) bool {
	switch node.ChunkType {
	case ragnode.ChunkPermissionRecord, ragnode.ChunkDecision, ragnode.ChunkProcedure:
		return true
	default:
		return false
	}
}

func allowUnapprovedAgenticNodes(ctx Context) bool {
	return ctx.Attributes != nil && ctx.Attributes[AttrAllowUnapprovedAgenticNodes] == "true"
}
