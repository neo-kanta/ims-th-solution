-- Table: permissions_function_definitions
-- Source: 20260506000002_permissions__create_function_definitions.up.sql
CREATE TABLE permissions_function_definitions (
    code            VARCHAR(100) PRIMARY KEY,
    module          VARCHAR(60)  NOT NULL,
    name            VARCHAR(160) NOT NULL,
    description     TEXT,
    deprecated_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_permissions_function_definitions_code
        CHECK (code = UPPER(code) AND code !~ '\s'),
    CONSTRAINT chk_permissions_function_definitions_module
        CHECK (module = LOWER(module) AND module !~ '\s')
);
