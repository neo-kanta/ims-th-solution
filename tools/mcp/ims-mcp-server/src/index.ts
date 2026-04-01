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

import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import pg from "pg";
import fs from "fs/promises";
import path from "path";

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

const DB_CONFIG = {
  host: process.env.IMS_DB_HOST ?? "localhost",
  port: parseInt(process.env.IMS_DB_PORT ?? "5437", 10),
  database: process.env.IMS_DB_NAME ?? "ims_dev",
  user: process.env.IMS_DB_USER ?? "ims_app",
  password: process.env.IMS_DB_PASSWORD ?? "ims_dev_password",
  max: 5,
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 5000,
};

// Path to the IMS project root — adjust if your MCP server lives elsewhere
const PROJECT_ROOT =
  process.env.IMS_PROJECT_ROOT ?? path.resolve(process.cwd());
const BACKEND_ROOT = path.join(PROJECT_ROOT, "backend");
const MIGRATIONS_DIR = path.join(PROJECT_ROOT, "database", "migrations");

// ---------------------------------------------------------------------------
// Database pool (lazy init)
// ---------------------------------------------------------------------------

let pool: pg.Pool | null = null;

function getPool(): pg.Pool {
  if (!pool) {
    pool = new pg.Pool(DB_CONFIG);
    pool.on("error", (err) => {
      console.error("[ims-mcp] Unexpected pool error:", err.message);
    });
  }
  return pool;
}

async function safeQuery(
  sql: string,
  params: unknown[] = [],
): Promise<pg.QueryResult> {
  const client = await getPool().connect();
  try {
    // Set read-only transaction for safety
    await client.query("SET TRANSACTION READ ONLY");
    const result = await client.query(sql, params);
    return result;
  } finally {
    client.release();
  }
}

// ---------------------------------------------------------------------------
// Workflow state machine (mirrors Go domain logic)
// ---------------------------------------------------------------------------

const WORKFLOW_STATES = [
  "NOT_STARTED",
  "INVESTMENT_DAY_START",
  "MANAGER_APPROVAL",
  "TRANSACTION_CLOSING",
  "ACCOUNTING_CLOSING",
] as const;

type WorkflowState = (typeof WORKFLOW_STATES)[number];

const VALID_TRANSITIONS: Record<WorkflowState, WorkflowState[]> = {
  NOT_STARTED: ["INVESTMENT_DAY_START"],
  INVESTMENT_DAY_START: ["MANAGER_APPROVAL"],
  MANAGER_APPROVAL: ["TRANSACTION_CLOSING"],
  TRANSACTION_CLOSING: ["ACCOUNTING_CLOSING"],
  ACCOUNTING_CLOSING: ["NOT_STARTED"], // next day cycle
};

// Rollback transitions (explicit, auditable)
const ROLLBACK_TRANSITIONS: Record<WorkflowState, WorkflowState[]> = {
  NOT_STARTED: [],
  INVESTMENT_DAY_START: ["NOT_STARTED"],
  MANAGER_APPROVAL: ["INVESTMENT_DAY_START"],
  TRANSACTION_CLOSING: ["MANAGER_APPROVAL"],
  ACCOUNTING_CLOSING: ["TRANSACTION_CLOSING"],
};

// ---------------------------------------------------------------------------
// Domain rule registry (structured, not ad-hoc)
// ---------------------------------------------------------------------------

interface DomainRule {
  module: string;
  code: string;
  description: string;
  severity: "error" | "warning";
  enforcement: "server" | "database" | "both";
}

