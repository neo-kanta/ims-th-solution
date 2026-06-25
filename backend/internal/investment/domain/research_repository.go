package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// ResearchReportListFilter describes the supported filters and pagination for
// list queries on research reports. Zero/nil values disable the corresponding
// filter.
type ResearchReportListFilter struct {
	ReportStatus   *vo.ReportStatus
	ReviewStatus   *vo.ReviewStatus
	Recommendation *vo.Recommendation
	InstrumentCode string
	OwnerUserID    *uuid.UUID
	ReportDateFrom *time.Time
	ReportDateTo   *time.Time
	Search         string
	Page           int
	Limit          int
}

// ResearchReportRepository persists ResearchReport aggregates.
// Implementations live in infrastructure/persistence/.
type ResearchReportRepository interface {
	// Create inserts a new report inside the supplied transaction.
	Create(ctx context.Context, tx pgx.Tx, r *entity.ResearchReport) error

	// GetByID returns the live (non-deleted) report or nil when missing.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ResearchReport, error)

	// GetByReportNo returns the live report by its business report_no
	// or nil when missing. Used to enforce uniqueness on create.
	GetByReportNo(ctx context.Context, reportNo string) (*entity.ResearchReport, error)

	// List paginates reports with optional filters.
	List(ctx context.Context, filter ResearchReportListFilter) ([]*entity.ResearchReport, int, error)

	// Update applies a metadata change.
	Update(ctx context.Context, tx pgx.Tx, r *entity.ResearchReport) error

	// SoftDelete stamps deleted_at and updated_at.
	SoftDelete(ctx context.Context, tx pgx.Tx, id uuid.UUID, deletedBy uuid.UUID, deletedAt time.Time) error

	// Invalidate atomically flips report_status and review_status to
	// INVALIDATED and stores the invalidation actor/reason/timestamp. The
	// repository MUST refuse to invalidate a row that is already INVALIDATED
	// or soft-deleted; callers are expected to pre-check entity.CanInvalidate
	// so a typed conflict error reaches the API caller.
	Invalidate(ctx context.Context, tx pgx.Tx, id uuid.UUID, actorID uuid.UUID, reason string, at time.Time) error
}
