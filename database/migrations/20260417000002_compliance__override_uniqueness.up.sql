-- ============================================================
-- compliance_overrides — enforce one override per breach
-- ============================================================
--
-- The override workflow commits two state changes atomically:
--   1. INSERT into compliance_overrides
--   2. UPDATE compliance_breaches SET status = 'OVERRIDDEN'
--
-- This unique constraint is the database-level safety net against
-- the race where two concurrent override requests both see a breach
-- in OPEN status and try to commit — the second one will fail with
-- SQLSTATE 23505, which the application maps to ErrOverrideAlreadyExists.
--
-- Combined with SELECT ... FOR UPDATE on the breach row inside the
-- same transaction, this guarantees at most one accepted override
-- per breach.

ALTER TABLE compliance_overrides
    ADD CONSTRAINT uq_compliance_ov_breach UNIQUE (breach_id);

-- The existing btree index idx_compliance_ov_breach is now redundant
-- with the unique constraint's backing index, but we keep it for
-- backwards compatibility with any query planner hints. PostgreSQL
-- will not double-scan.
