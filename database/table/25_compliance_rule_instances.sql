-- Table: compliance_rule_instances
-- Source: 20260417000001_compliance__create_rules_tables.up.sql
CREATE TABLE IF NOT EXISTS compliance_rule_instances (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_type_id      VARCHAR(100) NOT NULL,           -- SPI type ID e.g. "concentration.single_issuer"
    name              VARCHAR(255) NOT NULL,
    description       TEXT,
    current_version   INT         NOT NULL DEFAULT 1,  -- pointer to active version
    is_active         BOOLEAN     NOT NULL DEFAULT true,
    effective_from    DATE        NOT NULL,
    effective_to      DATE,                             -- NULL = open-ended
    created_by        UUID        REFERENCES iam_users(id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_compliance_ri_name UNIQUE (name)
);
