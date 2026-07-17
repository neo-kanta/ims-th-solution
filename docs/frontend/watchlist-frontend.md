# Watchlist Frontend Implementation Guide

**Status:** frontend implementation handoff. Backend Watchlist is implemented, but the Nuxt frontend still needs to replace the current static mock page.

**Target route:** `/watchlist`

**Backend base path:** `/api/v1/watchlists`

## 1. Purpose And Scope

The Watchlist frontend lets authenticated users monitor canonical securities with market-price threshold rules and review generated alert events.

v1 supports only:

- `PERSONAL` watchlist items owned by the authenticated user.
- `PORTFOLIO` watchlist items attached to one portfolio.
- `MARKET_PRICE` thresholds.
- `ABOVE` and `BELOW` crossing rules.
- Alert listing and acknowledgement.

The frontend must not present unsupported v1 concepts. Do not build NAV, AUM, valuation, risk, or portfolio-performance thresholds. Do not describe watchlist prices as NAV. Decimal price and threshold values are string decimals at the API boundary, not JSON numbers.

The real Watchlist domain API is `/api/v1/watchlists`. Do not use `/api/v1/market-data/screen/watchlist` for item management, threshold rule management, alert listing, acknowledgement, or evaluation. That market-data path is only a Market Data screen feed of canonical securities joined to quote snapshots.

## 2. Source Of Truth

Use these files as the source of truth while implementing:

- `docs/handoff/watchlist-backend.md`: current backend handoff and implementation notes.
- `docs/api/watchlist-api.md`: full Watchlist API contract and examples.
- `docs/architecture/watchlist-backend-design.md`: domain boundaries, permission model, and alert semantics.
- `backend/internal/watchlist/transport/http/router.go`: mounted routes and route-level permissions.
- `backend/internal/watchlist/transport/http/dto.go`: response and request DTO names and JSON fields.
- `backend/internal/watchlist/transport/http/handler.go`: query parsing, default pagination, default include flags, and error codes.

When documents and code disagree, prefer the implemented backend code and then update the API docs in a separate backend documentation task.

## 3. Current Frontend State And Gaps

Current state:

- `frontend/app/pages/watchlist/index.vue` is a static mock. It already uses `layout: "dashboard"`, `middleware: ["auth", "permission"]`, and `meta: { permission: "WATCHLIST_VIEW" }`, but the body is hardcoded.
- The mock uses IDs like `WL-001`, NAV wording, hardcoded currency symbols, raw tables, and local arrays. Replace all of that.
- `frontend/app/pages/market-data/index.vue` correctly delegates to `frontend/app/features/market-data/MarketDataPage.vue`.
- `frontend/app/features/market-data/composables/useMarketData.ts` intentionally calls `GET /market-data/screen/watchlist` for the Market Data tab. That code is not the Watchlist feature and should not be reused for domain watchlists.
- `frontend/app/api/ims-api.d.ts` currently exposes `/market-data/screen/watchlist` and `ScreenWatchlistResponseDTO`, but it does not expose `/watchlists` paths or Watchlist response schemas.

Primary frontend gaps:

- Regenerate the OpenAPI client after backend Swagger exposes Watchlist paths and concrete Watchlist DTO schemas.
- Confirm the generated response types are usable. Path generation alone is not enough if Swagger still documents Watchlist success responses as generic `httputil.SuccessResponse`.
- Replace the static route with a thin route shell.
- Add feature-owned Watchlist code under `frontend/app/features/watchlist/`.
- Add navigation and i18n keys.
- Add API, formatter, permission, and component tests.

## 4. Target File Layout

Replace the route file:

```text
frontend/app/pages/watchlist/index.vue
```

Add this redirect-only route if the backend notification action URL remains `/watchlists/alerts/{id}`:

```text
frontend/app/pages/watchlists/alerts/[id].vue
```

Add feature-owned files:

