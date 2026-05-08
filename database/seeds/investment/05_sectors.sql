-- Phase 0 reference data: GICS Level 1 (sectors) and Level 2 (industry groups).
--
-- GICS numeric codes are sourced from the published GICS taxonomy. Labels are
-- project-authored short names to avoid reproducing MSCI proprietary
-- descriptions; the gics_code column is the canonical machine identifier.
--
-- Order: Level 1 sectors are inserted first so Level 2 industry groups can
-- resolve their parent_code via FK.

-- Level 1 — 11 GICS sectors
INSERT INTO investment__sectors (code, parent_code, level, gics_code, name, description, sort_order, is_active) VALUES
    ('ENERGY',                NULL, 1, '10', 'Energy',                  'Energy producers and equipment.',          10, true),
    ('MATERIALS',             NULL, 1, '15', 'Materials',               'Chemicals, metals and construction materials.', 20, true),
    ('INDUSTRIALS',           NULL, 1, '20', 'Industrials',             'Capital goods, transport and commercial services.', 30, true),
    ('CONSUMER_DISCRETIONARY', NULL, 1, '25', 'Consumer Discretionary', 'Cyclical consumer goods and services.',    40, true),
    ('CONSUMER_STAPLES',      NULL, 1, '30', 'Consumer Staples',        'Defensive consumer goods.',                50, true),
    ('HEALTH_CARE',           NULL, 1, '35', 'Health Care',             'Pharma, biotech and health-care services.', 60, true),
    ('FINANCIALS',            NULL, 1, '40', 'Financials',              'Banks, insurers and capital markets.',     70, true),
    ('INFORMATION_TECHNOLOGY', NULL, 1, '45', 'Information Technology', 'Software, semiconductors and tech hardware.', 80, true),
    ('COMMUNICATION_SERVICES', NULL, 1, '50', 'Communication Services', 'Telecoms, media and interactive services.', 90, true),
    ('UTILITIES',             NULL, 1, '55', 'Utilities',               'Regulated and unregulated utilities.',     100, true),
    ('REAL_ESTATE',           NULL, 1, '60', 'Real Estate',             'REITs and real-estate management.',        110, true)
ON CONFLICT (code) DO NOTHING;

-- Level 2 — 24 GICS industry groups
INSERT INTO investment__sectors (code, parent_code, level, gics_code, name, description, sort_order, is_active) VALUES
    ('ENERGY_GROUP',                       'ENERGY',                  2, '1010', 'Energy',                          'Integrated oil & gas, exploration and equipment.',    10, true),
    ('MATERIALS_GROUP',                    'MATERIALS',               2, '1510', 'Materials',                       'Chemicals, metals and mining, paper & forest.',       20, true),
    ('CAPITAL_GOODS',                      'INDUSTRIALS',             2, '2010', 'Capital Goods',                   'Construction, machinery, electrical equipment.',      30, true),
    ('COMMERCIAL_PROFESSIONAL_SERVICES',   'INDUSTRIALS',             2, '2020', 'Commercial & Professional Services', 'Outsourced and professional services.',          40, true),
    ('TRANSPORTATION',                     'INDUSTRIALS',             2, '2030', 'Transportation',                  'Air, marine, road, rail and infrastructure.',         50, true),
    ('AUTOMOBILES_COMPONENTS',             'CONSUMER_DISCRETIONARY',  2, '2510', 'Automobiles & Components',         'Auto OEMs and parts suppliers.',                     60, true),
    ('CONSUMER_DURABLES_APPAREL',          'CONSUMER_DISCRETIONARY',  2, '2520', 'Consumer Durables & Apparel',      'Household durables, leisure goods, apparel.',         70, true),
    ('CONSUMER_SERVICES',                  'CONSUMER_DISCRETIONARY',  2, '2530', 'Consumer Services',                'Hotels, restaurants and education services.',         80, true),
    ('CONSUMER_DISCRETIONARY_DISTRIBUTION', 'CONSUMER_DISCRETIONARY', 2, '2550', 'Consumer Discretionary Distribution & Retail', 'Distributors and retail.',                90, true),
    ('CONSUMER_STAPLES_DISTRIBUTION',      'CONSUMER_STAPLES',        2, '3010', 'Consumer Staples Distribution & Retail', 'Food and staples retail and distribution.',     100, true),
    ('FOOD_BEVERAGE_TOBACCO',              'CONSUMER_STAPLES',        2, '3020', 'Food, Beverage & Tobacco',         'Food producers, beverages, tobacco.',               110, true),
    ('HOUSEHOLD_PERSONAL_PRODUCTS',        'CONSUMER_STAPLES',        2, '3030', 'Household & Personal Products',    'Household and personal-care goods.',                120, true),
    ('HEALTH_CARE_EQUIPMENT_SERVICES',     'HEALTH_CARE',             2, '3510', 'Health Care Equipment & Services', 'Equipment, providers and services.',                130, true),
    ('PHARMA_BIOTECH_LIFE_SCIENCES',       'HEALTH_CARE',             2, '3520', 'Pharmaceuticals, Biotechnology & Life Sciences', 'Drug-makers and life-science tools.',  140, true),
    ('BANKS',                              'FINANCIALS',              2, '4010', 'Banks',                            'Diversified and regional banks.',                    150, true),
    ('FINANCIAL_SERVICES',                 'FINANCIALS',              2, '4020', 'Financial Services',               'Capital markets, brokerage, exchanges.',             160, true),
    ('INSURANCE',                          'FINANCIALS',              2, '4030', 'Insurance',                        'Life, P&C, multi-line and brokers.',                170, true),
    ('SOFTWARE_SERVICES',                  'INFORMATION_TECHNOLOGY',  2, '4510', 'Software & Services',              'Application software, IT services.',                180, true),
    ('TECH_HARDWARE_EQUIPMENT',            'INFORMATION_TECHNOLOGY',  2, '4520', 'Technology Hardware & Equipment',  'Hardware, networking and storage.',                 190, true),
    ('SEMICONDUCTORS_EQUIPMENT',           'INFORMATION_TECHNOLOGY',  2, '4530', 'Semiconductors & Semiconductor Equipment', 'Chips and chip equipment.',                  200, true),
    ('TELECOMMUNICATION_SERVICES',         'COMMUNICATION_SERVICES',  2, '5010', 'Telecommunication Services',        'Wireline and wireless telecoms.',                   210, true),
    ('MEDIA_ENTERTAINMENT',                'COMMUNICATION_SERVICES',  2, '5020', 'Media & Entertainment',             'Media, entertainment and interactive services.',    220, true),
    ('UTILITIES_GROUP',                    'UTILITIES',               2, '5510', 'Utilities',                         'Electric, gas, water and renewable utilities.',     230, true),
    ('REAL_ESTATE_GROUP',                  'REAL_ESTATE',             2, '6010', 'Equity Real Estate Investment Trusts (REITs)', 'Equity REITs across sub-types.',           240, true)
ON CONFLICT (code) DO NOTHING;
