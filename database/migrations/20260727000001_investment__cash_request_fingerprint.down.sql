-- =============================================================================
-- Rollback: drop the canonical request fingerprint column + its lookup index.
-- =============================================================================

DROP INDEX IF EXISTS idx_inv_cash_req_fingerprint;

ALTER TABLE investment__portfolio_cash_requests
    DROP COLUMN IF EXISTS request_fingerprint;
