package entity

import (
	"time"

	"github.com/google/uuid"
)

// TradeConfirmationImportBatchStatus enumerates the persisted batch status
// values mirrored by the DB CHECK chk_inv_confirmation_import_status.
type TradeConfirmationImportBatchStatus string

const (
	TradeConfirmationImportBatchCompleted          TradeConfirmationImportBatchStatus = "COMPLETED"
	TradeConfirmationImportBatchCompletedWithError TradeConfirmationImportBatchStatus = "COMPLETED_WITH_ERRORS"
	TradeConfirmationImportBatchFailed             TradeConfirmationImportBatchStatus = "FAILED"
)

// TradeConfirmationImportItemStatus enumerates per-row outcomes.
type TradeConfirmationImportItemStatus string

const (
	TradeConfirmationImportItemAccepted TradeConfirmationImportItemStatus = "ACCEPTED"
	TradeConfirmationImportItemRejected TradeConfirmationImportItemStatus = "REJECTED"
)

// TradeConfirmationImportBatch is the header of a broker EOD / file-driven
// trade-confirmation import. Acts as the audit anchor for the per-row items.
type TradeConfirmationImportBatch struct {
	ID              uuid.UUID
	SourceFilename  string
	Status          TradeConfirmationImportBatchStatus
	TotalRecords    int
	AcceptedRecords int
	RejectedRecords int
	ErrorMessage    string
	CreatedBy       uuid.UUID
	CreatedAt       time.Time
	CompletedAt     *time.Time
}

// TradeConfirmationImportItem records the outcome of one row inside a batch.
// ConfirmationID is populated when Status == ACCEPTED.
type TradeConfirmationImportItem struct {
	ID             uuid.UUID
	BatchID        uuid.UUID
	RowIndex       int
	Status         TradeConfirmationImportItemStatus
	ErrorMessage   string
	ConfirmationID *uuid.UUID
	RawPayload     []byte
	CreatedAt      time.Time
}
