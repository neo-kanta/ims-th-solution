-- Table: investment__process_step_assignments
-- Source: 20260427000002_investment__create_process_assignment_tables.up.sql
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
