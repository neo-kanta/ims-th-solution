-- =============================================================================
-- Audit query helper view — financial reconstruction
-- =============================================================================
-- iam_audit_events is intentionally narrow (actor, event_type, target, JSONB
-- metadata). Financial-action emits stash business_date, fund/contract IDs,
-- before/after snapshots, and delegation context inside the JSONB blob. This
-- view extracts the most-queried fields into typed columns so auditors and
-- the in-app investigation UI can filter by business date / fund without
-- writing jsonb_extract_path_text every time.
--
-- The view is READ-ONLY. Promoting these to real columns is a future
-- migration; this is the cheaper, non-destructive step that satisfies the
-- "one business-day reconstruction" requirement today.
-- =============================================================================

CREATE OR REPLACE VIEW audit_events_financial_v AS
SELECT
    ae.id,
    ae.actor_id,
    ae.event_type,
    ae.target_type,
    ae.target_id,
    ae.ip_address,
    ae.user_agent,
    ae.created_at,
    -- Business date stored by the investment adapter as YYYY-MM-DD text.
    NULLIF(ae.metadata ->> 'business_date', '')::DATE                        AS business_date,
    -- module / resource fields surface the entry's owning subsystem.
    NULLIF(ae.metadata ->> 'module', '')                                     AS module,
    NULLIF(ae.metadata ->> 'resource_type', '')                              AS resource_type,
    NULLIF(ae.metadata ->> 'resource_id', '')                                AS resource_id,
    -- Fund/contract/portfolio ids when investment commands populate them
    -- inside details. The COALESCE chain tolerates either top-level metadata
    -- keys or nested details keys depending on emit site.
    COALESCE(
        NULLIF(ae.metadata #>> '{details,fund_id}', ''),
        NULLIF(ae.metadata ->> 'fund_id', '')
    )::UUID                                                                  AS fund_id,
    COALESCE(
        NULLIF(ae.metadata #>> '{details,contract_id}', ''),
        NULLIF(ae.metadata ->> 'contract_id', '')
    )::UUID                                                                  AS contract_id,
    COALESCE(
        NULLIF(ae.metadata #>> '{details,portfolio_id}', ''),
        NULLIF(ae.metadata ->> 'portfolio_id', '')
    )::UUID                                                                  AS portfolio_id,
    -- Before/after snapshots and delegation principal extracted as JSONB so
    -- consumers can drill in without parsing strings.
    ae.metadata #> '{details,before}'                                        AS before_value,
    ae.metadata #> '{details,after}'                                         AS after_value,
    COALESCE(
        NULLIF(ae.metadata #>> '{details,delegation_principal_id}', ''),
        NULLIF(ae.metadata ->> 'delegation_principal_id', '')
    )::UUID                                                                  AS delegation_principal_id,
    ae.metadata                                                              AS raw_metadata
FROM iam_audit_events ae;

COMMENT ON VIEW audit_events_financial_v IS
    'Read-only view of iam_audit_events with business_date, contract/fund/portfolio IDs, before/after snapshots and delegation principal extracted from metadata JSONB. Use for fund-day reconstruction queries until typed columns ship.';
