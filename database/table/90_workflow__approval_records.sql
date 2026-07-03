-- Table: workflow__approval_records
-- Source: 20260422000001_workflow__create_tables.up.sql
CREATE TABLE workflow__approval_records (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

-- Parent reference (denormalized for query independence)
workflow_day_id UUID NOT NULL REFERENCES workflow__day_states (id),
contract_id UUID,
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
