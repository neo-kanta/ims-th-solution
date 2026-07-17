# IMS Chat Assistant — Requirements Specification

| | |
|---|---|
| **Product** | IMS Chat Assistant (MCP-grounded financial chatbot) |
| **Parent system** | IMS Thailand Solution (`ims-th-solution`) |
| **Document status** | Draft v0.3 |
| **Last updated** | 2026-06-06 |
| **Maturity** | Proof of Concept (production-grade architecture) |
| **Owner** | _product owner TBD_ |
| **Author** | Engineering |

> **Requirement keywords** follow RFC 2119: **MUST**, **MUST NOT**, **SHOULD**, **MAY**.
> Each requirement carries a **priority** (`M`=Must / `S`=Should / `C`=Could) and a
> **status** (`Implemented` / `Partial` / `Planned`). Status reflects the build as of
> this document's date and is maintained in §15 (Traceability).

---

## 1. Purpose & Scope

### 1.1 Purpose
Provide a conversational assistant that lets authorised IMS users ask natural-language
questions about investment data — funds, portfolios, holdings, positions, NAV, prices,
balances — and receive answers that are **provably grounded in the IMS system of record**.
The assistant is a *read* layer over existing IMS data; it is not a new system of record.

### 1.2 In scope
- A chat experience (web UI + streaming API) embedded in the existing IMS application.
- A provider-agnostic gateway to multiple LLM vendors, selectable by configuration.
- Access to IMS data **exclusively** through the Model Context Protocol (MCP).
- Financial-accuracy controls, permission enforcement, audit, and provenance.

### 1.3 Out of scope
- Building or modifying IMS business logic (accounting, OMS, PAM, valuation engines).
- Allowing the assistant to mutate IMS state in normal operation (write is gated off).
- Replacing the dashboard, reporting, or any authoritative financial workflow.

---

## 2. Background & Context

The IMS Thailand Solution is a Go modular-monolith backend with a Nuxt 4 frontend,
PostgreSQL, and Redis. It already owns the authoritative investment data and the
permission/audit machinery. The Chat Assistant adds a **conversational read interface**
on top, implemented as a new `chat` module that reaches business data only via MCP and
reuses the platform's IAM and audit capabilities through their existing contracts.

The defining requirement is **trust**: in a financial setting an answer that contains a
fabricated or miscalculated figure is worse than no answer. Every control in this
document exists to guarantee that displayed figures originate from the IMS and can be
traced back to their source.

---

## 3. Actors & Stakeholders

| Actor | Description |
|---|---|
| **Authenticated IMS user** | Fund manager, analyst, operations staff. Asks questions; sees only data their IAM permissions allow. |
| **Administrator** | Configures the active LLM provider, the write-enabled flag, and MCP servers. |
| **Auditor / Compliance** | Reviews the immutable audit trail of every chat turn and tool call. |
| **Platform operator** | Runs and monitors the stack (containers, health, logs). |
| **IMS REST API** | The authoritative data source the MCP tools call. |
| **LLM provider** | External model service (Anthropic / OpenAI / Google / OpenAI-compatible). |
| **MCP server(s)** | Process(es) exposing read-only IMS tools to the assistant. |

---

## 4. Glossary

| Term | Meaning |
|---|---|
| **LLM** | Large Language Model. |
| **Provider** | A specific LLM vendor/endpoint behind the gateway. |
| **MCP** | Model Context Protocol — the standard by which the backend discovers and calls external tools. |
| **MCP client** | The component (inside the `chat` module) that connects to MCP servers. |
| **MCP server** | A process exposing tools; here, read-only wrappers over IMS data. |
| **Tool** | A callable function exposed via MCP (e.g. `get_portfolio_holdings`). |
| **Tool call / tool result** | The model's request to run a tool, and the data returned. |
| **Agent loop** | The server-side cycle: model → tool calls → results → model → final answer. |
| **Provenance** | The stored link from a displayed figure to the tool call + raw result that produced it. |
| **Turn** | One user message and the assistant's full response (including any tool calls). |
| **SSE** | Server-Sent Events — the streaming response format. |

---

## 5. Goals & Non-Goals

### 5.1 Goals
- **G1** — Answer questions about IMS data accurately and in plain language.
- **G2** — Be model-agnostic: switch LLM vendor by configuration only.
- **G3** — Reach all business data through MCP, never via direct cross-module coupling.
- **G4** — Make financial accuracy a structural guarantee, not a matter of trusting the model.
- **G5** — Enforce permissions server-side and record an immutable audit trail.

