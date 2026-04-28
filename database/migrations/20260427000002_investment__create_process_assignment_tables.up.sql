-- =============================================================================
-- Investment Module - Process Assignment Tables
-- =============================================================================
-- Phase 2 adds assignment settings for the investment process:
--   Analysis Report -> Investment Decisions -> Investment Execution -> Investment Review
--
-- The tables support:
--   - stable process-step definitions
--   - process groups, such as Group A and Group B
--   - group members mapped to IAM users
--   - process-step authorization by group or by direct user assignment
--
-- Runtime investment records are intentionally not created here. These tables
-- are configuration/authorization data that future command handlers should
-- check before allowing investment process actions.
-- =============================================================================

-- ---------------------------------------------------------------------------
-- 1. Process step definitions
-- ---------------------------------------------------------------------------
CREATE TABLE investment__process_steps (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    step_key                VARCHAR(60) NOT NULL,
    name                    VARCHAR(120) NOT NULL,
    description             TEXT,
    sequence_no             SMALLINT    NOT NULL,

    requires_workflow_open  BOOLEAN     NOT NULL DEFAULT true,
    blocks_after_rejection  BOOLEAN     NOT NULL DEFAULT true,

    is_active               BOOLEAN     NOT NULL DEFAULT true,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by              UUID        REFERENCES iam_users(id),
    updated_by              UUID        REFERENCES iam_users(id),

    CONSTRAINT uq_investment_process_steps_key UNIQUE (step_key),
    CONSTRAINT uq_investment_process_steps_sequence UNIQUE (sequence_no),

    CONSTRAINT chk_investment_process_steps_key
        CHECK (step_key IN (
            'ANALYSIS_REPORT',
            'INVESTMENT_DECISION',
            'INVESTMENT_EXECUTION',
            'INVESTMENT_REVIEW'
        )),

    CONSTRAINT chk_investment_process_steps_sequence
        CHECK (sequence_no > 0)
);

CREATE INDEX idx_investment_process_steps_active ON investment__process_steps (is_active, sequence_no);

CREATE TRIGGER trg_investment_process_steps_updated_at
    BEFORE UPDATE ON investment__process_steps
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE investment__process_steps IS 'Stable investment process steps used for authorization and workflow guards.';
COMMENT ON COLUMN investment__process_steps.blocks_after_rejection IS 'When true, manager/high-level rejection blocks this process step.';

-- ---------------------------------------------------------------------------
-- 2. Process groups
-- ---------------------------------------------------------------------------
CREATE TABLE investment__process_groups (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    group_key       VARCHAR(80) NOT NULL,
    name            VARCHAR(120) NOT NULL,
    description     TEXT,

    is_active       BOOLEAN     NOT NULL DEFAULT true,
    effective_from  DATE        NOT NULL,
    effective_to    DATE,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      UUID        REFERENCES iam_users(id),
    updated_by      UUID        REFERENCES iam_users(id),

    CONSTRAINT uq_investment_process_groups_key UNIQUE (group_key),
    CONSTRAINT uq_investment_process_groups_name UNIQUE (name),

    CONSTRAINT chk_investment_process_groups_effective_range
        CHECK (effective_to IS NULL OR effective_to >= effective_from)
);

CREATE INDEX idx_investment_process_groups_active ON investment__process_groups (is_active, effective_from, effective_to);

CREATE TRIGGER trg_investment_process_groups_updated_at
    BEFORE UPDATE ON investment__process_groups
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE investment__process_groups IS 'Named investment process working groups, such as Group A or Group B.';