```text
frontend/app/features/watchlist/
  WatchlistPage.vue
  services/
    watchlistApi.ts
  composables/
    useWatchlist.ts
    useWatchlistAlerts.ts
  types.ts
  lib/
    formatters.ts
  components/
    WatchlistItemsTable.vue
    WatchlistFilters.vue
    WatchlistItemDrawer.vue
    ThresholdRuleForm.vue
    WatchlistAlertsPanel.vue
    AcknowledgeAlertDialog.vue
    WatchlistScopeSelector.vue
```

Add i18n message modules:

```text
frontend/app/shared/i18n/messages/en/watchlist.ts
frontend/app/shared/i18n/messages/th/watchlist.ts
frontend/app/shared/i18n/messages/zh/watchlist.ts
```

Register each locale module in:

```text
frontend/app/shared/i18n/messages/en/index.ts
frontend/app/shared/i18n/messages/th/index.ts
frontend/app/shared/i18n/messages/zh/index.ts
```

Important i18n key placement:

- Put Watchlist feature copy under a top-level `watchlist` object in the new `watchlist.ts` files.
- Put the sidebar label `navigation.watchlist` in each locale's existing `common.ts` `navigation` object, because the locale index files merge message modules with shallow object spreads.
- Do not export a top-level `navigation` object from `watchlist.ts` unless the i18n merge strategy is changed first; it can overwrite existing navigation keys.

Suggested tests:

```text
frontend/tests/watchlist-api.test.ts
frontend/tests/watchlist-formatters.test.ts
frontend/tests/watchlist-permissions.test.ts
frontend/tests/watchlist-components.test.ts
```

## 5. Route, Navigation, And Permissions

### Route Shell

`frontend/app/pages/watchlist/index.vue` should be a thin Nuxt route shell:

```vue
<script setup lang="ts">
import WatchlistPage from "~/features/watchlist/WatchlistPage.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  meta: { permission: "WATCHLIST_VIEW" },
});
</script>

<template>
  <WatchlistPage />
</template>
```

Keep page state, API calls, and UI logic out of the route file.

### Notification Deep Links

Backend Watchlist notifications currently create action URLs shaped like:

```text
/watchlists/alerts/{alert_event_id}
```

The target frontend page is `/watchlist`, not `/watchlists/alerts/{id}`. To prevent notification links from 404ing, do one of these before shipping Watchlist notifications:

- Preferred frontend-compatible option: update the backend notifier to emit `/watchlist?alert_id={alert_event_id}`.
- If the backend URL is not changed, add a frontend route shim at `frontend/app/pages/watchlists/alerts/[id].vue` that requires `WATCHLIST_VIEW` and redirects to `/watchlist?alert_id={id}`.

Do not build a second full Watchlist screen at `/watchlists/alerts/[id]`.

`WatchlistPage.vue` should read `route.query.alert_id` and attempt to select or highlight the matching alert after alerts load. The current Watchlist API has list and acknowledge endpoints, but no single-alert detail endpoint or alert-ID filter. If exact deep linking to any historical alert is required, request a backend `GET /watchlists/alerts/{id}` endpoint or an alert-ID filter before relying on the list page alone. If the alert is not present in the loaded list, show a localized "alert not found or no longer visible" state without rendering the raw alert UUID.

### Sidebar Navigation

Add a sidebar item in `frontend/app/features/shell/navigation.ts`. The current navigation structure has an Investment section, so place it there unless product decides to create a separate Monitoring section.

Recommended item:

```ts
{
  label: t("navigation.watchlist", "Watchlist"),
  to: "/watchlist",
  icon: "notifications",
  requiredPermissions: ["WATCHLIST_VIEW"],
}
```

The navigation builder already filters items when none of the required permissions are present.

Add `watchlist: "Watchlist"` to `navigation` in each locale's `common.ts` so the `navigation.watchlist` key does not overwrite the existing navigation object.

### Optional Header Tabs

Do not add top header tabs unless the Watchlist feature actually needs them. If it later needs dashboard header tabs, follow `docs/frontend/dashboard-tab-registry-refactor.md`:

```text
frontend/app/features/watchlist/navigation/dashboardTabs.ts
frontend/app/features/shell/tabs/dashboardTabRegistry.ts
```

