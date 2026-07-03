-- Table: compliance_rule_bindings
-- Source: 20260417000001_compliance__create_rules_tables.up.sql
CREATE TABLE IF NOT EXISTS compliance_rule_bindings (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_instance_id UUID        NOT NULL REFERENCES compliance_rule_instances(id),
    rule_set_id      UUID        REFERENCES compliance_rule_sets(id),  -- NULL if direct binding

    -- Scope
    scope_type       VARCHAR(50)  NOT NULL,          -- GLOBAL | JURISDICTION | FUND_CATEGORY | CONTRACT | PORTFOLIO | ASSET_CLASS | INSTRUMENT_TYPE
    scope_id         UUID,                            -- NULL for GLOBAL scope

    -- Severity override (may cap the rule type's default)
    severity         VARCHAR(30) NOT NULL,            -- BLOCK | WARN | REQUIRE_APPROVAL | MONITOR
    priority         INT         NOT NULL DEFAULT 100, -- lower = higher priority; tie-breaking

    -- Effective dating
    effective_from   DATE        NOT NULL,
    effective_to     DATE,                            -- NULL = open-ended

    is_active        BOOLEAN     NOT NULL DEFAULT true,
    created_by       UUID        REFERENCES iam_users(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
