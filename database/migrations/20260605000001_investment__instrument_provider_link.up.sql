-- =============================================================================
-- Investment Module — instrument ↔ market-data provider link
-- =============================================================================
-- Adds nullable provider-symbol columns to investment__instruments so the
-- intraday valuation service can resolve an internal instrument to the
-- ticker each provider expects without joining reference_data.
--
-- This is the simpler of two valid mapping stores: reference_data already has
-- securities_master + security_provider_mappings, but those rows are seeded
-- independently of investment__instruments. Keeping a small redundant pair on
-- the instrument row gives the intraday valuation a self-contained read path
-- while leaving reference_data intact for richer mapping workflows.
--
-- Columns are nullable; instruments without a mapping fall back to
-- "<primary_ticker>.BK" for Thai SET equities or the bare ticker otherwise.
-- =============================================================================

ALTER TABLE investment__instruments
    ADD COLUMN IF NOT EXISTS provider_symbol_alpha_vantage VARCHAR(64),
    ADD COLUMN IF NOT EXISTS provider_symbol_yahoo         VARCHAR(64);

COMMENT ON COLUMN investment__instruments.provider_symbol_alpha_vantage IS
    'Alpha Vantage ticker for live quote lookups. NULL = derive from primary_ticker.';
COMMENT ON COLUMN investment__instruments.provider_symbol_yahoo IS
    'Yahoo Finance ticker for live quote lookups. NULL = derive from primary_ticker.';

CREATE INDEX IF NOT EXISTS idx_inv_instruments_provider_yahoo
    ON investment__instruments (provider_symbol_yahoo)
    WHERE provider_symbol_yahoo IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_inv_instruments_provider_alpha
    ON investment__instruments (provider_symbol_alpha_vantage)
    WHERE provider_symbol_alpha_vantage IS NOT NULL AND deleted_at IS NULL;
