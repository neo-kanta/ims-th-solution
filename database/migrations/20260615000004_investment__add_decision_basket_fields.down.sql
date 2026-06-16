-- Revert basket-fields migration.
-- NOTE: side/instrument_code NOT NULL cannot be restored if null rows exist.
ALTER TABLE investment__decisions
    DROP CONSTRAINT IF EXISTS chk_inv_decision_decision_type,
    DROP CONSTRAINT IF EXISTS chk_inv_decision_process_type,
    DROP CONSTRAINT IF EXISTS chk_inv_decision_product_type,
    DROP CONSTRAINT IF EXISTS chk_inv_decision_side;

ALTER TABLE investment__decisions
    ADD CONSTRAINT chk_inv_decision_side CHECK (side IN ('BUY','SELL'));

ALTER TABLE investment__decisions
    ADD CONSTRAINT chk_inv_decision_quantity_or_amount
        CHECK (quantity IS NOT NULL OR amount IS NOT NULL);

ALTER TABLE investment__decisions
    DROP COLUMN IF EXISTS decision_type,
    DROP COLUMN IF EXISTS process_type,
    DROP COLUMN IF EXISTS product_type,
    DROP COLUMN IF EXISTS strategy_code,
    DROP COLUMN IF EXISTS amendment_no;
