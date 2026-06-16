DROP INDEX IF EXISTS uq_inv_funds_contract_code_alive;

ALTER TABLE investment__funds
    DROP COLUMN IF EXISTS contract_code;
