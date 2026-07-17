# Portfolio V2 Frontend - DDD Design

**Status:** target design, not current implementation  
**Primary route family:** `/portfolios` or `/investment/portfolios`  
**Frontend principle:** route shells are thin; feature modules own use cases, API adapters, state, and view models.

## 1. Purpose

Portfolio V2 frontend makes the user experience match the real business center of the application: portfolio management.

The current UI is partly fund-first. V2 changes the mental model:

- Users choose a portfolio, not a fund, to do operational work.
- Fund/product data appears only as optional context.
- Real, simulation, and model portfolios are clearly separated.
- Decisions, executions, confirmations, ledger, holdings, cash, valuation, compliance, and watchlists are portfolio-scoped.

## 2. Frontend Bounded Contexts

### Portfolio Workspace

Owns:

- Portfolio directory.
- Portfolio detail shell.
- Portfolio lifecycle controls.
- Portfolio tabs and route layout.
- Portfolio type behavior: `LIVE`, `SIMULATION`, `MODEL`.
- Cross-feature portfolio context provider.

Does not own:

- Compliance rule authoring.
- Approval engine state machine.
- Market-data symbol ingestion.
- Fund issuance screens.

### Portfolio Ledger

Owns:

- Transaction list.
- Transaction post form.
- Simulation preview.
- Reversal flow.
- Cash and holdings read models.

### Portfolio Decision

Owns:

- Decision list scoped by portfolio.
- Decision create/edit/submit.
- Basket/rebalance/switch line editor.
- Approval status panel for portfolio decisions.

### Portfolio Execution And Confirmation

Owns:

- Execution queue.
- Fill/cancel actions.
- Confirmation import/record/resolve.
- Mismatch review panel.

### Portfolio Valuation

Owns:

- Valuation snapshot list.
- Latest valuation panel.
- Run valuation action.
- Indicative simulation valuation display.

## 3. Target Route Map

Preferred route map:

```text
/portfolios
/portfolios/new
/portfolios/:portfolioCode
/portfolios/:portfolioCode/overview
/portfolios/:portfolioCode/holdings
/portfolios/:portfolioCode/cash
/portfolios/:portfolioCode/ledger
/portfolios/:portfolioCode/ledger/new
/portfolios/:portfolioCode/decisions
/portfolios/:portfolioCode/decisions/new
/portfolios/:portfolioCode/decisions/:decisionId
/portfolios/:portfolioCode/executions
/portfolios/:portfolioCode/confirmations
/portfolios/:portfolioCode/valuations
/portfolios/:portfolioCode/compliance
/portfolios/:portfolioCode/watchlists
/portfolios/:portfolioCode/settings
```

Portfolio route params use `portfolioCode`, not UUID. The frontend should generate links from `portfolio.code` and URL-encode the value when building API paths.

If product wants to keep the Investment sidebar group, use:

```text
/investment/portfolios/...
```

Do not keep fund-first workspace URLs as the primary operational route in V2.

## 4. Target Feature Layout

```text
frontend/app/features/portfolio-workspace/
  PortfolioWorkspacePage.vue
  PortfolioWorkspaceLayout.vue
  navigation/
    dashboardTabs.ts
  services/
    portfolioApi.ts
  composables/
    usePortfolioDirectory.ts
    usePortfolioContext.ts
    usePortfolioAllowedActions.ts
  types.ts
  lib/
    portfolioLabels.ts
    portfolioGuards.ts
  components/
    PortfolioHeader.vue
    PortfolioTypeBadge.vue
    PortfolioStatusBadge.vue
    PortfolioDirectoryTable.vue
    PortfolioCreateDrawer.vue
    PortfolioLifecycleActions.vue

frontend/app/features/portfolio-ledger/
  services/
    portfolioLedgerApi.ts
  composables/
    usePortfolioLedger.ts
    usePortfolioCash.ts
    usePortfolioHoldings.ts
    useTransactionSimulation.ts
  components/
    PortfolioHoldingsTable.vue
    PortfolioCashPanel.vue
    TransactionLedgerTable.vue
    TransactionPostDrawer.vue
    TransactionSimulationPanel.vue
    ReverseTransactionDialog.vue

frontend/app/features/portfolio-decision/
  services/
    portfolioDecisionApi.ts
  composables/
    usePortfolioDecisions.ts
    useDecisionDraft.ts
    useDecisionLifecycle.ts
  components/
    DecisionList.vue
    DecisionEditor.vue
    DecisionLinesEditor.vue
    DecisionSubmitPanel.vue
    DecisionApprovalPanel.vue

frontend/app/features/portfolio-execution/
  services/
    portfolioExecutionApi.ts
  composables/
    usePortfolioExecutions.ts
    useTradeConfirmations.ts
  components/
    ExecutionQueue.vue
    ExecutionFillDrawer.vue
    ConfirmationTable.vue
    ConfirmationResolveDialog.vue

frontend/app/features/portfolio-valuation/
  services/
    portfolioValuationApi.ts
  composables/
    usePortfolioValuations.ts
  components/
    LatestValuationPanel.vue
    ValuationHistoryTable.vue
    RunValuationDrawer.vue
```

