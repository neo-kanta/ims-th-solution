# Module: audit

This module is the active backend boundary for audit and change-log functionality.

## Intended Scope

- audit event querying
- retention and export workflows
- filterable audit reporting
- immutable change-log access patterns

## Current Status

This module is wired into the running backend.

It owns:

- immutable audit event definitions
- audit event persistence
- audit recorder implementation used by other modules
- automatic request-ID enrichment for recorded events
- audit list and export query flows
- self-audited audit access and export flows
- admin audit HTTP handlers and route registration

## Boundary Notes

- IAM and other modules should emit audit events through the recorder port exposed by this module
- authentication, session, and account lifecycle logic remain outside this module
- the current persistence layer still writes to the existing `iam_audit_events` table for backward compatibility
