-- Table: investment__trade_confirmation_import_items
-- Source: 20260604093800_investment__trade_confirmation_import_batches.up.sql
CREATE TABLE IF NOT EXISTS investment__trade_confirmation_import_items (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id        UUID        NOT NULL REFERENCES investment__trade_confirmation_import_batches(id) ON DELETE CASCADE,
    row_index       INTEGER     NOT NULL,
    status          VARCHAR(30) NOT NULL,
    error_message   TEXT        NOT NULL DEFAULT '',
    confirmation_id UUID        REFERENCES investment__trade_confirmations(id) ON DELETE SET NULL,
    raw_payload     JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_inv_confirmation_import_items_batch_row
        UNIQUE (batch_id, row_index),
    CONSTRAINT chk_inv_confirmation_import_item_status
        CHECK (status IN ('ACCEPTED', 'REJECTED'))
);
