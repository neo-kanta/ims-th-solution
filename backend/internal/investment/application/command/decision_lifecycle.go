package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// CreateDecisionRequest is the input for creating a brand-new DRAFT decision.
type CreateDecisionRequest struct {
	// FundID is nil for a decision on a fund-less portfolio.
	FundID           *uuid.UUID
	PortfolioID      uuid.UUID
	InstrumentID     *uuid.UUID
	InstrumentCode   string
	BusinessDate     time.Time
	Side             vo.OrderSide
	Quantity         *decimal.Decimal
	Amount           *decimal.Decimal
	LimitPrice       *decimal.Decimal
	Currency         string
	Exchange         string
	ResearchReportID *uuid.UUID
	Rationale        string
	ActorID          uuid.UUID
}

// UpdateDecisionRequest is the input for editing a DRAFT decision.
type UpdateDecisionRequest struct {
	DecisionID       uuid.UUID
	InstrumentID     *uuid.UUID
	InstrumentCode   *string
	BusinessDate     *time.Time
	Side             *vo.OrderSide
	Quantity         *decimal.Decimal
	Amount           *decimal.Decimal
	LimitPrice       *decimal.Decimal
	Currency         *string
	Exchange         *string
	ResearchReportID *uuid.UUID
	Rationale        *string
	ActorID          uuid.UUID
}

// CancelDecisionRequest is the input for cancelling a DRAFT or PENDING decision.
type CancelDecisionRequest struct {
	DecisionID uuid.UUID
	Reason     string
	ActorID    uuid.UUID
}

// DecisionCommandHandler owns the lifecycle commands for investment decisions:
// create (DRAFT), update (DRAFT-only), submit (DRAFT → PENDING_APPROVAL via
// approval engine), cancel (DRAFT/PENDING → CANCELLED), and the
// final-decision callback bridged from the approval engine
// (ApplyApprovalDecision).
type DecisionCommandHandler struct {
	pool              *pgxpool.Pool
	decisions         domain.DecisionRepository
	reports           domain.ResearchReportRepository
	workflow          contract.WorkflowStateProvider
	audit             contract.AuditLogger
	now               func() time.Time
	runTx             func(ctx context.Context, fn func(pgx.Tx) error) error
	approval          contract.ApprovalSubmitter
	approvalCanceller contract.ApprovalCanceller
	compliance        contract.ComplianceChecker
	funds             domain.FundRepository
	portfolios        domain.PortfolioRepository
}

// NewDecisionCommandHandler wires the handler.
func NewDecisionCommandHandler(
	pool *pgxpool.Pool,
	decisions domain.DecisionRepository,
	reports domain.ResearchReportRepository,
	workflow contract.WorkflowStateProvider,
	audit contract.AuditLogger,
	now func() time.Time,
) *DecisionCommandHandler {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	h := &DecisionCommandHandler{
		pool:      pool,
		decisions: decisions,
		reports:   reports,
		workflow:  workflow,
		audit:     audit,
		now:       now,
	}
	h.runTx = func(ctx context.Context, fn func(pgx.Tx) error) error {
		return withTransaction(ctx, pool, fn)
	}
	return h
}

// SetApprovalSubmitter injects the approval submitter post-construction to
// avoid a circular construction dependency between investment and approval.
func (h *DecisionCommandHandler) SetApprovalSubmitter(s contract.ApprovalSubmitter) {
	if h != nil {
		h.approval = s
	}
}

// SetApprovalCanceller injects the approval canceller post-construction.
func (h *DecisionCommandHandler) SetApprovalCanceller(c contract.ApprovalCanceller) {
	if h != nil {
		h.approvalCanceller = c
	}
}

// SetComplianceChecker injects the pre-trade IRG compliance gate post-construction.
func (h *DecisionCommandHandler) SetComplianceChecker(c contract.ComplianceChecker) {
	if h != nil {
		h.compliance = c
	}
}

// SetFundRepository injects the fund repository post-construction (used for the
// config-driven report-required gate in Submit).
func (h *DecisionCommandHandler) SetFundRepository(f domain.FundRepository) {
	if h != nil {
		h.funds = f
	}
}

// SetPortfolioRepository injects the authoritative portfolio repository used
// by Submit to determine LIVE/SIMULATION/MODEL policy. Portfolio type is never
// accepted from a caller or transport request.
func (h *DecisionCommandHandler) SetPortfolioRepository(p domain.PortfolioRepository) {
	if h != nil {
		h.portfolios = p
	}
}

