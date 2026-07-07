-- =============================================================================
-- Investment Module - Drop contract_id From Operational Tables
-- (Portfolio V2 Phase 3, part 1 - see docs/handoff/portfolio-v2-cleanup-readiness.md)
-- =============================================================================
-- Purpose: contract_id on investment__decisions/executions/trade_confirmations
-- has always been a bare mirror of fund_id (no FK of its own), and is now
-- DB-enforced equal by the 20260703000001 CHECK constraints. All backend
-- .ContractID reads on Decision/Execution/TradeConfirmation have been
-- repointed to .FundID; V1 response DTOs no longer serialize contract_id for
-- these three resources. fund_id itself is untouched - it remains the real
-- FK into investment__funds and stays until Phase 3's larger fund_id removal.
--
-- Preflight (run manually before applying in any environment with existing
-- data - should return zero rows; a non-zero result means the equality this
-- migration assumes has already been violated and needs investigating first):
--
--   SELECT id FROM investment__decisions          WHERE contract_id <> fund_id;
--   SELECT id FROM investment__executions         WHERE contract_id <> fund_id;
--   SELECT id FROM investment__trade_confirmations WHERE contract_id <> fund_id;
-- =============================================================================

-- idx_inv_confirmation_contract_date was the table's only (..., business_date)
-- composite index; recreate it on fund_id so ListByFundDate keeps index
-- support (decisions/executions already have their own fund_id+date indexes).
DROP INDEX IF EXISTS idx_inv_confirmation_contract_date;

CREATE INDEX IF NOT EXISTS idx_inv_confirmation_fund_date
    ON investment__trade_confirmations (fund_id, business_date DESC);

ALTER TABLE investment__decisions
    DROP CONSTRAINT IF EXISTS chk_inv_decisions_contract_is_fund,
    DROP COLUMN IF EXISTS contract_id;

ALTER TABLE investment__executions
    DROP CONSTRAINT IF EXISTS chk_inv_executions_contract_is_fund,
    DROP COLUMN IF EXISTS contract_id;

ALTER TABLE investment__trade_confirmations
    DROP CONSTRAINT IF EXISTS chk_inv_confirmations_contract_is_fund,
    DROP COLUMN IF EXISTS contract_id;
