# Approval + Permission User-Facing DTO Reference

**As of:** 2026-06-17  
**Rule:** No raw UUID may appear as visible text in normal Approval or Permission UI.

---

## No-UUID Policy

### Allowed visible values

| Value type | Example |
|---|---|
| username | `alice.nakamura` |
| accountCode | `U001` |
| displayName | `Alice Nakamura` |
| group code/name | `FUND_MANAGER_REVIEWERS` |
| team code/name | `INVESTMENT_TEAM` |
| subject number | `RPT-2026-042` |
| subject title | `Q2 Equity Analysis` |
| business code/name | `KTB-EQUITY` |
| displayLabel | `RPT-2026-042 — Q2 Equity Analysis` |
| "Unknown User" | fallback when name unavailable |
| "Deleted User" | fallback for removed actors |
| "Unknown Subject" | fallback when subject descriptor unavailable |

### Forbidden visible values

Raw UUIDs of any kind: user UUID, actor UUID, signer UUID, delegated user UUID, target_id, contract_id, fund_id, portfolio_id.

---

## UserDescriptor Shape

Returned by approval/permission endpoints for any user reference.

```json
{
  "user_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "username": "alice.nakamura",
  "display_name": "Alice Nakamura",
  "account_code": "U001"
}
```

Frontend rendering rule:
1. Show `display_name` if non-empty
2. Else show `username`
3. Else show `account_code`
4. Else show "Unknown User"

Never show `user_id`.

---

## SubjectDescriptor Shape

Returned in approval request responses as `subject`.

```json
{
  "subject_type": "RESEARCH_REPORT",
  "subject_id": "ab9cc8e4-61e8-4b51-b632-a71cb3162e6a",
  "subject_number": "RPT-2026-042",
  "subject_title": "Q2 Equity Analysis",
  "display_label": "RPT-2026-042 — Q2 Equity Analysis"
}
```

Frontend rendering rule:
1. Show `display_label` if non-empty
2. Else show `subject_number + subject_title`
3. Else show "Unknown Subject"

Never show `subject_id`.

---

## TargetDescriptor Shape (Permission module)

The `ChangeItem` in permission requests currently exposes `target_table` and `target_id`.

**Policy (P0 FIX 5):** `target_id` must not be shown as visible text.

Current rendering after fix (`PermissionRequestDetailScreen.vue`):
```vue
<span>{{ item.target_table }}</span>
```

Future improvement (P1): Add `target_display_label` to the backend DTO:
```json
{
  "target_table": "iam__function_permissions",
  "target_id": "<uuid>",
  "target_display_label": "Investment.Reports.View (iam__function_permissions)"
}
```

Frontend fallback chain:
1. `item.target_display_label` if present
2. Else `item.target_name` or `item.target_code`
3. Else "Unknown Subject"

---

## Before / After Examples

### Approval request detail — before

```json
{
  "submitter_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "subject_id": "ab9cc8e4-61e8-4b51-b632-a71cb3162e6a"
}
```
Frontend would render two raw UUIDs.

### Approval request detail — after

```json
{
  "submitter": {
    "user_id": "3fa85f64-...",
    "username": "alice.nakamura",
    "display_name": "Alice Nakamura"
  },
  "subject": {
    "subject_type": "RESEARCH_REPORT",
    "subject_id": "ab9cc8e4-...",
    "subject_number": "RPT-2026-042",
    "subject_title": "Q2 Equity Analysis",
    "display_label": "RPT-2026-042 — Q2 Equity Analysis"
  }
}
```
Frontend renders `Alice Nakamura` and `RPT-2026-042 — Q2 Equity Analysis`.

### Permission change item — before

```html
<span>iam__function_permissions 3fa85f64-5717-4562-b3fc-2c963f66afa6</span>
```

### Permission change item — after

```html
<span>iam__function_permissions</span>
```

---

## Approval Timeline / Signature Stamps — Fallback Chain

From `approvalMappers.ts`:

```ts
// Actor (event)
actor?.display_name || actor_name || "System"

// Signer (signature stamp)
signer?.display_name || signer_display_name || "Unknown User"
```

These fallbacks ensure the timeline always shows a human-readable name.