Register a `watchlistDashboardTabs` provider in the registry. Provider `setup()` must remain eager through `useDashboardHeaderTabs`; do not make provider setup lazy inside computed values.

### Permission Model

Frontend permission gates are UX only. The backend still enforces every permission.

| Capability | Permission | Frontend behavior |
| --- | --- | --- |
| View page, list items, list alerts | `WATCHLIST_VIEW` | Required route meta and sidebar visibility. |
| Create/update/delete items and threshold rules | `WATCHLIST_MANAGE` | Show Add, Edit, Disable, Delete, and Save controls only when present. |
| Acknowledge alerts | `WATCHLIST_ALERT_ACK` | Show acknowledge actions only when present and alert is unacknowledged. |
| Manual evaluation | `WATCHLIST_EVALUATE` | Hide from normal users. Only expose in an explicitly operational/admin surface. |
| Future admin override | `WATCHLIST_ADMIN` | Optional future behavior only; do not assume cross-owner access in v1. |

Use `useAuthStore().hasPermission(code)` or `IMSPermissionGuard` for component-level gating. Never rely on disabled buttons alone for authorization-sensitive actions.

## 6. API Service Contract

Create:

```text
frontend/app/features/watchlist/services/watchlistApi.ts
```

Use the existing typed OpenAPI pattern:

```ts
import {
  assertOpenApiResponse,
  unwrapOpenApiResponse,
  useOpenApiClient,
} from "~/api/openapi";
import type { paths, components } from "~/api/ims-api";
```

Use `unwrapOpenApiResponse()` for endpoints returning an envelope with data. `DELETE /watchlists/items/{id}` returns `204 No Content`, so validate it with `assertOpenApiResponse()` or an equivalent success check rather than unwrapping data.

### Generated-Type Prerequisite

Before implementing the service, the backend Swagger output must include these paths:

- `GET /watchlists`
- `POST /watchlists/items`
- `PATCH /watchlists/items/{id}`
- `DELETE /watchlists/items/{id}`
- `GET /watchlists/alerts`
- `POST /watchlists/alerts/{id}/acknowledge`
- `POST /watchlists/evaluate`

Path presence is not enough. The generated client must also expose concrete Watchlist request and response schemas from `backend/internal/watchlist/transport/http/dto.go`, including:

- `ListItemsResponseData`
- `ListAlertsResponseData`
- `WatchlistItemResponse`
- `ThresholdRuleResponse`
- `AlertEventResponse`
- `AcknowledgeAlertRequest`
- `CreateItemRequest`
- `UpdateItemRequest`
- `EvaluateRequest`
- `EvaluateResponseData`

Current backend Swagger annotations may describe Watchlist success responses as generic `httputil.SuccessResponse`. Because `SuccessResponse.Data` is `interface{}`, `make api-client` can add `/watchlists` paths while still producing response types that do not expose `ListItemsResponseData`, `WatchlistItemResponse`, or `AlertEventResponse`. The frontend implementation should not start from those unusable generic response types.

Backend prerequisite:

- Tighten Watchlist Swagger annotations or schema models so generated OpenAPI includes concrete `data` payload shapes for each success response.
- Keep `DELETE /watchlists/items/{id}` as `204 No Content`.
- Verify `frontend/app/api/ims-api.d.ts` can reference the concrete Watchlist schemas through `components["schemas"]` or through operation response types.

Current generated file `frontend/app/api/ims-api.d.ts` does not expose those Watchlist paths. From the repository root, run:

```bash
make api-client
```

Only implement the feature service after `frontend/app/api/ims-api.d.ts` exposes `/watchlists` and concrete Watchlist DTOs. Avoid hand-rolled API types unless the team explicitly accepts a temporary bridge. If a temporary bridge is approved, keep it in `frontend/app/features/watchlist/types.ts`, mirror `dto.go` exactly, document the bridge in that file, and remove it after backend Swagger plus `make api-client` provides generated types.

### Service Methods

The service should wrap exactly these domain endpoints:

| Method | Endpoint | Permission | Returns |
| --- | --- | --- | --- |
| `listItems(query)` | `GET /watchlists` | `WATCHLIST_VIEW` | `ListItemsResponseData` |
| `createItem(body)` | `POST /watchlists/items` | `WATCHLIST_MANAGE` | `WatchlistItemResponse` |
| `updateItem(id, body)` | `PATCH /watchlists/items/{id}` | `WATCHLIST_MANAGE` | `WatchlistItemResponse` |
| `deleteItem(id)` | `DELETE /watchlists/items/{id}` | `WATCHLIST_MANAGE` | `void` |
| `listAlerts(query)` | `GET /watchlists/alerts` | `WATCHLIST_VIEW` | `ListAlertsResponseData` |
| `acknowledgeAlert(id, body)` | `POST /watchlists/alerts/{id}/acknowledge` | `WATCHLIST_ALERT_ACK` | `AlertEventResponse` |
| `manualEvaluate(body)` | `POST /watchlists/evaluate` | `WATCHLIST_EVALUATE` | `EvaluateResponseData` |

Do not add methods that call `/market-data/screen/watchlist`.

### Item List Query

`GET /watchlists` supports:

- `scope_type?: "PERSONAL" | "PORTFOLIO"`
- `portfolio_id?: string`
- `security_id?: string`
- `include_disabled?: boolean`, default `false`
- `include_thresholds?: boolean`, default `true`
- `include_quote?: boolean`, default `true`
- `limit?: number`, default `50`, max `200`
- `offset?: number`, default `0`

If `portfolio_id` is supplied, `scope_type` must be `PORTFOLIO`.

### Alert List Query

`GET /watchlists/alerts` supports:

- `scope_type?: "PERSONAL" | "PORTFOLIO"`
- `portfolio_id?: string`
- `security_id?: string`
- `rule_id?: string`
- `acknowledged?: boolean`
- `created_from?: string`, RFC3339
- `created_to?: string`, RFC3339
- `limit?: number`, default `50`, max `200`
- `offset?: number`, default `0`

If `portfolio_id` is supplied, `scope_type` must be `PORTFOLIO`.

### Create And Update Bodies

`POST /watchlists/items` body:

```json
{
  "scope_type": "PERSONAL",
  "portfolio_id": null,
  "security_id": "canonical-security-uuid",
  "pinned": true,
  "note": "Monitor upside breakout",
  "threshold_rules": [
    {
      "metric_type": "MARKET_PRICE",
      "direction": "ABOVE",
      "threshold_value": "190.00000000",
      "currency": "USD",
      "cooldown_minutes": 60,
      "status": "ENABLED"
    }
  ]
}
```

`PATCH /watchlists/items/{id}` body supports partial item metadata:

- `pinned?: boolean`
- `note?: string`
- `status?: "ACTIVE" | "DISABLED"`
- `threshold_rules?: ThresholdRuleRequest[]`

Important PATCH rule behavior from the backend handoff:

- Omitting `threshold_rules` leaves existing rules untouched.
- Passing `threshold_rules` replaces existing enabled rules.
- Passing an empty array removes all enabled rules.
- Include rule `id` when updating an existing rule.

The edit drawer must preserve existing rules in the request when the user is editing rules, and must omit `threshold_rules` when the user changes only item metadata.

Current note-clearing gap: backend `UpdateItemRequest.Note` is a pointer string and the update handler changes the note only when that pointer is non-nil. In JSON decoding, both an omitted `note` field and `"note": null` become nil, so `null` does not clear the note. Until the backend defines a clear-note contract, the frontend should treat note update as "send a string to replace, omit to leave unchanged." If product requires clearing, either send an empty string only after backend confirms that convention or add a backend contract change for explicit clearing.

### Manual Evaluation

`POST /watchlists/evaluate` is operational/admin only. It should not appear as a normal Watchlist user action.

If an ops-only surface is approved later, body fields are:

- `scope_type?: "PERSONAL" | "PORTFOLIO"`
- `portfolio_id?: string`
- `security_id?: string`
- `item_id?: string`
- `rule_id?: string`
- `dry_run?: boolean`

