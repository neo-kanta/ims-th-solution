-- =============================================================================
-- Investment Module — Add contract_code to funds
-- =============================================================================
-- Adds a nullable contract_code column to investment__funds.
-- This exposes a human-readable cross-module contract identifier, separate from
-- the fund's internal code, for display in the investment workspace toolbar.
-- Nullable because existing funds predate this column; back-filled by ops.
--
-- The unique partial index prevents duplicate contract codes on live funds while
-- allowing multiple soft-deleted funds to share a code.
-- =============================================================================

ALTER TABLE investment__funds
    ADD COLUMN contract_code VARCHAR(40);

CREATE UNIQUE INDEX uq_inv_funds_contract_code_alive
    ON investment__funds (contract_code)
    WHERE deleted_at IS NULL AND contract_code IS NOT NULL;

COMMENT ON COLUMN investment__funds.contract_code IS
    'Optional cross-module contract label shown in the workspace toolbar. '
    'Distinct from code; back-filled manually for funds predating this column.';
