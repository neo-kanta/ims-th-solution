<script setup lang="ts">
import { computed, ref } from "vue";

import AppCard from "~/shared/ui/AppCard.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import IMSApprovalStamp from "~/shared/ui/IMSApprovalStamp.vue";

import ApprovalStatusBadge from "./ApprovalStatusBadge.vue";
import ApprovalTimeline from "./ApprovalTimeline.vue";
import ApprovalActionPanel from "./ApprovalActionPanel.vue";
import { prettify } from "../lib/approvalStatus";
import { toStamps } from "../lib/approvalMappers";
import type { ApprovalRequestDetail } from "../types";

const props = defineProps<{
  detail: ApprovalRequestDetail;
  submitting?: boolean;
  actionError?: string | null;
  currentUserId?: string;
}>();

const emit = defineEmits<{
  approve: [comment: string];
  reject: [reason: string];
  withdraw: [];
  revoke: [reason: string];
}>();

const request = computed(() => props.detail.request);
const viewerTask = computed(() => props.detail.viewer_task ?? null);
const allowedActions = computed<string[]>(() => props.detail.allowed_actions ?? []);
const canAct = computed(() => allowedActions.value.includes("approve") || allowedActions.value.includes("reject"));
const canWithdraw = computed(() => allowedActions.value.includes("withdraw"));
const canRevoke = computed(() => allowedActions.value.includes("revoke"));
const stamps = computed(() => toStamps(props.detail.signatures ?? []));
const isDelegatedView = computed(() => Boolean(viewerTask.value?.is_delegated_action));

const revokeReason = ref("");
const showRevokeForm = ref(false);

function fmtDate(value?: string | null): string {
  if (!value) return "—";
  try {
    return new Intl.DateTimeFormat("en-CA", {
      timeZone: "Asia/Bangkok",
      year: "numeric",
      month: "short",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      hourCycle: "h23",
    }).format(new Date(value));
  } catch {
    return value;
  }
}
</script>

<template>
  <div v-if="request" class="request-detail">
    <!-- Header -->
    <AppCard>
      <div class="request-detail__header">
        <div>
          <div class="request-detail__number">{{ request.request_number }}</div>
          <h2 class="request-detail__title">
            {{ request.subject_title || request.subject_reference || prettify(request.subject_type ?? '') }}
          </h2>
          <p class="request-detail__sub">
            {{ prettify(request.process_type ?? '') }} ·
            {{ prettify(request.subject_type ?? '') }}
          </p>
        </div>
        <ApprovalStatusBadge :status="request.status ?? ''" />
      </div>

      <dl class="request-detail__grid">
        <div>
          <dt>Submitter</dt>
          <dd>{{ request.submitter_name || '—' }}</dd>
        </div>
        <div>
          <dt>Submitted</dt>
          <dd>{{ fmtDate(request.submitted_at) }}</dd>
        </div>
        <div>
          <dt>Current stage</dt>
          <dd>{{ request.current_stage_number }}</dd>
        </div>
        <div>
          <dt>Reference</dt>
          <dd>{{ request.subject_reference || '—' }}</dd>
        </div>
        <div v-if="request.subject?.display_label || request.subject_title">
          <dt>Subject</dt>
          <dd>{{ request.subject?.display_label || request.subject_title }}</dd>
        </div>
        <div v-if="request.final_decision_at">
          <dt>Final decision</dt>
          <dd>{{ fmtDate(request.final_decision_at) }}</dd>
        </div>
      </dl>

      <p v-if="request.rejection_reason" class="request-detail__rejection">
        <strong>Rejection reason:</strong> {{ request.rejection_reason }}
      </p>
    </AppCard>

    <!-- Delegated banner -->
    <div v-if="isDelegatedView" class="request-detail__delegated">
      You are acting as a delegate (proxy) for the originally assigned approver. Your sign-off will be recorded as a delegated approval.
    </div>

    <!-- Action panel -->
    <AppCard title="Decision">
      <ApprovalActionPanel
        :can-act="canAct"
        :submitting="submitting"
        :error="actionError"
        @approve="emit('approve', $event)"
        @reject="emit('reject', $event)"
      />
      <div v-if="canWithdraw || canRevoke" class="request-detail__secondary-actions">
        <AppButton
          v-if="canWithdraw"
          variant="ghost"
          size="sm"
          :disabled="submitting"
          @click="emit('withdraw')"
        >
          Withdraw request
        </AppButton>
        <template v-if="canRevoke">
          <AppButton
            v-if="!showRevokeForm"
            variant="ghost"
            size="sm"
            :disabled="submitting"
            class="btn-danger-ghost"
            @click="showRevokeForm = true"
          >
            Revoke approval
          </AppButton>
          <div v-else class="request-detail__revoke-form">
            <label class="request-detail__revoke-label">Reason for revocation</label>
            <textarea
              v-model="revokeReason"
              class="request-detail__revoke-textarea"
              rows="2"
              placeholder="Enter reason (required)"
            />
            <div class="request-detail__revoke-actions">
              <AppButton
                variant="danger"
                size="sm"
                :disabled="submitting || !revokeReason.trim()"
                @click="emit('revoke', revokeReason); showRevokeForm = false"
              >
                Confirm revoke
              </AppButton>
              <AppButton variant="ghost" size="sm" @click="showRevokeForm = false; revokeReason = ''">
                Cancel
              </AppButton>
            </div>
          </div>
        </template>
      </div>
    </AppCard>

    <!-- Signatures / stamps -->
    <AppCard title="Signatures &amp; stamps">
      <div v-if="stamps.length === 0" class="request-detail__empty">No approvals stamped yet.</div>
      <div v-else class="request-detail__stamps">
        <IMSApprovalStamp
          v-for="(s, i) in stamps"
          :key="i"
          :name="s.name"
          :role="s.role"
          :timestamp="s.timestamp"
          :delegated="s.delegated"
          :status="s.status"
        />
      </div>
    </AppCard>

    <!-- Timeline -->
    <AppCard title="Approval timeline">
      <ApprovalTimeline :events="detail.timeline ?? []" />
    </AppCard>
  </div>
