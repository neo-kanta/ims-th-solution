package command

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

const (
	// cashApprovalProcessType/SubjectType identify a LIVE cash movement in the
	// shared approval engine. The approval's SubjectID is the cash request id.
	cashApprovalProcessType = "PORTFOLIO_CASH_TRANSACTION"
	cashApprovalSubjectType = "CASH_TRANSACTION"

	// maxIdempotencyKeyLen bounds the client-supplied key to the column width so
	// an over-long header is rejected at the boundary (400) rather than failing
	// the INSERT with a scrubbed 500.
	maxIdempotencyKeyLen = 255

	// derivedKeyPrefix namespaces a SERVER-derived idempotency key
	// ("auto:" || fingerprint) so a no-key request still dedups via the partial
	// unique index on (portfolio_id, idempotency_key). It is reserved: a client
	// key beginning with this prefix is rejected so a client can never forge a
	// collision with a derived key. "auto:" + 64 hex = 69 chars ≤ column width.
	derivedKeyPrefix = "auto:"

	// fingerprintFieldSep delimits canonical fingerprint fields. It is a unit
	// separator (0x1f) — not producible from a UUID / decimal / ISO code / date,
	// and all such fixed-format fields precede the free-text memo, so no memo
	// content can shift the field layout.
	fingerprintFieldSep = "\x1f"
)

// LIVE cash-transaction approval gate (IMS-PORTFOLIO-FUND-OPTIONAL Stage 2).
//
// A LIVE-portfolio cash movement (CASH_IN/CASH_OUT/FEE/DIVIDEND) does not post
// directly to the append-only ledger. Instead it is staged in the mutable
// investment__portfolio_cash_requests table (PENDING) and pushed into the
// shared approval engine. The real ledger row is materialized only after
// approval, inside ApplyCashRequestApproval (invoked by the approval callback).