## 7. Main Screens And Workflows

### WatchlistPage

`WatchlistPage.vue` owns the page composition:

- Wrap the screen in `AppPage`.
- Use `useWatchlist()` for item list, filters, create/update/delete, and optimistic or post-save refresh behavior.
- Use `useWatchlistAlerts()` for alert list and acknowledgement.
- Use `useI18n()` for all user-facing copy.
- Use `useAppToast()` for success and failure feedback.

Suggested page layout:

- Header actions: Add item if `WATCHLIST_MANAGE` exists.
- Scope selector: Personal and Portfolio. Portfolio requires a selected portfolio before list/create calls can include `portfolio_id`.
- Filters: scope, portfolio, security, status/include disabled, acknowledgement state for alerts, and date range for alerts.
- Main table: watchlist items with threshold summaries and latest quote.
- Right rail or lower panel: recent alerts with acknowledge actions.
- Drawer: create/edit item and threshold rules.

### Items Table

`WatchlistItemsTable.vue` should use `AppDataTable`, not a raw table unless `AppDataTable` cannot support a required layout.

Recommended columns:

- Security: `security.display_symbol`, `security.name`, optional `asset_type`.
- Scope: localized `PERSONAL` or `PORTFOLIO`.
- Portfolio: `portfolio.display_name` for portfolio-scoped rows; omit for personal rows.
- Current market price: `quote.price` and `quote.currency` through `AppMoney`.
- Change percent: `quote.change_percent` through `AppPercent`.
- Thresholds: one or more `MARKET_PRICE` rules with direction, value, currency, status, and last state.
- Quote status: fresh/stale using `quote.stale`, `quote.stale_reason`, or rule `last_quote_stale`.
- Last evaluated and last alerted timestamps.
- Actions: edit/delete only with `WATCHLIST_MANAGE`.

Rows should use resource IDs only for keys and API calls. Do not display item IDs, rule IDs, security IDs, portfolio IDs, owner IDs, or acknowledgement IDs as labels.

### Filters

`WatchlistFilters.vue` should hold filter controls and emit typed filter state:

- Scope type: all, personal, portfolio.
- Portfolio selector: enabled only for portfolio scope.
- Security search/filter: selected canonical security ID plus display label.
- Include disabled toggle.
- Alert acknowledgement filter: all, unacknowledged, acknowledged.
- Created date range for alert history.

### Selector Data Sources

Selectors must be backed by existing typed APIs or composable patterns. Do not use free-form UUID text inputs as the normal user workflow.

Security selector:

- Do not import `frontend/app/features/market-data/services/marketDataCatalogApi.ts` directly from the Watchlist feature; that service is useful as an implementation pattern but belongs to Market Data.
- Add a Watchlist-owned wrapper around `GET /reference-data/securities/search`, or extract a shared reference-data catalog service if multiple features need the same selector behavior.
- Store and submit the selected canonical `security_id`.
- Display `display_symbol`, `name`, `asset_type`, `currency`, and `exchange_mic` when available.

Portfolio selector:

- Do not import `frontend/app/features/investment-ledger/composables/usePortfolioDirectory.ts` directly from the Watchlist feature; that would blur feature ownership.
- Copy that pattern into a Watchlist-owned wrapper around `GET /investment/portfolios`, or extract a shared portfolio directory composable if Watchlist and investment ledger both need it.
- The existing pattern calls `investmentLedgerApi.listPortfolios({ limit: 200 })`, backed by `GET /investment/portfolios`.
- Store and submit the selected portfolio ID only in API payloads and query params.
- Display portfolio names/codes from the portfolio list, and use Watchlist response `portfolio.display_name` once rows are returned.
- If the portfolio list cannot load, show a localized empty/error state and do not allow `PORTFOLIO` item creation.

### Item Drawer

`WatchlistItemDrawer.vue` should handle both create and edit flows:

