# Dashboard Tab Registry — Refactor Handoff

**Date:** 2026-06-15
**Scope:** Frontend only — `frontend/app/`
**Status:** Complete. Build passes, 105 tests pass, zero new type errors.

---

## Problem solved

`app/layouts/dashboard.vue` was the single point of truth for all header-tab logic across three unrelated features (compliance, investment workspace, market data). Every time any feature needed to add, remove, rename, or badge-count a tab, the global layout had to change. It also directly imported feature-specific composables (`useComplianceRuleDirectory`, `useComplianceBreachesList`, `useMarketDataCatalog`).

---

## What was changed

### New files

```
app/features/shell/tabs/types.ts
app/features/shell/tabs/dashboardTabRegistry.ts
app/features/shell/composables/useDashboardHeaderTabs.ts
app/features/compliance/navigation/dashboardTabs.ts
app/features/investment-workspace/navigation/dashboardTabs.ts
app/features/market-data/navigation/dashboardTabs.ts
```

### Modified files

```
app/layouts/dashboard.vue
```

No backend, infra, or database files were touched.

---

## New architecture

### Layer 1 — Shared types (`shell/tabs/types.ts`)

Defines the three types that hold the pattern together:

```ts
// Context passed to every provider at setup time.
type DashboardTabContext = {
  route: RouteLocationNormalizedLoaded;   // reactive route from useRoute()
  t: (key: AppTranslationKey, ...) => string;  // from useI18n()
};

// What a provider produces during setup.
type DashboardTabSetup = {
  items: ComputedRef<TabItem[]>;       // reactive tab list
  activeKey: ComputedRef<string>;      // reactive active tab key
  ariaLabel: string;                   // static accessibility label
  onRouteEnter?: () => void | Promise<void>;  // preload hook
};

// The contract every feature provider must implement.
type DashboardTabProvider = {
  id: string;
  matches: (route: RouteLocationNormalizedLoaded) => boolean;
  setup: (ctx: DashboardTabContext) => DashboardTabSetup;
};
```

Also re-exports `TabItem` from `AppTabs.vue` so provider files have a single clean import.

### Layer 2 — Feature providers (one per feature)

Each feature owns a file at `features/<feature>/navigation/dashboardTabs.ts`.
It exports a single `DashboardTabProvider` object.

| Provider file | Owns |
|---|---|
| `compliance/navigation/dashboardTabs.ts` | Tab items, active-key logic (path-based), preload for rule directory and open breaches |
| `investment-workspace/navigation/dashboardTabs.ts` | Tab items (fundId-interpolated), active-key logic (path segment 4), no preload |
| `market-data/navigation/dashboardTabs.ts` | Tab items, active-key logic (query-string `?tab=`), preload for REVIEW_REQUIRED candidates |

### Layer 3 — Registry (`shell/tabs/dashboardTabRegistry.ts`)

The only file that imports all three providers. Order determines match priority (first match wins).

```ts
export const dashboardTabRegistry: DashboardTabProvider[] = [
  complianceDashboardTabs,
  investmentWorkspaceDashboardTabs,
  marketDataDashboardTabs,
];
```

### Layer 4 — Composable (`shell/composables/useDashboardHeaderTabs.ts`)

Called once from `dashboard.vue` setup. It:

1. Calls `useRoute()` and `useI18n()`.
2. Calls **every** provider's `setup(ctx)` **eagerly** — this is mandatory because `useState` and other Nuxt composables must be called while the component instance is active (layout setup). Route-gating the `setup()` calls would cause "Nuxt instance unavailable" errors on navigation.
3. Exposes `{ hasHeaderTabs, activeTabItems, activeTabValue, activeTabAriaLabel }`.
4. Watches `route.path` and calls the matched provider's `onRouteEnter()` — with `{ immediate: true }` to preload on first load.

### Layer 5 — Layout (`layouts/dashboard.vue`)

Now contains zero feature-specific imports or tab logic. Its tab section is:

```ts
const { hasHeaderTabs, activeTabItems, activeTabValue, activeTabAriaLabel } = useDashboardHeaderTabs();
```

The route watch was simplified to only reset `pageTitle`; preloading is the composable's responsibility.

