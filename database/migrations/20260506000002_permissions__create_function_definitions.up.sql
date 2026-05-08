-- =============================================================================
-- Permissions Module - Function Definitions Catalog (Phase 0)
-- =============================================================================
-- Introduces a master catalog of every function permission code known to the
-- system. cmd/seed upserts into this table from each module's permission
-- provider; existing permissions_function_rights rows then reference the
-- catalog by FK so unknown codes can no longer be granted.
--
-- Backfill: any code already present in permissions_function_rights at the
-- time this migration runs is auto-imported with module='legacy'. The seeder
-- overwrites these rows with proper module/name/description on the next run.
-- =============================================================================

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

CREATE INDEX idx_permissions_function_definitions_module
    ON permissions_function_definitions (module)
    WHERE deprecated_at IS NULL;

CREATE INDEX idx_permissions_function_definitions_active
    ON permissions_function_definitions (deprecated_at);

CREATE TRIGGER trg_permissions_function_definitions_updated_at
    BEFORE UPDATE ON permissions_function_definitions
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE permissions_function_definitions IS 'Master catalog of every function permission code known to the system. Owned by cmd/seed; updated from each module''s permission provider.';
COMMENT ON COLUMN permissions_function_definitions.module IS 'Module that owns the code, lowercase (e.g., workflow, investment, market_data).';
COMMENT ON COLUMN permissions_function_definitions.deprecated_at IS 'When set, the code is retained for historical grants but new grants should be rejected at the application layer.';

-- ---------------------------------------------------------------------------
-- Backfill: pull any code already granted in permissions_function_rights into
-- the catalog so the FK can be added without violating existing grants.
-- ---------------------------------------------------------------------------
INSERT INTO permissions_function_definitions (code, module, name, description)
SELECT DISTINCT
    permission_code,
    'legacy',
    permission_code,
    'Auto-imported from existing permissions_function_rights. Overwrite via cmd/seed when the owning module declares this code.'
FROM permissions_function_rights
ON CONFLICT (code) DO NOTHING;

-- ---------------------------------------------------------------------------
-- FK from grants to catalog. ON UPDATE CASCADE permits a future code rename
-- to propagate; ON DELETE RESTRICT prevents accidental removal of a code that
-- is still granted.
-- ---------------------------------------------------------------------------
ALTER TABLE permissions_function_rights
    ADD CONSTRAINT fk_permissions_function_rights_definition
    FOREIGN KEY (permission_code)
    REFERENCES permissions_function_definitions (code)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;