// submitCashForApproval stages a LIVE cash movement for approval. It is called
// from Handle after prepareTransaction has already validated the request and
// run the pre-post preconditions (EvaluatePost), so a movement that would be
// rejected outright is caught at submit time rather than approval time — the
// authoritative re-check ALSO runs at materialization.
func (h *PostTransactionHandler) submitCashForApproval(
	ctx context.Context,
	req PostTransactionRequest,
	prep *preparedTransaction,
) (*PostTransactionResult, error) {
	// Fail CLOSED: a LIVE cash movement must never post immediately when the
	// approval gate is not wired — that would bypass the required approval.
	if h.cashRequests == nil || h.approval == nil {
		return nil, &domain.ErrPostPreconditionFailed{
			Violation: string(policy.PostViolationApprovalUnavailable),
			Detail:    "LIVE cash movements require the approval engine, which is not configured",
		}
	}
	// The staged amount is the positive magnitude; sign is derived from the
	// transaction type at materialization. prep.gross is the client-supplied
	// gross_amount for a cash movement (deriveAmounts).
	if prep.gross.Sign() <= 0 {
		return nil, &domain.ErrInvalidDecisionRequest{
			Field:  "gross_amount",
			Detail: "a positive amount is required for a cash movement",
		}
	}

	// Normalise + validate the optional idempotency key at the boundary.
	key := strings.TrimSpace(req.IdempotencyKey)
	if len(key) > maxIdempotencyKeyLen {
		return nil, &domain.ErrInvalidDecisionRequest{
			Field:  "idempotency_key",
			Detail: fmt.Sprintf("must be at most %d characters", maxIdempotencyKeyLen),
		}
	}
	if strings.HasPrefix(key, derivedKeyPrefix) {
		// Reserved for server-derived keys — reject so a client cannot forge a
		// collision with the no-key dedup path.
		return nil, &domain.ErrInvalidDecisionRequest{
			Field:  "idempotency_key",
			Detail: fmt.Sprintf("must not start with the reserved %q prefix", derivedKeyPrefix),
		}
	}

	// Canonical request fingerprint (Finding 3). Deterministic over the payload,
	// so a genuine retry produces the SAME fingerprint even when it carries no
	// client key. Persisted once and never mutated.
	fingerprint := computeCashRequestFingerprint(
		req.ActorID, prep.portfolio.ID, req.TransactionType,
		prep.gross, req.Fees, req.Currency, req.BusinessDate, req.Reason,
	)

	// Effective dedup key: the client key when supplied, otherwise a server-
	// derived key over the fingerprint so a no-key retry of the identical request
	// still collapses onto the original via the partial unique index. A different
	// no-key payload has a different fingerprint → different derived key → a new
	// request, preserving the "no-key is not cross-payload dedup" behaviour.
	effectiveKey := key
	if effectiveKey == "" {
		effectiveKey = derivedKeyPrefix + fingerprint
	}

	// Idempotent replay: a prior request already staged for this (portfolio,
	// effective key)? Return / reconcile it instead of creating a duplicate.
	existing, err := h.cashRequests.GetByIdempotencyKey(ctx, prep.portfolio.ID, effectiveKey)
	if err != nil {
		return nil, fmt.Errorf("looking up cash request by idempotency key: %w", err)
	}
	if existing != nil {
		return h.resumeCashRequest(ctx, req, prep, existing, fingerprint)
	}

	now := h.now()
	cr := &entity.PortfolioCashRequest{
		ID:                 h.nextID(),
		PortfolioID:        prep.portfolio.ID,
		FundID:             prep.portfolio.FundID,
		TransactionType:    req.TransactionType,
		Amount:             prep.gross,
		Currency:           req.Currency,
		Fees:               req.Fees,
		ValueDate:          req.BusinessDate,
		Memo:               req.Reason,
		Status:             vo.CashRequestStatusPending,
		IdempotencyKey:     strPtr(effectiveKey),
		RequestFingerprint: strPtr(fingerprint),
		SubmittedBy:        req.ActorID,
		SubmittedAt:        now,
		Version:            1,
		CreatedAt:          now,
		UpdatedAt:          now,
		CreatedBy:          req.ActorID,
		UpdatedBy:          req.ActorID,
	}

	// 1. Persist the PENDING request first — the approval engine's subject
	//    access port resolves the subject by id during SubmitForApproval, so the
	//    row must exist before we submit.
	if err := h.runTransaction(ctx, func(dbtx pgx.Tx) error {
		return h.cashRequests.Create(ctx, dbtx, cr)
	}); err != nil {
		// A concurrent retry with the same effective key won the insert race
		// (partial unique index on (portfolio_id, idempotency_key)). Resolve to
		// the winner rather than surfacing a duplicate-key 500. This covers both
		// the client-key and the no-key (derived-key) paths.
		if isUniqueViolation(err) {
			existing, gErr := h.cashRequests.GetByIdempotencyKey(ctx, prep.portfolio.ID, effectiveKey)
			if gErr != nil {
				return nil, fmt.Errorf("resolving idempotent cash request after unique violation: %w", gErr)
			}
			if existing != nil {
				return h.resumeCashRequest(ctx, req, prep, existing, fingerprint)
			}
		}
		return nil, fmt.Errorf("creating cash request: %w", err)
	}

	// 2 + 3. Fresh row → submit to the approval engine and record its id. No
	// prior approval can exist for a row we just inserted, so reconcile=false.
	return h.submitAndLinkCashApproval(ctx, req, prep, cr, false)
}

