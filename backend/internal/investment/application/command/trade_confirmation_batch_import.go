package command

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ConfirmationBatchImportRow is one wire-level row inside a batch import. Each
// row produces a PENDING_REVIEW trade-confirmation when validation passes.
//
// All decimal-looking fields are received as strings to preserve precision
// from the broker file. ExecutionID is required; FundID / ContractID /
// PortfolioID / BusinessDate / Side / Currency are taken from the resolved
// execution row, so callers do not need to supply them.
type ConfirmationBatchImportRow struct {
	ExecutionID       uuid.UUID `json:"execution_id"`
	ConfirmedQuantity string    `json:"confirmed_quantity,omitempty"`
	ConfirmedAmount   string    `json:"confirmed_amount,omitempty"`
	ConfirmedPrice    string    `json:"confirmed_price,omitempty"`
	BrokerReference   string    `json:"broker_reference,omitempty"`
}

// ImportConfirmationBatchRequest is the input to the batch-import command.
type ImportConfirmationBatchRequest struct {
	SourceFilename string
	Rows           []ConfirmationBatchImportRow
	ActorID        uuid.UUID
}

// ConfirmationBatchImportRowResult records one row's outcome.
type ConfirmationBatchImportRowResult struct {
	RowIndex       int
	Accepted       bool
	ConfirmationID *uuid.UUID
	Error          string
}

// ImportConfirmationBatchResult is the response to the importer.
type ImportConfirmationBatchResult struct {
	BatchID         uuid.UUID
	TotalRecords    int
	AcceptedRecords int
	RejectedRecords int
	Status          entity.TradeConfirmationImportBatchStatus
	Rows            []ConfirmationBatchImportRowResult
}

// ConfirmationBatchImportHandler ingests N rows of broker trade confirmations
// in a single batch. Per-row failures are isolated — a bad row does not
// roll back accepted rows; the whole batch row count, accepted count, and
// rejected count are persisted on the batch header so the operator can see
// at a glance what landed and what bounced.
//
// Each accepted row creates a PENDING_REVIEW confirmation pointing at the
// batch via import_batch_id. Mismatch detection and discrepancy-reason
// handling stay on the existing Record/Resolve commands — imports never
// flip a confirmation straight to MATCHED.
type ConfirmationBatchImportHandler struct {
	pool          *pgxpool.Pool
	executions    domain.ExecutionRepository
	confirmations domain.TradeConfirmationRepository
	batches       domain.TradeConfirmationImportRepository
	audit         contract.AuditLogger
	now           func() time.Time
}

func NewConfirmationBatchImportHandler(
	pool *pgxpool.Pool,
	executions domain.ExecutionRepository,
	confirmations domain.TradeConfirmationRepository,
	batches domain.TradeConfirmationImportRepository,
	audit contract.AuditLogger,
	now func() time.Time,
) *ConfirmationBatchImportHandler {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &ConfirmationBatchImportHandler{
		pool:          pool,
		executions:    executions,
		confirmations: confirmations,
		batches:       batches,
		audit:         audit,
		now:           now,
	}
}

// Handle processes the batch. All work runs inside one DB transaction so
// either the whole header + items + accepted confirmations land, or nothing
// does. Audit-error is surfaced after commit — the batch is the financial
// truth even if the audit emit briefly fails.
func (h *ConfirmationBatchImportHandler) Handle(ctx context.Context, req ImportConfirmationBatchRequest) (*ImportConfirmationBatchResult, error) {
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	if len(req.Rows) == 0 {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "rows", Detail: "at least one row is required"}
	}

	now := h.now()
	batch := &entity.TradeConfirmationImportBatch{
		ID:             uuid.New(),
		SourceFilename: strings.TrimSpace(req.SourceFilename),
		Status:         entity.TradeConfirmationImportBatchCompleted,
		TotalRecords:   len(req.Rows),
		CreatedBy:      req.ActorID,
		CreatedAt:      now,
	}

	rowResults := make([]ConfirmationBatchImportRowResult, 0, len(req.Rows))
	withinBatchBrokerRefs := map[string]int{}

	txErr := withTransaction(ctx, h.pool, func(tx pgx.Tx) error {
		if err := h.batches.CreateBatch(ctx, tx, batch); err != nil {
			return err
		}
		for idx, row := range req.Rows {
			payload, _ := json.Marshal(row)
			rowResult := ConfirmationBatchImportRowResult{RowIndex: idx}

			conf, rowErr := h.processRow(ctx, tx, batch, idx, row, withinBatchBrokerRefs)
			if rowErr != nil {
				batch.RejectedRecords++
				rowResult.Accepted = false
				rowResult.Error = rowErr.Error()
				if itemErr := h.batches.CreateItem(ctx, tx, &entity.TradeConfirmationImportItem{
					ID:           uuid.New(),
					BatchID:      batch.ID,
					RowIndex:     idx,
					Status:       entity.TradeConfirmationImportItemRejected,
					ErrorMessage: rowErr.Error(),
					RawPayload:   payload,
					CreatedAt:    now,
				}); itemErr != nil {
					return itemErr
				}
				rowResults = append(rowResults, rowResult)
				continue
			}

			batch.AcceptedRecords++
			rowResult.Accepted = true
			confID := conf.ID
			rowResult.ConfirmationID = &confID
			if itemErr := h.batches.CreateItem(ctx, tx, &entity.TradeConfirmationImportItem{
				ID:             uuid.New(),
				BatchID:        batch.ID,
				RowIndex:       idx,
				Status:         entity.TradeConfirmationImportItemAccepted,
				ConfirmationID: &confID,
				RawPayload:     payload,
				CreatedAt:      now,
			}); itemErr != nil {
				return itemErr
			}
			rowResults = append(rowResults, rowResult)
		}

		// Finalise summary.
		completedAt := h.now()
		batch.CompletedAt = &completedAt
		if batch.RejectedRecords > 0 && batch.AcceptedRecords > 0 {
			batch.Status = entity.TradeConfirmationImportBatchCompletedWithError
		} else if batch.RejectedRecords > 0 && batch.AcceptedRecords == 0 {
			batch.Status = entity.TradeConfirmationImportBatchFailed
		} else {
			batch.Status = entity.TradeConfirmationImportBatchCompleted
		}
		return h.batches.UpdateBatchSummary(ctx, tx, batch)
	})
	if txErr != nil {
		return nil, txErr
	}

	if err := h.audit.LogActionStrict(ctx, contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_CONFIRMATION_BATCH_IMPORTED",
		Module:       "investment",
		ResourceType: "INVESTMENT_TRADE_CONFIRMATION_IMPORT_BATCH",
		ResourceID:   batch.ID.String(),
		Details: map[string]any{
			"source_filename":  batch.SourceFilename,
			"total":            batch.TotalRecords,
			"accepted":         batch.AcceptedRecords,
			"rejected":         batch.RejectedRecords,
			"status":           string(batch.Status),
		},
		BusinessDate: now,
	}); err != nil {
		// Confirmations + items are committed. Surface the audit failure so the
		// caller can retry or alert; the batch row remains visible in DB.
		return nil, fmt.Errorf("auditing batch import: %w", err)
	}

	return &ImportConfirmationBatchResult{
		BatchID:         batch.ID,
		TotalRecords:    batch.TotalRecords,
		AcceptedRecords: batch.AcceptedRecords,
		RejectedRecords: batch.RejectedRecords,
		Status:          batch.Status,
		Rows:            rowResults,
	}, nil
}

