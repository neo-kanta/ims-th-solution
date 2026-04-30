package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
)

// PostgresPriceRepository implements domain.PriceSnapshotRepository.
//
// Append-only at the storage layer (Postgres RULE blocks UPDATE/DELETE).
// This repo only exposes Insert and read operations.
type PostgresPriceRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresPriceRepository wires the repo.
func NewPostgresPriceRepository(pool *pgxpool.Pool) *PostgresPriceRepository {
	return &PostgresPriceRepository{pool: pool}
}

const priceSelect = `
	SELECT id, instrument_id, business_date,
	       price, currency, price_source, COALESCE(provider_ref, ''),
	       is_stale, COALESCE(stale_reason, ''),
	       captured_at, created_at, created_by
	FROM investment__price_snapshots`

func (r *PostgresPriceRepository) Insert(ctx context.Context, tx pgx.Tx, p *entity.PriceSnapshot) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__price_snapshots (
			id, instrument_id, business_date,
			price, currency, price_source, provider_ref,
			is_stale, stale_reason,
			captured_at, created_at, created_by
		) VALUES (
			$1, $2, $3,
			$4, $5, $6, NULLIF($7,''),
			$8, NULLIF($9,''),
			$10, $11, $12
		)`,
		p.ID, p.InstrumentID, p.BusinessDate,
		p.Price, p.Currency, p.PriceSource, p.ProviderRef,
		p.IsStale, p.StaleReason,
		p.CapturedAt, p.CreatedAt, p.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting price snapshot: %w", err)
	}
	return nil
}

// GetLatest returns the most recent price for the instrument on or before
// asOf. Returns nil (no error) when no row exists.
func (r *PostgresPriceRepository) GetLatest(ctx context.Context, instrumentID uuid.UUID, asOf time.Time) (*entity.PriceSnapshot, error) {
	row := r.pool.QueryRow(ctx,
		priceSelect+`
		 WHERE instrument_id = $1 AND business_date <= $2
		 ORDER BY business_date DESC, created_at DESC
		 LIMIT 1`,
		instrumentID, asOf,
	)
	return scanPrice(row)
}

func (r *PostgresPriceRepository) ListByInstrument(ctx context.Context, instrumentID uuid.UUID, from, to time.Time, limit int) ([]*entity.PriceSnapshot, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		priceSelect+`
		 WHERE instrument_id = $1 AND business_date >= $2 AND business_date <= $3
		 ORDER BY business_date DESC LIMIT $4`,
		instrumentID, from, to, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("listing prices: %w", err)
	}
	defer rows.Close()

	out := []*entity.PriceSnapshot{}
	for rows.Next() {
		p, err := scanPrice(rows)
		if err != nil {
			return nil, err
		}
		if p != nil {
			out = append(out, p)
		}
	}
	return out, rows.Err()
}

func scanPrice(s scanner) (*entity.PriceSnapshot, error) {
	var p entity.PriceSnapshot
	err := s.Scan(
		&p.ID, &p.InstrumentID, &p.BusinessDate,
		&p.Price, &p.Currency, &p.PriceSource, &p.ProviderRef,
		&p.IsStale, &p.StaleReason,
		&p.CapturedAt, &p.CreatedAt, &p.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning price: %w", err)
	}
	return &p, nil
}

var _ domain.PriceSnapshotRepository = (*PostgresPriceRepository)(nil)
