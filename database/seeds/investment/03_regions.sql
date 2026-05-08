-- Phase 0 reference data: regions used to bucket countries.

INSERT INTO investment__regions (code, name, description, sort_order, is_active) VALUES
    ('GLOBAL',     'Global',                'Cross-region scope.',                            10, true),
    ('THAILAND',   'Thailand',              'Domestic market scope.',                         20, true),
    ('APAC',       'Asia-Pacific',          'Asia-Pacific including Japan, ANZ, ASEAN, GC.',  30, true),
    ('EMEA',       'EMEA',                  'Europe, Middle East and Africa.',                40, true),
    ('AMER',       'Americas',              'North, Central and South America.',              50, true),
    ('DEVELOPED',  'Developed Markets',     'Aggregate of MSCI-classified developed markets.', 60, true),
    ('EMERGING',   'Emerging Markets',      'Aggregate of MSCI-classified emerging markets.',  70, true)
ON CONFLICT (code) DO NOTHING;
