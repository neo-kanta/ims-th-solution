package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// CreateResearchReportRequest is the input for creating a research report
// (always in DRAFT / NOT_SUBMITTED).
type CreateResearchReportRequest struct {
	ReportNo             string
	ReportDate           time.Time
	EffectiveDate        *time.Time
	OwnerUserID          uuid.UUID
	AuthorUserID         uuid.UUID
	ApplicableContractID *uuid.UUID

	InstrumentType string
	InstrumentCode string
	InstrumentName string
	Market         string
	Currency       string

	Recommendation vo.Recommendation
	ReportTitle    string

	CompanyOverview    string
	CompanyOutlook     string
	ESGComment         string
	FinancialStatus    string
	InvestmentAnalysis string

	ActorID uuid.UUID
}

// UpdateResearchReportRequest patches mutable content fields on a research
// report. Only non-nil pointer fields are applied.
//
// Lifecycle fields (ReportStatus, ReviewStatus, RejectionReason) are
// intentionally NOT updatable here. Status transitions must go through the
// dedicated commands (Submit / CancelSubmit / ApplyApprovalDecision) so that
// every status flip runs through the policy gate and emits a status-specific
// audit event. Accepting them on the generic edit path would let any caller
// with INVESTMENT_RESEARCH_UPDATE flip REJECTED → ACTIVE or EXPIRED → ACTIVE
// without going through approval.
type UpdateResearchReportRequest struct {
	ReportID uuid.UUID

	ReportDate           *time.Time
	EffectiveDate        *time.Time
	OwnerUserID          *uuid.UUID
	AuthorUserID         *uuid.UUID
	ApplicableContractID *uuid.UUID

	InstrumentType *string
	InstrumentCode *string
	InstrumentName *string
	Market         *string
	Currency       *string

	Recommendation *vo.Recommendation
	ReportTitle    *string

	CompanyOverview    *string
	CompanyOutlook     *string
	ESGComment         *string
	FinancialStatus    *string
	InvestmentAnalysis *string

	PostSubmissionNote *string

	ActorID uuid.UUID
}

// ResearchReportCommandHandler owns the write-side use cases for research
// reports: create / update / soft-delete and the simple submit / cancel-submit
// status transitions.
//
// runTx is an injectable transaction runner. Production wiring delegates
// to withTransaction(ctx, pool, fn); tests can substitute a runner that
// passes a nil pgx.Tx through to the in-memory repository fake. Keeping
// the hook field-local to this handler avoids any change to the shared
// withTransaction helper or to unrelated commands.
type ResearchReportCommandHandler struct {
	pool    *pgxpool.Pool
	reports domain.ResearchReportRepository
	audit   contract.AuditLogger
	now     func() time.Time
	runTx   func(ctx context.Context, fn func(pgx.Tx) error) error

	// approval is an optional hook into the generic Approval Module. When set
	// (wired in production), submitting a report creates a real approval
	// request and the report's lifecycle is driven by the approval outcome.
	// When nil (tests / approval not wired), Submit behaves as before.
	approval contract.ApprovalSubmitter

	// approvalCanceller terminates any active approval request for a subject
	// when the subject itself is cancelled or withdrawn. Optional — safe to
	// leave nil in tests or when approval is not wired.
	approvalCanceller contract.ApprovalCanceller
}

// SetApprovalSubmitter injects the approval submitter after construction. This
// keeps the constructor signature (and its existing test call sites) unchanged
// while letting production wiring connect the Approval Module.
func (h *ResearchReportCommandHandler) SetApprovalSubmitter(s contract.ApprovalSubmitter) {
	if h != nil {
		h.approval = s
	}
}

// SetApprovalCanceller injects the approval canceller after construction.
func (h *ResearchReportCommandHandler) SetApprovalCanceller(c contract.ApprovalCanceller) {
	if h != nil {
		h.approvalCanceller = c
	}
}

