-- Table: compliance_restriction_list_entries
-- Source: 20260424000001_compliance__create_restriction_list.up.sql
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
