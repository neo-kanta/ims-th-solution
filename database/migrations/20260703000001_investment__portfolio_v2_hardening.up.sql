-- =============================================================================
-- Investment Module — Portfolio V2 Hardening (docs/api/portfolio-v2-api-ddd.md
-- section 12, Phase 0)
-- =============================================================================
-- Purpose: make the current schema safer for the Portfolio V2 migration
-- WITHOUT removing any legacy fund_id/contract_id column. This migration is
-- purely additive/defensive:
--   1. portfolio_type (LIVE / SIMULATION / MODEL) on investment__portfolios.
--   2. Global (not just per-fund) active portfolio code uniqueness, so V2
--      portfolioCode routes can resolve a code without ambiguity.
--   3. Composite FK (portfolio_id, fund_id) on every operational table that
--      still carries fund_id, so a row can never point at a portfolio_id
--      that belongs to a different fund than the row's own fund_id.
--   4. CHECK (contract_id = fund_id) on decisions/executions/confirmations,
--      since contract_id today is always meant to equal fund_id (Fund.id IS
--      the cross-module contract_id — see 20260428000002 header comment).
--
-- Preflight (run manually before applying in any environment with existing
-- data — this migration will fail with a constraint-violation error if any
-- of these return rows, which is the intended fail-safe):
--
--   -- duplicate active portfolio codes across funds
--   SELECT code, COUNT(*)
--   FROM investment__portfolios
--   WHERE deleted_at IS NULL
--   GROUP BY code
--   HAVING COUNT(*) > 1;
--
--   -- portfolio transaction fund drift
--   SELECT t.id, t.portfolio_id, t.fund_id, p.fund_id AS portfolio_fund_id
--   FROM investment__portfolio_transactions t
--   JOIN investment__portfolios p ON p.id = t.portfolio_id
--   WHERE t.fund_id <> p.fund_id;
--
--   -- decision fund/contract drift
--   SELECT d.id, d.portfolio_id, d.fund_id, d.contract_id, p.fund_id AS portfolio_fund_id
--   FROM investment__decisions d
--   JOIN investment__portfolios p ON p.id = d.portfolio_id
--   WHERE d.fund_id <> p.fund_id OR d.contract_id <> d.fund_id;
--
--   -- execution fund/contract drift
--   SELECT e.id, e.portfolio_id, e.fund_id, e.contract_id, p.fund_id AS portfolio_fund_id
--   FROM investment__executions e
--   JOIN investment__portfolios p ON p.id = e.portfolio_id
--   WHERE e.fund_id <> p.fund_id OR e.contract_id <> e.fund_id;
--
--   -- trade confirmation fund/contract drift
--   SELECT c.id, c.portfolio_id, c.fund_id, c.contract_id, p.fund_id AS portfolio_fund_id
--   FROM investment__trade_confirmations c
--   JOIN investment__portfolios p ON p.id = c.portfolio_id
--   WHERE c.fund_id <> p.fund_id OR c.contract_id <> c.fund_id;
--
-- Preflight run against the local dev DB (ims_dev) on 2026-07-03: all five
-- queries returned zero rows (portfolios=7, transactions/decisions/
-- executions/confirmations=0), so no dirty-data remediation was required
-- before writing this migration.
-- =============================================================================

-- ---------------------------------------------------------------------------
-- 1. portfolio_type
-- ---------------------------------------------------------------------------
ALTER TABLE investment__portfolios
    ADD COLUMN IF NOT EXISTS portfolio_type VARCHAR(20) NOT NULL DEFAULT 'LIVE';

ALTER TABLE investment__portfolios
    ADD CONSTRAINT chk_inv_portfolios_type
        CHECK (portfolio_type IN ('LIVE', 'SIMULATION', 'MODEL'));

COMMENT ON COLUMN investment__portfolios.portfolio_type IS
    'Portfolio V2 type (docs/api/portfolio-v2-api-ddd.md section 4). Existing rows default to LIVE.';

-- ---------------------------------------------------------------------------
-- 2. Global active portfolio code uniqueness
-- ---------------------------------------------------------------------------
-- Existing uq_inv_portfolios_fund_code_alive (fund_id, code) stays — this adds
-- a stricter global constraint required by V2's portfolioCode route identity
-- (docs/api/portfolio-v2-api-ddd.md section 5: "portfolioCode must be unique
-- among active/non-deleted portfolios").
CREATE UNIQUE INDEX IF NOT EXISTS uq_inv_portfolios_code_alive
ON investment__portfolios (code)
WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 3. Composite identity anchor: (id, fund_id) on portfolios
-- ---------------------------------------------------------------------------
-- Required so operational tables can FK on (portfolio_id, fund_id) and have
-- Postgres enforce that the pair actually belongs together.
ALTER TABLE investment__portfolios
    ADD CONSTRAINT uq_inv_portfolios_id_fund UNIQUE (id, fund_id);

-- ---------------------------------------------------------------------------
-- 4. Composite FK (portfolio_id, fund_id) -> portfolios(id, fund_id)
-- ---------------------------------------------------------------------------
-- Prevents any operational row from drifting to a fund_id that does not
-- match its own portfolio's fund_id, while fund_id/contract_id still exist
-- as legacy compatibility columns.
ALTER TABLE investment__portfolio_transactions
    ADD CONSTRAINT fk_inv_txn_portfolio_fund
        FOREIGN KEY (portfolio_id, fund_id)
        REFERENCES investment__portfolios (id, fund_id);

ALTER TABLE investment__decisions
    ADD CONSTRAINT fk_inv_decision_portfolio_fund
        FOREIGN KEY (portfolio_id, fund_id)
        REFERENCES investment__portfolios (id, fund_id);

ALTER TABLE investment__executions
    ADD CONSTRAINT fk_inv_execution_portfolio_fund
        FOREIGN KEY (portfolio_id, fund_id)
        REFERENCES investment__portfolios (id, fund_id);

ALTER TABLE investment__trade_confirmations
    ADD CONSTRAINT fk_inv_confirmation_portfolio_fund
        FOREIGN KEY (portfolio_id, fund_id)
        REFERENCES investment__portfolios (id, fund_id);

-- ---------------------------------------------------------------------------
-- 5. contract_id must equal fund_id while contract_id still exists
-- ---------------------------------------------------------------------------
-- Fund.id IS the cross-module contract_id (see 20260428000002 header
-- comment); this CHECK makes that invariant explicit at the DB layer until
-- contract_id is dropped from investment operational tables (V2 Phase 3).
ALTER TABLE investment__decisions
    ADD CONSTRAINT chk_inv_decisions_contract_is_fund
        CHECK (contract_id = fund_id);

ALTER TABLE investment__executions
    ADD CONSTRAINT chk_inv_executions_contract_is_fund
        CHECK (contract_id = fund_id);

ALTER TABLE investment__trade_confirmations
    ADD CONSTRAINT chk_inv_confirmations_contract_is_fund
        CHECK (contract_id = fund_id);
