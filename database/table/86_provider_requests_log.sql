-- Table: provider_requests_log
-- Source: 20260429000200_marketdata__create_tables.up.sql
CREATE TABLE provider_requests_log (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_name   VARCHAR(40)  NOT NULL,
    symbol          VARCHAR(64)  NOT NULL,
    operation       VARCHAR(20)  NOT NULL,
    status          VARCHAR(20)  NOT NULL,
    status_code     INTEGER,
    error_code      VARCHAR(80),
    error_message   TEXT,
    duration_ms     INTEGER      NOT NULL DEFAULT 0,
    requested_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_provider_requests_log_provider_name
        CHECK (provider_name <> ''),
    CONSTRAINT chk_provider_requests_log_operation
        CHECK (operation IN ('quote','history')),
    CONSTRAINT chk_provider_requests_log_status
        CHECK (status IN ('success','error'))
);