- Scope selector: `PERSONAL` or `PORTFOLIO`.
- Portfolio field: required only for `PORTFOLIO`; use the portfolio selector data source above.
- Security selector: use reference-data security search, choose a canonical security, and send `security_id`.
- Pinned toggle.
- Note field.
- Embedded `ThresholdRuleForm.vue`.
- Save button gated by `WATCHLIST_MANAGE`.
- Delete/disable action behind `AppConfirmDialog`.

The drawer must not let the user choose NAV, AUM, valuation, risk, or portfolio-performance metrics. The only metric field is `MARKET_PRICE`.

### Threshold Rule Form

`ThresholdRuleForm.vue` should support:

- Metric: fixed `MARKET_PRICE`, usually hidden or shown as read-only copy.
- Direction: `ABOVE` or `BELOW`.
- Threshold value: positive decimal string.
- Currency: default from `security.currency` or quote currency when available.
- Cooldown minutes: default `60`.
- Status: `ENABLED` or `DISABLED`.

Keep threshold form values as strings until submission. Validate string decimals, but do not convert request payload decimals to numbers.

### Alerts Panel

`WatchlistAlertsPanel.vue` should list alert events visible to the actor:

- Security label from descriptor.
- Portfolio label from descriptor when present.
- Direction and threshold crossing copy.
- Observed price and threshold value through `AppMoney`.
- Stale quote badge and reason when present.
- Notification status.
- Acknowledgement state.
- Acknowledged by display name and timestamp when populated.
- Acknowledge action only with `WATCHLIST_ALERT_ACK` and only for unacknowledged alerts.

`AcknowledgeAlertDialog.vue` must accept an optional note and refresh or patch alert state after success. Current `AppConfirmDialog` supports fixed title, description, and confirm/cancel actions, but it does not expose a body slot for an input. Either extend `AppConfirmDialog` with a body slot before using it for acknowledgement notes, or build an equivalent Watchlist-owned modal pattern for this dialog.

## 8. Display And Formatting Rules

### IDs And Labels

Never render raw UUIDs as user-facing labels. Use:

- `security.display_symbol` and `security.name`.
- `portfolio.display_name`.
- `created_by_user.display_name`.
- `acknowledged_by_user.display_name`.

If a user descriptor is missing, omit the label instead of showing `created_by_user_id` or `acknowledged_by`.

Keep IDs in state for:

- Vue keys.
- API path params.
- selected filters.
- mutation payloads.
- cache normalization.

### Decimal Strings

The API returns price, threshold, previous close, and percentage values as decimal strings. Keep raw API values as strings in feature state.

Use:

- `AppMoney` for money and price display: pass decimal strings such as `quote.price`, `quote.previous_close`, `threshold_value`, and `observed_price` as `:amount`, and pass currency codes such as `THB` or `USD` as `:currency`.
- `AppPercent` for `quote.change_percent`.
- formatter helpers in `lib/formatters.ts` for null handling, direction labels, threshold summaries, stale text, and timestamp text.

Do not hardcode display symbols such as `$` or the baht symbol; pass currency codes such as `THB` or `USD` to `AppMoney`.

### Status Labels

Use localized labels for enum values:

- Scope: `PERSONAL`, `PORTFOLIO`.
- Item status: `ACTIVE`, `DISABLED`.
- Direction: `ABOVE`, `BELOW`.
- Rule status: `ENABLED`, `DISABLED`.
- Rule state: `UNKNOWN`, `NON_BREACHED`, `BREACHED`.
- Notification status: `PENDING`, `CREATED`, `SUPPRESSED`, `FAILED`, `SKIPPED`.
- Acknowledgement state: `UNACKNOWLEDGED`, `ACKNOWLEDGED`.

Map these through `formatters.ts` and i18n keys, not inline template strings.

## 9. State And Error Behavior

### Loading

- Use `AppPage` `loading` for the first page load.
- Use `AppDataTable` `loading` for table refreshes.
- Disable mutation buttons while a mutation is in flight.
- Keep item and alert loading states separate so one panel can refresh without blocking the whole page.

### Empty States

Use `AppEmptyState`:

