-- Revert: restore NOT NULL on contract_id.
--
-- NOT SAFELY REVERSIBLE once portfolio-only compliance rows exist. A NULL
-- contract_id here means "this check ran for a portfolio with no fund" — it
-- is not a data gap, so there is no honest backfill value (a sentinel
-- contract UUID would misrepresent the audit trail as contract-scoped when
-- it was not). This down migration intentionally does not invent one.
--
-- The guard below raises an explicit, actionable error instead of letting
-- the ALTER TABLE fail with a bare not-null-violation. If it fires, the
-- NOT NULL constraint cannot be restored without deleting or otherwise
-- rewriting the affected portfolio-only rows, which this migration will not
-- do silently.
DO $$
DECLARE
    orphan_checks  bigint;
    orphan_breaches bigint;
BEGIN
    SELECT count(*) INTO orphan_checks
        FROM compliance_check_records WHERE contract_id IS NULL;
    SELECT count(*) INTO orphan_breaches
        FROM compliance_breaches WHERE contract_id IS NULL;

    IF orphan_checks > 0 OR orphan_breaches > 0 THEN
        RAISE EXCEPTION
            'cannot restore contract_id NOT NULL: % compliance_check_records and % compliance_breaches rows have NULL contract_id (portfolio-only Compliance V2 checks). This down migration is not reversible while those rows exist — see migration comment.',
            orphan_checks, orphan_breaches
            USING ERRCODE = 'check_violation';
    END IF;
END;
$$;

ALTER TABLE compliance_check_records
    ALTER COLUMN contract_id SET NOT NULL;

ALTER TABLE compliance_breaches
    ALTER COLUMN contract_id SET NOT NULL;
