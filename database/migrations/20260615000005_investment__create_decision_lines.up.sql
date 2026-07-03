-- =============================================================================
-- Phase: 2026-06-15 — Investment Decision Lines
-- =============================================================================
-- Stores the individual instrument lines for BASKET_ORDER, REBALANCE, and
-- SWITCH decision types. SINGLE_ORDER decisions keep their instrument/side/qty
-- on the parent investment__decisions row and leave this table empty.
-- =============================================================================

CREATE TABLE IF NOT EXISTS investment__decision_lines (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_id     UUID            NOT NULL REFERENCES investment__decisions(id) ON DELETE CASCADE,
    line_number     INTEGER         NOT NULL,

    instrument_id   UUID            REFERENCES investment__instruments(id) ON DELETE RESTRICT,
    instrument_code VARCHAR(40)     NOT NULL,
    product_type    VARCHAR(20)     NOT NULL DEFAULT 'MUTUAL_FUND',

    side            VARCHAR(20)     NOT NULL,
    quantity        DECIMAL(28,8),
    amount          DECIMAL(28,8),
    target_weight   DECIMAL(10,6),
    limit_price     DECIMAL(28,8),
    currency        CHAR(3)         NOT NULL,

    notes           TEXT,

    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_by      UUID            NOT NULL REFERENCES iam_users(id),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by      UUID            NOT NULL REFERENCES iam_users(id),

    CONSTRAINT uq_inv_decision_line_number    UNIQUE (decision_id, line_number),
    CONSTRAINT chk_inv_line_side              CHECK (side IN ('BUY','SELL','SWITCH','SWITCH_IN','SWITCH_OUT','REBALANCE_BUY','REBALANCE_SELL')),
    CONSTRAINT chk_inv_line_currency          CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_inv_line_product_type      CHECK (product_type IN ('MUTUAL_FUND','ETF','STOCK','BOND','CASH'))
);

CREATE INDEX IF NOT EXISTS idx_inv_decision_line_decision
    ON investment__decision_lines (decision_id);

COMMENT ON TABLE investment__decision_lines IS
    'Per-instrument lines for BASKET_ORDER, REBALANCE, and SWITCH decisions. Empty for SINGLE_ORDER.';
COMMENT ON COLUMN investment__decision_lines.target_weight IS
    'Target portfolio weight (0–1) for REBALANCE lines.';