- No watchlist items: show a create action only when `WATCHLIST_MANAGE` exists.
- No alerts: explain that no threshold crossings have produced visible alerts yet.
- No portfolio access: show a permissions-focused empty state without exposing portfolio IDs.

### Error Handling

API errors use the platform envelope:

```json
{
  "error": "message",
  "code": "WATCHLIST_DUPLICATE_ITEM",
  "details": {}
}
```

`OpenApiRequestError.details` may include this object. Map known `code` values to localized user messages and keep raw `details` for logs or developer tooling only.

Recommended user-facing handling:

| Code | UI behavior |
| --- | --- |
| `VALIDATION_ERROR` | Show field-level validation when possible, otherwise a localized invalid-request message. |
| `WATCHLIST_INVALID_QUERY` | Reset or correct the offending filter. |
| `WATCHLIST_INVALID_SECURITY` | Tell the user to select an active canonical security. |
| `WATCHLIST_DUPLICATE_ITEM` | Show that the security is already in the selected scope; offer to open the existing row if it is in the current list. |
| `WATCHLIST_DUPLICATE_THRESHOLD` | Show that the same market-price rule already exists for the item. |
| `WATCHLIST_INVALID_THRESHOLD` | Highlight direction, value, currency, or cooldown fields. |
| `WATCHLIST_FORBIDDEN_SCOPE` | Show that the user cannot access this personal or portfolio scope. Do not render IDs from `details`. |
| `WATCHLIST_ITEM_NOT_FOUND` | Remove stale local row state and show a not-found toast. |
| `WATCHLIST_ALERT_NOT_FOUND` | Remove stale alert state and show a not-found toast. |
| `WATCHLIST_ALERT_ALREADY_ACKNOWLEDGED` | Refresh the alert list and show that it has already been acknowledged. |
| `WATCHLIST_RULE_DISABLED` | Ops-only targeted evaluation: show that the selected rule is disabled. Normal users should not see manual evaluation controls. |
| `WATCHLIST_STALE_QUOTE` | Ops-only targeted evaluation: show stale quote failure. Normal users should only see stale badges on data. |
| `WATCHLIST_PROVIDER_UNAVAILABLE` | Ops-only targeted evaluation or refresh: show provider unavailable message. |
| `FORBIDDEN` | Watchlist-coded forbidden response, if present. Hide future action where possible; if encountered after a click, show a permission message. |
| HTTP `403` with no `code` | Route-level permission middleware may return `{ "error": "insufficient permissions" }` without a code. Treat it as missing function permission and show the same localized permission message. |

### Stale Quotes

Stale quote is a data-quality state, not an automatic page error.

Display stale state when:

- `item.quote?.stale` is true.
- a threshold rule has `last_quote_stale` true.
- an alert event has `stale` true.

Show `stale_reason` when populated and safe, but keep it concise. Do not call stale data "stale NAV"; v1 evaluates market prices only.

### Forbidden Portfolio Scope

When `WATCHLIST_FORBIDDEN_SCOPE` is returned for a portfolio action:

- Keep the user on the page.
- Clear or reject the invalid portfolio selection.
- Show a localized permission message.
- Do not show `portfolio_id` from error details.

### Duplicate Item

When `WATCHLIST_DUPLICATE_ITEM` is returned:

- Keep the drawer open.
- Show a scope-specific duplicate message.
- If the duplicate item is already visible in the current list, select or scroll to that row.
- Do not display `security_id` or `portfolio_id` from error details.

## 10. Testing Checklist

Add focused Vitest coverage before treating the feature as complete.

API service tests:

- Calls each `/watchlists` endpoint with typed path/query/body params.
- Uses `unwrapOpenApiResponse()` for data responses.
- Handles `DELETE /watchlists/items/{id}` as `204 No Content`.
- Does not call `/market-data/screen/watchlist`.
- Maps `OpenApiRequestError` details with Watchlist error codes.

Formatter tests:

- Decimal strings remain valid display inputs.
- `AppMoney` and `AppPercent` wrapper helpers preserve null and empty states.
- Direction, scope, item status, rule state, acknowledgement state, and notification status labels are localized.
- UUID fields are never selected as display labels.
- Stale quote labels use `stale_reason` without saying NAV.

