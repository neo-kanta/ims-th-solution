# IMS Thailand — MCP Server

A Model Context Protocol (MCP) server that gives Claude Code / Claude.ai live access to your IMS project's database, workflow state machine, permission model, domain rules, and codebase structure.

## What This Gives You

| Tool                 | What It Does                                                           |
| -------------------- | ---------------------------------------------------------------------- |
| `db_query`           | Run read-only SQL against your IMS PostgreSQL database                 |
| `db_schema`          | List all tables, columns, types, constraints                           |
| `db_table_info`      | Deep-dive a single table (indexes, FKs, row count, comments)           |
| `db_migrations`      | Show which migrations have been applied                                |
| `workflow_states`    | Show the full workflow state machine with transitions                  |
| `workflow_validate`  | Check if a state transition is valid (forward, rollback, or invalid)   |
| `perm_user_rights`   | Show all permissions for a user (groups, function rights, data rights) |
| `perm_check`         | Check if a user has a specific permission code                         |
| `perm_groups`        | List all groups with members and rights                                |
| `domain_rules`       | List IMS business validation rules by module                           |
| `audit_log`          | Query the audit trail                                                  |
| `project_modules`    | Scan backend modules and their implementation status                   |
| `project_migrations` | List migration files from the filesystem                               |
| `project_api_routes` | Scan registered API routes from the codebase                           |

Plus **resources** that expose `AIREAD.md` and `CLAUDE.md` for context.

## Prerequisites

- Node.js 20+ (LTS)
- Your IMS PostgreSQL database running (via `docker compose up -d postgres`)
- Claude Code installed (`npm install -g @anthropic-ai/claude-code`)

## Step-by-Step Setup

### Step 1: Install dependencies

```bash
cd mcp/ims-mcp-server
npm install
```

### Step 2: Build the server

```bash
npm run build
```

### Step 3: Verify it compiles

```bash
ls -la build/index.js
# Should exist and be executable
```

### Step 4: Make sure your database is running

```bash
cd ../../infra
docker compose up -d postgres
# Wait a few seconds
make migrate-up   # from project root
```

### Step 5: Configure Claude Code

Create or edit `.mcp.json` in your **IMS project root**:

```json
{
  "mcpServers": {
    "ims": {
      "command": "node",
      "args": ["mcp/ims-mcp-server/build/index.js"],
      "env": {
        "IMS_DB_HOST": "localhost",
        "IMS_DB_PORT": "5437",
        "IMS_DB_NAME": "ims_dev",
        "IMS_DB_USER": "ims_app",
        "IMS_DB_PASSWORD": "ims_dev_password",
        "IMS_PROJECT_ROOT": "."
      }
    }
  }
}
```

### Step 6: Restart Claude Code

Claude Code reads `.mcp.json` on startup. After adding the config:

```bash
# If using Claude Code CLI:
claude

# It should show "ims" as a connected MCP server
```

### Step 7: Test it

In Claude Code, try:

- "Use the ims MCP to show me the database schema"
- "What workflow states exist and what transitions are valid?"
- "Check if the admin user has any permissions"
- "List all domain rules for the stock_investment module"
- "What backend modules have TODO markers?"

## Configuration Reference

| Environment Variable | Default            | Description              |
| -------------------- | ------------------ | ------------------------ |
| `IMS_DB_HOST`        | `localhost`        | PostgreSQL host          |
| `IMS_DB_PORT`        | `5437`             | PostgreSQL port          |
| `IMS_DB_NAME`        | `ims_dev`          | Database name            |
| `IMS_DB_USER`        | `ims_app`          | Database user            |
| `IMS_DB_PASSWORD`    | `ims_dev_password` | Database password        |
| `IMS_PROJECT_ROOT`   | `process.cwd()`    | Path to IMS project root |

## Security Notes

- **Read-only by design.** The `db_query` tool only allows SELECT/WITH/EXPLAIN. All queries run inside a read-only transaction.
- **Connection pool capped at 5.** Won't overwhelm your dev database.
- **No writes to the database** from any tool.
- **Credentials stay in `.mcp.json`** which should be in `.gitignore`.

## Adding to .gitignore

Add this to your project's `.gitignore`:

```
# MCP server build artifacts
mcp/ims-mcp-server/build/
mcp/ims-mcp-server/node_modules/

# MCP config contains credentials
.mcp.json
```

## Architecture

```
mcp/ims-mcp-server/
├── src/
│   └── index.ts          # All tools, resources, and server setup
├── build/                 # Compiled JS (gitignored)
├── package.json
├── tsconfig.json
└── README.md
```

Single-file design is intentional — MCP servers should be simple, focused, and easy to audit. If tools grow beyond ~20, split into `src/tools/db.ts`, `src/tools/workflow.ts`, etc.

## Extending

To add a new tool:

```typescript
server.tool(
  "my_new_tool", // tool name
  "Description for Claude to understand when to use this", // description
  {
    param1: z.string().describe("What this param does"),
  },
  async ({ param1 }) => {
    // Your logic here
    return {
      content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
    };
  },
);
```

## Troubleshooting

**"Cannot connect to database"**
→ Make sure PostgreSQL is running: `docker compose up -d postgres`
→ Check the port matches (default 5437)

**"Migration table not found"**
→ Run `make migrate-up` from the project root first

**"MCP server not showing in Claude Code"**
→ Check `.mcp.json` is in the project root (where you run `claude`)
→ Restart Claude Code after adding/changing `.mcp.json`

**"Permission denied on build/index.js"**
→ Run `chmod 755 build/index.js` or rebuild with `npm run build`