// resumeCashRequest returns or reconciles an existing cash request found via its
// idempotency key, making a retried submit a no-op on already-created state.
func (h *PostTransactionHandler) resumeCashRequest(
	ctx context.Context,
	req PostTransactionRequest,
	prep *preparedTransaction,
	cr *entity.PortfolioCashRequest,
	fingerprint string,
) (*PostTransactionResult, error) {
	// Cross-actor key reuse: a client key belongs to the submitter that first
	// used it. A DIFFERENT actor reusing the same (portfolio, key) must be
	// rejected — never resume or disclose another user's request. (On the no-key
	// derived-key path the actor is baked into the fingerprint, so a different
	// actor produces a different derived key and never reaches this row; the
	// check is a belt-and-suspenders guard there.)
	if cr.SubmittedBy != req.ActorID {
		return nil, &domain.ErrInvalidDecisionRequest{
			Field:  "idempotency_key",
			Detail: "key was already used by another user; use a new key",
		}
	}

	// Idempotency-key safety: the same key must identify the SAME movement. A
	// canonical-fingerprint mismatch (different amount/type/fees/currency/value-
	// date/memo) must NOT silently return the original request (a money-path
	// hazard) — reject it so the client uses a fresh key.
	if !cashRequestMatchesPayload(cr, req, prep, fingerprint) {
		return nil, &domain.ErrInvalidDecisionRequest{
			Field:  "idempotency_key",
			Detail: "key was already used for a different cash movement; use a new key",
		}
	}

	// Already linked to an approval, or in a terminal state → idempotent replay
	// of the original response. Always render the cash request (Pending=true) so
	// the transport returns the same 202 cash-request body it returned the first
	// time, never a nil transaction.
	if cr.ApprovalRequestID != nil || cr.Status.IsTerminal() {
		return &PostTransactionResult{CashRequest: cr, Pending: true}, nil
	}
	// PENDING with a NULL approval_request_id: a prior attempt created the row
	// but failed to record an approval id (a step-3 / link failure). Reconcile —
	// re-link to the approval it already submitted if one exists, else submit.
	return h.submitAndLinkCashApproval(ctx, req, prep, cr, true)
}

// cashRequestMatchesPayload reports whether a stored request describes the SAME
// movement as the incoming request. It compares the canonical fingerprint when
// one is stored (the normal path) and falls back to a field-by-field comparison
// for legacy rows persisted before the fingerprint column existed.
func cashRequestMatchesPayload(
	cr *entity.PortfolioCashRequest,
	req PostTransactionRequest,
	prep *preparedTransaction,
	fingerprint string,
) bool {
	if cr.RequestFingerprint != nil {
		return *cr.RequestFingerprint == fingerprint
	}
	return cr.TransactionType == req.TransactionType &&
		cr.Amount.Equal(prep.gross) &&
		cr.Fees.Equal(req.Fees) &&
		cr.Currency == req.Currency
}