// Create persists a brand-new DRAFT decision and emits an audit record.
// Reference rules on the research report (if any) are enforced at submission
// time, not creation — drafts may reference a report whose review has not yet
// completed.
func (h *DecisionCommandHandler) Create(ctx context.Context, req CreateDecisionRequest) (*entity.Decision, error) {
	if err := validateCreateDecision(req); err != nil {
		return nil, err
	}

	// Light validation of the optional research report reference. We accept
	// references to non-final reports while the decision is still DRAFT — the
	// hard policy check runs at Submit time so drafts can be assembled
	// concurrently with the report's approval cycle.
	if req.ResearchReportID != nil {
		rep, err := h.reports.GetByID(ctx, *req.ResearchReportID)
		if err != nil {
			return nil, fmt.Errorf("loading research report: %w", err)
		}
		if rep == nil || rep.IsDeleted() {
			return nil, &domain.ErrDecisionReferenceInvalid{
				ReportID: req.ResearchReportID.String(),
				Reason:   string(policy.ViolationReportMissing),
			}
		}
	}

	now := h.now()
	d := &entity.Decision{
		ID:               uuid.New(),
		FundID:           req.FundID,
		PortfolioID:      req.PortfolioID,
		InstrumentID:     req.InstrumentID,
		InstrumentCode:   strings.ToUpper(strings.TrimSpace(req.InstrumentCode)),
		BusinessDate:     req.BusinessDate.UTC(),
		Side:             req.Side,
		Quantity:         req.Quantity,
		Amount:           req.Amount,
		LimitPrice:       req.LimitPrice,
		Currency:         strings.ToUpper(strings.TrimSpace(req.Currency)),
		Exchange:         strings.TrimSpace(req.Exchange),
		ResearchReportID: req.ResearchReportID,
		Rationale:        req.Rationale,
		Status:           vo.DecisionLifecycleDraft,
		SubmitterUserID:  req.ActorID,
		// This request shape only ever builds a single-order decision header
		// (basket/rebalance/switch decisions are out of scope for Create).
		// These three columns are NOT NULL with a DB-side DEFAULT
		// (20260615000004_investment__add_decision_basket_fields.up.sql), but
		// the DEFAULT only applies when a column is omitted from the INSERT —
		// the persistence layer's INSERT always lists them explicitly, so an
		// unset Go zero value inserts '' and trips
		// chk_inv_decision_decision_type/process_type/product_type. Set the
		// same values the migration documents as the defaults.
		DecisionType: vo.DecisionTypeSingleOrder,
		ProcessType:  vo.DecisionProcessInvestment,
		ProductType:  vo.DecisionProductMutualFund,
		CreatedAt:    now,
		CreatedBy:    req.ActorID,
		UpdatedAt:    now,
		UpdatedBy:    req.ActorID,
	}

	if req.ResearchReportID != nil {
		// Capture the human-friendly number for read paths; load again only
		// because the earlier load was a presence check.
		rep, _ := h.reports.GetByID(ctx, *req.ResearchReportID)
		if rep != nil {
			d.ResearchReportNo = rep.ReportNo
		}
	}

	err := h.runTx(ctx, func(tx pgx.Tx) error {
		num, err := h.decisions.NextDecisionNumber(ctx, tx, d.BusinessDate)
		if err != nil {
			return err
		}
		d.DecisionNumber = num
		return h.decisions.Create(ctx, tx, d)
	})
	if err != nil {
		return nil, err
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_DECISION_CREATED",
		Module:       "investment",
		ResourceType: "INVESTMENT_DECISION",
		ResourceID:   d.ID.String(),
		Details: map[string]any{
			"decision_number": d.DecisionNumber,
			"fund_id":         d.FundID,
			"side":            string(d.Side),
			"instrument_code": d.InstrumentCode,
			"business_date":   d.BusinessDate.Format("2006-01-02"),
		},
		BusinessDate: now,
	})
	return d, nil
}

