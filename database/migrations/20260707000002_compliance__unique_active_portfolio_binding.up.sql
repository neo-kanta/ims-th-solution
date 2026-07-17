-- =============================================================================
-- Compliance / IRG - Portfolio Compliance V2 (additive)
-- =============================================================================
-- Prevent binding the same rule instance to the same portfolio twice while
-- both bindings are active. Application-layer duplicate checks (see
-- application/command/create_rule_binding.go) are the primary defense; this
-- partial unique index is the database-level safety net against the same
-- concurrent-request race the compliance_overrides uniqueness migration
-- (20260417000002) guards against.
--
-- Scoped to PORTFOLIO bindings only, matching the current binding-CRUD use
-- case (task: prevent duplicate active binding for same rule_instance_id +
-- PORTFOLIO + portfolio_id). GLOBAL/CONTRACT/other scopes are unaffected.

CREATE UNIQUE INDEX uq_compliance_rb_active_portfolio
    ON compliance_rule_bindings (rule_instance_id, scope_id)
    WHERE is_active = true AND scope_type = 'PORTFOLIO';
