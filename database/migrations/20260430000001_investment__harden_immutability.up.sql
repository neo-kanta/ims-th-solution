-- =============================================================================
-- Investment Module — Harden append-only invariant with RAISE EXCEPTION triggers
-- =============================================================================
-- The original migrations guarded the seven immutable tables with
--   CREATE RULE no_<op>_inv_<table> AS ON <op> TO <table> DO INSTEAD NOTHING;
-- which silently discards UPDATE/DELETE attempts. A bug in repository code or
-- a panicked operator running raw SQL would have no signal that the write was
-- dropped.
--
-- This migration replaces those rules with BEFORE UPDATE / BEFORE DELETE
-- triggers that RAISE EXCEPTION, so any attempted mutation fails loudly with
-- SQLSTATE 23514 (check_violation). Behaviour for legitimate INSERTs is
-- unchanged.
--
-- Reversible: the .down.sql restores the original RULE definitions verbatim.
-- =============================================================================

CREATE OR REPLACE FUNCTION investment__reject_mutation() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION '% on immutable table % is not allowed',
        TG_OP, TG_TABLE_NAME
        USING ERRCODE = 'check_violation';
END;
$$;

-- ---------------------------------------------------------------------------
-- portfolio_transactions
-- ---------------------------------------------------------------------------
DROP RULE IF EXISTS no_update_inv_portfolio_transactions ON investment__portfolio_transactions;
DROP RULE IF EXISTS no_delete_inv_portfolio_transactions ON investment__portfolio_transactions;

CREATE TRIGGER trg_inv_portfolio_transactions_no_update
    BEFORE UPDATE ON investment__portfolio_transactions
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();
CREATE TRIGGER trg_inv_portfolio_transactions_no_delete
    BEFORE DELETE ON investment__portfolio_transactions
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();

-- ---------------------------------------------------------------------------
-- cash_movements
-- ---------------------------------------------------------------------------
DROP RULE IF EXISTS no_update_inv_cash_movements ON investment__cash_movements;
DROP RULE IF EXISTS no_delete_inv_cash_movements ON investment__cash_movements;

CREATE TRIGGER trg_inv_cash_movements_no_update
    BEFORE UPDATE ON investment__cash_movements
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();
CREATE TRIGGER trg_inv_cash_movements_no_delete
    BEFORE DELETE ON investment__cash_movements
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();

-- ---------------------------------------------------------------------------
-- price_snapshots
-- ---------------------------------------------------------------------------
DROP RULE IF EXISTS no_update_inv_price_snapshots ON investment__price_snapshots;
DROP RULE IF EXISTS no_delete_inv_price_snapshots ON investment__price_snapshots;

CREATE TRIGGER trg_inv_price_snapshots_no_update
    BEFORE UPDATE ON investment__price_snapshots
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();
CREATE TRIGGER trg_inv_price_snapshots_no_delete
    BEFORE DELETE ON investment__price_snapshots
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();

-- ---------------------------------------------------------------------------
-- valuation_snapshots
-- ---------------------------------------------------------------------------
DROP RULE IF EXISTS no_update_inv_valuation_snapshots ON investment__valuation_snapshots;
DROP RULE IF EXISTS no_delete_inv_valuation_snapshots ON investment__valuation_snapshots;

CREATE TRIGGER trg_inv_valuation_snapshots_no_update
    BEFORE UPDATE ON investment__valuation_snapshots
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();
CREATE TRIGGER trg_inv_valuation_snapshots_no_delete
    BEFORE DELETE ON investment__valuation_snapshots
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();

-- ---------------------------------------------------------------------------
-- valuation_holding_lines
-- ---------------------------------------------------------------------------
DROP RULE IF EXISTS no_update_inv_valuation_lines ON investment__valuation_holding_lines;
DROP RULE IF EXISTS no_delete_inv_valuation_lines ON investment__valuation_holding_lines;

CREATE TRIGGER trg_inv_valuation_holding_lines_no_update
    BEFORE UPDATE ON investment__valuation_holding_lines
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();
CREATE TRIGGER trg_inv_valuation_holding_lines_no_delete
    BEFORE DELETE ON investment__valuation_holding_lines
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();

-- ---------------------------------------------------------------------------
-- nav_snapshots
-- ---------------------------------------------------------------------------
DROP RULE IF EXISTS no_update_inv_nav_snapshots ON investment__nav_snapshots;
DROP RULE IF EXISTS no_delete_inv_nav_snapshots ON investment__nav_snapshots;

CREATE TRIGGER trg_inv_nav_snapshots_no_update
    BEFORE UPDATE ON investment__nav_snapshots
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();
CREATE TRIGGER trg_inv_nav_snapshots_no_delete
    BEFORE DELETE ON investment__nav_snapshots
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();

-- ---------------------------------------------------------------------------
-- aum_snapshots
-- ---------------------------------------------------------------------------
DROP RULE IF EXISTS no_update_inv_aum_snapshots ON investment__aum_snapshots;
DROP RULE IF EXISTS no_delete_inv_aum_snapshots ON investment__aum_snapshots;

CREATE TRIGGER trg_inv_aum_snapshots_no_update
    BEFORE UPDATE ON investment__aum_snapshots
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();
CREATE TRIGGER trg_inv_aum_snapshots_no_delete
    BEFORE DELETE ON investment__aum_snapshots
    FOR EACH ROW EXECUTE FUNCTION investment__reject_mutation();
