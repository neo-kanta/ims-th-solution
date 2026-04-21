package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
)

// OverrideBreachRequest is the input for a compliance officer override.
type OverrideBreachRequest struct {
	BreachID      uuid.UUID
	Reason        string     // mandatory written justification
	OverriddenBy  uuid.UUID  // actor with IRG_OVERRIDE_BREACH permission
	DelegatedFrom *uuid.UUID // optional: original officer if this is a delegation
	ApprovedBy    *uuid.UUID // optional: second-level approver
}

// OverrideBreachHandler processes a compliance officer breach override.
//
// The whole flow — loading/locking the breach, validating its state, inserting
// the override record, and transitioning the breach to OVERRIDDEN — runs inside
// one database transaction delegated to OverrideRepository.CommitOverride.
// The handler itself performs no persistence and owns no transaction logic;
// it builds the domain entity, supplies the shared UTC timestamp, and surfaces
// typed domain errors back to the transport layer.
type OverrideBreachHandler struct {
	overrideRepo domain.OverrideRepository
}

// NewOverrideBreachHandler creates the handler.
func NewOverrideBreachHandler(overrideRepo domain.OverrideRepository) *OverrideBreachHandler {
	return &OverrideBreachHandler{overrideRepo: overrideRepo}
}

// Handle atomically creates an override record and transitions the breach to
// OVERRIDDEN. The entire write is either fully committed or fully rolled back —
// there is no intermediate state where an override exists without the matching
// breach transition, or vice versa.
//
// Error cases surfaced as typed domain errors:
//   - *domain.ErrBreachNotFound        — no breach with the supplied ID.
//   - *domain.ErrBreachNotOpen         — breach is not in OPEN status.
//   - *domain.ErrOverrideAlreadyExists — a concurrent request already overrode
//                                        this breach (caught by UNIQUE(breach_id)).
func (h *OverrideBreachHandler) Handle(ctx context.Context, req OverrideBreachRequest) (*entity.Override, error) {
	if req.BreachID == uuid.Nil {
		return nil, &domain.ErrInvalidOverrideRequest{Field: "breach_id"}
	}
	if req.Reason == "" {
		return nil, &domain.ErrInvalidOverrideRequest{Field: "reason"}
	}
	if req.OverriddenBy == uuid.Nil {
		return nil, &domain.ErrInvalidOverrideRequest{Field: "overridden_by"}
	}

	// Single UTC timestamp shared between override.CreatedAt and breach.ResolvedAt.
	now := time.Now().UTC()
	override := &entity.Override{
		ID:            uuid.New(),
		BreachID:      req.BreachID,
		Reason:        req.Reason,
		OverriddenBy:  req.OverriddenBy,
		DelegatedFrom: req.DelegatedFrom,
		ApprovedBy:    req.ApprovedBy,
		CreatedAt:     now,
	}

	if err := h.overrideRepo.CommitOverride(ctx, override); err != nil {
		return nil, err
	}
	return override, nil
}
