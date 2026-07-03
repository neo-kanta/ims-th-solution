-- =============================================================================
-- Phase: 2026-06-15 — Investment Decision — Basket / Multi-Line Support
-- =============================================================================
-- Adds decision_type, process_type, product_type, strategy_code, amendment_no
-- to investment__decisions so the table can represent SINGLE_ORDER,
-- BASKET_ORDER, REBALANCE, and SWITCH decision headers.
--
-- Side CHECK is relaxed to include SWITCH_IN/OUT and REBALANCE_BUY/SELL.
-- instrument_code and side are made nullable because basket headers have no
-- single instrument or direction — those details live in decision_lines.
-- The quantity/amount check is removed; the application layer enforces it.
-- =============================================================================

-- Add classification and metadata columns
ALTER TABLE investment__decisions
    ADD COLUMN IF NOT EXISTS decision_type VARCHAR(20) NOT NULL DEFAULT 'SINGLE_ORDER',
    ADD COLUMN IF NOT EXISTS process_type  VARCHAR(20) NOT NULL DEFAULT 'INVESTMENT_DECISION',
    ADD COLUMN IF NOT EXISTS product_type  VARCHAR(20) NOT NULL DEFAULT 'MUTUAL_FUND',
    ADD COLUMN IF NOT EXISTS strategy_code VARCHAR(40),
    ADD COLUMN IF NOT EXISTS amendment_no  INTEGER     NOT NULL DEFAULT 0;

-- Allow basket headers to have no single instrument / side
ALTER TABLE investment__decisions
    ALTER COLUMN instrument_code DROP NOT NULL,
    ALTER COLUMN side            DROP NOT NULL;

-- Drop old strict side check; replace with a wider value set
ALTER TABLE investment__decisions
    DROP CONSTRAINT IF EXISTS chk_inv_decision_side;

ALTER TABLE investment__decisions
    ADD CONSTRAINT chk_inv_decision_side CHECK (
        side IS NULL
        OR side IN ('BUY','SELL','SWITCH','SWITCH_IN','SWITCH_OUT','REBALANCE_BUY','REBALANCE_SELL')
    );

-- Quantity/amount check no longer required at DB level; basket headers omit both
ALTER TABLE investment__decisions
    DROP CONSTRAINT IF EXISTS chk_inv_decision_quantity_or_amount;

-- Classification constraints
ALTER TABLE investment__decisions
    ADD CONSTRAINT chk_inv_decision_decision_type CHECK (
        decision_type IN ('SINGLE_ORDER','BASKET_ORDER','REBALANCE','SWITCH')
    );

ALTER TABLE investment__decisions
    ADD CONSTRAINT chk_inv_decision_process_type CHECK (
        process_type IN ('INVESTMENT_DECISION','ORDER_CANCEL','ORDER_AMEND')
    );

ALTER TABLE investment__decisions
    ADD CONSTRAINT chk_inv_decision_product_type CHECK (
        product_type IN ('MUTUAL_FUND','ETF','STOCK','BOND','CASH','MIXED')
    );

COMMENT ON COLUMN investment__decisions.decision_type  IS 'SINGLE_ORDER | BASKET_ORDER | REBALANCE | SWITCH';
COMMENT ON COLUMN investment__decisions.process_type   IS 'INVESTMENT_DECISION | ORDER_CANCEL | ORDER_AMEND';
COMMENT ON COLUMN investment__decisions.product_type   IS 'MUTUAL_FUND | ETF | STOCK | BOND | CASH | MIXED';
COMMENT ON COLUMN investment__decisions.strategy_code  IS 'Optional reference to a research/strategy allocation.';
COMMENT ON COLUMN investment__decisions.amendment_no   IS 'Incremented each time an approved decision is amended.';
