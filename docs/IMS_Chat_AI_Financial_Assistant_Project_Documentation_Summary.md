# IMS Chat AI Financial Assistant — Documentation Summary

**Parent system:** IMS Thailand Solution (`ims-th-solution`)
**Scope:** Chat module / AI Financial Assistant only
**Version:** v1.0 · **Status:** Existing Code Review + Documentation · **Date:** 6 June 2026
**Companion file:** `IMS_Chat_AI_Financial_Assistant_Project_Documentation.docx`

> Confidential — Internal Use Only. This summary reflects the code as it exists today. No production-readiness claim is made beyond what is evidenced from the source.

---

## What it is

A read-oriented, permission-aware, auditable **natural-language window onto IMS investment data** — funds, portfolios, holdings, positions, NAV, prices, balances. It is **not** a system of record, trading bot, robo-advisor, accounting/valuation/compliance engine, or recommendation engine.

**Core principle:** AI explains and summarizes; MCP tools + the IMS REST API provide the facts; backend code enforces permission, provenance, audit, and validation. The model must not invent, estimate, or freely calculate financial figures.

---

## Architecture (verified)

```
User → Frontend Chat UI (Nuxt/Vue)
     → Backend Chat API (ChatHandler, SSE)
         → Agent Loop (SendMessage.Run, temp=0, ≤6 steps)
             → LLM Provider (Anthropic adapter)        [text + tool requests]
             → MCP Client (ToolGateway)
                 → ims-mcp server (read-only tools)
                     → IMS REST API (/api/v1/investment/*)
                         → System-of-record data (PostgreSQL)
```

- The frontend never talks to the model or MCP directly.
- The model never touches IMS data directly — it only ever sees figures **inside tool results it requested**.
- Every tool call runs with the **user's own JWT**, forwarded via MCP `_meta` (`ims_auth_token`), never shown to the model and never written to audit.

The .docx contains 7 Mermaid diagrams: high-level architecture, message lifecycle, agent loop, MCP tool-call flow, provenance flow, permission+audit flow, and frontend streaming flow.

---

## What is genuinely implemented (strengths)

- **Read-only by construction** — `ims-mcp` registers only GET wrappers; defense-in-depth allow (`list_*`, `get_*`) / deny (`*create*`, `*update*`, …) globs; `CHAT_WRITE_ENABLED` defaults **false**.
- **Two pre-execution gates** — write-policy gate + fail-closed permission gate (`INVESTMENT_FUND_VIEW` / `INVESTMENT_PORTFOLIO_VIEW`), then the REST endpoint re-checks under the user's token (row-level data scope).
- **Provenance store** — verbatim tool output persisted in `chat_tool_invocations.raw_result`; the MCP layer never reshapes/rounds numbers and stamps `tool/endpoint/http_status/as_of`.
- **Strict audit** — `CHAT_TURN_STARTED` / `CHAT_TOOL_CALL` / `CHAT_TURN_COMPLETED` / `CHAT_TURN_FAILED`; audit failure fails the turn. Central audit is the spine; chat tables are the detail.
- **Append-only persistence** — `chat_sessions`, `chat_messages`, `chat_tool_invocations`; UPDATE blocked by DB triggers.
- **Determinism** — temperature `0`; bounded loop (6 steps); history cap (40 msgs); 16 KiB content cap.
- **Provider-agnostic seam** — single `ChatProvider` interface; build-enforced **MCP-only module boundary** (`boundary_test.go`).
- **Streaming UX** — SSE backend + `fetch()`/`ReadableStream` frontend (so the JWT can be sent); typed via generated OpenAPI client.
- **Secrets posture** — keys live in **gitignored** `infra/.env`; committed `.env.example` files keep keys empty.

---

## Production gaps (honest)

