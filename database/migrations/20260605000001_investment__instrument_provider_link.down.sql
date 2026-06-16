-- Reverse of 20260605000001_investment__instrument_provider_link.up.sql.

DROP INDEX IF EXISTS idx_inv_instruments_provider_yahoo;
DROP INDEX IF EXISTS idx_inv_instruments_provider_alpha;

ALTER TABLE investment__instruments
    DROP COLUMN IF EXISTS provider_symbol_alpha_vantage,
    DROP COLUMN IF EXISTS provider_symbol_yahoo;
