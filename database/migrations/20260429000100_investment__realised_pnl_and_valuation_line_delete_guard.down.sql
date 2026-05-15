DROP RULE IF EXISTS no_delete_inv_valuation_lines ON investment__valuation_holding_lines;

ALTER TABLE investment__portfolio_transactions
    DROP COLUMN IF EXISTS realised_pnl_base;
