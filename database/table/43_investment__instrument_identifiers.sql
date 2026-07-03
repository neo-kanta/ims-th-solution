-- Table: investment__instrument_identifiers
-- Source: 20260428000003_investment__create_instruments.up.sql
CREATE TABLE investment__instrument_identifiers (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    instrument_id   UUID         NOT NULL REFERENCES investment__instruments(id) ON DELETE CASCADE,
    id_type         VARCHAR(20)  NOT NULL,
    id_value        VARCHAR(60)  NOT NULL,
    provider_code   VARCHAR(40),
    is_primary      BOOLEAN      NOT NULL DEFAULT false,

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by      UUID         REFERENCES iam_users(id),
    updated_by      UUID         REFERENCES iam_users(id),

    CONSTRAINT chk_inv_instrument_identifiers_type
        CHECK (id_type IN ('ISIN','CUSIP','SEDOL','BBG','RIC','MORNINGSTAR','PROVIDER'))
);
