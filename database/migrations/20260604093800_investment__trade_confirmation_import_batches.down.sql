-- Reverse the trade-confirmation import batch tables.
-- Drop the FK first so the batch table can be dropped.

ALTER TABLE investment__trade_confirmations
    DROP CONSTRAINT IF EXISTS fk_inv_confirmation_import_batch;

DROP TABLE IF EXISTS investment__trade_confirmation_import_items;
DROP TABLE IF EXISTS investment__trade_confirmation_import_batches;
