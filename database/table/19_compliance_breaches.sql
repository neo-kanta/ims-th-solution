-- Table: compliance_breaches
-- Source: 20260417000001_compliance__create_rules_tables.up.sql
CREATE TABLE IF NOT EXISTS compliance_breaches (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    check_record_id  UUID        NOT NULL REFERENCES compliance_check_records(id),
    check_group_id   UUID        NOT NULL,
    portfolio_id     UUID        NOT NULL,
    contract_id      UUID        NOT NULL,
    rule_type_id     VARCHAR(100) NOT NULL,
    rule_instance_id UUID        NOT NULL REFERENCES compliance_rule_instances(id),
    severity         VARCHAR(30) NOT NULL,
    verdict          VARCHAR(10) NOT NULL,
    status           VARCHAR(30) NOT NULL DEFAULT 'OPEN',   -- OPEN | OVERRIDDEN | RESOLVED
    evidence         JSONB,
    message          TEXT,
    business_date    DATE        NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at      TIMESTAMPTZ,
    resolved_by      UUID        REFERENCES iam_users(id),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
