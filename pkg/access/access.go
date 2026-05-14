package access

import (
	"fmt"

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

const sensitivityOrdinalUnknown = 3

const (
	ReasonAllowed       = "allowed"
	ReasonAccessControl = "access-control"
)

const (
	// AttrAllowUnapprovedAgenticNodes lets internal validation paths inspect
	// unapproved controlling nodes. Production Agentic packing remains fail-closed.
	AttrAllowUnapprovedAgenticNodes = "allow_unapproved_agentic_nodes"
	// AttrTenantID scopes access to tenant-tagged RAG nodes.
	AttrTenantID = "tenant_id"
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

type edgeSensitivityController interface {
	CanUseEdgeBetween(ctx Context, edge ragnode.Edge, sourceSensitivity, targetSensitivity ragnode.Sensitivity) Decision
}

type DefaultController struct{}

func NewController() DefaultController {
	return DefaultController{}
}

func (DefaultController) CanSeeNode(ctx Context, node ragnode.RAGNode) Decision {
	if node.TenantID != "" && ctx.Attributes[AttrTenantID] != node.TenantID {
		return deny("node tenant %q does not match caller tenant", node.TenantID)
	}
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

// CanUseEdge allows all edges at the interface level. Callers with a full node map should
// prefer CanUseEdgeBetween to enforce sensitivity-derived edge access control.
func (DefaultController) CanUseEdge(ctx Context, edge ragnode.Edge) Decision {
	return Decision{Allowed: true, Reason: ReasonAllowed}
}

// CanUseEdgeBetween enforces that the caller can traverse an edge by checking that their
// clearance permits access to BOTH endpoint sensitivities. Use this when the caller has
// the full node map available.
func (DefaultController) CanUseEdgeBetween(ctx Context, edge ragnode.Edge, sourceSensitivity, targetSensitivity ragnode.Sensitivity) Decision {
	if sensitivityOrdinal(sourceSensitivity) > clearanceOrdinal(ctx.Clearance) {
		return deny("edge source sensitivity %q exceeds caller clearance %q", sourceSensitivity, ctx.Clearance)
	}
	if sensitivityOrdinal(targetSensitivity) > clearanceOrdinal(ctx.Clearance) {
		return deny("edge target sensitivity %q exceeds caller clearance %q", targetSensitivity, ctx.Clearance)
	}
	return Decision{Allowed: true, Reason: ReasonAllowed}
}

// EdgeDecision applies endpoint sensitivity checks when ctrl supports
// CanUseEdgeBetween, falling back to Controller.CanUseEdge for legacy
// Controller implementations. New controllers should add CanUseEdgeBetween to
// enforce sensitivity-derived edge access without changing the Controller
// interface contract.
func EdgeDecision(ctrl Controller, ctx Context, edge ragnode.Edge, sourceSensitivity, targetSensitivity ragnode.Sensitivity) Decision {
	if checker, ok := ctrl.(edgeSensitivityController); ok {
		return checker.CanUseEdgeBetween(ctx, edge, sourceSensitivity, targetSensitivity)
	}
	return ctrl.CanUseEdge(ctx, edge)
}

func (DefaultController) RedactNode(ctx Context, node ragnode.RAGNode) ragnode.RAGNode {
	return ragnode.RAGNode{Item: citewise.Item{ID: node.ID}}
}

// ValidateSensitivity returns false if s is not a recognized Sensitivity value.
// Callers building CandidateSet nodes should call this before submitting to the pipeline
// and log a warning when it returns false; CitewiseRAG will treat unknown values as
// restricted (fail-closed) but will not surface a warning itself.
func ValidateSensitivity(s ragnode.Sensitivity) bool {
	switch s {
	case ragnode.SensitivityPublic, ragnode.SensitivityInternal,
		ragnode.SensitivityConfidential, ragnode.SensitivityRestricted, "":
		return true
	default:
		return false
	}
}

// sensitivityOrdinal maps known sensitivities to clearance levels. Unknown values map to
// restricted via sensitivityOrdinalUnknown; callers should use ValidateSensitivity to
// detect this before submission.
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
		return sensitivityOrdinalUnknown
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

// deny returns a Decision that blocks access with an access-control reason and a
// formatted detail message for audit logs.
func deny(format string, args ...any) Decision {
	return Decision{Allowed: false, Reason: ReasonAccessControl, Detail: fmt.Sprintf(format, args...)}
}