// NewResearchReportCommandHandler wires the handler.
func NewResearchReportCommandHandler(
	pool *pgxpool.Pool,
	reports domain.ResearchReportRepository,
	audit contract.AuditLogger,
	now func() time.Time,
) *ResearchReportCommandHandler {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &ResearchReportCommandHandler{
		pool:    pool,
		reports: reports,
		audit:   audit,
		now:     now,
		runTx: func(ctx context.Context, fn func(pgx.Tx) error) error {
			return withTransaction(ctx, pool, fn)
		},
	}
}

// Create persists a new research report. report_no must be unique among
// non-deleted rows; if blank, the caller is expected to supply one — auto
// numbering belongs to a follow-up.
func (h *ResearchReportCommandHandler) Create(
	ctx context.Context,
	req CreateResearchReportRequest,
) (*entity.ResearchReport, error) {
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidResearchReportRequest{Field: "actor_id", Detail: "is required"}
	}

	// PoC: default to actor. Role-based override TBD when iam roles for
	// research are defined.
	if req.AuthorUserID == uuid.Nil {
		req.AuthorUserID = req.ActorID
	}
	if req.OwnerUserID == uuid.Nil {
		req.OwnerUserID = req.ActorID
	}

	// Length checks fire before the policy so the operator gets a sharp
	// "max 60 characters" error instead of a generic "is required" when
	// they overrun a column limit.
	if e := validateLengths([]fieldSpec{
		{"report_no", strings.TrimSpace(req.ReportNo), 60},
		{"instrument_code", strings.TrimSpace(req.InstrumentCode), 40},
		{"instrument_name", req.InstrumentName, 255},
		{"instrument_type", req.InstrumentType, 40},
		{"market", req.Market, 40},
		{"report_title", req.ReportTitle, 255},
	}); e != nil {
		return nil, e
	}
	if e := validateCurrency(req.Currency); e != nil {
		return nil, e
	}

	policyErr := policy.ValidateResearchReport(policy.ResearchReportInput{
		ReportNo:           req.ReportNo,
		ReportDate:         req.ReportDate,
		OwnerUserID:        req.OwnerUserID,
		AuthorUserID:       req.AuthorUserID,
		InstrumentCode:     req.InstrumentCode,
		Recommendation:     req.Recommendation,
		InvestmentAnalysis: req.InvestmentAnalysis,
	})
	if policyErr != nil {
		return nil, &domain.ErrInvalidResearchReportRequest{Field: policyErr.Field, Detail: policyErr.Detail}
	}

	reportNo := strings.TrimSpace(req.ReportNo)
	if reportNo == "" {
		return nil, &domain.ErrInvalidResearchReportRequest{Field: "report_no", Detail: "is required"}
	}

	existing, err := h.reports.GetByReportNo(ctx, reportNo)
	if err != nil {
		return nil, fmt.Errorf("checking existing report_no: %w", err)
	}
	if existing != nil {
		return nil, &domain.ErrResearchReportNoAlreadyExists{ReportNo: reportNo}
	}

	now := h.now()
	actor := req.ActorID
	r := &entity.ResearchReport{
		ID:                   uuid.New(),
		ReportNo:             reportNo,
		ReportDate:           req.ReportDate.UTC(),
		EffectiveDate:        toUTCPtr(req.EffectiveDate),
		OwnerUserID:          req.OwnerUserID,
		AuthorUserID:         req.AuthorUserID,
		ApplicableContractID: req.ApplicableContractID,
		InstrumentType:       strings.TrimSpace(req.InstrumentType),
		InstrumentCode:       strings.ToUpper(strings.TrimSpace(req.InstrumentCode)),
		InstrumentName:       strings.TrimSpace(req.InstrumentName),
		Market:               strings.TrimSpace(req.Market),
		Currency:             strings.ToUpper(strings.TrimSpace(req.Currency)),
		Recommendation:       req.Recommendation,
		ReportTitle:          strings.TrimSpace(req.ReportTitle),
		CompanyOverview:      req.CompanyOverview,
		CompanyOutlook:       req.CompanyOutlook,
		ESGComment:           req.ESGComment,
		FinancialStatus:      req.FinancialStatus,
		InvestmentAnalysis:   req.InvestmentAnalysis,
		ReportStatus:         vo.ReportStatusDraft,
		ReviewStatus:         vo.ReviewStatusNotSubmitted,
		CreatedAt:            now,
		UpdatedAt:            now,
		CreatedBy:            &actor,
		UpdatedBy:            &actor,
	}

	if err := h.runTx(ctx, func(tx pgx.Tx) error {
		return h.reports.Create(ctx, tx, r)
	}); err != nil {
		return nil, err
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_RESEARCH_CREATED",
		Module:       "investment",
		ResourceType: "INVESTMENT_RESEARCH_REPORT",
		ResourceID:   r.ID.String(),
		Details:      map[string]any{"report_no": r.ReportNo, "instrument_code": r.InstrumentCode},
		BusinessDate: now,
	})

	return r, nil
}

