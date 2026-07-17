package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
)

// DeactivateRuleBindingRequest deactivates an existing binding. Ownership
// (which portfolio/scope the caller is allowed to touch) is enforced by the
// caller before invoking this handler — see the portfolio contract adapter,
// which verifies the binding's scope matches the resolved portfolio.
type DeactivateRuleBindingRequest struct {
	BindingID uuid.UUID
}

// DeactivateRuleBindingHandler sets a binding's is_active flag to false.
type DeactivateRuleBindingHandler struct {
	bindingRepo domain.RuleBindingRepository
}

// NewDeactivateRuleBindingHandler wires the handler.
func NewDeactivateRuleBindingHandler(bindingRepo domain.RuleBindingRepository) *DeactivateRuleBindingHandler {
	return &DeactivateRuleBindingHandler{bindingRepo: bindingRepo}
}

// Handle deactivates the binding. Returns *domain.ErrBindingNotFound if it
// does not exist.
func (h *DeactivateRuleBindingHandler) Handle(ctx context.Context, req DeactivateRuleBindingRequest) error {
	if h == nil || h.bindingRepo == nil {
		return fmt.Errorf("deactivate rule binding handler not initialised")
	}
	if req.BindingID == uuid.Nil {
		return fmt.Errorf("binding_id is required")
	}
	existing, err := h.bindingRepo.GetByID(ctx, req.BindingID)
	if err != nil {
		return fmt.Errorf("loading rule binding: %w", err)
	}
	if existing == nil {
		return &domain.ErrBindingNotFound{ID: req.BindingID.String()}
	}
	if err := h.bindingRepo.Deactivate(ctx, req.BindingID); err != nil {
		return fmt.Errorf("deactivating rule binding: %w", err)
	}
	return nil
}