// Update applies a partial edit to a DRAFT decision. Refused once the
// decision is past DRAFT — use Cancel + re-create instead.
func (h *DecisionCommandHandler) Update(ctx context.Context, req UpdateDecisionRequest) (*entity.Decision, error) {
	if req.DecisionID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "decision_id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}

	d, err := h.decisions.GetByID(ctx, req.DecisionID)
	if err != nil {
		return nil, fmt.Errorf("loading decision: %w", err)
	}
	if d == nil {
		return nil, &domain.ErrDecisionNotFound{DecisionID: req.DecisionID.String()}
	}
	if !d.CanEdit() {
		return nil, &domain.ErrDecisionLifecycle{
			DecisionID:    d.ID.String(),
			CurrentStatus: string(d.Status),
			Detail:        "only DRAFT decisions can be edited",
		}
	}

	if req.InstrumentID != nil {
		d.InstrumentID = req.InstrumentID
	}
	if req.InstrumentCode != nil {
		d.InstrumentCode = strings.ToUpper(strings.TrimSpace(*req.InstrumentCode))
	}
	if req.BusinessDate != nil {
		d.BusinessDate = req.BusinessDate.UTC()
	}
	if req.Side != nil {
		d.Side = *req.Side
	}
	if req.Quantity != nil {
		d.Quantity = req.Quantity
	}
	if req.Amount != nil {
		d.Amount = req.Amount
	}
	if req.LimitPrice != nil {
		d.LimitPrice = req.LimitPrice
	}
	if req.Currency != nil {
		d.Currency = strings.ToUpper(strings.TrimSpace(*req.Currency))
	}
	if req.Exchange != nil {
		d.Exchange = strings.TrimSpace(*req.Exchange)
	}
	if req.ResearchReportID != nil {
		// Allow caller to reset by sending uuid.Nil through pointer? In
		// pointer-based partial updates we treat any non-nil pointer as a
		// set. Verify the report still exists.
		rep, err := h.reports.GetByID(ctx, *req.ResearchReportID)
		if err != nil {
			return nil, fmt.Errorf("loading research report: %w", err)
		}
		if rep == nil || rep.IsDeleted() {
			return nil, &domain.ErrDecisionReferenceInvalid{
				DecisionID: d.ID.String(),
				ReportID:   req.ResearchReportID.String(),
				Reason:     string(policy.ViolationReportMissing),
			}
		}
		d.ResearchReportID = req.ResearchReportID
		d.ResearchReportNo = rep.ReportNo
	}
	if req.Rationale != nil {
		d.Rationale = *req.Rationale
	}

	now := h.now()
	d.UpdatedAt = now
	d.UpdatedBy = req.ActorID

	if err := h.runTx(ctx, func(tx pgx.Tx) error {
		return h.decisions.Update(ctx, tx, d)
	}); err != nil {
		return nil, err
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_DECISION_UPDATED",
		Module:       "investment",
		ResourceType: "INVESTMENT_DECISION",
		ResourceID:   d.ID.String(),
		BusinessDate: now,
	})
	return d, nil
}