// computeCashRequestFingerprint returns a deterministic hex sha256 over the
// canonical LIVE cash-movement payload. The field set and order are fixed:
//
//	submitter_id | portfolio_id | transaction_type | amount | fees | currency |
//	value_date (UTC yyyy-mm-dd) | memo (trimmed)
//
// Money values use StringFixed(8) to match the DECIMAL(28,8) column so an honest
// retry that sent "1000" vs "1000.00" hashes identically. The fields are joined
// with a 0x1f unit separator that no fixed-format field can contain.
func computeCashRequestFingerprint(
	submitterID, portfolioID uuid.UUID,
	txType vo.TransactionType,
	amount, fees decimal.Decimal,
	currency string,
	valueDate time.Time,
	memo string,
) string {
	canonical := strings.Join([]string{
		submitterID.String(),
		portfolioID.String(),
		string(txType),
		amount.StringFixed(8),
		fees.StringFixed(8),
		strings.ToUpper(strings.TrimSpace(currency)),
		valueDate.UTC().Format("2006-01-02"),
		strings.TrimSpace(memo),
	}, fingerprintFieldSep)
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

// submitAndLinkCashApproval submits the staged cash request to the approval
// engine (or, when reconcile is set, re-links to an approval a prior attempt
// already submitted) and records the approval request id on the row.
func (h *PostTransactionHandler) submitAndLinkCashApproval(
	ctx context.Context,
	req PostTransactionRequest,
	prep *preparedTransaction,
	cr *entity.PortfolioCashRequest,
	reconcile bool,
) (*PostTransactionResult, error) {
	if reconcile {
		// Recover an approval a prior attempt already submitted for this subject
		// so a retry never creates a SECOND approval request. Fail closed when we
		// cannot query approval state — a blind re-submit would duplicate it.
		if h.approvalStatus == nil {
			return nil, &domain.ErrPostPreconditionFailed{
				Violation: string(policy.PostViolationApprovalUnavailable),
				Detail:    "cannot reconcile a pending cash request without the approval status provider",
			}
		}
		stage, err := h.approvalStatus.GetApprovalStage(ctx, cashApprovalSubjectType, cr.ID)
		if err != nil {
			return nil, fmt.Errorf("looking up existing approval for cash request %s: %w", cr.ID, err)
		}
		if stage != nil && stage.RequestID != uuid.Nil {
			return h.linkCashApproval(ctx, req, cr, stage.RequestID)
		}
		// No approval exists yet (the prior attempt failed at/before submit) —
		// fall through and submit a fresh one.
	}

	// contractType/contractID mirror how decisions choose scope: FUND when the
	// portfolio is fund-bound, else COMPANY (fund-optional) so the COMPANY-global
	// process config resolves.
	contractType := "COMPANY"
	var contractID *uuid.UUID
	if prep.portfolio.FundID != nil {
		contractType = "FUND"
		contractID = prep.portfolio.FundID
	}
	portfolioID := prep.portfolio.ID
	res, err := h.approval.SubmitForApproval(ctx, contract.ApprovalSubmission{
		ProcessType:      cashApprovalProcessType,
		SubjectType:      cashApprovalSubjectType,
		SubjectID:        cr.ID,
		SubjectTitle:     fmt.Sprintf("Cash %s %s %s", cr.TransactionType, cr.Amount.String(), cr.Currency),
		SubjectReference: prep.portfolio.Code,
		ContractType:     contractType,
		ContractID:       contractID,
		PortfolioID:      &portfolioID,
		SubmitterID:      req.ActorID,
	})
	if err != nil || res == nil {
		// AMBIGUOUS OUTCOME. A returned error or a nil result does NOT prove the
		// approval was not committed: the submit may have committed on the engine
		// side and then failed the post-commit read (timeout), or the engine's
		// duplicate-active guard fired because a prior attempt already committed,
		// or the failure is transient. Never blindly compensate — that would
		// strand a committed approval. Verify first, then re-link or (only when
		// verified absent) cancel.
		return h.resolveAmbiguousSubmit(ctx, req, cr, err)
	}
	return h.linkCashApproval(ctx, req, cr, res.RequestID)
}

// resolveAmbiguousSubmit handles a SubmitForApproval that returned an error or a
// nil result. It queries the approval engine for an approval actually committed
// against this subject and:
//   - re-links to it when one exists (no duplicate, no orphan, no lost request);
//   - compensates by cancelling the staged row ONLY when verified that no
//     approval exists (the submit truly failed);
//   - otherwise leaves the row PENDING (never cancels state it cannot reason
//     about) and returns a retryable error.
func (h *PostTransactionHandler) resolveAmbiguousSubmit(
	ctx context.Context,
	req PostTransactionRequest,
	cr *entity.PortfolioCashRequest,
	submitErr error,
) (*PostTransactionResult, error) {
	baseErr := submitErr
	if baseErr == nil {
		baseErr = fmt.Errorf("approval engine returned no result for cash request %s", cr.ID)
	}

	// Without the status provider we cannot verify whether an approval committed.
	// Fail closed WITHOUT compensating: a lingering PENDING row is safe (the
	// approval callback is terminal/idempotent-guarded and the effective key lets
	// a retry reconcile), whereas cancelling could strand a committed approval.
	if h.approvalStatus == nil {
		return nil, fmt.Errorf(
			"submitting cash movement for approval (left PENDING; cannot verify approval state without the status provider, retry to reconcile): %w",
			baseErr)
	}

	stage, sErr := h.approvalStatus.GetApprovalStage(ctx, cashApprovalSubjectType, cr.ID)
	if sErr != nil {
		// Cannot determine whether an approval exists → do not compensate; leave
		// the row PENDING for a reconciling retry and surface both errors.
		return nil, errors.Join(
			fmt.Errorf("submitting cash movement for approval: %w", baseErr),
			fmt.Errorf("verifying approval state for cash request %s failed (left PENDING for retry): %w", cr.ID, sErr),
		)
	}
	if stage != nil && stage.RequestID != uuid.Nil {
		// The approval actually committed — re-link instead of compensating.
		return h.linkCashApproval(ctx, req, cr, stage.RequestID)
	}

	// Verified: no approval exists for this subject. The submit truly failed, so
	// it is safe to compensate by cancelling the staged row. Surface a
	// compensation failure rather than swallowing it.
	if cErr := h.cancelStagedCashRequest(ctx, req.ActorID, cr, "approval submission failed; auto-cancelled."); cErr != nil {
		return nil, errors.Join(
			fmt.Errorf("submitting cash movement for approval: %w", baseErr),
			fmt.Errorf("compensating cancel of cash request %s failed (left PENDING): %w", cr.ID, cErr),
		)
	}
	return nil, fmt.Errorf("submitting cash movement for approval: %w", baseErr)
}

// linkCashApproval records the approval request id on the staged row and audits
// the submission. On a link (step-3) failure the approval already EXISTS but its
// id was not recorded; because every request now carries an effective
// idempotency key (client-supplied or server-derived), the row is always
// retry-reconcilable, so we leave it PENDING+NULL and return a retryable error.
// We deliberately never compensate here — cancelling would strand the committed
// approval; a retry re-links via GetApprovalStage and never creates a second.
func (h *PostTransactionHandler) linkCashApproval(
	ctx context.Context,
	req PostTransactionRequest,
	cr *entity.PortfolioCashRequest,
	approvalRequestID uuid.UUID,
) (*PostTransactionResult, error) {
	rid := approvalRequestID
	cr.ApprovalRequestID = &rid
	cr.UpdatedBy = req.ActorID
	cr.UpdatedAt = h.now()
	if err := h.runTransaction(ctx, func(dbtx pgx.Tx) error {
		return h.cashRequests.Update(ctx, dbtx, cr)
	}); err != nil {
		// Leave the row PENDING with a NULL approval_request_id so a retry
		// reconciles by re-linking to the already-submitted approval.
		return nil, fmt.Errorf(
			"recording approval request id for cash request %s (retry to reconcile): %w",
			cr.ID, err)
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_CASH_REQUEST_SUBMITTED",
		Module:       "investment",
		ResourceType: "INVESTMENT_CASH_REQUEST",
		ResourceID:   cr.ID.String(),
		Details: map[string]any{
			"portfolio_id":        cr.PortfolioID,
			"fund_id":             cr.FundID,
			"transaction_type":    string(cr.TransactionType),
			"amount":              cr.Amount.String(),
			"currency":            cr.Currency,
			"approval_request_id": cr.ApprovalRequestID,
		},
		BusinessDate: cr.ValueDate,
	})

	return &PostTransactionResult{CashRequest: cr, Pending: true}, nil
}

