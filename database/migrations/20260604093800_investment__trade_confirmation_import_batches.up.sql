-- =============================================================================
-- Investment Trade Confirmation Import Batches
-- =============================================================================
-- Tracks broker EOD / file-driven imports of trade confirmations. One batch
-- groups N rows; each successful row inserts an investment__trade_confirmations
-- row with import_batch_id pointing at this table.
--
-- Per-row outcomes are stored in investment__trade_confirmation_import_items
-- so failures are inspectable without re-running the import.
-- =============================================================================

CREATE TABLE IF NOT EXISTS investment__trade_confirmation_import_batches (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    source_filename     VARCHAR(255) NOT NULL DEFAULT '',
    status              VARCHAR(30)  NOT NULL DEFAULT 'COMPLETED',
    total_records       INTEGER      NOT NULL DEFAULT 0,
    accepted_records    INTEGER      NOT NULL DEFAULT 0,
    rejected_records    INTEGER      NOT NULL DEFAULT 0,
    error_message       TEXT         NOT NULL DEFAULT '',
    created_by          UUID         NOT NULL REFERENCES iam_users(id),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ,

    CONSTRAINT chk_inv_confirmation_import_status
        CHECK (status IN (
            'COMPLETED',
            'COMPLETED_WITH_ERRORS',
            'FAILED'
        )),
    CONSTRAINT chk_inv_confirmation_import_counts
        CHECK (
            total_records    >= 0
            AND accepted_records >= 0
            AND rejected_records >= 0
            AND accepted_records + rejected_records <= total_records
        )
);

CREATE INDEX IF NOT EXISTS idx_inv_confirmation_import_batches_created_at
    ON investment__trade_confirmation_import_batches (created_at DESC);

COMMENT ON TABLE  investment__trade_confirmation_import_batches IS
    'Broker trade-confirmation import jobs. One row per import call; per-row results live in investment__trade_confirmation_import_items.';
COMMENT ON COLUMN investment__trade_confirmation_import_batches.source_filename IS
    'Originating filename or remote source identifier (free text; empty for ad-hoc JSON imports).';

CREATE TABLE IF NOT EXISTS investment__trade_confirmation_import_items (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id            UUID         NOT NULL REFERENCES investment__trade_confirmation_import_batches(id) ON DELETE CASCADE,
    row_index           INTEGER      NOT NULL,
    status              VARCHAR(30)  NOT NULL,
    error_message       TEXT         NOT NULL DEFAULT '',
    confirmation_id     UUID         REFERENCES investment__trade_confirmations(id) ON DELETE SET NULL,
    raw_payload         JSONB        NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_inv_confirmation_import_items_batch_row
        UNIQUE (batch_id, row_index),
    CONSTRAINT chk_inv_confirmation_import_item_status
        CHECK (status IN ('ACCEPTED', 'REJECTED'))
);

CREATE INDEX IF NOT EXISTS idx_inv_confirmation_import_items_batch_status
    ON investment__trade_confirmation_import_items (batch_id, status);

COMMENT ON TABLE investment__trade_confirmation_import_items IS
    'Per-row outcome of a trade-confirmation batch import. raw_payload preserves the submitted row body for forensic replay.';

-- Hook the existing investment__trade_confirmations.import_batch_id to the new
-- batches table. ON DELETE SET NULL preserves historical confirmations if a
-- batch row is removed for any reason; the import items table cascades.
ALTER TABLE investment__trade_confirmations
    ADD CONSTRAINT fk_inv_confirmation_import_batch
        FOREIGN KEY (import_batch_id)
        REFERENCES investment__trade_confirmation_import_batches(id)
        ON DELETE SET NULL;
