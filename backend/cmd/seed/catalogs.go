package main

import (
	approvalperm "github.com/neo-kanta/ims-th-solution/backend/internal/approval/permission"
	auditperm "github.com/neo-kanta/ims-th-solution/backend/internal/audit/permission"
	iamperm "github.com/neo-kanta/ims-th-solution/backend/internal/iam/permission"
	investmentperm "github.com/neo-kanta/ims-th-solution/backend/internal/investment/permission"
	leaveperm "github.com/neo-kanta/ims-th-solution/backend/internal/leave_delegation/permission"
	marketdataperm "github.com/neo-kanta/ims-th-solution/backend/internal/market_data/permission"
	notificationperm "github.com/neo-kanta/ims-th-solution/backend/internal/notification/permission"
	permissionsperm "github.com/neo-kanta/ims-th-solution/backend/internal/permissions/permission"
	workflowperm "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/permission"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// defaultPermissionCatalogs returns every module-owned PermissionCatalog
// registered with the seeder. Modules without operator-grantable codes
// (currently market_data) appear so the catalog test surfaces missing
// registrations rather than silent gaps.
func defaultPermissionCatalogs() []contract.PermissionCatalog {
	return []contract.PermissionCatalog{
		iamperm.Provider{},
		permissionsperm.Provider{},
		auditperm.Provider{},
		workflowperm.Provider{},
		investmentperm.Provider{},
		marketdataperm.Provider{},
		approvalperm.Provider{},
		leaveperm.Provider{},
		notificationperm.Provider{},
	}
}
