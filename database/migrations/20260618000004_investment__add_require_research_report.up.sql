-- =============================================================================
-- Investment Module — Add require_research_report_for_decision flag to funds
-- =============================================================================
-- Adds a backend-enforced policy gate that mandates a linked research report
-- before a decision can be submitted for approval. This is distinct from
-- require_pretrade_preview which only controls the Operation-tab UX.
-- =============================================================================

ALTER TABLE investment__funds
    ADD COLUMN IF NOT EXISTS require_research_report_for_decision BOOLEAN NOT NULL DEFAULT FALSE;