Permission tests:

- `/watchlist` route meta requires `WATCHLIST_VIEW`.
- Sidebar item requires `WATCHLIST_VIEW`.
- Create/edit/delete controls require `WATCHLIST_MANAGE`.
- Acknowledge controls require `WATCHLIST_ALERT_ACK`.
- Manual evaluate controls are absent for normal users and gated by `WATCHLIST_EVALUATE` if an ops-only surface is introduced.

Component behavior tests:

- Page shows loading, error, and empty states.
- Filters call `listItems` and `listAlerts` with correct query params.
- Portfolio ID is sent only for `PORTFOLIO` scope.
- Item drawer submits decimal strings, not numbers.
- Edit drawer omits `threshold_rules` when rules are not edited.
- Edit drawer sends the full replacement array when rules are edited.
- Empty threshold array removes all rules only after explicit user intent.
- Duplicate item and forbidden scope errors show friendly messages without raw UUID labels.
- Alert acknowledgement updates or refreshes the alert list.

Recommended verification commands from `frontend/`:

```bash
npm run test
npm run build
```

Use `npx nuxi typecheck` when practical, but account for any pre-existing unrelated typecheck errors separately.

## 11. Implementation Checklist

Future frontend coding agent checklist:

- [ ] Confirm backend Swagger exposes all Watchlist paths.
- [ ] Confirm backend Swagger exposes concrete Watchlist success data schemas, not only generic `httputil.SuccessResponse`.
- [ ] Run `make api-client` from the repository root.
- [ ] Verify `frontend/app/api/ims-api.d.ts` includes `/watchlists` and Watchlist DTO schemas.
- [ ] Replace `frontend/app/pages/watchlist/index.vue` with the thin route shell.
- [ ] Add `frontend/app/features/watchlist/types.ts` using generated `components["schemas"]` aliases, or an explicitly approved temporary local bridge that mirrors `dto.go`.
- [ ] Add `services/watchlistApi.ts` using `useOpenApiClient()`, `unwrapOpenApiResponse()`, and no market-data screen watchlist calls.
- [ ] Add `composables/useWatchlist.ts` for item filters, pagination, create/update/delete, and refresh.
- [ ] Add `composables/useWatchlistAlerts.ts` for alert filters, pagination, and acknowledgement.
- [ ] Support Watchlist notification deep links by changing backend action URLs to `/watchlist?alert_id=...` or adding a frontend `/watchlists/alerts/[id].vue` redirect shim.
- [ ] Add `lib/formatters.ts` for enum labels, descriptor labels, stale text, timestamps, and null-safe display helpers.
- [ ] Back the security selector with a Watchlist-owned wrapper around `GET /reference-data/securities/search`, or extract a shared reference-data catalog service; do not import the Market Data feature service directly.
- [ ] Back the portfolio selector with a Watchlist-owned wrapper around `GET /investment/portfolios`, or extract a shared portfolio directory composable; do not import another feature's composable directly.
- [ ] Decide whether note clearing is needed; if yes, confirm an empty-string convention or request an explicit backend clear-note contract.
- [ ] Build `WatchlistPage.vue` with `AppPage`, shared UI components, localized copy, and permission-gated actions.
- [ ] Build the suggested feature components under `components/`.
- [ ] Add the sidebar item in `frontend/app/features/shell/navigation.ts` with `requiredPermissions: ["WATCHLIST_VIEW"]`.
- [ ] Add `watchlist.ts` i18n modules for `en`, `th`, and `zh`, then register them in each locale index.
- [ ] Add `navigation.watchlist` to each locale's existing `common.ts` `navigation` object; do not add a top-level `navigation` object in `watchlist.ts`.
- [ ] Add tests for API service, formatters, permissions, and page/component behavior.
- [ ] Run frontend tests and build.
- [ ] Manually verify that no raw UUIDs render in the Watchlist UI.
- [ ] Manually verify that `/market-data/screen/watchlist` remains only in the Market Data feature, not in the Watchlist feature.
