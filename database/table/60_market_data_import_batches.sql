-- Table: market_data_import_batches
-- Source: 20260526000001_marketdata__create_import_batches.up.sql
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
