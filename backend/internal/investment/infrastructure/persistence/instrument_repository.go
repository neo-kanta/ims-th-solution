package persistence

import (
	"context"
	"encoding/json"
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

// PostgresInstrumentRepository implements domain.InstrumentRepository.
type PostgresInstrumentRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresInstrumentRepository wires the repo.
func NewPostgresInstrumentRepository(pool *pgxpool.Pool) *PostgresInstrumentRepository {
	return &PostgresInstrumentRepository{pool: pool}
}

const instrumentSelect = `
	SELECT id, primary_ticker, name,
	       asset_class_id, asset_subtype_id, currency, country_id, region_id,
	       COALESCE(primary_exchange, ''), sector_id, fund_category_id,
	       lot_size, tick_size, is_tradable, status,
	       attributes,
	       created_at, updated_at, created_by, updated_by, deleted_at
	FROM investment__instruments`

func (r *PostgresInstrumentRepository) Create(ctx context.Context, tx pgx.Tx, inst *entity.Instrument) error {
	attrsJSON, err := json.Marshal(inst.Attributes)
	if err != nil {
		return fmt.Errorf("marshalling instrument attributes: %w", err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO investment__instruments (
			id, primary_ticker, name,
			asset_class_id, asset_subtype_id, currency, country_id, region_id,
			primary_exchange, sector_id, fund_category_id,
			lot_size, tick_size, is_tradable, status,
			attributes,
			created_at, updated_at, created_by, updated_by
		) VALUES (
			$1, $2, $3,
			$4, $5, $6, $7, $8,
			NULLIF($9,''), $10, $11,
			$12, $13, $14, $15,
			$16,
			$17, $18, $19, $20
		)`,
		inst.ID, inst.PrimaryTicker, inst.Name,
		inst.AssetClassID, inst.AssetSubtypeID, inst.Currency, inst.CountryID, inst.RegionID,
		inst.PrimaryExchange, inst.SectorID, inst.FundCategoryID,
		inst.LotSize, inst.TickSize, inst.IsTradable, string(inst.Status),
		attrsJSON,
		inst.CreatedAt, inst.UpdatedAt, inst.CreatedBy, inst.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting instrument: %w", err)
	}
	return nil
}

func (r *PostgresInstrumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Instrument, error) {
	row := r.pool.QueryRow(ctx, instrumentSelect+" WHERE id = $1 AND deleted_at IS NULL", id)
	return scanInstrument(row)
}

func (r *PostgresInstrumentRepository) List(ctx context.Context, filter domain.InstrumentListFilter) ([]*entity.Instrument, int, error) {
	conds := []string{"deleted_at IS NULL"}
	args := []any{}
	idx := 1
	if filter.AssetClassID != nil {
		conds = append(conds, fmt.Sprintf("asset_class_id = $%d", idx))
		args = append(args, *filter.AssetClassID)
		idx++
	}
	if filter.AssetSubtypeID != nil {
		conds = append(conds, fmt.Sprintf("asset_subtype_id = $%d", idx))
		args = append(args, *filter.AssetSubtypeID)
		idx++
	}
	if filter.SectorID != nil {
		conds = append(conds, fmt.Sprintf("sector_id = $%d", idx))
		args = append(args, *filter.SectorID)
		idx++
	}
	if filter.CountryID != nil {
		conds = append(conds, fmt.Sprintf("country_id = $%d", idx))
		args = append(args, *filter.CountryID)
		idx++
	}
	if filter.Status != nil {
		conds = append(conds, fmt.Sprintf("status = $%d", idx))
		args = append(args, string(*filter.Status))
		idx++
	}
	if filter.Search != "" {
		conds = append(conds, fmt.Sprintf("(primary_ticker ILIKE $%d OR name ILIKE $%d)", idx, idx))
		args = append(args, "%"+filter.Search+"%")
		idx++
	}
	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM investment__instruments "+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting instruments: %w", err)
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
		instrumentSelect+" "+where+
			fmt.Sprintf(" ORDER BY primary_ticker LIMIT $%d OFFSET $%d", idx, idx+1),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing instruments: %w", err)
	}
	defer rows.Close()

	out := make([]*entity.Instrument, 0, limit)
	for rows.Next() {
		inst, err := scanInstrument(rows)
		if err != nil {
			return nil, 0, err
		}
		if inst != nil {
			out = append(out, inst)
		}
	}
	return out, total, rows.Err()
}

func (r *PostgresInstrumentRepository) Update(ctx context.Context, tx pgx.Tx, inst *entity.Instrument) error {
	attrsJSON, err := json.Marshal(inst.Attributes)
	if err != nil {
		return fmt.Errorf("marshalling attributes: %w", err)
	}
	_, err = tx.Exec(ctx, `
		UPDATE investment__instruments
		   SET name              = $2,
		       primary_exchange  = NULLIF($3,''),
		       sector_id         = $4,
		       fund_category_id  = $5,
		       lot_size          = $6,
		       tick_size         = $7,
		       is_tradable       = $8,
		       status            = $9,
		       attributes        = $10,
		       updated_at        = $11,
		       updated_by        = $12
		 WHERE id = $1 AND deleted_at IS NULL`,
		inst.ID,
		inst.Name, inst.PrimaryExchange, inst.SectorID, inst.FundCategoryID,
		inst.LotSize, inst.TickSize, inst.IsTradable, string(inst.Status),
		attrsJSON, inst.UpdatedAt, inst.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("updating instrument: %w", err)
	}
	return nil
}

func (r *PostgresInstrumentRepository) SoftDelete(ctx context.Context, tx pgx.Tx, id uuid.UUID, deletedBy uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE investment__instruments
		   SET deleted_at = NOW(),
		       updated_at = NOW(),
		       updated_by = $2
		 WHERE id = $1 AND deleted_at IS NULL`,
		id, deletedBy,
	)
	return err
}

func scanInstrument(s scanner) (*entity.Instrument, error) {
	var inst entity.Instrument
	var statusStr string
	var attrsJSON []byte
	err := s.Scan(
		&inst.ID, &inst.PrimaryTicker, &inst.Name,
		&inst.AssetClassID, &inst.AssetSubtypeID, &inst.Currency, &inst.CountryID, &inst.RegionID,
		&inst.PrimaryExchange, &inst.SectorID, &inst.FundCategoryID,
		&inst.LotSize, &inst.TickSize, &inst.IsTradable, &statusStr,
		&attrsJSON,
		&inst.CreatedAt, &inst.UpdatedAt, &inst.CreatedBy, &inst.UpdatedBy, &inst.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning instrument: %w", err)
	}
	inst.Status = vo.InstrumentStatus(statusStr)
	if len(attrsJSON) > 0 {
		_ = json.Unmarshal(attrsJSON, &inst.Attributes)
	}
	return &inst, nil
}

var _ domain.InstrumentRepository = (*PostgresInstrumentRepository)(nil)
