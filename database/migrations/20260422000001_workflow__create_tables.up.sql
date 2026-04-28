-- =============================================================================
-- Workflow Management Module — Core Tables
-- =============================================================================
-- Aggregate key: (contract_id, business_date)
-- business_date is a DATE in Asia/Bangkok calendar semantics.
-- All TIMESTAMPTZ values are stored in UTC.
-- Transition log is append-only: no UPDATE or DELETE ever.
-- =============================================================================

-- ---------------------------------------------------------------------------
-- 1. Current-state snapshot: one row per (contract_id, business_date)
-- ---------------------------------------------------------------------------
CREATE TABLE workflow__day_states (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

-- Aggregate key
contract_id UUID NOT NULL, business_date DATE NOT NULL,

-- State machine position
current_state VARCHAR(30) NOT NULL DEFAULT 'NOT_STARTED',

-- Per-stage timestamps (NULL until that stage is reached)
opened_at TIMESTAMPTZ,
opened_by UUID,
manager_approved_at TIMESTAMPTZ,
manager_approved_by UUID,
transactions_locked_at TIMESTAMPTZ, -- set simultaneously with manager_approved_at
transaction_closed_at TIMESTAMPTZ,
transaction_closed_by UUID,
accounting_closed_at TIMESTAMPTZ,
accounting_closed_by UUID,

-- Rollback tracking (Accounting Closing)
pending_reclose BOOLEAN NOT NULL DEFAULT false,
reclose_count SMALLINT NOT NULL DEFAULT 0,

-- Optimistic locking
version INTEGER NOT NULL DEFAULT 1,

-- Standard audit columns (UTC)

created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by              UUID        NOT NULL,
    updated_by              UUID        NOT NULL,

    CONSTRAINT uq_wf_day_states_contract_date
        UNIQUE (contract_id, business_date),

    CONSTRAINT chk_wf_day_state
        CHECK (current_state IN (
            'NOT_STARTED',
            'DAY_OPEN',
            'MANAGER_APPROVED',
            'TRANSACTION_CLOSED',
            'ACCOUNTING_CLOSED'
        )),

    CONSTRAINT chk_wf_reclose_count
        CHECK (reclose_count >= 0 AND reclose_count <= 10),

    CONSTRAINT chk_wf_version_positive
        CHECK (version >= 1)
);

CREATE INDEX idx_wf_day_states_contract ON workflow__day_states (contract_id);

CREATE INDEX idx_wf_day_states_business_date ON workflow__day_states (business_date);

-- Hot-path index: "list all contracts in state X on date Y"
CREATE INDEX idx_wf_day_states_state_date ON workflow__day_states (current_state, business_date);

-- Partial index: fast scan for pending re-close (small cardinality, high priority)
CREATE INDEX idx_wf_day_states_pending_reclose ON workflow__day_states (contract_id, business_date)
WHERE
    pending_reclose = true;

-- ---------------------------------------------------------------------------
-- 2. Immutable transition log: append-only forensic trail
-- ---------------------------------------------------------------------------
CREATE TABLE workflow__transition_log (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

-- Parent reference (denormalized columns allow log queries without joining day_states)
workflow_day_id UUID NOT NULL REFERENCES workflow__day_states (id),
contract_id UUID NOT NULL,
business_date DATE NOT NULL,

-- State change
from_state VARCHAR(30) NOT NULL,
to_state VARCHAR(30) NOT NULL,
action VARCHAR(50) NOT NULL,

-- Actor
actor_id UUID, -- NULL for fully automated actions
actor_type VARCHAR(10) NOT NULL DEFAULT 'HUMAN',
actor_username VARCHAR(255) NOT NULL DEFAULT '',

-- Reason (mandatory for cancel/rollback actions; optional otherwise)
reason TEXT,

-- Extensible payload (IRG snapshot, attestation data, external refs, etc.)
metadata JSONB NOT NULL DEFAULT '{}',

-- Observability

occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    request_id      VARCHAR(255) NOT NULL DEFAULT '',

    CONSTRAINT chk_wf_transition_actor_type
        CHECK (actor_type IN ('HUMAN', 'SYSTEM')),

    CONSTRAINT chk_wf_transition_from_state
        CHECK (from_state IN (
            'NOT_STARTED', 'DAY_OPEN', 'MANAGER_APPROVED',
            'TRANSACTION_CLOSED', 'ACCOUNTING_CLOSED'
        )),

    CONSTRAINT chk_wf_transition_to_state
        CHECK (to_state IN (
            'NOT_STARTED', 'DAY_OPEN', 'MANAGER_APPROVED',
            'TRANSACTION_CLOSED', 'ACCOUNTING_CLOSED'
        ))
);

CREATE INDEX idx_wf_transition_log_workflow_day ON workflow__transition_log (workflow_day_id);

CREATE INDEX idx_wf_transition_log_contract_date ON workflow__transition_log (contract_id, business_date);

CREATE INDEX idx_wf_transition_log_occurred_at ON workflow__transition_log (occurred_at DESC);

-- ---------------------------------------------------------------------------
-- 3. Approval records: extensibility anchor for maker-checker (future)
-- ---------------------------------------------------------------------------
CREATE TABLE workflow__approval_records (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

-- Parent reference (denormalized for query independence)
workflow_day_id UUID NOT NULL REFERENCES workflow__day_states (id),
contract_id UUID NOT NULL,
business_date DATE NOT NULL,

-- Approver snapshot (captured at approval time; immutable)
approver_id UUID NOT NULL,
approver_username VARCHAR(255) NOT NULL DEFAULT '',
approver_role VARCHAR(50) NOT NULL DEFAULT '',

-- Approval lifecycle
approval_status VARCHAR(20) NOT NULL DEFAULT 'APPROVED',

-- Zero-transaction attestation (populated when approving with no trades)
is_zero_transaction BOOLEAN NOT NULL DEFAULT false,
attestation_reason TEXT, -- >= 30 chars when is_zero_transaction = true

-- Timestamps

approved_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at       TIMESTAMPTZ,        -- set by CANCEL_APPROVAL (Batch 2)
    revoked_by       UUID,               -- set by CANCEL_APPROVAL (Batch 2)

    notes            TEXT,

    CONSTRAINT chk_wf_approval_status
        CHECK (approval_status IN ('APPROVED', 'REVOKED'))
);

CREATE INDEX idx_wf_approval_records_workflow_day ON workflow__approval_records (workflow_day_id);

CREATE INDEX idx_wf_approval_records_contract_date ON workflow__approval_records (contract_id, business_date);