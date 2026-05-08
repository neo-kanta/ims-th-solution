ALTER TABLE investment__portfolio_transactions
    ADD COLUMN IF NOT EXISTS realised_pnl_base DECIMAL(28,8) NOT NULL DEFAULT 0;

DROP RULE IF EXISTS no_delete_inv_valuation_lines ON investment__valuation_holding_lines;

CREATE RULE no_delete_inv_valuation_lines
    AS ON DELETE TO investment__valuation_holding_lines DO INSTEAD NOTHING;
