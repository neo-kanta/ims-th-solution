-- Compliance restriction list entries
--
-- One row per (ticker, list_type, source) pairing that is effective on a given
-- business date. The list_type column drives the restriction.list_enforcement
-- rule (BLACKLIST / WHITELIST / ALERT / DISPOSAL); GRAYLIST is accepted as a
-- legacy alias of ALERT by the rule evaluator.
--
-- Source indicates where the restriction originated (e.g. GLOBAL, CONTRACT,
-- REGULATORY). Effective dating lets operators queue future entries without
-- touching production data at switch-over time.

CREATE TABLE IF NOT EXISTS compliance_restriction_list_entries (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    ticker            VARCHAR(64) NOT NULL,
    issuer_id         UUID,                      -- optional: classify at issuer granularity
    list_type         VARCHAR(32) NOT NULL,      -- BLACKLIST | WHITELIST | ALERT | DISPOSAL | GRAYLIST
    reason            TEXT        NOT NULL,
    source            VARCHAR(64) NOT NULL,      -- GLOBAL | CONTRACT | REGULATORY | ...
    effective_from    DATE        NOT NULL,
    effective_to      DATE,                      -- NULL = open-ended
    created_by        UUID        REFERENCES iam_users(id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_compliance_rl_list_type CHECK (
        list_type IN ('BLACKLIST', 'WHITELIST', 'ALERT', 'DISPOSAL', 'GRAYLIST')
    ),
    CONSTRAINT chk_compliance_rl_date_range CHECK (
        effective_to IS NULL OR effective_to >= effective_from
    )
);

COMMENT ON TABLE  compliance_restriction_list_entries              IS 'Active restriction list rows consumed by restriction.list_enforcement.';
COMMENT ON COLUMN compliance_restriction_list_entries.list_type    IS 'BLACKLIST (hard block) / WHITELIST (allow-only) / ALERT (warn) / DISPOSAL (buy blocked, sell allowed). GRAYLIST is a legacy alias of ALERT.';
COMMENT ON COLUMN compliance_restriction_list_entries.source       IS 'Origin of the restriction: GLOBAL, CONTRACT, REGULATORY, etc.';
COMMENT ON COLUMN compliance_restriction_list_entries.effective_to IS 'NULL means the entry is open-ended; otherwise the entry is inactive on or after this date.';

-- Fast ticker/date lookup — the hot-path query in the Postgres adapter.
CREATE INDEX IF NOT EXISTS idx_compliance_rl_ticker_active
    ON compliance_restriction_list_entries (ticker, list_type)
    WHERE effective_to IS NULL;

CREATE INDEX IF NOT EXISTS idx_compliance_rl_effective_range
    ON compliance_restriction_list_entries (effective_from, effective_to);

CREATE INDEX IF NOT EXISTS idx_compliance_rl_issuer
    ON compliance_restriction_list_entries (issuer_id)
    WHERE issuer_id IS NOT NULL;
