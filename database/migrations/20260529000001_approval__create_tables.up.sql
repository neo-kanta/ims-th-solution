-- =============================================================================
-- Approval Module — generic, reusable approval engine
-- =============================================================================
-- Owners: the approval module (backend/internal/approval).
--
-- This is a GENERIC approval engine that can approve any IMS business object
-- (research report, investment decision, workflow operation, leave, delegation)
-- via a polymorphic (subject_type, subject_id) reference. It is intentionally
-- SEPARATE from the permission-management `approval_workflow_*` tables, which
-- are scoped only to `permission_change_requests`.
--
-- Naming follows the repo module convention `<module>__<entity>` (double
-- underscore), matching `investment__*`, `workflow__*`, `marketdata__*`.
--
-- Cross-entity references to users are FK'd to iam_users. The approval SUBJECT
-- (subject_id) is deliberately NOT a foreign key — it is polymorphic across
-- modules and validated by the application layer.
-- =============================================================================

-- Human-readable request numbers: APR-000001, APR-000002, ...
CREATE SEQUENCE IF NOT EXISTS approval__request_no_seq START WITH 1 INCREMENT BY 1;

-- ─────────────────────────────────────────────────────────────────────────────
-- 1. Approval groups (reusable approver pools)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE approval__groups (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    group_code  VARCHAR(60)  NOT NULL,
    group_name  VARCHAR(160) NOT NULL,
    remarks     TEXT         NOT NULL DEFAULT '',
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_by  UUID         REFERENCES iam_users(id),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by  UUID         REFERENCES iam_users(id),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_groups_code UNIQUE (group_code)
);

CREATE TRIGGER trg_approval_groups_updated_at
    BEFORE UPDATE ON approval__groups
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────────────────────────────────────────
-- 2. Approval group members
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE approval__group_members (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id       UUID         NOT NULL REFERENCES approval__groups(id) ON DELETE CASCADE,
    user_id        UUID         NOT NULL REFERENCES iam_users(id) ON DELETE RESTRICT,
    priority_order INTEGER      NOT NULL DEFAULT 1,
    member_type    VARCHAR(20)  NOT NULL DEFAULT 'MEMBER',
    status         VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    is_active      BOOLEAN      NOT NULL DEFAULT true,
    created_by     UUID         REFERENCES iam_users(id),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by     UUID         REFERENCES iam_users(id),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_group_members UNIQUE (group_id, user_id),
    CONSTRAINT chk_approval_group_member_type   CHECK (member_type IN ('MEMBER','SUPERVISOR')),
    CONSTRAINT chk_approval_group_member_status CHECK (status IN ('PENDING','APPROVED','REVOKED')),
    CONSTRAINT chk_approval_group_member_priority CHECK (priority_order >= 0)
);

CREATE INDEX idx_approval_group_members_group
    ON approval__group_members (group_id, priority_order);

CREATE TRIGGER trg_approval_group_members_updated_at
    BEFORE UPDATE ON approval__group_members
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────────────────────────────────────────
-- 3. Approval teams (per contract/fund team approval)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE approval__teams (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    team_code           VARCHAR(60)  NOT NULL,
    team_name           VARCHAR(160) NOT NULL,
    remarks             TEXT         NOT NULL DEFAULT '',
    has_co_manager      BOOLEAN      NOT NULL DEFAULT false,
    min_required_stamps INTEGER      NOT NULL DEFAULT 1,
    max_allowed_stamps  INTEGER      NOT NULL DEFAULT 1,
    is_active           BOOLEAN      NOT NULL DEFAULT true,
    created_by          UUID         REFERENCES iam_users(id),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by          UUID         REFERENCES iam_users(id),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_teams_code UNIQUE (team_code),
    CONSTRAINT chk_approval_teams_stamps CHECK (
        min_required_stamps >= 1 AND max_allowed_stamps >= min_required_stamps
    )
);

