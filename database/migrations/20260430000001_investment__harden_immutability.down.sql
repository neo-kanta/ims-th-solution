-- =============================================================================
-- Investment Module — Roll back immutability triggers; restore original RULEs
-- =============================================================================
-- Drops the 14 triggers and the helper function created by the .up.sql, then
-- recreates the original RULE-based guards verbatim so the schema lands in
-- exactly the state migration 20260429000100 left it in.
-- =============================================================================

DROP TRIGGER IF EXISTS trg_inv_portfolio_transactions_no_update    ON investment__portfolio_transactions;
DROP TRIGGER IF EXISTS trg_inv_portfolio_transactions_no_delete    ON investment__portfolio_transactions;
DROP TRIGGER IF EXISTS trg_inv_cash_movements_no_update            ON investment__cash_movements;
DROP TRIGGER IF EXISTS trg_inv_cash_movements_no_delete            ON investment__cash_movements;
DROP TRIGGER IF EXISTS trg_inv_price_snapshots_no_update           ON investment__price_snapshots;
DROP TRIGGER IF EXISTS trg_inv_price_snapshots_no_delete           ON investment__price_snapshots;
DROP TRIGGER IF EXISTS trg_inv_valuation_snapshots_no_update       ON investment__valuation_snapshots;
DROP TRIGGER IF EXISTS trg_inv_valuation_snapshots_no_delete       ON investment__valuation_snapshots;
DROP TRIGGER IF EXISTS trg_inv_valuation_holding_lines_no_update   ON investment__valuation_holding_lines;
DROP TRIGGER IF EXISTS trg_inv_valuation_holding_lines_no_delete   ON investment__valuation_holding_lines;
DROP TRIGGER IF EXISTS trg_inv_nav_snapshots_no_update             ON investment__nav_snapshots;
DROP TRIGGER IF EXISTS trg_inv_nav_snapshots_no_delete             ON investment__nav_snapshots;
DROP TRIGGER IF EXISTS trg_inv_aum_snapshots_no_update             ON investment__aum_snapshots;
DROP TRIGGER IF EXISTS trg_inv_aum_snapshots_no_delete             ON investment__aum_snapshots;

DROP FUNCTION IF EXISTS investment__reject_mutation();

-- ---------------------------------------------------------------------------
-- Restore original RULE-based guards. Definitions match
-- 20260428000004 (ledger), 20260428000005 (pricing/valuation), and
-- 20260429000100 (valuation_holding_lines DELETE).
-- ---------------------------------------------------------------------------
CREATE RULE no_update_inv_portfolio_transactions
    AS ON UPDATE TO investment__portfolio_transactions DO INSTEAD NOTHING;
CREATE RULE no_delete_inv_portfolio_transactions
    AS ON DELETE TO investment__portfolio_transactions DO INSTEAD NOTHING;

CREATE RULE no_update_inv_cash_movements
    AS ON UPDATE TO investment__cash_movements DO INSTEAD NOTHING;
CREATE RULE no_delete_inv_cash_movements
    AS ON DELETE TO investment__cash_movements DO INSTEAD NOTHING;

CREATE RULE no_update_inv_price_snapshots
    AS ON UPDATE TO investment__price_snapshots DO INSTEAD NOTHING;
CREATE RULE no_delete_inv_price_snapshots
    AS ON DELETE TO investment__price_snapshots DO INSTEAD NOTHING;

CREATE RULE no_update_inv_valuation_snapshots
    AS ON UPDATE TO investment__valuation_snapshots DO INSTEAD NOTHING;
CREATE RULE no_delete_inv_valuation_snapshots
    AS ON DELETE TO investment__valuation_snapshots DO INSTEAD NOTHING;

CREATE RULE no_update_inv_valuation_lines
    AS ON UPDATE TO investment__valuation_holding_lines DO INSTEAD NOTHING;
CREATE RULE no_delete_inv_valuation_lines
    AS ON DELETE TO investment__valuation_holding_lines DO INSTEAD NOTHING;

CREATE RULE no_update_inv_nav_snapshots
    AS ON UPDATE TO investment__nav_snapshots DO INSTEAD NOTHING;
CREATE RULE no_delete_inv_nav_snapshots
    AS ON DELETE TO investment__nav_snapshots DO INSTEAD NOTHING;

CREATE RULE no_update_inv_aum_snapshots
    AS ON UPDATE TO investment__aum_snapshots DO INSTEAD NOTHING;
CREATE RULE no_delete_inv_aum_snapshots
    AS ON DELETE TO investment__aum_snapshots DO INSTEAD NOTHING;
