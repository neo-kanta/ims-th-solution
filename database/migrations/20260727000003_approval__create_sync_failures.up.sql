-- =============================================================================
-- Approval module — durable subject-sync failure replay (G2 item 4)
-- =============================================================================
-- An approval decision (approve/reject/revoke) commits inside its own DB
-- transaction and is never rolled back for a failure in the POST-commit
-- subject-sync callback (contract.ApprovalSubjectCallback.OnApproved/
-- OnRejected) that notifies the owning business module (e.g. investment) to
-- materialize the real effect. Previously such a failure was only logged to
-- the audit trail with a "retryable": true flag — durable, but not
-- operator-actionable: nothing tracked how many times a failure had been
-- retried, and nothing let an operator trigger a bounded retry.
--
-- This table is that durable, operator-visible record. One row per failed
-- callback invocation attempt sequence for a given (approval_request_id,
-- outcome) pair; an operator-triggered retry increments attempt_count and
-- flips status to RESOLVED on success or EXHAUSTED once max_attempts is used
-- up without success.
-- =============================================================================

BEGIN;

CREATE TABLE approval__sync_failures (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_request_id   UUID NOT NULL REFERENCES approval__requests(id) ON DELETE CASCADE,
    subject_type          VARCHAR(50) NOT NULL,
    subject_id            UUID NOT NULL,
    outcome               VARCHAR(20) NOT NULL,
    reason                TEXT,

    attempt_count         INT NOT NULL DEFAULT 1,
    max_attempts          INT NOT NULL DEFAULT 5,
    last_error            TEXT NOT NULL,
    status                VARCHAR(20) NOT NULL DEFAULT 'PENDING',

    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at           TIMESTAMPTZ,
    resolved_by           UUID,

    CONSTRAINT chk_sync_failure_outcome CHECK (outcome IN ('APPROVED', 'REJECTED', 'REVOKED')),
    CONSTRAINT chk_sync_failure_status CHECK (status IN ('PENDING', 'RESOLVED', 'EXHAUSTED')),
    CONSTRAINT chk_sync_failure_attempts CHECK (attempt_count >= 1 AND attempt_count <= max_attempts),
    CONSTRAINT chk_sync_failure_resolved CHECK (
        (status = 'PENDING' AND resolved_at IS NULL AND resolved_by IS NULL)
        OR (status <> 'PENDING')
    )
);

-- Operator inbox query: unresolved failures, oldest first.
CREATE INDEX idx_sync_failures_status ON approval__sync_failures (status, created_at);
-- Reconciliation lookup: every sync-failure record for one approval request.
CREATE INDEX idx_sync_failures_request ON approval__sync_failures (approval_request_id);

COMMENT ON TABLE approval__sync_failures IS
    'Durable, operator-visible record of an approval decision whose post-commit subject-sync callback failed. Never blocks or reverses the approval decision itself; exists purely so an operator can discover and bounded-retry the missed business-module notification.';

COMMIT;
