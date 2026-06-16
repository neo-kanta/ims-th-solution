-- Table: investment__portfolio_positions
-- Source: 20260428000004_investment__create_ledger.up.sql
CREATE TABLE investment__portfolio_positions (
    id                    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

    portfolio_id          UUID         NOT NULL REFERENCES investment__portfolios(id) ON DELETE RESTRICT,
    instrument_id         UUID         NOT NULL REFERENCES investment__instruments(id) ON DELETE RESTRICT,

    quantity              DECIMAL(28,8) NOT NULL DEFAULT 0,
    average_cost          DECIMAL(28,8) NOT NULL DEFAULT 0,
    cost_basis            DECIMAL(28,8) NOT NULL DEFAULT 0,

    last_transaction_id   UUID         REFERENCES investment__portfolio_transactions(id) ON DELETE RESTRICT,
    last_business_date    DATE,

    version               INTEGER      NOT NULL DEFAULT 1,
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_inv_position_portfolio_instrument
        UNIQUE (portfolio_id, instrument_id),

    CONSTRAINT chk_inv_position_quantity_non_negative
        CHECK (quantity >= 0),

    CONSTRAINT chk_inv_position_avg_cost_non_negative
        CHECK (average_cost >= 0),

    CONSTRAINT chk_inv_position_version_positive
        CHECK (version >= 1)
);