// Submit moves the decision from DRAFT to PENDING_APPROVAL after enforcing:
//   - workflow day is open / not locked,
//   - report reference policy passes when a report is linked.
//
// When the approval engine is wired, a real approval request is created and
// the decision is transitioned only on success.
func (h *DecisionCommandHandler) Submit(ctx context.Context, decisionID, actorID uuid.UUID) (*entity.Decision, error) {
	if decisionID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "decision_id", Detail: "is required"}
	}
	if actorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	d, err := h.decisions.GetByID(ctx, decisionID)
	if err != nil {
		return nil, fmt.Errorf("loading decision: %w", err)
	}
	if d == nil {
		return nil, &domain.ErrDecisionNotFound{DecisionID: decisionID.String()}
	}
	if !d.CanSubmit() {
		return nil, &domain.ErrDecisionLifecycle{
			DecisionID:    d.ID.String(),
			CurrentStatus: string(d.Status),
			Detail:        "only DRAFT decisions can be submitted",
		}
	}

	// Workflow gate — refuse to submit when the day is closed or locked. A
	// fund-less decision has no fund-scoped business day to check.
	if h.workflow != nil && d.FundID != nil {
		allowed, err := h.workflow.IsTradeAllowed(ctx, *d.FundID, d.BusinessDate)
		if err != nil {
			return nil, fmt.Errorf("checking workflow trade gate: %w", err)
		}
		if !allowed {
			return nil, &domain.ErrDecisionLifecycle{
				DecisionID:    d.ID.String(),
				CurrentStatus: string(d.Status),
				Detail:        "workflow day not open for decision submission",
			}
		}
	}

	// Config-driven report gate: if the fund requires a research report for
	// decision submission, a report must be linked before proceeding.
	if h.funds != nil && d.FundID != nil {
		fund, err := h.funds.GetByID(ctx, *d.FundID)
		if err != nil {
			return nil, fmt.Errorf("loading fund for report gate: %w", err)
		}
		if fund != nil && fund.RequireResearchReportForDecision && d.ResearchReportID == nil {
			return nil, &domain.ErrDecisionLifecycle{
				DecisionID:    d.ID.String(),
				CurrentStatus: string(d.Status),
				Detail:        "fund requires a research report before decision can be submitted",
			}
		}
	}

	// contractID stays uuid.Nil for a fund-less decision — every downstream
	// contract-scoped check here (report reference, compliance) already
	// treats uuid.Nil as "no contract scope applies" rather than an error.
	contractID := uuid.Nil
	if d.FundID != nil {
		contractID = *d.FundID
	}

	// Reference rule enforcement runs before the compliance check so that a
	// releasable-BLOCK path also validates the linked report. The compliance
	// early-return at line ~476 would otherwise skip this check.
	if d.ResearchReportID != nil {
		rep, err := h.reports.GetByID(ctx, *d.ResearchReportID)
		if err != nil {
			return nil, fmt.Errorf("loading research report: %w", err)
		}
		violation := policy.CanReferenceResearchReport(policy.ReportReferenceInput{
			Report:       rep,
			ContractID:   contractID,
			Side:         d.Side,
			BusinessDate: d.BusinessDate,
		})
		if violation != "" {
			return nil, &domain.ErrDecisionReferenceInvalid{
				DecisionID: d.ID.String(),
				ReportID:   d.ResearchReportID.String(),
				Reason:     string(violation),
			}
		}
	}

	// Pre-trade IRG compliance check. Runs after the workflow gate and before
	// the approval engine so a BLOCK verdict never reaches the approval queue.
	// For BASKET_ORDER / REBALANCE / SWITCH the Ticker field is empty — the
	// compliance engine evaluates header-level rules only (per-line checks are Phase 2).
	if h.compliance != nil {
		portfolioType, err := resolveCompliancePortfolioType(ctx, h.portfolios, d.PortfolioID)
		if err != nil {
			return nil, err
		}
		auditGap := func(gateErr error) error {
			return auditComplianceControlGap(ctx, h.audit, complianceControlGapAuditInput{
				Phase:          "DECISION_SUBMISSION",
				DecisionID:     d.ID,
				PortfolioID:    d.PortfolioID,
				PortfolioType:  portfolioType,
				ActorID:        actorID,
				BusinessDate:   d.BusinessDate,
				InstrumentCode: d.InstrumentCode,
			}, gateErr)
		}
		qty, price, err := decisionComplianceOrderValues(d)
		if err != nil {
			return nil, err
		}
		checkGroupID := uuid.New()
		result, err := h.compliance.CheckProposedOrder(ctx, contract.ProposedOrderCheck{
			CheckGroupID: checkGroupID,
			PortfolioID:  d.PortfolioID,
			ContractID:   contractID,
			BusinessDate: d.BusinessDate,
			Actor:        actorID.String(),
			OrderID:      d.ID,
			Ticker:       d.InstrumentCode,
			Side:         mapOrderSideToContract(d.Side),
			Quantity:     qty,
			Price:        price,
			Currency:     d.Currency,
			Exchange:     d.Exchange,
		})
		if err != nil {
			return nil, fmt.Errorf("pre-trade compliance check: %w", err)
		}
		if result == nil {
			gateErr := &ErrComplianceUnavailable{
				CheckGroupID: checkGroupID.String(),
				Reason:       "pre-trade compliance check returned nil result",
			}
			if err := auditGap(gateErr); err != nil {
				return nil, err
			}
			return nil, gateErr
		}
		if gateErr := liveComplianceStatusError(portfolioType, result); gateErr != nil {
			if err := auditGap(gateErr); err != nil {
				return nil, err
			}
			return nil, gateErr
		}
		switch result.Verdict {
		case contract.ComplianceVerdictPass, contract.ComplianceVerdictWarn, contract.ComplianceVerdictBlock:
			// Supported business verdicts continue through the existing policy.
		default:
			gateErr := &ErrComplianceUnavailable{
				CheckGroupID: result.CheckGroupID.String(),
				Reason:       fmt.Sprintf("unsupported compliance verdict %q", result.Verdict),
			}
			if err := auditGap(gateErr); err != nil {
				return nil, err
			}
			return nil, gateErr
		}
		cgid := result.CheckGroupID
		d.ComplianceCheckGroupID = &cgid

		if result.Verdict == contract.ComplianceVerdictBlock {
			allReleasable := true
			for _, b := range result.Breaches {
				if !b.Overridable {
					allReleasable = false
					break
				}
			}
			if !allReleasable || h.approval == nil {
				// Persist the check group ID for audit trail before rejecting.
				nowBlk := h.now()
				d.UpdatedAt = nowBlk
				d.UpdatedBy = actorID
				_ = h.runTx(ctx, func(tx pgx.Tx) error {
					return h.decisions.Update(ctx, tx, d)
				})
				return nil, &domain.ErrComplianceRejected{
					DecisionID:   d.ID.String(),
					CheckGroupID: result.CheckGroupID.String(),
					Message:      summarizeBreaches(result.Breaches),
				}
			}
			// All breaches are overridable — submit COMPLIANCE_RELEASE approval
			// first; only persist the status change after the submission succeeds.
			// If submission fails the decision stays in DRAFT and can be retried.
			nowCR := h.now()
			res, err := h.approval.SubmitForApproval(ctx, contract.ApprovalSubmission{
				ProcessType:      "COMPLIANCE_RELEASE",
				SubjectType:      "COMPLIANCE_RELEASE",
				SubjectID:        d.ID,
				SubjectTitle:     fmt.Sprintf("Compliance Release — %s %s %s", d.Side, d.InstrumentCode, d.DecisionNumber),
				SubjectReference: d.DecisionNumber,
				ContractType:     "FUND",
				ContractID:       d.FundID,
				PortfolioID:      &d.PortfolioID,
				SubmitterID:      actorID,
			})
			if err != nil {
				return nil, fmt.Errorf("submitting compliance release approval: %w", err)
			}
			d.Status = vo.DecisionLifecyclePendingComplianceRelease
			d.UpdatedAt = nowCR
			d.UpdatedBy = actorID
			if res != nil {
				rid := res.RequestID
				// Store the compliance-release request ID in its own field so
				// the investment-decision approval ID is not overwritten later.
				d.ComplianceReleaseApprovalRequestID = &rid
				d.ApprovalStatus = res.Status
			}
			if err := h.runTx(ctx, func(tx pgx.Tx) error {
				return h.decisions.Update(ctx, tx, d)
			}); err != nil {
				return nil, err
			}
			h.audit.LogAction(contract.AuditEntry{
				ActorID:      actorID.String(),
				Action:       "INVESTMENT_DECISION_COMPLIANCE_RELEASE_SUBMITTED",
				Module:       "investment",
				ResourceType: "INVESTMENT_DECISION",
				ResourceID:   d.ID.String(),
				Details:      map[string]any{"check_group_id": result.CheckGroupID.String()},
				BusinessDate: nowCR,
			})
			return d, nil // decision is pending compliance release; skip investment approval
		}
		// PASS and WARN both continue to the investment decision approval engine.
	}

	// Submit through approval engine when wired. ContractID is nil for a
	// fund-less decision — the approval engine's contract-scope check
	// already tolerates a nil ContractID (see other ContractType usages).
	if h.approval != nil {
		res, err := h.approval.SubmitForApproval(ctx, contract.ApprovalSubmission{
			ProcessType:      "INVESTMENT_DECISION",
			SubjectType:      "INVESTMENT_DECISION",
			SubjectID:        d.ID,
			SubjectTitle:     fmt.Sprintf("%s %s %s", d.Side, d.InstrumentCode, d.DecisionNumber),
			SubjectReference: d.DecisionNumber,
			ContractType:     "FUND",
			ContractID:       d.FundID,
			PortfolioID:      &d.PortfolioID,
			SubmitterID:      actorID,
		})
		if err != nil {
			return nil, err
		}
		if res != nil {
			id := res.RequestID
			d.ApprovalRequestID = &id
			d.ApprovalStatus = res.Status
		}
	}

	now := h.now()
	d.Status = vo.DecisionLifecyclePendingApproval
	d.SubmittedAt = &now
	d.UpdatedAt = now
	d.UpdatedBy = actorID
	if err := h.runTx(ctx, func(tx pgx.Tx) error {
		return h.decisions.Update(ctx, tx, d)
	}); err != nil {
		return nil, err
	}
	if err := h.audit.LogActionStrict(ctx, contract.AuditEntry{
		ActorID:      actorID.String(),
		Action:       "INVESTMENT_DECISION_SUBMITTED",
		Module:       "investment",
		ResourceType: "INVESTMENT_DECISION",
		ResourceID:   d.ID.String(),
		Details: map[string]any{
			"decision_number":    d.DecisionNumber,
			"approval_request":   d.ApprovalRequestID,
			"research_report_id": d.ResearchReportID,
		},
		BusinessDate: now,
	}); err != nil {
		return nil, fmt.Errorf("auditing decision submit: %w", err)
	}
	return d, nil
}

