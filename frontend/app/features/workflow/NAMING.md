# Workflow action naming

The backend module owns the canonical wire codes. The frontend uses backend
codes on the wire and spec labels in i18n. Do not introduce new wire codes
on the frontend.

| Wire code (backend, do not change) | Spec label (i18n in `messages/{en,th,zh}/workflow.ts`) | Stage cluster        |
| ---------------------------------- | ------------------------------------------------------ | -------------------- |
| `OPEN_DAY`                         | Investment Day Start                                   | Day Open             |
| `APPROVE`                          | Manager Approval                                       | Manager Approval     |
| `CLOSE_TRANSACTIONS`               | Transaction Closing                                    | Transaction Closing  |
| `CLOSE_ACCOUNTING`                 | Accounting Closing                                     | Accounting Closing   |
| `CANCEL_DAY_START`                 | Cancel Investment Day Start                            | Day Open (reverse)   |
| `CANCEL_APPROVAL`                  | Cancel Manager Approval                                | Manager Approval (r) |
| `CANCEL_TRANSACTION_CLOSE`         | Cancel Transaction Closing                             | Transaction (r)      |
| `ROLLBACK_ACCOUNTING_CLOSE`        | Rollback Accounting Closing                            | Accounting (r)       |

State codes (`WorkflowStateResponse.currentState`):
`NOT_STARTED`, `DAY_OPEN`, `MANAGER_APPROVED`, `TRANSACTION_CLOSED`, `ACCOUNTING_CLOSED`.

Permission codes are exhaustively declared in
`features/workflow/permissions.ts`. The lookup is the single source of
truth for the WORKFLOW_* permission required to dispatch each action.
