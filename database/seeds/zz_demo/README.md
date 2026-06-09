# Demo seed — `zz_demo/`

End-to-end demo dataset that brings the **My Funds** cockpit to life.
Loaded last by `make seed` because the `zz_` prefix sorts after every other
folder under `database/seeds/`, so reference data is in place before this
seed runs.

## What it creates

| File | Adds |
| ---- | ---- |
| `01_demo_funds.sql`              | 6 funds, 7 portfolios, 15 instruments (Thai blue-chips, US tech, govt + corp bonds, MMF) |
| `02_demo_positions_cash.sql`     | Positions and cash balances for every portfolio (numbers reconcile to the AUMs in 03) |
| `03_demo_prices_valuations.sql`  | Latest price snapshots, valuation snapshots, NAV snapshots, and AUM snapshots |
| `04_demo_workflow_state.sql`     | Today's workflow `day_states` + transition log so funds show varied stages |
| `05_demo_compliance_breaches.sql`| 1 BLOCK + 3 WARN open breaches (matches the cockpit "1 / 3 warnings" KPI) |
| `06_demo_data_permissions.sql`   | Per-user data scopes so admin/ben/green/neo see different fund subsets |

## Demo funds

| Code         | Name                                  | Asset class    | Status | Manager | Workflow today    | Notable signal              |
| ------------ | ------------------------------------- | -------------- | ------ | ------- | ----------------- | --------------------------- |
| TH-GOV-LTF   | Thai Government LTF — Alpha series    | Fixed income   | ACTIVE | admin   | DAY_OPEN          | NAV-per-unit; LOW risk      |
| BBL-EQUITY   | Large-cap Momentum — Equity           | Equity         | ACTIVE | admin   | MANAGER_APPROVED  | 1 BLOCK + 1 WARN breach     |
| SCB-FIXED    | Corporate Bonds 2026 — Fixed Income   | Fixed income   | ACTIVE | ben     | ACCOUNTING_CLOSED | Locked for the day          |
| KTB-BALANCED | Balanced — Quarterly Rebalance Q2     | Mixed          | ACTIVE | admin   | TRANSACTION_CLOSED| 1 WARN breach, 2 sleeves    |
| GLOBAL-TECH  | Global Tech Thematic                  | Equity (USD)   | ACTIVE | green   | DAY_OPEN          | Stale valuation (>72h)      |
| MMF-CASH     | SCB Money Market Fund                 | Cash           | ACTIVE | admin   | (no day state)    | NAV-per-unit; "Not started" |

## Demo accounts

All four users share the dev password `admin123` and force-password-change:

- `admin`  — wildcard data scope (sees all 6 funds)
- `ben`    — SCB-FIXED + TH-GOV-LTF
- `green`  — GLOBAL-TECH + BBL-EQUITY
- `neo`    — KTB-BALANCED + GLOBAL-TECH

## Idempotency

Every row uses a stable UUID (`d0001000-...` for funds, `d0002000-...` for
portfolios, etc.). Mutable tables `UPSERT` via natural unique constraints;
append-only tables (`price_snapshots`, `valuation_snapshots`, `nav_snapshots`,
`aum_snapshots`, `transition_log`, `check_records`, `breaches`) use
`ON CONFLICT (id) DO NOTHING` so re-running `make seed` is safe and never
mutates an immutable row.

## Refreshing the demo

```bash
make seed
```

To wipe the demo dataset entirely:

```sql
DELETE FROM compliance_breaches         WHERE id::text LIKE 'd000d000-%';
DELETE FROM compliance_check_records    WHERE id::text LIKE 'd000c000-%';
DELETE FROM workflow__transition_log    WHERE id::text LIKE 'd000b000-%';
DELETE FROM workflow__day_states        WHERE id::text LIKE 'd000a000-%';
DELETE FROM investment__aum_snapshots   WHERE id::text LIKE 'd0009000-%';
DELETE FROM investment__nav_snapshots   WHERE id::text LIKE 'd0008000-%';
DELETE FROM investment__valuation_snapshots WHERE id::text LIKE 'd0007000-%';
DELETE FROM investment__price_snapshots WHERE id::text LIKE 'd0006000-%';
DELETE FROM investment__cash_balances   WHERE id::text LIKE 'd0005000-%';
DELETE FROM investment__portfolio_positions WHERE id::text LIKE 'd0004000-%';
DELETE FROM investment__instruments     WHERE id::text LIKE 'd0003000-%';
DELETE FROM investment__portfolios      WHERE id::text LIKE 'd0002000-%';
DELETE FROM investment__funds           WHERE id::text LIKE 'd0001000-%';
DELETE FROM permissions_data_rights     WHERE id::text LIKE 'd000e000-%';
```

(Order matters because of FK constraints.) The append-only tables and most
ledger tables are DB-protected against `DELETE`; you'll need to drop the RULE
or recreate the DB if you really want a clean slate.
