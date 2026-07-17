# IMS Chat AI Financial Assistant — User & Operator Guide

A plain-language guide for people who are new to this assistant. It explains what
it does, how to use it, **why an answer sometimes gets blocked**, and how to get
it fully working in a local/dev environment.

> The Chat Assistant is **read-only** and **grounded**: it can look things up in
> IMS, but it cannot change anything, and it will refuse to show a financial
> number it cannot trace back to real IMS data.

---

## 1. What it is (in one minute)

- It is a chat box that answers questions about **your** IMS data — funds,
  portfolios, holdings, NAV, valuations, balances.
- It does **not** make decisions, trade, or change data. It only reads.
- Every financial number it shows must come from an IMS tool result. If it can't
  verify a number, it **won't show the answer** — by design (see §4).

Open it from the sidebar → **Chat**, or go to `/chat`.

---

## 2. How to use it

1. Type a question and press **Enter** (Shift+Enter for a new line).
2. Watch the **tool chips** appear while it works (e.g. `get_fund_nav running → ok`).
3. Read the answer. Open **“Sources & provenance”** under the answer to see which
   tools produced the data and which figure came from which tool.
4. Use **New chat** (top of the history rail) to start fresh, or click any past
   conversation in the rail to **resume** it.
5. **Stop** appears while an answer is streaming if you want to cancel it.

### Good questions (these return verifiable numbers)
- “List the funds I can access.”
- “Show my portfolios.”
- “What is the latest NAV for fund *&lt;name or id&gt;*?”
- “Show the latest valuation for portfolio *&lt;id&gt;*.”
- “What is the weighted average price / allocation / exposure for portfolio *&lt;id&gt;*?”

> Tip: name a specific fund or portfolio. The assistant first lists, then looks
> up details by id — naming the thing helps it fetch the right figure.

---

## 3. Reading an answer

| Element | Meaning |
|---|---|
| **Tool chips** (`get_fund_nav` ok/error/denied) | Which IMS tools ran this turn and how each finished. |
| **“Figures verified”** chip | Every number in the answer was traced to a tool result. |
| **“Some figures could not be verified”** chip | The answer was blocked (see §4). |
| **Sources & provenance** panel | The tools used (name, server, as-of) and, per figure, which tool it came from. |
| **Couldn’t verify: …** (red) | The exact number(s) that failed verification. |

---

## 4. “Some figures could not be verified” — what it means

This is **not** an error or a connection problem. It is the safety guardrail.

After the model writes an answer, the backend **checks every financial figure**
in it against the data the tools actually returned this turn. If a number isn’t
present in any tool result (and isn’t produced by a calculation tool), the answer
is **blocked** and replaced with a safe message. The red **“Couldn’t verify: …”**
chips show you exactly which number(s) tripped it.

### Why this happens (common cases)

1. **The tool doesn’t return that number.** `list_funds` returns fund ids, codes,
   names, base currency, inception date and status — it has **no NAV, AUM, price
   or value**. If the model adds a value while listing funds, that value can’t be
   verified, so it’s blocked. → Ask a figure question that hits a figure tool
   instead (e.g. *“latest NAV for fund X”* → `get_fund_nav`).
2. **There’s no data yet.** If your fund/portfolio list comes back empty (nothing
   seeded, or your user can’t see it), the model tends to “fill in” example
   numbers — which the guardrail correctly blocks. → Seed data and check
   permissions (§6).
3. **The model did the math itself.** Totals/ratios/averages must come from a
   calculation tool, not the model. → Ask for the specific portfolio so it calls
   `calc_*` / `get_portfolio_valuation`.

### How to see the exact figure that was blocked
- In the UI: the red **“Couldn’t verify:”** chips under the blocked answer.
- In logs: the backend prints `chat numeric validation blocked answer`.
- In audit: the `CHAT_NUMERIC_VALIDATION_FAILED` event lists `unverified_figures`.
- In the DB: that message’s `provenance_map.validation.violations`.

> The guardrail is intentional. A blocked answer means the assistant protected
> you from an unverified financial number. The fix is almost always “give it
> real data / ask a tool-backed question,” **not** turning the guardrail off.

---

## 5. What it will and won’t do