const DOMAIN_RULES: DomainRule[] = [
  // Workflow rules
  {
    module: "workflow",
    code: "WF-001",
    description:
      "Investment transactions must not occur before Investment Day Start",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "workflow",
    code: "WF-002",
    description:
      "Manager Approval can only happen after all required transactions/reviews for that day are done",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "workflow",
    code: "WF-003",
    description:
      "After Manager Approval, managers can no longer perform transactions for that day",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "workflow",
    code: "WF-004",
    description:
      "Transaction Closing and Accounting Closing are separate states — cannot skip",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "workflow",
    code: "WF-005",
    description:
      "Rollback/cancel operations must be modeled explicitly with audit trail",
    severity: "error",
    enforcement: "both",
  },
  {
    module: "workflow",
    code: "WF-006",
    description:
      "All workflow actions must log actor, time, contract, action, and result",
    severity: "error",
    enforcement: "both",
  },

  // Stock Investment rules
  {
    module: "stock_investment",
    code: "SI-001",
    description:
      "Minimum trading unit validation — order quantity must be a multiple of the minimum trading unit",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "stock_investment",
    code: "SI-002",
    description:
      "Price tick validation — order price must align with exchange tick size rules",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "stock_investment",
    code: "SI-003",
    description:
      "Buy/sell recommendation must be linked to an approved analysis report",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "stock_investment",
    code: "SI-004",
    description:
      "Sell quantity must not exceed allowed inventory (available shares)",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "stock_investment",
    code: "SI-005",
    description:
      "Submission must go through approval workflow before execution",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "stock_investment",
    code: "SI-006",
    description:
      "If report approval is required, investment decision cannot bypass it",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "stock_investment",
    code: "SI-007",
    description:
      "Contract/manager visibility restrictions — users can only see contracts they are authorized for",
    severity: "error",
    enforcement: "both",
  },
  {
    module: "stock_investment",
    code: "SI-008",
    description:
      "Stock pool / investment scope check — instrument must be in the fund's allowed investment universe",
    severity: "error",
    enforcement: "server",
  },

  // Leave / Delegation rules
  {
    module: "leave_delegation",
    code: "LD-001",
    description:
      "Users on approved leave are restricted from normal login/operations",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "leave_delegation",
    code: "LD-002",
    description:
      "Delegated users inherit functional AND data permissions during the leave window only",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "leave_delegation",
    code: "LD-003",
    description:
      "Agent setup supports priority-based delegation — up to 6 agents per contract",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "leave_delegation",
    code: "LD-004",
    description:
      "Leave, temporary leave, and leave cancellation flows must all exist",
    severity: "error",
    enforcement: "server",
  },

  // Permissions rules
  {
    module: "permissions",
    code: "PM-001",
    description: "Permission checks must be server-side — never frontend-only",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "permissions",
    code: "PM-002",
    description:
      "Data permissions are per-contract/fund — only authorized users can see or act on contract data",
    severity: "error",
    enforcement: "both",
  },
  {
    module: "permissions",
    code: "PM-003",
    description:
      "Function permissions are assigned to groups, users are assigned to groups",
    severity: "error",
    enforcement: "both",
  },

  // Compliance / IRG hooks
  {
    module: "compliance",
    code: "IRG-001",
    description:
      "Pre-trade checking must be enforced before any order execution",
    severity: "error",
    enforcement: "server",
  },
  {
    module: "compliance",
    code: "IRG-002",
    description:
      "Extension points must exist for blacklist/whitelist, investment ratio, instrument restriction, credit/rating rules",
    severity: "warning",
    enforcement: "server",
  },
];

// ---------------------------------------------------------------------------
// MCP Server
// ---------------------------------------------------------------------------

const server = new McpServer({
  name: "ims-thailand",
  version: "1.0.0",
});

// ===========================
// TOOL: db_query
// ===========================
server.tool(
  "db_query",
  "Run a read-only SQL query against the IMS PostgreSQL database. Returns rows as JSON. Only SELECT statements are allowed.",
  {
    sql: z.string().describe("SQL SELECT query to execute"),
    params: z
      .array(z.string())
      .optional()
      .describe("Query parameters ($1, $2, etc.)"),
  },
  async ({ sql, params }) => {
    // Safety: block writes
    const normalized = sql.trim().toUpperCase();
    if (
      !normalized.startsWith("SELECT") &&
      !normalized.startsWith("WITH") &&
      !normalized.startsWith("EXPLAIN")
    ) {
      return {
        content: [
          {
            type: "text",
            text: "ERROR: Only SELECT, WITH (CTE), and EXPLAIN queries are allowed. This is a read-only tool.",
          },
        ],
      };
    }

    try {
      const result = await safeQuery(sql, params ?? []);
      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(
              {
                rowCount: result.rowCount,
                fields: result.fields.map((f) => ({
                  name: f.name,
                  dataTypeID: f.dataTypeID,
                })),
                rows: result.rows.slice(0, 100), // cap at 100 rows
              },
              null,
              2,
            ),
          },
        ],
      };
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      return { content: [{ type: "text", text: `SQL Error: ${message}` }] };
    }
  },
);

