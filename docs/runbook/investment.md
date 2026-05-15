# Runbook — Investment Module

Operational guide for the IMS Investment module. Each scenario lists symptoms, the SQL or metric to confirm the diagnosis, and remediation steps.

> All `psql` snippets assume `\timing on` is acceptable in your session and that you are connected as the application user (`ims_app`). Replace the placeholder UUIDs.

---

## 1. Alpha Vantage provider down

### Symptoms

- Sustained spike in `market_data_ingest_total{provider="alpha_vantage", outcome="failed"}` and `outcome="rate_limited"`.
- `market_data_provider_latency_seconds{provider="alpha_vantage"}` shows wall-clock spikes or timeouts.
- Frontend portfolio valuation surfaces "stale price" warnings; counter `investment_valuation_stale_inputs_total` rising.

### Confirm

```bash
# 1. Did the provider return non-success in the last hour?
curl -fsS "$METRICS_URL/metrics" | grep -E 'market_data_ingest_total\{provider="alpha_vantage",outcome="(failed|rate_limited)"\}'

# 2. Is the outage upstream or in our keys?
curl -fsS "https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=IBM&apikey=$ALPHA_VANTAGE_API_KEY" | jq .
```

### Remediate

1. **Cap the blast radius.** Pause the scheduler so no further ingestion attempts pile up:
   ```sql
   UPDATE workflow_scheduler_settings SET is_paused = true WHERE id = 'market-data-ingest';
   ```
2. **Manual price post (operator-authorised only).** Use the `/investment/instruments/{id}/prices` endpoint with a fresh price you sourced from the broker:
   ```bash
   curl -X POST "$API/investment/instruments/$INSTRUMENT_ID/prices" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
              "business_date": "2026-05-06",
              "currency": "THB",
              "close_price": "157.25",
              "source": "MANUAL_BROKER_QUOTE"
            }'
   ```
   - Required permission: `INVESTMENT_PRICE_POST_MANUAL` (Phase 1).
   - Only a head-of-trading or risk-control approver may authorise.
   - Audit row is automatic; verify with section 6 below.
3. **Resume the scheduler** once the upstream is healthy:
   ```sql
   UPDATE workflow_scheduler_settings SET is_paused = false WHERE id = 'market-data-ingest';
   ```

### Escalation

Persistent (> 2h) Alpha Vantage outage → notify the Trading Operations on-call and engage the secondary provider plan (`docs/adr/00X-secondary-market-data-provider.md` — TBD).

---

## 2. Locked-day reversal

### When

A trade lands in `investment__ledger` for a portfolio whose business date is already `MANAGER_APPROVED` or later, and a reversal is required (e.g. trade error caught next day, bad cost basis).

### Confirm

```sql
-- Find the original transaction
SELECT id, transaction_type, business_date, posted_at, posted_by, reversed_by_id
FROM   investment__ledger
WHERE  portfolio_id = :portfolio_id
  AND  external_ref = :external_ref
ORDER  BY posted_at DESC;

-- Confirm the day is locked
SELECT current_state
FROM   workflow__day_states
WHERE  contract_id = :contract_id
  AND  business_date = :business_date;
-- Expected: MANAGER_APPROVED, TRANSACTION_CLOSED, or ACCOUNTING_CLOSED
```

### Remediate

