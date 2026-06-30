-- Table: investment__trade_confirmation_import_batches
-- Source: 20260604093800_investment__trade_confirmation_import_batches.up.sql
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
            total_records >= 0
            AND accepted_records >= 0
            AND rejected_records >= 0
            AND accepted_records + rejected_records <= total_records
        )
);
