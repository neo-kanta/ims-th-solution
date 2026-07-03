-- =============================================================================
-- Market Data Import Batches
-- =============================================================================
-- Tracks chunked market data import jobs. Each batch covers a list of symbols
-- split into chunks of bounded size that can be retried and processed
-- independently.
-- =============================================================================

CREATE TABLE market_data_import_batches (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_code      VARCHAR(40)  NOT NULL,
    import_type        VARCHAR(40)  NOT NULL,
    status             VARCHAR(30)  NOT NULL DEFAULT 'PENDING',
    idempotency_key    VARCHAR(255),
    total_records      INTEGER      NOT NULL DEFAULT 0,
    accepted_records   INTEGER      NOT NULL DEFAULT 0,
    rejected_records   INTEGER      NOT NULL DEFAULT 0,
    warning_records    INTEGER      NOT NULL DEFAULT 0,
    created_by         UUID,
    started_at         TIMESTAMPTZ,
    completed_at       TIMESTAMPTZ,
    error_message      TEXT,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_market_data_import_batches_status
        CHECK (status IN (
            'PENDING',
            'RUNNING',
            'COMPLETED',
            'COMPLETED_WITH_WARNINGS',
            'PARTIAL_FAILED',
            'FAILED',
            'CANCELLED'
        )),
    CONSTRAINT chk_market_data_import_batches_import_type
        CHECK (import_type IN (
            'QUOTE_SYNC',
            'HISTORY_SYNC',
            'QUOTE_AND_HISTORY_SYNC',
            'MANUAL_FILE'
        ))
);

CREATE UNIQUE INDEX uq_market_data_import_batches_idempotency_key
    ON market_data_import_batches (idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX idx_market_data_import_batches_status_created
    ON market_data_import_batches (status, created_at DESC);

COMMENT ON TABLE market_data_import_batches IS 'Chunked market data import jobs (provider sync and manual file imports).';
COMMENT ON COLUMN market_data_import_batches.idempotency_key IS 'Optional client-provided key; duplicate submissions return the existing batch.';

CREATE TABLE market_data_import_chunks (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id           UUID         NOT NULL REFERENCES market_data_import_batches(id) ON DELETE CASCADE,
    chunk_index        INTEGER      NOT NULL,
    status             VARCHAR(30)  NOT NULL DEFAULT 'PENDING',
    total_records      INTEGER      NOT NULL DEFAULT 0,
    accepted_records   INTEGER      NOT NULL DEFAULT 0,
    rejected_records   INTEGER      NOT NULL DEFAULT 0,
    warning_records    INTEGER      NOT NULL DEFAULT 0,
    attempt_count      INTEGER      NOT NULL DEFAULT 0,
    locked_at          TIMESTAMPTZ,
    started_at         TIMESTAMPTZ,
    completed_at       TIMESTAMPTZ,
    error_message      TEXT,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_market_data_import_chunks_batch_index UNIQUE (batch_id, chunk_index),
    CONSTRAINT chk_market_data_import_chunks_status
        CHECK (status IN (
            'PENDING',
            'RUNNING',
            'COMPLETED',
            'COMPLETED_WITH_WARNINGS',
            'FAILED',
            'RATE_LIMITED',
            'CANCELLED'
        ))
);

CREATE INDEX idx_market_data_import_chunks_batch_status
    ON market_data_import_chunks (batch_id, status);

COMMENT ON TABLE market_data_import_chunks IS 'Individual chunks of a market data import batch processed independently.';

CREATE TABLE market_data_import_chunk_items (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    chunk_id           UUID         NOT NULL REFERENCES market_data_import_chunks(id) ON DELETE CASCADE,
    security_id        UUID,
    symbol             VARCHAR(128) NOT NULL,
    provider_symbol    VARCHAR(128),
    status             VARCHAR(30)  NOT NULL DEFAULT 'PENDING',
    error_code         VARCHAR(80),
    error_message      TEXT,
    raw_payload        JSONB        NOT NULL DEFAULT '{}'::jsonb,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_market_data_import_chunk_items_status
        CHECK (status IN (
            'PENDING',
            'ACCEPTED',
            'REJECTED',
            'WARNING',
            'RATE_LIMITED',
            'UNMAPPED',
            'REVIEW_REQUIRED',
            'FAILED'
        ))
);

CREATE INDEX idx_market_data_import_chunk_items_chunk_status
    ON market_data_import_chunk_items (chunk_id, status);

CREATE INDEX idx_market_data_import_chunk_items_symbol
    ON market_data_import_chunk_items (symbol);

COMMENT ON TABLE market_data_import_chunk_items IS 'Per-symbol outcome rows for a market data import chunk.';
COMMENT ON COLUMN market_data_import_chunk_items.raw_payload IS 'Provider raw response or rejected row body. Not exposed via standard frontend APIs.';