// processRow validates and inserts a single row inside the batch transaction.
// Returns the persisted confirmation on success, or an error describing why
// the row was rejected. The error message goes onto the per-row item record
// and into the response — operators see exactly what was wrong without
// re-running the import.
func (h *ConfirmationBatchImportHandler) processRow(
	ctx context.Context,
	tx pgx.Tx,
	batch *entity.TradeConfirmationImportBatch,
	rowIndex int,
	row ConfirmationBatchImportRow,
	withinBatchBrokerRefs map[string]int,
) (*entity.TradeConfirmation, error) {
	if row.ExecutionID == uuid.Nil {
		return nil, fmt.Errorf("execution_id is required")
	}

	exec, err := h.executions.GetByID(ctx, row.ExecutionID)
	if err != nil {
		return nil, fmt.Errorf("loading execution: %w", err)
	}
	if exec == nil {
		return nil, fmt.Errorf("execution %s not found", row.ExecutionID)
	}
	if exec.Status == vo.ExecutionStatusCancelled {
		return nil, fmt.Errorf("cannot confirm a cancelled execution %s", row.ExecutionID)
	}

	q, err := parseOptionalDecimal(row.ConfirmedQuantity, "confirmed_quantity")
	if err != nil {
		return nil, err
	}
	amt, err := parseOptionalDecimal(row.ConfirmedAmount, "confirmed_amount")
	if err != nil {
		return nil, err
	}
	price, err := parseOptionalDecimal(row.ConfirmedPrice, "confirmed_price")
	if err != nil {
		return nil, err
	}

	brokerRef := strings.TrimSpace(row.BrokerReference)
	if brokerRef != "" {
		if priorIdx, ok := withinBatchBrokerRefs[brokerRef]; ok {
			return nil, fmt.Errorf("duplicate broker_reference within batch (also at row %d)", priorIdx)
		}
		existing, err := h.confirmations.GetByBrokerReference(ctx, brokerRef)
		if err != nil {
			return nil, fmt.Errorf("dedupe check failed: %w", err)
		}
		if existing != nil {
			return nil, fmt.Errorf("broker_reference %q already recorded on confirmation %s", brokerRef, existing.ID)
		}
		withinBatchBrokerRefs[brokerRef] = rowIndex
	}

	c := &entity.TradeConfirmation{
		ID:                uuid.New(),
		ExecutionID:       exec.ID,
		DecisionID:        exec.DecisionID,
		FundID:            exec.FundID,
		PortfolioID:       exec.PortfolioID,
		ContractID:        exec.ContractID,
		BusinessDate:      exec.BusinessDate,
		ConfirmedQuantity: q,
		ConfirmedAmount:   amt,
		ConfirmedPrice:    price,
		Currency:          exec.Currency,
		BrokerReference:   brokerRef,
		ImportBatchID:     &batch.ID,
		Status:            vo.TradeConfirmationPendingReview,
		CreatedAt:         batch.CreatedAt,
		CreatedBy:         batch.CreatedBy,
		UpdatedAt:         batch.CreatedAt,
		UpdatedBy:         batch.CreatedBy,
	}
	if err := h.confirmations.Create(ctx, tx, c); err != nil {
		return nil, fmt.Errorf("create confirmation: %w", err)
	}
	return c, nil
}

func parseOptionalDecimal(raw, field string) (*decimal.Decimal, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	v, err := decimal.NewFromString(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid decimal %q", field, raw)
	}
	return &v, nil
}
