-- Table: investment__research_reports
-- Source: 20260515000001_investment__create_research_reports.up.sql
CREATE TABLE investment__research_reports (
    id                      UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    report_no               VARCHAR(60)  NOT NULL,
    report_date             DATE         NOT NULL,
    effective_date          DATE,

    owner_user_id           UUID         NOT NULL,
    author_user_id          UUID         NOT NULL,
    applicable_contract_id  UUID,

    instrument_type         VARCHAR(40)  NOT NULL DEFAULT '',
    instrument_code         VARCHAR(40)  NOT NULL,
    instrument_name         VARCHAR(255) NOT NULL DEFAULT '',
    market                  VARCHAR(40)  NOT NULL DEFAULT '',
    currency                VARCHAR(3)   NOT NULL DEFAULT '',

    recommendation          VARCHAR(10)  NOT NULL,
    report_title            VARCHAR(255) NOT NULL DEFAULT '',

    company_overview        TEXT         NOT NULL DEFAULT '',
    company_outlook         TEXT         NOT NULL DEFAULT '',
    esg_comment             TEXT         NOT NULL DEFAULT '',
    financial_status        TEXT         NOT NULL DEFAULT '',
    investment_analysis     TEXT         NOT NULL,

    rejection_reason        TEXT         NOT NULL DEFAULT '',
    post_submission_note    TEXT         NOT NULL DEFAULT '',

    report_status           VARCHAR(20)  NOT NULL DEFAULT 'DRAFT',
    review_status           VARCHAR(20)  NOT NULL DEFAULT 'NOT_SUBMITTED',

    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by              UUID,
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by              UUID,
    deleted_at              TIMESTAMPTZ,

    CONSTRAINT chk_inv_research_recommendation
        CHECK (recommendation IN ('BUY', 'SELL', 'HOLD')),

    CONSTRAINT chk_inv_research_report_status
        CHECK (report_status IN ('DRAFT', 'ACTIVE', 'EXPIRED', 'REJECTED')),

    CONSTRAINT chk_inv_research_review_status
        CHECK (review_status IN ('NOT_SUBMITTED', 'SUBMITTED', 'REVIEW_COMPLETED')),

    CONSTRAINT chk_inv_research_currency
        CHECK (currency = '' OR currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_inv_research_investment_analysis_length
        CHECK (char_length(investment_analysis) >= 25)
);
