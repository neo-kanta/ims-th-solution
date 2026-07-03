-- Table: investment__executions
-- Source: 20260601000001_investment__create_decisions_executions_confirmations.up.sql
CREATE TABLE IF NOT EXISTS investment__executions (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_id         UUID            NOT NULL REFERENCES investment__decisions(id) ON DELETE RESTRICT,

    fund_id             UUID            NOT NULL REFERENCES investment__funds(id) ON DELETE RESTRICT,
    portfolio_id        UUID            NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,
    contract_id         UUID            NOT NULL,
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

    status              VARCHAR(24)     NOT NULL DEFAULT 'PENDING',
    trader_user_id      UUID            REFERENCES iam_users(id),
    broker_reference    VARCHAR(80),

    executed_at         TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,
    cancelled_by        UUID            REFERENCES iam_users(id),
    cancellation_reason TEXT,

    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_by          UUID            NOT NULL REFERENCES iam_users(id),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by          UUID            NOT NULL REFERENCES iam_users(id),

    CONSTRAINT chk_inv_execution_side
        CHECK (side IN ('BUY','SELL')),

    CONSTRAINT chk_inv_execution_status
        CHECK (status IN ('PENDING','EXECUTED','PARTIALLY_EXECUTED','CANCELLED')),

    CONSTRAINT chk_inv_execution_currency
        CHECK (currency ~ '^[A-Z]{3}$')
);
