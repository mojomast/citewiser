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

// ContextFromGovOne constructs a CitewiseRAG access.Context from a GovOne caller identity.
//
// callerID is the authenticated user or service account ID.
// role is the caller's GovOne RBAC role, which maps to a CitewiseRAG clearance level:
//   - super_admin, tenant_admin -> restricted (sees all sensitivity levels)
//   - governance_lead, operator -> confidential
//   - auditor, viewer           -> internal
//   - (unknown)                 -> public
//
// tenantID scopes the context to tenant-tagged RAG nodes. It is added to groups,
// trusted approvers, and attributes so that DefaultController can enforce tenant isolation.
// Pass an empty string for single-tenant or cross-tenant super-admin contexts.
//
// additionalGroups supplements the tenant-derived group membership with any extra
// GovOne group IDs (e.g., department codes, project IDs, or feature-flag groups)
// that downstream access policies may check via ctx.Groups.
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
