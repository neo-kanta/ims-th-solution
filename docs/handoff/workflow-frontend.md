# Workflow Frontend Handoff

**Branch:** `neo-develop`
**As of:** 2026-06-15
**Backend API status:** Blocker resolved (ContractCatalog wired). New endpoints not yet built — see `workflow-backend.md` for remaining work.

---

## 1. Purpose

Build the Workflow feature page so investment managers and supervisors can:

- See the current workflow state for a business date and contract
- Execute workflow operations (start day, approve, close transactions, close accounting)
- Review the full transition history / audit timeline
- Configure who is allowed to approve each workflow operation

**This page is a prerequisite for all investment transaction operations.** Users must be able to understand the current workflow state within 3 seconds of opening the page.

---

## 2. Location and Structure

### File locations

```
frontend/app/features/workflow/
  components/         ← Vue components
  composables/        ← already has useWorkflowOperations.ts, useWorkflowStore.ts
  store/              ← existing Pinia store
  types.ts            ← extend or replace with new API shapes
  permissions.ts      ← existing permission codes
  NAMING.md           ← read before adding files
  navigation/         ← add dashboardTabs.ts here for tab registration
```

```
frontend/app/pages/workflow/
  index.vue           ← existing page shell — will be redesigned
```

### Tab registration

**Do NOT add tabs directly to `dashboard.vue`.**

Use the dashboard tab registry pattern (introduced 2026-06-15):
1. Create `frontend/app/features/workflow/navigation/dashboardTabs.ts` implementing `DashboardTabProvider`
2. Register it in `frontend/app/features/shell/tabs/dashboardTabRegistry.ts`

See `docs/frontend/dashboard-tab-registry-refactor.md` for the full pattern.

### Existing components (already in repo)

These exist and may be adapted or replaced:

| File | Current purpose |
|------|-----------------|
| `WorkflowDayRail.vue` | Side-rail showing day state (old UUID-based) |
| `WorkflowContractPicker.vue` | Contract selector (may be reused) |
| `WorkflowOperationConfirmDialog.vue` | Confirmation dialog before state change |
| `WorkflowOperationForm.vue` | Form for executing operations |
| `WorkflowOperationHistoryTimeline.vue` | Timeline of past transitions |
| `WorkflowStageTracker.vue` | Visual stage progress indicator |

The existing composables `useWorkflowOperations.ts` and `useWorkflowStore.ts` use the old UUID-based API (`/workflow/day-states/{contractId}/...`). They will need to be updated to use the new business-readable API once the backend endpoints are built.

---

## 3. Required Tabs

Register three tabs on the Workflow feature page:

| Tab | Route/anchor | Purpose |
|-----|--------------|---------|
| **Overview** | default | Current state, allowed actions, timeline summary |
| **Audit** | `#audit` | Full transition history, actor details, admin override flags |
| **Settings** | `#settings` | Configure approved roles / users per operation |

---

## 4. Backend API Expected

> **Status:** These endpoints are NOT yet built. Backend implementation is in progress.
> Do not implement the frontend against mocked data that deviates from this contract.

### `GET /api/workflow/daily`

```
GET /api/workflow/daily?businessDate=YYYY-MM-DD&contractCode=ABC
Authorization: Bearer <token>
```

Returns the full daily workflow state (see `workflow-backend.md` Section 7 for the full response shape).

Key response fields to consume in the UI:

```typescript
interface WorkflowDailyResponse {
  businessDate: string           // "YYYY-MM-DD"
  contractCode: string
  currentState: WorkflowState
  isToday: boolean
  allowedOperations: string[]
  blockedReasons: string[]
  timeline: WorkflowTimelineEntry[]
  approvers: WorkflowApproverEntry[]
  settings: WorkflowSettings
  auditSummary: WorkflowAuditSummary
  moduleReadiness: WorkflowModuleReadiness
}
```

### `POST /api/workflow/daily/execute`

```
POST /api/workflow/daily/execute
Content-Type: application/json

{
  "businessDate": "YYYY-MM-DD",
  "contractCode": "ABC",
  "operationType": "MANAGER_APPROVE",
  "remark": ""
}
```

### `GET /api/workflow/settings`

```
GET /api/workflow/settings?contractCode=ABC
```

### `PUT /api/workflow/settings`

```
PUT /api/workflow/settings?contractCode=ABC
Content-Type: application/json

{
  "defaultApproverRole": "Admin",
  "approvalMode": "ANY_OF",
  "configuredApprovers": [...]
}
```

### API client

- Do **NOT** write manual fetch code.
- Wait for `make swagger && make api-client` to generate `frontend/app/api/ims-api.d.ts`.
- Access backend calls via `useOpenApiClient()` or `useApi()` composable.

---

## 5. UI Requirements

### Date and contract selection

- User selects `businessDate` (date picker, default today in Asia/Bangkok)
- User selects `contractCode` from a list or types it (business-readable, e.g., "ABC")
- **Never require the user to know or input a UUID**

### Current state display

- Show current state as a human-readable badge/chip: `NOT STARTED`, `INVESTMENT DAY STARTED`, `MANAGER APPROVED`, `TRANSACTION CLOSED`, `ACCOUNTING CLOSED`
- Show the stage tracker (already exists as `WorkflowStageTracker.vue`)
- Highlight the current stage

### Allowed operations

