package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
)

// PostgresTaxonomyRepository implements domain.AssetTaxonomyRepository.
type PostgresTaxonomyRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresTaxonomyRepository wires the repo.
func NewPostgresTaxonomyRepository(pool *pgxpool.Pool) *PostgresTaxonomyRepository {
	return &PostgresTaxonomyRepository{pool: pool}
}

func (r *PostgresTaxonomyRepository) ListAssetClasses(ctx context.Context, includeInactive bool) ([]*entity.AssetClass, error) {
	q := `SELECT id, code, name, COALESCE(description,''), display_order, is_active, created_at, updated_at
	      FROM investment__asset_classes`
	if !includeInactive {
		q += " WHERE is_active = true"
	}
	q += " ORDER BY display_order, code"

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("listing asset classes: %w", err)
	}
	defer rows.Close()

	out := []*entity.AssetClass{}
	for rows.Next() {
		var a entity.AssetClass
		if err := rows.Scan(&a.ID, &a.Code, &a.Name, &a.Description,
			&a.DisplayOrder, &a.IsActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &a)
	}
	return out, rows.Err()
}

func (r *PostgresTaxonomyRepository) ListAssetSubtypes(ctx context.Context, assetClassID *uuid.UUID, includeInactive bool) ([]*entity.AssetSubtype, error) {
	q := `SELECT id, asset_class_id, code, name, COALESCE(description,''), display_order, is_active, created_at, updated_at
	      FROM investment__asset_subtypes WHERE 1=1`
	args := []any{}
	idx := 1
	if !includeInactive {
		q += " AND is_active = true"
	}
	if assetClassID != nil {
		q += fmt.Sprintf(" AND asset_class_id = $%d", idx)
		args = append(args, *assetClassID)
		idx++
	}
	q += " ORDER BY display_order, code"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("listing asset subtypes: %w", err)
	}
	defer rows.Close()

	out := []*entity.AssetSubtype{}
	for rows.Next() {
		var s entity.AssetSubtype
		if err := rows.Scan(&s.ID, &s.AssetClassID, &s.Code, &s.Name, &s.Description,
			&s.DisplayOrder, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}

func (r *PostgresTaxonomyRepository) ListRegions(ctx context.Context) ([]*entity.Region, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name, is_active, created_at, updated_at
		  FROM investment__regions
		 WHERE is_active = true ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("listing regions: %w", err)
	}
	defer rows.Close()

	out := []*entity.Region{}
	for rows.Next() {
		var x entity.Region
		if err := rows.Scan(&x.ID, &x.Code, &x.Name, &x.IsActive, &x.CreatedAt, &x.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &x)
	}
	return out, rows.Err()
}

func (r *PostgresTaxonomyRepository) ListCountries(ctx context.Context, regionID *uuid.UUID) ([]*entity.Country, error) {
	q := `SELECT id, iso_code, name, region_id, is_active, created_at, updated_at
	      FROM investment__countries WHERE is_active = true`
	args := []any{}
	if regionID != nil {
		q += " AND region_id = $1"
		args = append(args, *regionID)
	}
	q += " ORDER BY iso_code"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("listing countries: %w", err)
	}
	defer rows.Close()

	out := []*entity.Country{}
	for rows.Next() {
		var x entity.Country
		if err := rows.Scan(&x.ID, &x.ISOCode, &x.Name, &x.RegionID, &x.IsActive, &x.CreatedAt, &x.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &x)
	}
	return out, rows.Err()
}

func (r *PostgresTaxonomyRepository) ListSectors(ctx context.Context, parentID *uuid.UUID, level *int) ([]*entity.Sector, error) {
	q := `SELECT id, code, name, parent_id, level, is_active, created_at, updated_at
	      FROM investment__sectors WHERE is_active = true`
	args := []any{}
	idx := 1
	if parentID != nil {
		q += fmt.Sprintf(" AND parent_id = $%d", idx)
		args = append(args, *parentID)
		idx++
	}
	if level != nil {
		q += fmt.Sprintf(" AND level = $%d", idx)
		args = append(args, *level)
		idx++
	}
	q += " ORDER BY level, code"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("listing sectors: %w", err)
	}
	defer rows.Close()

	out := []*entity.Sector{}
	for rows.Next() {
		var x entity.Sector
		if err := rows.Scan(&x.ID, &x.Code, &x.Name, &x.ParentID, &x.Level, &x.IsActive, &x.CreatedAt, &x.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &x)
	}
	return out, rows.Err()
}

func (r *PostgresTaxonomyRepository) ListFundCategories(ctx context.Context) ([]*entity.FundCategory, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name, asset_class_id, COALESCE(description,''), is_active, created_at, updated_at
		  FROM investment__fund_categories
		 WHERE is_active = true ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("listing fund categories: %w", err)
	}
	defer rows.Close()

	out := []*entity.FundCategory{}
	for rows.Next() {
		var x entity.FundCategory
		if err := rows.Scan(&x.ID, &x.Code, &x.Name, &x.AssetClassID, &x.Description, &x.IsActive, &x.CreatedAt, &x.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &x)
	}
	return out, rows.Err()
}

func (r *PostgresTaxonomyRepository) ListInvestmentStyles(ctx context.Context) ([]*entity.InvestmentStyle, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name, is_active, created_at, updated_at
		  FROM investment__investment_styles
		 WHERE is_active = true ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("listing investment styles: %w", err)
	}
	defer rows.Close()

	out := []*entity.InvestmentStyle{}
	for rows.Next() {
		var x entity.InvestmentStyle
		if err := rows.Scan(&x.ID, &x.Code, &x.Name, &x.IsActive, &x.CreatedAt, &x.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &x)
	}
	return out, rows.Err()
}

var _ domain.AssetTaxonomyRepository = (*PostgresTaxonomyRepository)(nil)
