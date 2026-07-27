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

// PostgresCashRequestRepository implements
// domain.PortfolioCashRequestRepository against
// investment__portfolio_cash_requests.
//
// This table is deliberately MUTABLE (status transitions) — there are no
// append-only RULEs on it, unlike investment__portfolio_transactions.
type PostgresCashRequestRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresCashRequestRepository wires the repo.
func NewPostgresCashRequestRepository(pool *pgxpool.Pool) *PostgresCashRequestRepository {
	return &PostgresCashRequestRepository{pool: pool}
}

const cashReqSelect = `
	SELECT id, portfolio_id, fund_id,
	       transaction_type, amount, currency, fees, value_date, COALESCE(memo, ''),
	       status, approval_request_id, resulting_txn_id, idempotency_key, request_fingerprint,
	       submitted_by, submitted_at, decided_by, decided_at,
	       version, created_at, updated_at, created_by, updated_by
	FROM investment__portfolio_cash_requests`

func (r *PostgresCashRequestRepository) Create(ctx context.Context, tx pgx.Tx, c *entity.PortfolioCashRequest) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__portfolio_cash_requests (
			id, portfolio_id, fund_id,
			transaction_type, amount, currency, fees, value_date, memo,
			status, approval_request_id, resulting_txn_id, idempotency_key, request_fingerprint,
			submitted_by, submitted_at, decided_by, decided_at,
			version, created_at, updated_at, created_by, updated_by
		) VALUES (
			$1, $2, $3,
			$4, $5, $6, $7, $8, NULLIF($9,''),
			$10, $11, $12, $13, $14,
			$15, $16, $17, $18,
			$19, $20, $21, $22, $23
		)`,
		c.ID, c.PortfolioID, c.FundID,
		string(c.TransactionType), c.Amount, c.Currency, c.Fees, c.ValueDate, c.Memo,
		string(c.Status), c.ApprovalRequestID, c.ResultingTxnID, c.IdempotencyKey, c.RequestFingerprint,
		c.SubmittedBy, c.SubmittedAt, c.DecidedBy, c.DecidedAt,
		c.Version, c.CreatedAt, c.UpdatedAt, c.CreatedBy, c.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting cash request: %w", err)
	}
	return nil
}

func (r *PostgresCashRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.PortfolioCashRequest, error) {
	row := r.pool.QueryRow(ctx, cashReqSelect+" WHERE id = $1", id)
	return scanCashRequest(row)
}

func (r *PostgresCashRequestRepository) GetByIdempotencyKey(
	ctx context.Context,
	portfolioID uuid.UUID,
	key string,
) (*entity.PortfolioCashRequest, error) {
	row := r.pool.QueryRow(ctx,
		cashReqSelect+" WHERE portfolio_id = $1 AND idempotency_key = $2", portfolioID, key)
	return scanCashRequest(row)
}

// GetByFingerprint returns the earliest request for a (portfolioID, fingerprint)
// pair or nil when none exists (read via pool). The fingerprint index is NOT
// unique — two distinct client keys may carry the same payload — so this is a
// reconciliation-only lookup and deterministically returns the FIRST-submitted
// match (ORDER BY submitted_at, id). Race-safe dedup is enforced by the
// (portfolio_id, idempotency_key) partial unique index, not by this method. The
// caller must not pass an empty fingerprint.
func (r *PostgresCashRequestRepository) GetByFingerprint(
	ctx context.Context,
	portfolioID uuid.UUID,
	fingerprint string,
) (*entity.PortfolioCashRequest, error) {
	row := r.pool.QueryRow(ctx,
		cashReqSelect+" WHERE portfolio_id = $1 AND request_fingerprint = $2"+
			" ORDER BY submitted_at ASC, id ASC LIMIT 1", portfolioID, fingerprint)
	return scanCashRequest(row)
}

func (r *PostgresCashRequestRepository) GetForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.PortfolioCashRequest, error) {
	row := tx.QueryRow(ctx, cashReqSelect+" WHERE id = $1 FOR UPDATE", id)
	return scanCashRequest(row)
}

func (r *PostgresCashRequestRepository) Update(ctx context.Context, tx pgx.Tx, c *entity.PortfolioCashRequest) error {
	// version is bumped here; the caller holds a FOR UPDATE lock (approval
	// callback) or is the sole writer (cancel path), so a plain id-keyed update
	// is race-safe.
	ct, err := tx.Exec(ctx, `
		UPDATE investment__portfolio_cash_requests
		   SET status = $2,
		       approval_request_id = $3,
		       resulting_txn_id = $4,
		       decided_by = $5,
		       decided_at = $6,
		       memo = NULLIF($7,''),
		       version = version + 1,
		       updated_by = $8,
		       updated_at = $9
		 WHERE id = $1`,
		c.ID,
		string(c.Status), c.ApprovalRequestID, c.ResultingTxnID,
		c.DecidedBy, c.DecidedAt, c.Memo,
		c.UpdatedBy, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating cash request: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("updating cash request: no row for id %s", c.ID)
	}
	c.Version++
	return nil
}

func (r *PostgresCashRequestRepository) ListByPortfolio(
	ctx context.Context,
	portfolioID uuid.UUID,
	f domain.CashRequestListFilter,
) ([]*entity.PortfolioCashRequest, int, error) {
	args := []any{portfolioID}
	where := "WHERE portfolio_id = $1"
	if f.Status != nil {
		args = append(args, string(*f.Status))
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}

	var total int
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM investment__portfolio_cash_requests "+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting cash requests: %w", err)
	}

	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := 0
	if f.Page > 1 {
		offset = (f.Page - 1) * limit
	}
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx,
		cashReqSelect+" "+where+
			fmt.Sprintf(" ORDER BY submitted_at DESC, id ASC LIMIT $%d OFFSET $%d", len(args)-1, len(args)),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing cash requests: %w", err)
	}
	defer rows.Close()

	out := make([]*entity.PortfolioCashRequest, 0, limit)
	for rows.Next() {
		c, err := scanCashRequest(rows)
		if err != nil {
			return nil, 0, err
		}
		if c != nil {
			out = append(out, c)
		}
	}
	return out, total, rows.Err()
}

func scanCashRequest(s scanner) (*entity.PortfolioCashRequest, error) {
	var c entity.PortfolioCashRequest
	var typeStr, statusStr string
	err := s.Scan(
		&c.ID, &c.PortfolioID, &c.FundID,
		&typeStr, &c.Amount, &c.Currency, &c.Fees, &c.ValueDate, &c.Memo,
		&statusStr, &c.ApprovalRequestID, &c.ResultingTxnID, &c.IdempotencyKey, &c.RequestFingerprint,
		&c.SubmittedBy, &c.SubmittedAt, &c.DecidedBy, &c.DecidedAt,
		&c.Version, &c.CreatedAt, &c.UpdatedAt, &c.CreatedBy, &c.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning cash request: %w", err)
	}
	c.TransactionType = vo.TransactionType(typeStr)
	c.Status = vo.CashRequestStatus(statusStr)
	return &c, nil
}

var _ domain.PortfolioCashRequestRepository = (*PostgresCashRequestRepository)(nil)
