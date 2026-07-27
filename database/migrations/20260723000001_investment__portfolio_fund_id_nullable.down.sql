-- Reverts 20260723000001_investment__portfolio_fund_id_nullable.up.sql.
-- Will fail if any fund-less portfolio rows exist — remediate (assign a fund
-- or delete the row) before rolling back.
ALTER TABLE investment__portfolios
    ALTER COLUMN fund_id SET NOT NULL;
