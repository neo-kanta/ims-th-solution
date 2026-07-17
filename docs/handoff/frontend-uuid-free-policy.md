# Frontend UUID-Free Display Policy

**As of:** 2026-06-17

Raw UUIDs (format `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`) must never appear as visible text in the IMS UI. This document explains the rule, its rationale, the approved patterns, and the regression tests that enforce it.

---

## Why UUIDs Must Not Be Displayed

1. **Users cannot interpret them.** A UUID provides zero business context to an operator, reviewer, or auditor.
2. **They leak internal identifiers.** Displaying database record IDs increases attack surface for enumeration attacks.
3. **They indicate a broken display path.** A UUID appearing as a label always means a name-resolution fallback was not implemented.

---

## The Rule

> A raw UUID string must never be placed in a rendered text node that is visible to the user.

UUIDs are only acceptable in:
- `href` / `to` attribute values (navigation routing)
- `data-*` attribute values (machine-readable markers for JS/testing)
- Hidden `<input>` values (form submission payloads, not displayed)
- JSON API responses consumed by developer tools (not end-user UI)

---

## Backend: UserDescriptor / SubjectDescriptor

All approval responses that reference a user or business subject now include a structured descriptor:

```json
{
  "id": "3fa85f64-...",
  "username": "alice",
  "display_name": "Alice Nakamura"
}
```

```json
{
  "type": "RESEARCH_REPORT",
  "id": "7d8e1234-...",
  "display_label": "Research: TH-BH 2026-06 Q2 Outlook"
}
```

The raw UUID fields (`submitter_id`, `actor_user_id`, `signer_user_id`) are retained as routing identifiers but **must never be used as display values in the frontend**.

**If `display_name` is blank** (user cannot be resolved), the backend returns `"Unknown User"`. The frontend must propagate this rather than falling back to the raw ID.

---

## Frontend: Safe Display Patterns

### Users
```vue
<!-- CORRECT: use descriptor -->
{{ event.actor?.display_name || event.actor_name || 'System' }}

<!-- WRONG: falls back to UUID -->
{{ event.actor_name || event.actor_user_id || 'System' }}
```

### Subjects
```vue
<!-- CORRECT: use subject descriptor -->
{{ request.subject?.display_label || request.subject_title || '—' }}

<!-- WRONG: exposes raw ID -->
{{ request.subject_title || request.subject_id }}
```

### Modules / Contracts
```vue
<!-- CORRECT: use business type code -->
{{ req.contract_type || req.process_type || 'SYSTEM' }}

<!-- WRONG: UUIDs appear when contract_id resolves but contract_type is empty -->
{{ req.contract_id || 'SYSTEM' }}
```

### Delegated-from / Proxy (`principalUserName`)

`principalUserName` drives the "for \<name\>" tag in `ApprovalStageCard.vue`. It must be `undefined` for non-delegated approvers and a human-readable name for delegated ones. Use `derivePrincipalName` from `approvalMappers.ts`:

```ts
// CORRECT: use derivePrincipalName — never returns a UUID
import { derivePrincipalName } from '../lib/approvalMappers';

principalUserName: derivePrincipalName(sig.is_proxy_signature, sig.proxy_for),
// → "Alice Nakamura" when resolved, "Unknown User" when not, undefined when not delegated

// WRONG: passes raw UUID as principalUserName
principalUserName: sig.proxy_for_user_id,          // raw UUID
principalUserName: tsk.delegated_from_user_id,     // raw UUID
```

```vue
<!-- CORRECT: descriptor -->
{{ delegated_from?.display_name }}

<!-- WRONG: raw UUID in metadata -->
{{ delegated_from_user_id }}
```

### Form Inputs — Member Selectors

When no user-search autocomplete is available, form fields that previously accepted raw UUID input must use a disabled input with the placeholder "User selector source not available". Never prompt an operator to type a UUID.

```vue
<!-- CORRECT -->
<AppFormField label="User">
  <AppInput placeholder="User selector source not available" disabled />
</AppFormField>

<!-- WRONG -->
<AppFormField label="User UUID">
  <AppInput placeholder="e.g. 00000000-0000..." />
</AppFormField>
```

---

## Regression Tests

### `frontend/tests/approval-no-uuid.test.ts`

Covers `toTimelineEvents`, `toStamps`, and `derivePrincipalName` functions, plus label/placeholder constant assertions:
- Verifies `actor` display uses `actor?.display_name` → `actor_name` → `"System"` (never UUID).
- Verifies `signer` display uses `signer?.display_name` → `signer_display_name` → `"Unknown User"`.
- Verifies `delegated_from` metadata uses `display_name`, not raw UUID.
- Verifies `derivePrincipalName` returns `undefined` when not delegated, descriptor's `display_name` when present, `"Unknown User"` when absent.
- Verifies `ApprovalGroupMemberTable` and `ApprovalTeamMemberTable` label/placeholder values don't contain "UUID".
- Uses an `assertNoUUID()` helper that throws if a UUID pattern appears in any label field.

### `frontend/tests/permission-no-uuid.test.ts`

Covers permission module display logic:
- Creator label: `created_by_name || "Unknown User"` (never `created_by` UUID).
- Target entity: type-only display, no `target_entity_id` UUID.

### Running the tests
```bash
cd frontend && npm run test
```

All 177 tests must pass, including the UUID-focused tests.

---

## How to Add a New Display Path Safely

When adding a new component that renders user or entity data:

1. **Check if the API response has a descriptor field** (`user_descriptor`, `submitter`, `actor`, `signer`, etc.).
2. **Use `display_name`** from the descriptor as the primary label.
3. **Fall back to a static string** like `"Unknown User"` or `"—"` when the descriptor is absent or its `display_name` is blank.
4. **Never fall back to an `_id` or `_user_id` field** as a display value.
5. **Write a test** that calls `assertNoUUID()` on the output of your mapping function.

---

## Fields That Still Contain UUIDs (Internal Use Only)

These fields exist for routing and must stay hidden from display:

| Field | Purpose |
|---|---|
| `request.id` | Used in `<NuxtLink :to="'/approval/requests/' + request.id">` |
| `request.submitter_id` | Used to check `req.submitter_id === currentUserId` (not displayed) |
| `task.assigned_user_id` | Used to route delegation checks |
| `request.contract_id` | Used in filter predicates (`x.contract_id === filterModule.value`) — never displayed |
| `event.actor_user_id` | Kept for API consumers; never rendered as text |
| `sig.proxy_for_user_id` | Route/lookup ID; display via `derivePrincipalName(sig.is_proxy_signature, sig.proxy_for)` |
| `tsk.delegated_from_user_id` | Route/lookup ID; display via `derivePrincipalName(tsk.is_delegated_action, tsk.delegated_from)` |
| `member.user_id` | Group/team member routing ID; display via `member.display_name` or `member.user_display_name` |