// Update applies the requested patches. Refuses when the report is deleted
// or when the review has already been completed.
func (h *ResearchReportCommandHandler) Update(
	ctx context.Context,
	req UpdateResearchReportRequest,
) (*entity.ResearchReport, error) {
	if req.ReportID == uuid.Nil {
		return nil, &domain.ErrInvalidResearchReportRequest{Field: "id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidResearchReportRequest{Field: "actor_id", Detail: "is required"}
	}

	// Length / format checks on whatever the caller is patching. We only
	// validate the fields the caller set (pointer != nil) so partial
	// updates don't fail on untouched columns.
	if req.InstrumentCode != nil {
		if e := validateLengths([]fieldSpec{
			{"instrument_code", strings.TrimSpace(*req.InstrumentCode), 40},
		}); e != nil {
			return nil, e
		}
	}
	if req.InstrumentName != nil {
		if e := validateLengths([]fieldSpec{
			{"instrument_name", *req.InstrumentName, 255},
		}); e != nil {
			return nil, e
		}
	}
	if req.InstrumentType != nil {
		if e := validateLengths([]fieldSpec{
			{"instrument_type", *req.InstrumentType, 40},
		}); e != nil {
			return nil, e
		}
	}
	if req.Market != nil {
		if e := validateLengths([]fieldSpec{
			{"market", *req.Market, 40},
		}); e != nil {
			return nil, e
		}
	}
	if req.ReportTitle != nil {
		if e := validateLengths([]fieldSpec{
			{"report_title", *req.ReportTitle, 255},
		}); e != nil {
			return nil, e
		}
	}
	if req.Currency != nil {
		if e := validateCurrency(*req.Currency); e != nil {
			return nil, e
		}
	}

	r, err := h.reports.GetByID(ctx, req.ReportID)
	if err != nil {
		return nil, fmt.Errorf("loading research report: %w", err)
	}
	if r == nil {
		return nil, &domain.ErrResearchReportNotFound{ReportID: req.ReportID.String()}
	}
	if !r.CanUpdate() {
		return nil, &domain.ErrResearchReportCannotUpdate{
			ReportID:     r.ID.String(),
			ReviewStatus: string(r.ReviewStatus),
		}
	}

	if req.ReportDate != nil {
		r.ReportDate = req.ReportDate.UTC()
	}
	if req.EffectiveDate != nil {
		r.EffectiveDate = toUTCPtr(req.EffectiveDate)
	}
	if req.OwnerUserID != nil {
		r.OwnerUserID = *req.OwnerUserID
	}
	if req.AuthorUserID != nil {
		r.AuthorUserID = *req.AuthorUserID
	}
	if req.ApplicableContractID != nil {
		r.ApplicableContractID = req.ApplicableContractID
	}
	if req.InstrumentType != nil {
		r.InstrumentType = strings.TrimSpace(*req.InstrumentType)
	}
	if req.InstrumentCode != nil {
		r.InstrumentCode = strings.ToUpper(strings.TrimSpace(*req.InstrumentCode))
	}
	if req.InstrumentName != nil {
		r.InstrumentName = strings.TrimSpace(*req.InstrumentName)
	}
	if req.Market != nil {
		r.Market = strings.TrimSpace(*req.Market)
	}
	if req.Currency != nil {
		r.Currency = strings.ToUpper(strings.TrimSpace(*req.Currency))
	}
	if req.Recommendation != nil {
		r.Recommendation = *req.Recommendation
	}
	if req.ReportTitle != nil {
		r.ReportTitle = strings.TrimSpace(*req.ReportTitle)
	}
	if req.CompanyOverview != nil {
		r.CompanyOverview = *req.CompanyOverview
	}
	if req.CompanyOutlook != nil {
		r.CompanyOutlook = *req.CompanyOutlook
	}
	if req.ESGComment != nil {
		r.ESGComment = *req.ESGComment
	}
	if req.FinancialStatus != nil {
		r.FinancialStatus = *req.FinancialStatus
	}
	if req.InvestmentAnalysis != nil {
		r.InvestmentAnalysis = *req.InvestmentAnalysis
	}
	if req.PostSubmissionNote != nil {
		r.PostSubmissionNote = *req.PostSubmissionNote
	}

	// Re-validate the post-patch entity against the same policy used on create
	// to catch e.g. truncating investment_analysis below the minimum length.
	if policyErr := policy.ValidateResearchReport(policy.ResearchReportInput{
		ReportNo:           r.ReportNo,
		ReportDate:         r.ReportDate,
		OwnerUserID:        r.OwnerUserID,
		AuthorUserID:       r.AuthorUserID,
		InstrumentCode:     r.InstrumentCode,
		Recommendation:     r.Recommendation,
		InvestmentAnalysis: r.InvestmentAnalysis,
	}); policyErr != nil {
		return nil, &domain.ErrInvalidResearchReportRequest{Field: policyErr.Field, Detail: policyErr.Detail}
	}

	now := h.now()
	actor := req.ActorID
	r.UpdatedAt = now
	r.UpdatedBy = &actor

	if err := h.runTx(ctx, func(tx pgx.Tx) error {
		return h.reports.Update(ctx, tx, r)
	}); err != nil {
		return nil, err
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_RESEARCH_UPDATED",
		Module:       "investment",
		ResourceType: "INVESTMENT_RESEARCH_REPORT",
		ResourceID:   r.ID.String(),
		BusinessDate: now,
	})

	return r, nil
}

