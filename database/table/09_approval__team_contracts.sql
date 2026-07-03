-- Table: approval__team_contracts
-- Source: 20260529000001_approval__create_tables.up.sql
CREATE TABLE approval__team_contracts (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id        UUID         NOT NULL REFERENCES approval__teams(id) ON DELETE CASCADE,
    contract_id    UUID         NOT NULL,
    effective_date DATE         NOT NULL DEFAULT CURRENT_DATE,
    is_active      BOOLEAN      NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