CREATE TRIGGER trg_approval_teams_updated_at
    BEFORE UPDATE ON approval__teams
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────────────────────────────────────────
-- 4. Approval team ⇄ contract/fund assignment
--    One ACTIVE contract may belong to only one ACTIVE team.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE approval__team_contracts (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id        UUID         NOT NULL REFERENCES approval__teams(id) ON DELETE CASCADE,
    contract_id    UUID         NOT NULL,
    effective_date DATE         NOT NULL DEFAULT CURRENT_DATE,
    is_active      BOOLEAN      NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- A contract can be actively assigned to only one team at a time.
CREATE UNIQUE INDEX uq_approval_team_contract_active
    ON approval__team_contracts (contract_id)
    WHERE is_active = true;

CREATE INDEX idx_approval_team_contracts_team
    ON approval__team_contracts (team_id);

CREATE TRIGGER trg_approval_team_contracts_updated_at
    BEFORE UPDATE ON approval__team_contracts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────────────────────────────────────────
-- 5. Approval team members
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE approval__team_members (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id        UUID         NOT NULL REFERENCES approval__teams(id) ON DELETE CASCADE,
    user_id        UUID         NOT NULL REFERENCES iam_users(id) ON DELETE RESTRICT,
    member_type    VARCHAR(20)  NOT NULL DEFAULT 'REVIEWER_AGENT',
    priority_order INTEGER      NOT NULL DEFAULT 1,
    is_active      BOOLEAN      NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_team_members UNIQUE (team_id, user_id),
    CONSTRAINT chk_approval_team_member_type
        CHECK (member_type IN ('ORDER_SUBMITTER','REVIEWER_AGENT')),
    CONSTRAINT chk_approval_team_member_priority CHECK (priority_order >= 0)
);

CREATE INDEX idx_approval_team_members_team
    ON approval__team_members (team_id, priority_order);

CREATE TRIGGER trg_approval_team_members_updated_at
    BEFORE UPDATE ON approval__team_members
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────────────────────────────────────────
-- 6. Approval process configuration (by process type + contract scope)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE approval__process_configs (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    process_code           VARCHAR(60)  NOT NULL,
    process_name           VARCHAR(160) NOT NULL,
    process_type           VARCHAR(40)  NOT NULL,
    contract_type          VARCHAR(20)  NOT NULL DEFAULT 'COMPANY',
    contract_id            UUID,
    effective_date         DATE         NOT NULL DEFAULT CURRENT_DATE,
    is_active              BOOLEAN      NOT NULL DEFAULT true,
    group_approval_enabled BOOLEAN      NOT NULL DEFAULT false,
    require_team_approval   BOOLEAN     NOT NULL DEFAULT false,
    created_by             UUID         REFERENCES iam_users(id),
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by             UUID         REFERENCES iam_users(id),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_process_code UNIQUE (process_code),
    CONSTRAINT chk_approval_process_type CHECK (process_type IN (
        'INVESTMENT_ANALYSIS_REPORT','INVESTMENT_DECISION','INVESTMENT_CANCELLATION',
        'WORKFLOW_OPERATION','LEAVE_REQUEST','LEAVE_CANCELLATION','DELEGATION_REQUEST'
    )),
    CONSTRAINT chk_approval_contract_type CHECK (contract_type IN ('FUND','DISCRETIONARY','COMPANY'))
);

CREATE INDEX idx_approval_process_configs_resolve
    ON approval__process_configs (process_type, contract_id, effective_date)
    WHERE is_active = true;

CREATE INDEX idx_approval_process_configs_type
    ON approval__process_configs (process_type, contract_type, effective_date)
    WHERE is_active = true;

CREATE TRIGGER trg_approval_process_configs_updated_at
    BEFORE UPDATE ON approval__process_configs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────────────────────────────────────────
-- 7. Approval process stages (max 3 per config)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE approval__process_stages (
    id                      UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    process_config_id       UUID         NOT NULL REFERENCES approval__process_configs(id) ON DELETE CASCADE,
    stage_number            INTEGER      NOT NULL,
    stage_name              VARCHAR(160) NOT NULL DEFAULT '',
    approver_mode           VARCHAR(20)  NOT NULL,
    approver_user_id        UUID         REFERENCES iam_users(id),
    approval_group_id       UUID         REFERENCES approval__groups(id),
    required_approval_count INTEGER      NOT NULL DEFAULT 1,
    is_final_stage          BOOLEAN      NOT NULL DEFAULT false,
    reject_policy           VARCHAR(20)  NOT NULL DEFAULT 'STOP',
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_process_stage UNIQUE (process_config_id, stage_number),
    CONSTRAINT chk_approval_stage_number CHECK (stage_number BETWEEN 1 AND 3),
    CONSTRAINT chk_approval_stage_mode CHECK (approver_mode IN (
        'SINGLE_USER','GROUP_PRIORITY','GROUP_ANY','TEAM_MINIMUM'
    )),
    CONSTRAINT chk_approval_stage_reject_policy CHECK (reject_policy IN ('STOP')),
    CONSTRAINT chk_approval_stage_required_count CHECK (required_approval_count >= 1)
);

CREATE INDEX idx_approval_process_stages_config
    ON approval__process_stages (process_config_id, stage_number);

CREATE TRIGGER trg_approval_process_stages_updated_at
    BEFORE UPDATE ON approval__process_stages
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────────────────────────────────────────
-- 8. Approval requests (one per submitted business object instance)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE approval__requests (
    id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    request_number       VARCHAR(40)  NOT NULL,
    process_type         VARCHAR(40)  NOT NULL,
    process_config_id    UUID         REFERENCES approval__process_configs(id) ON DELETE RESTRICT,
    subject_type         VARCHAR(40)  NOT NULL,
    subject_id           UUID         NOT NULL,
    subject_title        VARCHAR(255) NOT NULL DEFAULT '',
    subject_reference    VARCHAR(120) NOT NULL DEFAULT '',
    contract_id          UUID,
    portfolio_id         UUID,
    submitter_id         UUID         NOT NULL REFERENCES iam_users(id) ON DELETE RESTRICT,
    submitted_at         TIMESTAMPTZ,
    current_stage_number INTEGER      NOT NULL DEFAULT 1,
    status               VARCHAR(20)  NOT NULL DEFAULT 'DRAFT',
    final_decision_by    UUID         REFERENCES iam_users(id),
    final_decision_at    TIMESTAMPTZ,
    rejection_reason     TEXT         NOT NULL DEFAULT '',
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_request_number UNIQUE (request_number),
    CONSTRAINT chk_approval_request_status CHECK (status IN (
        'DRAFT','SUBMITTED','PENDING_APPROVAL','APPROVED','REJECTED','CANCELLED','WITHDRAWN'
    ))
);

CREATE INDEX idx_approval_requests_subject
    ON approval__requests (subject_type, subject_id);

CREATE INDEX idx_approval_requests_status
    ON approval__requests (status, updated_at DESC);

CREATE INDEX idx_approval_requests_submitter
    ON approval__requests (submitter_id, created_at DESC);

-- Only one ACTIVE (not terminal) request per subject — prevents duplicate submit.
CREATE UNIQUE INDEX uq_approval_request_active_subject
    ON approval__requests (subject_type, subject_id)
    WHERE status IN ('DRAFT','SUBMITTED','PENDING_APPROVAL');

CREATE TRIGGER trg_approval_requests_updated_at
    BEFORE UPDATE ON approval__requests
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────────────────────────────────────────
-- 9. Approval tasks (per-approver work items)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE approval__tasks (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_request_id    UUID         NOT NULL REFERENCES approval__requests(id) ON DELETE CASCADE,
    stage_number           INTEGER      NOT NULL,
    assigned_user_id       UUID         REFERENCES iam_users(id),
    assigned_group_id      UUID         REFERENCES approval__groups(id),
    assigned_team_id       UUID         REFERENCES approval__teams(id),
    delegated_from_user_id UUID         REFERENCES iam_users(id),
    status                 VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    acted_by               UUID         REFERENCES iam_users(id),
    acted_at               TIMESTAMPTZ,
    action_comment         TEXT         NOT NULL DEFAULT '',
    is_delegated_action    BOOLEAN      NOT NULL DEFAULT false,
    due_at                 TIMESTAMPTZ,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_approval_task_status CHECK (status IN (
        'PENDING','APPROVED','REJECTED','SKIPPED','CANCELLED'
    )),
    CONSTRAINT chk_approval_task_stage CHECK (stage_number BETWEEN 1 AND 3)
);

CREATE INDEX idx_approval_tasks_request
    ON approval__tasks (approval_request_id, stage_number);

CREATE INDEX idx_approval_tasks_assignee
    ON approval__tasks (assigned_user_id, status);

CREATE TRIGGER trg_approval_tasks_updated_at
    BEFORE UPDATE ON approval__tasks
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────────────────────────────────────────
-- 10. Approval events (immutable timeline / audit trail)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE approval__events (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_request_id    UUID         NOT NULL REFERENCES approval__requests(id) ON DELETE CASCADE,
    event_type             VARCHAR(30)  NOT NULL,
    stage_number           INTEGER,
    actor_user_id          UUID         REFERENCES iam_users(id),
    delegated_from_user_id UUID         REFERENCES iam_users(id),
    comment                TEXT         NOT NULL DEFAULT '',
    metadata_json          JSONB        NOT NULL DEFAULT '{}',
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_approval_event_type CHECK (event_type IN (
        'SUBMITTED','TASK_CREATED','APPROVED','REJECTED','CANCELLED','WITHDRAWN',
        'DELEGATED','STAGE_COMPLETED','REQUEST_COMPLETED'
    ))
);

CREATE INDEX idx_approval_events_request
    ON approval__events (approval_request_id, created_at);

-- Immutable: approval events are an audit trail and must never be mutated.
CREATE OR REPLACE FUNCTION prevent_approval_events_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'approval__events are immutable and cannot be modified or deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_approval_events_no_update
    BEFORE UPDATE ON approval__events
    FOR EACH ROW EXECUTE FUNCTION prevent_approval_events_modification();

CREATE TRIGGER trg_approval_events_no_delete_guard
    BEFORE DELETE ON approval__events
    FOR EACH ROW EXECUTE FUNCTION prevent_approval_events_modification();

-- ─────────────────────────────────────────────────────────────────────────────
-- 11. Approval signature / stamp records
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE approval__signature_records (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_request_id UUID         NOT NULL REFERENCES approval__requests(id) ON DELETE CASCADE,
    stage_number        INTEGER      NOT NULL,
    signer_user_id      UUID         NOT NULL REFERENCES iam_users(id),
    signer_display_name VARCHAR(255) NOT NULL DEFAULT '',
    signer_title        VARCHAR(160),
    signed_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_proxy_signature  BOOLEAN      NOT NULL DEFAULT false,
    proxy_for_user_id   UUID         REFERENCES iam_users(id),
    signature_label     VARCHAR(20)  NOT NULL DEFAULT 'NORMAL',
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_approval_signature_label CHECK (signature_label IN ('NORMAL','DELEGATED'))
);

CREATE INDEX idx_approval_signature_records_request
    ON approval__signature_records (approval_request_id, stage_number);

-- =============================================================================
COMMENT ON TABLE approval__groups            IS 'Reusable approver groups for configurable approval chains.';
COMMENT ON TABLE approval__group_members     IS 'Members of approval groups; only APPROVED + active members are eligible approvers.';
COMMENT ON TABLE approval__teams             IS 'Per contract/fund approval teams with min/max stamp requirements.';
COMMENT ON TABLE approval__team_contracts    IS 'Active assignment of a team to a contract/fund (one active team per contract).';
COMMENT ON TABLE approval__team_members      IS 'Members of approval teams (ORDER_SUBMITTER / REVIEWER_AGENT).';
COMMENT ON TABLE approval__process_configs   IS 'Approval process configuration resolved by process_type + contract scope + effective_date.';
COMMENT ON TABLE approval__process_stages    IS 'Up to 3 ordered approval stages per process config.';
COMMENT ON TABLE approval__requests          IS 'Approval instance for one submitted business object (polymorphic subject).';
COMMENT ON TABLE approval__tasks             IS 'Per-approver pending work items for an approval request stage.';
COMMENT ON TABLE approval__events            IS 'Immutable, ordered approval timeline / audit trail.';
COMMENT ON TABLE approval__signature_records IS 'Digital stamp/signature records, including delegated (proxy) markers.';
