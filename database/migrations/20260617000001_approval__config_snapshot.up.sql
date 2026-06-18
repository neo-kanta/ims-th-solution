-- Add config_snapshot to approval__requests so that in-flight approvals use
-- the stage configuration that was active at submit time, preventing
-- retroactive changes from admin config edits.
ALTER TABLE approval__requests
    ADD COLUMN IF NOT EXISTS config_snapshot JSONB NOT NULL DEFAULT '[]'::jsonb;

COMMENT ON COLUMN approval__requests.config_snapshot IS
    'Frozen copy of approval__process_stages rows at the moment this request was submitted. '
    'Runtime stage advancement reads from this snapshot to prevent in-flight mutation.';
