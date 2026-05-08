-- Phase 0 reference data: asset subtypes (12 entries, FK by code-lookup).
-- Schema: investment__asset_subtypes (id UUID PK, asset_class_id UUID FK to
-- investment__asset_classes(id), code, name, description, display_order,
-- is_active). UNIQUE on (asset_class_id, code).
-- The asset_class_id is resolved by code via SELECT in each row so the seed
-- doesn't need hand-typed UUIDs.

INSERT INTO investment__asset_subtypes (asset_class_id, code, name, description, display_order, is_active)
SELECT ac.id, v.code, v.name, v.description, v.display_order, true
FROM (VALUES
    ('EQUITY',       'COMMON_STOCK',    'Common Stock',         'Ordinary equity shares with voting rights.',          10),
    ('EQUITY',       'PREFERRED_STOCK', 'Preferred Stock',      'Non-voting equity with priority dividend treatment.', 20),
    ('ETF',          'EQUITY_ETF',      'Equity ETF',           'Exchange-traded fund holding equities.',              10),
    ('ETF',          'BOND_ETF',        'Bond ETF',             'Exchange-traded fund holding fixed-income.',          20),
    ('FUND',         'MUTUAL_FUND',     'Mutual Fund',          'Open-ended pooled investment vehicle.',               10),
    ('FUND',         'MMF_FUND',        'Money Market Fund',    'Cash-equivalent pooled vehicle.',                     20),
    ('FUND',         'INDEX_FUND',      'Index Fund',           'Passive fund tracking a published index.',            30),
    ('FIXED_INCOME', 'GOV_BOND',        'Government Bond',      'Sovereign or quasi-sovereign debt.',                  10),
    ('FIXED_INCOME', 'CORP_BOND',       'Corporate Bond',       'Debt issued by non-government corporates.',           20),
    ('DERIVATIVE',   'OPTION',          'Option',               'Listed option contract.',                             10),
    ('DERIVATIVE',   'FUTURE',          'Future',               'Listed futures contract.',                            20),
    ('DERIVATIVE',   'WARRANT',         'Warrant',              'Equity warrant.',                                     30)
) AS v(asset_class_code, code, name, description, display_order)
JOIN investment__asset_classes ac ON ac.code = v.asset_class_code
ON CONFLICT (asset_class_id, code) DO NOTHING;
