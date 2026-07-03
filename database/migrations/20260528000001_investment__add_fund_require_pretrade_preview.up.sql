-- =============================================================================
-- Investment / Fund — require_pretrade_preview flag
-- =============================================================================
-- Adds a per-fund toggle that controls the trade-ticket UX:
--   * true  — the Operation tab must run a pre-trade simulation before allowing
--             a posted transaction. Users see per-rule verdicts and only the
--             Post button activates for PASS or WARN.
--   * false — the Operation tab posts directly; the server still enforces the
--             pre-trade gates inside the post handler, but no preview UI runs.
--
-- Default is false so existing rows keep today's behaviour (server-side
-- gating only). The frontend Create-fund form lets the operator opt in.
-- =============================================================================

ALTER TABLE investment__funds
    ADD COLUMN IF NOT EXISTS require_pretrade_preview BOOLEAN NOT NULL DEFAULT false;

COMMENT ON COLUMN investment__funds.require_pretrade_preview IS
    'When true, the trade ticket must run pre-trade simulation before allowing a transaction post.';