Existing feature folders can be migrated gradually. New V2 work should avoid adding more fund-first dependencies to `my-funds` and `investment-workspace`.

## 5. Route Shell Pattern

Page files should stay thin:

```vue
<script setup lang="ts">
import PortfolioWorkspacePage from "~/features/portfolio-workspace/PortfolioWorkspacePage.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  meta: { permission: "PORTFOLIO_VIEW" },
});
</script>

<template>
  <PortfolioWorkspacePage />
</template>
```

Rules:

- No API calls in route files.
- No feature state in route files.
- No raw UUID display or UUID route generation in route files.
- Route files only bind page meta and delegate.

## 6. Portfolio Context

Every child feature under a portfolio route should read context from one composable:

```ts
type PortfolioContext = {
  portfolioCode: Ref<string>;
  portfolioId: ComputedRef<string | null>;
  portfolio: Ref<PortfolioResponse | null>;
  portfolioType: ComputedRef<"LIVE" | "SIMULATION" | "MODEL">;
  isLive: ComputedRef<boolean>;
  isSimulation: ComputedRef<boolean>;
  isModel: ComputedRef<boolean>;
  canPostLedger: ComputedRef<boolean>;
  canRunOfficialValuation: ComputedRef<boolean>;
  reload: () => Promise<void>;
};
```

`usePortfolioContext()` owns:

- Reading `portfolioCode` from the route.
- Loading the portfolio descriptor.
- Resolving internal `portfolio.id` for feature logic that truly needs it.
- Interpreting `portfolio_type`.
- Computing allowed feature behavior.
- Hiding legacy fund/contract details from child features.

`portfolioId` is internal state from the loaded DTO. Do not use it to build user-facing routes.

Do not pass `fund_id` or `contract_id` between frontend feature modules in V2.

## 7. Portfolio Type UI Rules

### LIVE

Show:

- Official holdings, cash, ledger, valuation.
- Approval/compliance/workflow status.
- Execution and confirmation actions.

Allow:

- Ledger post if user has permission and workflow allows.
- Official valuation run.
- Decision submit.

### SIMULATION

Show:

- Simulation badge beside the portfolio name.
- What-if ledger posts and simulation results.
- Indicative valuation only.

Disable or relabel:

- Official accounting close.
- Official valuation run.
- Compliance release wording that implies a real trade.

### MODEL

Show:

- Target allocation / model lines.
- Drift comparison if linked to a live portfolio later.

Disable:

- Ledger post.
- Cash movement.
- Execution.
- Confirmation.
- Official valuation.

## 8. Navigation Design

Sidebar target:

```ts
{
  label: t("navigation.portfolios", "Portfolios"),
  to: "/portfolios",
  icon: "portfolio",
  requiredPermissions: ["PORTFOLIO_VIEW"],
}
```

Dashboard header tabs should be owned by `portfolio-workspace/navigation/dashboardTabs.ts`.

Suggested tabs:

