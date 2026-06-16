-- Reverse the INVALIDATED lifecycle additions for investment__research_reports.
--
-- Existing INVALIDATED rows (if any) are rejected by the restored CHECKs and
-- would fail the migration; flip them to REJECTED before downgrading.

ALTER TABLE investment__research_reports
    DROP CONSTRAINT IF EXISTS chk_inv_research_invalidation_coherent;

ALTER TABLE investment__research_reports
    DROP CONSTRAINT IF EXISTS chk_inv_research_review_status;

ALTER TABLE investment__research_reports
    ADD CONSTRAINT chk_inv_research_review_status
        CHECK (review_status IN ('NOT_SUBMITTED', 'SUBMITTED', 'REVIEW_COMPLETED'));

ALTER TABLE investment__research_reports
    DROP CONSTRAINT IF EXISTS chk_inv_research_report_status;

ALTER TABLE investment__research_reports
    ADD CONSTRAINT chk_inv_research_report_status
        CHECK (report_status IN ('DRAFT', 'ACTIVE', 'EXPIRED', 'REJECTED'));

ALTER TABLE investment__research_reports
    DROP COLUMN IF EXISTS invalidation_reason,
    DROP COLUMN IF EXISTS invalidated_by,
    DROP COLUMN IF EXISTS invalidated_at;
