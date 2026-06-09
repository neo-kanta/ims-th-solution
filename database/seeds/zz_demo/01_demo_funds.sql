-- =============================================================================
-- Demo seed — Funds, Portfolios, Instruments
-- =============================================================================
-- Brings the "My Funds" cockpit to life with six representative Thai PoC funds
-- spanning equity, fixed income, mixed, money-market, and a USD thematic fund.
-- Idempotent: every row uses a stable UUID and is upserted by primary key
-- (or, for projection rows, by their natural unique constraint).
--
-- Conventions used in this file:
--   * UUIDs are layered by entity kind:
--       d0001000-...  funds
--       d0002000-...  portfolios
--       d0003000-...  instruments
--   * Manager assignments distribute "Manager" and "Member" badges across the
--     admin user (a000...0001), ben (...0010), green (...0011), neo (...0012).
--   * All currencies are ISO 4217 uppercase 3-letter; the schema CHECK
--     enforces this so seed errors surface immediately.
-- =============================================================================

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Demo funds (six funds — one breached, one stale, one locked, etc.)
-- ---------------------------------------------------------------------------
INSERT INTO investment__funds (
    id, code, name, short_name, fund_category_id, base_currency,
    inception_date, manager_user_id, benchmark, risk_profile, has_units,
    status, created_by, updated_by
)
SELECT v.id::uuid, v.code, v.name, v.short_name, fc.id, v.base_currency,
       v.inception_date::date, v.manager_user_id::uuid, v.benchmark,
       v.risk_profile, v.has_units, v.status,
       'a0000000-0000-0000-0000-000000000001'::uuid,
       'a0000000-0000-0000-0000-000000000001'::uuid
FROM (VALUES
    ('d0001000-0000-0000-0000-000000000001', 'TH-GOV-LTF',  'Thai Government LTF — Alpha series',     'fund-alpha',
        'BOND_FUND',    'THB', '2024-01-15', 'a0000000-0000-0000-0000-000000000001', 'ThaiBMA Govt Bond Index', 'LOW',    true,  'ACTIVE'),
    ('d0001000-0000-0000-0000-000000000002', 'BBL-EQUITY',  'Large-cap Momentum — Equity',            'lp-momentum',
        'EQUITY_FUND',  'THB', '2023-06-01', 'a0000000-0000-0000-0000-000000000001', 'SET50',                   'MEDIUM', false, 'ACTIVE'),
    ('d0001000-0000-0000-0000-000000000003', 'SCB-FIXED',   'Corporate Bonds 2026 — Fixed Income',    'corp-bonds-2026',
        'BOND_FUND',    'THB', '2022-09-12', 'a0000000-0000-0000-0000-000000000010', 'ThaiBMA Corp BBB+',       'LOW',    false, 'ACTIVE'),
    ('d0001000-0000-0000-0000-000000000004', 'KTB-BALANCED','Balanced — Quarterly Rebalance Q2',      'rebalance-q2',
        'MIXED_FUND',   'THB', '2024-04-01', 'a0000000-0000-0000-0000-000000000001', 'Balanced 50/40/10',       'MEDIUM', false, 'ACTIVE'),
    ('d0001000-0000-0000-0000-000000000005', 'GLOBAL-TECH', 'Global Tech Thematic',                   'global-tech',
        'THEMATIC_FUND','USD', '2025-02-20', 'a0000000-0000-0000-0000-000000000011', 'NASDAQ-100',              'HIGH',   false, 'ACTIVE'),
    ('d0001000-0000-0000-0000-000000000006', 'MMF-CASH',    'SCB Money Market Fund',                  'money-market',
        'MMF_FUND',     'THB', '2024-10-10', 'a0000000-0000-0000-0000-000000000001', '1M BIBOR',                'LOW',    true,  'ACTIVE')
) AS v(id, code, name, short_name, fund_category_code, base_currency, inception_date, manager_user_id, benchmark, risk_profile, has_units, status)
JOIN investment__fund_categories fc ON fc.code = v.fund_category_code
ON CONFLICT (id) DO UPDATE SET
    code             = EXCLUDED.code,
    name             = EXCLUDED.name,
    short_name       = EXCLUDED.short_name,
    fund_category_id = EXCLUDED.fund_category_id,
    base_currency    = EXCLUDED.base_currency,
    manager_user_id  = EXCLUDED.manager_user_id,
    benchmark        = EXCLUDED.benchmark,
    risk_profile     = EXCLUDED.risk_profile,
    has_units        = EXCLUDED.has_units,
    status           = EXCLUDED.status,
    updated_at       = NOW(),
    updated_by       = EXCLUDED.updated_by;

