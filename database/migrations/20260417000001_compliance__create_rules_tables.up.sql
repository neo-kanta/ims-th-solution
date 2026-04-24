-- Compliance / IRG (Investment Regulation Guard) schema
-- Four concerns kept strictly separate:
--   1. compliance_rule_instances      — what rule type + human config
--   2. compliance_rule_instance_versions — immutable parameter snapshots
--   3. compliance_rule_bindings       — scope mapping + effective dating
--   4. compliance_check_records       — append-only evaluation audit
-- Plus: rule_sets, breaches, overrides.

-- ============================================================
-- RULE INSTANCES
-- ============================================================
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

COMMENT ON TABLE  compliance_rule_instances                  IS 'Configured compliance rule instances linking a rule type to human-readable metadata.';
COMMENT ON COLUMN compliance_rule_instances.rule_type_id     IS 'Stable SPI type identifier; must match a registered RuleEvaluator.';
COMMENT ON COLUMN compliance_rule_instances.current_version  IS 'Active version number (FK into compliance_rule_instance_versions).';

CREATE INDEX idx_compliance_ri_type    ON compliance_rule_instances(rule_type_id);
CREATE INDEX idx_compliance_ri_active  ON compliance_rule_instances(is_active);

CREATE TRIGGER trg_compliance_ri_updated_at
    BEFORE UPDATE ON compliance_rule_instances
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- RULE INSTANCE VERSIONS  (immutable parameter snapshots)
-- ============================================================
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

COMMENT ON TABLE  compliance_rule_instance_versions             IS 'Immutable versioned parameter snapshots. Never mutated after creation.';
COMMENT ON COLUMN compliance_rule_instance_versions.parameters  IS 'JSON parameters validated against the rule type ParameterSchema at creation time.';

CREATE INDEX idx_compliance_riv_instance ON compliance_rule_instance_versions(rule_instance_id);

-- Prevent parameter mutation after initial insert.
CREATE RULE no_update_compliance_riv AS ON UPDATE TO compliance_rule_instance_versions DO INSTEAD NOTHING;
CREATE RULE no_delete_compliance_riv AS ON DELETE TO compliance_rule_instance_versions DO INSTEAD NOTHING;

-- ============================================================
-- RULE SETS  (named groupings for bulk binding)
-- ============================================================
CREATE TABLE IF NOT EXISTS compliance_rule_sets (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    is_active   BOOLEAN     NOT NULL DEFAULT true,
    created_by  UUID        REFERENCES iam_users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_compliance_rs_name UNIQUE (name)
);

COMMENT ON TABLE compliance_rule_sets IS 'Named collections of rule instances that can be bound to a scope as a unit.';

CREATE TRIGGER trg_compliance_rs_updated_at
    BEFORE UPDATE ON compliance_rule_sets
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- RULE SET MEMBERS
-- ============================================================
CREATE TABLE IF NOT EXISTS compliance_rule_set_members (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_set_id      UUID        NOT NULL REFERENCES compliance_rule_sets(id) ON DELETE CASCADE,
    rule_instance_id UUID        NOT NULL REFERENCES compliance_rule_instances(id),
    added_by         UUID        REFERENCES iam_users(id),
    added_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_compliance_rsm UNIQUE (rule_set_id, rule_instance_id)
);

CREATE INDEX idx_compliance_rsm_set      ON compliance_rule_set_members(rule_set_id);
CREATE INDEX idx_compliance_rsm_instance ON compliance_rule_set_members(rule_instance_id);

-- ============================================================
-- RULE BINDINGS  (scope → rule mapping with effective dating)
-- ============================================================
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

COMMENT ON TABLE  compliance_rule_bindings              IS 'Maps a rule instance to a compliance scope. Effective-dated and severity-capped.';
COMMENT ON COLUMN compliance_rule_bindings.scope_type   IS 'Specificity: PORTFOLIO > CONTRACT > FUND_CATEGORY > JURISDICTION > GLOBAL.';
COMMENT ON COLUMN compliance_rule_bindings.severity     IS 'Caps the rule raw verdict. MONITOR = PASS always; WARN = WARN max; BLOCK = full block.';
COMMENT ON COLUMN compliance_rule_bindings.priority     IS 'Lower number = higher priority within equal specificity level.';

CREATE INDEX idx_compliance_rb_instance   ON compliance_rule_bindings(rule_instance_id);
CREATE INDEX idx_compliance_rb_scope      ON compliance_rule_bindings(scope_type, scope_id);
CREATE INDEX idx_compliance_rb_active_eff ON compliance_rule_bindings(is_active, effective_from, effective_to);

