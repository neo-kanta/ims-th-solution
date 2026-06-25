package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// PostgresPortfolioStatusHistoryRepository implements domain.PortfolioStatusHistoryRepository.
type PostgresPortfolioStatusHistoryRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresPortfolioStatusHistoryRepository wires the repo.
func NewPostgresPortfolioStatusHistoryRepository(pool *pgxpool.Pool) *PostgresPortfolioStatusHistoryRepository {
	return &PostgresPortfolioStatusHistoryRepository{pool: pool}
}

// Append inserts one immutable lifecycle row inside the caller's transaction.
func (r *PostgresPortfolioStatusHistoryRepository) Append(
	ctx context.Context,
	tx pgx.Tx,
	h *entity.PortfolioStatusHistory,
) error {
	var fromStatus *string
	if h.FromStatus != nil {
		s := string(*h.FromStatus)
		fromStatus = &s
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__portfolio_status_history
		    (id, portfolio_id, from_status, to_status, actor_id, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		h.ID, h.PortfolioID, fromStatus, string(h.ToStatus),
		h.ActorID, h.Reason, h.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("appending portfolio status history: %w", err)
	}
	return nil
}

// ListByPortfolio returns the full lifecycle log for a portfolio, newest first.
func (r *PostgresPortfolioStatusHistoryRepository) ListByPortfolio(
	ctx context.Context,
	portfolioID uuid.UUID,
) ([]*entity.PortfolioStatusHistory, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, portfolio_id, from_status, to_status, actor_id, reason, created_at
		  FROM investment__portfolio_status_history
		 WHERE portfolio_id = $1
		 ORDER BY created_at DESC`,
		portfolioID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing portfolio status history: %w", err)
	}
	defer rows.Close()

	var out []*entity.PortfolioStatusHistory
	for rows.Next() {
		h, err := scanPortfolioStatusHistory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func scanPortfolioStatusHistory(s scanner) (*entity.PortfolioStatusHistory, error) {
	var h entity.PortfolioStatusHistory
	var fromStr *string
	var toStr string
	err := s.Scan(
		&h.ID, &h.PortfolioID, &fromStr, &toStr,
		&h.ActorID, &h.Reason, &h.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning portfolio status history: %w", err)
	}
	if fromStr != nil {
		s := vo.PortfolioStatus(*fromStr)
		h.FromStatus = &s
	}
	h.ToStatus = vo.PortfolioStatus(toStr)
	return &h, nil
}

var _ domain.PortfolioStatusHistoryRepository = (*PostgresPortfolioStatusHistoryRepository)(nil)