// Cancel transitions a DRAFT or PENDING_APPROVAL decision to CANCELLED.
// A reason is required and is recorded on the row + audit log.
func (h *DecisionCommandHandler) Cancel(ctx context.Context, req CancelDecisionRequest) (*entity.Decision, error) {
	if req.DecisionID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "decision_id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "reason", Detail: "is required"}
	}
	d, err := h.decisions.GetByID(ctx, req.DecisionID)
	if err != nil {
		return nil, fmt.Errorf("loading decision: %w", err)
	}
	if d == nil {
		return nil, &domain.ErrDecisionNotFound{DecisionID: req.DecisionID.String()}
	}
	if !d.CanCancel() {
		return nil, &domain.ErrDecisionLifecycle{
			DecisionID:    d.ID.String(),
			CurrentStatus: string(d.Status),
			Detail:        "cannot cancel from this status",
		}
	}
	if h.approvalCanceller != nil && d.Status == vo.DecisionLifecyclePendingApproval {
		if err := h.approvalCanceller.CancelApprovalBySubject(ctx, "INVESTMENT_DECISION", d.ID, req.ActorID); err != nil {
			return nil, fmt.Errorf("cancelling approval request: %w", err)
		}
	}
	now := h.now()
	d.Status = vo.DecisionLifecycleCancelled
	d.CancelledAt = &now
	d.CancelledBy = &req.ActorID
	d.CancellationReason = reason
	d.UpdatedAt = now
	d.UpdatedBy = req.ActorID
	if err := h.runTx(ctx, func(tx pgx.Tx) error {
		return h.decisions.Update(ctx, tx, d)
	}); err != nil {
		return nil, err
	}
	if err := h.audit.LogActionStrict(ctx, contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_DECISION_CANCELLED",
		Module:       "investment",
		ResourceType: "INVESTMENT_DECISION",
		ResourceID:   d.ID.String(),
		Details:      map[string]any{"reason": reason},
		BusinessDate: now,
	}); err != nil {
		return nil, fmt.Errorf("auditing decision cancel: %w", err)
	}
	return d, nil
}