// ===========================
// TOOL: db_schema
// ===========================
server.tool(
  "db_schema",
  "List all tables in the IMS database with their columns, data types, and constraints. Useful for understanding the current database structure.",
  {
    schema_name: z
      .string()
      .optional()
      .default("public")
      .describe("PostgreSQL schema name (default: public)"),
  },
  async ({ schema_name }) => {
    try {
      const result = await safeQuery(
        `
        SELECT
          t.table_name,
          c.column_name,
          c.data_type,
          c.column_default,
          c.is_nullable,
          c.character_maximum_length,
          tc.constraint_type,
          kcu.constraint_name
        FROM information_schema.tables t
        JOIN information_schema.columns c
          ON c.table_schema = t.table_schema AND c.table_name = t.table_name
        LEFT JOIN information_schema.key_column_usage kcu
          ON kcu.table_schema = t.table_schema
          AND kcu.table_name = t.table_name
          AND kcu.column_name = c.column_name
        LEFT JOIN information_schema.table_constraints tc
          ON tc.constraint_name = kcu.constraint_name
          AND tc.table_schema = kcu.table_schema
        WHERE t.table_schema = $1
          AND t.table_type = 'BASE TABLE'
        ORDER BY t.table_name, c.ordinal_position
        `,
        [schema_name],
      );

      // Group by table
      const tables: Record<string, unknown[]> = {};
      for (const row of result.rows) {
        const tbl = row.table_name as string;
        if (!tables[tbl]) tables[tbl] = [];
        tables[tbl].push(row);
      }

      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(
              {
                schema: schema_name,
                table_count: Object.keys(tables).length,
                tables,
              },
              null,
              2,
            ),
          },
        ],
      };
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      return { content: [{ type: "text", text: `Schema Error: ${message}` }] };
    }
  },
);

// ===========================
// TOOL: db_table_info
// ===========================
server.tool(
  "db_table_info",
  "Get detailed information about a specific table: columns, indexes, foreign keys, row count estimate.",
  {
    table_name: z.string().describe("Table name to inspect"),
  },
  async ({ table_name }) => {
    try {
      // Columns
      const cols = await safeQuery(
        `
        SELECT column_name, data_type, column_default, is_nullable, character_maximum_length
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = $1
        ORDER BY ordinal_position
        `,
        [table_name],
      );

      // Indexes
      const indexes = await safeQuery(
        `
        SELECT indexname, indexdef
        FROM pg_indexes
        WHERE schemaname = 'public' AND tablename = $1
        `,
        [table_name],
      );

      // Foreign keys
      const fks = await safeQuery(
        `
        SELECT
          tc.constraint_name,
          kcu.column_name,
          ccu.table_name AS foreign_table_name,
          ccu.column_name AS foreign_column_name
        FROM information_schema.table_constraints AS tc
        JOIN information_schema.key_column_usage AS kcu
          ON tc.constraint_name = kcu.constraint_name
        JOIN information_schema.constraint_column_usage AS ccu
          ON ccu.constraint_name = tc.constraint_name
        WHERE tc.constraint_type = 'FOREIGN KEY'
          AND tc.table_name = $1
        `,
        [table_name],
      );

      // Row count estimate
      const countEst = await safeQuery(
        `SELECT reltuples::bigint AS estimate FROM pg_class WHERE relname = $1`,
        [table_name],
      );

      // Comments
      const comments = await safeQuery(
        `
        SELECT col_description(oid, a.attnum) as column_comment, a.attname as column_name
        FROM pg_class c
        JOIN pg_attribute a ON a.attrelid = c.oid
        WHERE c.relname = $1 AND a.attnum > 0 AND NOT a.attisdropped
        ORDER BY a.attnum
        `,
        [table_name],
      );

      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(
              {
                table: table_name,
                row_count_estimate: countEst.rows[0]?.estimate ?? "unknown",
                columns: cols.rows,
                indexes: indexes.rows,
                foreign_keys: fks.rows,
                column_comments: comments.rows.filter((r) => r.column_comment),
              },
              null,
              2,
            ),
          },
        ],
      };
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      return {
        content: [{ type: "text", text: `Table Info Error: ${message}` }],
      };
    }
  },
);

