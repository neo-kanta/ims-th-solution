-- Phase 0 reference data: investment styles.
-- Schema: investment__investment_styles (id UUID PK, code UNIQUE, name,
-- is_active). The user's canonical schema does not carry description /
-- display_order on this table.

INSERT INTO investment__investment_styles (code, name, is_active) VALUES
    ('GROWTH', 'Growth', true),
    ('VALUE',  'Value',  true),
    ('BLEND',  'Blend',  true),
    ('INCOME', 'Income', true),
    ('INDEX',  'Index',  true)
ON CONFLICT (code) DO NOTHING;