### 5.2 Non-Goals
- **NG1** — The assistant is not a trading or approval interface.
- **NG2** — The assistant does not compute or derive financial figures itself.
- **NG3** — The assistant is not a general-purpose web search or market-data oracle.

---

## 6. Hard Constraints (Non-Negotiable)

These are binding invariants. A change that violates one is a defect, not a trade-off.

- **HC1 (Accuracy)** — Any numeric financial value (price, quantity, NAV, P&L, balance,
  ratio, transaction date) **MUST** originate from an MCP tool result and be passed
  through verbatim. The model **MUST NOT** fabricate or estimate figures.
- **HC2 (No model arithmetic on money)** — Calculations on monetary values **MUST** be
  performed in code or by a dedicated tool, never in the model's free-text reasoning.
- **HC3 (Provenance)** — Every displayed figure **MUST** be traceable to the tool call,
  arguments, and raw result that produced it, and that trace **MUST** be stored.
- **HC4 (MCP-only data access)** — The `chat` module **MUST** reach all business data only
  through MCP. It **MUST NOT** import another business module's internal packages. Only
  the IAM and audit contracts may be used in-process.
- **HC5 (Read-only by default)** — Only read/query tools are exposed. Mutating tools
  **MUST NOT** execute unless an explicit write-enabled flag is set **AND** the user passes
  a server-side permission check.
- **HC6 (Untrusted tool output)** — Tool results **MUST** be treated as data, never as
  instructions; embedded instructions **MUST NOT** be able to escalate to a write tool.
- **HC7 (Server-side enforcement)** — Permission checks, tool gating, and accuracy
  controls **MUST** run on the backend. Frontend controls are UX only.
- **HC8 (Determinism where it matters)** — The orchestration model **MUST** run at low/zero
  temperature, and tool arguments **MUST** be validated against the tool's input schema.
- **HC9 (Currency-aware money)** — Monetary values **MUST** carry a currency and **MUST**
  cross the MCP boundary as strings (never JSON numbers, to avoid float precision loss).

---

## 7. Functional Requirements

### 7.1 Model Gateway (multi-provider)

| ID | Priority | Requirement | Status |
|---|---|---|---|
| FR-GW-01 | M | The system **MUST** define a single provider interface that every LLM adapter implements; the agent loop **MUST** be written once against it with no vendor-specific code. | Implemented |
| FR-GW-02 | M | The active provider **MUST** be selectable at runtime by configuration (`LLM_PROVIDER`) with no other code change. | Partial |
| FR-GW-03 | M | The system **MUST** support Anthropic (Claude). | Implemented |
| FR-GW-04 | S | The system **SHOULD** support OpenAI, Google Gemini, and any OpenAI-compatible endpoint (e.g. Ollama/OpenRouter) behind the same interface. | Planned |
| FR-GW-05 | M | Each adapter **MUST** translate the canonical tool definitions into its own function-calling schema, and translate the model's tool-call requests back into the canonical form. | Implemented (Anthropic) |
| FR-GW-06 | M | Calls to providers **MUST** be stateless; the backend **MUST** own conversation history and resend it each turn. | Implemented |
| FR-GW-07 | S | A turn **MAY** override the model/provider per request; overrides apply to the session. | Partial |

### 7.2 MCP Integration & Tools

| ID | Priority | Requirement | Status |
|---|---|---|---|
| FR-MCP-01 | M | The backend **MUST** host an MCP client that connects to one or more MCP servers and discovers their tools. | Implemented |
| FR-MCP-02 | M | The client **MUST** support local stdio MCP servers. | Implemented |
| FR-MCP-03 | S | The client **SHOULD** support remote MCP servers over Streamable HTTP / SSE. | Planned |
| FR-MCP-04 | M | MCP servers and tool allow/deny rules **MUST** be configurable (file or built-in default). | Implemented |
| FR-MCP-05 | M | The system **MUST** provide read-only IMS tools covering at least: list funds, list portfolios, portfolio holdings, latest fund NAV. | Implemented |
| FR-MCP-06 | M | Each IMS tool **MUST** obtain data from the existing IMS REST API using the **calling user's identity**, so existing permission and data-scoping apply and figures match the dashboard. | Implemented |
| FR-MCP-07 | M | Tool results **MUST** include provenance metadata (tool name, endpoint, as-of timestamp). | Implemented |
| FR-MCP-08 | M | A tool that errors or has no data **MUST** return a structured error/empty result (not a fabricated success). | Implemented |
| FR-MCP-09 | M | A read-only allowlist **MUST** be enforced by the client independently of what a server exposes (deny patterns win). | Implemented |