`isComplianceRoute` remains as a local computed in `dashboard.vue` because the breadcrumb logic (`breadcrumbRepo`, `breadcrumbIcon`, `handleRepoClick`) still needs it. It was left in-place per scope decision — breadcrumb abstraction was assessed as over-engineering risk.

---

## How to add tabs for a new feature

**Step 1** — Create `app/features/<your-feature>/navigation/dashboardTabs.ts`:

```ts
import { computed } from "vue";
import type { DashboardTabProvider } from "~/features/shell/tabs/types";

export const yourFeatureDashboardTabs: DashboardTabProvider = {
  id: "your-feature",

  // Return true for every route that should show these tabs.
  matches: (route) => route.path.startsWith("/your-feature"),

  setup({ route, t }) {
    const items = computed(() => [
      { key: "overview", label: t("yourFeature.tabs.overview"), to: "/your-feature", icon: "list" },
      // ... more tabs
    ]);

    const activeKey = computed(() => {
      if (route.path === "/your-feature") return "overview";
      // ... more path cases
      return "";
    });

    // Optional: data to prefetch when the user enters this section.
    async function onRouteEnter() {
      // void yourComposable.ensureLoaded();
    }

    return { items, activeKey, ariaLabel: "Your feature tabs", onRouteEnter };
  },
};
```

**Step 2** — Register in `app/features/shell/tabs/dashboardTabRegistry.ts`:

```ts
import { yourFeatureDashboardTabs } from "~/features/your-feature/navigation/dashboardTabs";

export const dashboardTabRegistry: DashboardTabProvider[] = [
  complianceDashboardTabs,
  investmentWorkspaceDashboardTabs,
  marketDataDashboardTabs,
  yourFeatureDashboardTabs,   // ← add here
];
```

`dashboard.vue` does not need to change.

---

## Design constraints to respect

**Eager provider setup is not optional.**
All `provider.setup(ctx)` calls happen synchronously during `useDashboardHeaderTabs`'s setup execution, which runs inside `dashboard.vue`'s component setup. Nuxt composables (`useState`, etc.) require an active component instance. If you ever refactor this to call `setup()` lazily (e.g., inside a computed), composables inside the provider will throw at runtime on page refresh or deep-link navigation.

**No fabricated counts.**
The investment workspace provider deliberately has no `count` on any tab. The backend does not yet expose per-fund, per-tab counts. Do not add hardcoded numbers; leave `count` absent until a real endpoint exists.

**Market-data active key is query-string based.**
All other providers use `route.path`. Market data uses `route.query.tab`. If you add a new feature with query-string tab selection, follow the same pattern in that feature's provider.

**`matches` must be tight.**
Investment workspace uses `route.path.startsWith("/investment") && !!route.params.fundId` — not just the prefix — to avoid showing workspace tabs on the fund list page `/investment/funds`. Be equally specific in your own provider.

---

## Validation results (2026-06-15)

| Check | Result |
|---|---|
| `npm run build` (Nuxt build) | ✅ Passed |
| `npm run test` (Vitest, 12 suites) | ✅ 105 / 105 passed |
| `npx nuxi typecheck` — files added/modified by this refactor | ✅ 0 errors |
| `npx nuxi typecheck` — pre-existing errors in other files | 16 errors in unrelated files (`NavHistoryGraph.vue`, `WorkspaceTabScaffold.vue`, `workflow/permissions.ts`, i18n `th`/`zh` messages, `AppSearch.vue`) |

The 16 pre-existing errors are not regressions from this change and should be addressed in separate PRs.

---

## Known follow-up items

| Item | Priority | Notes |
|---|---|---|
| Fix pre-existing `nuxi typecheck` errors | Medium | 16 errors in unrelated files; none block runtime |
| Breadcrumb provider pattern | Low | `isComplianceRoute`, `breadcrumbRepo`, `breadcrumbIcon`, `handleRepoClick` are still feature-specific in `dashboard.vue`; a parallel `BreadcrumbProvider` registry would clean this up, but was deferred as lower value / higher risk |
| `SIDEBAR_COLLAPSED_WIDTH` unused constant | Low | Pre-existing dead code at line 70 of `dashboard.vue`; safe to remove in a cleanup pass |
| Investment tab badge counts | Deferred | No backend endpoint yet; leave `count` absent on all investment workspace tabs until the API lands |
