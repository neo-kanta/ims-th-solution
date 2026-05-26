import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, "..", "..");
const migrationDir = path.join(repoRoot, "database", "migrations");
const outDir = path.join(repoRoot, "database", "erd");
const drawioPath = path.join(outDir, "ims-th-solution-erd.drawio");
const notesPath = path.join(outDir, "ims-th-solution-erd-notes.md");
const GENERATED_DATE =
  process.env.IMS_ERD_GENERATED_DATE ||
  new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Bangkok",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());

const COLORS = {
  iam: { main: "#4F46E5", fill: "#EEF2FF", stroke: "#3730A3" },
  permissions: { main: "#7C3AED", fill: "#F5F3FF", stroke: "#5B21B6" },
  permissionWorkflow: { main: "#BE185D", fill: "#FDF2F8", stroke: "#9D174D" },
  workflow: { main: "#D97706", fill: "#FFF7ED", stroke: "#B45309" },
  compliance: { main: "#DC2626", fill: "#FEF2F2", stroke: "#991B1B" },
  investment: { main: "#059669", fill: "#ECFDF5", stroke: "#047857" },
  market: { main: "#0891B2", fill: "#ECFEFF", stroke: "#0E7490" },
  neutral: { main: "#475569", fill: "#F8FAFC", stroke: "#334155" },
  warning: { main: "#F59E0B", fill: "#FFFBEB", stroke: "#B45309" },
};

const PAGE = {
  width: 2600,
  marginX: 60,
  marginY: 42,
  headerH: 118,
  gapX: 34,
  gapY: 34,
  cardW: 420,
};

const STRUCTURAL_USER_COLUMNS = new Set([
  "user_id",
  "manager_user_id",
  "owner_user_id",
  "author_user_id",
]);

const AUDIT_USER_COLUMNS = new Set([
  "created_by",
  "updated_by",
  "assigned_by",
  "assigned_to",
  "granted_by",
  "added_by",
  "resolved_by",
  "overridden_by",
  "delegated_from",
  "delegated_from_user_id",
  "approved_by",
  "approver_user_id",
  "actor_user_id",
  "changed_by",
  "closed_by",
  "merged_by",
  "rejected_by",
  "requested_by",
  "decided_by",
  "superseded_by",
]);

const MANUAL_NOTES = [
  {
    title: "Fund.id is the contract_id",
    body:
      "investment__funds.id is the cross-module contract key used by workflow, compliance, scheduler, and permissions data rights. Most of those links are intentionally not declared as FKs to keep module boundaries loose.",
  },
  {
    title: "Two permission namespaces now coexist",
    body:
      "The legacy permissions_* RBAC/grant tables remain active while the newer permission_* and approval_workflow_* tables add request, approval, merge, labeling, checks, and notification workflow. Be explicit about which path is authoritative during migration.",
  },
  {
    title: "Soft delete is not uniform",
    body:
      "funds, portfolios, instruments, and research reports use partial unique indexes with deleted_at IS NULL. iam_users and permissions_groups still have full unique constraints, so soft-deleted usernames/group names cannot be reused without a migration.",
  },
  {
    title: "Append-only enforcement differs by module",
    body:
      "investment ledger/snapshot tables use rejecting triggers. iam_audit_events uses rejecting triggers. compliance_check_records and compliance_overrides rely on privilege revokes, which do not stop the table owner. workflow transition_log is append-only by design comments, but not enforced by a DB trigger today.",
  },
  {
    title: "Polymorphic scope and subject columns need application validation",
    body:
      "scope_type/scope_id appears in compliance bindings, workflow settings/rules, investment process assignments, and AUM snapshots. The new permission grants also use subject_type/subject_id for USER/GROUP/ROLE targets. The database cannot enforce all of these conditional references directly.",
  },
  {
    title: "Research reports are deliberately loose in PoC scope",
    body:
      "owner_user_id, author_user_id, applicable_contract_id, and instrument_code have no FKs yet. This keeps the feature scaffold flexible, but it allows orphaned user/contract/instrument references.",
  },
  {
    title: "Permissions data rights has a type mismatch",
    body:
      "permissions_data_rights.contract_id is VARCHAR(50), while the investment contract key is investment__funds.id UUID. This is likely legacy or an integration boundary, and should be revisited before hard FK enforcement.",
  },
  {
    title: "Market data is symbol-driven",
    body:
      "market_symbols and market_data_snapshots are not linked to investment__instruments. Mapping is by ticker/provider symbols, so reconciliation logic must handle mismatches and provider aliases.",
  },
  {
    title: "JSONB carries important domain state",
    body:
      "instrument attributes, compliance parameters/evidence/snapshots, workflow metadata, and market raw_payload are JSONB. They are flexible, but they move some schema guarantees from DB constraints into code/tests.",
  },
  {
    title: "Optimistic locking marks mutable projections",
    body:
      "version columns appear on users, funds, portfolios, day_states, portfolio_positions, and cash_balances. These are mutable current-state projections and should update with compare-and-increment semantics.",
  },
  {
    title: "Rule versioning preserves reproducibility",
    body:
      "compliance_rule_instances points at the current version, while check records pin rule_instance_version and parameter_snapshot. Historical compliance decisions should be replayable against the exact parameters used.",
  },
];

