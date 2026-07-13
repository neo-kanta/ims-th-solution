package domain

import "fmt"

// ErrRuleTypeNotFound indicates a rule type ID is not registered in the SPI registry.
type ErrRuleTypeNotFound struct {
	TypeID string
}

func (e *ErrRuleTypeNotFound) Error() string {
	return fmt.Sprintf("compliance: rule type not registered: %s", e.TypeID)
}

// ErrRuleInstanceNotFound indicates a rule instance was not found.
type ErrRuleInstanceNotFound struct {
	ID string
}

func (e *ErrRuleInstanceNotFound) Error() string {
	return fmt.Sprintf("compliance: rule instance not found: %s", e.ID)
}

// ErrParameterValidation indicates rule parameters failed schema validation.
type ErrParameterValidation struct {
	RuleTypeID string
	Detail     string
}

func (e *ErrParameterValidation) Error() string {
	return fmt.Sprintf("compliance: parameter validation failed for %s: %s", e.RuleTypeID, e.Detail)
}

// ErrBreachNotOverridable indicates a hard-regulatory rule cannot be overridden.
type ErrBreachNotOverridable struct {
	BreachID   string
	RuleTypeID string
}

func (e *ErrBreachNotOverridable) Error() string {
	return fmt.Sprintf("compliance: breach %s (rule %s) is not overridable", e.BreachID, e.RuleTypeID)
}

// ErrStaleData indicates required data is missing or stale. Fail-closed.
type ErrStaleData struct {
	DataSource string
	Detail     string
}

func (e *ErrStaleData) Error() string {
	return fmt.Sprintf("compliance: stale or missing data from %s: %s", e.DataSource, e.Detail)
}

// ErrBindingNotFound indicates a binding was not found.
type ErrBindingNotFound struct {
	ID string
}

func (e *ErrBindingNotFound) Error() string {
	return fmt.Sprintf("compliance: binding not found: %s", e.ID)
}

// ErrBreachNotFound indicates a breach row was not found when locking for override.
type ErrBreachNotFound struct {
	BreachID string
}

func (e *ErrBreachNotFound) Error() string {
	return fmt.Sprintf("compliance: breach not found: %s", e.BreachID)
}

// ErrBreachNotOpen indicates the target breach is not in OPEN status and cannot be overridden.
// The breach may already be OVERRIDDEN (duplicate request) or RESOLVED (state changed).
type ErrBreachNotOpen struct {
	BreachID      string
	CurrentStatus string
}

func (e *ErrBreachNotOpen) Error() string {
	return fmt.Sprintf("compliance: breach %s is %s, only OPEN breaches can be overridden",
		e.BreachID, e.CurrentStatus)
}

// ErrOverrideAlreadyExists indicates another override for the same breach has already
// been committed — either by a concurrent request that won the race, or by a prior
// override attempt. Surfaced from the UNIQUE(breach_id) constraint.
type ErrOverrideAlreadyExists struct {
	BreachID string
}

func (e *ErrOverrideAlreadyExists) Error() string {
	return fmt.Sprintf("compliance: override already exists for breach %s", e.BreachID)
}

// ErrRuleInstanceInactive indicates a rule instance exists but is not active
// and therefore cannot be bound to a new scope.
type ErrRuleInstanceInactive struct {
	ID string
}

func (e *ErrRuleInstanceInactive) Error() string {
	return fmt.Sprintf("compliance: rule instance %s is not active", e.ID)
}

// ErrDuplicateActiveBinding indicates an active binding already exists for
// the same rule instance + scope. Surfaced either by the application-layer
// pre-check or the DB unique index (compliance_rule_bindings, PORTFOLIO scope
// only — migration 20260707000002).
type ErrDuplicateActiveBinding struct {
	RuleInstanceID string
	ScopeType      string
	ScopeID        string
}

func (e *ErrDuplicateActiveBinding) Error() string {
	return fmt.Sprintf(
		"compliance: an active binding already exists for rule instance %s at scope %s/%s",
		e.RuleInstanceID, e.ScopeType, e.ScopeID,
	)
}

// ErrInvalidOverrideRequest indicates the override request failed input
// validation (missing/empty required field). Client-fixable → HTTP 400.
// Distinct from internal errors (DB failures, tx errors) which must surface as 500.
type ErrInvalidOverrideRequest struct {
	Field  string
	Detail string
}

func (e *ErrInvalidOverrideRequest) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("compliance: invalid override request: %s: %s", e.Field, e.Detail)
	}
	return fmt.Sprintf("compliance: invalid override request: %s is required", e.Field)
}

// ErrInvalidPreTradeRequest indicates a pre-trade check request failed input
// validation (missing/invalid field on the proposed order). Client-fixable →
// HTTP 400. Distinct from internal errors (DB failures, pipeline errors)
// which must surface as 500.
type ErrInvalidPreTradeRequest struct {
	Detail string
}

func (e *ErrInvalidPreTradeRequest) Error() string {
	return fmt.Sprintf("compliance: invalid pre-trade request: %s", e.Detail)
}
