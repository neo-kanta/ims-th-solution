-- Table: investment__decisions
-- Source: 20260601000001_investment__create_decisions_executions_confirmations.up.sql,
--         updated by 20260615000004_investment__add_decision_basket_fields.up.sql,
--         updated by 20260618000002_investment__add_compliance_release_status.up.sql,
--         updated by 20260618000003_investment__add_compliance_release_approval_id.up.sql,
--         updated by 20260703000001_investment__portfolio_v2_hardening.up.sql,
--         updated by 20260703000002_investment__drop_contract_id_from_operational_tables.up.sql
-- Purpose: Investment decision aggregate, before execution. Portfolio is the operational source of truth.
-- Identity: decision_number is the business identity; database links use portfolio_id.
--           fund_id is a legacy fund-wrapper compatibility field.
CREATE TABLE IF NOT EXISTS investment__decisions (
    -- Identity
    id                                      UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_number                         VARCHAR(40)     NOT NULL,
    portfolio_id                            UUID            NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,

    -- Optional fund wrapper / legacy compatibility
    fund_id                                 UUID            NOT NULL REFERENCES investment__funds(id) ON DELETE RESTRICT,

    -- Instrument / order header
    instrument_id                           UUID            REFERENCES investment__instruments(id) ON DELETE RESTRICT,
    instrument_code                         VARCHAR(40),
    business_date                           DATE            NOT NULL,
    side                                    VARCHAR(10),
    quantity                                DECIMAL(28,8),
    amount                                  DECIMAL(28,8),
    limit_price                             DECIMAL(28,8),
    currency                                CHAR(3)         NOT NULL,
    exchange                                VARCHAR(40),

    -- Research / strategy
    research_report_id                      UUID            REFERENCES investment__research_reports(id) ON DELETE RESTRICT,
    research_report_no                      VARCHAR(60),
    decision_type                           VARCHAR(20)     NOT NULL DEFAULT 'SINGLE_ORDER',
    process_type                            VARCHAR(20)     NOT NULL DEFAULT 'INVESTMENT_DECISION',
    product_type                            VARCHAR(20)     NOT NULL DEFAULT 'MUTUAL_FUND',
    strategy_code                           VARCHAR(40),
    amendment_no                            INTEGER         NOT NULL DEFAULT 0,
    rationale                               TEXT,

    -- Lifecycle / status
    status                                  VARCHAR(24)     NOT NULL DEFAULT 'DRAFT',
    approval_request_id                     UUID,
    compliance_release_approval_request_id  UUID,
    approval_status                         VARCHAR(24),
    compliance_check_group_id               UUID,
    submitter_user_id                       UUID            NOT NULL REFERENCES iam_users(id),
    submitted_at                            TIMESTAMPTZ,
    cancelled_at                            TIMESTAMPTZ,
    cancelled_by                            UUID            REFERENCES iam_users(id),
    cancellation_reason                     TEXT,
    ready_for_execution_at                  TIMESTAMPTZ,

    -- Audit
    created_at                              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_by                              UUID            NOT NULL REFERENCES iam_users(id),
    updated_at                              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by                              UUID            NOT NULL REFERENCES iam_users(id),

    -- Constraints
    CONSTRAINT uq_inv_decision_number
        UNIQUE (decision_number),

    CONSTRAINT chk_inv_decision_side
        CHECK (
            side IS NULL
            OR side IN ('BUY','SELL','SWITCH','SWITCH_IN','SWITCH_OUT','REBALANCE_BUY','REBALANCE_SELL')
        ),

    CONSTRAINT chk_inv_decision_status
        CHECK (status IN (
            'DRAFT',
            'PENDING_APPROVAL',
            'APPROVED',
            'REJECTED',
            'CANCELLED',
            'READY_FOR_EXECUTION',
            'EXECUTED',
            'PENDING_COMPLIANCE_RELEASE'
        )),

    CONSTRAINT chk_inv_decision_currency
        CHECK (currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_decision_decision_type
        CHECK (decision_type IN ('SINGLE_ORDER','BASKET_ORDER','REBALANCE','SWITCH')),

    CONSTRAINT chk_inv_decision_process_type
        CHECK (process_type IN ('INVESTMENT_DECISION','ORDER_CANCEL','ORDER_AMEND')),

    CONSTRAINT chk_inv_decision_product_type
        CHECK (product_type IN ('MUTUAL_FUND','ETF','STOCK','BOND','CASH','MIXED')),

    CONSTRAINT fk_inv_decision_portfolio_fund
        FOREIGN KEY (portfolio_id, fund_id)
        REFERENCES investment__portfolios (id, fund_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_inv_decision_fund_date
    ON investment__decisions (fund_id, business_date DESC);

CREATE INDEX IF NOT EXISTS idx_inv_decision_portfolio_date
    ON investment__decisions (portfolio_id, business_date DESC);

CREATE INDEX IF NOT EXISTS idx_inv_decision_status
    ON investment__decisions (status);

CREATE INDEX IF NOT EXISTS idx_inv_decision_research_report
    ON investment__decisions (research_report_id)
    WHERE research_report_id IS NOT NULL;

-- Comments
COMMENT ON TABLE investment__decisions IS
    'Investment decision aggregate (proposed trade prior to execution).';
COMMENT ON COLUMN investment__decisions.decision_type IS
    'SINGLE_ORDER | BASKET_ORDER | REBALANCE | SWITCH';
COMMENT ON COLUMN investment__decisions.process_type IS
    'INVESTMENT_DECISION | ORDER_CANCEL | ORDER_AMEND';
COMMENT ON COLUMN investment__decisions.product_type IS
    'MUTUAL_FUND | ETF | STOCK | BOND | CASH | MIXED';
COMMENT ON COLUMN investment__decisions.strategy_code IS
    'Optional reference to a research/strategy allocation.';
COMMENT ON COLUMN investment__decisions.amendment_no IS
    'Incremented each time an approved decision is amended.';