### 7.3 Conversation, Streaming & Sessions

| ID | Priority | Requirement | Status |
|---|---|---|---|
| FR-CV-01 | M | A user **MUST** be able to send a message and receive a streamed, incremental response. | Implemented |
| FR-CV-02 | M | Streaming **MUST** use a transport that carries the bearer token and a POST body (i.e. `fetch`+`ReadableStream`, not native `EventSource`). | Implemented |
| FR-CV-03 | M | Conversations **MUST** be persisted as sessions and messages; a new message **MAY** continue an existing session. | Implemented |
| FR-CV-04 | M | The system **MUST** cap resent history (oldest-message trim) to bound context size and cost. | Implemented |
| FR-CV-05 | M | The agent loop **MUST** be bounded by a maximum number of tool/Generate steps per turn. | Implemented |
| FR-CV-06 | S | The UI **SHOULD** surface tool activity ("calling get_portfolio_holdings…") and a per-figure source panel. | Planned |

### 7.4 Financial Accuracy

| ID | Priority | Requirement | Status |
|---|---|---|---|
| FR-FA-01 | M | The system prompt **MUST** instruct the model to fetch data via tools and never invent or compute figures. | Implemented |
| FR-FA-02 | M | The model **MUST** only receive figures inside tool results it explicitly requested; raw results **MUST** be fed back verbatim. | Implemented |
| FR-FA-03 | M | Tool raw results **MUST** be stored as the provenance source of record for the turn. | Implemented |
| FR-FA-04 | M | When tools cannot supply the requested data, the assistant **MUST** state so plainly and **MUST NOT** produce a number. | Implemented |
| FR-FA-05 | S | A post-generation guardrail **SHOULD** verify that numeric financial figures in the final answer trace to a tool result for the turn, and block/redact otherwise. | Planned |
| FR-FA-06 | S | Aggregations (sums, weighted averages, exposure %) **SHOULD** be produced by code/dedicated tools, surfaced with their inputs. | Planned |

### 7.5 Security, Permissions & Write Policy

| ID | Priority | Requirement | Status |
|---|---|---|---|
| FR-SEC-01 | M | The chat endpoint **MUST** require authentication. | Implemented |
| FR-SEC-02 | M | Before executing any tool, the system **MUST** perform a server-side permission check for the user; denial **MUST** prevent execution. | Implemented |
| FR-SEC-03 | M | Data-level scoping (which funds/portfolios a user may see) **MUST** be enforced by the authoritative endpoints the tools call. | Implemented |
| FR-SEC-04 | M | Mutating tools **MUST NOT** execute unless a write-enabled flag is set (default off) **AND** the user passes the permission check. | Implemented |
| FR-SEC-05 | M | Tool output **MUST** be treated as untrusted data; it **MUST NOT** be able to trigger a write tool. | Implemented (read-only) / Partial (injection neutralisation) |
| FR-SEC-06 | M | The user's bearer token forwarded to tools **MUST NOT** be exposed to the model nor written to audit/provenance. | Implemented |
| FR-SEC-07 | M | User-facing errors **MUST** be safe (no stack traces, keys, or raw provider errors). | Implemented |

### 7.6 Audit & Provenance

| ID | Priority | Requirement | Status |
|---|---|---|---|
| FR-AUD-01 | M | The system **MUST** record an immutable audit entry for the start and completion of every turn. | Implemented |
| FR-AUD-02 | M | The system **MUST** record an immutable audit entry for every tool call, including outcome (succeeded/failed/denied). | Implemented |
| FR-AUD-03 | M | Every audit entry **MUST** carry correlation references (session id, message id, and tool-invocation id where relevant) so an auditor can pivot from the central trail to the chat detail tables. | Implemented |
| FR-AUD-04 | M | Tool calls, arguments, and raw results **MUST** be persisted (append-only) as the provenance detail store. | Implemented |
| FR-AUD-05 | M | Chat detail stores **MUST NOT** be orphaned from the central audit trail. | Implemented |

### 7.7 Frontend Chat Experience

| ID | Priority | Requirement | Status |
|---|---|---|---|
| FR-UI-01 | M | A chat page **MUST** be reachable from the authenticated app navigation. | Implemented |
| FR-UI-02 | M | The UI **MUST** render the streamed answer incrementally and indicate progress. | Implemented |
| FR-UI-03 | M | The UI **MUST** use the generated typed API client (no hand-written request/response types). | Implemented |
| FR-UI-04 | M | User-facing copy **MUST** be available in EN/TH/ZH via the i18n system. | Implemented |
| FR-UI-05 | S | The UI **SHOULD** display a tool-call trace and an expandable per-figure "source" panel reading the provenance map. | Planned |