// ===========================
// TOOL: db_migrations
// ===========================
server.tool(
  "db_migrations",
  "Show the migration history from the schema_migrations table — which migrations have been applied to the database.",
  {},
  async () => {
    try {
      const result = await safeQuery(
        `SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 50`,
      );
      return {
        content: [{ type: "text", text: JSON.stringify(result.rows, null, 2) }],
      };
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      return {
        content: [
          {
            type: "text",
            text: `Migration query failed (table may not exist yet): ${message}`,
          },
        ],
      };
    }
  },
);

// ===========================
// TOOL: workflow_states
// ===========================
server.tool(
  "workflow_states",
  "Show the IMS workflow state machine: all valid states, forward transitions, rollback transitions, and business rules.",
  {},
  async () => {
    return {
      content: [
        {
          type: "text",
          text: JSON.stringify(
            {
              states: WORKFLOW_STATES,
              forward_transitions: VALID_TRANSITIONS,
              rollback_transitions: ROLLBACK_TRANSITIONS,
              lifecycle:
                "NOT_STARTED → INVESTMENT_DAY_START → MANAGER_APPROVAL → TRANSACTION_CLOSING → ACCOUNTING_CLOSING → (next day)",
              rules: DOMAIN_RULES.filter((r) => r.module === "workflow"),
            },
            null,
            2,
          ),
        },
      ],
    };
  },
);

// ===========================
// TOOL: workflow_validate
// ===========================
server.tool(
  "workflow_validate",
  "Check if a workflow state transition is valid. Returns whether it's a forward move, rollback, or invalid.",
  {
    from_state: z
      .enum(WORKFLOW_STATES as unknown as [string, ...string[]])
      .describe("Current workflow state"),
    to_state: z
      .enum(WORKFLOW_STATES as unknown as [string, ...string[]])
      .describe("Target workflow state"),
  },
  async ({ from_state, to_state }) => {
    const from = from_state as WorkflowState;
    const to = to_state as WorkflowState;

    const isForward = VALID_TRANSITIONS[from]?.includes(to) ?? false;
    const isRollback = ROLLBACK_TRANSITIONS[from]?.includes(to) ?? false;

    let status: string;
    let detail: string;

    if (isForward) {
      status = "VALID_FORWARD";
      detail = `Transition ${from} → ${to} is a valid forward move.`;
    } else if (isRollback) {
      status = "VALID_ROLLBACK";
      detail = `Transition ${from} → ${to} is a valid rollback. Must be audited with reason.`;
    } else {
      status = "INVALID";
      detail = `Transition ${from} → ${to} is NOT allowed. Valid forward: [${VALID_TRANSITIONS[from]?.join(", ")}], valid rollback: [${ROLLBACK_TRANSITIONS[from]?.join(", ")}]`;
    }

    return {
      content: [
        {
          type: "text",
          text: JSON.stringify(
            { from_state, to_state, status, detail },
            null,
            2,
          ),
        },
      ],
    };
  },
);

