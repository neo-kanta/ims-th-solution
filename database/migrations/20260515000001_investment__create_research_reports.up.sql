-- =============================================================================
-- Investment Module — Research Reports (basic CRUD foundation)
-- =============================================================================
-- Owners: investment module. Stores analyst-authored research reports with a
-- DRAFT / ACTIVE / EXPIRED / REJECTED report lifecycle and a separate
-- NOT_SUBMITTED / SUBMITTED / REVIEW_COMPLETED review status.
--
-- Foreign keys to author/owner users and to instruments/contracts are
-- intentionally omitted at this stage — those references are kept as plain
-- UUID columns and will be hardened with FK constraints in a follow-up once
-- the referenced tables stabilise (per the PoC scaffolding plan).
--
-- Soft delete: `deleted_at`. Application-level reads MUST filter `deleted_at
-- IS NULL`.
-- =============================================================================

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

CREATE UNIQUE INDEX uq_inv_research_report_no_alive
    ON investment__research_reports (report_no) WHERE deleted_at IS NULL;

CREATE INDEX idx_inv_research_report_status
    ON investment__research_reports (report_status) WHERE deleted_at IS NULL;

CREATE INDEX idx_inv_research_review_status
    ON investment__research_reports (review_status) WHERE deleted_at IS NULL;

CREATE INDEX idx_inv_research_owner
    ON investment__research_reports (owner_user_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_inv_research_instrument_code
    ON investment__research_reports (instrument_code) WHERE deleted_at IS NULL;

CREATE INDEX idx_inv_research_report_date
    ON investment__research_reports (report_date) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_inv_research_reports_updated_at
    BEFORE UPDATE ON investment__research_reports
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE  investment__research_reports IS 'Investment research / analysis reports authored by research analysts. PoC scope: simple CRUD only — no approval workflow, AI summary, evidence engine, or thesis drift detection yet.';
COMMENT ON COLUMN investment__research_reports.report_no IS 'Human-readable report identifier (e.g. RR-2026-0001). Unique among non-deleted rows.';
COMMENT ON COLUMN investment__research_reports.owner_user_id IS 'Owner of the report — typically the analyst accountable for it. Plain UUID (no FK during PoC).';
COMMENT ON COLUMN investment__research_reports.author_user_id IS 'Author of the report content. May equal owner_user_id.';
COMMENT ON COLUMN investment__research_reports.applicable_contract_id IS 'Optional fund/contract scope for the recommendation. Plain UUID (no FK during PoC).';
COMMENT ON COLUMN investment__research_reports.instrument_code IS 'Free-text ticker / instrument identifier. Plain string (no FK to investment__instruments during PoC).';
COMMENT ON COLUMN investment__research_reports.report_status IS 'DRAFT (editable) | ACTIVE | EXPIRED | REJECTED. Lifecycle is managed by the report author / admin in this PoC; future iterations may automate transitions.';
COMMENT ON COLUMN investment__research_reports.review_status IS 'NOT_SUBMITTED | SUBMITTED | REVIEW_COMPLETED. Drives the simple submit / cancel-submit endpoints; full approval workflow is out of scope for this PoC.';