-- ---------------------------------------------------------------------------
-- 2. Demo portfolios (7 portfolios; KTB-BALANCED has 2 sleeves)
-- ---------------------------------------------------------------------------
INSERT INTO investment__portfolios (
    id, fund_id, code, name, description, base_currency, valuation_currency,
    strategy_code, style_id, manager_user_id, benchmark, risk_profile,
    inception_date, status, has_units, tax_lot_method,
    created_by, updated_by
)
SELECT v.id::uuid, v.fund_id::uuid, v.code, v.name, v.description,
       v.base_currency, v.valuation_currency, v.strategy_code,
       (SELECT id FROM investment__investment_styles WHERE code = v.style_code),
       v.manager_user_id::uuid, v.benchmark, v.risk_profile,
       v.inception_date::date, v.status, v.has_units, v.tax_lot_method,
       'a0000000-0000-0000-0000-000000000001'::uuid,
       'a0000000-0000-0000-0000-000000000001'::uuid
FROM (VALUES
    ('d0002000-0000-0000-0000-000000000001', 'd0001000-0000-0000-0000-000000000001', 'A02-CORE',
        'Alpha Core Bond Sleeve', 'Long-duration Thai government bonds.', 'THB', 'THB',
        'CORE',       'INCOME',  'a0000000-0000-0000-0000-000000000001', 'ThaiBMA Govt Bond Index', 'LOW',
        '2024-01-15', 'ACTIVE', true,  'AVERAGE'),
    ('d0002000-0000-0000-0000-000000000002', 'd0001000-0000-0000-0000-000000000002', 'B14-CORE',
        'Momentum Equity Core',   'Large-cap Thai equity with momentum tilt.', 'THB', 'THB',
        'MOMENTUM',   'GROWTH',  'a0000000-0000-0000-0000-000000000001', 'SET50',                   'MEDIUM',
        '2023-06-01', 'ACTIVE', false, 'AVERAGE'),
    ('d0002000-0000-0000-0000-000000000003', 'd0001000-0000-0000-0000-000000000003', 'C07-CORE',
        'Corporate Bond Core',    'Investment-grade Thai corporate bonds.', 'THB', 'THB',
        'CORE',       'INCOME',  'a0000000-0000-0000-0000-000000000010', 'ThaiBMA Corp BBB+',       'LOW',
        '2022-09-12', 'ACTIVE', false, 'AVERAGE'),
    ('d0002000-0000-0000-0000-000000000004', 'd0001000-0000-0000-0000-000000000004', 'D03-EQUITY',
        'Balanced Equity Sleeve', 'Thai blue-chip equity sleeve.', 'THB', 'THB',
        'BALANCED',   'BLEND',   'a0000000-0000-0000-0000-000000000001', 'SET50',                   'MEDIUM',
        '2024-04-01', 'ACTIVE', false, 'AVERAGE'),
    ('d0002000-0000-0000-0000-000000000005', 'd0001000-0000-0000-0000-000000000004', 'D03-BOND',
        'Balanced Bond Sleeve',   'Mix of Thai government and corporate bonds.', 'THB', 'THB',
        'BALANCED',   'INCOME',  'a0000000-0000-0000-0000-000000000001', 'ThaiBMA Composite',       'LOW',
        '2024-04-01', 'ACTIVE', false, 'AVERAGE'),
    ('d0002000-0000-0000-0000-000000000006', 'd0001000-0000-0000-0000-000000000005', 'GT-CORE',
        'Global Tech Core',       'US-listed large-cap tech equity.', 'USD', 'USD',
        'THEMATIC',   'GROWTH',  'a0000000-0000-0000-0000-000000000011', 'NASDAQ-100',              'HIGH',
        '2025-02-20', 'ACTIVE', false, 'AVERAGE'),
    ('d0002000-0000-0000-0000-000000000007', 'd0001000-0000-0000-0000-000000000006', 'MM-CORE',
        'Money Market Core',      'Thai short-duration money market.', 'THB', 'THB',
        'CASH_EQ',    'INCOME',  'a0000000-0000-0000-0000-000000000001', '1M BIBOR',                'LOW',
        '2024-10-10', 'ACTIVE', true,  'AVERAGE')
) AS v(id, fund_id, code, name, description, base_currency, valuation_currency,
       strategy_code, style_code, manager_user_id, benchmark, risk_profile,
       inception_date, status, has_units, tax_lot_method)
