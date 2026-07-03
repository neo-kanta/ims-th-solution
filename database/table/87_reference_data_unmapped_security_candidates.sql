-- Table: reference_data_unmapped_security_candidates
-- Source: 20260526000002_referencedata__create_securities.up.sql
CREATE TABLE reference_data_unmapped_security_candidates (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id               UUID,
    provider_code          VARCHAR(40)  NOT NULL,
    provider_symbol        VARCHAR(128) NOT NULL,
    provider_name          TEXT,
    provider_asset_type    VARCHAR(40),
    provider_exchange      VARCHAR(64),
    provider_currency      CHAR(3),
    isin                   VARCHAR(12),
    raw_payload            JSONB        NOT NULL DEFAULT '{}'::jsonb,
    candidate_status       VARCHAR(30)  NOT NULL DEFAULT 'REVIEW_REQUIRED',
    suggested_security_id  UUID         REFERENCES securities_master(id) ON DELETE SET NULL,
    confidence_score       NUMERIC(5,2),
    rejected_reason        TEXT,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    resolved_at            TIMESTAMPTZ,

    CONSTRAINT chk_unmapped_candidates_status
        CHECK (candidate_status IN ('REVIEW_REQUIRED','MAPPED','REJECTED','DUPLICATE','CONFLICTED')),
    CONSTRAINT chk_unmapped_candidates_provider_code
        CHECK (provider_code <> ''),
    CONSTRAINT chk_unmapped_candidates_provider_symbol
        CHECK (provider_symbol <> ''),
    CONSTRAINT chk_unmapped_candidates_currency
        CHECK (provider_currency IS NULL OR provider_currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_unmapped_candidates_isin_len
        CHECK (isin IS NULL OR char_length(isin) = 12)
);