CREATE TRIGGER trg_compliance_rb_updated_at
    BEFORE UPDATE ON compliance_rule_bindings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- CHECK RECORDS  (APPEND-ONLY — immutable audit trail)
-- ============================================================
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

COMMENT ON TABLE  compliance_check_records                     IS 'Immutable audit log of every rule evaluation. NEVER updated or deleted.';
COMMENT ON COLUMN compliance_check_records.check_group_id      IS 'Ties all records from a single check request together.';
COMMENT ON COLUMN compliance_check_records.parameter_snapshot  IS 'Frozen parameters used — enables reproducibility without current state.';
COMMENT ON COLUMN compliance_check_records.data_snapshot_hash  IS 'SHA-256 of the DataBundle — proves what market data was used.';

CREATE INDEX idx_compliance_cr_group       ON compliance_check_records(check_group_id);
CREATE INDEX idx_compliance_cr_portfolio   ON compliance_check_records(portfolio_id, business_date);
CREATE INDEX idx_compliance_cr_order       ON compliance_check_records(order_id) WHERE order_id IS NOT NULL;
CREATE INDEX idx_compliance_cr_timing      ON compliance_check_records(timing, final_verdict);
CREATE INDEX idx_compliance_cr_instance    ON compliance_check_records(rule_instance_id);
CREATE INDEX idx_compliance_cr_checked_at  ON compliance_check_records(checked_at DESC);

-- Enforce append-only at the database level.
REVOKE UPDATE, DELETE ON compliance_check_records FROM PUBLIC;

-- ============================================================
-- BREACHES
-- ============================================================
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

COMMENT ON TABLE  compliance_breaches         IS 'Active/historical breach records created from BLOCK or WARN verdicts.';
COMMENT ON COLUMN compliance_breaches.status  IS 'OPEN = unresolved, OVERRIDDEN = compliance officer approved, RESOLVED = rule subsequently passed.';

CREATE INDEX idx_compliance_br_portfolio  ON compliance_breaches(portfolio_id, business_date);
CREATE INDEX idx_compliance_br_status     ON compliance_breaches(status) WHERE status = 'OPEN';
CREATE INDEX idx_compliance_br_record     ON compliance_breaches(check_record_id);
CREATE INDEX idx_compliance_br_instance   ON compliance_breaches(rule_instance_id);

CREATE TRIGGER trg_compliance_br_updated_at
    BEFORE UPDATE ON compliance_breaches
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- OVERRIDES  (APPEND-ONLY — immutable compliance officer actions)
-- ============================================================
CREATE TABLE IF NOT EXISTS compliance_overrides (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    breach_id        UUID        NOT NULL REFERENCES compliance_breaches(id),
    reason           TEXT        NOT NULL,              -- mandatory justification
    overridden_by    UUID        NOT NULL REFERENCES iam_users(id),
    delegated_from   UUID        REFERENCES iam_users(id),
    approved_by      UUID        REFERENCES iam_users(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- NO updated_at — append-only
);

COMMENT ON TABLE  compliance_overrides              IS 'Immutable audit of compliance officer override actions.';
COMMENT ON COLUMN compliance_overrides.reason       IS 'Mandatory written justification for the override.';
COMMENT ON COLUMN compliance_overrides.overridden_by IS 'Must hold IRG_OVERRIDE_BREACH permission.';

CREATE INDEX idx_compliance_ov_breach ON compliance_overrides(breach_id);
CREATE INDEX idx_compliance_ov_officer ON compliance_overrides(overridden_by);

REVOKE UPDATE, DELETE ON compliance_overrides FROM PUBLIC;

-- ============================================================
-- SEED: Compliance permission codes
-- ============================================================
-- These are referenced by permissions_function_rights.permission_code.
-- Insert as constants into the permissions system seed table so admin groups
-- can be granted these rights via the UI without needing another migration.
INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_at)
SELECT g.id, p.code, true, NOW()
FROM permissions_groups g
CROSS JOIN (VALUES
    ('IRG_VIEW_RULES'),
    ('IRG_EDIT_RULE_INSTANCE'),
    ('IRG_EDIT_BINDING'),
    ('IRG_OVERRIDE_BREACH'),
    ('IRG_ADMIN_RULE_TYPE')
) AS p(code)
WHERE g.name = 'Admin'
ON CONFLICT (group_id, permission_code) DO NOTHING;
