-- Rollback: Compliance / IRG tables
-- Order matters: delete from most-dependent to least-dependent.

-- Remove seeded permissions
DELETE FROM permissions_function_rights
WHERE permission_code IN (
    'IRG_VIEW_RULES',
    'IRG_EDIT_RULE_INSTANCE',
    'IRG_EDIT_BINDING',
    'IRG_OVERRIDE_BREACH',
    'IRG_ADMIN_RULE_TYPE'
);

DROP TABLE IF EXISTS compliance_overrides          CASCADE;
DROP TABLE IF EXISTS compliance_breaches           CASCADE;
DROP TABLE IF EXISTS compliance_check_records      CASCADE;
DROP TABLE IF EXISTS compliance_rule_bindings      CASCADE;
DROP TABLE IF EXISTS compliance_rule_set_members   CASCADE;
DROP TABLE IF EXISTS compliance_rule_sets          CASCADE;
DROP TABLE IF EXISTS compliance_rule_instance_versions CASCADE;
DROP TABLE IF EXISTS compliance_rule_instances     CASCADE;