const PAGE_GROUPS = [
  {
    name: "01 - IAM + Core Permissions",
    description: "Identity, sessions, MFA, audit trail, and the original permissions_* RBAC/grant tables.",
    modules: ["iam", "permissions"],
    tables: [
      "iam_users",
      "iam_sessions",
      "iam_mfa_enrollments",
      "iam_mfa_recovery_codes",
      "iam_login_attempts",
      "iam_signing_keys",
      "iam_audit_events",
      "permissions_groups",
      "permissions_accounts_groups",
      "permissions_function_definitions",
      "permissions_function_rights",
      "permissions_data_rights",
    ],
  },
  {
    name: "02 - Permission Approval Workflow",
    description: "Financial-grade permission request workflow, approvals, comments, checks, labels, audit logs, and notifications.",
    modules: ["permissionWorkflow"],
    tables: [
      "permission_roles",
      "permission_user_role_assignments",
      "permission_role_assignment_policies",
      "permission_change_requests",
      "permission_change_items",
      "approval_workflow_settings",
      "approval_workflow_steps",
      "permission_request_approval_steps",
      "permission_request_step_approvers",
      "permission_request_comments",
      "permission_request_checks",
      "permission_request_revisions",
      "approval_workflow_events",
      "permission_labels",
      "permission_change_request_labels",
      "permission_function_definitions",
      "permission_function_rights",
      "permission_data_rights",
      "audit_logs",
      "notification_settings",
    ],
  },
  {
    name: "03 - Workflow + Scheduler",
    description: "Business-day state machine, transition audit, scheduler configuration, run audit, and control decisions.",
    modules: ["workflow"],
    tables: [
      "workflow__day_states",
      "workflow__transition_log",
      "workflow__approval_records",
      "workflow__day_settings",
      "workflow__scheduler_contracts",
      "workflow__schedule_rules",
      "workflow__scheduler_runs",
      "workflow__scheduler_run_items",
      "workflow__control_decisions",
    ],
  },
  {
    name: "04 - Compliance",
    description: "Rule definitions, versioned parameters, scoped bindings, immutable check records, breaches, and overrides.",
    modules: ["compliance"],
    tables: [
      "compliance_rule_instances",
      "compliance_rule_instance_versions",
      "compliance_rule_sets",
      "compliance_rule_set_members",
      "compliance_rule_bindings",
      "compliance_check_records",
      "compliance_breaches",
      "compliance_overrides",
      "compliance_restriction_list_entries",
    ],
  },
  {
    name: "05 - Investment Reference + Master",
    description: "Taxonomy, fund/contract master, portfolios, instrument master, external identifiers, and research reports.",
    modules: ["investment"],
    tables: [
      "investment__asset_classes",
      "investment__asset_subtypes",
      "investment__regions",
      "investment__countries",
      "investment__sectors",
      "investment__fund_categories",
      "investment__investment_styles",
      "investment__funds",
      "investment__portfolios",
      "investment__instruments",
      "investment__instrument_identifiers",
      "investment__research_reports",
    ],
  },
  {
    name: "06 - Investment Process + Ledger",
    description: "Process authorization, append-only trade/cash ledgers, mutable projections, pricing, valuation, NAV, and AUM snapshots.",
    modules: ["investment"],
    tables: [
      "investment__process_steps",
      "investment__process_groups",
      "investment__process_group_members",
      "investment__process_step_assignments",
      "investment__portfolio_transactions",
      "investment__portfolio_positions",
      "investment__cash_movements",
      "investment__cash_balances",
      "investment__price_snapshots",
      "investment__valuation_snapshots",
      "investment__valuation_holding_lines",
      "investment__nav_snapshots",
      "investment__aum_snapshots",
    ],
  },
  {
    name: "07 - Market Data",
    description: "Provider symbol mapping, price/quote snapshots, and provider request observability.",
    modules: ["market"],
    tables: ["market_symbols", "market_data_snapshots", "provider_requests_log"],
  },
];

const INFERRED_RELATIONSHIPS = [
  {
    fromTable: "workflow__day_states",
    fromColumn: "contract_id",
    toTable: "investment__funds",
    toColumn: "id",
    note: "contract_id = fund.id, no FK",
  },
  {
    fromTable: "workflow__scheduler_contracts",
    fromColumn: "contract_id",
    toTable: "investment__funds",
    toColumn: "id",
    note: "scheduler bridge contract, no FK",
  },
  {
    fromTable: "workflow__scheduler_contracts",
    fromColumn: "fund_id",
    toTable: "investment__funds",
    toColumn: "id",
    note: "bridge fund_id, no FK",
  },
  {
    fromTable: "workflow__transition_log",
    fromColumn: "contract_id",
    toTable: "investment__funds",
    toColumn: "id",
    note: "denormalized contract, no FK",
  },
  {
    fromTable: "workflow__approval_records",
    fromColumn: "contract_id",
    toTable: "investment__funds",
    toColumn: "id",
    note: "denormalized contract, no FK",
  },
  {
    fromTable: "workflow__control_decisions",
    fromColumn: "contract_id",
    toTable: "investment__funds",
    toColumn: "id",
    note: "decision scope, no FK",
  },
  {
    fromTable: "workflow__scheduler_run_items",
    fromColumn: "contract_id",
    toTable: "investment__funds",
    toColumn: "id",
    note: "scheduler target, no FK",
  },
  {
    fromTable: "compliance_check_records",
    fromColumn: "contract_id",
    toTable: "investment__funds",
    toColumn: "id",
    note: "compliance contract, no FK",
  },
  {
    fromTable: "compliance_breaches",
    fromColumn: "contract_id",
    toTable: "investment__funds",
    toColumn: "id",
    note: "breach contract, no FK",
  },
  {
    fromTable: "compliance_check_records",
    fromColumn: "portfolio_id",
    toTable: "investment__portfolios",
    toColumn: "id",
    note: "compliance portfolio, no FK",
  },
  {
    fromTable: "compliance_breaches",
    fromColumn: "portfolio_id",
    toTable: "investment__portfolios",
    toColumn: "id",
    note: "breach portfolio, no FK",
  },
  {
    fromTable: "permissions_data_rights",
    fromColumn: "contract_id",
    toTable: "investment__funds",
    toColumn: "id",
    note: "VARCHAR contract grant, fund.id is UUID",
  },
  {
    fromTable: "permission_function_rights",
    fromColumn: "subject_id",
    toTable: "iam_users",
    toColumn: "id",
    note: "when subject_type = USER",
  },
  {
    fromTable: "permission_function_rights",
    fromColumn: "subject_id",
    toTable: "permissions_groups",
    toColumn: "id",
    note: "when subject_type = GROUP",
  },
  {
    fromTable: "permission_function_rights",
    fromColumn: "subject_id",
    toTable: "permission_roles",
    toColumn: "id",
    note: "when subject_type = ROLE",
  },
  {
    fromTable: "permission_data_rights",
    fromColumn: "subject_id",
    toTable: "iam_users",
    toColumn: "id",
    note: "when subject_type = USER",
  },
  {
    fromTable: "permission_data_rights",
    fromColumn: "subject_id",
    toTable: "permissions_groups",
    toColumn: "id",
    note: "when subject_type = GROUP",
  },
  {
    fromTable: "permission_data_rights",
    fromColumn: "subject_id",
    toTable: "permission_roles",
    toColumn: "id",
    note: "when subject_type = ROLE",
  },
  {
    fromTable: "approval_workflow_steps",
    fromColumn: "required_role_code",
    toTable: "permission_roles",
    toColumn: "role_code",
    note: "role lookup by code, no FK",
  },
  {
    fromTable: "permission_request_step_approvers",
    fromColumn: "approver_role_code",
    toTable: "permission_roles",
    toColumn: "role_code",
    note: "role lookup by code, no FK",
  },
  {
    fromTable: "investment__aum_snapshots",
    fromColumn: "scope_id",
    toTable: "investment__funds",
    toColumn: "id",
    note: "when scope_type = FUND",
  },
  {
    fromTable: "investment__aum_snapshots",
    fromColumn: "scope_id",
    toTable: "investment__portfolios",
    toColumn: "id",
    note: "when scope_type = PORTFOLIO",
  },
  {
    fromTable: "investment__research_reports",
    fromColumn: "owner_user_id",
    toTable: "iam_users",
    toColumn: "id",
    note: "plain UUID in PoC",
  },
  {
    fromTable: "investment__research_reports",
    fromColumn: "author_user_id",
    toTable: "iam_users",
    toColumn: "id",
    note: "plain UUID in PoC",
  },
  {
    fromTable: "investment__research_reports",
    fromColumn: "applicable_contract_id",
    toTable: "investment__funds",
    toColumn: "id",
    note: "plain UUID in PoC",
  },
  {
    fromTable: "investment__research_reports",
    fromColumn: "instrument_code",
    toTable: "investment__instruments",
    toColumn: "primary_ticker",
    note: "plain string in PoC",
  },
  {
    fromTable: "market_symbols",
    fromColumn: "symbol",
    toTable: "investment__instruments",
    toColumn: "primary_ticker",
    note: "symbol mapping, no FK",
  },
];