// ApplyApprovalDecision is invoked by the approval module via the registered
// subject callback when a decision's approval reaches a final outcome.
// Approved → APPROVED + ready_for_execution timestamp; rejected → REJECTED.
func (h *DecisionCommandHandler) ApplyApprovalDecision(ctx context.Context, decisionID uuid.UUID, approved bool, reason string) error {
	d, err := h.decisions.GetByID(ctx, decisionID)
	if err != nil {
		return fmt.Errorf("loading decision: %w", err)
	}
	if d == nil {
		return &domain.ErrDecisionNotFound{DecisionID: decisionID.String()}
	}
	now := h.now()
	if approved {
		d.Status = vo.DecisionLifecycleApproved
		d.ApprovalStatus = "APPROVED"
		d.ReadyForExecutionAt = &now
	} else {
		d.Status = vo.DecisionLifecycleRejected
		d.ApprovalStatus = "REJECTED"
		d.CancellationReason = strings.TrimSpace(reason)
	}
	d.UpdatedAt = now
	d.UpdatedBy = d.SubmitterUserID // system update; preserve submitter audit
	if err := h.runTx(ctx, func(tx pgx.Tx) error {
		return h.decisions.Update(ctx, tx, d)
	}); err != nil {
		return err
	}
	action := "INVESTMENT_DECISION_APPROVAL_APPROVED"
	if !approved {
		action = "INVESTMENT_DECISION_APPROVAL_REJECTED"
	}
	if err := h.audit.LogActionStrict(ctx, contract.AuditEntry{
		Action:       action,
		Module:       "investment",
		ResourceType: "INVESTMENT_DECISION",
		ResourceID:   d.ID.String(),
		Details:      map[string]any{"approved": approved, "reason": reason},
		BusinessDate: now,
	}); err != nil {
		return fmt.Errorf("auditing decision approval callback: %w", err)
	}
	return nil
}

