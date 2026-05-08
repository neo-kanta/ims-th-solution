-- Phase 0 reference data: asset subtypes (12 entries, mapped to asset classes).

INSERT INTO investment__asset_subtypes (code, asset_class, name, description, sort_order, is_active) VALUES
    ('COMMON_STOCK',    'EQUITY',       'Common Stock',                'Ordinary equity shares with voting rights.',          10, true),
    ('PREFERRED_STOCK', 'EQUITY',       'Preferred Stock',             'Non-voting equity with priority dividend treatment.', 20, true),
    ('EQUITY_ETF',      'ETF',          'Equity ETF',                  'Exchange-traded fund holding equities.',              10, true),
    ('BOND_ETF',        'ETF',          'Bond ETF',                    'Exchange-traded fund holding fixed-income.',          20, true),
    ('MUTUAL_FUND',     'FUND',         'Mutual Fund',                 'Open-ended pooled investment vehicle.',               10, true),
    ('MMF_FUND',        'FUND',         'Money Market Fund',           'Cash-equivalent pooled vehicle.',                     20, true),
    ('INDEX_FUND',      'FUND',         'Index Fund',                  'Passive fund tracking a published index.',            30, true),
    ('GOV_BOND',        'FIXED_INCOME', 'Government Bond',             'Sovereign or quasi-sovereign debt.',                  10, true),
    ('CORP_BOND',       'FIXED_INCOME', 'Corporate Bond',              'Debt issued by non-government corporates.',          20, true),
    ('OPTION',          'DERIVATIVE',   'Option',                      'Listed option contract.',                            10, true),
    ('FUTURE',          'DERIVATIVE',   'Future',                      'Listed futures contract.',                           20, true),
    ('WARRANT',         'DERIVATIVE',   'Warrant',                     'Equity warrant.',                                    30, true)
ON CONFLICT (code) DO NOTHING;
