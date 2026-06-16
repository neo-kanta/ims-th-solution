-- =============================================================================
-- Demo seed — Price snapshots, valuation snapshots, NAV, and AUM
-- =============================================================================
-- These tables are append-only at the DB layer (RULE no_update / no_delete).
--
-- Idempotency strategy (re-runnable on ANY calendar day):
--   * Row ids use gen_random_uuid() rather than fixed UUIDs. Fixed ids paired
--     with a CURRENT_DATE-relative business_date are NOT safe across days — on a
--     later day the natural key (which includes business_date) no longer matches
--     the ON CONFLICT target, so the fixed primary key collides with the row
--     written on the first run (SQLSTATE 23505 on *_pkey).
--   * Conflicts resolve on the NATURAL unique key (DO NOTHING), so re-running on
--     the same day is a no-op and re-running on a new day appends that day's
--     snapshot — the correct behaviour for append-only price/valuation history.
--   * The NAV snapshot resolves its valuation_snapshot_id by natural key rather
--     than a literal id, so it always points at the matching valuation row.
--
-- Business-date strategy:
--   * Most price/valuation rows use CURRENT_DATE.
--   * GLOBAL-TECH uses CURRENT_DATE - 3 to demonstrate the "Stale NAV" badge.
-- =============================================================================

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Price snapshots — one per instrument, source = 'DEMO_SEED'
-- ---------------------------------------------------------------------------
INSERT INTO investment__price_snapshots (
    id, instrument_id, business_date, price, currency, price_source,
    provider_ref, is_stale, stale_reason, captured_at, created_by
)
SELECT gen_random_uuid(), v.instrument_id::uuid,
       (CURRENT_DATE - v.date_offset::int)::date,
       v.price::decimal(28,8), v.currency,
       'DEMO_SEED', 'demo-seed-snapshot', v.is_stale,
       CASE WHEN v.is_stale THEN 'Demo: provider feed offline >24h' ELSE NULL END,
       NOW() - (v.date_offset::int || ' days')::interval,
       'a0000000-0000-0000-0000-000000000001'::uuid
FROM (VALUES
    -- Thai equity — current
    ('d0003000-0000-0000-0000-000000000001', 0,   34.50, 'THB', false),
    ('d0003000-0000-0000-0000-000000000002', 0,  158.80, 'THB', false),
    ('d0003000-0000-0000-0000-000000000003', 0,   63.50, 'THB', false),
    ('d0003000-0000-0000-0000-000000000004', 0,   59.40, 'THB', false),
    ('d0003000-0000-0000-0000-000000000005', 0,  110.50, 'THB', false),

    -- US tech equity — stale by 3 days for the GLOBAL-TECH demo
    ('d0003000-0000-0000-0000-000000000006', 3,  183.50, 'USD', true),
    ('d0003000-0000-0000-0000-000000000007', 3,  415.00, 'USD', true),
    ('d0003000-0000-0000-0000-000000000008', 3,  128.50, 'USD', true),
    ('d0003000-0000-0000-0000-000000000009', 3,  168.20, 'USD', true),

    -- Thai government bonds
    ('d0003000-0000-0000-0000-00000000000a', 0, 1032.00, 'THB', false),
    ('d0003000-0000-0000-0000-00000000000b', 0, 1308.00, 'THB', false),
    ('d0003000-0000-0000-0000-00000000000c', 0,   99.85, 'USD', false),

    -- Thai corporate bonds
    ('d0003000-0000-0000-0000-00000000000d', 0, 1002.00, 'THB', false),
    ('d0003000-0000-0000-0000-00000000000e', 0,  999.00, 'THB', false),

    -- MMF
    ('d0003000-0000-0000-0000-00000000000f', 0,   10.12, 'THB', false)
) AS v(instrument_id, date_offset, price, currency, is_stale)
ON CONFLICT (instrument_id, business_date, price_source) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Valuation snapshots — one per portfolio for CURRENT_DATE
--    GLOBAL-TECH uses CURRENT_DATE - 3 to expose stale-NAV semantics.
-- ---------------------------------------------------------------------------
INSERT INTO investment__valuation_snapshots (
    id, portfolio_id, business_date, valuation_ccy,
    market_value, cost_basis, unrealised_pnl, realised_pnl, roi, aum,
    cash_balance, price_set_hash, has_stale_inputs, is_indicative, source,
    created_by
)
SELECT gen_random_uuid(), v.portfolio_id::uuid,
       (CURRENT_DATE - v.date_offset::int)::date, v.valuation_ccy,
       v.market_value::decimal(28,8), v.cost_basis::decimal(28,8),
       v.unrealised_pnl::decimal(28,8), 0::decimal(28,8),
       v.roi::decimal(18,8), v.aum::decimal(28,8),
       v.cash_balance::decimal(28,8),
       'demo-seed-' || substr(v.portfolio_id, 1, 12),
       v.has_stale_inputs, true, 'INTERNAL',
       'a0000000-0000-0000-0000-000000000001'::uuid