-- ---------------------------------------------------------------------------
-- 3. Process group members
-- ---------------------------------------------------------------------------
CREATE TABLE investment__process_group_members (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    group_id        UUID        NOT NULL REFERENCES investment__process_groups(id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,

    member_role     VARCHAR(30) NOT NULL DEFAULT 'MEMBER',
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    effective_from  DATE        NOT NULL,
    effective_to    DATE,

    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by     UUID        REFERENCES iam_users(id),

    CONSTRAINT uq_investment_process_group_member UNIQUE (group_id, user_id),

    CONSTRAINT chk_investment_process_group_member_role
        CHECK (member_role IN ('LEAD', 'MEMBER', 'BACKUP')),

    CONSTRAINT chk_investment_process_group_member_effective_range
        CHECK (effective_to IS NULL OR effective_to >= effective_from)
);

CREATE INDEX idx_investment_process_group_members_group ON investment__process_group_members (group_id);
CREATE INDEX idx_investment_process_group_members_user ON investment__process_group_members (user_id);
CREATE INDEX idx_investment_process_group_members_active ON investment__process_group_members (is_active, effective_from, effective_to);

COMMENT ON TABLE investment__process_group_members IS 'Maps IAM users to investment process groups.';

-- ---------------------------------------------------------------------------
-- 4. Process step assignments
-- ---------------------------------------------------------------------------
CREATE TABLE investment__process_step_assignments (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    process_step_id     UUID        NOT NULL REFERENCES investment__process_steps(id) ON DELETE CASCADE,

    -- GROUP assignment covers all active group members.
    -- USER assignment allows direct one-person assignment without a group.
    assignment_type     VARCHAR(20) NOT NULL,
    process_group_id    UUID        REFERENCES investment__process_groups(id) ON DELETE CASCADE,
    user_id             UUID        REFERENCES iam_users(id) ON DELETE CASCADE,

    -- Scope allows global defaults first, then contract/fund overrides later.
    scope_type          VARCHAR(30) NOT NULL DEFAULT 'GLOBAL',
    scope_id            UUID,

    can_execute         BOOLEAN     NOT NULL DEFAULT true,
    is_active           BOOLEAN     NOT NULL DEFAULT true,
    priority            INTEGER     NOT NULL DEFAULT 100,

    effective_from      DATE        NOT NULL,
    effective_to        DATE,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          UUID        REFERENCES iam_users(id),
    updated_by          UUID        REFERENCES iam_users(id),

    CONSTRAINT chk_investment_process_assignment_type
        CHECK (assignment_type IN ('GROUP', 'USER')),

    CONSTRAINT chk_investment_process_assignment_target
        CHECK (
            (assignment_type = 'GROUP' AND process_group_id IS NOT NULL AND user_id IS NULL)
            OR (assignment_type = 'USER' AND user_id IS NOT NULL AND process_group_id IS NULL)
        ),

    CONSTRAINT chk_investment_process_assignment_scope
        CHECK (scope_type IN ('GLOBAL', 'CONTRACT', 'FUND_GROUP')),

    CONSTRAINT chk_investment_process_assignment_scope_id
        CHECK (
            (scope_type = 'GLOBAL' AND scope_id IS NULL)
            OR (scope_type <> 'GLOBAL' AND scope_id IS NOT NULL)
        ),

    CONSTRAINT chk_investment_process_assignment_priority
        CHECK (priority >= 0),

    CONSTRAINT chk_investment_process_assignment_effective_range
        CHECK (effective_to IS NULL OR effective_to >= effective_from)
);

CREATE INDEX idx_investment_process_step_assignments_step ON investment__process_step_assignments (process_step_id);
CREATE INDEX idx_investment_process_step_assignments_group ON investment__process_step_assignments (process_group_id);
CREATE INDEX idx_investment_process_step_assignments_user ON investment__process_step_assignments (user_id);
CREATE INDEX idx_investment_process_step_assignments_scope ON investment__process_step_assignments (scope_type, scope_id);
CREATE INDEX idx_investment_process_step_assignments_active ON investment__process_step_assignments (is_active, effective_from, effective_to);

CREATE UNIQUE INDEX uq_investment_process_assignment_group_global
ON investment__process_step_assignments (process_step_id, process_group_id)
WHERE assignment_type = 'GROUP' AND scope_type = 'GLOBAL';

CREATE UNIQUE INDEX uq_investment_process_assignment_group_scoped
ON investment__process_step_assignments (process_step_id, process_group_id, scope_type, scope_id)
WHERE assignment_type = 'GROUP' AND scope_type <> 'GLOBAL';

CREATE UNIQUE INDEX uq_investment_process_assignment_user_global
ON investment__process_step_assignments (process_step_id, user_id)
WHERE assignment_type = 'USER' AND scope_type = 'GLOBAL';

CREATE UNIQUE INDEX uq_investment_process_assignment_user_scoped
ON investment__process_step_assignments (process_step_id, user_id, scope_type, scope_id)
WHERE assignment_type = 'USER' AND scope_type <> 'GLOBAL';

CREATE TRIGGER trg_investment_process_step_assignments_updated_at
    BEFORE UPDATE ON investment__process_step_assignments
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE investment__process_step_assignments IS 'Authorizes groups or individual users to execute investment process steps.';
COMMENT ON COLUMN investment__process_step_assignments.can_execute IS 'When false, an assignment can explicitly deny execution in a scoped override.';
