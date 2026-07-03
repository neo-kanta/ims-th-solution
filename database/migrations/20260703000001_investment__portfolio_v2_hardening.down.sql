-- Reverse of 20260703000001_investment__portfolio_v2_hardening.up.sql.
-- Drop constraints/indexes/column in exact reverse order.

ALTER TABLE investment__trade_confirmations
    DROP CONSTRAINT IF EXISTS chk_inv_confirmations_contract_is_fund;

ALTER TABLE investment__executions
    DROP CONSTRAINT IF EXISTS chk_inv_executions_contract_is_fund;

ALTER TABLE investment__decisions
    DROP CONSTRAINT IF EXISTS chk_inv_decisions_contract_is_fund;

ALTER TABLE investment__trade_confirmations
    DROP CONSTRAINT IF EXISTS fk_inv_confirmation_portfolio_fund;

ALTER TABLE investment__executions
    DROP CONSTRAINT IF EXISTS fk_inv_execution_portfolio_fund;

ALTER TABLE investment__decisions
    DROP CONSTRAINT IF EXISTS fk_inv_decision_portfolio_fund;

ALTER TABLE investment__portfolio_transactions
    DROP CONSTRAINT IF EXISTS fk_inv_txn_portfolio_fund;

ALTER TABLE investment__portfolios
    DROP CONSTRAINT IF EXISTS uq_inv_portfolios_id_fund;

DROP INDEX IF EXISTS uq_inv_portfolios_code_alive;

ALTER TABLE investment__portfolios
    DROP CONSTRAINT IF EXISTS chk_inv_portfolios_type;

ALTER TABLE investment__portfolios
    DROP COLUMN IF EXISTS portfolio_type;
