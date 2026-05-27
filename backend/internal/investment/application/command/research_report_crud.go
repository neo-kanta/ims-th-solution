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

// UpdateResearchReportRequest patches mutable fields on a research report.
// Only non-nil pointer fields are applied.
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

	RejectionReason    *string
	PostSubmissionNote *string

	ReportStatus *vo.ReportStatus

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
	if req.RejectionReason != nil {
		r.RejectionReason = *req.RejectionReason
	}
	if req.PostSubmissionNote != nil {
		r.PostSubmissionNote = *req.PostSubmissionNote
	}
	if req.ReportStatus != nil {
		if !req.ReportStatus.IsValid() {
			return nil, &domain.ErrInvalidResearchReportRequest{Field: "report_status", Detail: "invalid"}
		}
		r.ReportStatus = *req.ReportStatus
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

// Submit moves review_status from NOT_SUBMITTED to SUBMITTED. No real
// approval workflow is invoked — that ships separately.
func (h *ResearchReportCommandHandler) Submit(
	ctx context.Context,
	id uuid.UUID,
	actorID uuid.UUID,
) (*entity.ResearchReport, error) {
	return h.transitionReview(ctx, id, actorID, vo.ReviewStatusSubmitted, "INVESTMENT_RESEARCH_SUBMITTED")
}

// CancelSubmit moves review_status from SUBMITTED back to NOT_SUBMITTED.
// Refuses once review has been completed.
func (h *ResearchReportCommandHandler) CancelSubmit(
	ctx context.Context,
	id uuid.UUID,
	actorID uuid.UUID,
) (*entity.ResearchReport, error) {
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

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      actorID.String(),
		Action:       auditAction,
		Module:       "investment",
		ResourceType: "INVESTMENT_RESEARCH_REPORT",
		ResourceID:   r.ID.String(),
		Details:      map[string]any{"review_status": string(target)},
		BusinessDate: now,
	})

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
