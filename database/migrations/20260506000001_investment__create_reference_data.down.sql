-- Reverse Phase 0 reference-data tables. Order respects FK dependencies.

DROP TRIGGER IF EXISTS trg_investment_investment_styles_updated_at ON investment__investment_styles;
DROP TABLE IF EXISTS investment__investment_styles;

DROP TRIGGER IF EXISTS trg_investment_fund_categories_updated_at ON investment__fund_categories;
DROP TABLE IF EXISTS investment__fund_categories;

DROP TRIGGER IF EXISTS trg_investment_sectors_updated_at ON investment__sectors;
DROP TABLE IF EXISTS investment__sectors;

DROP TRIGGER IF EXISTS trg_investment_countries_updated_at ON investment__countries;
DROP TABLE IF EXISTS investment__countries;

DROP TRIGGER IF EXISTS trg_investment_regions_updated_at ON investment__regions;
DROP TABLE IF EXISTS investment__regions;

DROP TRIGGER IF EXISTS trg_investment_asset_subtypes_updated_at ON investment__asset_subtypes;
DROP TABLE IF EXISTS investment__asset_subtypes;

DROP TRIGGER IF EXISTS trg_investment_asset_classes_updated_at ON investment__asset_classes;
DROP TABLE IF EXISTS investment__asset_classes;
