ALTER TABLE market_data_snapshots
    DROP CONSTRAINT IF EXISTS chk_market_data_snapshots_provider,
    DROP CONSTRAINT IF EXISTS chk_market_data_snapshots_provider_name,
    ADD CONSTRAINT chk_market_data_snapshots_provider_name
        CHECK (provider_name <> '');

ALTER TABLE provider_requests_log
    DROP CONSTRAINT IF EXISTS chk_provider_requests_log_provider,
    DROP CONSTRAINT IF EXISTS chk_provider_requests_log_provider_name,
    ADD CONSTRAINT chk_provider_requests_log_provider_name
        CHECK (provider_name <> '');
