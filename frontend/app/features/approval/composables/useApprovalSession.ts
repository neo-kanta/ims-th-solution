import { ref, computed } from "vue";
import { approvalApi, approvalErrorMessage } from "../services/approvalApi";
import { derivePrincipalName } from "../lib/approvalMappers";
import type {
  ApprovalTarget,
  ApprovalInstance,
  ApprovalStatus,
  ApprovalAction,
  ApprovalStage,
  ApprovalApprover,
  ApprovalHistoryItem,
} from "../types";

export function useApprovalSession(target: ApprovalTarget) {
  const approval = ref<ApprovalInstance | null>(null);
  const rawTimeline = ref<any[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const actionPending = ref(false);

  // Group and Team member caches to avoid redundant API queries
  const groupCache = new Map<string, any[]>();
  const teamCache = new Map<string, any[]>();

  async function getGroupMembers(groupId: string) {
    if (groupCache.has(groupId)) return groupCache.get(groupId)!;
    try {
      const members = await approvalApi.listGroupMembers(groupId);
      groupCache.set(groupId, members);
      return members;
    } catch {
      return [];
    }
  }

  async function getTeamMembers(teamId: string) {
    if (teamCache.has(teamId)) return teamCache.get(teamId)!;
    try {
      const members = await approvalApi.listTeamMembers(teamId);
      teamCache.set(teamId, members);
      return members;
    } catch {
      return [];
    }
  }

  // Derive permissions dynamically based on backend state
  const permissions = computed(() => {
    const defaultPerms = {
      canView: true,
      canSubmit: false,
      canCancelSubmit: false,
      canApprove: false,
      canReject: false,
      canRevokeApproval: false,
      canExport: false,
    };

    if (!approval.value) return defaultPerms;

    // Check backend-level allowed actions for this request
    const allowed = approval.value.history ? (approval.value as any)._allowedActions || [] : [];
    const status = approval.value.status;

    return {
      canView: true,
      canSubmit: status === "NOT_SUBMITTED" || status === "DRAFT",
      canCancelSubmit: allowed.includes("withdraw") || allowed.includes("cancel"),
      canApprove: allowed.includes("approve"),
      canReject: allowed.includes("reject"),
      canRevokeApproval: allowed.includes("revoke") && status === "APPROVED",
      canExport: status === "APPROVED",
    };
  });

  async function refresh() {
    if (!target.recordId || !target.recordType) return;
    loading.value = true;
    error.value = null;

    try {
      // 1. Fetch current subject status
      const subjectStatus = await approvalApi.subjectStatus(target.recordType, target.recordId);

      if (subjectStatus.has_request && subjectStatus.request?.id) {
        const reqId = subjectStatus.request.id;

        // Fetch request details and timeline
        const [detail, timelineEvents] = await Promise.all([
          approvalApi.request(reqId),
          approvalApi.timeline(reqId),
        ]);
        rawTimeline.value = timelineEvents;

        const request = detail.request!;
        const signatures = detail.signatures ?? [];
        const tasks = detail.tasks ?? [];

        // Fetch process config definition
        let stagesDef: any[] = [];
        if (request.process_config_id) {
          try {
            const procConfig = await approvalApi.getProcess(request.process_config_id);
            stagesDef = procConfig.stages ?? [];
          } catch {
            // Fallback: construct stages dynamically from tasks and signatures
            const stagesCount = Math.max(
              ...signatures.map((s) => s.stage_number ?? 1),
              ...tasks.map((t) => t.stage_number ?? 1),
              1
            );
            stagesDef = Array.from({ length: stagesCount }, (_, i) => ({
              stage_number: i + 1,
              stage_name: `Stage ${i + 1}`,
              approver_mode: "SINGLE_USER",
            }));
          }
        }

        // Map stages and enrich with candidate/actual approver details
        const stages: ApprovalStage[] = [];
        for (const stage of stagesDef) {
          const sNo = stage.stage_number ?? 1;
          const sName = stage.stage_name ?? `Stage ${sNo}`;
          const mode: "SINGLE" | "GROUP_ANY" | "TEAM_STAMP" =
            stage.approver_mode === "TEAM_MINIMUM"
              ? "TEAM_STAMP"
              : stage.approver_mode === "GROUP_ANY" || stage.approver_mode === "GROUP_PRIORITY"
              ? "GROUP_ANY"
              : "SINGLE";

          // Calculate current stamp status
          const stageSignatures = signatures.filter((sig) => sig.stage_number === sNo);
          const stageTasks = tasks.filter((task) => task.stage_number === sNo && task.status === "PENDING");

          // Determine stage status
          let stageStatus: "PENDING" | "APPROVED" | "REJECTED" | "SKIPPED" = "SKIPPED";
          if (request.status === "APPROVED") {
            stageStatus = "APPROVED";
          } else if (request.status === "REJECTED" && request.current_stage_number === sNo) {
            stageStatus = "REJECTED";
          } else if (request.current_stage_number === sNo) {
            stageStatus = "PENDING";
          } else if (request.current_stage_number !== undefined && sNo < request.current_stage_number) {
            stageStatus = "APPROVED";
          }

          // Build list of approvers for the stage
          const approvers: ApprovalApprover[] = [];

          if (mode === "SINGLE") {
            // Find acting signer or pending assignee
            const sig = stageSignatures[0];
            const tsk = stageTasks[0];

            if (sig) {
              approvers.push({
                userId: sig.signer_user_id ?? "",
                displayName: sig.signer_display_name ?? "Approver",
                status: "APPROVED",
                actedAt: sig.signed_at,
                isAgent: sig.is_proxy_signature,
                principalUserName: derivePrincipalName(sig.is_proxy_signature, sig.proxy_for),
                remark: sig.signature_label === "DELEGATED" ? "Signed via delegation" : undefined,
              });
            } else if (tsk) {
              approvers.push({
                userId: tsk.assigned_user_id ?? "",
                displayName: tsk.assigned_user_name ?? "Assigned Approver",
                status: "WAITING",
                isAgent: tsk.is_delegated_action,
                principalUserName: derivePrincipalName(tsk.is_delegated_action, tsk.delegated_from),
              });
            } else {
              approvers.push({
                userId: stage.approver_user_id ?? "",
                displayName: "Approver",
                status: "WAITING",
              });
            }
          } else if (mode === "GROUP_ANY" && stage.approval_group_id) {
            const groupMembers = await getGroupMembers(stage.approval_group_id);
            for (const member of groupMembers) {
              const sig = stageSignatures.find((s) => s.signer_user_id === member.user_id);
              const tsk = stageTasks.find((t) => t.assigned_user_id === member.user_id);

              approvers.push({
                userId: member.user_id ?? "",
                displayName: member.user_display_name || member.username || "Group Member",
                status: sig ? "APPROVED" : tsk ? "WAITING" : "SKIPPED",
                actedAt: sig?.signed_at,
                isAgent: sig?.is_proxy_signature ?? tsk?.is_delegated_action,
                principalUserName: derivePrincipalName(
                  sig?.is_proxy_signature ?? tsk?.is_delegated_action,
                  sig?.proxy_for ?? tsk?.delegated_from,
                ),
              });
            }
          } else if (mode === "TEAM_STAMP" && stage.required_approval_count) {
            // Team stamp requires mapping of team members
            const teamId = stageTasks[0]?.assigned_team_id || stage.assigned_team_id;
            const teamMembers = teamId ? await getTeamMembers(teamId) : [];

            // Add anyone who has signed
            for (const sig of stageSignatures) {
              approvers.push({
                userId: sig.signer_user_id ?? "",
                displayName: sig.signer_display_name ?? "Team Signer",
                status: "APPROVED",
                actedAt: sig.signed_at,
                isAgent: sig.is_proxy_signature,
                principalUserName: derivePrincipalName(sig.is_proxy_signature, sig.proxy_for),
              });
            }

            // Add other pending team members
            for (const member of teamMembers) {
              if (!stageSignatures.some((s) => s.signer_user_id === member.user_id)) {
                const tsk = stageTasks.find((t) => t.assigned_user_id === member.user_id);
                approvers.push({
                  userId: member.user_id ?? "",
                  displayName: member.user_display_name || member.username || "Team Member",
                  status: tsk ? "WAITING" : "SKIPPED",
                  isAgent: tsk?.is_delegated_action,
                  principalUserName: derivePrincipalName(tsk?.is_delegated_action, tsk?.delegated_from),
                });
              }
            }
          }

          stages.push({
            stageNo: sNo,
            stageName: sName,
            approvalMode: mode,
            status: stageStatus,
            requiredStampCount: stage.required_approval_count,
            currentStampCount: stageSignatures.length,
            approvers,
          });
        }

        // Map history events
        const history: ApprovalHistoryItem[] = timelineEvents.map((evt) => {
          let action: ApprovalAction = "APPROVE";
          if (evt.event_type === "SUBMITTED") action = "SUBMIT";
          else if (evt.event_type === "CANCELLED" || evt.event_type === "WITHDRAWN") action = "CANCEL_SUBMISSION";
          else if (evt.event_type === "REJECTED") action = "REJECT";
          else if (evt.event_type === "REVOKED") action = "REVOKE_APPROVAL";

          return {
            id: evt.id ?? "",
            action,
            actionBy: evt.actor_user_id ?? "",
            actionByDisplayName: evt.actor_name ?? "User",
            actedAt: evt.created_at ?? "",
            stageNo: evt.stage_number ?? undefined,
            remark: evt.comment ?? undefined,
            isAgent: Boolean(evt.delegated_from?.id || evt.delegated_from_user_id),
            principalUserName: derivePrincipalName(
              Boolean(evt.delegated_from?.id || evt.delegated_from_user_id),
              evt.delegated_from,
            ),
          };
        });

        // Determine next approvers
        const nextApproverNames = tasks
          .filter((t) => t.stage_number === request.current_stage_number && t.status === "PENDING")
          .map((t) => t.assigned_user_name ?? "Unknown Approver");

        const instance: ApprovalInstance = {
          approvalId: request.id!,
          target: {
            ...target,
            title: request.subject_title || target.title,
          },
          status: (request.status as ApprovalStatus) || "DRAFT",
          currentStageNo: request.current_stage_number,
          nextApproverNames,
          submittedBy: request.submitter_name,
          submittedAt: request.submitted_at,
          completedAt: request.final_decision_at,
          stages,
          history,
        };

        // Attach allowed actions to instance for computing permissions
        (instance as any)._allowedActions = detail.allowed_actions ?? [];

        approval.value = instance;
      } else {
        rawTimeline.value = [];
        // No request exists, draft status
        // Fetch active process configuration if available to preview stage pipeline
        let previewStages: ApprovalStage[] = [];
        try {
          const procConfigs = await approvalApi.listProcesses({
            process_type: target.processType,
            active_only: true,
          });
          const config = procConfigs[0];
          if (config && config.stages) {
            previewStages = config.stages.map((stage) => {
              const mode: "SINGLE" | "GROUP_ANY" | "TEAM_STAMP" =
                stage.approver_mode === "TEAM_MINIMUM"
                  ? "TEAM_STAMP"
                  : stage.approver_mode === "GROUP_ANY" || stage.approver_mode === "GROUP_PRIORITY"
                  ? "GROUP_ANY"
                  : "SINGLE";

              return {
                stageNo: stage.stage_number ?? 1,
                stageName: stage.stage_name ?? `Stage ${stage.stage_number}`,
                approvalMode: mode,
                status: "SKIPPED",
                requiredStampCount: stage.required_approval_count,
                currentStampCount: 0,
                approvers: [],
              };
            });
          }
        } catch {
          // Ignore preview load failures
        }

        approval.value = {
          approvalId: "",
          target,
          status: "NOT_SUBMITTED",
          nextApproverNames: [],
          stages: previewStages,
          history: [],
        };
      }
    } catch (err) {
      error.value = approvalErrorMessage(err, "Failed to load approval session.");
    } finally {
      loading.value = false;
    }
  }

  async function submit(remark?: string) {
    if (actionPending.value) return;
    actionPending.value = true;
    error.value = null;
    try {
      const payload = {
        contract_id: target.recordType === "FUND" || target.recordType === "CONTRACT" ? target.recordId : undefined,
        contract_type: target.moduleCode,
        process_type: target.processType,
        subject_type: target.recordType,
        subject_id: target.recordId,
        subject_title: target.title || `Approval request for ${target.recordId}`,
        subject_reference: target.recordId,
      };
      await approvalApi.submit(payload);
      await refresh();
    } catch (err) {
      error.value = approvalErrorMessage(err, "Failed to submit request.");
      throw err;
    } finally {
      actionPending.value = false;
    }
  }

  async function cancelSubmission(remark?: string) {
    if (!approval.value?.approvalId || actionPending.value) return;
    actionPending.value = true;
    error.value = null;
    try {
      // Call cancel or withdraw based on permissions
      const allowed = (approval.value as any)._allowedActions || [];
      if (allowed.includes("withdraw")) {
        await approvalApi.withdraw(approval.value.approvalId);
      } else {
        await approvalApi.cancel(approval.value.approvalId);
      }
      await refresh();
    } catch (err) {
      error.value = approvalErrorMessage(err, "Failed to cancel submission.");
      throw err;
    } finally {
      actionPending.value = false;
    }
  }

  async function approve(remark?: string) {
    if (!approval.value?.approvalId || actionPending.value) return;
    actionPending.value = true;
    error.value = null;
    try {
      // Retrieve the pending task ID for the current user
      const detail = await approvalApi.request(approval.value.approvalId);
      const taskId = detail.viewer_task?.id;
      if (!taskId) throw new Error("No pending task found for your user.");

      await approvalApi.approve(taskId, { comment: remark });
      await refresh();
    } catch (err) {
      error.value = approvalErrorMessage(err, "Failed to approve request.");
      throw err;
    } finally {
      actionPending.value = false;
    }
  }

  async function reject(reason: string) {
    if (!approval.value?.approvalId || actionPending.value) return;
    actionPending.value = true;
    error.value = null;
    try {
      const detail = await approvalApi.request(approval.value.approvalId);
      const taskId = detail.viewer_task?.id;
      if (!taskId) throw new Error("No pending task found for your user.");

      await approvalApi.reject(taskId, { reason });
      await refresh();
    } catch (err) {
      error.value = approvalErrorMessage(err, "Failed to reject request.");
      throw err;
    } finally {
      actionPending.value = false;
    }
  }

  async function revokeApproval(remark?: string) {
    if (!approval.value?.approvalId || actionPending.value) return;
    actionPending.value = true;
    error.value = null;
    try {
      await approvalApi.revoke(approval.value.approvalId, { reason: remark });
      await refresh();
    } catch (err) {
      error.value = approvalErrorMessage(err, "Failed to revoke approval.");
      throw err;
    } finally {
      actionPending.value = false;
    }
  }

  return {
    approval,
    rawTimeline,
    loading,
    error,
    actionPending,
    permissions,
    refresh,
    submit,
    cancelSubmission,
    approve,
    reject,
    revokeApproval,
  };
}
