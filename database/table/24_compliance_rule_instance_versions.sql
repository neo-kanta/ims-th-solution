-- Table: compliance_rule_instance_versions
-- Source: 20260417000001_compliance__create_rules_tables.up.sql
CREATE TABLE IF NOT EXISTS compliance_rule_instance_versions (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_instance_id UUID        NOT NULL REFERENCES compliance_rule_instances(id) ON DELETE CASCADE,
    version_number   INT         NOT NULL,
    parameters       JSONB       NOT NULL,             -- validated against rule type's JSON Schema
    change_note      TEXT,
    created_by       UUID        REFERENCES iam_users(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_compliance_riv_version UNIQUE (rule_instance_id, version_number)
);
