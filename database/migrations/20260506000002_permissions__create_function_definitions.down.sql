-- Reverse the permissions function-definitions catalog migration.

ALTER TABLE permissions_function_rights
    DROP CONSTRAINT IF EXISTS fk_permissions_function_rights_definition;

DROP TRIGGER IF EXISTS trg_permissions_function_definitions_updated_at ON permissions_function_definitions;
DROP TABLE IF EXISTS permissions_function_definitions;
