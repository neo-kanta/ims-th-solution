-- Table: workflow__transition_log
-- Source: 20260422000001_workflow__create_tables.up.sql
CREATE TABLE workflow__transition_log (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

-- Parent reference (denormalized columns allow log queries without joining day_states)
workflow_day_id UUID NOT NULL REFERENCES workflow__day_states (id),
contract_id UUID,
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
            'NOT_STARTED', 'DAY_OPEN', 'INVESTMENT_DAY_STARTED',
            'MANAGER_APPROVED', 'MANAGER_APPROVED_END_OF_DAY',
            'TRANSACTION_CLOSED', 'ACCOUNTING_CLOSED'
        )),

    CONSTRAINT chk_wf_transition_to_state
        CHECK (to_state IN (
            'NOT_STARTED', 'DAY_OPEN', 'INVESTMENT_DAY_STARTED',
            'MANAGER_APPROVED', 'MANAGER_APPROVED_END_OF_DAY',
            'TRANSACTION_CLOSED', 'ACCOUNTING_CLOSED'
        ))
);
