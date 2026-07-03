-- Table: investment__process_steps
-- Source: 20260427000002_investment__create_process_assignment_tables.up.sql
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
