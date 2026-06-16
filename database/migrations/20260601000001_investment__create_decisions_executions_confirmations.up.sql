-- =============================================================================
-- Phase: 2026-06-01 — Investment Decision, Execution and Trade Confirmation
-- =============================================================================
--
-- Adds three new aggregates that close the 4-step investment process loop
--   analysis  -> decision  -> execution  -> trade confirmation (review)
--
-- Notes
--   * Decision is the trader's proposed trade. Persisted from DRAFT and driven
--     through approval (via the generic Approval Module) before becoming
--     executable. References research report when one is required.
--   * Execution captures the actual order send / fill. Created only after the
--     decision has been approved and submitted for execution.
--   * Trade confirmation records the broker-confirmed quantities/prices and
--     reconciles them against the execution. Used by the workflow closing
--     gate (close_transactions) to refuse closing while pending or while a
--     mismatch lacks a documented review reason.
--
-- All three tables are append-friendly (status flips happen via UPDATE on
-- mutable rows, but historical content is preserved through audit events
-- elsewhere). The transactions ledger remains the immutable financial truth.
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. Investment decisions
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS investment__decisions (
    id                          UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_number             VARCHAR(40)     NOT NULL,

    fund_id                     UUID            NOT NULL REFERENCES investment__funds(id)       ON DELETE RESTRICT,
    portfolio_id                UUID            NOT NULL REFERENCES investment__portfolios(id)  ON DELETE RESTRICT,
    contract_id                 UUID            NOT NULL,
    instrument_id               UUID            REFERENCES investment__instruments(id)          ON DELETE RESTRICT,
    instrument_code             VARCHAR(40)     NOT NULL,
    business_date               DATE            NOT NULL,

    side                        VARCHAR(10)     NOT NULL,
    quantity                    DECIMAL(28,8),
    amount                      DECIMAL(28,8),
    limit_price                 DECIMAL(28,8),
    currency                    CHAR(3)         NOT NULL,
    exchange                    VARCHAR(40),

    research_report_id          UUID            REFERENCES investment__research_reports(id)     ON DELETE RESTRICT,
    research_report_no          VARCHAR(60),

    rationale                   TEXT,
    status                      VARCHAR(24)     NOT NULL DEFAULT 'DRAFT',
    approval_request_id         UUID,
    approval_status             VARCHAR(24),
    compliance_check_group_id   UUID,

    submitter_user_id           UUID            NOT NULL REFERENCES iam_users(id),
    submitted_at                TIMESTAMPTZ,
    cancelled_at                TIMESTAMPTZ,
    cancelled_by                UUID            REFERENCES iam_users(id),
    cancellation_reason         TEXT,

    ready_for_execution_at      TIMESTAMPTZ,

    created_at                  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_by                  UUID            NOT NULL REFERENCES iam_users(id),
    updated_at                  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by                  UUID            NOT NULL REFERENCES iam_users(id),

    CONSTRAINT uq_inv_decision_number
        UNIQUE (decision_number),

    CONSTRAINT chk_inv_decision_side
        CHECK (side IN ('BUY','SELL')),

    CONSTRAINT chk_inv_decision_status
        CHECK (status IN (
            'DRAFT','PENDING_APPROVAL','APPROVED','REJECTED',
            'CANCELLED','READY_FOR_EXECUTION','EXECUTED'
        )),

    CONSTRAINT chk_inv_decision_currency
        CHECK (currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_decision_quantity_or_amount
        CHECK (quantity IS NOT NULL OR amount IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_inv_decision_fund_date
    ON investment__decisions (fund_id, business_date DESC);

CREATE INDEX IF NOT EXISTS idx_inv_decision_portfolio_date
    ON investment__decisions (portfolio_id, business_date DESC);

CREATE INDEX IF NOT EXISTS idx_inv_decision_status
    ON investment__decisions (status);

CREATE INDEX IF NOT EXISTS idx_inv_decision_research_report
    ON investment__decisions (research_report_id)
    WHERE research_report_id IS NOT NULL;

-- -----------------------------------------------------------------------------
-- 2. Investment executions (minimal model)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS investment__executions (
    id                          UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_id                 UUID            NOT NULL REFERENCES investment__decisions(id) ON DELETE RESTRICT,

    fund_id                     UUID            NOT NULL REFERENCES investment__funds(id)       ON DELETE RESTRICT,
    portfolio_id                UUID            NOT NULL REFERENCES investment__portfolios(id)  ON DELETE RESTRICT,
    contract_id                 UUID            NOT NULL,
    instrument_id               UUID            REFERENCES investment__instruments(id)          ON DELETE RESTRICT,
    instrument_code             VARCHAR(40)     NOT NULL,
    business_date               DATE            NOT NULL,

    side                        VARCHAR(10)     NOT NULL,
    ordered_quantity            DECIMAL(28,8),
    ordered_amount              DECIMAL(28,8),
    executed_quantity           DECIMAL(28,8),
    executed_amount             DECIMAL(28,8),
    execution_price             DECIMAL(28,8),
    currency                    CHAR(3)         NOT NULL,

    status                      VARCHAR(24)     NOT NULL DEFAULT 'PENDING',
    trader_user_id              UUID            REFERENCES iam_users(id),
    broker_reference            VARCHAR(80),

    executed_at                 TIMESTAMPTZ,
    cancelled_at                TIMESTAMPTZ,
    cancelled_by                UUID            REFERENCES iam_users(id),
    cancellation_reason         TEXT,

    created_at                  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_by                  UUID            NOT NULL REFERENCES iam_users(id),
    updated_at                  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by                  UUID            NOT NULL REFERENCES iam_users(id),

    CONSTRAINT chk_inv_execution_side
        CHECK (side IN ('BUY','SELL')),

    CONSTRAINT chk_inv_execution_status
        CHECK (status IN ('PENDING','EXECUTED','PARTIALLY_EXECUTED','CANCELLED')),

    CONSTRAINT chk_inv_execution_currency
        CHECK (currency ~ '^[A-Z]{3}$')
);

CREATE INDEX IF NOT EXISTS idx_inv_execution_decision
    ON investment__executions (decision_id);

CREATE INDEX IF NOT EXISTS idx_inv_execution_fund_date
    ON investment__executions (fund_id, business_date DESC);

CREATE INDEX IF NOT EXISTS idx_inv_execution_portfolio_date
    ON investment__executions (portfolio_id, business_date DESC);

CREATE INDEX IF NOT EXISTS idx_inv_execution_status
    ON investment__executions (status);

-- -----------------------------------------------------------------------------
-- 3. Trade confirmations (review of broker fills against decision/execution)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS investment__trade_confirmations (
    id                          UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id                UUID            NOT NULL REFERENCES investment__executions(id)  ON DELETE RESTRICT,
    decision_id                 UUID            NOT NULL REFERENCES investment__decisions(id)   ON DELETE RESTRICT,

    fund_id                     UUID            NOT NULL REFERENCES investment__funds(id)       ON DELETE RESTRICT,
    portfolio_id                UUID            NOT NULL REFERENCES investment__portfolios(id)  ON DELETE RESTRICT,
    contract_id                 UUID            NOT NULL,
    business_date               DATE            NOT NULL,

    confirmed_quantity          DECIMAL(28,8),
    confirmed_amount            DECIMAL(28,8),
    confirmed_price             DECIMAL(28,8),
    currency                    CHAR(3)         NOT NULL,

    broker_reference            VARCHAR(80),
    import_batch_id             UUID,

    status                      VARCHAR(24)     NOT NULL DEFAULT 'PENDING_REVIEW',
    discrepancy_reason          TEXT,
    reviewed_at                 TIMESTAMPTZ,
    reviewed_by                 UUID            REFERENCES iam_users(id),

    created_at                  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_by                  UUID            NOT NULL REFERENCES iam_users(id),
    updated_at                  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by                  UUID            NOT NULL REFERENCES iam_users(id),

    CONSTRAINT chk_inv_confirmation_status
        CHECK (status IN ('PENDING_REVIEW','MATCHED','MISMATCHED','REVIEWED')),

    CONSTRAINT chk_inv_confirmation_currency
        CHECK (currency ~ '^[A-Z]{3}$'),

    -- If marked mismatched or reviewed, a reason must be supplied.
    CONSTRAINT chk_inv_confirmation_reason_required
        CHECK (
            status NOT IN ('MISMATCHED','REVIEWED')
            OR (discrepancy_reason IS NOT NULL AND length(trim(discrepancy_reason)) > 0)
        )
);

CREATE INDEX IF NOT EXISTS idx_inv_confirmation_execution
    ON investment__trade_confirmations (execution_id);

CREATE INDEX IF NOT EXISTS idx_inv_confirmation_decision
    ON investment__trade_confirmations (decision_id);

CREATE INDEX IF NOT EXISTS idx_inv_confirmation_contract_date
    ON investment__trade_confirmations (contract_id, business_date DESC);

CREATE INDEX IF NOT EXISTS idx_inv_confirmation_status
    ON investment__trade_confirmations (status);

COMMENT ON TABLE investment__decisions               IS 'Investment decision aggregate (proposed trade prior to execution).';
COMMENT ON TABLE investment__executions              IS 'Minimal execution record. One decision can have at most one current execution in this phase.';
COMMENT ON TABLE investment__trade_confirmations     IS 'Broker confirmation reconciled against the execution. Required for transaction closing.';
