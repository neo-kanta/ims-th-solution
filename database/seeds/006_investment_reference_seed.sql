-- =============================================================================
-- Investment reference / classification seed
-- =============================================================================
-- Seeds the investment-domain taxonomy required for instrument classification
-- and fund categorisation. Idempotent: re-running is safe (ON CONFLICT DO NOTHING).
--
-- Seeded by: SYSTEM admin user a0000000-0000-0000-0000-000000000001 (matches
-- existing seed convention in 001_initial_seed.sql).
-- =============================================================================

BEGIN;

-- ---------------------------------------------------------------------------
-- Asset classes
-- ---------------------------------------------------------------------------
INSERT INTO
    investment__asset_classes (
        code,
        name,
        description,
        display_order,
        created_by,
        updated_by
    )
VALUES (
        'EQUITY',
        'Equity',
        'Common and preferred shares of listed corporations.',
        10,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'FIXED_INCOME',
        'Fixed Income',
        'Government, corporate and structured debt securities.',
        20,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'FUND',
        'Fund',
        'Mutual funds, money-market funds and other pooled vehicles.',
        30,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'ETF',
        'ETF',
        'Exchange-traded funds.',
        40,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'CASH',
        'Cash',
        'Cash and cash equivalents.',
        50,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'ALTERNATIVE',
        'Alternative',
        'Private equity, hedge funds, real estate, infrastructure.',
        60,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'DERIVATIVE',
        'Derivative',
        'Futures, options, swaps and other derivative contracts.',
        70,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ) ON CONFLICT (code) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Asset subtypes
-- ---------------------------------------------------------------------------
INSERT INTO
    investment__asset_subtypes (
        asset_class_id,
        code,
        name,
        display_order,
        created_by,
        updated_by
    )
SELECT ac.id, v.code, v.name, v.display_order, 'a0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001'
FROM (
        VALUES (
                'EQUITY',
                'COMMON_STOCK',
                'Common Stock',
                10
            ),
            (
                'EQUITY',
                'PREFERRED_STOCK',
                'Preferred Stock',
                20
            ),
            (
                'EQUITY',
                'DEPOSITARY_RECEIPT',
                'Depositary Receipt',
                30
            ),
            (
                'FIXED_INCOME',
                'GOV_BOND',
                'Government Bond',
                10
            ),
            (
                'FIXED_INCOME',
                'CORP_BOND',
                'Corporate Bond',
                20
            ),
            (
                'FIXED_INCOME',
                'STRUCTURED_NOTE',
                'Structured Note',
                30
            ),
            (
                'FUND',
                'MUTUAL_FUND',
                'Mutual Fund',
                10
            ),
            (
                'FUND',
                'INDEX_FUND',
                'Index Fund',
                20
            ),
            (
                'FUND',
                'MMF',
                'Money Market Fund',
                30
            ),
            (
                'FUND',
                'REIT',
                'REIT (Fund Type)',
                40
            ),
            (
                'FUND',
                'INFRA_FUND',
                'Infrastructure Fund',
                50
            ),
            (
                'ETF',
                'EQUITY_ETF',
                'Equity ETF',
                10
            ),
            (
                'ETF',
                'BOND_ETF',
                'Bond ETF',
                20
            ),
            (
                'ETF',
                'COMMODITY_ETF',
                'Commodity ETF',
                30
            ),
            (
                'CASH',
                'CASH_DEPOSIT',
                'Cash Deposit',
                10
            ),
            (
                'ALTERNATIVE',
                'PRIVATE_EQUITY',
                'Private Equity',
                10
            ),
            (
                'ALTERNATIVE',
                'HEDGE_FUND',
                'Hedge Fund',
                20
            ),
            (
                'ALTERNATIVE',
                'REAL_ESTATE',
                'Real Estate',
                30
            ),
            (
                'DERIVATIVE',
                'FUTURES',
                'Futures',
                10
            ),
            (
                'DERIVATIVE',
                'OPTION',
                'Option',
                20
            ),
            (
                'DERIVATIVE',
                'SWAP',
                'Swap',
                30
            )
    ) AS v (
        class_code,
        code,
        name,
        display_order
    )
    JOIN investment__asset_classes ac ON ac.code = v.class_code ON CONFLICT (asset_class_id, code) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Regions
