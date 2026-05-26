package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
)

func (r *PostgresRepository) CreateBatch(ctx context.Context, plan domain.ImportBatchPlan) (*domain.ImportBatch, []domain.ImportChunk, error) {
	if r == nil || r.pool == nil {
		return nil, nil, fmt.Errorf("postgres pool not initialised")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("beginning import batch tx: %w", err)
	}
	defer tx.Rollback(ctx)

	total := 0
	for _, c := range plan.Chunks {
		total += len(c.Symbols)
	}

	var (
		batchID           string
		batchCreatedAt    time.Time
		createdByNullable any
		idempotencyKeyArg any
	)
	if plan.CreatedBy != "" {
		createdByNullable = plan.CreatedBy
	}
	if plan.IdempotencyKey != "" {
		idempotencyKeyArg = plan.IdempotencyKey
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO market_data_import_batches (
			provider_code, import_type, status, idempotency_key, total_records, created_by
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`,
		plan.ProviderCode,
		plan.ImportType,
		domain.ImportBatchStatusPending,
		idempotencyKeyArg,
		total,
		createdByNullable,
	).Scan(&batchID, &batchCreatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("inserting import batch: %w", err)
	}

	chunks := make([]domain.ImportChunk, 0, len(plan.Chunks))
	for _, c := range plan.Chunks {
		var (
			chunkID        string
			chunkCreatedAt time.Time
		)
		err = tx.QueryRow(ctx, `
			INSERT INTO market_data_import_chunks (
				batch_id, chunk_index, status, total_records
			) VALUES ($1, $2, $3, $4)
			RETURNING id, created_at`,
			batchID,
			c.ChunkIndex,
			domain.ImportChunkStatusPending,
			len(c.Symbols),
		).Scan(&chunkID, &chunkCreatedAt)
		if err != nil {
			return nil, nil, fmt.Errorf("inserting import chunk: %w", err)
		}
		for _, symbol := range c.Symbols {
			_, err = tx.Exec(ctx, `
				INSERT INTO market_data_import_chunk_items (
					chunk_id, symbol, status
				) VALUES ($1, $2, $3)`,
				chunkID,
				symbol,
				domain.ImportItemStatusPending,
			)
			if err != nil {
				return nil, nil, fmt.Errorf("inserting import chunk item: %w", err)
			}
		}
		chunks = append(chunks, domain.ImportChunk{
			ID:           chunkID,
			BatchID:      batchID,
			ChunkIndex:   c.ChunkIndex,
			Status:       domain.ImportChunkStatusPending,
			TotalRecords: len(c.Symbols),
			CreatedAt:    chunkCreatedAt,
			Symbols:      append([]string(nil), c.Symbols...),
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("committing import batch tx: %w", err)
	}

	batch := &domain.ImportBatch{
		ID:             batchID,
		ProviderCode:   plan.ProviderCode,
		ImportType:     plan.ImportType,
		Status:         domain.ImportBatchStatusPending,
		IdempotencyKey: plan.IdempotencyKey,
		TotalRecords:   total,
		TotalChunks:    len(chunks),
		CreatedBy:      plan.CreatedBy,
		CreatedAt:      batchCreatedAt,
	}
	return batch, chunks, nil
}

func (r *PostgresRepository) FindBatchByIdempotencyKey(ctx context.Context, key string) (*domain.ImportBatch, error) {
	if r == nil || r.pool == nil || key == "" {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, `
		SELECT id, provider_code, import_type, status, COALESCE(idempotency_key, ''),
		       total_records, accepted_records, rejected_records, warning_records,
		       COALESCE(created_by::text, ''), started_at, completed_at,
		       COALESCE(error_message, ''), created_at
		  FROM market_data_import_batches
		 WHERE idempotency_key = $1
		 LIMIT 1`,
		key,
	)
	batch, err := scanBatch(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	chunkCount, err := r.countChunks(ctx, batch.ID)
	if err != nil {
		return nil, err
	}
	batch.TotalChunks = chunkCount
	return batch, nil
}

func (r *PostgresRepository) GetBatch(ctx context.Context, batchID string) (*domain.ImportBatch, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, `
		SELECT id, provider_code, import_type, status, COALESCE(idempotency_key, ''),
		       total_records, accepted_records, rejected_records, warning_records,
		       COALESCE(created_by::text, ''), started_at, completed_at,
		       COALESCE(error_message, ''), created_at
		  FROM market_data_import_batches
		 WHERE id = $1`,
		batchID,
	)
	batch, err := scanBatch(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	chunkCount, err := r.countChunks(ctx, batchID)
	if err != nil {
		return nil, err
	}
	batch.TotalChunks = chunkCount
	return batch, nil
}

func (r *PostgresRepository) countChunks(ctx context.Context, batchID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM market_data_import_chunks WHERE batch_id = $1`,
		batchID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting import chunks: %w", err)
	}
	return count, nil
}

func (r *PostgresRepository) ListChunks(ctx context.Context, batchID string) ([]domain.ImportChunk, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, batch_id, chunk_index, status, total_records, accepted_records,
		       rejected_records, warning_records, attempt_count, locked_at, started_at,
		       completed_at, COALESCE(error_message, ''), created_at
		  FROM market_data_import_chunks
		 WHERE batch_id = $1
		 ORDER BY chunk_index ASC`,
		batchID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing import chunks: %w", err)
	}
	defer rows.Close()
	out := []domain.ImportChunk{}
	for rows.Next() {
		var c domain.ImportChunk
		if err := rows.Scan(
			&c.ID, &c.BatchID, &c.ChunkIndex, &c.Status, &c.TotalRecords, &c.AcceptedRecords,
			&c.RejectedRecords, &c.WarningRecords, &c.AttemptCount, &c.LockedAt, &c.StartedAt,
			&c.CompletedAt, &c.ErrorMessage, &c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning import chunk: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) ListChunkItems(ctx context.Context, chunkID string) ([]domain.ImportChunkItem, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, chunk_id, COALESCE(security_id::text, ''), symbol,
		       COALESCE(provider_symbol, ''), status, COALESCE(error_code, ''),
		       COALESCE(error_message, ''), created_at
		  FROM market_data_import_chunk_items
		 WHERE chunk_id = $1
		 ORDER BY created_at ASC`,
		chunkID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing import chunk items: %w", err)
	}
	defer rows.Close()
	out := []domain.ImportChunkItem{}
	for rows.Next() {
		var item domain.ImportChunkItem
		if err := rows.Scan(
			&item.ID, &item.ChunkID, &item.SecurityID, &item.Symbol,
			&item.ProviderSymbol, &item.Status, &item.ErrorCode,
			&item.ErrorMessage, &item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning import chunk item: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) ListBatchErrors(ctx context.Context, batchID string) ([]domain.ImportChunkItem, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT i.id, i.chunk_id, COALESCE(i.security_id::text, ''), i.symbol,
		       COALESCE(i.provider_symbol, ''), i.status, COALESCE(i.error_code, ''),
		       COALESCE(i.error_message, ''), i.created_at
		  FROM market_data_import_chunk_items i
		  JOIN market_data_import_chunks c ON c.id = i.chunk_id
		 WHERE c.batch_id = $1
		   AND i.status IN ('REJECTED','FAILED','RATE_LIMITED','UNMAPPED','REVIEW_REQUIRED','WARNING')
		 ORDER BY i.created_at ASC`,
		batchID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing import batch errors: %w", err)
	}
	defer rows.Close()
	out := []domain.ImportChunkItem{}
	for rows.Next() {
		var item domain.ImportChunkItem
		if err := rows.Scan(
			&item.ID, &item.ChunkID, &item.SecurityID, &item.Symbol,
			&item.ProviderSymbol, &item.Status, &item.ErrorCode,
			&item.ErrorMessage, &item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning import error row: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) UpdateBatchStatus(ctx context.Context, batchID string, update domain.ImportBatchStatusUpdate) error {
	if r == nil || r.pool == nil {
		return nil
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE market_data_import_batches
		   SET status = COALESCE(NULLIF($2, ''), status),
		       accepted_records = COALESCE($3, accepted_records),
		       rejected_records = COALESCE($4, rejected_records),
		       warning_records = COALESCE($5, warning_records),
		       started_at = COALESCE($6, started_at),
		       completed_at = COALESCE($7, completed_at),
		       error_message = COALESCE($8, error_message)
		 WHERE id = $1`,
		batchID,
		update.Status,
		intPtr(update.AcceptedRecords),
		intPtr(update.RejectedRecords),
		intPtr(update.WarningRecords),
		timePtr(update.StartedAt),
		timePtr(update.CompletedAt),
		stringPtr(update.ErrorMessage),
	)
	if err != nil {
		return fmt.Errorf("updating import batch status: %w", err)
	}
	return nil
}

func (r *PostgresRepository) UpdateChunkStatus(ctx context.Context, chunkID string, update domain.ImportChunkStatusUpdate) error {
	if r == nil || r.pool == nil {
		return nil
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE market_data_import_chunks
		   SET status = COALESCE(NULLIF($2, ''), status),
		       accepted_records = COALESCE($3, accepted_records),
		       rejected_records = COALESCE($4, rejected_records),
		       warning_records = COALESCE($5, warning_records),
		       attempt_count = COALESCE($6, attempt_count),
		       locked_at = COALESCE($7, locked_at),
		       started_at = COALESCE($8, started_at),
		       completed_at = COALESCE($9, completed_at),
		       error_message = COALESCE($10, error_message)
		 WHERE id = $1`,
		chunkID,
		update.Status,
		intPtr(update.AcceptedRecords),
		intPtr(update.RejectedRecords),
		intPtr(update.WarningRecords),
		intPtr(update.AttemptCount),
		timePtr(update.LockedAt),
		timePtr(update.StartedAt),
		timePtr(update.CompletedAt),
		stringPtr(update.ErrorMessage),
	)
	if err != nil {
		return fmt.Errorf("updating import chunk status: %w", err)
	}
	return nil
}

func (r *PostgresRepository) RecordChunkItem(ctx context.Context, item domain.ImportChunkItem) error {
	if r == nil || r.pool == nil {
		return nil
	}
	var providerSymbol any
	if item.ProviderSymbol != "" {
		providerSymbol = item.ProviderSymbol
	}
	var errorCode any
	if item.ErrorCode != "" {
		errorCode = item.ErrorCode
	}
	var errorMessage any
	if item.ErrorMessage != "" {
		errorMessage = item.ErrorMessage
	}
	var securityID any
	if item.SecurityID != "" {
		securityID = item.SecurityID
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE market_data_import_chunk_items
		   SET status = $2,
		       error_code = $3,
		       error_message = $4,
		       provider_symbol = COALESCE($5, provider_symbol),
		       security_id = COALESCE($6::uuid, security_id)
		 WHERE chunk_id = $7 AND symbol = $1`,
		item.Symbol,
		item.Status,
		errorCode,
		errorMessage,
		providerSymbol,
		securityID,
		item.ChunkID,
	)
	if err != nil {
		return fmt.Errorf("recording import chunk item: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ResetChunkItems(ctx context.Context, chunkID string) error {
	if r == nil || r.pool == nil {
		return nil
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE market_data_import_chunk_items
		   SET status = $2,
		       error_code = NULL,
		       error_message = NULL
		 WHERE chunk_id = $1`,
		chunkID,
		domain.ImportItemStatusPending,
	)
	if err != nil {
		return fmt.Errorf("resetting import chunk items: %w", err)
	}
	return nil
}

func scanBatch(row pgx.Row) (*domain.ImportBatch, error) {
	var b domain.ImportBatch
	err := row.Scan(
		&b.ID, &b.ProviderCode, &b.ImportType, &b.Status, &b.IdempotencyKey,
		&b.TotalRecords, &b.AcceptedRecords, &b.RejectedRecords, &b.WarningRecords,
		&b.CreatedBy, &b.StartedAt, &b.CompletedAt, &b.ErrorMessage, &b.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func intPtr(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func timePtr(v *time.Time) any {
	if v == nil {
		return nil
	}
	return *v
}

func stringPtr(v *string) any {
	if v == nil {
		return nil
	}
	return *v
}

var _ domain.ImportBatchRepository = (*PostgresRepository)(nil)
