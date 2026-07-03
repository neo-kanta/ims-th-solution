package query

import "context"

// IAMPort is the minimal IAM surface consumed by integration dashboard queries.
type IAMPort interface {
	HasFunctionPermission(ctx context.Context, userID string, code string) (bool, error)
	GetAccessibleContracts(ctx context.Context, userID string) ([]string, error)
}