// SoftDelete marks the report as deleted. Refuses to delete a report that
// has been submitted or whose review has been completed.
func (h *ResearchReportCommandHandler) SoftDelete(
	ctx context.Context,
	id uuid.UUID,
	actorID uuid.UUID,
) error {
	if id == uuid.Nil {
		return &domain.ErrInvalidResearchReportRequest{Field: "id", Detail: "is required"}
	}
	if actorID == uuid.Nil {
		return &domain.ErrInvalidResearchReportRequest{Field: "actor_id", Detail: "is required"}
	}

	r, err := h.reports.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("loading research report: %w", err)
	}
	if r == nil {
		return &domain.ErrResearchReportNotFound{ReportID: id.String()}
	}
	if !r.CanDelete() {
		return &domain.ErrResearchReportCannotDelete{
			ReportID:     r.ID.String(),
			ReviewStatus: string(r.ReviewStatus),
		}
	}

	now := h.now()
	if err := h.runTx(ctx, func(tx pgx.Tx) error {
		return h.reports.SoftDelete(ctx, tx, id, actorID, now)
	}); err != nil {
		return fmt.Errorf("soft-deleting research report: %w", err)
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      actorID.String(),
		Action:       "INVESTMENT_RESEARCH_DELETED",
		Module:       "investment",
		ResourceType: "INVESTMENT_RESEARCH_REPORT",
		ResourceID:   id.String(),
		BusinessDate: now,
	})

	return nil
}

// InvalidateRequest is the input for the Invalidate command.
type InvalidateResearchReportRequest struct {
	ReportID uuid.UUID
	ActorID  uuid.UUID
	Reason   string
}

// minInvalidationReasonLen mirrors the DB CHECK constraint requirement that
// invalidation reason is at least 20 characters.
const minInvalidationReasonLen = 20

