-- Table: compliance_rule_set_members
-- Source: 20260417000001_compliance__create_rules_tables.up.sql
CREATE TABLE IF NOT EXISTS compliance_rule_set_members (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_set_id      UUID        NOT NULL REFERENCES compliance_rule_sets(id) ON DELETE CASCADE,
    rule_instance_id UUID        NOT NULL REFERENCES compliance_rule_instances(id),
    added_by         UUID        REFERENCES iam_users(id),
    added_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_compliance_rsm UNIQUE (rule_set_id, rule_instance_id)
);