// ===========================
// TOOL: perm_user_rights
// ===========================
server.tool(
  "perm_user_rights",
  "Show all permissions for a user: their groups, function rights, and data rights (contract visibility).",
  {
    username: z.string().describe("Username to look up"),
  },
  async ({ username }) => {
    try {
      const userResult = await safeQuery(
        `SELECT id, username, display_name, is_active, is_on_leave FROM iam_users WHERE username = $1`,
        [username],
      );
      if (userResult.rowCount === 0) {
        return {
          content: [{ type: "text", text: `User '${username}' not found.` }],
        };
      }
      const user = userResult.rows[0];
      const userId = user.id;

      // Groups
      const groups = await safeQuery(
        `
        SELECT g.id, g.name, g.description
        FROM permissions_groups g
        JOIN permissions_accounts_groups ag ON ag.group_id = g.id
        WHERE ag.user_id = $1
        `,
        [userId],
      );

      // Function rights (via groups)
      const funcRights = await safeQuery(
        `
        SELECT DISTINCT fr.permission_code, fr.is_granted, g.name as group_name
        FROM permissions_function_rights fr
        JOIN permissions_groups g ON g.id = fr.group_id
        JOIN permissions_accounts_groups ag ON ag.group_id = g.id
        WHERE ag.user_id = $1
        ORDER BY fr.permission_code
        `,
        [userId],
      );

      // Data rights
      const dataRights = await safeQuery(
        `SELECT contract_id, is_granted FROM permissions_data_rights WHERE user_id = $1`,
        [userId],
      );

      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(
              {
                user,
                groups: groups.rows,
                function_rights: funcRights.rows,
                data_rights: dataRights.rows,
              },
              null,
              2,
            ),
          },
        ],
      };
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      return {
        content: [
          { type: "text", text: `Permission lookup error: ${message}` },
        ],
      };
    }
  },
);

// ===========================
// TOOL: perm_check
// ===========================
server.tool(
  "perm_check",
  "Check if a specific user has a specific function permission code. Returns granted/denied with context.",
  {
    username: z.string().describe("Username to check"),
    permission_code: z
      .string()
      .describe(
        "Permission code (e.g. 'WORKFLOW_START_DAY', 'INVESTMENT_CREATE_ORDER')",
      ),
  },
  async ({ username, permission_code }) => {
    try {
      const result = await safeQuery(
        `
        SELECT fr.permission_code, fr.is_granted, g.name as group_name
        FROM permissions_function_rights fr
        JOIN permissions_groups g ON g.id = fr.group_id
        JOIN permissions_accounts_groups ag ON ag.group_id = g.id
        JOIN iam_users u ON u.id = ag.user_id
        WHERE u.username = $1 AND fr.permission_code = $2
        `,
        [username, permission_code],
      );

      if (result.rowCount === 0) {
        return {
          content: [
            {
              type: "text",
              text: JSON.stringify(
                {
                  username,
                  permission_code,
                  result: "DENIED",
                  reason: "No matching permission found for this user",
                },
                null,
                2,
              ),
            },
          ],
        };
      }

      const granted = result.rows.some((r) => r.is_granted);
      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(
              {
                username,
                permission_code,
                result: granted ? "GRANTED" : "DENIED",
                sources: result.rows,
              },
              null,
              2,
            ),
          },
        ],
      };
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      return {
        content: [{ type: "text", text: `Permission check error: ${message}` }],
      };
    }
  },
);

// ===========================
// TOOL: perm_groups
// ===========================
server.tool(
  "perm_groups",
  "List all permission groups with their members and assigned function rights.",
  {},
  async () => {
    try {
      const groups = await safeQuery(
        `SELECT id, name, description, is_active FROM permissions_groups ORDER BY name`,
      );

      const details = [];
      for (const group of groups.rows) {
        const members = await safeQuery(
          `
          SELECT u.username, u.display_name
          FROM iam_users u
          JOIN permissions_accounts_groups ag ON ag.user_id = u.id
          WHERE ag.group_id = $1
          `,
          [group.id],
        );

        const rights = await safeQuery(
          `SELECT permission_code, is_granted FROM permissions_function_rights WHERE group_id = $1 ORDER BY permission_code`,
          [group.id],
        );

        details.push({
          ...group,
          members: members.rows,
          function_rights: rights.rows,
        });
      }

      return {
        content: [{ type: "text", text: JSON.stringify(details, null, 2) }],
      };
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      return { content: [{ type: "text", text: `Groups error: ${message}` }] };
    }
  },
);

