package domain

import (
	"context"
	"time"
)

const (
	ImportTypeQuoteSync           = "QUOTE_SYNC"
	ImportTypeHistorySync         = "HISTORY_SYNC"
	ImportTypeQuoteAndHistorySync = "QUOTE_AND_HISTORY_SYNC"
	ImportTypeManualFile          = "MANUAL_FILE"
)

const (
	ImportBatchStatusPending               = "PENDING"
	ImportBatchStatusRunning               = "RUNNING"
	ImportBatchStatusCompleted             = "COMPLETED"
	ImportBatchStatusCompletedWithWarnings = "COMPLETED_WITH_WARNINGS"
	ImportBatchStatusPartialFailed         = "PARTIAL_FAILED"
	ImportBatchStatusFailed                = "FAILED"
	ImportBatchStatusCancelled             = "CANCELLED"
)

const (
	ImportChunkStatusPending               = "PENDING"
	ImportChunkStatusRunning               = "RUNNING"
	ImportChunkStatusCompleted             = "COMPLETED"
	ImportChunkStatusCompletedWithWarnings = "COMPLETED_WITH_WARNINGS"
	ImportChunkStatusFailed                = "FAILED"
	ImportChunkStatusRateLimited           = "RATE_LIMITED"
	ImportChunkStatusCancelled             = "CANCELLED"
)

const (
	ImportItemStatusPending        = "PENDING"
	ImportItemStatusAccepted       = "ACCEPTED"
	ImportItemStatusRejected       = "REJECTED"
	ImportItemStatusWarning        = "WARNING"
	ImportItemStatusRateLimited    = "RATE_LIMITED"
	ImportItemStatusUnmapped       = "UNMAPPED"
	ImportItemStatusReviewRequired = "REVIEW_REQUIRED"
	ImportItemStatusFailed         = "FAILED"
)

const (
	ImportErrorCodeRateLimited     = "RATE_LIMITED"
	ImportErrorCodeUnmapped        = "UNMAPPED"
	ImportErrorCodeReviewRequired  = "REVIEW_REQUIRED"
	ImportErrorCodeProviderFailure = "PROVIDER_FAILURE"
	ImportErrorCodeTimeout         = "TIMEOUT"
)

type ImportBatch struct {
	ID              string     `json:"batch_id"`
	ProviderCode    string     `json:"provider"`
	ImportType      string     `json:"import_type"`
	Status          string     `json:"status"`
	IdempotencyKey  string     `json:"idempotency_key,omitempty"`
	TotalRecords    int        `json:"total_symbols"`
	AcceptedRecords int        `json:"accepted_records"`
	RejectedRecords int        `json:"rejected_records"`
	WarningRecords  int        `json:"warning_records"`
	TotalChunks     int        `json:"total_chunks"`
	CreatedBy       string     `json:"created_by,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type ImportChunk struct {
	ID              string     `json:"chunk_id"`
	BatchID         string     `json:"batch_id"`
	ChunkIndex      int        `json:"chunk_index"`
	Status          string     `json:"status"`
	TotalRecords    int        `json:"total_records"`
	AcceptedRecords int        `json:"accepted_records"`
	RejectedRecords int        `json:"rejected_records"`
	WarningRecords  int        `json:"warning_records"`
	AttemptCount    int        `json:"attempt_count"`
	LockedAt        *time.Time `json:"locked_at,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	Symbols         []string   `json:"symbols,omitempty"`
}

type ImportChunkItem struct {
	ID             string    `json:"item_id"`
	ChunkID        string    `json:"chunk_id"`
	SecurityID     string    `json:"security_id,omitempty"`
	Symbol         string    `json:"symbol"`
	ProviderSymbol string    `json:"provider_symbol,omitempty"`
	Status         string    `json:"status"`
	ErrorCode      string    `json:"error_code,omitempty"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type ImportBatchPlan struct {
	ProviderCode   string
	ImportType     string
	IncludeQuote   bool
	IncludeHistory bool
	HistoryLimit   int
	IdempotencyKey string
	CreatedBy      string
	Chunks         []ImportChunkPlan
}

type ImportChunkPlan struct {
	ChunkIndex int
	Symbols    []string
}

type ImportBatchRepository interface {
	CreateBatch(ctx context.Context, plan ImportBatchPlan) (*ImportBatch, []ImportChunk, error)
	FindBatchByIdempotencyKey(ctx context.Context, key string) (*ImportBatch, error)
	GetBatch(ctx context.Context, batchID string) (*ImportBatch, error)
	ListChunks(ctx context.Context, batchID string) ([]ImportChunk, error)
	ListChunkItems(ctx context.Context, chunkID string) ([]ImportChunkItem, error)
	ListBatchErrors(ctx context.Context, batchID string) ([]ImportChunkItem, error)
	UpdateBatchStatus(ctx context.Context, batchID string, update ImportBatchStatusUpdate) error
	UpdateChunkStatus(ctx context.Context, chunkID string, update ImportChunkStatusUpdate) error
	RecordChunkItem(ctx context.Context, item ImportChunkItem) error
	ResetChunkItems(ctx context.Context, chunkID string) error
}

type ImportBatchStatusUpdate struct {
	Status          string
	AcceptedRecords *int
	RejectedRecords *int
	WarningRecords  *int
	StartedAt       *time.Time
	CompletedAt     *time.Time
	ErrorMessage    *string
}

type ImportChunkStatusUpdate struct {
	Status          string
	AcceptedRecords *int
	RejectedRecords *int
	WarningRecords  *int
	AttemptCount    *int
	LockedAt        *time.Time
	StartedAt       *time.Time
	CompletedAt     *time.Time
	ErrorMessage    *string
}
