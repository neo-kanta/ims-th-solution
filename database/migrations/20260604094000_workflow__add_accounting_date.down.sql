ALTER TABLE workflow__day_states
    DROP CONSTRAINT IF EXISTS chk_wf_accounting_date_after_business_date;

ALTER TABLE workflow__day_states
    DROP COLUMN IF EXISTS prev_accounting_date,
    DROP COLUMN IF EXISTS accounting_date;
