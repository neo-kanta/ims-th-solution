# Approval Config Snapshot — Design Reference

**As of:** 2026-06-17

The config snapshot ensures that in-flight approval requests cannot be retroactively altered by an admin editing the approval process configuration.

---

## Problem

When `ApproveTask` is called to advance a multi-stage request, it previously called:
```go
cfg, err := s.repo.GetConfig(ctx, r.ProcessConfigID)
next, hasNext := nextStage(cfgStages(cfg), task.StageNumber)
```

If an admin changed the process config between stage 1 approval and stage 2 advancement (e.g., added a stage, changed approver modes, removed a stage), the in-flight request would follow the new config — not the one that was active when the submitter agreed to the workflow.

This violates the principle that a submitted request is a commitment under the rules that existed at submission time.

---

## Design

### Entity

```go
// entity/request.go
type StageSnapshot struct {
    StageNumber           int    `json:"stage_number"`
    StageName             string `json:"stage_name"`
    ApproverMode          string `json:"approver_mode"`
    ApproverUserID        string `json:"approver_user_id,omitempty"`
    ApprovalGroupID       string `json:"approval_group_id,omitempty"`
    RequiredApprovalCount int    `json:"required_approval_count"`
    IsFinalStage          bool   `json:"is_final_stage"`
    RejectPolicy          string `json:"reject_policy"`
}

// ApprovalRequest
ConfigSnapshot []StageSnapshot
```

### Persistence

Migration: `20260617000001_approval__config_snapshot.up.sql`
```sql
ALTER TABLE approval__requests
    ADD COLUMN IF NOT EXISTS config_snapshot JSONB NOT NULL DEFAULT '[]'::jsonb;
```

`requestSelect` includes `COALESCE(r.config_snapshot, '[]'::jsonb)`. `scanRequest` unmarshals the JSON into `[]entity.StageSnapshot`. `CreateRequest` marshals and inserts it.

### Submit

`SubmitApproval` populates the snapshot immediately after resolving the config:
```go
req.ConfigSnapshot = entity.StageSnapshotFromConfig(cfg.Stages)
```

### Stage Advancement

`resolveStages()` is the single point that resolves stages for a request:
```go
func (s *ApprovalRuntimeService) resolveStages(ctx, r) ([]entity.ApprovalProcessStage, error) {
    if len(r.ConfigSnapshot) > 0 {
        // Use frozen snapshot
        return stagesFromSnapshot(r.ConfigSnapshot), nil
    }
    // Backwards-compatible fallback for pre-snapshot requests
    cfg, err := s.repo.GetConfig(ctx, derefUUID(r.ProcessConfigID))
    return cfgStages(cfg), nil
}
```

Both `loadActionContext` and the stage-advancement block in `ApproveTask` now call `resolveStages()` instead of calling `GetConfig` directly.

---

## Backwards Compatibility

Requests created before the migration have `config_snapshot = '[]'` (empty array). `resolveStages()` detects `len == 0` and falls back to the live config. This is safe because:
1. Those requests were already routed under the live config before this fix.
2. If the config was edited between that request's submit and this migration, the live config is the best available approximation.

New requests (after migration) always have a non-empty snapshot and are fully immutable.

---

## What Is NOT Snapshotted

The snapshot stores only the fields needed for stage routing and completion logic:
- `stage_number`, `stage_name`, `approver_mode`, `approver_user_id`, `approval_group_id`
- `required_approval_count`, `is_final_stage`, `reject_policy`

Fields NOT snapshotted:
- `process_config_id` (already on the request row as a FK reference)
- Stage `id` (UUID), `created_at`, `updated_at` — not used during advancement
- Group/team membership — these are resolved live at task-creation time (intentional: delegation changes should take effect on future stages)

---

## Operational Notes

- Config changes take effect only for new submissions. In-flight requests are isolated by their snapshot.
- To retroactively fix a broken in-flight request (e.g., wrong approver configured), an admin must revoke the request, correct the config, and ask the submitter to re-submit.
- The snapshot column is a JSONB column with a non-null default; no backfill is required for existing rows.
