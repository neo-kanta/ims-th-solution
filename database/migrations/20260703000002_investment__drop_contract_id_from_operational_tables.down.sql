-- Reverses 20260703000002_investment__drop_contract_id_from_operational_tables.up.sql

ALTER TABLE investment__decisions
    ADD COLUMN IF NOT EXISTS contract_id UUID;
UPDATE investment__decisions SET contract_id = fund_id WHERE contract_id IS NULL;
ALTER TABLE investment__decisions
    ALTER COLUMN contract_id SET NOT NULL,
    ADD CONSTRAINT chk_inv_decisions_contract_is_fund CHECK (contract_id = fund_id);

ALTER TABLE investment__executions
    ADD COLUMN IF NOT EXISTS contract_id UUID;
UPDATE investment__executions SET contract_id = fund_id WHERE contract_id IS NULL;
ALTER TABLE investment__executions
    ALTER COLUMN contract_id SET NOT NULL,
    ADD CONSTRAINT chk_inv_executions_contract_is_fund CHECK (contract_id = fund_id);

ALTER TABLE investment__trade_confirmations
    ADD COLUMN IF NOT EXISTS contract_id UUID;
UPDATE investment__trade_confirmations SET contract_id = fund_id WHERE contract_id IS NULL;
ALTER TABLE investment__trade_confirmations
    ALTER COLUMN contract_id SET NOT NULL,
    ADD CONSTRAINT chk_inv_confirmations_contract_is_fund CHECK (contract_id = fund_id);

DROP INDEX IF EXISTS idx_inv_confirmation_fund_date;

CREATE INDEX IF NOT EXISTS idx_inv_confirmation_contract_date
    ON investment__trade_confirmations (contract_id, business_date DESC);
