-- Phase 0 reference data: GICS Level 1 (sectors) and Level 2 (industry groups).
--
-- Schema: investment__sectors (id UUID PK, code, name, parent_id UUID FK
-- to self, level SMALLINT, is_active). UNIQUE on (parent_id, code). Labels
-- are project-authored short names to avoid reproducing MSCI proprietary
-- descriptions; the GICS numeric codes are dropped from this seed because
-- the canonical schema does not carry a gics_code column.
--
-- Idempotency: NULLs are distinct under the default UNIQUE constraint on
-- (parent_id, code), so ON CONFLICT will not fire for level-1 rows whose
-- parent_id is NULL. WHERE NOT EXISTS guards against duplicate inserts on
-- repeat runs.

-- Level 1 — 11 GICS sectors
INSERT INTO investment__sectors (code, name, parent_id, level, is_active)
SELECT v.code, v.name, NULL, 1, true
FROM (VALUES
    ('ENERGY',                'Energy'),
    ('MATERIALS',             'Materials'),
    ('INDUSTRIALS',           'Industrials'),
    ('CONSUMER_DISCRETIONARY','Consumer Discretionary'),
    ('CONSUMER_STAPLES',      'Consumer Staples'),
    ('HEALTH_CARE',           'Health Care'),
    ('FINANCIALS',            'Financials'),
    ('INFORMATION_TECHNOLOGY','Information Technology'),
    ('COMMUNICATION_SERVICES','Communication Services'),
    ('UTILITIES',             'Utilities'),
    ('REAL_ESTATE',           'Real Estate')
) AS v(code, name)
WHERE NOT EXISTS (
    SELECT 1 FROM investment__sectors s
    WHERE s.parent_id IS NULL AND s.code = v.code
);

-- Level 2 — 24 GICS industry groups (parent_id resolved via JOIN on parent code)
INSERT INTO investment__sectors (code, name, parent_id, level, is_active)
SELECT v.code, v.name, p.id, 2, true
FROM (VALUES
    ('ENERGY_GROUP',                       'Energy',                                      'ENERGY'),
    ('MATERIALS_GROUP',                    'Materials',                                   'MATERIALS'),
    ('CAPITAL_GOODS',                      'Capital Goods',                               'INDUSTRIALS'),
    ('COMMERCIAL_PROFESSIONAL_SERVICES',   'Commercial & Professional Services',          'INDUSTRIALS'),
    ('TRANSPORTATION',                     'Transportation',                              'INDUSTRIALS'),
    ('AUTOMOBILES_COMPONENTS',             'Automobiles & Components',                    'CONSUMER_DISCRETIONARY'),
    ('CONSUMER_DURABLES_APPAREL',          'Consumer Durables & Apparel',                 'CONSUMER_DISCRETIONARY'),
    ('CONSUMER_SERVICES',                  'Consumer Services',                           'CONSUMER_DISCRETIONARY'),
    ('CONSUMER_DISCRETIONARY_DISTRIBUTION','Consumer Discretionary Distribution & Retail','CONSUMER_DISCRETIONARY'),
    ('CONSUMER_STAPLES_DISTRIBUTION',      'Consumer Staples Distribution & Retail',      'CONSUMER_STAPLES'),
    ('FOOD_BEVERAGE_TOBACCO',              'Food, Beverage & Tobacco',                    'CONSUMER_STAPLES'),
    ('HOUSEHOLD_PERSONAL_PRODUCTS',        'Household & Personal Products',               'CONSUMER_STAPLES'),
    ('HEALTH_CARE_EQUIPMENT_SERVICES',     'Health Care Equipment & Services',            'HEALTH_CARE'),
    ('PHARMA_BIOTECH_LIFE_SCIENCES',       'Pharmaceuticals, Biotechnology & Life Sciences','HEALTH_CARE'),
    ('BANKS',                              'Banks',                                       'FINANCIALS'),
    ('FINANCIAL_SERVICES',                 'Financial Services',                          'FINANCIALS'),
    ('INSURANCE',                          'Insurance',                                   'FINANCIALS'),
    ('SOFTWARE_SERVICES',                  'Software & Services',                         'INFORMATION_TECHNOLOGY'),
    ('TECH_HARDWARE_EQUIPMENT',            'Technology Hardware & Equipment',             'INFORMATION_TECHNOLOGY'),
    ('SEMICONDUCTORS_EQUIPMENT',           'Semiconductors & Semiconductor Equipment',    'INFORMATION_TECHNOLOGY'),
    ('TELECOMMUNICATION_SERVICES',         'Telecommunication Services',                  'COMMUNICATION_SERVICES'),
    ('MEDIA_ENTERTAINMENT',                'Media & Entertainment',                       'COMMUNICATION_SERVICES'),
    ('UTILITIES_GROUP',                    'Utilities',                                   'UTILITIES'),
    ('REAL_ESTATE_GROUP',                  'Equity Real Estate Investment Trusts (REITs)','REAL_ESTATE')
) AS v(code, name, parent_code)
JOIN investment__sectors p ON p.code = v.parent_code AND p.level = 1
ON CONFLICT (parent_id, code) DO NOTHING;