// Invalidate marks the report as INVALIDATED (terminal). This is a one-way
// transition; an invalidated report can no longer be referenced by an
// investment decision, edited, submitted, cancelled, or deleted.
//
// Authorisation: callers must hold the INVESTMENT_RESEARCH_INVALIDATE
// function permission — enforced at the route, not here.
//
// Audit: emitted post-commit by the caller via h.audit.LogAction so a failed
// audit insert does not roll back the invalidation (task 5 will tighten this
// for financial actions; this command is treated as financial-grade and the
// audit error IS returned).
func (h *ResearchReportCommandHandler) Invalidate(
	ctx context.Context,
	req InvalidateResearchReportRequest,
) (*entity.ResearchReport, error) {
	if req.ReportID == uuid.Nil {
		return nil, &domain.ErrInvalidResearchReportRequest{Field: "id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidResearchReportRequest{Field: "actor_id", Detail: "is required"}
	}
	reason := strings.TrimSpace(req.Reason)
	if len(reason) < minInvalidationReasonLen {
		return nil, &domain.ErrInvalidResearchReportRequest{
			Field: "reason",
			Detail: fmt.Sprintf(
				"must be at least %d characters (got %d)",
				minInvalidationReasonLen, len(reason),
			),
		}
	}

	r, err := h.reports.GetByID(ctx, req.ReportID)
	if err != nil {
		return nil, fmt.Errorf("loading research report: %w", err)
	}
	if r == nil {
		return nil, &domain.ErrResearchReportNotFound{ReportID: req.ReportID.String()}
	}
	if !r.CanInvalidate() {
		return nil, &domain.ErrResearchReportCannotInvalidate{
			ReportID:     r.ID.String(),
			ReportStatus: string(r.ReportStatus),
			ReviewStatus: string(r.ReviewStatus),
		}
	}

	now := h.now()
	if err := h.runTx(ctx, func(tx pgx.Tx) error {
		return h.reports.Invalidate(ctx, tx, r.ID, req.ActorID, reason, now)
	}); err != nil {
		return nil, err
	}

	// Surface the audit failure for this financial action — invalidating a
	// research report must always leave an audit trail.
	if err := h.audit.LogActionStrict(ctx, contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_RESEARCH_INVALIDATED",
		Module:       "investment",
		ResourceType: "INVESTMENT_RESEARCH_REPORT",
		ResourceID:   r.ID.String(),
		Details: map[string]any{
			"reason":        reason,
			"report_no":     r.ReportNo,
			"prior_status":  string(r.ReportStatus),
			"prior_review":  string(r.ReviewStatus),
		},
		BusinessDate: now,
	}); err != nil {
		return nil, fmt.Errorf("auditing invalidation: %w", err)
	}

	// Refresh and return the canonical row.
	out, err := h.reports.GetByID(ctx, r.ID)
	if err != nil {
		return nil, fmt.Errorf("re-loading invalidated report: %w", err)
	}
	return out, nil
}

