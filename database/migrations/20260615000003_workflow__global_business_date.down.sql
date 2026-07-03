DROP INDEX IF EXISTS idx_wf_transition_log_business_date;

ALTER TABLE workflow__day_states
    DROP CONSTRAINT IF EXISTS uq_wf_day_states_business_date;

ALTER TABLE workflow__day_states
    ADD CONSTRAINT uq_wf_day_states_contract_date UNIQUE (contract_id, business_date);

ALTER TABLE workflow__approval_records
    ALTER COLUMN contract_id SET NOT NULL;

ALTER TABLE workflow__transition_log
    ALTER COLUMN contract_id SET NOT NULL;

ALTER TABLE workflow__day_states
    ALTER COLUMN contract_id SET NOT NULL;