FROM (VALUES
    -- A02-CORE (TH-GOV-LTF): bonds 1,857,600,000 + 654,000,000 = 2,511,600,000
    -- cost 2,500,400,000; unrealised 11,200,000; cash 60,240,265; AUM 2,571,840,265
    ('d0002000-0000-0000-0000-000000000001', 0, 'THB',
        2511600000.00, 2500400000.00,   11200000.00, 0.00448, 2571840265.00,   60240265.00, false),

    -- B14-CORE (BBL-EQUITY): five equities 1,210,565,000; cost 1,182,950,000
    -- unrealised 27,615,000; cash 73,935,000; AUM 1,284,500,000
    ('d0002000-0000-0000-0000-000000000002', 0, 'THB',
        1210565000.00, 1182950000.00,   27615000.00, 0.02335, 1284500000.00,   73935000.00, false),

    -- C07-CORE (SCB-FIXED): bonds 500,400,000; cost 498,400,000; unreal 2,000,000
    -- cash 41,320,400; AUM 541,720,400
    ('d0002000-0000-0000-0000-000000000003', 0, 'THB',
         500400000.00,  498400000.00,    2000000.00, 0.00401,  541720400.00,   41320400.00, false),

    -- D03-EQUITY (KTB-BALANCED equity sleeve): 511,160,000; cost 494,200,000
    -- unreal 16,960,000; cash 34,517,378; AUM 545,677,378
    ('d0002000-0000-0000-0000-000000000004', 0, 'THB',
         511160000.00,  494200000.00,   16960000.00, 0.03432,  545677378.00,   34517378.00, false),

    -- D03-BOND (KTB-BALANCED bond sleeve): 404,085,600; cost 401,421,600
    -- unreal 2,664,000; cash 36,447,827; AUM 440,533,427
    ('d0002000-0000-0000-0000-000000000005', 0, 'THB',
         404085600.00,  401421600.00,    2664000.00, 0.00664,  440533427.00,   36447827.00, false),

    -- GT-CORE (GLOBAL-TECH, USD): 4 names 106,381,000; cost 104,400,000
    -- unreal 1,981,000; cash 3,500,000; AUM 109,881,000 — STALE
    ('d0002000-0000-0000-0000-000000000006', 3, 'USD',
         106381000.00,  104400000.00,    1981000.00, 0.01897,  109881000.00,    3500000.00, true),

    -- MM-CORE (MMF-CASH): K-CASH 809,600,000; cost 808,000,000; unreal 1,600,000
    -- cash 25,000,000; AUM 834,600,000
    ('d0002000-0000-0000-0000-000000000007', 0, 'THB',
         809600000.00,  808000000.00,    1600000.00, 0.00198,  834600000.00,   25000000.00, false)
) AS v(portfolio_id, date_offset, valuation_ccy, market_value, cost_basis,
       unrealised_pnl, roi, aum, cash_balance, has_stale_inputs)