// ApplyComplianceReleaseDecision is invoked by the approval module when a
// COMPLIANCE_RELEASE approval request reaches a final outcome.
//
// Approved → the compliance breach is considered released; the decision is
// automatically re-submitted to the standard INVESTMENT_DECISION approval engine
// so the portfolio manager can proceed.
//
// Rejected → the decision is transitioned to CANCELLED (it cannot be traded
// while a compliance block stands unresolved).
func (h *DecisionCommandHandler) ApplyComplianceReleaseDecision(ctx context.Context, decisionID uuid.UUID, approved bool, reason string) error {
	d, err := h.decisions.GetByID(ctx, decisionID)
	if err != nil {
		return fmt.Errorf("loading decision for compliance release: %w", err)
	}
	if d == nil {
		return &domain.ErrDecisionNotFound{DecisionID: decisionID.String()}
	}
	if d.Status != vo.DecisionLifecyclePendingComplianceRelease {
		return &domain.ErrDecisionLifecycle{
			DecisionID:    d.ID.String(),
			CurrentStatus: string(d.Status),
			Detail:        "decision is no longer in PENDING_COMPLIANCE_RELEASE state",
		}
	}
	now := h.now()
	if !approved {
		d.Status = vo.DecisionLifecycleCancelled
		d.CancellationReason = fmt.Sprintf("compliance release rejected: %s", strings.TrimSpace(reason))
		d.CancelledAt = &now
		d.CancelledBy = &d.SubmitterUserID
		d.UpdatedAt = now
		d.UpdatedBy = d.SubmitterUserID
		if err := h.runTx(ctx, func(tx pgx.Tx) error {
			return h.decisions.Update(ctx, tx, d)
		}); err != nil {
			return err
		}
		h.audit.LogAction(contract.AuditEntry{
			Action:       "INVESTMENT_DECISION_COMPLIANCE_RELEASE_REJECTED",
			Module:       "investment",
			ResourceType: "INVESTMENT_DECISION",
			ResourceID:   d.ID.String(),
			Details:      map[string]any{"reason": reason},
			BusinessDate: now,
		})
		return nil
	}
	// Compliance release approved — re-validate before re-submitting to the
	// investment decision approval engine. No fund-scoped business day to
	// check for a fund-less decision.
	if h.workflow != nil && d.FundID != nil {
		allowed, wfErr := h.workflow.IsTradeAllowed(ctx, *d.FundID, d.BusinessDate)
		if wfErr != nil {
			return fmt.Errorf("checking workflow trade gate: %w", wfErr)
		}
		if !allowed {
			return &domain.ErrDecisionLifecycle{
				DecisionID:    d.ID.String(),
				CurrentStatus: string(d.Status),
				Detail:        "workflow day no longer open; compliance release cannot advance decision",
			}
		}
	}
	if d.ResearchReportID != nil {
		rep, repErr := h.reports.GetByID(ctx, *d.ResearchReportID)
		if repErr != nil {
			return fmt.Errorf("loading research report during compliance release: %w", repErr)
		}
		reportContractID := uuid.Nil
		if d.FundID != nil {
			reportContractID = *d.FundID
		}
		violation := policy.CanReferenceResearchReport(policy.ReportReferenceInput{
			Report:       rep,
			ContractID:   reportContractID,
			Side:         d.Side,
			BusinessDate: d.BusinessDate,
		})
		if violation != "" {
			return &domain.ErrDecisionReferenceInvalid{
				DecisionID: d.ID.String(),
				ReportID:   d.ResearchReportID.String(),
				Reason:     string(violation),
			}
		}
	}
	if h.approval == nil {
		return fmt.Errorf("approval engine not wired; cannot continue decision %s after compliance release", decisionID)
	}
	res, err := h.approval.SubmitForApproval(ctx, contract.ApprovalSubmission{
		ProcessType:      "INVESTMENT_DECISION",
		SubjectType:      "INVESTMENT_DECISION",
		SubjectID:        d.ID,
		SubjectTitle:     fmt.Sprintf("%s %s %s", d.Side, d.InstrumentCode, d.DecisionNumber),
		SubjectReference: d.DecisionNumber,
		ContractType:     "FUND",
		ContractID:       d.FundID,
		PortfolioID:      &d.PortfolioID,
		SubmitterID:      d.SubmitterUserID,
	})
	if err != nil {
		return fmt.Errorf("re-submitting decision after compliance release: %w", err)
	}
	d.Status = vo.DecisionLifecyclePendingApproval
	if res != nil {
		rid := res.RequestID
		d.ApprovalRequestID = &rid
		d.ApprovalStatus = res.Status
	}
	d.UpdatedAt = now
	d.UpdatedBy = d.SubmitterUserID
	if err := h.runTx(ctx, func(tx pgx.Tx) error {
		return h.decisions.Update(ctx, tx, d)
	}); err != nil {
		return err
	}
	if err := h.audit.LogActionStrict(ctx, contract.AuditEntry{
		Action:       "INVESTMENT_DECISION_COMPLIANCE_RELEASED_SUBMITTED_FOR_APPROVAL",
		Module:       "investment",
		ResourceType: "INVESTMENT_DECISION",
		ResourceID:   d.ID.String(),
		Details:      map[string]any{"approval_request_id": d.ApprovalRequestID},
		BusinessDate: now,
	}); err != nil {
		return fmt.Errorf("auditing compliance release approval submission: %w", err)
	}
	return nil
}