| Tab | Route | Visible For |
|---|---|---|
| Overview | `/portfolios/:portfolioCode/overview` | all |
| Holdings | `/portfolios/:portfolioCode/holdings` | `LIVE`, `SIMULATION` |
| Cash | `/portfolios/:portfolioCode/cash` | `LIVE`, `SIMULATION` |
| Ledger | `/portfolios/:portfolioCode/ledger` | `LIVE`, `SIMULATION` |
| Decisions | `/portfolios/:portfolioCode/decisions` | `LIVE`, `SIMULATION` |
| Executions | `/portfolios/:portfolioCode/executions` | `LIVE` |
| Confirmations | `/portfolios/:portfolioCode/confirmations` | `LIVE` |
| Valuations | `/portfolios/:portfolioCode/valuations` | `LIVE`, `SIMULATION` |
| Compliance | `/portfolios/:portfolioCode/compliance` | `LIVE` |
| Watchlists | `/portfolios/:portfolioCode/watchlists` | `LIVE`, `SIMULATION` |
| Settings | `/portfolios/:portfolioCode/settings` | all |

## 9. API Adapter Pattern

Each feature owns a small adapter over `useApi()`.

Example:

```ts
export const portfolioApi = {
  list(params: ListPortfoliosParams) {
    return apiGet<PortfolioListResponse>("/api/v2/portfolios", { params });
  },
  get(portfolioCode: string) {
    return apiGet<PortfolioResponse>(`/api/v2/portfolios/${encodeURIComponent(portfolioCode)}`);
  },
  create(body: CreatePortfolioRequest) {
    return apiPost<PortfolioResponse>("/api/v2/portfolios", body);
  },
};
```

Rules:

- API adapters return DTOs, not component state.
- Portfolio-scoped API adapters accept `portfolioCode` for route-scoped calls.
- Composables transform DTOs into view models.
- Components consume view models and emit commands.
- Components do not know legacy V1 fields.

## 10. View Models

### PortfolioListItem

```ts
type PortfolioListItem = {
  id: string;
  code: string;
  name: string;
  type: "LIVE" | "SIMULATION" | "MODEL";
  status: string;
  managerLabel: string;
  valuationCurrency: string;
  latestAumLabel: string;
  riskLabel: string;
};
```

### TransactionRow

```ts
type TransactionRow = {
  id: string;
  dateLabel: string;
  typeLabel: string;
  instrumentLabel: string;
  quantityLabel: string;
  priceLabel: string;
  cashImpactLabel: string;
  statusLabel: string;
  canReverse: boolean;
};
```

### DecisionRow

```ts
type DecisionRow = {
  id: string;
  decisionNumber: string;
  businessDateLabel: string;
  typeLabel: string;
  instrumentSummary: string;
  amountLabel: string;
  lifecycleStatus: string;
  approvalStatusLabel: string;
  canOpen: boolean;
};
```

## 11. State Ownership

| State | Owner |
|---|---|
| Current selected portfolio | Route param + `usePortfolioContext` |
| Portfolio list filters | `usePortfolioDirectory` |
| Ledger filters | `usePortfolioLedger` |
| Transaction draft | `TransactionPostDrawer` local state until submit |
| Simulation preview | `useTransactionSimulation` |
| Decision draft | `useDecisionDraft` |
| Valuation run draft | `RunValuationDrawer` local state |
| Toasts | `useAppToast` |
| Global loading indicator | `useGlobalProgress` |

Do not create a single global investment store containing all portfolio state.

## 12. User Workflows

### Create Portfolio

1. User opens `/portfolios`.
2. User opens create drawer.
3. User selects type: `LIVE`, `SIMULATION`, or `MODEL`.
4. Frontend validates required fields based on type.
5. API creates portfolio.
6. Directory reloads.
7. User is routed to `/portfolios/${portfolio.code}/overview`.

### Post Transaction

1. User opens portfolio ledger tab.
2. User opens post drawer.
3. Frontend validates portfolio type:
   - `LIVE`: allow normal post.
   - `SIMULATION`: show simulation/non-official treatment.
   - `MODEL`: block with clear disabled state.
4. User runs simulation preview.
5. User posts.
6. Ledger, holdings, and cash reload.