// ===========================
// TOOL: domain_rules
// ===========================
server.tool(
  "domain_rules",
  "List the business validation rules for a specific module (or all modules). These are the IMS domain rules that the backend MUST enforce.",
  {
    module: z
      .string()
      .optional()
      .describe(
        "Module name to filter (workflow, stock_investment, leave_delegation, permissions, compliance). Leave empty for all.",
      ),
  },
  async ({ module }) => {
    const filtered = module
      ? DOMAIN_RULES.filter((r) => r.module === module)
      : DOMAIN_RULES;

    return {
      content: [
        {
          type: "text",
          text: JSON.stringify(
            {
              filter: module ?? "all",
              rule_count: filtered.length,
              rules: filtered,
            },
            null,
            2,
          ),
        },
      ],
    };
  },
);

// ===========================
// TOOL: audit_log
// ===========================
server.tool(
  "audit_log",
  "Query the audit log table for recent entries. Useful for debugging workflow actions, permission changes, and domain events.",
  {
    limit: z.number().optional().default(20).describe("Max rows to return"),
    actor_username: z.string().optional().describe("Filter by actor username"),
    action_type: z.string().optional().describe("Filter by action type"),
  },
  async ({ limit, actor_username, action_type }) => {
    try {
      let sql = `
        SELECT * FROM audit_log
        WHERE 1=1
      `;
      const params: unknown[] = [];
      let paramIdx = 1;

      if (actor_username) {
        sql += ` AND actor_username = $${paramIdx++}`;
        params.push(actor_username);
      }
      if (action_type) {
        sql += ` AND action_type = $${paramIdx++}`;
        params.push(action_type);
      }
      sql += ` ORDER BY created_at DESC LIMIT $${paramIdx}`;
      params.push(limit);

      const result = await safeQuery(sql, params);
      return {
        content: [{ type: "text", text: JSON.stringify(result.rows, null, 2) }],
      };
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      return {
        content: [
          {
            type: "text",
            text: `Audit log query failed (table may not exist yet): ${message}`,
          },
        ],
      };
    }
  },
);

// ===========================
// TOOL: project_modules
// ===========================
server.tool(
  "project_modules",
  "List all backend domain modules, their directory structure, and whether they have route registrations, domain entities, or TODO markers.",
  {},
  async () => {
    try {
      const internalDir = path.join(BACKEND_ROOT, "internal");
      const entries = await fs.readdir(internalDir, { withFileTypes: true });
      const modules = [];

      for (const entry of entries) {
        if (!entry.isDirectory()) continue;

        const modPath = path.join(internalDir, entry.name);
        const modFile = path.join(modPath, "module.go");

        let hasModuleGo = false;
        let hasRoutes = false;
        let hasTODO = false;
        let subfolders: string[] = [];

        try {
          const modContent = await fs.readFile(modFile, "utf-8");
          hasModuleGo = true;
          hasRoutes =
            modContent.includes("RegisterRoutes") &&
            !modContent.includes("// TODO");
          hasTODO = modContent.includes("TODO");
        } catch {
          // module.go doesn't exist
        }

        try {
          const subs = await fs.readdir(modPath, { withFileTypes: true });
          subfolders = subs.filter((s) => s.isDirectory()).map((s) => s.name);
        } catch {
          // can't read directory
        }

        modules.push({
          name: entry.name,
          has_module_go: hasModuleGo,
          has_routes_registered: hasRoutes,
          has_todo: hasTODO,
          subfolders,
        });
      }

      return {
        content: [{ type: "text", text: JSON.stringify(modules, null, 2) }],
      };
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      return {
        content: [
          {
            type: "text",
            text: `Project modules scan error: ${message}. Make sure IMS_PROJECT_ROOT is set correctly.`,
          },
        ],
      };
    }
  },
);

