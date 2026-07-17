package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

// pgUniqueViolation is the SQLSTATE for unique_violation. Mirrors the
// constant in infrastructure/persistence/override_repository.go — the
// application layer needs its own copy since it does not import persistence.
const pgUniqueViolation = "23505"

// CreateRuleBindingRequest binds an existing rule instance to a scope.
//
// The compliance_rule_bindings table already supports every scope in
// vo.ScopeType; this command is scope-agnostic. Portfolio Compliance V2 always
// calls it with ScopeType=PORTFOLIO and ScopeID=the portfolio's ID (see
// transport's portfolio contract adapter).
type CreateRuleBindingRequest struct {
	RuleInstanceID uuid.UUID
	ScopeType      vo.ScopeType
	ScopeID        *uuid.UUID // nil only for GLOBAL scope
	Severity       vo.Severity
	Priority       int // 0 defaults to 100 (lower = higher priority)
	EffectiveFrom  time.Time
	EffectiveTo    *time.Time
	CreatedBy      uuid.UUID
}

// CreateRuleBindingHandler persists a new rule binding after validating the
// rule instance exists/is active and no active duplicate binding already
// exists for the same rule instance + scope.
type CreateRuleBindingHandler struct {
	bindingRepo  domain.RuleBindingRepository
	instanceRepo domain.RuleInstanceRepository
	now          func() time.Time
}

// NewCreateRuleBindingHandler wires the handler.
func NewCreateRuleBindingHandler(
	bindingRepo domain.RuleBindingRepository,
	instanceRepo domain.RuleInstanceRepository,
) *CreateRuleBindingHandler {
	return &CreateRuleBindingHandler{
		bindingRepo:  bindingRepo,
		instanceRepo: instanceRepo,
		now:          func() time.Time { return time.Now().UTC() },
	}
}

// Handle validates and persists the binding.
//
// Error contract (checked by callers via errors.As):
//   - *ErrInvalidCreateBindingRequest — structural field errors (400).
//   - *domain.ErrRuleInstanceNotFound — rule_instance_id does not exist (404/422).
//   - *domain.ErrRuleInstanceInactive — rule_instance_id exists but is inactive (409).
//   - *domain.ErrDuplicateActiveBinding — active binding already exists (409).
//   - wrapped error — infrastructure failure (500).
func (h *CreateRuleBindingHandler) Handle(ctx context.Context, req CreateRuleBindingRequest) (*entity.RuleBinding, error) {
	if h == nil || h.bindingRepo == nil || h.instanceRepo == nil {
		return nil, fmt.Errorf("create rule binding handler not initialised")
	}

	if err := validateCreateBindingRequest(req); err != nil {
		return nil, err
	}

	instance, err := h.instanceRepo.GetByID(ctx, req.RuleInstanceID)
	if err != nil {
		return nil, fmt.Errorf("loading rule instance: %w", err)
	}
	if instance == nil {
		return nil, &domain.ErrRuleInstanceNotFound{ID: req.RuleInstanceID.String()}
	}
	if !instance.IsActive {
		return nil, &domain.ErrRuleInstanceInactive{ID: req.RuleInstanceID.String()}
	}

	// Application-layer duplicate check. The DB unique index
	// (uq_compliance_rb_active_portfolio, PORTFOLIO scope only) is the
	// authoritative guard against a concurrent-request race; this check gives
	// a clean typed error on the common non-concurrent path and covers scopes
	// the index does not (GLOBAL/CONTRACT/etc. are out of scope for now).
	active := true
	scopeType := req.ScopeType
	existing, _, err := h.bindingRepo.List(ctx, domain.BindingFilter{
		RuleInstanceID: &req.RuleInstanceID,
		ScopeType:      &scopeType,
		ScopeID:        req.ScopeID,
		IsActive:       &active,
	})
	if err != nil {
		return nil, fmt.Errorf("checking for duplicate binding: %w", err)
	}
	for _, b := range existing {
		if scopeIDEqual(b.Scope.ID, req.ScopeID) {
			return nil, &domain.ErrDuplicateActiveBinding{
				RuleInstanceID: req.RuleInstanceID.String(),
				ScopeType:      string(req.ScopeType),
				ScopeID:        scopeIDString(req.ScopeID),
			}
		}
	}

	priority := req.Priority
	if priority == 0 {
		priority = 100
	}

	now := h.now()
	binding := &entity.RuleBinding{
		ID:             uuid.New(),
		RuleInstanceID: req.RuleInstanceID,
		Scope:          vo.Scope{Type: req.ScopeType, ID: req.ScopeID},
		Severity:       req.Severity,
		Priority:       priority,
		EffectiveWindow: vo.EffectiveWindow{
			ValidFrom: req.EffectiveFrom,
			ValidTo:   req.EffectiveTo,
		},
		IsActive:  true,
		CreatedBy: req.CreatedBy,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.bindingRepo.Create(ctx, binding); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return nil, &domain.ErrDuplicateActiveBinding{
				RuleInstanceID: req.RuleInstanceID.String(),
				ScopeType:      string(req.ScopeType),
				ScopeID:        scopeIDString(req.ScopeID),
			}
		}
		return nil, fmt.Errorf("persisting rule binding: %w", err)
	}

	return binding, nil
}

func validateCreateBindingRequest(req CreateRuleBindingRequest) error {
	if req.RuleInstanceID == uuid.Nil {
		return &ErrInvalidCreateBindingRequest{Field: "rule_instance_id"}
	}
	switch req.ScopeType {
	case vo.ScopeGlobal, vo.ScopeJurisdiction, vo.ScopeFundCategory,
		vo.ScopeContract, vo.ScopePortfolio, vo.ScopeAssetClass, vo.ScopeInstrumentType:
		// known scope type
	default:
		return &ErrInvalidCreateBindingRequest{Field: "scope_type", Detail: "unknown scope type"}
	}
	if req.ScopeType == vo.ScopeGlobal && req.ScopeID != nil {
		return &ErrInvalidCreateBindingRequest{Field: "scope_id", Detail: "must be empty for GLOBAL scope"}
	}
	if req.ScopeType != vo.ScopeGlobal && (req.ScopeID == nil || *req.ScopeID == uuid.Nil) {
		return &ErrInvalidCreateBindingRequest{Field: "scope_id", Detail: "is required for non-GLOBAL scope"}
	}
	if !req.Severity.IsValid() {
		return &ErrInvalidCreateBindingRequest{Field: "severity", Detail: "unknown severity"}
	}
	if req.EffectiveFrom.IsZero() {
		return &ErrInvalidCreateBindingRequest{Field: "effective_from"}
	}
	if req.EffectiveTo != nil && req.EffectiveTo.Before(req.EffectiveFrom) {
		return &ErrInvalidCreateBindingRequest{Field: "effective_to", Detail: "must be on or after effective_from"}
	}
	if req.CreatedBy == uuid.Nil {
		return &ErrInvalidCreateBindingRequest{Field: "created_by"}
	}
	return nil
}

func scopeIDEqual(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func scopeIDString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

// ErrInvalidCreateBindingRequest flags missing or malformed input on the
// create-rule-binding command. Client-fixable → HTTP 400 at the transport layer.
type ErrInvalidCreateBindingRequest struct {
	Field  string
	Detail string
}

func (e *ErrInvalidCreateBindingRequest) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("compliance: invalid create-rule-binding request: %s: %s", e.Field, e.Detail)
	}
	return fmt.Sprintf("compliance: invalid create-rule-binding request: %s is required", e.Field)
}
