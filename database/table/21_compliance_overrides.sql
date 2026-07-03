-- Table: compliance_overrides
-- Source: 20260417000001_compliance__create_rules_tables.up.sql
CREATE TABLE IF NOT EXISTS compliance_overrides (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    breach_id        UUID        NOT NULL REFERENCES compliance_breaches(id),
    reason           TEXT        NOT NULL,              -- mandatory justification
    overridden_by    UUID        NOT NULL REFERENCES iam_users(id),
    delegated_from   UUID        REFERENCES iam_users(id),
    approved_by      UUID        REFERENCES iam_users(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- NO updated_at — append-only
    CONSTRAINT uq_compliance_ov_breach UNIQUE (breach_id)
);
