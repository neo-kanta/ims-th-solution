-- =============================================================================
-- Investment Module — Rollback Pricing & Valuation
-- =============================================================================

DROP RULE IF EXISTS no_delete_inv_aum_snapshots         ON investment__aum_snapshots;
DROP RULE IF EXISTS no_update_inv_aum_snapshots         ON investment__aum_snapshots;
DROP RULE IF EXISTS no_delete_inv_nav_snapshots         ON investment__nav_snapshots;
DROP RULE IF EXISTS no_update_inv_nav_snapshots         ON investment__nav_snapshots;
DROP RULE IF EXISTS no_update_inv_valuation_lines       ON investment__valuation_holding_lines;
DROP RULE IF EXISTS no_delete_inv_valuation_snapshots   ON investment__valuation_snapshots;
DROP RULE IF EXISTS no_update_inv_valuation_snapshots   ON investment__valuation_snapshots;
DROP RULE IF EXISTS no_delete_inv_price_snapshots       ON investment__price_snapshots;
DROP RULE IF EXISTS no_update_inv_price_snapshots       ON investment__price_snapshots;

DROP TABLE IF EXISTS investment__aum_snapshots;
DROP TABLE IF EXISTS investment__nav_snapshots;
DROP TABLE IF EXISTS investment__valuation_holding_lines;
DROP TABLE IF EXISTS investment__valuation_snapshots;
DROP TABLE IF EXISTS investment__price_snapshots;