// cancelStagedCashRequest rolls a staged PENDING row to CANCELLED as
// compensation for a failed submit/link. It returns its own error so callers
// can surface a compensation failure instead of swallowing it.
func (h *PostTransactionHandler) cancelStagedCashRequest(
	ctx context.Context,
	actorID uuid.UUID,
	cr *entity.PortfolioCashRequest,
	note string,
) error {
	cancelNow := h.now()
	cr.Status = vo.CashRequestStatusCancelled
	cr.DecidedBy = &actorID
	cr.DecidedAt = &cancelNow
	cr.Memo = strings.TrimSpace(note + " " + cr.Memo)
	cr.UpdatedBy = actorID
	cr.UpdatedAt = cancelNow
	return h.runTransaction(ctx, func(dbtx pgx.Tx) error {
		return h.cashRequests.Update(ctx, dbtx, cr)
	})
}

// strPtr returns a pointer to a non-empty string, or nil for the empty string.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	v := s
	return &v
}

// ApplyCashRequestApproval is invoked by the approval subject callback when a
// CASH_TRANSACTION approval reaches a final outcome. It is idempotent (safe to
// call more than once for the same request) and all-or-nothing (the ledger
// insert and the request status flip commit in a single DB transaction).
//
// Note: the approval engine calls this AFTER the approval itself has committed,
// and CANNOT roll the approval back on error (see runPostAction). A failure of
// the authoritative re-check therefore leaves ZERO ledger rows AND leaves the
// request PENDING (not APPROVED) — the approval engine records the failure as a
// retryable APPROVAL_SYNC_FAILED audit event for operator replay.
func (h *PostTransactionHandler) ApplyCashRequestApproval(
	ctx context.Context,
	requestID uuid.UUID,
	approved bool,
	reason string,
) error {
	if h == nil || h.cashRequests == nil {
		return fmt.Errorf("cash request approval handler not initialised")
	}

	var (
		materialized bool
		rejected     bool
		resultTxnID  uuid.UUID
	)

	err := h.runTransaction(ctx, func(dbtx pgx.Tx) error {
		// Lock the request row first — this serializes duplicate approval
		// callbacks so two concurrent invocations cannot both materialize.
		cr, err := h.cashRequests.GetForUpdate(ctx, dbtx, requestID)
		if err != nil {
			return fmt.Errorf("locking cash request: %w", err)
		}
		if cr == nil {
			return &domain.ErrCashRequestNotFound{RequestID: requestID.String()}
		}
		// Idempotency guard: a terminal or already-materialized request is a
		// no-op (duplicate callback safe).
		if cr.Status.IsTerminal() || cr.IsMaterialized() {
			return nil
		}

		now := h.now()

		if !approved {
			cr.Status = vo.CashRequestStatusRejected
			cr.DecidedBy = &cr.SubmittedBy // system decision; preserve submitter attribution
			cr.DecidedAt = &now
			cr.Memo = mergeReason(cr.Memo, reason)
			cr.UpdatedBy = cr.SubmittedBy
			cr.UpdatedAt = now
			if err := h.cashRequests.Update(ctx, dbtx, cr); err != nil {
				return err
			}
			rejected = true
			return nil
		}

		// Approved — re-run the authoritative preconditions (portfolio still
		// active? workflow day still open/unlocked?) via the shared prepare +
		// execute path, then materialize the real ledger row and flip the
		// request status in this same transaction.
		postReq := PostTransactionRequest{
			PortfolioID:     cr.PortfolioID,
			TransactionType: cr.TransactionType,
			Currency:        cr.Currency,
			Fees:            cr.Fees,
			GrossAmount:     &cr.Amount,
			BusinessDate:    cr.ValueDate,
			Reason:          cr.Memo,
			ActorID:         cr.SubmittedBy, // attribute the ledger post to the submitter
		}
		txID := h.nextID()
		prep, err := h.prepareTransaction(ctx, postReq, false, false, false, txID)
		if err != nil {
			return err
		}
		ledger := h.buildLedgerEntity(postReq, prep, txID, now)
		if err := h.executePostTx(ctx, dbtx, postReq, prep, ledger); err != nil {
			return err
		}

		cr.Status = vo.CashRequestStatusApproved
		cr.ResultingTxnID = &txID
		cr.DecidedBy = &cr.SubmittedBy
		cr.DecidedAt = &now
		cr.UpdatedBy = cr.SubmittedBy
		cr.UpdatedAt = now
		if err := h.cashRequests.Update(ctx, dbtx, cr); err != nil {
			return err
		}
		materialized = true
		resultTxnID = txID
		return nil
	})
	if err != nil {
		return err
	}

	switch {
	case materialized:
		_ = h.audit.LogAction(contract.AuditEntry{
			Action:       "INVESTMENT_CASH_REQUEST_APPROVED",
			Module:       "investment",
			ResourceType: "INVESTMENT_CASH_REQUEST",
			ResourceID:   requestID.String(),
			Details: map[string]any{
				"approved":         true,
				"resulting_txn_id": resultTxnID.String(),
			},
		})
	case rejected:
		_ = h.audit.LogAction(contract.AuditEntry{
			Action:       "INVESTMENT_CASH_REQUEST_REJECTED",
			Module:       "investment",
			ResourceType: "INVESTMENT_CASH_REQUEST",
			ResourceID:   requestID.String(),
			Details:      map[string]any{"approved": false, "reason": reason},
		})
	}
	return nil
}

