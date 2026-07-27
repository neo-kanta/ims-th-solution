-- =============================================================================
-- Investment module — canonical request fingerprint for LIVE cash-approval
-- =============================================================================
-- IMS-PORTFOLIO-FUND-OPTIONAL Stage 2 (work package G2, Codex finding 3).
--
-- The idempotency key (20260726000001) only dedups when the CLIENT supplies an
-- Idempotency-Key header. This adds a deterministic, server-computed fingerprint
-- over the canonical request payload
--   {submitter_id, portfolio_id, transaction_type, amount, fees, currency,
--    value_date, memo}
-- (hex sha256, 64 chars) so that:
--
--   1. A keyed retry can be validated: same key + SAME fingerprint is a true
--      idempotent replay; same key + DIFFERENT fingerprint is rejected instead
--      of silently resolving to the original movement (a money-path hazard).
--
--   2. A submission WITHOUT a client key still dedups: the application derives a
--      server-side idempotency key ('auto:' || fingerprint) and writes it into
--      idempotency_key, so a genuine retry of the byte-identical request collapses
--      onto the original via the existing partial unique index. A different-payload
--      no-key submission gets a different fingerprint → different derived key → a
--      new request, as before.
--
-- The column is NULLABLE (no NOT NULL / no backfill): rows created before this
-- migration keep NULL and the application falls back to a field-by-field
-- comparison for them. New rows always populate it. This keeps the migration
-- additive and the down migration a clean, honest drop.
--
-- Forward-only migration: the base table (20260724000001) and the idempotency
-- column (20260726000001) are already applied, so this adds the column
-- additively rather than editing an applied migration (repo forward-migration
-- discipline, see G1).
-- =============================================================================

ALTER TABLE investment__portfolio_cash_requests
    ADD COLUMN request_fingerprint VARCHAR(64);

-- Reconciliation lookup: find a portfolio's request(s) by canonical fingerprint.
-- NOT unique — two DISTINCT client keys may legitimately carry the same payload
-- fingerprint; race-safe dedup is enforced by the (portfolio_id, idempotency_key)
-- partial unique index, not by this index.
CREATE INDEX idx_inv_cash_req_fingerprint
    ON investment__portfolio_cash_requests (portfolio_id, request_fingerprint)
    WHERE request_fingerprint IS NOT NULL;

COMMENT ON COLUMN investment__portfolio_cash_requests.request_fingerprint IS
    'Deterministic hex sha256 over the canonical request payload {submitter_id, portfolio_id, transaction_type, amount, fees, currency, value_date, memo}. Written once at creation and never mutated. Used to (1) validate that a keyed retry carries the SAME movement and (2) derive a server-side idempotency key so no-key retries of an identical request still dedup. NULL for rows created before this column existed.';
