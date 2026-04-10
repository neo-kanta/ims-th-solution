# IMS Thailand MCP Server

This package contains the local Model Context Protocol server for IMS Thailand.

It is intended for AI-assisted development workflows that need read access to:

- the IMS PostgreSQL schema and data
- migration history
- workflow definitions
- permission models
- project structure and module metadata

## What It Exposes

The server is designed to give an AI client enough project context to answer implementation questions safely without directly editing your runtime system.

Examples of the exposed capabilities include:

- database schema inspection
- read-only SQL queries
- migration inspection
- workflow-state inspection and validation
- permission and group lookups
- audit-log queries
- project module and route scanning

## Package Layout

```text
tools/mcp/ims-mcp-server/
|-- src/
|   `-- index.ts
|-- build/                # Generated output after npm run build
|-- package.json
|-- tsconfig.json
`-- README.md
```

## Requirements

- Node.js 20+
- Access to the IMS development database

## Install

```bash
cd tools/mcp/ims-mcp-server
npm install
```

## Build

```bash
npm run build
```

The compiled entrypoint will be:

```text
build/index.js
```

## Run Locally

```bash
npm start
```

## Environment Variables

| Variable | Default | Purpose |
| --- | --- | --- |
| `IMS_DB_HOST` | `localhost` | PostgreSQL host |
| `IMS_DB_PORT` | `5437` | PostgreSQL port |
| `IMS_DB_NAME` | `ims_dev` | Database name |
| `IMS_DB_USER` | `ims_app` | Database user |
| `IMS_DB_PASSWORD` | `ims_dev_password` | Database password |
| `IMS_PROJECT_ROOT` | `process.cwd()` | Repository root for filesystem scans |

## Example MCP Configuration

Place this in your local MCP client config, adjusting paths as needed:

```json
{
  "mcpServers": {
    "ims": {
      "command": "node",
      "args": ["tools/mcp/ims-mcp-server/build/index.js"],
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

## Local Setup Flow

From the repository root:

```bash
make dev
make migrate-up
```

Then build the MCP server:

```bash
cd tools/mcp/ims-mcp-server
npm run build
```

## Security Notes

- database query support should remain read-only
- do not point this tool at a production database unless you explicitly intend to
- keep local MCP config files out of version control if they contain credentials

## Troubleshooting

- Cannot connect to the database:
  - make sure PostgreSQL is running
  - verify host, port, user, password, and database name
- No migration data:
  - run `make migrate-up`
- MCP client cannot see the server:
  - confirm the `build/index.js` path is correct
  - restart the MCP client after updating its config
