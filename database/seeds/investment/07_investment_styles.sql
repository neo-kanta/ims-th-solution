-- Phase 0 reference data: investment styles.

INSERT INTO investment__investment_styles (code, name, description, sort_order, is_active) VALUES
    ('GROWTH', 'Growth', 'Targets companies with above-average earnings growth.',         10, true),
    ('VALUE',  'Value',  'Targets companies trading below estimated intrinsic value.',    20, true),
    ('BLEND',  'Blend',  'Combines growth and value characteristics.',                    30, true),
    ('INCOME', 'Income', 'Targets dividend yield and stable cash distributions.',         40, true),
    ('INDEX',  'Index',  'Replicates the composition of a published index.',              50, true)
ON CONFLICT (code) DO NOTHING;
