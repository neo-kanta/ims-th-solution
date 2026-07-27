-- Reverts 20260723000002_investment__trading_fund_id_nullable.up.sql.
-- Will fail if any fund-less transaction/decision/execution/confirmation
-- rows exist — remediate before rolling back.
ALTER TABLE investment__trade_confirmations ALTER COLUMN fund_id SET NOT NULL;
ALTER TABLE investment__executions ALTER COLUMN fund_id SET NOT NULL;
ALTER TABLE investment__decisions ALTER COLUMN fund_id SET NOT NULL;
ALTER TABLE investment__portfolio_transactions ALTER COLUMN fund_id SET NOT NULL;
