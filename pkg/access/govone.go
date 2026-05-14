package access

// GovOneRole represents a GovOne RBAC role string.
type GovOneRole string

const (
	RoleSuperAdmin     GovOneRole = "super_admin"
	RoleTenantAdmin    GovOneRole = "tenant_admin"
	RoleGovernanceLead GovOneRole = "governance_lead"
	RoleOperator       GovOneRole = "operator"
	RoleAuditor        GovOneRole = "auditor"
	RoleViewer         GovOneRole = "viewer"
)

// ContextFromGovOne constructs a CitewiseRAG access context from a GovOne
// caller identity. tenantID is included in groups, trusted approvers, and
// attributes so tenant-scoped RAG nodes can be access-gated consistently.
func ContextFromGovOne(callerID string, role GovOneRole, tenantID string, additionalGroups []string) Context {
	groups := append([]string(nil), additionalGroups...)
	trustedApprovers := []string(nil)
	attributes := map[string]string{}
	if tenantID != "" {
		groups = append([]string{tenantID}, groups...)
		trustedApprovers = []string{tenantID}
		attributes[AttrTenantID] = tenantID
	}
	return Context{
		CallerID:         callerID,
		Groups:           groups,
		Clearance:        clearanceForGovOneRole(role),
		TrustedApprovers: trustedApprovers,
		Attributes:       attributes,
	}
}

func clearanceForGovOneRole(role GovOneRole) Clearance {
	switch role {
	case RoleSuperAdmin, RoleTenantAdmin:
		return ClearanceRestricted
	case RoleGovernanceLead, RoleOperator:
		return ClearanceConfidential
	case RoleAuditor, RoleViewer:
		return ClearanceInternal
	default:
		return ClearancePublic
	}
}
