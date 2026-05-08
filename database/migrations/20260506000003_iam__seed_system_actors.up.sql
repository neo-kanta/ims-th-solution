-- =============================================================================
-- IAM Module - System Actors (Phase 0)
-- =============================================================================
-- Seeds two stable, never-login system users used as the actor on records
-- written by automated subsystems. Both rows are immutable identity anchors:
--   - system.market_data : Phase 5 ingestion writer
--   - system.scheduler   : ValuationRunner and other scheduled jobs
--
-- Security:
--   - password_hash is the literal sentinel '!system' which cannot match any
--     bcrypt-verified credential. Login flows additionally reject usernames
--     that begin with 'system.'.
--   - is_active stays true so created_by / updated_by FKs resolve, but the
--     accounts MUST NOT be granted permission groups; their authority comes
--     from the calling subsystem, not from a permission grant.
--
-- IDs are reserved as the all-zeros UUID with the trailing octet incremented:
--   00000000-0000-0000-0000-000000000001 = system.market_data
--   00000000-0000-0000-0000-000000000002 = system.scheduler
-- =============================================================================

INSERT INTO iam_users (
    id,
    username,
    display_name,
    email,
    password_hash,
    is_active,
    force_password_change,
    created_by,
    updated_by
)
VALUES
    (
        '00000000-0000-0000-0000-000000000001',
        'system.market_data',
        'System: Market Data Ingestion',
        NULL,
        '!system',
        true,
        false,
        NULL,
        NULL
    ),
    (
        '00000000-0000-0000-0000-000000000002',
        'system.scheduler',
        'System: Scheduler',
        NULL,
        '!system',
        true,
        false,
        NULL,
        NULL
    )
ON CONFLICT ON CONSTRAINT uq_iam_users_username DO NOTHING;