ON CONFLICT (id) DO UPDATE SET
    fund_id            = EXCLUDED.fund_id,
    code               = EXCLUDED.code,
    name               = EXCLUDED.name,
    description        = EXCLUDED.description,
    base_currency      = EXCLUDED.base_currency,
    valuation_currency = EXCLUDED.valuation_currency,
    strategy_code      = EXCLUDED.strategy_code,
    style_id           = EXCLUDED.style_id,
    manager_user_id    = EXCLUDED.manager_user_id,
    benchmark          = EXCLUDED.benchmark,
    risk_profile       = EXCLUDED.risk_profile,
    status             = EXCLUDED.status,
    has_units          = EXCLUDED.has_units,
    tax_lot_method     = EXCLUDED.tax_lot_method,
    updated_at         = NOW(),
    updated_by         = EXCLUDED.updated_by;

-- ---------------------------------------------------------------------------
-- 3. Demo instruments — Thai equity, US tech, gov/corp bonds, MMF
-- ---------------------------------------------------------------------------
--
-- Notes:
--   * primary_exchange is set so the alive-uniqueness index doesn't collide.
--   * Sector lookups use level-1 codes from the reference seed; the sector
--     code differs slightly per existing reference SQL (e.g. CONS_DISC vs
--     CONSUMER_DISCRETIONARY) — we go with the 006_investment_reference_seed
--     short codes here because that is what runs first.
INSERT INTO investment__instruments (
    id, primary_ticker, name, asset_class_id, asset_subtype_id, currency,
    country_id, region_id, primary_exchange, sector_id, fund_category_id,
    lot_size, tick_size, is_tradable, status, attributes, created_by, updated_by
)
SELECT v.id::uuid, v.primary_ticker, v.name, ac.id, ast.id, v.currency,
       co.id, co.region_id, v.primary_exchange,
       (SELECT s.id FROM investment__sectors s WHERE s.code = v.sector_code AND s.level = 1 LIMIT 1),
       NULL,
       v.lot_size, v.tick_size::decimal(20,8), true, 'ACTIVE', v.attributes::jsonb,
       'a0000000-0000-0000-0000-000000000001'::uuid,
       'a0000000-0000-0000-0000-000000000001'::uuid
