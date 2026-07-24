# IMS Thailand Frontend

This package contains the Nuxt 4 frontend for IMS Thailand. It provides the authenticated app shell, login flow, dashboard, and feature-owned screens for investment (research, decisions, ledger, portfolio workspace/decision), compliance, approval, permissions, market data, watchlist, notifications, chat, and settings.

## Stack

- Nuxt 4
- Vue 3
- Pinia
- TypeScript
- VueUse
- Day.js

## Current Architecture

Nuxt is configured with `srcDir: "app/"`, so application code lives under `frontend/app`.

The frontend is organized around three main ownership zones:

- `app/features/` for feature-owned components, composables, and feature helpers
- `app/shared/` for shared UI, routing helpers, and i18n
- `app/pages/` for route entry points that stay thin wherever practical

## Directory Layout

```text
frontend/
|-- app/
|   |-- app.vue
|   |-- assets/css/                  # Global styles
|   |-- composables/                 # Thin Nuxt composables such as useI18n/useApi
|   |-- features/
|   |   |-- approval/                # Approval inbox, groups/teams, subject timelines
|   |   |-- auth/                    # Auth helpers and types
|   |   |-- chat/                    # AI assistant chat UI
|   |   |-- compliance/              # IRG rule builder, breach inbox, checks
|   |   |-- dashboard/               # Dashboard screen and dashboard-specific logic
|   |   |-- investment-decision/     # V1 investment decision workflow
|   |   |-- investment-ledger/       # V1 ledger views
|   |   |-- investment-research/     # Analysis report workflow
|   |   |-- investment-workspace/    # V1 fund/portfolio workspace
|   |   |-- market-data/             # Quote/history screens
|   |   |-- my-funds/                # Manager's fund views
|   |   |-- notifications/           # Notification center
|   |   |-- operator/                # Portfolio V2 operator decision/execution screens
|   |   |-- permissions/             # Permission governance UI
|   |   |-- portfolio-decision/      # Portfolio V2 (portfolio-code) decision workflow
|   |   |-- portfolio-workspace/     # Portfolio V2 (portfolio-code) directory/holdings/cash
|   |   |-- settings/                # Settings/admin screen and service clients
|   |   |-- shell/                   # Navigation and dashboard tab registry
|   |   |-- watchlist/               # Personal/portfolio watchlist and alerts
|   |   `-- workflow/                # Business-day workflow state UI
|   |-- layouts/                     # default, auth, dashboard
|   |-- middleware/                  # auth and permission guards
|   |-- pages/                       # Route shells and file-based routing
|   |-- shared/
|   |   |-- i18n/                    # Locale registry, message modules, formatting helpers
|   |   |-- routing/                 # Shared route-access helpers
|   |   `-- ui/                      # Shared UI components
|   |-- stores/                      # Cross-feature state such as auth
|   `-- types/                       # Shared app-level types
|-- public/
|-- tests/                           # Vitest coverage for key helpers
|-- nuxt.config.ts
`-- package.json
```

## i18n

The frontend uses an internal shared i18n layer, not `@nuxtjs/i18n`.

Why:

- the app currently needs scalable shared translations more than localized routing
- the existing locale switcher and app shell are already custom
- feature-owned message files are easier to maintain inside this codebase

Translations live in:

```text
app/shared/i18n/messages/
|-- en/
|-- th/
`-- zh/
```

Each locale is split by concern, for example:

- `common.ts`
- `dashboard.ts`
- `settings.ts`
- `placeholders.ts`

When adding user-facing copy:

1. add the message to `en`
2. add the same key to `th` and `zh`
3. consume it through `useI18n().t(...)`
4. avoid hardcoded English strings in active screens

## Runtime Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `NUXT_PUBLIC_API_BASE_URL` | `http://localhost:8080/api/v1` | Client-visible API base URL |
| `NUXT_API_BASE_URL` | Falls back to `NUXT_PUBLIC_API_BASE_URL` | Server-side API base URL |
| `NUXT_PUBLIC_APP_NAME` | `IMS Thailand` | Application name in the UI |

Example:

```bash
NUXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
NUXT_API_BASE_URL=http://localhost:8080/api/v1
NUXT_PUBLIC_APP_NAME=IMS Thailand
```

## Development

Install dependencies:

```bash
npm install
```

Run the dev server:

```bash
npm run dev
```

The app runs at `http://localhost:3000` by default.

## Scripts

```bash
npm run dev
npm run build
npm run preview
npm run generate
npm run api:generate
npm run test
```

## Generated API Types

The frontend can generate typed API paths and schemas from the backend Swagger output:

```bash
npm run api:generate
```

From the repository root, `make api-client` first refreshes `backend/docs/swagger.json`/`backend/docs/v2/v2_swagger.json` and then regenerates `app/api/ims-api.d.ts`.

Use `useOpenApiClient()` from `app/api/openapi.ts` when a feature needs a typed V1 client backed by Swagger:

```ts
const client = useOpenApiClient();
const me = unwrapOpenApiResponse(await client.GET("/auth/me"));
```

Use `useOpenApiClientV2()` from the same module for the additive Portfolio V2
(portfolio-code) API at `/api/v2/portfolios/*`:

```ts
const clientV2 = useOpenApiClientV2();
const portfolio = unwrapOpenApiResponse(
  await clientV2.GET("/portfolios/{portfolioCode}", { params: { path: { portfolioCode } } }),
);
```

## Auth and Routing

- Auth state is owned by `app/stores/useAuthStore.ts`
- API requests attach the bearer token through `app/composables/useApi.ts`
- Route protection lives in `app/middleware/auth.ts` and `app/middleware/permission.ts`
- Shared route meta parsing lives in `app/shared/routing/routeAccess.ts`

Current limitation:

- the frontend still uses bearer-token persistence in frontend-managed storage because that is what the current backend contract supports
- a full HttpOnly-cookie session redesign would require backend changes and is not faked here

## Build Output

`npm run build` produces the Nuxt server bundle under `.output/`, with the server entry at:

```text
.output/server/index.mjs
```

## Notes for New Feature Work

When adding a new frontend feature:

- create feature-owned code under `app/features/<feature>/`
- keep route files small and route-oriented
- put reusable components in `app/shared/ui/` only when they are truly shared
- put translations in `app/shared/i18n/messages/<locale>/`
- prefer updating the feature README or root docs when you introduce a new runtime pattern