// Submit moves review_status from NOT_SUBMITTED to SUBMITTED. When the Approval
// Module is wired (h.approval != nil), it also creates a real approval request
// (process INVESTMENT_ANALYSIS_REPORT). Approval is created first so that, if no
// approval process is configured or the resolver fails, the report is NOT moved
// to SUBMITTED and the caller receives a clean error.
func (h *ResearchReportCommandHandler) Submit(
	ctx context.Context,
	id uuid.UUID,
	actorID uuid.UUID,
) (*entity.ResearchReport, error) {
	if h.approval != nil {
		if id == uuid.Nil {
			return nil, &domain.ErrInvalidResearchReportRequest{Field: "id", Detail: "is required"}
		}
		if actorID == uuid.Nil {
			return nil, &domain.ErrInvalidResearchReportRequest{Field: "actor_id", Detail: "is required"}
		}
		r, err := h.reports.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("loading research report: %w", err)
		}
		if r == nil {
			return nil, &domain.ErrResearchReportNotFound{ReportID: id.String()}
		}
		if !r.CanSubmit() {
			return nil, &domain.ErrResearchReportCannotSubmit{ReportID: r.ID.String(), ReviewStatus: string(r.ReviewStatus)}
		}
		title := strings.TrimSpace(r.ReportTitle)
		if title == "" {
			title = r.ReportNo
		}
		contractType := "COMPANY"
		if r.ApplicableContractID != nil {
			contractType = "FUND"
		}
		if _, err := h.approval.SubmitForApproval(ctx, contract.ApprovalSubmission{
			ProcessType:      "INVESTMENT_ANALYSIS_REPORT",
			SubjectType:      "RESEARCH_REPORT",
			SubjectID:        r.ID,
			SubjectTitle:     title,
			SubjectReference: r.ReportNo,
			ContractType:     contractType,
			ContractID:       r.ApplicableContractID,
			SubmitterID:      actorID,
		}); err != nil {
			return nil, err
		}
	}
	return h.transitionReview(ctx, id, actorID, vo.ReviewStatusSubmitted, "INVESTMENT_RESEARCH_SUBMITTED")
}

// ApplyApprovalDecision updates a research report's lifecycle when its approval
// request reaches a final decision. Approved → ACTIVE / REVIEW_COMPLETED;
// rejected → REJECTED with the reason stored and review reset so it can be
// revised. Invoked by the Approval Module via the registered subject callback.
func (h *ResearchReportCommandHandler) ApplyApprovalDecision(
	ctx context.Context,
	reportID uuid.UUID,
	approved bool,
	reason string,
) error {
	r, err := h.reports.GetByID(ctx, reportID)
	if err != nil {
		return fmt.Errorf("loading research report: %w", err)
	}
	if r == nil {
		return &domain.ErrResearchReportNotFound{ReportID: reportID.String()}
	}
	now := h.now()
	if approved {
		r.ReviewStatus = vo.ReviewStatusReviewCompleted
		r.ReportStatus = vo.ReportStatusActive
	} else {
		r.ReportStatus = vo.ReportStatusRejected
		r.ReviewStatus = vo.ReviewStatusNotSubmitted
		r.RejectionReason = strings.TrimSpace(reason)
	}
	r.UpdatedAt = now
	r.UpdatedBy = nil
	if err := h.runTx(ctx, func(tx pgx.Tx) error { return h.reports.Update(ctx, tx, r) }); err != nil {
		return err
	}
	action := "INVESTMENT_RESEARCH_APPROVAL_APPROVED"
	if !approved {
		action = "INVESTMENT_RESEARCH_APPROVAL_REJECTED"
	}
	// Approval-callback drives the report lifecycle — surface the audit error
	// so a silent loss never leaves the operator confused about who approved
	// the report.
	if err := h.audit.LogActionStrict(ctx, contract.AuditEntry{
		Action:       action,
		Module:       "investment",
		ResourceType: "INVESTMENT_RESEARCH_REPORT",
		ResourceID:   r.ID.String(),
		Details:      map[string]any{"approved": approved, "reason": reason},
		BusinessDate: now,
	}); err != nil {
		return fmt.Errorf("auditing approval callback: %w", err)
	}
	return nil
}

// CancelSubmit moves review_status from SUBMITTED back to NOT_SUBMITTED.
// Refuses once review has been completed. When an approval request is in-flight,
// it is cancelled first so the two states stay in sync.
func (h *ResearchReportCommandHandler) CancelSubmit(
	ctx context.Context,
	id uuid.UUID,
	actorID uuid.UUID,
) (*entity.ResearchReport, error) {
	if h.approvalCanceller != nil {
		if err := h.approvalCanceller.CancelApprovalBySubject(ctx, "RESEARCH_REPORT", id, actorID); err != nil {
			return nil, fmt.Errorf("cancelling approval request: %w", err)
		}
	}
	return h.transitionReview(ctx, id, actorID, vo.ReviewStatusNotSubmitted, "INVESTMENT_RESEARCH_SUBMIT_CANCELLED")
}