---

## 8. Non-Functional Requirements

| ID | Priority | Requirement | Status |
|---|---|---|---|
| NFR-01 | M | **Configuration**: API keys, provider selection, write flag, and MCP servers **MUST** come from environment/config; no secrets in code or committed files. | Implemented |
| NFR-02 | M | **Observability**: structured JSON logging for requests, tool calls, and errors. | Implemented |
| NFR-03 | M | **Resilience**: a misconfigured provider or unreachable MCP server **MUST** degrade gracefully (chat disabled / no-tools) without taking down the rest of the API. | Implemented |
| NFR-04 | M | **Deployment**: the full stack (backend, frontend, Postgres, Redis, MCP server) **MUST** come up via `docker compose`. | Implemented |
| NFR-05 | S | **Testability**: unit tests for the agent loop, permission/ write gating, and provider tool translation; one integration test for a full grounded round trip. | Partial |
| NFR-06 | M | **Data hygiene**: timestamps stored UTC; money never `float64`; currency-aware money type. | Implemented |
| NFR-07 | S | **Health**: liveness/readiness endpoints for the backend (and ideally MCP). | Partial |
| NFR-08 | M | **Boundary enforcement**: an automated check (arch-lint/test) **MUST** fail the build if the chat module imports another business module's internals. | Implemented |

---

## 9. Data Model (chat-owned)

All tables are append-only where noted; UPDATE is blocked by trigger; DELETE only via
session cascade. The central `iam_audit_events` table retains the audit spine independently.

- **`chat_sessions`** — `id, user_id, provider, model, created_at, updated_at`.
- **`chat_messages`** — `id, session_id, role(system|user|assistant|tool), content,
  provenance_map JSONB, raw_provider_payload JSONB, created_at`. *(append-only)*
- **`chat_tool_invocations`** — `id, session_id, message_id, tool_call_id, tool_name,
  mcp_server, arguments JSONB, raw_result JSONB, is_error, state(requested|succeeded|
  failed|denied), error, started_at, finished_at`. *(append-only; provenance source of record)*

Audit reuses the existing immutable `iam_audit_events` table; chat emits
`CHAT_TURN_STARTED`, `CHAT_TURN_COMPLETED`, `CHAT_TURN_FAILED`, and `CHAT_TOOL_CALL`
events with correlation metadata.

---

## 10. External Interface — `POST /api/v1/chat`

- **Auth**: Bearer JWT (required).
- **Request body**: `{ content (required), session_id?, model?, provider? }`.
- **Response**: `Content-Type: text/event-stream`. Each frame's `data:` payload is a
  `ChatStreamEvent` with a `kind` discriminator:
  - `session_started` → `session_id`, `message_id`
  - `text` → incremental `text`
  - `tool_call` → `tool_name`, `tool_status` (`running|ok|error|denied`)
  - `done` → `session_id`, `message_id`, `stop_reason`
  - `error` → safe `error` message
- The contract is published in Swagger/OpenAPI and consumed via the generated typed
  client; it **MUST NOT** be hand-maintained on the frontend.

---

## 11. Architecture & Module Placement

```
Nuxt UI ──fetch/ReadableStream (SSE)──> Go backend `chat` module
                                          ├─ application/ : agent loop, ports
                                          ├─ provider/    : LLM adapters (ChatProvider)
                                          ├─ mcp/         : MCP client (stdio/http)
                                          ├─ persistence/ : sessions, messages, tool invocations
                                          └─ adapters     : audit (audit contract), permission gate (IAM)
                                                  │ MCP (stdio)
                                                  ▼
                                          IMS MCP server (read-only)
                                                  │ REST + caller's JWT
                                                  ▼
                                          Existing IMS REST API ──> PostgreSQL
```

- The `chat` module follows the repo's modular-monolith placement rules.
- The MCP server is a **separate process** spawned over stdio; the reference
  implementation is the Go binary `backend/cmd/ims-mcp`. Its tools call the IMS REST API
  with the user's token — keeping the chat module free of business-module imports while
  guaranteeing identical, permission-scoped figures.

---

## 12. Configuration (environment)