func validateCreateDecision(req CreateDecisionRequest) error {
	if req.ActorID == uuid.Nil {
		return &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	if req.PortfolioID == uuid.Nil {
		return &domain.ErrInvalidDecisionRequest{Field: "portfolio_id", Detail: "is required"}
	}
	if req.BusinessDate.IsZero() {
		return &domain.ErrInvalidDecisionRequest{Field: "business_date", Detail: "is required"}
	}
	if !req.Side.IsValid() {
		return &domain.ErrInvalidDecisionRequest{Field: "side", Detail: "must be BUY or SELL"}
	}
	if strings.TrimSpace(req.InstrumentCode) == "" {
		return &domain.ErrInvalidDecisionRequest{Field: "instrument_code", Detail: "is required"}
	}
	if strings.TrimSpace(req.Currency) == "" {
		return &domain.ErrInvalidDecisionRequest{Field: "currency", Detail: "is required"}
	}
	if req.Quantity == nil && req.Amount == nil {
		return &domain.ErrInvalidDecisionRequest{Field: "quantity", Detail: "quantity or amount is required"}
	}
	if req.Quantity != nil && !req.Quantity.IsPositive() {
		return &domain.ErrInvalidDecisionRequest{Field: "quantity", Detail: "must be positive"}
	}
	if req.Amount != nil && !req.Amount.IsPositive() {
		return &domain.ErrInvalidDecisionRequest{Field: "amount", Detail: "must be positive"}
	}
	if req.LimitPrice != nil && !req.LimitPrice.IsPositive() {
		return &domain.ErrInvalidDecisionRequest{Field: "limit_price", Detail: "must be positive"}
	}
	return nil
}

// decisionComplianceOrderValues converts the decision ticket into the
// quantity + unit-price shape required by the compliance contract.
//
// An explicit amount is the authoritative proposed notional. When quantity
// is also available, amount/quantity is therefore the effective unit price;
// this keeps Quantity*Price equal to the amount the operator entered. A
// quantity-only decision falls back to its positive limit price. Amount-only
// decisions fail closed because quantity-based rules cannot be evaluated
// safely without units.
func decisionComplianceOrderValues(d *entity.Decision) (decimal.Decimal, decimal.Decimal, error) {
	if d.Quantity == nil || !d.Quantity.IsPositive() {
		return decimal.Zero, decimal.Zero, &domain.ErrInvalidDecisionRequest{
			Field:  "quantity",
			Detail: "a positive quantity is required for pre-trade compliance",
		}
	}

	qty := *d.Quantity
	if d.Amount != nil {
		if !d.Amount.IsPositive() {
			return decimal.Zero, decimal.Zero, &domain.ErrInvalidDecisionRequest{
				Field:  "amount",
				Detail: "must be positive",
			}
		}
		return qty, d.Amount.Div(qty), nil
	}

	if d.LimitPrice == nil || !d.LimitPrice.IsPositive() {
		return decimal.Zero, decimal.Zero, &domain.ErrInvalidDecisionRequest{
			Field:  "limit_price",
			Detail: "a positive limit price or amount is required for pre-trade compliance",
		}
	}
	return qty, *d.LimitPrice, nil
}
