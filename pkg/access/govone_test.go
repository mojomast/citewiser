package access

import "testing"

func TestContextFromGovOneMapsRoles(t *testing.T) {
	cases := []struct {
		role GovOneRole
		want Clearance
	}{
		{RoleSuperAdmin, ClearanceRestricted},
		{RoleTenantAdmin, ClearanceRestricted},
		{RoleGovernanceLead, ClearanceConfidential},
		{RoleOperator, ClearanceConfidential},
		{RoleAuditor, ClearanceInternal},
		{RoleViewer, ClearanceInternal},
		{"", ClearancePublic},
	}
	for _, tc := range cases {
		ctx := ContextFromGovOne("caller", tc.role, "tenant-a", []string{"extra"})
		if ctx.Clearance != tc.want {
			t.Fatalf("role %q clearance got %q want %q", tc.role, ctx.Clearance, tc.want)
		}
		if len(ctx.TrustedApprovers) != 1 || ctx.TrustedApprovers[0] != "tenant-a" {
			t.Fatalf("trusted approvers got %v", ctx.TrustedApprovers)
		}
		if ctx.Attributes[AttrTenantID] != "tenant-a" {
			t.Fatalf("tenant attribute got %q", ctx.Attributes[AttrTenantID])
		}
		if len(ctx.Groups) != 2 || ctx.Groups[0] != "tenant-a" || ctx.Groups[1] != "extra" {
			t.Fatalf("groups got %v", ctx.Groups)
		}
	}
}