// ===========================
// TOOL: project_migrations
// ===========================
server.tool(
  "project_migrations",
  "List all migration files in the database/migrations directory (filesystem view, not DB state).",
  {},
  async () => {
    try {
      const files = await fs.readdir(MIGRATIONS_DIR);
      const migrations = files
        .filter((f) => f.endsWith(".sql"))
        .sort()
        .map((f) => {
          const match = f.match(/^(\d+)_([^_]+)__(.+)\.(up|down)\.sql$/);
          return {
            filename: f,
            version: match?.[1] ?? "unknown",
            module: match?.[2] ?? "unknown",
            name: match?.[3] ?? "unknown",
            direction: match?.[4] ?? "unknown",
          };
        });

      // Group by version
      const grouped: Record<
        string,
        { up: string; down: string; module: string; name: string }
      > = {};
      for (const m of migrations) {
        if (!grouped[m.version]) {
          grouped[m.version] = {
            up: "",
            down: "",
            module: m.module,
            name: m.name,
          };
        }
        if (m.direction === "up") grouped[m.version].up = m.filename;
        if (m.direction === "down") grouped[m.version].down = m.filename;
      }

      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(
              {
                migrations_dir: MIGRATIONS_DIR,
                total_files: files.length,
                migrations: grouped,
              },
              null,
              2,
            ),
          },
        ],
      };
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      return {
        content: [
          {
            type: "text",
            text: `Migration scan error: ${message}. Make sure IMS_PROJECT_ROOT is set correctly.`,
          },
        ],
      };
    }
  },
);

// ===========================
// TOOL: project_api_routes
// ===========================
server.tool(
  "project_api_routes",
  "Scan the backend codebase for registered API routes by looking for chi router patterns in transport/router.go and module.go files.",
  {},
  async () => {
    try {
      const internalDir = path.join(BACKEND_ROOT, "internal");
      const entries = await fs.readdir(internalDir, { withFileTypes: true });
      const routes: Array<{ module: string; file: string; routes: string[] }> =
        [];

      for (const entry of entries) {
        if (!entry.isDirectory()) continue;
        const modPath = path.join(internalDir, entry.name);

        // Check module.go and transport/router.go
        const filesToCheck = [
          path.join(modPath, "module.go"),
          path.join(modPath, "transport", "router.go"),
        ];

        for (const filePath of filesToCheck) {
          try {
            const content = await fs.readFile(filePath, "utf-8");
            // Extract route patterns: r.Get, r.Post, r.Put, r.Delete, r.Route, r.Group
            const routePatterns = content.match(
              /r\.(Get|Post|Put|Delete|Patch|Route|Group|Handle|Method)\s*\(\s*"[^"]*"/g,
            );
            if (routePatterns && routePatterns.length > 0) {
              routes.push({
                module: entry.name,
                file: path.relative(BACKEND_ROOT, filePath),
                routes: routePatterns,
              });
            }
          } catch {
            // file doesn't exist — skip
          }
        }
      }

      return {
        content: [{ type: "text", text: JSON.stringify(routes, null, 2) }],
      };
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      return {
        content: [{ type: "text", text: `Route scan error: ${message}` }],
      };
    }
  },
);

// ---------------------------------------------------------------------------
// Resources: Expose key project files as MCP resources
// ---------------------------------------------------------------------------

server.resource("airead", "file:///AIREAD.md", async (uri) => {
  try {
    const content = await fs.readFile(
      path.join(PROJECT_ROOT, "AIREAD.md"),
      "utf-8",
    );
    return {
      contents: [{ uri: uri.href, text: content, mimeType: "text/markdown" }],
    };
  } catch {
    return { contents: [{ uri: uri.href, text: "AIREAD.md not found" }] };
  }
});

server.resource("claude-md", "file:///CLAUDE.md", async (uri) => {
  try {
    const content = await fs.readFile(
      path.join(PROJECT_ROOT, "CLAUDE.md"),
      "utf-8",
    );
    return {
      contents: [{ uri: uri.href, text: content, mimeType: "text/markdown" }],
    };
  } catch {
    return { contents: [{ uri: uri.href, text: "CLAUDE.md not found" }] };
  }
});

// ---------------------------------------------------------------------------
// Start
// ---------------------------------------------------------------------------

async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("[ims-mcp] IMS Thailand MCP server started (stdio transport)");
}

main().catch((err) => {
  console.error("[ims-mcp] Fatal:", err);
  process.exit(1);
});