- Show action buttons only for operations listed in `allowedOperations`
- Disable all action buttons if `moduleReadiness` has any `{ ready: false }` entry that blocks the current operation
- Show `blockedReasons` as an info/warning banner

### Timeline

- Show `timeline` entries in reverse chronological order
- Display: `operationType`, `fromState → toState`, `executedAt` (formatted for Asia/Bangkok), `executedByAccountCode`
- **Never display raw UUIDs in the timeline**
- Show an "Admin Override" badge if `isAdminOverride: true`

### Approvers display

- Show configured approvers from `approvers` array
- Show "Admin (default)" if `isAdminOverride: true` on the approver entry

### Audit summary

- Show `auditSummary.lastActorUsername`, `lastAction`, `lastActionAt` in the page header or info section

### Module readiness

- If `moduleReadiness.approval.ready === false`, show an info banner: "Approval module not yet active — workflow operates with Admin group as default approver."
- Do not block the page entirely due to module readiness warnings

### Settings tab

- List current approvers per operation
- Allow adding/removing approvers by username or account code
- Save via `PUT /api/workflow/settings`
- Show `Admin` as the default approver for all operations

---

## 6. Suggested Component Structure

```
frontend/app/features/workflow/
  components/
    WorkflowPage.vue                  ← top-level page wrapper
    WorkflowOverviewTab.vue           ← Overview tab content
    WorkflowAuditTab.vue              ← Audit/history tab content
    WorkflowSettingsTab.vue           ← Settings/approvers tab content
    WorkflowStateBadge.vue            ← state chip/badge (NOT_STARTED, etc.)
    WorkflowTimeline.vue              ← transition history list
    WorkflowActionPanel.vue           ← action buttons (start, approve, close)
    WorkflowApproverSelector.vue      ← approver search/add widget
    WorkflowAuditTable.vue            ← full audit table for Audit tab
    WorkflowModuleReadinessBanner.vue ← warning banner for scaffold modules
  composables/
    useWorkflow.ts                    ← main data fetching composable
    useWorkflowExecute.ts             ← mutation composable for execute endpoint
    useWorkflowSettings.ts            ← settings read/write composable
  services/
    workflow.api.ts                   ← typed API calls (generated client wrappers)
  types/
    workflow.types.ts                 ← TypeScript types matching API shapes
  navigation/
    dashboardTabs.ts                  ← DashboardTabProvider registration
  store/
    workflow.store.ts                 ← Pinia store (if cross-component state needed)
```

---

## 7. UX Behavior

| Scenario | Expected behavior |
|----------|-------------------|
| User opens Workflow page | Date defaults to today; contract defaults to user's primary fund if available |
| Date/contract changes | Triggers new `GET /api/workflow/daily` fetch; loading skeleton shown |
| State is `NOT_STARTED` | "Start Investment Day" button visible and enabled |
| State is `INVESTMENT_DAY_STARTED` | "Manager Approve" and "Cancel Day Start" visible |
| State is `MANAGER_APPROVED` | "Close Transactions" visible; no further investment transactions allowed |
| User clicks action button | Confirmation dialog (reuse `WorkflowOperationConfirmDialog.vue`) |
| Remark required (cancel/rollback) | Remark field shown in dialog; submit disabled if empty |
| `POST /execute` succeeds | Refetch `GET /api/workflow/daily`; show success toast |
| `POST /execute` fails | Show error from `errcode` envelope; do not close dialog |
| `moduleReadiness.approval.ready === false` | Show info banner; do not block operations that don't require approval module |
| `blockedReasons` non-empty | Show orange warning banner listing each reason |
| Admin user | All operations available regardless of approver settings |
| Non-admin without permission | Buttons disabled with tooltip "Not authorized for this operation" |

---

## 8. Theme and Accessibility

- Read color tokens from `frontend/app/assets/css/main.css` — do not hardcode hex colors
- Support dark and light themes using CSS custom properties already defined in the project
- Use semantic HTML (`<button>`, `<table>`, `<nav>`) for accessibility
- Responsive: must work on 1024px+ desktop (primary target) and degrade gracefully on tablet

---

## 9. Acceptance Criteria

- [ ] User opens Workflow page without any UUID in the URL or query params
- [ ] Selecting a date and contract code loads the correct workflow state
- [ ] Current state is understandable within 3 seconds of page load
- [ ] Overview tab shows: current state badge, stage tracker, allowed action buttons, blocked reasons
- [ ] Clicking an action button shows confirmation dialog; confirm executes; page refreshes
- [ ] Audit tab shows full transition timeline with `accountCode`, timestamps, and admin override badges
- [ ] Settings tab shows configured approvers and allows updates via `PUT /api/workflow/settings`
- [ ] Dark and light themes both render correctly
- [ ] Layout is responsive at 1024px+ desktop width
- [ ] `npm run build` passes with no TypeScript errors
- [ ] No raw UUIDs visible in any normal user-facing UI element

---

## 10. Notes

- The existing `WorkflowDayRail.vue` uses the old UUID path API (`/workflow/day-states/{contractId}/...`). When the new `GET /api/workflow/daily` endpoint is live, update or replace this component.
- The existing `frontend/app/features/dashboard/composables/useDashboardWorkflow.ts` may be reused or replaced by `useWorkflow.ts` — evaluate which is the right consolidation point.
- `frontend/app/pages/workflow/index.vue` currently shows a placeholder ("Go to per-fund workflow"). This file will be replaced by the new `WorkflowPage.vue` driven by the daily API.
