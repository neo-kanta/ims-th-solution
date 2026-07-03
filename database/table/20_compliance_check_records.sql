-- Table: compliance_check_records
-- Source: 20260417000001_compliance__create_rules_tables.up.sql
CREATE TABLE IF NOT EXISTS compliance_check_records (
    id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    check_group_id       UUID        NOT NULL,         -- groups all records from one check request
    timing               VARCHAR(30) NOT NULL,          -- PRE_TRADE | POST_TRADE | PERIODIC
    order_id             UUID,                          -- NULL for periodic/post-trade scans
    portfolio_id         UUID        NOT NULL,
    contract_id          UUID        NOT NULL,
    ticker               VARCHAR(20),                   -- empty for portfolio-wide checks
    rule_type_id         VARCHAR(100) NOT NULL,
    rule_instance_id     UUID        NOT NULL REFERENCES compliance_rule_instances(id),
    rule_instance_version INT        NOT NULL,
    parameter_snapshot   JSONB       NOT NULL,          -- frozen copy of params used
    verdict              VARCHAR(10) NOT NULL,           -- raw verdict from rule: PASS | WARN | BLOCK
    effective_severity   VARCHAR(30) NOT NULL,           -- severity from binding
    final_verdict        VARCHAR(10) NOT NULL,           -- after severity cap
    evidence             JSONB,
    message              TEXT,
    data_snapshot_hash   VARCHAR(64) NOT NULL,           -- SHA-256 hex of DataBundle
    eval_duration_ms     BIGINT      NOT NULL DEFAULT 0,
    checked_by           VARCHAR(100) NOT NULL,          -- actor user ID or "system"
    business_date        DATE        NOT NULL,
    checked_at           TIMESTAMPTZ NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- NO updated_at — append-only
);
