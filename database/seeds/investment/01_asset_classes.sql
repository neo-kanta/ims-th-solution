-- Phase 0 reference data: asset classes.
-- Idempotent: re-running this seed leaves any operator-renamed rows alone.

INSERT INTO investment__asset_classes (code, name, description, sort_order, is_active) VALUES
    ('EQUITY',       'Equity',          'Common and preferred shares of listed and unlisted companies.', 10, true),
    ('FIXED_INCOME', 'Fixed Income',    'Government and corporate debt instruments.',                     20, true),
    ('FUND',         'Fund',            'Pooled investment vehicles (mutual funds, money market funds).', 30, true),
    ('ETF',          'ETF',             'Exchange-traded funds.',                                          40, true),
    ('CASH',         'Cash & Equivalent', 'Cash positions and overnight cash equivalents.',                50, true),
    ('ALTERNATIVE',  'Alternative',     'Real assets and other non-traditional asset classes.',            60, true),
    ('DERIVATIVE',   'Derivative',      'Listed and OTC derivatives.',                                     70, true)
ON CONFLICT (code) DO NOTHING;
