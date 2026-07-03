-- Table: market_data_import_chunk_items
-- Source: 20260526000001_marketdata__create_import_batches.up.sql
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
