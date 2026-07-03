-- =============================================================================
-- Investment module — permission seed
-- =============================================================================
-- Seeds INVESTMENT_* function permission codes against the Admin group and a
-- new "Investment Operator" group. Idempotent.
-- =============================================================================

BEGIN;

-- Grant the full investment permission set to Admin.
INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT g.id, p.code, true, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
CROSS JOIN (VALUES
    ('INVESTMENT_FUND_VIEW'),
    ('INVESTMENT_FUND_MANAGE'),
    ('INVESTMENT_PORTFOLIO_VIEW'),
    ('INVESTMENT_PORTFOLIO_MANAGE'),
    ('INVESTMENT_INSTRUMENT_VIEW'),
    ('INVESTMENT_INSTRUMENT_MANAGE'),
    ('INVESTMENT_REFERENCE_VIEW'),
    ('INVESTMENT_LEDGER_VIEW'),
    ('INVESTMENT_LEDGER_POST'),
    ('INVESTMENT_LEDGER_SIMULATE'),
    ('INVESTMENT_LEDGER_FORCE_POST'),
    ('INVESTMENT_LEDGER_REVERSE'),
    ('INVESTMENT_VALUATION_VIEW'),
    ('INVESTMENT_VALUATION_RUN'),
    ('INVESTMENT_PRICE_POST'),
    ('INVESTMENT_RESEARCH_VIEW'),
    ('INVESTMENT_RESEARCH_CREATE'),
    ('INVESTMENT_RESEARCH_UPDATE'),
    ('INVESTMENT_RESEARCH_DELETE'),
    ('INVESTMENT_RESEARCH_SUBMIT'),
    ('INVESTMENT_RESEARCH_CANCEL_SUBMIT')
) AS p(code)
WHERE g.name = 'Admin'
ON CONFLICT (group_id, permission_code) DO UPDATE
SET is_granted = EXCLUDED.is_granted,
    created_by = COALESCE(permissions_function_rights.created_by, EXCLUDED.created_by);

-- A non-admin operator group with read + post-only rights (no force-post).
INSERT INTO permissions_groups (id, name, description, is_active, created_by, updated_by)
VALUES
    (
        'b0000000-0000-0000-0000-000000000020',
        'Investment Operator',
        'Operators who manage portfolios, instruments, and post simulated transactions.',
        true,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    )
ON CONFLICT ON CONSTRAINT uq_permissions_groups_name DO UPDATE
SET description = EXCLUDED.description,
    is_active   = EXCLUDED.is_active,
    updated_by  = EXCLUDED.updated_by,
    updated_at  = NOW();

INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT g.id, p.code, true, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
CROSS JOIN (VALUES
    ('INVESTMENT_FUND_VIEW'),
    ('INVESTMENT_PORTFOLIO_VIEW'),
    ('INVESTMENT_PORTFOLIO_MANAGE'),
    ('INVESTMENT_INSTRUMENT_VIEW'),
    ('INVESTMENT_REFERENCE_VIEW'),
    ('INVESTMENT_LEDGER_VIEW'),
    ('INVESTMENT_LEDGER_POST'),
    ('INVESTMENT_LEDGER_SIMULATE'),
    ('INVESTMENT_LEDGER_REVERSE'),
    ('INVESTMENT_VALUATION_VIEW'),
    ('INVESTMENT_VALUATION_RUN'),
    ('INVESTMENT_RESEARCH_VIEW'),
    ('INVESTMENT_RESEARCH_CREATE'),
    ('INVESTMENT_RESEARCH_UPDATE'),
    ('INVESTMENT_RESEARCH_SUBMIT'),
    ('INVESTMENT_RESEARCH_CANCEL_SUBMIT')
) AS p(code)
WHERE g.name = 'Investment Operator'
ON CONFLICT (group_id, permission_code) DO UPDATE
SET is_granted = EXCLUDED.is_granted;

COMMIT;
