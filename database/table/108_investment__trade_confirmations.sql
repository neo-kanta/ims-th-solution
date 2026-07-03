-- Table: investment__trade_confirmations
-- Source: 20260601000001_investment__create_decisions_executions_confirmations.up.sql,
--         updated by 20260604093800_investment__trade_confirmation_import_batches.up.sql,
--         updated by 20260703000001_investment__portfolio_v2_hardening.up.sql
-- Purpose: Broker confirmation reconciled against execution before ledger posting.
-- Identity: database links use portfolio_id. fund_id and contract_id are legacy compatibility fields
--           and must remain equal while they exist.
CREATE TABLE IF NOT EXISTS investment__trade_confirmations (
    -- Identity
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id        UUID            NOT NULL REFERENCES investment__executions(id) ON DELETE RESTRICT,
    decision_id         UUID            NOT NULL REFERENCES investment__decisions(id) ON DELETE RESTRICT,
    portfolio_id        UUID            NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,

    -- Optional fund wrapper / legacy compatibility
    fund_id             UUID            NOT NULL REFERENCES investment__funds(id) ON DELETE RESTRICT,
    contract_id         UUID            NOT NULL,

    -- Confirmation economics
    business_date       DATE            NOT NULL,
    confirmed_quantity  DECIMAL(28,8),
    confirmed_amount    DECIMAL(28,8),
    confirmed_price     DECIMAL(28,8),
    currency            CHAR(3)         NOT NULL,

    -- Broker/import references
    broker_reference    VARCHAR(80),
    import_batch_id     UUID,

    -- Lifecycle / status
    status              VARCHAR(24)     NOT NULL DEFAULT 'PENDING_REVIEW',
    discrepancy_reason  TEXT,
    reviewed_at         TIMESTAMPTZ,
    reviewed_by         UUID            REFERENCES iam_users(id),

    -- Audit
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_by          UUID            NOT NULL REFERENCES iam_users(id),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by          UUID            NOT NULL REFERENCES iam_users(id),

    -- Constraints
    CONSTRAINT chk_inv_confirmation_status
        CHECK (status IN ('PENDING_REVIEW','MATCHED','MISMATCHED','REVIEWED')),

    CONSTRAINT chk_inv_confirmation_currency
        CHECK (currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_confirmation_reason_required
        CHECK (
            status NOT IN ('MISMATCHED','REVIEWED')
            OR (discrepancy_reason IS NOT NULL AND length(trim(discrepancy_reason)) > 0)
        ),

    CONSTRAINT fk_inv_confirmation_import_batch
        FOREIGN KEY (import_batch_id)
        REFERENCES investment__trade_confirmation_import_batches(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_inv_confirmation_portfolio_fund
        FOREIGN KEY (portfolio_id, fund_id)
        REFERENCES investment__portfolios (id, fund_id),

    CONSTRAINT chk_inv_confirmations_contract_is_fund
        CHECK (contract_id = fund_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_inv_confirmation_execution
    ON investment__trade_confirmations (execution_id);

CREATE INDEX IF NOT EXISTS idx_inv_confirmation_decision
    ON investment__trade_confirmations (decision_id);

CREATE INDEX IF NOT EXISTS idx_inv_confirmation_contract_date
    ON investment__trade_confirmations (contract_id, business_date DESC);

CREATE INDEX IF NOT EXISTS idx_inv_confirmation_status
    ON investment__trade_confirmations (status);

-- Comments
COMMENT ON TABLE investment__trade_confirmations IS
    'Broker confirmation reconciled against the execution. Required for transaction closing.';