| Variable | Purpose | Default |
|---|---|---|
| `LLM_PROVIDER` | Active provider | `anthropic` |
| `ANTHROPIC_API_KEY` | Provider credential (required for Anthropic) | — |
| `ANTHROPIC_MODEL` | Model id | `claude-haiku-4-5-…` |
| `CHAT_MAX_TOKENS_PER_TURN` | Output budget per turn | `1024` |
| `CHAT_WRITE_ENABLED` | Enable mutating tools | `false` |
| `CHAT_MCP_SERVERS_CONFIG` | Path to `mcp-servers.yaml` (optional) | — (built-in default) |
| `CHAT_MCP_IMS_BIN` | Path to the IMS MCP binary | next to server binary |
| `IMS_API_BASE_URL` | REST base URL the MCP tools call | `http://localhost:8080/api/v1` |

Secrets **MUST** be provided via gitignored files (`backend/.env`, `infra/.env`) or the
deployment secret store — never committed.

---

## 13. Acceptance Criteria

| ID | Criterion | Status |
|---|---|---|
| AC-1 | Switching `LLM_PROVIDER` changes the active model with no other code change. | Partial (Anthropic live; others pending FR-GW-04) |
| AC-2 | The same MCP tool works across all configured providers (verified by translation test). | Partial |
| AC-3 | A question whose data the tools can't supply yields an honest "I don't have that," never a fabricated number. | **Met** |
| AC-4 | Every displayed figure is traceable to a logged tool call + raw result. | **Met** (stored; UI surfacing pending FR-UI-05) |
| AC-5 | No write tool executes unless the write flag is set **and** the user passes a permission check. | **Met** |
| AC-6 | `docker compose up` brings up backend + frontend + Postgres (+ Redis + MCP); chat works end-to-end with ≥1 MCP server. | **Met** |

---

## 14. Assumptions & Key Decisions

**Assumptions**
- A1 — Chat is a module inside the existing monolith, not a standalone app.
- A2 — Audit and permissions are reused via existing contracts, not re-implemented.
- A3 — History is resent each turn with a simple count cap; full compaction is later.
- A4 — Money is currency-aware and crosses MCP as strings.

**Decisions (confirmed)**
- D1 — **Data path**: MCP tools reuse the existing REST API with the caller's identity
  (vs. direct SQL) — chosen for accuracy parity and permission reuse.
- D2 — **Topology**: the IMS MCP server runs as a local stdio child process (vs. a
  standalone HTTP service) for this phase.
- D3 — **Streaming transport**: `fetch`+`ReadableStream`, not `EventSource`.
- D4 — **Audit linkage**: split storage (central spine + chat detail) joined by
  correlation ids.
- D5 — **Boundary**: enforced by an automated arch-lint test.

---

## 15. Traceability — Implementation Status by Phase

| Phase | Theme | Status |
|---|---|---|
| **Slice A** | Streaming chat skeleton: UI + SSE API + Anthropic adapter + audit seam | ✅ Complete |
| **Slice B** | Provider gateway: OpenAI / Gemini / OpenAI-compatible behind one interface | ⏳ Interface + Anthropic done; other adapters pending |
| **Slice C** | MCP client + tool registry + agent loop + read-only IMS MCP server + permissions + audit + provenance | ✅ Complete |
| **Slice D** | Guardrails (figure-level accuracy validation + injection neutralisation) | ⏳ Not started (plumbing in place) |
| **UI polish** | Tool-call trace + per-figure source panel | ⏳ Not started |

---

## 16. Out of Scope & Future Work

- Additional LLM provider adapters (FR-GW-04) and cross-provider translation tests (AC-2).
- Post-generation financial-accuracy guardrail and injection neutralisation (FR-FA-05, FR-SEC-05).
- UI tool-trace and source panel (FR-UI-05).
- Remote (HTTP/SSE) MCP servers (FR-MCP-03).
- Computation tools for aggregations (FR-FA-06).
- Token-aware history compaction (beyond the count cap).
- Write-enabled tool catalog and the approval flow that would gate it.

---

## 17. Risks & Mitigations

| Risk | Mitigation |
|---|---|
| Model fabricates a figure | Tools are the only source of figures; verbatim passthrough; provenance stored; honest-refusal verified; planned figure-level guardrail (FR-FA-05). |
| Prompt injection via tool data | Tool output treated as data; read-only by default; write double-gated; planned neutralisation (FR-SEC-05). |
| Permission bypass | Server-side gate before execution **and** authoritative REST scoping with the user's token. |
| Secret leakage | Keys in gitignored env/secret store; token never shown to model or audited. |
| Provider/MCP outage | Graceful degradation; the rest of the API stays up. |
| Cost / context blow-up | Zero temperature, max-tokens cap, history cap, bounded agent steps. |
| Cross-module coupling erosion | Automated boundary arch-lint test in CI (NFR-08). |

---

_End of document._