**Will:** list funds/portfolios; report the latest NAV; report an authoritative
portfolio valuation (market value, cost basis, unrealised P&L, ROI, AUM); compute
weighted-average price, allocation %, and exposure % with decimal-safe math; show
provenance for every figure; keep an audit trail of every turn and tool call.

**Won’t:** trade, post, approve, change any data (read-only by default); invent or
estimate a number; do free-hand arithmetic on money; browse the web; give
investment advice; show data you don’t have permission to see.

---

## 6. Setup & troubleshooting (local / dev)

The assistant needs four things to return real data. Check them in order.

### 6.1 An LLM API key (so chat is enabled at all)
The chat module only mounts when a provider key is present.
- Local backend (`make dev`): put it in **`backend/.env`**
  ```
  LLM_PROVIDER=anthropic
  ANTHROPIC_API_KEY=sk-ant-...
  ANTHROPIC_MODEL=claude-haiku-4-5-20251001
  ```
- Docker stack: put it in **`infra/.env`** (compose auto-loads it).

Both files are **gitignored** — safe to hold a real key, it won’t be committed.
If the key is missing, `/chat` isn’t mounted and the page can’t connect.

### 6.2 Database migrations (so turns persist)
Run **`make migrate-up`**. The chat tables plus the `correlation_id` columns must
exist, or saving a message fails and the turn dies.

### 6.3 Seed data + permissions (so there’s something to answer about)
- Run **`make seed`** — demo data lives under `database/seeds/zz_demo/`
  (`01_demo_funds.sql`, `03_demo_prices_valuations.sql`, `06_demo_data_permissions.sql`, …).
- Make sure your logged-in user has the right **function permissions and data
  scope**:
  | To use… | The user needs |
  |---|---|
  | `list_funds`, `get_fund_nav` | `INVESTMENT_FUND_VIEW` |
  | `list_portfolios`, `get_portfolio_holdings` | `INVESTMENT_PORTFOLIO_VIEW` |
  | `get_portfolio_valuation`, `calc_*` | `INVESTMENT_VALUATION_VIEW` |
  Tools run **as you** — if you can’t see a fund in the app, the assistant can’t either.

### 6.4 The MCP tool server (so it can fetch data)
The assistant reaches IMS data through the read-only **`ims-mcp`** binary.
- Docker: it’s built into the image automatically.
- Local: build it next to the server, e.g. `cd backend && go build -o ims-mcp ./cmd/ims-mcp`
  (or set `CHAT_MCP_IMS_BIN` to its path), then restart the backend.
- Check **`GET /health`** → the `chat` block shows `mcp_connected` and `tool_count`.
  If `mcp_connected: false`, the assistant can still chat but has **no tools**, so
  it can’t fetch figures.

### Quick checklist when something’s off
| Symptom | Likely cause | Fix |
|---|---|---|
| Page says “not connected” / `/chat` 404 | No API key, or backend not restarted | Set key in `backend/.env` / `infra/.env`, restart |
| Turn errors immediately | Migration not applied | `make migrate-up`, restart |
| Answer blocked: “could not be verified” | Model stated a number not in tool data, or no data | Ask a tool-backed figure question; `make seed`; check permissions |
| “you do not have permission to access this data” | Missing `INVESTMENT_*_VIEW` / data scope | Grant permission in IAM / settings |
| Tool chips never appear | MCP not connected | Build `ims-mcp`; check `/health` `chat.mcp_connected` |
| History rail empty | No sessions yet, or session endpoints missing | Send one message; ensure backend is current |

---

## 7. Privacy & safety notes

- The assistant runs every tool call **with your identity** — it can never show
  data you couldn’t already see in IMS.
- Your API key/token is **never** shown to the model, written to the audit trail,
  or sent to the browser.
- Tool output is treated as **untrusted data**: instructions hidden inside tool
  results are ignored, and the assistant won’t reveal secrets or run write tools
  because some text told it to.
- Every turn and tool call is written to an append-only **audit trail**; auditors
  with `IAM_AUDIT_VIEW` can review any session (and that review is itself audited).

---

## 8. One-line summary

Ask about your IMS funds/portfolios/NAV/valuations in plain language. The
assistant fetches the numbers from IMS, shows you where each one came from, and
**refuses to show any number it can’t verify** — so what you see is always real
IMS data, never a guess.