FROM (VALUES
    -- Thai blue-chip equities (SET)
    ('d0003000-0000-0000-0000-000000000001', 'PTT',     'PTT Public Company Ltd',         'EQUITY',       'COMMON_STOCK', 'THB', 'TH', 'SET',    'ENERGY',      100, '0.25', '{"asset_kind":"equity"}'),
    ('d0003000-0000-0000-0000-000000000002', 'KBANK',   'Kasikornbank PCL',               'EQUITY',       'COMMON_STOCK', 'THB', 'TH', 'SET',    'FINANCIALS',  100, '0.50', '{"asset_kind":"equity","issuer":"KBANK"}'),
    ('d0003000-0000-0000-0000-000000000003', 'AOT',     'Airports of Thailand PCL',       'EQUITY',       'COMMON_STOCK', 'THB', 'TH', 'SET',    'INDUSTRIALS', 100, '0.10', '{"asset_kind":"equity"}'),
    ('d0003000-0000-0000-0000-000000000004', 'CPALL',   'CP All PCL',                     'EQUITY',       'COMMON_STOCK', 'THB', 'TH', 'SET',    'CONS_STAPLES',100, '0.25', '{"asset_kind":"equity"}'),
    ('d0003000-0000-0000-0000-000000000005', 'SCB',     'SCB X PCL',                      'EQUITY',       'COMMON_STOCK', 'THB', 'TH', 'SET',    'FINANCIALS',  100, '0.25', '{"asset_kind":"equity"}'),

    -- US tech equities
    ('d0003000-0000-0000-0000-000000000006', 'AAPL',    'Apple Inc.',                     'EQUITY',       'COMMON_STOCK', 'USD', 'US', 'NASDAQ', 'INFO_TECH',     1, '0.01', '{"asset_kind":"equity"}'),
    ('d0003000-0000-0000-0000-000000000007', 'MSFT',    'Microsoft Corp.',                'EQUITY',       'COMMON_STOCK', 'USD', 'US', 'NASDAQ', 'INFO_TECH',     1, '0.01', '{"asset_kind":"equity"}'),
    ('d0003000-0000-0000-0000-000000000008', 'NVDA',    'NVIDIA Corp.',                   'EQUITY',       'COMMON_STOCK', 'USD', 'US', 'NASDAQ', 'INFO_TECH',     1, '0.01', '{"asset_kind":"equity"}'),
    ('d0003000-0000-0000-0000-000000000009', 'GOOGL',   'Alphabet Inc. Class A',          'EQUITY',       'COMMON_STOCK', 'USD', 'US', 'NASDAQ', 'COMM_SVCS',     1, '0.01', '{"asset_kind":"equity"}'),

    -- Thai government bonds
    ('d0003000-0000-0000-0000-00000000000a', 'TH-LB30DA','Thai Govt Bond LB30DA (10Y)',   'FIXED_INCOME', 'GOV_BOND',    'THB', 'TH', 'TBMA',  NULL, 1, '0.01', '{"asset_kind":"bond","coupon_pct":2.85,"maturity":"2034-12-15"}'),
    ('d0003000-0000-0000-0000-00000000000b', 'TH-LB40DA','Thai Govt Bond LB40DA (20Y)',   'FIXED_INCOME', 'GOV_BOND',    'THB', 'TH', 'TBMA',  NULL, 1, '0.01', '{"asset_kind":"bond","coupon_pct":3.45,"maturity":"2044-06-15"}'),
    ('d0003000-0000-0000-0000-00000000000c', 'US-T-10Y', 'US Treasury 10Y',               'FIXED_INCOME', 'GOV_BOND',    'USD', 'US', 'OTC',   NULL, 1, '0.01', '{"asset_kind":"bond","coupon_pct":4.25,"maturity":"2034-08-15"}'),

    -- Thai corporate bonds
    ('d0003000-0000-0000-0000-00000000000d', 'KBANK-23B','KBANK Senior Unsecured 2028',   'FIXED_INCOME', 'CORP_BOND',   'THB', 'TH', 'TBMA',  'FINANCIALS',  1, '0.01', '{"asset_kind":"bond","coupon_pct":3.60,"maturity":"2028-03-30","issuer":"KBANK","rating":"AA-"}'),
    ('d0003000-0000-0000-0000-00000000000e', 'CP-26C',   'CP All Senior 2026',            'FIXED_INCOME', 'CORP_BOND',   'THB', 'TH', 'TBMA',  'CONS_STAPLES',1, '0.01', '{"asset_kind":"bond","coupon_pct":3.20,"maturity":"2026-11-12","issuer":"CPALL","rating":"A+"}'),

    -- Mutual fund / MMF
    ('d0003000-0000-0000-0000-00000000000f', 'K-CASH',   'KAsset Cash Plus Fund',          'FUND',         'MMF',         'THB', 'TH', 'OTC',   NULL, 1, '0.0001', '{"asset_kind":"fund"}')
) AS v(id, primary_ticker, name, asset_class_code, asset_subtype_code, currency, country_iso, primary_exchange, sector_code, lot_size, tick_size, attributes)
JOIN investment__asset_classes ac ON ac.code = v.asset_class_code
JOIN investment__asset_subtypes ast ON ast.asset_class_id = ac.id AND ast.code = v.asset_subtype_code
JOIN investment__countries co ON co.iso_code = v.country_iso
ON CONFLICT (id) DO UPDATE SET
    primary_ticker   = EXCLUDED.primary_ticker,
    name             = EXCLUDED.name,
    asset_class_id   = EXCLUDED.asset_class_id,
    asset_subtype_id = EXCLUDED.asset_subtype_id,
    currency         = EXCLUDED.currency,
    country_id       = EXCLUDED.country_id,
    region_id        = EXCLUDED.region_id,
    primary_exchange = EXCLUDED.primary_exchange,
    sector_id        = EXCLUDED.sector_id,
    lot_size         = EXCLUDED.lot_size,
    tick_size        = EXCLUDED.tick_size,
    is_tradable      = EXCLUDED.is_tradable,
    status           = EXCLUDED.status,
    attributes       = EXCLUDED.attributes,
    updated_at       = NOW(),
    updated_by       = EXCLUDED.updated_by;

COMMIT;
