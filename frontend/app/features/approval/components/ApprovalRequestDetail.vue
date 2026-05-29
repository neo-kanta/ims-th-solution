<script setup lang="ts">
import { computed } from "vue";

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
}>();

const request = computed(() => props.detail.request);
const viewerTask = computed(() => props.detail.viewer_task ?? null);
const canAct = computed(() => Boolean(viewerTask.value));
const stamps = computed(() => toStamps(props.detail.signatures ?? []));
const isDelegatedView = computed(() => Boolean(viewerTask.value?.is_delegated_action));

const canWithdraw = computed(() => {
  const r = request.value;
  if (!r) return false;
  const active = r.status === "PENDING_APPROVAL" || r.status === "SUBMITTED";
  return active && !!props.currentUserId && r.submitter_id === props.currentUserId;
});

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
  <div class="request-detail">
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
        <div v-if="request.contract_id">
          <dt>Contract / fund</dt>
          <dd class="request-detail__mono">{{ request.contract_id }}</dd>
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
      <div v-if="canWithdraw" class="request-detail__withdraw">
        <AppButton variant="ghost" size="sm" :disabled="submitting" @click="emit('withdraw')">
          Withdraw request
        </AppButton>
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
.request-detail__withdraw {
  margin-top: var(--space-3, 12px);
}
</style>