1. **Force-post the reversal** with the dedicated permission. Phase 1 ships `POST /investment/portfolios/{id}/transactions/{txn_id}/reverse?force=true`:
   ```bash
   curl -X POST "$API/investment/portfolios/$PORTFOLIO_ID/transactions/$TXN_ID/reverse" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
              "reversal_business_date": "2026-05-07",
              "reason": "Trade error: wrong instrument identifier",
              "force": true
            }'
   ```
   - Required permission: `INVESTMENT_FORCE_POST` (Phase 1).
   - Reversal dating semantics:
     - **Cash leg** is dated `reversal_business_date` (today's open day).
     - **Position math** is unwound at the **original** business date — average-cost is reverted to its pre-trade state.
2. **Verify the reversal landed** correctly:
   ```sql
   -- The original transaction should reference the reversal.
   SELECT id, transaction_type, reversed_by_id, reversed_at
   FROM   investment__ledger
   WHERE  id = :original_txn_id;
   -- Expected: reversed_by_id is non-null, transaction_type unchanged.

   -- The reversal row should reference the original.
   SELECT id, transaction_type, reverses_id, business_date, reason
   FROM   investment__ledger
   WHERE  reverses_id = :original_txn_id;
   -- Expected: transaction_type = 'REVERSAL'; reverses_id matches.

   -- Position is back to pre-trade state at original date.
   SELECT business_date, quantity, avg_cost
   FROM   investment__positions
   WHERE  portfolio_id = :portfolio_id
     AND  instrument_id = :instrument_id
   ORDER  BY business_date DESC LIMIT 5;
   ```
3. **Audit trail check** — the force-post must produce a row keyed `actor_id = <head-of-trading>` and `action = 'INVESTMENT_FORCE_POST'`:
   ```sql
   SELECT created_at, actor_id, target_type, target_id, metadata
   FROM   audit__events
   WHERE  target_type = 'investment_ledger'
     AND  target_id   = :original_txn_id::text
   ORDER  BY created_at DESC;
   ```

### Escalation

Reversal lands but `investment_force_post_total{type="reversal"}` does NOT increment → metrics wiring regression. Escalate to Engineering.

---

## 3. Valuation looks wrong

### Symptoms

- A portfolio's NAV diverges from PAM (or from a manually-computed sum-of-positions).
- Operator pings asking "why is fund X's AUM off by $Y?"

### Confirm

```sql
-- Pull the snapshot's input hashes; a price/FX set hash that doesn't
-- match the underlying snapshots is the most common root cause.
SELECT v.id, v.business_date, v.nav_total,
       v.price_set_hash, v.fx_set_hash, v.run_id
FROM   investment__valuations v
WHERE  v.portfolio_id = :portfolio_id
ORDER  BY v.business_date DESC LIMIT 5;

-- Inspect the price snapshots used
SELECT ps.instrument_id, ps.business_date, ps.close_price, ps.currency, ps.source
FROM   investment__price_snapshots ps
WHERE  ps.business_date = :business_date
  AND  ps.instrument_id IN (
       SELECT instrument_id FROM investment__positions
        WHERE portfolio_id = :portfolio_id AND business_date = :business_date
       );

-- Inspect the FX snapshots
SELECT fr.from_ccy, fr.to_ccy, fr.business_date, fr.rate, fr.source
FROM   investment__fx_rates fr
WHERE  fr.business_date = :business_date;
```

### Remediate

1. If `price_set_hash` doesn't match a recompute over the same snapshots → bug; open Engineering ticket and capture hashes + run_id in the report.
2. If a snapshot is wrong (e.g. stale source) → manual correction via section 1's manual price-post path, then re-run valuation:
   ```bash
   curl -X POST "$API/investment/portfolios/$PORTFOLIO_ID/valuations/run" \
        -H "Authorization: Bearer $TOKEN" \
        -d '{"business_date": "2026-05-06"}'
   ```
3. If FX is wrong → wait for ingestion or post a manual FX rate (Phase 1: `POST /reference/fx-rates`).

### Escalation

Two consecutive runs produce different `price_set_hash` for the same `business_date` → ingestion is non-deterministic. Engineering must investigate.

---

## 4. Negative quantity tripped

### Symptoms

- 5xx from a post or reversal command with envelope `{ "error_code": "INTERNAL_ERROR", ... }`.
- Server log shows `negative quantity guard` or `position projector divergence`.
- `investment_projector_retry_total` is climbing.

### Confirm

```sql
-- Find the offending position row
SELECT portfolio_id, instrument_id, business_date, quantity, avg_cost, version
FROM   investment__positions
WHERE  quantity < 0
ORDER  BY business_date DESC;

-- Cross-check against the ledger
SELECT id, transaction_type, quantity, business_date, posted_at, version_at_post
FROM   investment__ledger
WHERE  portfolio_id = :portfolio_id
  AND  instrument_id = :instrument_id
ORDER  BY business_date DESC, posted_at DESC LIMIT 20;
```

### Diagnose

The defensive guard fires when the projector tries to write a row whose computed quantity is negative — almost always a B1-class race (concurrent posts updating the same `(portfolio_id, instrument_id, business_date)` row without optimistic locking). Confirm with:

```bash
curl -fsS "$METRICS_URL/metrics" | grep investment_projector_retry_total
# Healthy: small steady number, mostly zero. Above ~5/min sustained → race.
```

### Remediate

1. The projector is designed to retry. If retries succeed, no operator action is needed; investigate the latency bump.
2. If a position row is stuck negative:
   - Lock the portfolio's day-state to prevent further posts.
   - Reverse the offending transaction (section 2 above).
   - Re-post the corrected version.
   - Verify position returns to a non-negative invariant.

### Escalation

Sustained `investment_projector_retry_total` > 1/sec or any persistent negative position row → page on-call Engineering.

---

## 5. Permission denied where it shouldn't be

### Symptoms

- User reports `403` (envelope `error_code: "FORBIDDEN"` or similar) on a route they used to access.
- `WORKFLOW_TRADE_NOT_ALLOWED` or `WORKFLOW_LOCKED` envelope from a write that should have worked.

### Confirm

```sql
-- Effective function permissions for a user, including group inheritance.
SELECT g.name AS group_name, fr.permission_code, fr.is_granted
FROM   permissions_function_rights fr
JOIN   permissions_accounts_groups ag ON ag.group_id = fr.group_id
JOIN   permissions_groups g ON g.id = fr.group_id
WHERE  ag.user_id = :user_id
  AND  fr.is_granted = true
ORDER  BY fr.permission_code;

-- Effective data scopes (contracts the user can see)
SELECT contract_id, is_granted, granted_at, granted_by
FROM   permissions_data_rights
WHERE  user_id = :user_id
ORDER  BY contract_id;

-- Workflow state for the contract+date the user is hitting
SELECT current_state, business_date, contract_id
FROM   workflow__day_states
WHERE  contract_id = :contract_id
  AND  business_date = :business_date;
```

### Remediate

1. Permission missing → grant via the permissions module; ensure the code exists in `permissions_function_definitions` first (`SELECT * FROM permissions_function_definitions WHERE code = '<CODE>';`).
2. Data scope missing → the user needs `permissions_data_rights` for the contract or fund.
3. Workflow gate (e.g., `WORKFLOW_LOCKED`) → the day is past `MANAGER_APPROVED`. Use the locked-day reversal path (section 2) or wait for the next business day.

### Escalation

User has the right grant in DB but the API still returns 403 → JWT or session-cache staleness. Have the user log out / log back in. If reproducible, escalate Engineering.

---

## 6. Audit query patterns

### Every actor on a portfolio in a date range

```sql
SELECT DISTINCT actor_id, MIN(created_at) AS first_seen, MAX(created_at) AS last_seen
FROM   audit__events
WHERE  target_type IN ('investment_portfolio', 'investment_ledger', 'investment_position')
  AND  metadata->>'portfolio_id' = :portfolio_id::text
  AND  created_at BETWEEN :start_ts AND :end_ts
GROUP  BY actor_id
ORDER  BY last_seen DESC;
```

### Every force-post in the last week

```sql
SELECT created_at, actor_id, target_type, target_id, metadata->>'reason' AS reason
FROM   audit__events
WHERE  event_type = 'INVESTMENT_FORCE_POST'
  AND  created_at >= NOW() - INTERVAL '7 days'
ORDER  BY created_at DESC;
```

### Every reversal and its original

```sql
SELECT r.id   AS reversal_id,
       r.business_date AS reversal_date,
       r.reverses_id   AS original_id,
       o.business_date AS original_date,
       r.posted_by     AS reversal_actor,
       r.reason
FROM   investment__ledger r
JOIN   investment__ledger o ON o.id = r.reverses_id
WHERE  r.transaction_type = 'REVERSAL'
ORDER  BY r.posted_at DESC;
```

---

## 7. New instrument onboarding

### Steps

1. **Create the instrument:**
   ```bash
   curl -X POST "$API/investment/instruments" \
        -H "Authorization: Bearer $TOKEN" \
        -d '{
              "ticker": "AOT",
              "name": "Airports of Thailand",
              "asset_class": "EQUITY",
              "asset_subtype": "COMMON_STOCK",
              "primary_exchange": "SET",
              "currency": "THB",
              "sector_code": "INDUSTRIALS",
              "country_code": "TH"
            }'
   ```
   - Required permission: `INVESTMENT_INSTRUMENT_CREATE` (Phase 1).
   - The reference codes (`asset_class`, `asset_subtype`, `sector_code`, `country_code`) MUST exist in their reference tables — see `database/seeds/investment/`.
2. **Add provider mapping** so the next ingestion run knows where to find prices:
   ```sql
   INSERT INTO investment__instrument_identifiers
       (instrument_id, id_type, id_value, source)
   VALUES
       (:instrument_id, 'PROVIDER', 'AOT.BK', 'alpha_vantage');
   ```
   - `id_type = 'PROVIDER'` is the agreed key the ingestion job filters on.
   - Optional: also add ISIN (`id_type='ISIN'`) and any house code.
3. **Wait for the next ingestion run** — by default the scheduler triggers once per business day. To trigger manually:
   ```bash
   curl -X POST "$API/market-data/ingest/run" \
        -H "Authorization: Bearer $TOKEN" \
        -d '{"only_instrument_ids": ["'"$INSTRUMENT_ID"'"]}'
   ```
4. **Confirm a price snapshot landed:**
   ```sql
   SELECT business_date, close_price, currency, source, ingested_at
   FROM   investment__price_snapshots
   WHERE  instrument_id = :instrument_id
   ORDER  BY business_date DESC LIMIT 3;
   ```

### Common pitfalls

- Wrong `currency` on the instrument vs the price snapshot → posts will fail with `PRICE_CURRENCY_MISMATCH`. Fix on the instrument record before re-ingesting.
- Provider symbol typo → `INSTRUMENT_NOT_MAPPED`. Update `investment__instrument_identifiers` and re-trigger ingestion.

---

## Quick links

- Metrics endpoint: `GET /metrics` (Prometheus exposition format).
- Health: `GET /health/live` (process up), `GET /health/ready` (DB + dependencies).
- Error envelope: `{ "error_code", "message", "details", "request_id" }` — every typed error in `pkg/errcode/codes.go`.
- Audit query API: `GET /admin/audit?...` (requires `IAM_AUDIT_VIEW`).
