# Approval Subject Access Port — Design Reference

**As of:** 2026-06-17

The approval engine is generic and has no knowledge of which business objects it governs. Subject-specific authorization (data-permission checks, object-level access) is delegated to owning modules via the `SubjectAccessPort` pattern.

---

## Interface

```go
// backend/pkg/contract/approval.go
type SubjectAccessor interface {
    CanViewApprovalSubject(ctx context.Context, actorID uuid.UUID, subjectType string, subjectID uuid.UUID) error
    CanSubmitApprovalSubject(ctx context.Context, actorID uuid.UUID, subjectType string, subjectID uuid.UUID) error
    CanActOnApprovalSubject(ctx context.Context, actorID uuid.UUID, subjectType string, subjectID uuid.UUID, action string) error
}
```

Each method returns `nil` for allow, non-nil for deny.

The approval engine's internal port wraps this interface:
```go
// backend/internal/approval/domain/ports.go
type ApprovalSubjectAccessPort interface {
    CanViewApprovalSubject(ctx context.Context, actorID uuid.UUID, st SubjectType, subjectID uuid.UUID) error
    CanSubmitApprovalSubject(ctx context.Context, actorID uuid.UUID, st SubjectType, subjectID uuid.UUID) error
    CanActOnApprovalSubject(ctx context.Context, actorID uuid.UUID, st SubjectType, subjectID uuid.UUID, action ApprovalAction) error
}
```

---

## Registration (main.go)

```go
investSubjectAccessor := investmentModule.SubjectAccessor(iamModule)
if investSubjectAccessor != nil {
    approvalModule.RegisterSubjectAccessPort("RESEARCH_REPORT", investSubjectAccessor)
    approvalModule.RegisterSubjectAccessPort("INVESTMENT_DECISION", investSubjectAccessor)
    // PORTFOLIO: intentionally NOT registered — see "PORTFOLIO Deferred" section below
}
```

`SubjectAccessor(iamModule)` is defined on `investment.Module` and returns an `*InvestmentSubjectAccessor` that delegates to the IAM data-permission check.

---

## Fail-Closed Guarantee

If no port is registered for a subject type, **all reads fail with ErrForbidden**. This prevents a misconfigured deployment from silently disclosing data.

```go
// runtime_service.go
func (s *ApprovalRuntimeService) checkSubjectView(ctx context.Context, actorID uuid.UUID, st, subjectID) error {
    port, ok := s.subjectAccessPorts[st]
    if !ok {
        return domain.Forbidden("no subject access port configured for subject type: " + string(st))
    }
    return port.CanViewApprovalSubject(ctx, actorID, st, subjectID)
}
```

Fail-closed means: on any error (transient IAM failure, nil dependency, missing subject), the accessor returns an opaque `"not found or not accessible"` error — never allows through.

---

## Investment Subject Accessor Implementation

**File:** `backend/internal/investment/infrastructure/adapter/investment_subject_access.go`

The accessor:
1. Calls `resolveContractID(ctx, subjectType, subjectID)` to map the business object to its associated contract UUID.
2. If `contractID == uuid.Nil`, the subject has no data-scope restriction — access is allowed.
3. Otherwise, calls `a.iam.HasDataPermission(ctx, actorID.String(), contractID.String())`.
4. Any error or `!ok` result → deny.

```
RESEARCH_REPORT     → research.GetByID(subjectID) → r.ApplicableContractID
INVESTMENT_DECISION → decisions.GetByID(subjectID) → d.ContractID
PORTFOLIO           → NOT SUPPORTED — switch hits default arm → "unsupported subject type" error
                      (this is why PORTFOLIO is not registered — see section below)
```

---

## PORTFOLIO Approval — Deferred (Option A, 2026-06-17)

**Status:** PORTFOLIO is NOT registered as an active approval subject. The `PortfolioApprovalCallback` method exists on the investment module but is not wired.

**Reason:** `InvestmentSubjectAccessor.resolveContractID` has no `case "PORTFOLIO":` arm. Registering the access port while the default arm denies all access would mean every PORTFOLIO approval attempt returns "not found or not accessible." A wired-but-broken subject is worse than a deferred one.

**To implement PORTFOLIO approval:**

1. Add `portfolios domain.PortfolioRepository` field to `InvestmentSubjectAccessor` (investment_subject_access.go)
2. Update `NewInvestmentSubjectAccessor` constructor to accept and store the portfolio repository
3. Add to `resolveContractID`:
   ```go
   case "PORTFOLIO":
       if a.portfolios == nil {
           return uuid.Nil, errors.New("portfolio repository unavailable")
       }
       p, err := a.portfolios.GetByID(ctx, subjectID)
       if err != nil { return uuid.Nil, err }
       if p == nil { return uuid.Nil, errors.New("portfolio not found") }
       return p.ContractID, nil  // or uuid.Nil if portfolio is the boundary
   ```
4. Implement `PortfolioSubjectValidator` implementing `ApprovalSubjectValidator`
5. In `main.go`: uncomment/add the three PORTFOLIO registrations
6. Update seed `013_approval_portfolio_onboarding_seed.sql` to set `is_active = true`
7. Add PORTFOLIO tests to `investment_subject_access_test.go`

**Do not register a subject type until all three are ready:** access port, descriptor callback, and validator.

---

## Adding a New Subject Type

1. Add the subject type to `vo.SubjectType` enum in `valueobject/enums.go`.
2. Implement the business access check in the owning module's adapter.
3. Register the port in `main.go`:
   ```go
   approvalModule.RegisterSubjectAccessPort("MY_SUBJECT_TYPE", myModuleAccessor)
   ```
4. Add a test in `runtime_service_test.go` covering: port missing → fail closed, deny → returns NotFound, allow → data visible.

---

## Security Notes

- **Existence oracle:** When subject access fails, `GetApprovalRequest` and `GetApprovalTimeline` return `ErrNotFound` (not `ErrForbidden`) so request existence cannot be inferred by probing the error type.
- **`GetSubjectApprovalStatus`** returns `(nil, nil)` on access denial, meaning the caller sees "no active request" regardless of whether one exists.
- **`ListRequests`** with `ViewerID != uuid.Nil` post-filters every result through the access port. Results are filtered silently rather than erroring, to avoid count-based leakage.
