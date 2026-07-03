-- Table: investment__decisions
-- Source: 20260601000001_investment__create_decisions_executions_confirmations.up.sql
CREATE TABLE IF NOT EXISTS investment__decisions (
    id                                      UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_number                         VARCHAR(40)     NOT NULL,

    fund_id                                 UUID            NOT NULL REFERENCES investment__funds(id) ON DELETE RESTRICT,
    portfolio_id                            UUID            NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,
    contract_id                             UUID            NOT NULL,
    instrument_id                           UUID            REFERENCES investment__instruments(id) ON DELETE RESTRICT,
    instrument_code                         VARCHAR(40),
    business_date                           DATE            NOT NULL,

    side                                    VARCHAR(10),
    quantity                                DECIMAL(28,8),
    amount                                  DECIMAL(28,8),
    limit_price                             DECIMAL(28,8),
    currency                                CHAR(3)         NOT NULL,
    exchange                                VARCHAR(40),

    research_report_id                      UUID            REFERENCES investment__research_reports(id) ON DELETE RESTRICT,
    research_report_no                      VARCHAR(60),

    decision_type                           VARCHAR(20)     NOT NULL DEFAULT 'SINGLE_ORDER',
    process_type                            VARCHAR(20)     NOT NULL DEFAULT 'INVESTMENT_DECISION',
    product_type                            VARCHAR(20)     NOT NULL DEFAULT 'MUTUAL_FUND',
    strategy_code                           VARCHAR(40),
    amendment_no                            INTEGER         NOT NULL DEFAULT 0,

    rationale                               TEXT,
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

    created_at                              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_by                              UUID            NOT NULL REFERENCES iam_users(id),
    updated_at                              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by                              UUID            NOT NULL REFERENCES iam_users(id),

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
        CHECK (product_type IN ('MUTUAL_FUND','ETF','STOCK','BOND','CASH','MIXED'))
);