</template>

<style scoped>
.request-detail {
  display: grid;
  gap: var(--space-5, 20px);
}
.request-detail__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4, 16px);
}
.request-detail__number {
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-tertiary, #6e7781);
}
.request-detail__title {
  margin: var(--space-1, 4px) 0 0;
  font-size: var(--font-size-lg, 1.125rem);
}
.request-detail__sub {
  margin: var(--space-1, 4px) 0 0;
  color: var(--text-secondary, #57606a);
  font-size: var(--font-size-sm, 0.875rem);
}
.request-detail__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: var(--space-4, 16px);
  margin: var(--space-4, 16px) 0 0;
}
.request-detail__grid dt {
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-tertiary, #6e7781);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.request-detail__grid dd {
  margin: 2px 0 0;
  font-weight: var(--font-weight-medium, 500);
}
.request-detail__mono {
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-xs, 0.75rem);
}
.request-detail__rejection {
  margin: var(--space-4, 16px) 0 0;
  color: var(--alert-danger-text, #cf222e);
  font-size: var(--font-size-sm, 0.875rem);
}
.request-detail__delegated {
  background: var(--alert-warning-bg, #fff8c5);
  border: 1px solid var(--alert-warning-border, #9a6700);
  color: var(--alert-warning-text, #9a6700);
  border-radius: var(--radius-md, 6px);
  padding: var(--space-3, 12px) var(--space-4, 16px);
  font-size: var(--font-size-sm, 0.875rem);
}
.request-detail__stamps {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-5, 20px);
  padding: var(--space-3, 12px) 0;
}
.request-detail__empty {
  color: var(--text-tertiary, #6e7781);
  font-size: var(--font-size-sm, 0.875rem);
}
.request-detail__secondary-actions {
  display: flex;
  gap: var(--space-3, 12px);
  align-items: flex-start;
  flex-wrap: wrap;
  margin-top: var(--space-3, 12px);
}
.request-detail__revoke-form {
  display: grid;
  gap: var(--space-2, 8px);
  width: 100%;
  max-width: 480px;
}
.request-detail__revoke-label {
  font-size: var(--font-size-sm, 0.875rem);
  font-weight: var(--font-weight-medium, 500);
}
.request-detail__revoke-textarea {
  width: 100%;
  padding: var(--space-2, 8px) var(--space-3, 12px);
  border: 1px solid var(--border-muted, #d0d7de);
  border-radius: var(--radius-sm, 4px);
  font-size: var(--font-size-sm, 0.875rem);
  resize: vertical;
}
.request-detail__revoke-actions {
  display: flex;
  gap: var(--space-2, 8px);
}

.btn-danger-ghost {
  color: var(--color-danger, #cf222e) !important;
}

.btn-danger-ghost:hover {
  background-color: var(--color-danger-subtle, #ffebe9) !important;
}
</style>
