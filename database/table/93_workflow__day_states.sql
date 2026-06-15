-- Table: workflow__day_states
-- Source: 20260422000001_workflow__create_tables.up.sql
CREATE TABLE workflow__day_states (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

-- Aggregate key
contract_id UUID, business_date DATE NOT NULL,

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

    CONSTRAINT uq_wf_day_states_business_date
        UNIQUE (business_date),

    CONSTRAINT chk_wf_day_state
        CHECK (current_state IN (
            'NOT_STARTED',
            'DAY_OPEN',
            'INVESTMENT_DAY_STARTED',
            'MANAGER_APPROVED',
            'MANAGER_APPROVED_END_OF_DAY',
            'TRANSACTION_CLOSED',
            'ACCOUNTING_CLOSED'
        )),

    CONSTRAINT chk_wf_reclose_count
        CHECK (reclose_count >= 0 AND reclose_count <= 10),

    CONSTRAINT chk_wf_version_positive
        CHECK (version >= 1)
);