### Submit Decision

1. User opens decisions tab.
2. User creates decision under current portfolio.
3. Frontend sends only portfolio-scoped request data.
4. Backend performs workflow, compliance, and approval routing.
5. Decision list updates to submitted state.

### Run Valuation

1. User opens valuations tab.
2. User opens run valuation drawer.
3. Frontend chooses mode:
   - `OFFICIAL` only for `LIVE`.
   - `INDICATIVE` for `LIVE` or `SIMULATION`.
4. Backend returns valuation snapshot.
5. Latest valuation panel reloads.

## 13. Error Handling

| Backend Error | Frontend Behavior |
|---|---|
| `PORTFOLIO_NOT_FOUND` | Route-level not-found state |
| `PORTFOLIO_TYPE_NOT_ALLOWED` | Disable command and show concise reason |
| `WORKFLOW_DAY_CLOSED` | Show workflow badge and block submit/post action |
| `COMPLIANCE_BLOCKED` | Show rule breaches in submission panel |
| `INSUFFICIENT_POSITION` | Highlight quantity field |
| `MISSING_PRICE_SNAPSHOT` | Show valuation input warning |
| `VERSION_CONFLICT` | Reload entity and ask user to retry |

Normal UI should never render raw UUIDs as labels. Use portfolio code/name, instrument ticker/name, decision number, and user display names.

## 14. i18n

Add portfolio-specific modules:

```text
frontend/app/shared/i18n/messages/en/portfolio.ts
frontend/app/shared/i18n/messages/th/portfolio.ts
frontend/app/shared/i18n/messages/zh/portfolio.ts
```

Recommended top-level keys:

```ts
export default {
  portfolio: {
    title: "...",
    type: {
      live: "...",
      simulation: "...",
      model: "..."
    },
    status: {},
    ledger: {},
    decisions: {},
    executions: {},
    confirmations: {},
    valuations: {},
    errors: {}
  }
};
```

Navigation labels still belong in the locale module that currently owns `navigation`.

## 15. Migration Checklist

### Phase 1 - Add Portfolio V2 UI Beside Existing UI

- Add portfolio directory route.
- Add portfolio workspace shell.
- Add `portfolio_type` badge behavior.
- Use V2 API adapters when backend is available.
- Keep existing fund pages unchanged.

### Phase 2 - Move Operational Workflows

- Move ledger operation flows from fund workspace to portfolio workspace.
- Move decisions/executions/confirmations under portfolio routes.
- Update compliance and watchlist selectors to prefer portfolio directory.
- Update navigation from "My funds" to "Portfolios" when ready.

### Phase 3 - Remove Legacy UI Coupling

- Stop passing `fund_id` and `contract_id` in frontend forms.
- Remove fund-first decision filters from default UI.
- Keep optional fund/product context only as display metadata if backend still returns it.

### Phase 4 - Fund/Product Reintroduction Later

Only after portfolio V2 is stable:

- Add fund wrapper pages.
- Add unit classes.
- Add fund unit subscription/redemption flows.
- Add official fund NAV per unit pages.

## 16. Frontend Acceptance Criteria

- A user can manage portfolio operations without selecting a fund.
- Portfolio URLs use portfolio code, not raw UUID.
- Every operational screen can be reached from a portfolio detail workspace.
- `LIVE`, `SIMULATION`, and `MODEL` behaviors are visible and enforced.
- Transaction and decision forms do not include `fund_id` or `contract_id`.
- Simulation results are visually distinct from official accounting data.
- The UI never uses fund routes as the only way to reach portfolio ledger work.
- Feature code stays under feature folders; route files remain thin.

## Source References

- `frontend/app/features/investment-workspace`
- `frontend/app/features/investment-ledger`
- `frontend/app/features/investment-decision`
- `frontend/app/features/my-funds`
- `frontend/app/features/compliance`
- `frontend/app/features/watchlist`
- `docs/api/portfolio-v2-api-ddd.md`
- `docs/frontend/dashboard-tab-registry-refactor.md`
- `docs/handoff/frontend-uuid-free-policy.md`
