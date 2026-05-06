-- Phase 0 reference data: fund categories.

INSERT INTO investment__fund_categories (code, name, description, sort_order, is_active) VALUES
    ('EQUITY_FUND', 'Equity Fund',     'Fund holding predominantly equities.',                10, true),
    ('BOND_FUND',   'Bond Fund',       'Fund holding predominantly fixed-income securities.', 20, true),
    ('MIXED_FUND',  'Mixed Fund',      'Balanced or allocation fund.',                        30, true),
    ('MMF_FUND',    'Money Market Fund', 'Cash and cash-equivalent fund.',                    40, true),
    ('INDEX_FUND',  'Index Fund',      'Passive fund tracking a published index.',            50, true)
ON CONFLICT (code) DO NOTHING;