function stripSqlComments(sql) {
  return sql.replace(/\/\*[\s\S]*?\*\//g, "").replace(/--.*$/gm, "");
}

function splitSqlStatements(sql) {
  const out = [];
  let buf = "";
  let inSingle = false;
  let inDouble = false;
  let inDollar = false;

  for (let i = 0; i < sql.length; i += 1) {
    const ch = sql[i];
    const next = sql[i + 1];

    if (!inSingle && !inDouble && ch === "$" && next === "$") {
      inDollar = !inDollar;
      buf += "$$";
      i += 1;
      continue;
    }
    if (!inDollar && !inDouble && ch === "'" && next === "'") {
      buf += "''";
      i += 1;
      continue;
    }
    if (!inDollar && !inDouble && ch === "'") {
      inSingle = !inSingle;
    } else if (!inDollar && !inSingle && ch === '"') {
      inDouble = !inDouble;
    }

    if (!inSingle && !inDouble && !inDollar && ch === ";") {
      const stmt = buf.trim();
      if (stmt) out.push(stmt);
      buf = "";
    } else {
      buf += ch;
    }
  }

  const trailing = buf.trim();
  if (trailing) out.push(trailing);
  return out;
}

function splitTopLevel(input) {
  const parts = [];
  let buf = "";
  let paren = 0;
  let bracket = 0;
  let inSingle = false;
  let inDouble = false;

  for (let i = 0; i < input.length; i += 1) {
    const ch = input[i];
    const next = input[i + 1];

    if (!inDouble && ch === "'" && next === "'") {
      buf += "''";
      i += 1;
      continue;
    }
    if (!inDouble && ch === "'") {
      inSingle = !inSingle;
    } else if (!inSingle && ch === '"') {
      inDouble = !inDouble;
    } else if (!inSingle && !inDouble && ch === "(") {
      paren += 1;
    } else if (!inSingle && !inDouble && ch === ")") {
      paren -= 1;
    } else if (!inSingle && !inDouble && ch === "[") {
      bracket += 1;
    } else if (!inSingle && !inDouble && ch === "]") {
      bracket -= 1;
    }

    if (!inSingle && !inDouble && paren === 0 && bracket === 0 && ch === ",") {
      const part = buf.trim();
      if (part) parts.push(part);
      buf = "";
    } else {
      buf += ch;
    }
  }

  const final = buf.trim();
  if (final) parts.push(final);
  return parts;
}

function findTopLevelKeyword(input, keywords) {
  let paren = 0;
  let bracket = 0;
  let inSingle = false;
  let inDouble = false;
  const upper = input.toUpperCase();
  const normalizedKeywords = keywords.map((kw) => kw.toUpperCase());

  for (let i = 0; i < input.length; i += 1) {
    const ch = input[i];
    const next = input[i + 1];

    if (!inDouble && ch === "'" && next === "'") {
      i += 1;
      continue;
    }
    if (!inDouble && ch === "'") {
      inSingle = !inSingle;
      continue;
    }
    if (!inSingle && ch === '"') {
      inDouble = !inDouble;
      continue;
    }
    if (!inSingle && !inDouble && ch === "(") paren += 1;
    if (!inSingle && !inDouble && ch === ")") paren -= 1;
    if (!inSingle && !inDouble && ch === "[") bracket += 1;
    if (!inSingle && !inDouble && ch === "]") bracket -= 1;

    if (inSingle || inDouble || paren !== 0 || bracket !== 0) continue;

    for (const kw of normalizedKeywords) {
      if (upper.startsWith(kw, i) && isBoundary(upper[i - 1]) && isBoundary(upper[i + kw.length])) {
        return i;
      }
    }
  }
  return -1;
}

function isBoundary(ch) {
  return !ch || /[^A-Z0-9_]/.test(ch);
}

function normalizeType(type) {
  return type
    .replace(/\s+/g, " ")
    .replace(/\bBOOLEAN\b/gi, "bool")
    .replace(/\bINTEGER\b/gi, "int")
    .replace(/\bINT\b/gi, "int")
    .replace(/\bTIMESTAMP WITH TIME ZONE\b/gi, "timestamptz")
    .replace(/\bTIMESTAMPTZ\b/gi, "timestamptz")
    .replace(/\bVARCHAR\b/gi, "varchar")
    .replace(/\bDECIMAL\b/gi, "decimal")
    .replace(/\bUUID\b/gi, "uuid")
    .replace(/\bJSONB\b/gi, "jsonb")
    .replace(/\bTEXT\b/gi, "text")
    .replace(/\bDATE\b/gi, "date")
    .replace(/\bTIME\b/gi, "time")
    .replace(/\bSMALLINT\b/gi, "smallint")
    .replace(/\bBIGINT\b/gi, "bigint")
    .replace(/\bINET\b/gi, "inet")
    .replace(/\bCHAR\b/gi, "char")
    .trim();
}

function parseColumnDef(def) {
  const match = def.trim().match(/^"?([A-Za-z_][\w]*)"?\s+([\s\S]+)$/);
  if (!match) return null;

  const name = match[1];
  const rest = match[2].trim();
  const typeEnd = findTopLevelKeyword(rest, [
    "PRIMARY KEY",
    "NOT NULL",
    "NULL",
    "DEFAULT",
    "REFERENCES",
    "CHECK",
    "CONSTRAINT",
    "UNIQUE",
    "COLLATE",
  ]);
  const rawType = typeEnd === -1 ? rest : rest.slice(0, typeEnd).trim();
  const column = {
    name,
    type: normalizeType(rawType),
    pk: /\bPRIMARY\s+KEY\b/i.test(rest),
    notNull: /\bNOT\s+NULL\b/i.test(rest),
    unique: /\bUNIQUE\b/i.test(rest),
    fk: null,
  };

  const ref = rest.match(/\bREFERENCES\s+([A-Za-z_][\w]*)\s*\(\s*([A-Za-z_][\w]*)\s*\)/i);
  if (ref) {
    column.fk = {
      toTable: ref[1],
      toColumn: ref[2],
      onDelete: (rest.match(/\bON\s+DELETE\s+([A-Z ]+?)(?=\s+ON|\s*$)/i)?.[1] || "").trim(),
      onUpdate: (rest.match(/\bON\s+UPDATE\s+([A-Z ]+?)(?=\s+ON|\s*$)/i)?.[1] || "").trim(),
    };
  }

  return column;
}

function parseColumnList(raw) {
  return splitTopLevel(raw)
    .map((item) => item.replace(/"/g, "").trim())
    .filter(Boolean);
}

function ensureTable(schema, name) {
  if (!schema.tables.has(name)) {
    schema.tables.set(name, {
      name,
      columns: [],
      columnMap: new Map(),
      uniqueConstraints: [],
      uniqueIndexes: [],
      indexes: [],
      triggers: new Set(),
      rules: new Set(),
      revokes: new Set(),
      comments: [],
      sourceMigrations: new Set(),
    });
  }
  return schema.tables.get(name);
}

function addColumn(table, column) {
  if (!column || table.columnMap.has(column.name)) return;
  table.columns.push(column);
  table.columnMap.set(column.name, column);
}

function addFk(schema, fromTable, fromColumn, fk, source) {
  const key = `${fromTable}.${fromColumn}->${fk.toTable}.${fk.toColumn}`;
  if (schema.fkKeys.has(key)) return;
  schema.fkKeys.add(key);
  schema.fks.push({
    fromTable,
    fromColumn,
    toTable: fk.toTable,
    toColumn: fk.toColumn,
    onDelete: fk.onDelete || "",
    onUpdate: fk.onUpdate || "",
    source,
  });
}

function markUnique(table, columns, name, kind = "constraint", where = "") {
  const record = { name, columns, kind, where };
  if (kind === "index") table.uniqueIndexes.push(record);
  else table.uniqueConstraints.push(record);

  if (columns.length === 1 && table.columnMap.has(columns[0])) {
    table.columnMap.get(columns[0]).unique = true;
  }
}

function parseSchema() {
  const schema = { tables: new Map(), fks: [], fkKeys: new Set(), migrations: [] };
  const migrationFiles = fs
    .readdirSync(migrationDir)
    .filter((file) => file.endsWith(".up.sql"))
    .sort();

  for (const file of migrationFiles) {
    const fullPath = path.join(migrationDir, file);
    const rawSql = fs.readFileSync(fullPath, "utf8");
    const sql = stripSqlComments(rawSql);
    const statements = splitSqlStatements(sql);
    schema.migrations.push(file);

    for (const stmt of statements) {
      parseCreateTable(schema, stmt, file);
      parseAlterTable(schema, stmt, file);
      parseCreateIndex(schema, stmt);
      parseTrigger(schema, stmt);
      parseRule(schema, stmt);
      parseRevoke(schema, stmt);
    }
  }

  for (const table of schema.tables.values()) {
    for (const col of table.columns) {
      if (col.fk) addFk(schema, table.name, col.name, col.fk, "column");
    }
  }

  return schema;
}

function parseCreateTable(schema, stmt, file) {
  const create = stmt.match(/^CREATE\s+TABLE(?:\s+IF\s+NOT\s+EXISTS)?\s+([A-Za-z_][\w]*)\s*\(/i);
  if (!create) return;

  const table = ensureTable(schema, create[1]);
  table.sourceMigrations.add(file);

  const open = stmt.indexOf("(", create[0].length - 1);
  const close = stmt.lastIndexOf(")");
  if (open === -1 || close === -1 || close <= open) return;

  const body = stmt.slice(open + 1, close);
  for (const part of splitTopLevel(body)) {
    const trimmed = part.trim();
    const upper = trimmed.toUpperCase();

    if (upper.startsWith("CONSTRAINT") || upper.startsWith("UNIQUE") || upper.startsWith("PRIMARY KEY") || upper.startsWith("FOREIGN KEY")) {
      parseTableConstraint(schema, table, trimmed);
      continue;
    }

    addColumn(table, parseColumnDef(trimmed));
  }
}

function parseTableConstraint(schema, table, constraint) {
  const name = constraint.match(/^CONSTRAINT\s+([A-Za-z_][\w]*)/i)?.[1] || "";

  const unique = constraint.match(/\bUNIQUE\s*\(([\s\S]+?)\)/i);
  if (unique) {
    markUnique(table, parseColumnList(unique[1]), name || "unique", "constraint");
  }

  const pk = constraint.match(/\bPRIMARY\s+KEY\s*\(([\s\S]+?)\)/i);
  if (pk) {
    for (const colName of parseColumnList(pk[1])) {
      const col = table.columnMap.get(colName);
      if (col) col.pk = true;
    }
  }

  const fk = constraint.match(/\bFOREIGN\s+KEY\s*\(([\s\S]+?)\)\s+REFERENCES\s+([A-Za-z_][\w]*)\s*\(\s*([A-Za-z_][\w]*)\s*\)/i);
  if (fk) {
    const fromCols = parseColumnList(fk[1]);
    for (const fromCol of fromCols) {
      const col = table.columnMap.get(fromCol);
      const fkRecord = { toTable: fk[2], toColumn: fk[3], onDelete: "", onUpdate: "" };
      if (col) col.fk = fkRecord;
      addFk(schema, table.name, fromCol, fkRecord, "constraint");
    }
  }
}

function parseAlterTable(schema, stmt, file) {
  const alter = stmt.match(/^ALTER\s+TABLE\s+([A-Za-z_][\w]*)\s+([\s\S]+)$/i);
  if (!alter) return;
  const table = ensureTable(schema, alter[1]);
  table.sourceMigrations.add(file);

  const rest = alter[2].trim();
  const notNull = rest.match(/^ALTER\s+COLUMN\s+([A-Za-z_][\w]*)\s+SET\s+NOT\s+NULL$/i);
  if (notNull && table.columnMap.has(notNull[1])) {
    table.columnMap.get(notNull[1]).notNull = true;
  }

  if (/^ADD\s+COLUMN/i.test(rest)) {
    for (const part of splitTopLevel(rest)) {
      const def = part.replace(/^ADD\s+COLUMN\s+(?:IF\s+NOT\s+EXISTS\s+)?/i, "").trim();
      if (!def || /^IF\s+NOT\s+EXISTS$/i.test(def)) continue;
      addColumn(table, parseColumnDef(def));
    }
  }

  const unique = rest.match(/\bADD\s+CONSTRAINT\s+([A-Za-z_][\w]*)\s+UNIQUE\s*\(([\s\S]+?)\)/i);
  if (unique) markUnique(table, parseColumnList(unique[2]), unique[1], "constraint");

  const fk = rest.match(/\bFOREIGN\s+KEY\s*\(([\s\S]+?)\)\s+REFERENCES\s+([A-Za-z_][\w]*)\s*\(\s*([A-Za-z_][\w]*)\s*\)/i);
  if (fk) {
    for (const fromCol of parseColumnList(fk[1])) {
      const col = table.columnMap.get(fromCol);
      const fkRecord = {
        toTable: fk[2],
        toColumn: fk[3],
        onDelete: (rest.match(/\bON\s+DELETE\s+([A-Z ]+?)(?=\s+ON|\s*$)/i)?.[1] || "").trim(),
        onUpdate: (rest.match(/\bON\s+UPDATE\s+([A-Z ]+?)(?=\s+ON|\s*$)/i)?.[1] || "").trim(),
      };
      if (col) col.fk = fkRecord;
      addFk(schema, table.name, fromCol, fkRecord, "alter");
    }
  }
}

function parseCreateIndex(schema, stmt) {
  const idx = stmt.match(/^CREATE\s+(UNIQUE\s+)?INDEX(?:\s+IF\s+NOT\s+EXISTS)?\s+([A-Za-z_][\w]*)\s+ON\s+([A-Za-z_][\w]*)/i);
  if (!idx) return;

  const isUnique = Boolean(idx[1]);
  const indexName = idx[2];
  const table = ensureTable(schema, idx[3]);
  const onIndex = stmt.toUpperCase().indexOf(" ON ");
  const firstParen = stmt.indexOf("(", onIndex);
  if (firstParen === -1) return;

  const lastParen = findMatchingParen(stmt, firstParen);
  if (lastParen === -1) return;
  const cols = parseColumnList(stmt.slice(firstParen + 1, lastParen));
  const where = stmt.slice(lastParen + 1).match(/\bWHERE\s+([\s\S]+)$/i)?.[1]?.trim() || "";

  const record = { name: indexName, columns: cols, where };
  table.indexes.push(record);
  if (isUnique) markUnique(table, cols, indexName, "index", where);
}

function findMatchingParen(input, openIndex) {
  let depth = 0;
  let inSingle = false;
  let inDouble = false;
  for (let i = openIndex; i < input.length; i += 1) {
    const ch = input[i];
    const next = input[i + 1];
    if (!inDouble && ch === "'" && next === "'") {
      i += 1;
      continue;
    }
    if (!inDouble && ch === "'") inSingle = !inSingle;
    else if (!inSingle && ch === '"') inDouble = !inDouble;
    else if (!inSingle && !inDouble && ch === "(") depth += 1;
    else if (!inSingle && !inDouble && ch === ")") {
      depth -= 1;
      if (depth === 0) return i;
    }
  }
  return -1;
}

function parseTrigger(schema, stmt) {
  const trigger = stmt.match(/^CREATE\s+TRIGGER\s+([A-Za-z_][\w]*)\s+BEFORE\s+(UPDATE|DELETE)\s+ON\s+([A-Za-z_][\w]*)/i);
  if (!trigger) return;
  const table = ensureTable(schema, trigger[3]);
  table.triggers.add(`${trigger[2].toUpperCase()}:${trigger[1]}`);
}

function parseRule(schema, stmt) {
  const rule = stmt.match(/^CREATE\s+RULE\s+([A-Za-z_][\w]*)\s+AS\s+ON\s+(UPDATE|DELETE)\s+TO\s+([A-Za-z_][\w]*)/i);
  if (!rule) return;
  const table = ensureTable(schema, rule[3]);
  table.rules.add(`${rule[2].toUpperCase()}:${rule[1]}`);
}

function parseRevoke(schema, stmt) {
  const revoke = stmt.match(/^REVOKE\s+([\s\S]+?)\s+ON\s+([A-Za-z_][\w]*)\s+FROM\s+PUBLIC/i);
  if (!revoke) return;
  const table = ensureTable(schema, revoke[2]);
  table.revokes.add(revoke[1].replace(/\s+/g, " ").trim().toUpperCase());
}

function moduleForTable(tableName) {
  if (tableName.startsWith("iam_")) return "iam";
  if (tableName.startsWith("permissions_")) return "permissions";
  if (
    tableName.startsWith("permission_") ||
    tableName.startsWith("approval_workflow_") ||
    tableName === "audit_logs" ||
    tableName === "notification_settings"
  ) {
    return "permissionWorkflow";
  }
  if (tableName.startsWith("workflow__")) return "workflow";
  if (tableName.startsWith("compliance_")) return "compliance";
  if (tableName.startsWith("investment__")) return "investment";
  if (tableName.startsWith("market_") || tableName === "provider_requests_log") return "market";
  return "neutral";
}

function tableBadges(table) {
  const badges = [];
  if (table.columns.some((col) => col.name === "deleted_at")) badges.push("SOFT");
  if (table.columns.some((col) => col.name === "version")) badges.push("LOCK");
  if (isImmutable(table)) badges.push("IMM");
  if (table.revokes.size > 0 && !isImmutable(table)) badges.push("REVOKE");
  return badges;
}

function isImmutable(table) {
  const hasRejectTrigger = [...table.triggers].some((item) => /NO_|AUDIT_EVENTS_NO_|REJECT/i.test(item));
  const hasNoMutationRule = [...table.rules].some((item) => /^UPDATE:NO_|^DELETE:NO_/i.test(item));
  return hasRejectTrigger || hasNoMutationRule;
}

function rowForColumn(column) {
  const flags = [];
  if (column.pk) flags.push("PK");
  if (column.fk) flags.push("FK");
  if (column.unique && !column.pk) flags.push("UQ");
  if (column.notNull && !column.pk) flags.push("NN");
  const prefix = flags.length ? `${flags.join(" ")} ` : "";
  const fk = column.fk ? ` -> ${column.fk.toTable}.${column.fk.toColumn}` : "";
  return `${prefix}${column.name} ${column.type}${fk}`;
}

function rowsForTable(table, mode = "detail") {
  if (mode === "compact") {
    const priority = table.columns.filter(
      (col) =>
        col.pk ||
        (col.fk && !AUDIT_USER_COLUMNS.has(col.name)) ||
        col.unique ||
        ["code", "name", "status", "scope_type", "contract_id", "portfolio_id", "business_date"].includes(col.name),
    );
    return priority.slice(0, 5).map(rowForColumn);
  }

  const rows = table.columns.map(rowForColumn);
  const constraints = [
    ...table.uniqueConstraints.map((item) => `UQ (${item.columns.join(", ")})`),
    ...table.uniqueIndexes.map((item) => `UQ IDX (${item.columns.join(", ")})${item.where ? " WHERE " + item.where : ""}`),
  ];
  for (const constraint of constraints.slice(0, 4)) rows.push(constraint);
  if (constraints.length > 4) rows.push(`... ${constraints.length - 4} more unique/index rules`);
  return rows;
}

function escapeXml(value) {
  return String(value)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

function cell(id, value, style, x, y, w, h, extra = "") {
  return `        <mxCell id="${escapeXml(id)}" value="${escapeXml(value)}" style="${escapeXml(style)}" vertex="1" parent="1"${extra}>
          <mxGeometry x="${x}" y="${y}" width="${w}" height="${h}" as="geometry" />
        </mxCell>`;
}

function edge(id, value, style, source, target) {
  return `        <mxCell id="${escapeXml(id)}" value="${escapeXml(value)}" style="${escapeXml(style)}" edge="1" parent="1" source="${escapeXml(source)}" target="${escapeXml(target)}">
          <mxGeometry relative="1" as="geometry" />
        </mxCell>`;
}

function tableCellId(prefix, tableName) {
  return `${prefix}_${tableName}`;
}

function tableCard(prefix, table, x, y, width, mode = "detail") {
  const module = moduleForTable(table.name);
  const color = COLORS[module] || COLORS.neutral;
  const badges = tableBadges(table);
  const rows = rowsForTable(table, mode);
  const maxRows = mode === "compact" ? 5 : 80;
  const shown = rows.slice(0, maxRows);
  if (rows.length > maxRows) shown.push(`... ${rows.length - maxRows} more columns/rules`);
  const body = shown.map((row) => `<div>${escapeHtml(row)}</div>`).join("");
  const badgeText = badges.length ? ` <span style="font-size:9px;font-weight:600;">[${badges.join(" ")}]</span>` : "";
  const html = `<div style="background:${color.main};color:#FFFFFF;padding:7px 9px;font-weight:700;font-size:13px;">${escapeHtml(table.name)}${badgeText}</div><div style="font-family:Consolas,monospace;font-size:10px;line-height:1.45;text-align:left;padding:7px 9px;color:#1E293B;">${body}</div>`;
  const height = mode === "compact" ? 116 : 48 + shown.length * 18;
  return {
    id: tableCellId(prefix, table.name),
    xml: cell(
      tableCellId(prefix, table.name),
      html,
      `rounded=1;whiteSpace=wrap;html=1;fillColor=#FFFFFF;strokeColor=${color.stroke};strokeWidth=2;arcSize=6;shadow=1;align=left;verticalAlign=top;`,
      x,
      y,
      width,
      height,
    ),
    height,
  };
}

function escapeHtml(value) {
  return String(value)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

function textBox(id, title, body, x, y, w, h, color = COLORS.neutral) {
  const html = `<div style="padding:10px 13px;"><div style="font-size:15px;font-weight:700;color:${color.stroke};margin-bottom:6px;">${escapeHtml(title)}</div><div style="font-size:11px;line-height:1.55;color:#334155;">${escapeHtml(body)}</div></div>`;
  return cell(
    id,
    html,
    `rounded=1;whiteSpace=wrap;html=1;fillColor=${color.fill};strokeColor=${color.stroke};strokeWidth=1.5;arcSize=7;shadow=0;align=left;verticalAlign=top;`,
    x,
    y,
    w,
    h,
  );
}

function headerCells(prefix, title, subtitle, width) {
  const html = `<div style="padding:8px 18px;"><div style="font-size:28px;font-weight:800;color:#0F172A;">${escapeHtml(title)}</div><div style="font-size:13px;color:#475569;margin-top:4px;">${escapeHtml(subtitle)}</div></div>`;
  return [
    cell(
      `${prefix}_header`,
      html,
      "text;html=1;strokeColor=none;fillColor=#FFFFFF;align=left;verticalAlign=middle;whiteSpace=wrap;rounded=1;arcSize=4;shadow=0;",
      40,
      28,
      width - 80,
      82,
    ),
  ];
}

function layoutCards(prefix, schema, tableNames, startY, cardW = PAGE.cardW) {
  const cells = [];
  const positions = new Map();
  const perRow = Math.max(1, Math.floor((PAGE.width - PAGE.marginX * 2 + PAGE.gapX) / (cardW + PAGE.gapX)));
  let y = startY;

  for (let i = 0; i < tableNames.length; i += perRow) {
    const row = tableNames.slice(i, i + perRow);
    const rowCards = [];
    let rowHeight = 0;
    for (let col = 0; col < row.length; col += 1) {
      const table = schema.tables.get(row[col]);
      if (!table) continue;
      const x = PAGE.marginX + col * (cardW + PAGE.gapX);
      const card = tableCard(prefix, table, x, y, cardW, "detail");
      rowCards.push(card);
      positions.set(table.name, { id: card.id, x, y, w: cardW, h: card.height });
      rowHeight = Math.max(rowHeight, card.height);
    }
    cells.push(...rowCards.map((card) => card.xml));
    y += rowHeight + PAGE.gapY;
  }

  return { cells, positions, height: y + PAGE.marginY };
}

function fkEdgesForPage(prefix, schema, positions) {
  const cells = [];
  let count = 0;
  for (const fk of schema.fks) {
    if (!positions.has(fk.fromTable) || !positions.has(fk.toTable)) continue;
    if (AUDIT_USER_COLUMNS.has(fk.fromColumn)) continue;
    const from = positions.get(fk.fromTable);
    const to = positions.get(fk.toTable);
    const label = fk.fromColumn;
    cells.push(
      edge(
        `${prefix}_fk_${count++}`,
        label,
        "edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;endArrow=block;endFill=1;strokeColor=#475569;strokeWidth=1.6;fontSize=10;fontColor=#334155;",
        from.id,
        to.id,
      ),
    );
  }

  for (const rel of INFERRED_RELATIONSHIPS) {
    if (!positions.has(rel.fromTable) || !positions.has(rel.toTable)) continue;
    const from = positions.get(rel.fromTable);
    const to = positions.get(rel.toTable);
    cells.push(
      edge(
        `${prefix}_inferred_${count++}`,
        rel.fromColumn,
        "edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;endArrow=open;endFill=0;dashed=1;dashPattern=7 4;strokeColor=#F59E0B;strokeWidth=1.8;fontSize=10;fontColor=#92400E;",
        from.id,
        to.id,
      ),
    );
  }
  return cells;
}

function diagramXml(id, name, width, height, cells) {
  return `  <diagram id="${escapeXml(id)}" name="${escapeXml(name)}">
    <mxGraphModel dx="1600" dy="1000" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="${width}" pageHeight="${height}" math="0" shadow="0" background="#F8FAFC">
      <root>
        <mxCell id="0" />
        <mxCell id="1" parent="0" />
${cells.join("\n")}
      </root>
    </mxGraphModel>
  </diagram>`;
}

function countByModule(schema) {
  const counts = new Map();
  for (const table of schema.tables.values()) {
    const module = moduleForTable(table.name);
    counts.set(module, (counts.get(module) || 0) + 1);
  }
  return counts;
}

function overviewPage(schema) {
  const width = 5200;
  const height = 3600;
  const cells = headerCells(
    "overview",
    "IMS-TH-SOLUTION Database ERD",
    `${schema.tables.size} tables generated from ${schema.migrations.length} PostgreSQL migration files. Multi-page draw.io ERD, source date ${GENERATED_DATE}.`,
    width,
  );

  const legendHtml = `<div style="padding:10px 14px;"><div style="font-size:15px;font-weight:700;color:#0F172A;margin-bottom:7px;">Legend</div><div style="font-size:11px;line-height:1.7;color:#334155;">PK = primary key<br/>FK = foreign key<br/>UQ = unique<br/>NN = not null<br/>SOFT = soft delete column<br/>LOCK = optimistic version column<br/>IMM = DB-level immutable guard<br/>Dashed amber line = inferred or polymorphic relation without FK</div></div>`;
  cells.push(
    cell(
      "overview_legend",
      legendHtml,
      "rounded=1;whiteSpace=wrap;html=1;fillColor=#FFFFFF;strokeColor=#CBD5E1;strokeWidth=1;arcSize=6;shadow=1;align=left;verticalAlign=top;",
      40,
      130,
      430,
      250,
    ),
  );

  const groups = [
    { id: "overview_iam", title: "IAM + Core Permissions", modules: ["iam", "permissions"], x: 40, y: 430, w: 1160, h: 760 },
    { id: "overview_permission_workflow", title: "Permission Approval Workflow", modules: ["permissionWorkflow"], x: 1240, y: 430, w: 1760, h: 760 },
    { id: "overview_workflow", title: "Workflow + Scheduler", modules: ["workflow"], x: 3040, y: 430, w: 850, h: 760 },
    { id: "overview_compliance", title: "Compliance", modules: ["compliance"], x: 3930, y: 430, w: 850, h: 760 },
    { id: "overview_investment", title: "Investment", modules: ["investment"], x: 40, y: 1240, w: 3040, h: 1800 },
    { id: "overview_market", title: "Market Data", modules: ["market"], x: 3120, y: 1240, w: 780, h: 460 },
  ];

  const positions = new Map();
  for (const group of groups) {
    const color = COLORS[group.modules[0]] || COLORS.neutral;
    cells.push(
      cell(
        group.id,
        `${group.title}`,
        `rounded=1;whiteSpace=wrap;html=1;fillColor=${color.fill};strokeColor=${color.stroke};strokeWidth=2;arcSize=6;shadow=0;verticalAlign=top;align=left;fontStyle=1;fontSize=15;fontColor=${color.stroke};spacingTop=8;spacingLeft=12;`,
        group.x,
        group.y,
        group.w,
        group.h,
      ),
    );

    const tables = [...schema.tables.values()]
      .filter((table) => group.modules.includes(moduleForTable(table.name)))
      .map((table) => table.name)
      .sort();
    const cardW = group.w > 1200 ? 260 : 250;
    const gapX = 18;
    const gapY = 18;
    const perRow = Math.max(1, Math.floor((group.w - 40 + gapX) / (cardW + gapX)));
    let y = group.y + 48;
    for (let i = 0; i < tables.length; i += perRow) {
      const row = tables.slice(i, i + perRow);
      let rowH = 0;
      for (let col = 0; col < row.length; col += 1) {
        const table = schema.tables.get(row[col]);
        const x = group.x + 20 + col * (cardW + gapX);
        const card = tableCard("overview", table, x, y, cardW, "compact");
        cells.push(card.xml);
        positions.set(table.name, { id: card.id, x, y, w: cardW, h: card.height });
        rowH = Math.max(rowH, card.height);
      }
      y += rowH + gapY;
    }
  }

  cells.push(...overviewCriticalNotes());
  cells.push(...overviewContextEdges());
  cells.push(...overviewInferredEdges(positions));

  return diagramXml("ims_overview", "00 - Overview", width, height, cells);
}

function overviewCriticalNotes() {
  const cells = [];
  const x = 3940;
  const y = 1240;
  const w = 1200;
  const h = 1270;
  const items = MANUAL_NOTES.slice(0, 8)
    .map((note, idx) => `<div style="margin-bottom:8px;"><b>${idx + 1}. ${escapeHtml(note.title)}</b><br/>${escapeHtml(note.body)}</div>`)
    .join("");
  const html = `<div style="padding:14px 18px;"><div style="font-size:17px;font-weight:800;color:#92400E;margin-bottom:10px;">Critical Things To Notice</div><div style="font-size:11px;line-height:1.58;color:#334155;">${items}</div></div>`;
  cells.push(
    cell(
      "overview_critical_notes",
      html,
      "rounded=1;whiteSpace=wrap;html=1;fillColor=#FFFBEB;strokeColor=#F59E0B;strokeWidth=2;arcSize=8;shadow=1;align=left;verticalAlign=top;",
      x,
      y,
      w,
      h,
    ),
  );
  return cells;
}

function overviewContextEdges() {
  const style =
    "edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;endArrow=block;endFill=1;strokeColor=#334155;strokeWidth=2;fontSize=12;fontColor=#0F172A;";
  return [
    edge("overview_context_1", "users and groups", style, "overview_iam", "overview_permission_workflow"),
    edge("overview_context_2", "users, roles, audit actors", style, "overview_iam", "overview_workflow"),
    edge("overview_context_3", "users, roles, audit actors", style, "overview_iam", "overview_compliance"),
    edge("overview_context_4", "users, roles, audit actors", style, "overview_iam", "overview_investment"),
    edge("overview_context_5", "contract_id = fund.id", style, "overview_investment", "overview_workflow"),
    edge("overview_context_6", "portfolio/contract checks", style, "overview_investment", "overview_compliance"),
    edge("overview_context_7", "fund/portfolio grants", style, "overview_permission_workflow", "overview_investment"),
    edge("overview_context_8", "symbol/provider data", style, "overview_market", "overview_investment"),
  ];
}

function overviewInferredEdges(positions) {
  const cells = [];
  const pairs = [
    ["workflow__day_states", "investment__funds", "contract_id"],
    ["compliance_check_records", "investment__portfolios", "portfolio_id"],
    ["permissions_data_rights", "investment__funds", "contract_id"],
    ["investment__research_reports", "investment__instruments", "instrument_code"],
    ["market_symbols", "investment__instruments", "symbol"],
  ];
  pairs.forEach(([from, to, label], idx) => {
    if (!positions.has(from) || !positions.has(to)) return;
    cells.push(
      edge(
        `overview_inferred_edge_${idx}`,
        label,
        "edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;endArrow=open;endFill=0;dashed=1;dashPattern=7 4;strokeColor=#F59E0B;strokeWidth=2;fontSize=10;fontColor=#92400E;",
        positions.get(from).id,
        positions.get(to).id,
      ),
    );
  });
  return cells;
}

function detailPage(schema, group, idx) {
  const prefix = `p${String(idx).padStart(2, "0")}`;
  const header = headerCells(prefix, group.name, group.description, PAGE.width);
  const noteCells = [];
  const noteText = detailPageNote(group.name);
  if (noteText) {
    noteCells.push(textBox(`${prefix}_note`, "Senior Engineer Note", noteText, PAGE.marginX, 126, PAGE.width - PAGE.marginX * 2, 86, COLORS.warning));
  }
  const layout = layoutCards(prefix, schema, group.tables, noteText ? 246 : 144);
  const edgeCells = fkEdgesForPage(prefix, schema, layout.positions);
  const height = Math.max(1500, layout.height + 30);
  return diagramXml(`ims_${prefix}`, group.name, PAGE.width, height, [...header, ...noteCells, ...layout.cells, ...edgeCells]);
}

function detailPageNote(name) {
  if (name.includes("Permission Approval")) {
    return "This page is the newer permission-request and approval workflow surface. It coexists with the older permissions_* grant tables, so rollout needs a clear migration/evaluation boundary.";
  }
  if (name.includes("Workflow")) {
    return "contract_id is not FK-enforced here; operationally it maps to investment__funds.id. transition_log is described as append-only, but this migration set does not add a DB no-update/no-delete trigger.";
  }
  if (name.includes("Compliance")) {
    return "Rule versioning is strong, but portfolio_id/contract_id/order_id are plain UUIDs. compliance_check_records and overrides are protected by privilege revokes, not the same rejecting trigger pattern used by investment.";
  }
  if (name.includes("Reference")) {
    return "investment__funds is the contract master. Research report user/contract/instrument references are intentionally plain values in PoC scope and are shown as dashed inferred links where possible.";
  }
  if (name.includes("Process")) {
    return "Ledger and snapshot tables are append-only; mutable projections are portfolio_positions and cash_balances. AUM snapshots use polymorphic scope_type/scope_id instead of two hard FKs.";
  }
  if (name.includes("IAM")) {
    return "iam_users has soft delete, but username uniqueness is still a full table constraint, so a deleted username is not reusable without schema change.";
  }
  if (name.includes("Market")) {
    return "Market data is intentionally provider/symbol oriented and has no FK to the investment instrument master. Reconciliation belongs in service logic.";
  }
  return "";
}

function criticalPage(schema) {
  const width = 2600;
  const height = 1900;
  const cells = headerCells(
    "critical",
    "Critical Analysis",
    "Operational and modeling details that are easy to miss when reading only table names.",
    width,
  );

  const cols = 2;
  const w = 1180;
  const gap = 40;
  const startY = 145;
  const boxH = 154;
  MANUAL_NOTES.forEach((note, idx) => {
    const col = idx % cols;
    const row = Math.floor(idx / cols);
    const x = 60 + col * (w + gap);
    const y = startY + row * (boxH + 28);
    cells.push(textBox(`critical_note_${idx}`, `${idx + 1}. ${note.title}`, note.body, x, y, w, boxH, COLORS.warning));
  });

  const counts = countByModule(schema);
  const inventory = [
    `Total tables: ${schema.tables.size}`,
    `IAM: ${counts.get("iam") || 0}`,
    `Core permissions: ${counts.get("permissions") || 0}`,
    `Permission workflow: ${counts.get("permissionWorkflow") || 0}`,
    `Workflow: ${counts.get("workflow") || 0}`,
    `Compliance: ${counts.get("compliance") || 0}`,
    `Investment: ${counts.get("investment") || 0}`,
    `Market data: ${counts.get("market") || 0}`,
    `Migration files parsed: ${schema.migrations.length}`,
  ].join("<br/>");
  const html = `<div style="padding:14px 18px;"><div style="font-size:17px;font-weight:800;color:#0F172A;margin-bottom:8px;">Inventory</div><div style="font-family:Consolas,monospace;font-size:12px;line-height:1.7;color:#334155;">${inventory}</div></div>`;
  cells.push(
    cell(
      "critical_inventory",
      html,
      "rounded=1;whiteSpace=wrap;html=1;fillColor=#FFFFFF;strokeColor=#CBD5E1;strokeWidth=1.5;arcSize=7;shadow=1;align=left;verticalAlign=top;",
      60,
      1560,
      1180,
      220,
    ),
  );

  const criticalPageName = `${String(PAGE_GROUPS.length + 1).padStart(2, "0")} - Critical Notes`;
  return diagramXml("ims_critical", criticalPageName, width, height, cells);
}

function buildDrawio(schema) {
  const diagrams = [overviewPage(schema), ...PAGE_GROUPS.map((group, idx) => detailPage(schema, group, idx + 1)), criticalPage(schema)];
  return `<mxfile host="app.diagrams.net" agent="Codex/IMS-TH" version="24.7.17" type="device">
${diagrams.join("\n")}
</mxfile>
`;
}

function buildNotes(schema) {
  const counts = countByModule(schema);
  const lines = [];
  lines.push("# IMS-TH-SOLUTION ERD Notes");
  lines.push("");
  lines.push(`Generated from \`database/migrations/*.up.sql\` on ${GENERATED_DATE}.`);
  lines.push("");
  lines.push("## Inventory");
  lines.push("");
  lines.push(`- Total tables: ${schema.tables.size}`);
  lines.push(`- IAM: ${counts.get("iam") || 0}`);
  lines.push(`- Core permissions: ${counts.get("permissions") || 0}`);
  lines.push(`- Permission workflow: ${counts.get("permissionWorkflow") || 0}`);
  lines.push(`- Workflow: ${counts.get("workflow") || 0}`);
  lines.push(`- Compliance: ${counts.get("compliance") || 0}`);
  lines.push(`- Investment: ${counts.get("investment") || 0}`);
  lines.push(`- Market data: ${counts.get("market") || 0}`);
  lines.push("");
  lines.push("## Draw.io Pages");
  lines.push("");
  lines.push("- 00 - Overview");
  for (const group of PAGE_GROUPS) lines.push(`- ${group.name}`);
  lines.push(`- ${String(PAGE_GROUPS.length + 1).padStart(2, "0")} - Critical Notes`);
  lines.push("");
  lines.push("## Critical Things To Notice");
  lines.push("");
  for (const note of MANUAL_NOTES) {
    lines.push(`### ${note.title}`);
    lines.push(note.body);
    lines.push("");
  }
  lines.push("## Inferred / Non-FK Relationships");
  lines.push("");
  for (const rel of INFERRED_RELATIONSHIPS) {
    lines.push(`- ${rel.fromTable}.${rel.fromColumn} -> ${rel.toTable}.${rel.toColumn}: ${rel.note}`);
  }
  lines.push("");
  return `${lines.join("\n")}\n`;
}

function main() {
  fs.mkdirSync(outDir, { recursive: true });
  const schema = parseSchema();
  fs.writeFileSync(drawioPath, buildDrawio(schema), "utf8");
  fs.writeFileSync(notesPath, buildNotes(schema), "utf8");
  console.log(`Wrote ${path.relative(repoRoot, drawioPath)} (${schema.tables.size} tables)`);
  console.log(`Wrote ${path.relative(repoRoot, notesPath)}`);
}

main();
