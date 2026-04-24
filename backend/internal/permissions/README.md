# Module: permissions

This module is the reserved backend boundary for a dedicated permissions subsystem.

## Intended Scope

- accounts and groups
- function-level rights
- data-scope and contract permissions
- permission administration workflows

## Current Status

This module is scaffolded only.

Important:

- route and repository placeholders exist here
- current permission checks in the running application are still enforced through IAM and platform middleware
- this folder should be treated as a future extraction target, not the current source of truth

## Guidance

If you start implementing this module for real, move responsibility deliberately instead of duplicating permission logic in both `iam` and `permissions`.