// CancelCashRequest cancels a PENDING cash request on behalf of its submitter.
// After cancel, approvers can no longer act on it (the approval request is
// cancelled and the ApprovalSubjectValidator refuses further action).
func (h *PostTransactionHandler) CancelCashRequest(
	ctx context.Context,
	requestID, actorID uuid.UUID,
	reason string,
) (*entity.PortfolioCashRequest, error) {
	if h == nil || h.cashRequests == nil {
		return nil, fmt.Errorf("cash request handler not initialised")
	}
	if actorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}

	cr, err := h.cashRequests.GetByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("loading cash request: %w", err)
	}
	if cr == nil {
		return nil, &domain.ErrCashRequestNotFound{RequestID: requestID.String()}
	}
	// v1: only the submitter may cancel.
	if cr.SubmittedBy != actorID {
		return nil, &domain.ErrCashRequestForbidden{
			RequestID: requestID.String(),
			Reason:    "only the submitter may cancel this request",
		}
	}
	if !cr.IsPending() {
		return nil, &domain.ErrCashRequestNotCancellable{
			RequestID:     requestID.String(),
			CurrentStatus: string(cr.Status),
		}
	}

	// Cancel the in-flight approval request first so approvers can no longer act.
	if h.approvalCanceller != nil {
		if err := h.approvalCanceller.CancelApprovalBySubject(ctx, "CASH_TRANSACTION", cr.ID, actorID); err != nil {
			return nil, fmt.Errorf("cancelling approval request: %w", err)
		}
	}

	now := h.now()
	if err := h.runTransaction(ctx, func(dbtx pgx.Tx) error {
		// Re-lock + re-check under the row lock to guard against an approve
		// racing the cancel.
		locked, err := h.cashRequests.GetForUpdate(ctx, dbtx, requestID)
		if err != nil {
			return err
		}
		if locked == nil {
			return &domain.ErrCashRequestNotFound{RequestID: requestID.String()}
		}
		if !locked.IsPending() {
			return &domain.ErrCashRequestNotCancellable{
				RequestID:     requestID.String(),
				CurrentStatus: string(locked.Status),
			}
		}
		locked.Status = vo.CashRequestStatusCancelled
		locked.DecidedBy = &actorID
		locked.DecidedAt = &now
		locked.Memo = mergeReason(locked.Memo, reason)
		locked.UpdatedBy = actorID
		locked.UpdatedAt = now
		if err := h.cashRequests.Update(ctx, dbtx, locked); err != nil {
			return err
		}
		cr = locked
		return nil
	}); err != nil {
		return nil, err
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      actorID.String(),
		Action:       "INVESTMENT_CASH_REQUEST_CANCELLED",
		Module:       "investment",
		ResourceType: "INVESTMENT_CASH_REQUEST",
		ResourceID:   cr.ID.String(),
		Details:      map[string]any{"reason": reason},
		BusinessDate: now,
	})
	return cr, nil
}

// mergeReason appends a decision/cancellation reason to an existing memo without
// losing the original submitter note.
func mergeReason(memo, reason string) string {
	reason = strings.TrimSpace(reason)
	memo = strings.TrimSpace(memo)
	switch {
	case reason == "":
		return memo
	case memo == "":
		return reason
	default:
		return memo + " | " + reason
	}
}