-- ---------------------------------------------------------------------------
INSERT INTO
    investment__regions (
        code,
        name,
        created_by,
        updated_by
    )
VALUES (
        'GLOBAL',
        'Global',
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'APAC',
        'Asia-Pacific',
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'EMEA',
        'Europe / Middle-East / Africa',
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'AMER',
        'Americas',
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'THAILAND',
        'Thailand',
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ) ON CONFLICT (code) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Countries (subset relevant to PoC)
-- ---------------------------------------------------------------------------
INSERT INTO
    investment__countries (
        iso_code,
        name,
        region_id,
        created_by,
        updated_by
    )
SELECT v.iso_code, v.name, r.id, 'a0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001'
FROM (
        VALUES ('TH', 'Thailand', 'THAILAND'),
            ('US', 'United States', 'AMER'),
            ('JP', 'Japan', 'APAC'),
            (
                'GB',
                'United Kingdom',
                'EMEA'
            ),
            ('DE', 'Germany', 'EMEA'),
            ('SG', 'Singapore', 'APAC'),
            ('HK', 'Hong Kong SAR', 'APAC'),
            ('CN', 'China', 'APAC'),
            ('AU', 'Australia', 'APAC'),
            ('KR', 'South Korea', 'APAC')
    ) AS v (iso_code, name, region_code)
    JOIN investment__regions r ON r.code = v.region_code ON CONFLICT (iso_code) DO NOTHING;

-- ---------------------------------------------------------------------------
-- GICS-style sector hierarchy (Level 1 only; sub-levels added later if needed)
-- ---------------------------------------------------------------------------
INSERT INTO
    investment__sectors (
        code,
        name,
        level,
        created_by,
        updated_by
    )
VALUES (
        'ENERGY',
        'Energy',
        1,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'MATERIALS',
        'Materials',
        1,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'INDUSTRIALS',
        'Industrials',
        1,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'CONS_DISC',
        'Consumer Discretionary',
        1,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'CONS_STAPLES',
        'Consumer Staples',
        1,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'HEALTH_CARE',
        'Health Care',
        1,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'FINANCIALS',
        'Financials',
        1,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'INFO_TECH',
        'Information Technology',
        1,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'COMM_SVCS',
        'Communication Services',
        1,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'UTILITIES',
        'Utilities',
        1,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'REAL_ESTATE',
        'Real Estate',
        1,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ) ON CONFLICT (parent_id, code) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Fund categories
-- ---------------------------------------------------------------------------
INSERT INTO
    investment__fund_categories (
        code,
        name,
        asset_class_id,
        created_by,
        updated_by
    )
SELECT v.code, v.name, ac.id, 'a0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001'
FROM (
        VALUES (
                'EQUITY_FUND',
                'Equity Fund',
                'EQUITY'
            ),
            (
                'BOND_FUND',
                'Bond Fund',
                'FIXED_INCOME'
            ),
            (
                'MIXED_FUND',
                'Mixed / Balanced Fund',
                'FUND'
            ),
            (
                'MMF_FUND',
                'Money Market Fund',
                'CASH'
            ),
            (
                'INDEX_FUND',
                'Index Fund',
                'EQUITY'
            ),
            (
                'SECTOR_FUND',
                'Sector Fund',
                'EQUITY'
            ),
            (
                'THEMATIC_FUND',
                'Thematic Fund',
                'EQUITY'
            ),
            (
                'FOF',
                'Fund of Funds',
                'FUND'
            ),
            (
                'PROP_PORT',
                'Proprietary Portfolio (no units)',
                'FUND'
            )
    ) AS v (code, name, class_code)
    LEFT JOIN investment__asset_classes ac ON ac.code = v.class_code ON CONFLICT (code) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Investment styles
-- ---------------------------------------------------------------------------
INSERT INTO
    investment__investment_styles (
        code,
        name,
        created_by,
        updated_by
    )
VALUES (
        'GROWTH',
        'Growth',
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'VALUE',
        'Value',
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'BLEND',
        'Blend',
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'INCOME',
        'Income',
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'INDEX',
        'Index',
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'ACTIVE',
        'Active',
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ) ON CONFLICT (code) DO NOTHING;

COMMIT;