-- =============================================================================
-- Investment module — allow fund-less portfolios (Portfolio V2)
-- =============================================================================
-- Drops the NOT NULL constraint on investment__portfolios.fund_id so a
-- portfolio can be created without being bound to a fund ("Bind with Fund: N"
-- on the Create Portfolio form). Access to a fund-less portfolio is then
-- governed by a portfolio_id-scoped grant in permission_data_rights (already
-- supported by that table/migration 20260521000001) instead of a fund-scoped
-- grant.
--
-- Deliberately NOT touched: investment__decisions/executions/trade_confirmations
-- keep fund_id NOT NULL and CHECK(contract_id = fund_id) (added by
-- 20260703000001). Trading on a fund-less portfolio is out of scope — the
-- application layer rejects decision/execution creation on such a portfolio
-- with a 422 rather than relying on these constraints. Because those child
-- tables always have a real fund_id, the composite FKs
-- (portfolio_id, fund_id) -> investment__portfolios(id, fund_id) added by
-- 20260703000001 are unaffected: a fund-less portfolio simply never has any
-- matching child rows.
-- =============================================================================

ALTER TABLE investment__portfolios
    ALTER COLUMN fund_id DROP NOT NULL;
