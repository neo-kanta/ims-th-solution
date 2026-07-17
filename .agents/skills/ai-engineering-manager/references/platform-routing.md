# Platform Routing

Read this reference whenever selecting agents or transferring work between
ChatGPT/Codex and Claude Code. Platform features evolve; verify current official
documentation before depending on an experimental or version-specific feature.

## Decision Table

| Work shape | Preferred topology |
| --- | --- |
| One cohesive implementation with shared context | One primary engineer |
| Noisy codebase exploration, docs research, logs, or test output | Focused subagent returning a summary |
| Several independent read-only reviews | Parallel subagents, then manager synthesis |
| Independent backend/frontend/test implementations | Isolated worktrees with exclusive file ownership |
| Workers must challenge or message each other | Coordinated team, only when supported and worth the cost |
| Sequential tasks with same-file edits | One agent or chained agents, never parallel writers |
| Cross-provider continuation | Repository handoff artifact plus exact Git state |

## ChatGPT and Codex

- Discover repository skills from `.agents/skills`.
- Invoke skills explicitly with `$skill-name` when the workflow must be used.
- Use `$frontier-backend-engineer` for Go/API/database work and
  `$frontier-frontend-engineer` for Nuxt/UI/client work.
- Use `AGENTS.md` for concise durable repository conventions; use skills for
  reusable procedures and `docs/MANAGER` for changing program state.
- Ask explicitly for subagents or parallel work when independent investigation
  would improve quality or preserve main-context focus.
- Keep write-heavy parallel work in isolated worktrees and integrate through one
  owner.
- Preserve the main task for requirements, decisions, and synthesis; return
  distilled findings instead of raw logs.

## Claude Code

- Discover project skills from `.claude/skills` and invoke with `/skill-name`.
- Use `/frontier-backend-engineer` for Go/API/database work and
  `/frontier-frontend-engineer` for Nuxt/UI/client work.
- Use `CLAUDE.md` for durable repository guidance and skills for procedures.
- Use focused subagents for self-contained investigation, verification, or
  high-volume output that should not fill the main context.
- Use Agent Teams only when workers need direct communication and independent
  tasks justify significant token and coordination cost. The feature is
  experimental and disabled by default, so obtain user approval before enabling
  it.
- Use isolated worktrees for simultaneous writers. Agent-team coordination alone
  does not prevent file conflicts.
- Remember that fresh workers do not inherit the manager transcript. Include the
  task contract and point them to repository context.

## Cross-Provider Rules

- Treat Git, source files, tests, and `docs/MANAGER` as the shared protocol. A
  ChatGPT transcript is not Claude memory, and a Claude transcript is not
  ChatGPT memory.
- Record the exact branch, commit, worktree, dirty files, staged files, commands,
  and unresolved decisions before transfer.
- Never copy secrets or real customer/financial data between providers.
- Ask the receiving agent to verify the handoff against source before acting.
- Keep provider-specific commands out of durable business memory; place them in
  this routing reference or the transient handoff.

## Official Sources Consulted

- OpenAI, Build skills: https://learn.chatgpt.com/docs/build-skills.md
- OpenAI, Multi-agent operations: https://learn.chatgpt.com/docs/agent-configuration/subagents.md
- OpenAI, Harness engineering: https://openai.com/index/harness-engineering/
- Anthropic, Extend Claude with skills: https://code.claude.com/docs/en/slash-commands
- Anthropic, Create custom subagents: https://code.claude.com/docs/en/sub-agents
- Anthropic, Agent teams: https://code.claude.com/docs/en/agent-teams
- Anthropic, Claude Code best practices: https://code.claude.com/docs/en/best-practices
