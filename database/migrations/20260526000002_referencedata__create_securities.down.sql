DROP INDEX IF EXISTS idx_market_symbols_security_id;
ALTER TABLE market_symbols DROP COLUMN IF EXISTS security_id;

DROP TABLE IF EXISTS reference_data_unmapped_security_candidates;
DROP TABLE IF EXISTS security_provider_mappings;
DROP TABLE IF EXISTS securities_master;
