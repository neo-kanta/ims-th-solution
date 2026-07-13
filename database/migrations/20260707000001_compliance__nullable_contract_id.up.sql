-- =============================================================================
-- Compliance / IRG - Portfolio Compliance V2 (additive)
-- =============================================================================
-- Portfolio is now the primary compliance identity; fund_id/contract_id is
-- optional at the portfolio boundary. compliance_check_records.contract_id and
-- compliance_breaches.contract_id were NOT NULL from the V1 schema, which
-- assumed every check was contract-scoped. Portfolio-only checks (no fund_id)
-- must persist contract_id as NULL instead of a zero UUID sentinel.
--
-- contract_id is NOT dropped - V1 contract-scoped checks keep writing it, and
-- existing rows are untouched (this only relaxes the constraint).

ALTER TABLE compliance_check_records
    ALTER COLUMN contract_id DROP NOT NULL;

ALTER TABLE compliance_breaches
    ALTER COLUMN contract_id DROP NOT NULL;

COMMENT ON COLUMN compliance_check_records.contract_id IS 'Fund/contract context. NULL for portfolio-only checks where the portfolio has no fund_id (Portfolio Compliance V2).';
COMMENT ON COLUMN compliance_breaches.contract_id IS 'Fund/contract context. NULL for portfolio-only checks where the portfolio has no fund_id (Portfolio Compliance V2).';
