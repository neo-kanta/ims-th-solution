# P0 Delivery Report — Permission + Approval Module

**Date:** 2026-06-17  
**Scope:** 7 P0 blockers from Codex commit review (Permission + Approval module)  
**Status:** ALL P0s FIXED, TESTED, DOCUMENTED — READY FOR REVIEW

---

## Summary

All 7 P0 blockers are resolved. No regressions. Build is clean.

| Test Suite | Count | Status |
|---|---|---|
| Backend (`go test ./...`) | 40+ | PASS |
| Frontend (`npm run test`) | 177 | PASS |
| Frontend build (`npm run build`) | — | CLEAN |

---

## P0 Fix Summary

### P0-1 — `ListRequests` leaks all approvals when actor context missing

**Root cause:** `ListRequests` discarded the `ok` return of `actorID(r)`. A request with no JWT context produced `actor = uuid.Nil`, which the service treats as a system bypass → all requests returned unfiltered.

**Fix:** Added 401 guard before service call:
```go
actor, ok := actorID(r)
if !ok {
    httputil.Unauthorized(w, "not authenticated")
    return
}
```
**File:** `backend/internal/approval/transport/handler/runtime_handler.go`  
**Test:** `TestListRequests_NoActor_Returns401` in `runtime_handler_test.go`

---

### P0-2 — `ApprovalGroupMemberTable` shows "User UUID" label and renders `m.user_id`

**Root cause:** Member rows had `<span class="member-table__uid">{{ m.user_id }}</span>`. Add-member form used `<AppFormField label="User UUID">` with free-text UUID input.

**Fix:**
- Removed UUID span from member rows
- Changed form label to `"User"`, disabled input with `placeholder="User selector source not available"`

**File:** `frontend/app/features/approval/components/ApprovalGroupMemberTable.vue`  
**Test:** `ApprovalGroupMemberTable – UUID-free label and placeholder` suite (4 tests) in `approval-no-uuid.test.ts`

---

### P0-3 — `ApprovalTeamMemberTable` shows "User UUID" label and renders `m.user_id`

**Root cause:** Member rows had `<span class="tm-table__uid">{{ m.user_id }}</span>`. Add-member input had `placeholder="User UUID"`.

**Fix:**
- Removed UUID span from member rows
- Changed placeholder to `"User selector source not available"`, set `disabled`

**File:** `frontend/app/features/approval/components/ApprovalTeamMemberTable.vue`  
**Test:** `ApprovalTeamMemberTable – UUID-free placeholder` suite (3 tests) in `approval-no-uuid.test.ts`

---

### P0-4 — Delegation/stage UI passes raw `proxy_for_user_id` as `principalUserName`

**Root cause:** Six assignment sites in `useApprovalSession.ts` set `principalUserName` to raw UUID fields (`sig.proxy_for_user_id`, `tsk.delegated_from_user_id`, `evt.delegated_from_user_id`). These values rendered in `ApprovalStageCard.vue` as `for {{ approver.principalUserName }}`.

**Fix:** Extracted pure function `derivePrincipalName(isDelegated, descriptor)` from `approvalMappers.ts`. Returns:
- `undefined` — when no delegation is in effect (hides the "for" tag via `v-if`)
- `descriptor.display_name` — when delegation present and resolved
- `"Unknown User"` — when delegation present but name unresolved

All six assignment sites now use `derivePrincipalName`.

**Files:**
- `frontend/app/features/approval/lib/approvalMappers.ts` — new `derivePrincipalName` function
- `frontend/app/features/approval/composables/useApprovalSession.ts` — all 6 sites updated

**Tests:** `derivePrincipalName – delegation display is UUID-free` suite (6 tests) in `approval-no-uuid.test.ts`

---

### P0-5 — Signature fallback uses `signer.String()`, persisting raw UUID when directory lookup fails

**Root cause:** `writeSignature()` initialized `display := signer.String()`. When directory lookup failed, the UUID was stored as `SignerDisplayName` in `approval__signature_records`. `userDescriptor()` passes non-empty names through as-is, so the UUID reached the API response.

**Fix:**
```go
// was: display := signer.String()
display := "Unknown User"
```

**File:** `backend/internal/approval/application/service/runtime_service.go`  
**Tests:**
- `TestWriteSignature_NilDirectory_SignerDisplayNameIsUnknownUser`
- `TestWriteSignature_ErrorDirectory_SignerDisplayNameIsUnknownUser`

in `runtime_service_test.go`

---

### P0-6 — Docs contradict Option A PORTFOLIO decision and current final-stage/revoke behavior

**Root cause:** Three stale claims in handoff docs:
1. P0-4 row claimed "Zero final stages → auto-mark" — code actually returns a validation error
2. P0-1 row incorrectly listed PORTFOLIO as registered
3. No documentation of `subjectAllowedActions` still emitting `"revoke"` while `computeAllowedActions` (the detail view source) does not

