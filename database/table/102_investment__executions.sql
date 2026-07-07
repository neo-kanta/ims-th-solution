-- Table: investment__executions
-- Source: 20260601000001_investment__create_decisions_executions_confirmations.up.sql,
--         updated by 20260703000001_investment__portfolio_v2_hardening.up.sql,
--         updated by 20260703000002_investment__drop_contract_id_from_operational_tables.up.sql
-- Purpose: Execution record for an approved investment decision. Portfolio is the operational source of truth.
-- Identity: database links use portfolio_id. fund_id is a legacy fund-wrapper compatibility field.
CREATE TABLE IF NOT EXISTS investment__executions (
    -- Identity
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_id         UUID            NOT NULL REFERENCES investment__decisions(id) ON DELETE RESTRICT,
    portfolio_id        UUID            NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,

    -- Optional fund wrapper / legacy compatibility
    fund_id             UUID            NOT NULL REFERENCES investment__funds(id) ON DELETE RESTRICT,

    -- Instrument / order
    instrument_id       UUID            REFERENCES investment__instruments(id) ON DELETE RESTRICT,
    instrument_code     VARCHAR(40)     NOT NULL,
    business_date       DATE            NOT NULL,
    side                VARCHAR(10)     NOT NULL,
    ordered_quantity    DECIMAL(28,8),
    ordered_amount      DECIMAL(28,8),
    executed_quantity   DECIMAL(28,8),
    executed_amount     DECIMAL(28,8),
    execution_price     DECIMAL(28,8),
    currency            CHAR(3)         NOT NULL,

    -- Lifecycle / status
    status              VARCHAR(24)     NOT NULL DEFAULT 'PENDING',
    trader_user_id      UUID            REFERENCES iam_users(id),
    broker_reference    VARCHAR(80),
    executed_at         TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,
    cancelled_by        UUID            REFERENCES iam_users(id),
    cancellation_reason TEXT,

    -- Audit
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_by          UUID            NOT NULL REFERENCES iam_users(id),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by          UUID            NOT NULL REFERENCES iam_users(id),

    -- Constraints
    CONSTRAINT chk_inv_execution_side
        CHECK (side IN ('BUY','SELL')),

    CONSTRAINT chk_inv_execution_status
        CHECK (status IN ('PENDING','EXECUTED','PARTIALLY_EXECUTED','CANCELLED')),

    CONSTRAINT chk_inv_execution_currency
        CHECK (currency ~ '^[A-Z]{3}$'),

    CONSTRAINT fk_inv_execution_portfolio_fund
        FOREIGN KEY (portfolio_id, fund_id)
        REFERENCES investment__portfolios (id, fund_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_inv_execution_decision
    ON investment__executions (decision_id);

CREATE INDEX IF NOT EXISTS idx_inv_execution_fund_date
    ON investment__executions (fund_id, business_date DESC);

CREATE INDEX IF NOT EXISTS idx_inv_execution_portfolio_date
    ON investment__executions (portfolio_id, business_date DESC);

CREATE INDEX IF NOT EXISTS idx_inv_execution_status
    ON investment__executions (status);

-- Comments
COMMENT ON TABLE investment__executions IS
    'Minimal execution record. One decision can have at most one current execution in this phase.';
