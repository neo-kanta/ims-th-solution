-- Workflow daily state is global per business date.
-- This migration is intentionally non-destructive: if historical contract-scoped
-- rows already contain more than one workflow day for the same date, it aborts
-- and lets an operator choose the canonical row before applying global uniqueness.

DO $$
DECLARE
    duplicate_dates TEXT;
BEGIN
    SELECT string_agg(business_date::TEXT, ', ' ORDER BY business_date)
      INTO duplicate_dates
      FROM (
          SELECT business_date
          FROM workflow__day_states
          GROUP BY business_date
          HAVING COUNT(*) > 1
      ) d;

    IF duplicate_dates IS NOT NULL THEN
        RAISE EXCEPTION
            'workflow__day_states cannot be made global; duplicate business_date rows exist: %',
            duplicate_dates;
    END IF;
END $$;

ALTER TABLE workflow__day_states
    ALTER COLUMN contract_id DROP NOT NULL;

ALTER TABLE workflow__transition_log
    ALTER COLUMN contract_id DROP NOT NULL;

ALTER TABLE workflow__approval_records
    ALTER COLUMN contract_id DROP NOT NULL;

ALTER TABLE workflow__day_states
    DROP CONSTRAINT IF EXISTS uq_wf_day_states_contract_date;

ALTER TABLE workflow__day_states
    ADD CONSTRAINT uq_wf_day_states_business_date UNIQUE (business_date);

CREATE INDEX IF NOT EXISTS idx_wf_transition_log_business_date
    ON workflow__transition_log (business_date, occurred_at ASC);
