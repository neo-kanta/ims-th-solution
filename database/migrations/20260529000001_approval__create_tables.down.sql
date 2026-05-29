-- =============================================================================
-- Approval Module — rollback generic approval engine
-- =============================================================================
-- Drop in reverse dependency order. CASCADE is implied by FK ON DELETE CASCADE
-- for child tables, but we drop explicitly to be deterministic.
-- =============================================================================

DROP TABLE IF EXISTS approval__signature_records;

-- Remove immutability guards before dropping the events table.
DROP TRIGGER IF EXISTS trg_approval_events_no_update ON approval__events;
DROP TRIGGER IF EXISTS trg_approval_events_no_delete_guard ON approval__events;
DROP TABLE IF EXISTS approval__events;
DROP FUNCTION IF EXISTS prevent_approval_events_modification();

DROP TABLE IF EXISTS approval__tasks;
DROP TABLE IF EXISTS approval__requests;
DROP TABLE IF EXISTS approval__process_stages;
DROP TABLE IF EXISTS approval__process_configs;
DROP TABLE IF EXISTS approval__team_members;
DROP TABLE IF EXISTS approval__team_contracts;
DROP TABLE IF EXISTS approval__teams;
DROP TABLE IF EXISTS approval__group_members;
DROP TABLE IF EXISTS approval__groups;

DROP SEQUENCE IF EXISTS approval__request_no_seq;
