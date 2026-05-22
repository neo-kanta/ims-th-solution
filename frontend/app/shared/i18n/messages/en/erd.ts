export const enErdMessages = {
  erd: {
    page: {
      title: "Database ERD",
      subtitle:
        "Entity-relationship diagram for the IMS Thailand investment platform.",
      sourceNote:
        "Generated from {count} PostgreSQL migration files in database/migrations.",
      sourceDate: "Snapshot as of {date}",
      modulesCount: "{count} bounded contexts",
      tablesCount: "{count} tables",
      openInDrawio: "Open in draw.io",
      downloadSvg: "Download SVG",
      downloadPng: "Download PNG",
      jumpToModule: "Jump to module",
      searchTable: "Search table",
      searchPlaceholder: "Search by table or column name…",
      noResults: "No tables match this filter.",
      lastUpdated: "Last updated {date}",
    },
    legend: {
      title: "Legend",
      pk: "Primary key",
      fk: "Foreign key",
      uq: "Unique constraint",
      nn: "Not null",
      idx: "Indexed column",
      gin: "JSONB / GIN index",
      check: "Check constraint",
      cascade: "ON DELETE CASCADE",
      restrict: "ON DELETE RESTRICT",
      setNull: "ON DELETE SET NULL",
      dashedLine: "Polymorphic or cross-module reference (no FK)",
      solidLine: "Enforced foreign key",
    },
    concepts: {
      immutable: "Append-only",
      immutableHint:
        "Database triggers block UPDATE and DELETE. Compensate via reversal rows.",
      softDelete: "Soft delete",
      softDeleteHint:
        "Rows are flagged via deleted_at instead of removed. Queries must filter WHERE deleted_at IS NULL.",
      optimisticLock: "Optimistic locking",
      optimisticLockHint:
        "Updates carry a version column; concurrent writes fail loudly.",
      polymorphic: "Polymorphic scope",
      polymorphicHint:
        "scope_type + scope_id pair can target GLOBAL, CONTRACT, FUND_CATEGORY, PORTFOLIO, etc. without a per-target FK.",
      versioned: "Versioned configuration",
      versionedHint:
        "Parameter snapshots stored separately so historical evaluations stay reproducible.",
      crossModule: "Cross-module reference",
      crossModuleHint:
        "UUID-only reference deliberately without an FK to keep modular monolith boundaries clean.",
      bitemporal: "Bi-temporal",
      bitemporalHint:
        "business_date (when the fact applies) is independent of created_at (when it was recorded).",
      eventSourcingLite: "Event-sourcing lite",
      eventSourcingLiteHint:
        "Immutable ledger of transactions projected into balance / position views.",
    },
    modules: {
      iam: {
        name: "IAM",
        short: "Identity & Auth",
        description:
          "User accounts, sessions, MFA enrollment, login attempts and JWT signing key rotation.",
        ownerHint:
          "Foundation module — referenced by created_by / updated_by across the platform.",
      },
      permissions: {
        name: "Permissions",
        short: "RBAC",
        description:
          "Permission groups, group membership, function rights and contract-scoped data rights backed by a code-first catalog.",
        ownerHint:
          "The Go seeder upserts permissions_function_definitions before SQL seeds run.",
      },
      audit: {
        name: "Audit",
        short: "Audit Trail",
        description:
          "Immutable audit event log shared across IAM, compliance and investment writes.",
        ownerHint: "Strictly append-only — triggers reject UPDATE and DELETE.",
      },
      compliance: {
        name: "Compliance",
        short: "Rules & Breaches",
        description:
          "IRG rule registry, parameter versioning, rule bindings to scopes, immutable check records and breach overrides.",
        ownerHint:
          "Versioning lets historical checks be re-evaluated against the exact parameters that fired.",
      },
      workflow: {
        name: "Workflow",
        short: "Business Day",
        description:
          "Per-contract business-day state machine, transition log, approvals, scheduler rules and control decisions.",
        ownerHint:
          "States flow NOT_STARTED → DAY_OPEN → MANAGER_APPROVED → TRANSACTION_CLOSED → ACCOUNTING_CLOSED.",
      },
      investment: {
        name: "Investment",
        short: "Funds & Ledger",
        description:
          "Reference data, funds, portfolios, instruments, append-only ledger, position / cash projections, valuation, NAV, AUM, process assignments and research reports.",
        ownerHint:
          "investment__funds.id IS the cross-module contract_id used by workflow and compliance.",
      },
      marketData: {
        name: "Market Data",
        short: "Provider Feeds",
        description:
          "Canonical market symbols, latest quote / daily snapshots and provider request operational log.",
        ownerHint:
          "Provider-agnostic — Alpha Vantage and Yahoo are wired in today.",
      },
    },
    critical: {
      title: "Critical patterns",
      contractIdTitle: "investment__funds.id is the contract_id",
      contractIdBody:
        "Workflow, compliance and approvals all reference contract_id as a plain UUID without an FK. This is intentional: modules cannot import each other's internal packages, so the join happens at the application layer.",
      immutableTitle: "Append-only tables",
      immutableBody:
        "Triggers reject UPDATE / DELETE on iam_audit_events, compliance_check_records, compliance_overrides, all investment ledger and snapshot tables, and rule_instance_versions. Reverse a posted transaction via reverses_transaction_id.",
      softDeleteTitle: "Soft-delete tables",
      softDeleteBody:
        "iam_users, investment__funds, portfolios, instruments and research_reports use deleted_at. Uniqueness indexes are filtered WHERE deleted_at IS NULL — queries must filter explicitly to avoid phantom duplicates.",
      jsonbTitle: "JSONB extensibility",
      jsonbBody:
        "Asset-class-specific fields live in investment__instruments.attributes (GIN-indexed). Rule parameters live in rule_instance_versions.parameters and check_records.parameter_snapshot.",
      versionedTitle: "Versioned configuration",
      versionedBody:
        "compliance_rule_instances stores only the current pointer; each parameter change inserts a new rule_instance_versions row. Check records pin rule_instance_version so re-running an audit is deterministic.",
      polymorphicTitle: "Polymorphic scope keys",
      polymorphicBody:
        "scope_type + scope_id appears on compliance_rule_bindings, workflow__day_settings, workflow__schedule_rules and investment__process_step_assignments. Application code validates the target type.",
      optimisticTitle: "Optimistic concurrency",
      optimisticBody:
        "iam_users, investment__funds, portfolios, portfolio_positions, cash_balances and workflow__day_states carry a version column that updates must increment-and-check.",
      selfRefTitle: "Self-references",
      selfRefBody:
        "investment__sectors.parent_id models a 4-level GICS tree. portfolio_transactions.reverses_transaction_id is UNIQUE — one-shot reversal only.",
      tokenFamilyTitle: "Token-family rotation",
      tokenFamilyBody:
        "iam_sessions.token_family enables refresh-token replay detection: once a stolen token is detected, the entire family is invalidated. Distinct from the 7-day absolute_expires_at hard cap.",
      catalogTitle: "Code-first permission catalog",
      catalogBody:
        "permissions_function_definitions is upserted by the Go seeder before SQL seeds run. function_rights FKs to it with ON UPDATE CASCADE so renaming a code is safe.",
    },
    stats: {
      title: "By the numbers",
      module: "Module",
      tables: "Tables",
      notes: "Notes",
      total: "Total",
      immutable: "Immutable tables",
      softDelete: "Soft-delete tables",
      optimisticLock: "Optimistic-locked tables",
    },
    tables: {
      iam_users: "Users",
      iam_sessions: "Refresh sessions",
      iam_audit_events: "Audit events",
      iam_mfa_enrollments: "MFA enrollments",
      iam_mfa_recovery_codes: "MFA recovery codes",
      iam_login_attempts: "Login attempts",
      iam_signing_keys: "JWT signing keys",
      permissions_groups: "Permission groups",
      permissions_accounts_groups: "User-group memberships",
      permissions_function_rights: "Function rights (per group)",
      permissions_data_rights: "Data rights (per user)",
      permissions_function_definitions: "Function permission catalog",
      compliance_rule_instances: "Rule instances",
      compliance_rule_instance_versions: "Rule parameter versions",
      compliance_rule_sets: "Rule sets",
      compliance_rule_set_members: "Rule-set memberships",
      compliance_rule_bindings: "Rule bindings (scope)",
      compliance_check_records: "Rule check records",
      compliance_breaches: "Breaches",
      compliance_overrides: "Breach overrides",
      compliance_restriction_list_entries: "Restriction list entries",
      workflow__day_states: "Business-day state",
      workflow__transition_log: "State transition log",
      workflow__approval_records: "Approval records",
      workflow__day_settings: "Day-window settings",
      workflow__scheduler_contracts: "Scheduler contract bridge",
      workflow__schedule_rules: "Schedule rules",
      workflow__scheduler_runs: "Scheduler runs",
      workflow__scheduler_run_items: "Scheduler run items",
      workflow__control_decisions: "Control decisions",
      investment__asset_classes: "Asset classes",
      investment__asset_subtypes: "Asset subtypes",
      investment__regions: "Regions",
      investment__countries: "Countries",
      investment__sectors: "Sectors (GICS tree)",
      investment__fund_categories: "Fund categories",
      investment__investment_styles: "Investment styles",
      investment__funds: "Funds (★ contract_id)",
      investment__portfolios: "Portfolios",
      investment__instruments: "Instruments",
      investment__instrument_identifiers: "Instrument identifiers",
      investment__portfolio_transactions: "Portfolio transactions (ledger)",
      investment__portfolio_positions: "Portfolio positions",
      investment__cash_movements: "Cash movements (ledger)",
      investment__cash_balances: "Cash balances",
      investment__price_snapshots: "Price snapshots",
      investment__valuation_snapshots: "Valuation snapshots",
      investment__valuation_holding_lines: "Valuation holding lines",
      investment__nav_snapshots: "NAV snapshots",
      investment__aum_snapshots: "AUM snapshots",
      investment__process_steps: "Investment process steps",
      investment__process_groups: "Investment process groups",
      investment__process_group_members: "Process group members",
      investment__process_step_assignments: "Process step assignments",
      investment__research_reports: "Research reports",
      market_symbols: "Market symbols",
      market_data_snapshots: "Market data snapshots",
      provider_requests_log: "Provider request log",
    },
    actions: {
      copySchema: "Copy table schema",
      copySchemaSuccess: "Schema copied to clipboard.",
      viewMigration: "View migration file",
      viewRelated: "View related tables",
      expandAll: "Expand all",
      collapseAll: "Collapse all",
      filterByModule: "Filter by module",
      clearFilters: "Clear filters",
    },
    statuses: {
      immutable: "Immutable",
      softDelete: "Soft-delete",
      optimisticLock: "Versioned",
      indexed: "Indexed",
      jsonb: "JSONB",
      reference: "Reference data",
      ledger: "Ledger",
      projection: "Projection",
      snapshot: "Snapshot",
      configuration: "Configuration",
    },
  },
} as const;
