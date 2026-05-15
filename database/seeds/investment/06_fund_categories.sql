-- Phase 0 reference data: fund categories.
-- Schema: investment__fund_categories (id UUID PK, code UNIQUE, name,
-- asset_class_id UUID FK to investment__asset_classes(id), description,
-- is_active). The asset_class_id is resolved via JOIN on asset-class code.

INSERT INTO investment__fund_categories (code, name, asset_class_id, description, is_active)
SELECT v.code, v.name, ac.id, v.description, true
FROM (VALUES
    ('EQUITY_FUND', 'Equity Fund',       'EQUITY',       'Fund holding predominantly equities.'),
    ('BOND_FUND',   'Bond Fund',         'FIXED_INCOME', 'Fund holding predominantly fixed-income securities.'),
    ('MIXED_FUND',  'Mixed Fund',        NULL,           'Balanced or allocation fund.'),
    ('MMF_FUND',    'Money Market Fund', 'CASH',         'Cash and cash-equivalent fund.'),
    ('INDEX_FUND',  'Index Fund',        NULL,           'Passive fund tracking a published index.')
) AS v(code, name, asset_class_code, description)
LEFT JOIN investment__asset_classes ac ON ac.code = v.asset_class_code
ON CONFLICT (code) DO NOTHING;
