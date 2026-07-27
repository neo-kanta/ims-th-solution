-- =============================================================================
-- Investment module — request-level idempotency for LIVE cash-approval submits
-- =============================================================================
-- IMS-PORTFOLIO-FUND-OPTIONAL Stage 2 (work package G2).
--
-- A client that times out waiting for the submit response and retries would,
-- before this change, re-run submitCashForApproval and create a SECOND cash
-- request AND a SECOND approval request. This migration adds an optional
-- client-supplied idempotency key (sourced from the `Idempotency-Key` HTTP
-- header) plus a PARTIAL unique index so a retry with the same key collapses
-- onto the original request instead of duplicating it.
--
-- The index is PARTIAL (WHERE idempotency_key IS NOT NULL) so that requests
-- submitted WITHOUT a key (the documented non-deduplicated path) never collide
-- with one another — many NULL keys per portfolio remain legal.
--
-- Forward-only migration: the base table (20260724000001) is already applied
-- locally, so we add the column additively rather than editing that file
-- (matches the repo's forward-migration discipline, see G1).
-- =============================================================================

ALTER TABLE investment__portfolio_cash_requests
    ADD COLUMN idempotency_key VARCHAR(255);

-- A given portfolio may reuse a key only for the SAME logical request. The
-- partial predicate keeps NULL (no-key) rows out of the uniqueness set.
CREATE UNIQUE INDEX uq_inv_cash_req_idempotency
    ON investment__portfolio_cash_requests (portfolio_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

COMMENT ON COLUMN investment__portfolio_cash_requests.idempotency_key IS
    'Optional client-supplied idempotency key (Idempotency-Key header). When present, a retry with the same (portfolio_id, key) returns the original request instead of creating a duplicate cash request / approval request. NULL means the request was not deduplicated.';
