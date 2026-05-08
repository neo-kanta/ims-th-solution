-- =============================================================================
-- Investment Module — Rollback Portfolio Ledger
-- =============================================================================

DROP RULE IF EXISTS no_delete_inv_cash_movements          ON investment__cash_movements;
DROP RULE IF EXISTS no_update_inv_cash_movements          ON investment__cash_movements;
DROP RULE IF EXISTS no_delete_inv_portfolio_transactions  ON investment__portfolio_transactions;
DROP RULE IF EXISTS no_update_inv_portfolio_transactions  ON investment__portfolio_transactions;

DROP TABLE IF EXISTS investment__cash_balances;
DROP TABLE IF EXISTS investment__cash_movements;
DROP TABLE IF EXISTS investment__portfolio_positions;
DROP TABLE IF EXISTS investment__portfolio_transactions;
