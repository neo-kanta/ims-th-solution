# Agent Work Package

Use this template for every delegated task. Remove irrelevant fields, but do not
remove scope, acceptance, validation, or reporting requirements.

```text
Task ID:
Role and platform:
Objective:
Why this task exists:

Read first:
- repository instructions
- manager memory/handoff/tasks
- task-specific source or design files

Evidence/reproduction:

Read scope:
Exclusive write scope:
Files/modules forbidden to change:

Constraints:
Non-goals:
Dependencies and inputs:
Allowed environment/tools:
Human approval required for:

Acceptance criteria:
- [ ]

Validation required:
- command or observable proof

Execution instructions:
- Invoke frontier-backend-engineer for Go/API/database work, or
  frontier-frontend-engineer for Nuxt/UI/client work.
- Do not cross the backend/frontend write boundary without manager approval.
- Preserve unrelated changes.
- Stop and escalate if the task requires a policy decision or scope expansion.
- Do not commit, push, deploy, or use production data unless explicitly allowed.

Return:
- outcome and behavior changed
- files changed
- exact commands and pass/fail results
- evidence for each acceptance criterion
- assumptions, risks, and deferred work
- Git state and next action
```

## Manager Tracking Record

```text
Task ID:
Owner/agent ID:
Branch/worktree:
State: pending | active | blocked | review | complete
Dependencies:
Write ownership:
Last evidence:
Next checkpoint:
```

## Review Brief

For an independent reviewer, replace implementation instructions with:

```text
Review the raw diff against the acceptance criteria. Prioritize correctness,
data integrity, permissions, lifecycle ordering, rollback, failure behavior,
and missing tests. Do not edit. Report only reproducible findings with file and
line references, then list residual test gaps. Do not assume the proposed
implementation or the manager's prior conclusions are correct.
```
