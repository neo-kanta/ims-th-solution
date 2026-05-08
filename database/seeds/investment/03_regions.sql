-- Phase 0 reference data: regions used to bucket countries.
-- Schema: investment__regions (id UUID PK, code UNIQUE, name, is_active).
-- The user's canonical schema does not carry description / display_order
-- on regions; the names below already convey ordering intent.

INSERT INTO investment__regions (code, name, is_active) VALUES
    ('GLOBAL',    'Global',            true),
    ('THAILAND',  'Thailand',          true),
    ('APAC',      'Asia-Pacific',      true),
    ('EMEA',      'EMEA',              true),
    ('AMER',      'Americas',          true),
    ('DEVELOPED', 'Developed Markets', true),
    ('EMERGING',  'Emerging Markets',  true)
ON CONFLICT (code) DO NOTHING;