**Fix:**
- `approval-module.md` §12: corrected P0-4 row, corrected P0-1 row, added P0-11 through P0-21, updated P0-16 with explicit two-function clarification
- `approval-module.md` §14: added "Revoke not surfaced in UI" and "User selector for group/team members" Known Gaps
- `approval-permission-p0-fixes.md`: added full P0-17 through P0-21 sections, updated verification checklist, updated remaining limitations table, added two-function revoke clarification to P0-16
- `frontend-uuid-free-policy.md`: added `principalUserName`/delegation pattern, Form Inputs section, updated test count, added new UUID-internal fields

---

### P0-7 — No regression tests covering these P0 issues

**Root cause:** No tests existed for the specific failure modes: ListRequests 401 bypass, member UUID display, principalUserName UUID, or signature UUID fallback.

**Fix — new tests:**

| File | Suite / Test | Covers |
|---|---|---|
| `backend/internal/approval/transport/handler/runtime_handler_test.go` | `TestListRequests_NoActor_Returns401` | P0-1 / P0-17 handler guard |
| `backend/internal/approval/application/service/runtime_service_test.go` | `TestWriteSignature_NilDirectory_*` | P0-5 / P0-21 signature UUID |
| `backend/internal/approval/application/service/runtime_service_test.go` | `TestWriteSignature_ErrorDirectory_*` | P0-5 / P0-21 error path |
| `frontend/tests/approval-no-uuid.test.ts` | `derivePrincipalName – delegation display is UUID-free` (6 tests) | P0-4 / P0-20 |
| `frontend/tests/approval-no-uuid.test.ts` | `ApprovalGroupMemberTable – UUID-free label and placeholder` (4 tests) | P0-2 / P0-18 |
| `frontend/tests/approval-no-uuid.test.ts` | `ApprovalTeamMemberTable – UUID-free placeholder` (3 tests) | P0-3 / P0-19 |

Total new tests: 3 backend + 13 frontend = 16 new regression guards.

---

## Files Changed

### Backend
| File | Change |
|---|---|
| `backend/internal/approval/transport/handler/runtime_handler.go` | Added 401 guard in `ListRequests` |
| `backend/internal/approval/application/service/runtime_service.go` | `display := "Unknown User"` in `writeSignature()` |
| `backend/internal/approval/transport/handler/runtime_handler_test.go` | NEW — 401 guard test |
| `backend/internal/approval/application/service/runtime_service_test.go` | Added 2 signature fallback tests + helpers |

### Frontend
| File | Change |
|---|---|
| `frontend/app/features/approval/components/ApprovalGroupMemberTable.vue` | Removed UUID span, disabled member input |
| `frontend/app/features/approval/components/ApprovalTeamMemberTable.vue` | Removed UUID span, disabled member input |
| `frontend/app/features/approval/lib/approvalMappers.ts` | Added `derivePrincipalName` export |
| `frontend/app/features/approval/composables/useApprovalSession.ts` | All 6 `principalUserName` sites use `derivePrincipalName` |
| `frontend/tests/approval-no-uuid.test.ts` | Added 3 new test suites (13 tests) |

### Docs
| File | Change |
|---|---|
| `docs/handoff/approval-module.md` | §12 and §14 updated; revoke two-function clarification added |
| `docs/handoff/approval-permission-p0-fixes.md` | P0-16 through P0-21 sections added; revoke clarification added |
| `docs/handoff/frontend-uuid-free-policy.md` | Delegation pattern, member selectors, test count updated |
| `docs/handoff/p0-delivery-report.md` | THIS FILE |

---

## Controlled Demo Scope

| Subject Type | Status |
|---|---|
| RESEARCH_REPORT approval | Registered, tested |
| INVESTMENT_DECISION approval | Registered, tested |
| Permission request flow | Covered by existing + new tests |
| PORTFOLIO | Deferred (Option A) — not registered in `main.go` |

---

## Known P1 Items (not in scope, documented)

| Item | Notes |
|---|---|
| Revoke button not surfaced | `computeAllowedActions` omits "revoke"; route wired. `subjectAllowedActions` still emits it but is not consumed by the detail view UI. |
| User selector for group/team members | Add-member inputs disabled; real autocomplete needed |
| PORTFOLIO full implementation | Deferred; access port and callback exist but are not wired |
| Old signature rows with UUID display name | Pre-fix rows retain UUID; no migration planned (fresh-demo-DB) |

---

## Verification Commands

```bash
# Backend
cd backend && go build ./...
cd backend && go test ./internal/approval/...
cd backend && go test ./...

# Frontend
cd frontend && npm run test      # must show 177 passed
cd frontend && npm run build     # must be clean (no errors)
```

**Verified 2026-06-17:** All commands pass.
