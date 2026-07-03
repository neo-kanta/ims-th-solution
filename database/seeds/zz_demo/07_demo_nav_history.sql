-- =============================================================================
-- Demo seed — 60-day NAV / valuation history
-- =============================================================================
-- Back-fills daily valuation_snapshots and (for unitised funds) nav_snapshots
-- for the previous 60 business days so the "Unit NAV history" chart in the
-- fund workspace renders a meaningful trend.
--
-- AUM / NAV-per-unit trend uses a deterministic mild-sine trajectory around
-- today's seeded values:
--     factor(t) = 1 + 0.005 * sin(t / 15) - 0.0002 * t
-- That keeps the latest values close to the anchor row from 03_demo_prices_
-- valuations.sql and drifts ~1.5% over the 60-day window — visible on a
-- chart, not so dramatic it looks fabricated.
--
-- Idempotent: the natural unique indexes on (portfolio_id, business_date,
-- source) and (portfolio_id, business_date) make re-runs no-ops.
--
-- GLOBAL-TECH's current valuation is dated CURRENT_DATE - 3 (stale demo),
-- so its history window is shifted by 3 days so we don't write rows that
-- would appear "fresher" than the canonical stale anchor.
-- =============================================================================

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Non-stale portfolios: 60 business days back from CURRENT_DATE - 1
-- ---------------------------------------------------------------------------
INSERT INTO investment__valuation_snapshots (
    id, portfolio_id, business_date, valuation_ccy,
    market_value, cost_basis, unrealised_pnl, realised_pnl, roi, aum,
    cash_balance, price_set_hash, has_stale_inputs, is_indicative, source,
    created_by
)
SELECT
    gen_random_uuid(),
    p.portfolio_id::uuid,
    (CURRENT_DATE - day_offset)::date,
    p.valuation_ccy,
    (p.market_value * factor)::decimal(28,8),
    p.cost_basis::decimal(28,8),
    (p.market_value * factor - p.cost_basis)::decimal(28,8),
    0::decimal(28,8),
    NULL,
    (p.aum * factor)::decimal(28,8),
    p.cash_balance::decimal(28,8),
    'demo-seed-hist-' || day_offset::text,
    false,
    true,
    'INTERNAL',
    'a0000000-0000-0000-0000-000000000001'::uuid
FROM (VALUES
    ('d0002000-0000-0000-0000-000000000001', 'THB', 2511600000.00, 2500400000.00, 2571840265.00,  60240265.00),
    ('d0002000-0000-0000-0000-000000000002', 'THB', 1210565000.00, 1182950000.00, 1284500000.00,  73935000.00),
    ('d0002000-0000-0000-0000-000000000003', 'THB',  500400000.00,  498400000.00,  541720400.00,  41320400.00),
    ('d0002000-0000-0000-0000-000000000004', 'THB',  511160000.00,  494200000.00,  545677378.00,  34517378.00),
    ('d0002000-0000-0000-0000-000000000005', 'THB',  404085600.00,  401421600.00,  440533427.00,  36447827.00),
    ('d0002000-0000-0000-0000-000000000007', 'THB',  809600000.00,  808000000.00,  834600000.00,  25000000.00)
) AS p(portfolio_id, valuation_ccy, market_value, cost_basis, aum, cash_balance)
CROSS JOIN generate_series(1, 60) AS day_offset
CROSS JOIN LATERAL (
    SELECT (1.0 + 0.005 * sin(day_offset / 15.0) - 0.0002 * day_offset)::numeric AS factor
) AS f
ON CONFLICT (portfolio_id, business_date, source) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. GLOBAL-TECH (USD, stale by 3 days): shift window so we never appear
--    fresher than the canonical stale anchor row.
-- ---------------------------------------------------------------------------
INSERT INTO investment__valuation_snapshots (
    id, portfolio_id, business_date, valuation_ccy,
    market_value, cost_basis, unrealised_pnl, realised_pnl, roi, aum,
    cash_balance, price_set_hash, has_stale_inputs, is_indicative, source,
    created_by
)
SELECT
    gen_random_uuid(),
    'd0002000-0000-0000-0000-000000000006'::uuid,
    (CURRENT_DATE - day_offset)::date,
    'USD',
    (106381000.00 * factor)::decimal(28,8),
    104400000.00::decimal(28,8),
    (106381000.00 * factor - 104400000.00)::decimal(28,8),
    0::decimal(28,8),
    NULL,
    (109881000.00 * factor)::decimal(28,8),
    3500000.00::decimal(28,8),
    'demo-seed-hist-gt-' || day_offset::text,
    true,                          -- stale = true throughout the historic window
    true,
    'INTERNAL',
    'a0000000-0000-0000-0000-000000000001'::uuid
FROM generate_series(4, 63) AS day_offset
CROSS JOIN LATERAL (
    SELECT (1.0 + 0.005 * sin(day_offset / 15.0) - 0.0002 * day_offset)::numeric AS factor
) AS f
ON CONFLICT (portfolio_id, business_date, source) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 3. NAV-per-unit history for the two unitised portfolios.
--    A02-CORE (TH-GOV-LTF): 250,000 units; nav/unit anchor 10,287.36106
--    MM-CORE  (MMF-CASH):   80,000,000 units; nav/unit anchor 10.43250
-- ---------------------------------------------------------------------------
INSERT INTO investment__nav_snapshots (
    id, portfolio_id, business_date, total_units, nav_per_unit,
    valuation_snapshot_id, is_indicative, created_by
)
SELECT
    gen_random_uuid(),
    p.portfolio_id::uuid,
    vs.business_date,
    p.total_units::decimal(28,8),
    (p.nav_anchor * (vs.aum / p.aum_anchor))::decimal(28,12),
    vs.id,
    true,
    'a0000000-0000-0000-0000-000000000001'::uuid
FROM (VALUES
    ('d0002000-0000-0000-0000-000000000001',    250000.00, 10287.36106, 2571840265.00),
    ('d0002000-0000-0000-0000-000000000007',  80000000.00,    10.43250,  834600000.00)
) AS p(portfolio_id, total_units, nav_anchor, aum_anchor)
JOIN investment__valuation_snapshots vs
    ON vs.portfolio_id = p.portfolio_id::uuid
   AND vs.source       = 'INTERNAL'
   AND vs.business_date BETWEEN CURRENT_DATE - 60 AND CURRENT_DATE - 1
ON CONFLICT (portfolio_id, business_date) DO NOTHING;

COMMIT;
