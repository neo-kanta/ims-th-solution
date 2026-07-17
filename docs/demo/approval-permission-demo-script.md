# Approval + Permission Controlled Demo Script

**As of:** 2026-06-17  
**Verdict:** READY FOR CONTROLLED DEMO (RESEARCH_REPORT + INVESTMENT_DECISION)

---

## Pre-Demo Checklist

- [ ] Backend running and healthy (`GET /health`)
- [ ] Database seeded (`make seed`)
- [ ] Login credentials for two roles ready: maker and checker
- [ ] Browser incognito windows open for each role

---

## Demo Flow 1 — Research Report Approval

### Step 1: Login as maker

Login as a user with `INVESTMENT_ANALYSIS_REPORT` submission rights (e.g., `alice.nakamura` / Portfolio Manager role).

### Step 2: Submit a research report for approval

1. Navigate to Investment → Research Reports
2. Select or create a research report
3. Submit for approval — approval request is created automatically
4. Note the request number (e.g., `RPT-2026-042`)

**Verify:** Maker cannot see Approve/Reject buttons on their own submission.

### Step 3: Login as checker

In a second browser window, login as a user in the approval group for stage 1 (e.g., a Fund Manager Reviewer).

### Step 4: Open approval inbox

Navigate to Approvals → Inbox.

**Verify:**
- Request appears with `subject_title` (e.g., "Q2 Equity Analysis"), not a raw UUID
- Submitter shown as display name (e.g., "Alice Nakamura"), not user UUID
- Request number visible (e.g., "RPT-2026-042")

### Step 5: Open the request detail

Click the request.

**Verify:**
- Subject descriptor shows report title and number — no raw UUID
- Timeline / stamps show actor display names — no raw UUID
- Approve and Reject buttons are visible for the current stage approver

### Step 6: Approve with reason

Click Approve, enter a reason, confirm.

**Verify:**
- Timeline updates with approver's display name and timestamp
- If multi-stage: request moves to next stage
- If final stage: request moves to APPROVED status

---

## Demo Flow 2 — Investment Decision Approval

Same flow as above, using an INVESTMENT_DECISION subject type.

Approval process type: `INVESTMENT_DECISION`  
Trigger: Decision Command Handler submit  
Inbox: same approval inbox

---

## Demo Flow 3 — Permission Request

### Step 1: Login as a permission change requestor

### Step 2: Create a permission change request

Navigate to Permissions → Change Requests → New Request.

### Step 3: Add permission change items

Use the Changes tab to add/modify permission entries.

**Verify in the Changes tab:**
- Item header shows `target_table` (e.g., `iam__function_permissions`) — no raw UUID
- `target_id` is not visible as text

### Step 4: Submit and route for approval

Submit the request. An approver reviews and approves it.

**Verify:**
- Creator shown as display name, not UUID
- No raw UUIDs visible in any field

---

## Approval Config Demo (if needed)

Navigate to Approvals → Process Config → Create Process.

**Verify:**
- "Applicable Scope" field shows "Scope selector source not available" (disabled) — no UUID input
- Stage builder SINGLE_USER mode shows "Approver" field as "User selector source not available" (disabled) — no UUID input

---

## Flows to AVOID in Demo

| Flow | Reason |
|---|---|
| PORTFOLIO approval | Deferred — not wired; all access would fail |
| Compliance / IRG review | Not ready |
| Production audit claim | Signature stamp immutability triggers not yet on `approval__signature_records` (P1) |
| Action-specific function rights | Permissions fetcher uses any-flag SQL, not per-action (P1) |
| Revoke button demo | Revoke not in `allowed_actions`; route-level gate only (P1) |
| DelegatedFrom actor name | Always shows "Unknown User" — name join missing from tasks SQL (P1) |

---

## Known Limitations (Safe to Mention)

- **Revoke:** The revoke API is functional (route middleware enforces permission), but the UI does not surface a Revoke button via `allowed_actions`. This is a P1 improvement.
- **PORTFOLIO approval:** Fully deferred. The infrastructure exists but is intentionally not wired.
- **Email notifications:** Notifier is wired but no-op in demo environment — in-app notifications work.

---

## Verdict

- Controlled demo: **SAFE** for RESEARCH_REPORT and INVESTMENT_DECISION approval flows, and Permission request flows.
- Compliance / IRG claim: **NOT YET** — P1 items above must be resolved first.
