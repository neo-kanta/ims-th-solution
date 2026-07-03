package permission_test

import (
	"testing"

	permcode "github.com/neo-kanta/ims-th-solution/backend/internal/permissions/permission"
)

// allCodes is the complete set of fine-grained route-guard codes declared by
// the permissions module. Any new constant used in a RequirePermission call
// must be added here so coverage and uniqueness checks stay current.
var allCodes = []string{
	permcode.CodeUsersView,
	permcode.CodeGroupsView,
	permcode.CodeRolesView,
	permcode.CodeFunctionRightsView,
	permcode.CodeDataRightsView,
	permcode.CodeChangeRequestCreate,
	permcode.CodeChangeRequestSubmit,
	permcode.CodeChangeRequestReview,
	permcode.CodeChangeRequestApprove,
	permcode.CodeChangeRequestMerge,
	permcode.CodeChangeRequestReject,
	permcode.CodeChangeRequestClose,
	permcode.CodeChangeRequestCancel,
	permcode.CodeAuditView,
	permcode.CodeAuditExport,
	permcode.CodeNotificationView,
	permcode.CodeNotificationEdit,
	permcode.CodeLabelsManage,
	permcode.CodeApprovalSettingsView,
}

// TestPermissionCodes_Unique asserts that no two permission constants share the
// same string value. Duplicate values create privilege-escalation paths where
// granting one permission silently grants another.
func TestPermissionCodes_Unique(t *testing.T) {
	t.Parallel()
	seen := make(map[string]string, len(allCodes))
	for _, code := range allCodes {
		if prev, dup := seen[code]; dup {
			t.Errorf("duplicate permission constant value %q: both %q and another constant share it (prev was %q)", code, code, prev)
		}
		seen[code] = code
	}
}

// TestPermissionCodes_KnownRegressions guards specific constants that were
// previously wrong due to copy-paste errors.
func TestPermissionCodes_KnownRegressions(t *testing.T) {
	t.Parallel()

	if permcode.CodeLabelsManage == permcode.CodeChangeRequestCreate {
		t.Errorf("CodeLabelsManage must not equal CodeChangeRequestCreate: label management and change-request creation are different permissions")
	}
	if permcode.CodeApprovalSettingsView == permcode.CodeChangeRequestReview {
		t.Errorf("CodeApprovalSettingsView must not equal CodeChangeRequestReview: approval-settings view and change-request review are different permissions")
	}
	if permcode.CodeLabelsManage != "permission.labels.manage" {
		t.Errorf("CodeLabelsManage = %q, want %q", permcode.CodeLabelsManage, "permission.labels.manage")
	}
	if permcode.CodeApprovalSettingsView != "permission.approval_settings.view" {
		t.Errorf("CodeApprovalSettingsView = %q, want %q", permcode.CodeApprovalSettingsView, "permission.approval_settings.view")
	}
}

// TestProviderPermissions_LegacyCodesOnly asserts that Provider.Permissions()
// returns only the two legacy uppercase codes. Fine-grained codes must NOT be
// added here because permissions_function_definitions enforces UPPER(code).
// Fine-grained codes are seeded via database/seeds/008_permission_management_seed.sql.
func TestProviderPermissions_LegacyCodesOnly(t *testing.T) {
	t.Parallel()
	p := permcode.Provider{}
	defs := p.Permissions()
	if len(defs) != 2 {
		t.Fatalf("Provider.Permissions() returned %d codes, want exactly 2 legacy uppercase codes", len(defs))
	}
	for _, d := range defs {
		if d.Code != "PERMISSIONS_VIEW" && d.Code != "PERMISSIONS_MANAGE" {
			t.Errorf("unexpected code in Provider.Permissions(): %q — only PERMISSIONS_VIEW and PERMISSIONS_MANAGE are legacy catalog codes", d.Code)
		}
	}
}

// TestPermissionCodes_NoWhitespace asserts that all constants are free of
// leading/trailing whitespace, which the DB CHECK constraint also enforces.
func TestPermissionCodes_NoWhitespace(t *testing.T) {
	t.Parallel()
	for _, code := range allCodes {
		if code == "" {
			t.Error("empty permission constant found")
			continue
		}
		if code[0] == ' ' || code[len(code)-1] == ' ' {
			t.Errorf("permission constant %q has leading or trailing whitespace", code)
		}
	}
}