| # | Gap | Severity |
|---|-----|----------|
| 1 | **No post-generation numeric validator** — nothing checks that answer figures trace to a tool result | **High** (top blocker) |
| 2 | **No calculation/aggregation tools** — multi-value math would fall to the model (disallowed but not prevented) | Medium |
| 3 | **No per-figure provenance / source panel UI** — frontend even ignores the `tool_call` SSE frames the backend sends | Medium |
| 4 | **Multi-provider partial** — only Anthropic implemented (OpenAI/Gemini/OpenAI-compatible stubbed) | Medium |
| 5 | **Prompt-injection defense partial** — tool output is structurally data, but no sanitisation layer and no regression tests | High |
| 6 | **Observability partial** — no chat/MCP health probe or metrics | Medium |
| 7 | **Testing partial** — only the agent loop + module boundary are tested | High |
| 8 | **Session history not implemented** — no server-backed list/resume (`GET /chat/sessions*` is a code comment only) | Low |
| 9 | `raw_provider_payload` always NULL; chat i18n copy stale ("Slice A / no tool access"); downstream money typing unverified | Low |

---

## Chat data model (3 append-only tables)

- **`chat_sessions`** — id, user_id→`iam_users`, provider, model, timestamps (mutable).
- **`chat_messages`** — id, session_id (cascade), role (`system|user|assistant|tool`), content, `provenance_map` (tool-call pointers), `raw_provider_payload` (NULL today), created_at.
- **`chat_tool_invocations`** — id, session_id, message_id, tool_call_id, tool_name, mcp_server, arguments (no token), `raw_result` (verbatim provenance), is_error/state/error, started/finished_at.

## Chat API

| Method | Path | Status |
|--------|------|--------|
| POST | `/api/v1/chat` | Implemented (SSE: `session_started`, `text`, `tool_call`, `done`, `error`) |
| GET | `/api/v1/chat/sessions` | Planned (comment only) |
| GET | `/api/v1/chat/sessions/{id}/messages` | Planned (comment only) |

## MCP tools (read-only)

| Tool | REST endpoint | Gate permission |
|------|---------------|-----------------|
| `list_funds` | `GET /investment/funds` | `INVESTMENT_FUND_VIEW` |
| `get_fund_nav` | `GET /investment/funds/{id}/nav/latest` | `INVESTMENT_FUND_VIEW` |
| `list_portfolios` | `GET /investment/portfolios` | `INVESTMENT_PORTFOLIO_VIEW` |
| `get_portfolio_holdings` | `GET /investment/portfolios/{id}/holdings` | `INVESTMENT_PORTFOLIO_VIEW` |

---

## Executive verdict

- **Credible AI financial assistant project?** Yes — the finance-appropriate architecture (grounding, isolation, gates, audit, temp 0) is real in code, not just prose.
- **Production-ready?** **Not yet** for unsupervised financial production; strong PoC/early-beta foundation.
- **Strongest point:** structural grounding + isolation (model can't reach IMS data directly; figures only via tool results; writes off by default; build-enforced boundary).
- **Biggest risk:** hallucinated/mis-stated figures in the final answer — unlikely given grounding, but not yet *structurally impossible* without a numeric validator.
- **Must finish before production:** (1) mandatory post-generation numeric validator, (2) deterministic calculation tools, (3) per-figure provenance UI, (4) prompt-injection suite + tool-output sanitisation, (5) observability + broad test matrix, (6) data retention/governance policy.

---

## Recommended next coding tasks (priority order)

1. **Post-generation numeric validator** — block/redact any answer figure not traceable to a current-turn tool result or deterministic calculation.
2. **Deterministic calculation/aggregation tools** — totals, weighted averages, P&L; forbid model arithmetic on money.
3. **Per-figure provenance + UI source panel** — bind numbers to result fields; render the `tool_call` frames the backend already emits.
4. **Prompt-injection regression suite + tool-output sanitisation** — user-input and tool-output vectors.
5. **Observability** — chat/MCP health & readiness probes, metrics (tool latency, step counts, denials), per-user cost budgets.
6. **Broaden tests** — MCP tools, Anthropic stream parser, repositories, SSE handler, audit-on-every-path, frontend components, golden no-hallucination dataset.
7. **Housekeeping** — implement session-history endpoints; refresh stale i18n copy; thread a correlation id into chat audit; persist `raw_provider_payload`; verify lossless money typing in the REST layer.
