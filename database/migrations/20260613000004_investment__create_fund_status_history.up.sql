-- =============================================================================
-- Investment Module — Fund status history table
-- =============================================================================
-- Immutable append-only record of every fund lifecycle transition.
-- One row per status change; from_status is NULL for the initial creation row.
-- No UPDATE or DELETE should ever touch this table.
-- =============================================================================

CREATE TABLE investment__fund_status_history (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    fund_id     UUID         NOT NULL
                    REFERENCES investment__funds(id) ON DELETE RESTRICT,

    from_status VARCHAR(20),   -- NULL for the initial DRAFT creation event
    to_status   VARCHAR(20)  NOT NULL,

    actor_id    UUID         REFERENCES iam_users(id),
    reason      TEXT,

    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_inv_fsh_to_status CHECK (to_status IN (
        'DRAFT', 'PENDING_APPROVAL', 'ACTIVE', 'SUSPENDED', 'CLOSED', 'REJECTED'
    )),
    CONSTRAINT chk_inv_fsh_from_status CHECK (from_status IS NULL OR from_status IN (
        'DRAFT', 'PENDING_APPROVAL', 'ACTIVE', 'SUSPENDED', 'CLOSED', 'REJECTED'
    ))
);

CREATE INDEX idx_inv_fsh_fund
    ON investment__fund_status_history (fund_id, created_at DESC);

CREATE INDEX idx_inv_fsh_actor
    ON investment__fund_status_history (actor_id, created_at DESC);

COMMENT ON TABLE  investment__fund_status_history IS 'Immutable lifecycle event log for funds (contracts). Never updated or deleted.';
COMMENT ON COLUMN investment__fund_status_history.from_status IS 'NULL only for the initial creation event.';
COMMENT ON COLUMN investment__fund_status_history.actor_id    IS 'NULL for system-initiated transitions (e.g. approval callback).';

REVOKE UPDATE, DELETE ON investment__fund_status_history FROM PUBLIC;
