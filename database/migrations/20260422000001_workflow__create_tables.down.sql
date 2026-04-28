-- =============================================================================
-- Workflow Management Module — Rollback
-- Drop order respects FK constraints (child tables first).
-- =============================================================================

DROP TABLE IF EXISTS workflow__approval_records;
DROP TABLE IF EXISTS workflow__transition_log;
DROP TABLE IF EXISTS workflow__day_states;
