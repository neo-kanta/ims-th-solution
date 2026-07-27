-- =============================================================================
-- Rollback: drop the request-level idempotency key + its partial unique index.
-- =============================================================================

DROP INDEX IF EXISTS uq_inv_cash_req_idempotency;

ALTER TABLE investment__portfolio_cash_requests
    DROP COLUMN IF EXISTS idempotency_key;
