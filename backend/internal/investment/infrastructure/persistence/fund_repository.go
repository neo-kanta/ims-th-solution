// Package persistence holds the pgx-backed implementations of the investment
// module's domain repository interfaces.
package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// PostgresFundRepository implements domain.FundRepository.
type PostgresFundRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresFundRepository wires the repo.
func NewPostgresFundRepository(pool *pgxpool.Pool) *PostgresFundRepository {
	return &PostgresFundRepository{pool: pool}
}

const fundSelect = `
	SELECT id, code, name, COALESCE(short_name, ''),
	       fund_category_id, base_currency, inception_date,
	       manager_user_id, COALESCE(benchmark, ''), COALESCE(risk_profile, ''),
	       has_units, require_pretrade_preview, COALESCE(external_pam_ref, ''),
	       status, version,
	       created_at, updated_at, created_by, updated_by, deleted_at
	FROM investment__funds`

// Create inserts a fund row inside the supplied transaction.
func (r *PostgresFundRepository) Create(ctx context.Context, tx pgx.Tx, f *entity.Fund) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__funds (
			id, code, name, short_name, fund_category_id, base_currency,
			inception_date, manager_user_id, benchmark, risk_profile,
			has_units, require_pretrade_preview, external_pam_ref, status, version,
			created_at, updated_at, created_by, updated_by
		) VALUES (
			$1, $2, $3, NULLIF($4,''), $5, $6,
			$7, $8, NULLIF($9,''), NULLIF($10,''),
			$11, $12, NULLIF($13,''), $14, $15,
			$16, $17, $18, $19
		)`,
		f.ID, f.Code, f.Name, f.ShortName, f.FundCategoryID, f.BaseCurrency,
		f.InceptionDate, f.ManagerUserID, f.Benchmark, string(f.RiskProfile),
		f.HasUnits, f.RequirePretradePreview, f.ExternalPAMRef, string(f.Status), f.Version,
		f.CreatedAt, f.UpdatedAt, f.CreatedBy, f.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting fund: %w", err)
	}
	return nil
}

// GetByID returns the fund by primary key, alive only.
func (r *PostgresFundRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Fund, error) {
	row := r.pool.QueryRow(ctx, fundSelect+" WHERE id = $1 AND deleted_at IS NULL", id)
	return scanFund(row)
}

// GetByCode returns the fund by its business code, alive only.
func (r *PostgresFundRepository) GetByCode(ctx context.Context, code string) (*entity.Fund, error) {
	row := r.pool.QueryRow(ctx, fundSelect+" WHERE code = $1 AND deleted_at IS NULL", code)
	return scanFund(row)
}

// List paginates funds with optional filters.
func (r *PostgresFundRepository) List(ctx context.Context, filter domain.FundListFilter) ([]*entity.Fund, int, error) {
	conds := []string{"deleted_at IS NULL"}
	args := []any{}
	idx := 1
	if filter.Status != nil {
		conds = append(conds, fmt.Sprintf("status = $%d", idx))
		args = append(args, string(*filter.Status))
		idx++
	}
	if filter.FundCategoryID != nil {
		conds = append(conds, fmt.Sprintf("fund_category_id = $%d", idx))
		args = append(args, *filter.FundCategoryID)
		idx++
	}
	if filter.ManagerUserID != nil {
		conds = append(conds, fmt.Sprintf("manager_user_id = $%d", idx))
		args = append(args, *filter.ManagerUserID)
		idx++
	}
	if len(filter.IDs) > 0 {
		conds = append(conds, fmt.Sprintf("id = ANY($%d)", idx))
		args = append(args, filter.IDs)
		idx++
	}
	if filter.AccessibleFundIDs != nil {
		if len(filter.AccessibleFundIDs) == 0 {
			return []*entity.Fund{}, 0, nil
		}
		conds = append(conds, fmt.Sprintf("id = ANY($%d)", idx))
		args = append(args, filter.AccessibleFundIDs)
		idx++
	}
	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM investment__funds "+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting funds: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := 0
	if filter.Page > 1 {
		offset = (filter.Page - 1) * limit
	}

	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx,
		fundSelect+" "+where+
			fmt.Sprintf(" ORDER BY code LIMIT $%d OFFSET $%d", idx, idx+1),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing funds: %w", err)
	}
	defer rows.Close()

	out := make([]*entity.Fund, 0, limit)
	for rows.Next() {
		f, err := scanFund(rows)
		if err != nil {
			return nil, 0, err
		}
		if f != nil {
			out = append(out, f)
		}
	}
	return out, total, rows.Err()
}

// Update applies a metadata change with optimistic version check.
func (r *PostgresFundRepository) Update(ctx context.Context, tx pgx.Tx, f *entity.Fund) error {
	tag, err := tx.Exec(ctx, `
		UPDATE investment__funds
		   SET name                     = $3,
		       short_name               = NULLIF($4,''),
		       fund_category_id         = $5,
		       manager_user_id          = $6,
		       benchmark                = NULLIF($7,''),
		       risk_profile             = NULLIF($8,''),
		       status                   = $9,
		       external_pam_ref         = NULLIF($10,''),
		       require_pretrade_preview = $14,
		       version                  = $11,
		       updated_at               = $12,
		       updated_by               = $13
		 WHERE id = $1 AND version = $2 AND deleted_at IS NULL`,
		f.ID, f.Version-1,
		f.Name, f.ShortName, f.FundCategoryID, f.ManagerUserID,
		f.Benchmark, string(f.RiskProfile), string(f.Status), f.ExternalPAMRef,
		f.Version, f.UpdatedAt, f.UpdatedBy, f.RequirePretradePreview,
	)
	if err != nil {
		return fmt.Errorf("updating fund: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return &domain.ErrFundVersionMismatch{
			FundID: f.ID.String(), ExpectedVersion: f.Version - 1, ActualVersion: -1,
		}
	}
	return nil
}

// SoftDelete stamps deleted_at with version check.
func (r *PostgresFundRepository) SoftDelete(ctx context.Context, tx pgx.Tx, id uuid.UUID, expectedVersion int, deletedBy uuid.UUID) error {
	tag, err := tx.Exec(ctx, `
		UPDATE investment__funds
		   SET deleted_at = NOW(),
		       updated_at = NOW(),
		       updated_by = $3,
		       version    = version + 1
		 WHERE id = $1 AND version = $2 AND deleted_at IS NULL`,
		id, expectedVersion, deletedBy,
	)
	if err != nil {
		return fmt.Errorf("soft-deleting fund: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return &domain.ErrFundVersionMismatch{
			FundID: id.String(), ExpectedVersion: expectedVersion, ActualVersion: -1,
		}
	}
	return nil
}

// CountActivePortfolios is used to refuse fund deletion when active portfolios remain.
func (r *PostgresFundRepository) CountActivePortfolios(ctx context.Context, fundID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM investment__portfolios
		 WHERE fund_id = $1 AND status = 'ACTIVE' AND deleted_at IS NULL`,
		fundID,
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("counting active portfolios: %w", err)
	}
	return n, nil
}

func scanFund(s scanner) (*entity.Fund, error) {
	var f entity.Fund
	var statusStr, riskStr string
	err := s.Scan(
		&f.ID, &f.Code, &f.Name, &f.ShortName,
		&f.FundCategoryID, &f.BaseCurrency, &f.InceptionDate,
		&f.ManagerUserID, &f.Benchmark, &riskStr,
		&f.HasUnits, &f.RequirePretradePreview, &f.ExternalPAMRef,
		&statusStr, &f.Version,
		&f.CreatedAt, &f.UpdatedAt, &f.CreatedBy, &f.UpdatedBy, &f.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning fund: %w", err)
	}
	f.Status = vo.FundStatus(statusStr)
	f.RiskProfile = vo.RiskProfile(riskStr)
	return &f, nil
}

var _ domain.FundRepository = (*PostgresFundRepository)(nil)
