-- =============================================================================
-- Investment module — allow ledger transactions and trading on fund-less
-- portfolios (Portfolio V2, phase 2)
-- =============================================================================
-- Phase 1 (20260723000001) allowed a fund-less investment__portfolios row.
-- This phase extends that to the money-movement tables so a fund-less
-- portfolio can actually post cash/ledger transactions and run the
-- Decision -> Execution -> Confirmation trading workflow.
--
-- Composite FKs (portfolio_id, fund_id) -> investment__portfolios(id, fund_id)
-- on all four tables already tolerate NULL (Postgres MATCH SIMPLE skips the
-- check when any FK column is NULL), so no FK changes are needed here — only
-- dropping each table's own NOT NULL.
--
-- Neither of these tables carries a contract_id column anymore (already
-- dropped in an earlier migration — verified against the live schema), so
-- there is no CHECK(contract_id = fund_id) to reconcile.
-- =============================================================================

ALTER TABLE investment__portfolio_transactions ALTER COLUMN fund_id DROP NOT NULL;
ALTER TABLE investment__decisions ALTER COLUMN fund_id DROP NOT NULL;
ALTER TABLE investment__executions ALTER COLUMN fund_id DROP NOT NULL;
ALTER TABLE investment__trade_confirmations ALTER COLUMN fund_id DROP NOT NULL;
