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
)

// PostgresTradeConfirmationImportRepository persists
// investment__trade_confirmation_import_batches and ..._items.
type PostgresTradeConfirmationImportRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTradeConfirmationImportRepository(pool *pgxpool.Pool) *PostgresTradeConfirmationImportRepository {
	return &PostgresTradeConfirmationImportRepository{pool: pool}
}

func (r *PostgresTradeConfirmationImportRepository) CreateBatch(ctx context.Context, tx pgx.Tx, b *entity.TradeConfirmationImportBatch) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__trade_confirmation_import_batches (
			id, source_filename, status,
			total_records, accepted_records, rejected_records,
			error_message, created_by, created_at, completed_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6,
			$7, $8, $9, $10
		)`,
		b.ID, b.SourceFilename, string(b.Status),
		b.TotalRecords, b.AcceptedRecords, b.RejectedRecords,
		b.ErrorMessage, b.CreatedBy, b.CreatedAt, b.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("insert import batch: %w", err)
	}
	return nil
}

func (r *PostgresTradeConfirmationImportRepository) UpdateBatchSummary(ctx context.Context, tx pgx.Tx, b *entity.TradeConfirmationImportBatch) error {
	tag, err := tx.Exec(ctx, `
		UPDATE investment__trade_confirmation_import_batches
		   SET status            = $2,
		       total_records     = $3,
		       accepted_records  = $4,
		       rejected_records  = $5,
		       error_message     = $6,
		       completed_at      = $7
		 WHERE id = $1`,
		b.ID, string(b.Status),
		b.TotalRecords, b.AcceptedRecords, b.RejectedRecords,
		b.ErrorMessage, b.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("update import batch summary: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("import batch %s missing during summary update", b.ID)
	}
	return nil
}

func (r *PostgresTradeConfirmationImportRepository) CreateItem(ctx context.Context, tx pgx.Tx, i *entity.TradeConfirmationImportItem) error {
	payload := i.RawPayload
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__trade_confirmation_import_items (
			id, batch_id, row_index, status, error_message,
			confirmation_id, raw_payload, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		i.ID, i.BatchID, i.RowIndex, string(i.Status), i.ErrorMessage,
		i.ConfirmationID, payload, i.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert import item: %w", err)
	}
	return nil
}

func (r *PostgresTradeConfirmationImportRepository) GetBatch(ctx context.Context, id uuid.UUID) (*entity.TradeConfirmationImportBatch, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, source_filename, status,
		       total_records, accepted_records, rejected_records,
		       error_message, created_by, created_at, completed_at
		  FROM investment__trade_confirmation_import_batches
		 WHERE id = $1`, id)

	b := &entity.TradeConfirmationImportBatch{}
	var statusStr string
	err := row.Scan(
		&b.ID, &b.SourceFilename, &statusStr,
		&b.TotalRecords, &b.AcceptedRecords, &b.RejectedRecords,
		&b.ErrorMessage, &b.CreatedBy, &b.CreatedAt, &b.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan import batch: %w", err)
	}
	b.Status = entity.TradeConfirmationImportBatchStatus(statusStr)
	return b, nil
}

func (r *PostgresTradeConfirmationImportRepository) ListBatchItems(ctx context.Context, batchID uuid.UUID) ([]*entity.TradeConfirmationImportItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, batch_id, row_index, status, error_message,
		       confirmation_id, raw_payload, created_at
		  FROM investment__trade_confirmation_import_items
		 WHERE batch_id = $1
		 ORDER BY row_index`, batchID)
	if err != nil {
		return nil, fmt.Errorf("list import items: %w", err)
	}
	defer rows.Close()
	out := []*entity.TradeConfirmationImportItem{}
	for rows.Next() {
		i := &entity.TradeConfirmationImportItem{}
		var statusStr string
		if err := rows.Scan(
			&i.ID, &i.BatchID, &i.RowIndex, &statusStr, &i.ErrorMessage,
			&i.ConfirmationID, &i.RawPayload, &i.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan import item: %w", err)
		}
		i.Status = entity.TradeConfirmationImportItemStatus(statusStr)
		out = append(out, i)
	}
	return out, rows.Err()
}

var _ domain.TradeConfirmationImportRepository = (*PostgresTradeConfirmationImportRepository)(nil)
