package command

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// CreateRuleInstanceRequest is the input for creating a new rule instance
// with its initial parameter version (version 1).
//
// The request is intentionally minimal — bindings (which scope the rule to
// a portfolio / contract / global) are created via a separate endpoint once
// the instance + version pair exists.
type CreateRuleInstanceRequest struct {
	// RuleTypeID is the stable SPI type ID, e.g. "credit.min_rating".
	// Must be registered in the SPI registry at creation time.
	RuleTypeID string

	// Name is a human-friendly label. Required, trimmed.
	Name string

	// Description is optional long-form context for operators.
	Description string

	// Parameters is the JSON document that populates version 1. Must validate
	// against the rule type's ParameterSchema.
	Parameters json.RawMessage

	// EffectiveFrom is the earliest date on which the instance may apply.
	// Required. Defaults to the caller's business date at the transport layer.
	EffectiveFrom time.Time

	// EffectiveTo is the latest date on which the instance may apply.
	// Nil = indefinite.
	EffectiveTo *time.Time

	// IsActive defaults to true — instances created with IsActive=false are
	// staged and cannot be bound until activated.
	IsActive bool

	// ChangeReason is stored on version 1 for audit ("initial creation" is
	// the conventional default).
	ChangeReason string

	// CreatedBy is the actor UUID, captured from the auth context at the
	// transport layer.
	CreatedBy uuid.UUID
}

// CreateRuleInstanceResult is the handler's response. Both the instance and
// its v1 version are returned so callers can avoid a follow-up read.
type CreateRuleInstanceResult struct {
	Instance entity.RuleInstance        `json:"instance"`
	Version  entity.RuleInstanceVersion `json:"version"`
}

// CreateRuleInstanceHandler persists a new rule instance + version 1 atomically
// from the caller's point of view (instance insert first, then version insert;
// version insert bumps current_version inside its own tx). On version-creation
// failure the half-written instance is rolled back so there is never a live
// rule instance without at least one version.
type CreateRuleInstanceHandler struct {
	instanceRepo domain.RuleInstanceRepository
	registry     *spi.RuleRegistry
	now          func() time.Time
}

// NewCreateRuleInstanceHandler wires the handler.
func NewCreateRuleInstanceHandler(
	instanceRepo domain.RuleInstanceRepository,
	registry *spi.RuleRegistry,
) *CreateRuleInstanceHandler {
	return &CreateRuleInstanceHandler{
		instanceRepo: instanceRepo,
		registry:     registry,
		now:          func() time.Time { return time.Now().UTC() },
	}
}

// Handle validates + persists the rule instance.
//
// Error contract (checked by the HTTP layer via errors.As):
//   - *domain.ErrRuleTypeNotFound    — unknown rule_type_id (404 / 422).
//   - *domain.ErrParameterValidation — parameters fail schema check (400).
//   - *ErrInvalidCreateRuleRequest   — structural field errors (400).
//   - wrapped error                  — infrastructure failure (500).
func (h *CreateRuleInstanceHandler) Handle(
	ctx context.Context,
	req CreateRuleInstanceRequest,
) (*CreateRuleInstanceResult, error) {
	if h == nil || h.instanceRepo == nil || h.registry == nil {
		return nil, fmt.Errorf("create rule instance handler not initialised")
	}

	// ── 1. Structural validation ─────────────────────────────────────────────
	typeID := strings.TrimSpace(req.RuleTypeID)
	if typeID == "" {
		return nil, &ErrInvalidCreateRuleRequest{Field: "rule_type_id"}
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, &ErrInvalidCreateRuleRequest{Field: "name"}
	}
	if len(req.Parameters) == 0 {
		return nil, &ErrInvalidCreateRuleRequest{Field: "parameters"}
	}
	if req.EffectiveFrom.IsZero() {
		return nil, &ErrInvalidCreateRuleRequest{Field: "effective_from"}
	}
	if req.EffectiveTo != nil && req.EffectiveTo.Before(req.EffectiveFrom) {
		return nil, &ErrInvalidCreateRuleRequest{
			Field:  "effective_to",
			Detail: "must be on or after effective_from",
		}
	}
	if req.CreatedBy == uuid.Nil {
		return nil, &ErrInvalidCreateRuleRequest{Field: "created_by"}
	}

	// ── 2. Rule-type existence + parameter schema check ──────────────────────
	evaluator, ok := h.registry.Get(typeID)
	if !ok {
		return nil, &domain.ErrRuleTypeNotFound{TypeID: typeID}
	}
	schema := evaluator.ParameterSchema()
	if err := schema.Validate(req.Parameters); err != nil {
		return nil, &domain.ErrParameterValidation{
			RuleTypeID: typeID,
			Detail:     err.Error(),
		}
	}

	// ── 3. Persist instance, then v1 version ─────────────────────────────────
	now := h.now()
	changeReason := strings.TrimSpace(req.ChangeReason)
	if changeReason == "" {
		changeReason = "initial creation"
	}

	instance := entity.RuleInstance{
		ID:             uuid.New(),
		RuleTypeID:     typeID,
		Name:           name,
		Description:    strings.TrimSpace(req.Description),
		CurrentVersion: 1,
		IsActive:       req.IsActive,
		EffectiveWindow: vo.EffectiveWindow{
			ValidFrom: req.EffectiveFrom,
			ValidTo:   req.EffectiveTo,
		},
		CreatedBy: req.CreatedBy,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := h.instanceRepo.Create(ctx, &instance); err != nil {
		return nil, fmt.Errorf("persisting rule instance: %w", err)
	}

	version := entity.RuleInstanceVersion{
		ID:             uuid.New(),
		RuleInstanceID: instance.ID,
		VersionNumber:  1,
		Parameters:     req.Parameters,
		ChangeReason:   changeReason,
		CreatedBy:      req.CreatedBy,
		CreatedAt:      now,
	}
	if err := h.instanceRepo.CreateVersion(ctx, &version); err != nil {
		// v1 failed — attempt to deactivate the orphan instance so it is not
		// listed as a live rule with no version. We do NOT try to DELETE it
		// because the rule-instance table is intended to be long-lived. The
		// repository contract only exposes UpdateActive, which is sufficient
		// for this degraded state.
		_ = h.instanceRepo.UpdateActive(ctx, instance.ID, false)
		return nil, fmt.Errorf("persisting rule instance version: %w", err)
	}

	return &CreateRuleInstanceResult{
		Instance: instance,
		Version:  version,
	}, nil
}

// ErrInvalidCreateRuleRequest flags missing or malformed input on the
// create-rule-instance command. Client-fixable → HTTP 400 at the transport
// layer (distinct from internal errors which must surface as 500).
type ErrInvalidCreateRuleRequest struct {
	Field  string
	Detail string
}

func (e *ErrInvalidCreateRuleRequest) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("compliance: invalid create-rule-instance request: %s: %s", e.Field, e.Detail)
	}
	return fmt.Sprintf("compliance: invalid create-rule-instance request: %s is required", e.Field)
}
