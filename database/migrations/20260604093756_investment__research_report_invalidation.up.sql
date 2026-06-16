-- =============================================================================
-- Investment Research Report — INVALIDATED lifecycle status
-- =============================================================================
-- Adds the INVALIDATED terminal state to both report_status and review_status
-- and records who invalidated the report, when, and why.
--
-- INVALIDATED is a one-way terminal state: an invalidated report cannot be
-- referenced by an investment decision, cannot be edited, cannot be deleted,
-- and cannot transition back to ACTIVE. Application policy enforces this.
-- =============================================================================

ALTER TABLE investment__research_reports
    ADD COLUMN IF NOT EXISTS invalidated_at      TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS invalidated_by      UUID,
    ADD COLUMN IF NOT EXISTS invalidation_reason TEXT NOT NULL DEFAULT '';

-- Relax report_status CHECK to include INVALIDATED.
ALTER TABLE investment__research_reports
    DROP CONSTRAINT IF EXISTS chk_inv_research_report_status;

ALTER TABLE investment__research_reports
    ADD CONSTRAINT chk_inv_research_report_status
        CHECK (report_status IN ('DRAFT', 'ACTIVE', 'EXPIRED', 'REJECTED', 'INVALIDATED'));

-- Relax review_status CHECK to include INVALIDATED.
ALTER TABLE investment__research_reports
    DROP CONSTRAINT IF EXISTS chk_inv_research_review_status;

ALTER TABLE investment__research_reports
    ADD CONSTRAINT chk_inv_research_review_status
        CHECK (review_status IN ('NOT_SUBMITTED', 'SUBMITTED', 'REVIEW_COMPLETED', 'INVALIDATED'));

-- Coherence: if either status is INVALIDATED, both must be — and the trio
-- (reason, by, at) must be populated. This is enforced as a single CHECK so
-- the DB itself prevents a partial invalidation row.
ALTER TABLE investment__research_reports
    ADD CONSTRAINT chk_inv_research_invalidation_coherent
        CHECK (
            (report_status <> 'INVALIDATED' AND review_status <> 'INVALIDATED')
            OR (
                report_status = 'INVALIDATED'
                AND review_status = 'INVALIDATED'
                AND invalidated_at IS NOT NULL
                AND invalidated_by IS NOT NULL
                AND length(trim(invalidation_reason)) >= 20
            )
        );

COMMENT ON COLUMN investment__research_reports.invalidated_at      IS 'UTC timestamp when the report was invalidated. NULL until invalidation.';
COMMENT ON COLUMN investment__research_reports.invalidated_by      IS 'iam_users.id of the actor who invalidated the report.';
COMMENT ON COLUMN investment__research_reports.invalidation_reason IS 'Free-text reason recorded at invalidation. Minimum 20 characters when status is INVALIDATED.';
