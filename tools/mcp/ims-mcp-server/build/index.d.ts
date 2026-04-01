#!/usr/bin/env node
/**
 * IMS Thailand — MCP Server
 *
 * Exposes the IMS project's PostgreSQL database, workflow states,
 * domain rules, permissions model, and project context as tools
 * for Claude Code / Claude.ai during development.
 *
 * Transport: stdio (for Claude Code local use)
 *
 * Tools provided:
 *   Database:
 *     - db_query           → Run read-only SQL against the IMS database
 *     - db_schema          → List all tables, columns, types, constraints
 *     - db_table_info      → Detailed info for a specific table
 *     - db_migrations      → Show applied migration history
 *
 *   Workflow:
 *     - workflow_states     → Show current workflow day states
 *     - workflow_validate   → Check if a state transition is valid
 *
 *   Permissions:
 *     - perm_user_rights    → Show all permissions for a user
 *     - perm_check          → Check if user has a specific permission
 *     - perm_groups         → List all permission groups and members
 *
 *   Domain:
 *     - domain_rules        → List business validation rules for a module
 *     - audit_log           → Query recent audit trail entries
 *
 *   Project:
 *     - project_modules     → List all backend modules and their status
 *     - project_api_routes  → List registered API routes
 *     - project_migrations  → List migration files (not DB state — file system)
 */
export {};
//# sourceMappingURL=index.d.ts.map