ON CONFLICT (portfolio_id, business_date, source) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 3. NAV snapshots — only for has_units = true portfolios
--      A02-CORE (TH-GOV-LTF): 250,000 units @ 10,287.36106 per unit
--      MM-CORE  (MMF-CASH):   80,000,000 units @ 10.4325 per unit
--    valuation_snapshot_id is resolved by natural key (portfolio + today's
--    business_date + INTERNAL source) so it survives the move to random ids.
-- ---------------------------------------------------------------------------
INSERT INTO investment__nav_snapshots (
    id, portfolio_id, business_date, total_units, nav_per_unit,
    valuation_snapshot_id, is_indicative, created_by
)
SELECT gen_random_uuid(), v.portfolio_id::uuid, CURRENT_DATE,
       v.total_units::decimal(28,8), v.nav_per_unit::decimal(28,12),
       (SELECT vs.id
          FROM investment__valuation_snapshots vs
         WHERE vs.portfolio_id = v.portfolio_id::uuid
           AND vs.business_date = CURRENT_DATE
           AND vs.source = 'INTERNAL'
         ORDER BY vs.created_at DESC
         LIMIT 1),
       true,
       'a0000000-0000-0000-0000-000000000001'::uuid
FROM (VALUES
    ('d0002000-0000-0000-0000-000000000001', 250000,    10287.361060000000),
    ('d0002000-0000-0000-0000-000000000007', 80000000,     10.432500000000)
) AS v(portfolio_id, total_units, nav_per_unit)
ON CONFLICT (portfolio_id, business_date) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 4. AUM snapshots — per portfolio AND per fund (rolled up across portfolios)
-- ---------------------------------------------------------------------------
INSERT INTO investment__aum_snapshots (
    id, scope_type, scope_id, business_date, aum, valuation_ccy, source, created_by
)
SELECT gen_random_uuid(), v.scope_type, v.scope_id::uuid,
       (CURRENT_DATE - v.date_offset::int)::date,
       v.aum::decimal(28,8), v.valuation_ccy, 'INTERNAL',
       'a0000000-0000-0000-0000-000000000001'::uuid
FROM (VALUES
    -- Per-portfolio AUM (mirrors valuation_snapshots.aum)
    ('PORTFOLIO', 'd0002000-0000-0000-0000-000000000001', 0, 'THB', 2571840265.00),
    ('PORTFOLIO', 'd0002000-0000-0000-0000-000000000002', 0, 'THB', 1284500000.00),
    ('PORTFOLIO', 'd0002000-0000-0000-0000-000000000003', 0, 'THB',  541720400.00),
    ('PORTFOLIO', 'd0002000-0000-0000-0000-000000000004', 0, 'THB',  545677378.00),
    ('PORTFOLIO', 'd0002000-0000-0000-0000-000000000005', 0, 'THB',  440533427.00),
    ('PORTFOLIO', 'd0002000-0000-0000-0000-000000000006', 3, 'USD',  109881000.00),
    ('PORTFOLIO', 'd0002000-0000-0000-0000-000000000007', 0, 'THB',  834600000.00),

    -- Per-fund AUM (rolled up; KTB-BALANCED sums its two sleeves)
    ('FUND',      'd0001000-0000-0000-0000-000000000001', 0, 'THB', 2571840265.00),
    ('FUND',      'd0001000-0000-0000-0000-000000000002', 0, 'THB', 1284500000.00),
    ('FUND',      'd0001000-0000-0000-0000-000000000003', 0, 'THB',  541720400.00),
    ('FUND',      'd0001000-0000-0000-0000-000000000004', 0, 'THB',  986210805.00),
    ('FUND',      'd0001000-0000-0000-0000-000000000005', 3, 'USD',  109881000.00),
    ('FUND',      'd0001000-0000-0000-0000-000000000006', 0, 'THB',  834600000.00)
) AS v(scope_type, scope_id, date_offset, valuation_ccy, aum)
ON CONFLICT (scope_type, scope_id, business_date, source) DO NOTHING;

COMMIT;
