-- Table: market_data_import_chunks
-- Source: 20260526000001_marketdata__create_import_batches.up.sql
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