func (h *ResearchReportCommandHandler) transitionReview(
	ctx context.Context,
	id uuid.UUID,
	actorID uuid.UUID,
	target vo.ReviewStatus,
	auditAction string,
) (*entity.ResearchReport, error) {
	if id == uuid.Nil {
		return nil, &domain.ErrInvalidResearchReportRequest{Field: "id", Detail: "is required"}
	}
	if actorID == uuid.Nil {
		return nil, &domain.ErrInvalidResearchReportRequest{Field: "actor_id", Detail: "is required"}
	}

	r, err := h.reports.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("loading research report: %w", err)
	}
	if r == nil {
		return nil, &domain.ErrResearchReportNotFound{ReportID: id.String()}
	}

	switch target {
	case vo.ReviewStatusSubmitted:
		if !r.CanSubmit() {
			return nil, &domain.ErrResearchReportCannotSubmit{
				ReportID:     r.ID.String(),
				ReviewStatus: string(r.ReviewStatus),
			}
		}
	case vo.ReviewStatusNotSubmitted:
		if !r.CanCancelSubmit() {
			return nil, &domain.ErrResearchReportCannotCancelSubmit{
				ReportID:     r.ID.String(),
				ReviewStatus: string(r.ReviewStatus),
			}
		}
	default:
		return nil, &domain.ErrInvalidResearchReportRequest{
			Field:  "target_review_status",
			Detail: "unsupported transition",
		}
	}

	now := h.now()
	actor := actorID
	r.ReviewStatus = target
	r.UpdatedAt = now
	r.UpdatedBy = &actor

	if err := h.runTx(ctx, func(tx pgx.Tx) error {
		return h.reports.Update(ctx, tx, r)
	}); err != nil {
		return nil, err
	}

	// Submit / cancel-submit are review lifecycle changes — financial-grade
	// because they gate the report's referenceability by decisions.
	if err := h.audit.LogActionStrict(ctx, contract.AuditEntry{
		ActorID:      actorID.String(),
		Action:       auditAction,
		Module:       "investment",
		ResourceType: "INVESTMENT_RESEARCH_REPORT",
		ResourceID:   r.ID.String(),
		Details:      map[string]any{"review_status": string(target)},
		BusinessDate: now,
	}); err != nil {
		return nil, fmt.Errorf("auditing review transition: %w", err)
	}

	return r, nil
}

func toUTCPtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	u := t.UTC()
	return &u
}

// fieldSpec describes one bounded text field for validateLengths. Keeping
// the struct private prevents accidental cross-package reuse; the lengths
// are owned by the research report aggregate.
type fieldSpec struct {
	field string
	value string
	max   int
}

// validateLengths returns the first overrun as a typed bad-request error,
// or nil when every field fits its budget. We measure the original string
// (not a trimmed copy) so an operator pasting trailing whitespace into a
// column-bounded field still sees the same error the DB would emit.
func validateLengths(specs []fieldSpec) *domain.ErrInvalidResearchReportRequest {
	for _, s := range specs {
		if len(s.value) > s.max {
			return &domain.ErrInvalidResearchReportRequest{
				Field:  s.field,
				Detail: fmt.Sprintf("max %d characters", s.max),
			}
		}
	}
	return nil
}

// validateCurrency enforces the same ISO-3 currency-code rule the Postgres
// CHECK constraint applies (currency = ” OR currency ~ '^[A-Z]{3}$'),
// uppercasing the input before checking so the API contract stays
// case-insensitive.
func validateCurrency(v string) *domain.ErrInvalidResearchReportRequest {
	u := strings.ToUpper(strings.TrimSpace(v))
	if u == "" {
		return nil
	}
	if len(u) != 3 {
		return &domain.ErrInvalidResearchReportRequest{
			Field:  "currency",
			Detail: "must be 3 uppercase letters",
		}
	}
	for _, r := range u {
		if r < 'A' || r > 'Z' {
			return &domain.ErrInvalidResearchReportRequest{
				Field:  "currency",
				Detail: "must be 3 uppercase letters",
			}
		}
	}
	return nil
}
