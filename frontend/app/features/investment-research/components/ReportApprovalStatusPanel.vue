<script setup lang="ts">
import { computed } from "vue";
import AppCard from "~/shared/ui/AppCard.vue";
import ApprovalStatusBadge from "~/features/approval/components/ApprovalStatusBadge.vue";
import type { ApprovalSubjectStatus } from "~/features/approval/types";

const props = defineProps<{
  status: ApprovalSubjectStatus | null;
}>();

const request = computed(() => props.status?.request ?? null);
const hasRequest = computed(() => Boolean(props.status?.has_request && request.value));
</script>

<template>
  <AppCard title="Approval Status">
    <div v-if="!hasRequest" class="report-approval-panel__empty">
      No approval request has been submitted for this report yet.
    </div>

    <template v-else>
      <div class="report-approval-panel__row">
        <span class="report-approval-panel__label">Request No.</span>
        <span class="report-approval-panel__mono">{{ request!.request_number }}</span>
      </div>
      <div class="report-approval-panel__row">
        <span class="report-approval-panel__label">Status</span>
        <ApprovalStatusBadge :status="request!.status ?? ''" />
      </div>
      <div class="report-approval-panel__row">
        <span class="report-approval-panel__label">Stage</span>
        <span>{{ request!.current_stage_number }}</span>
      </div>
      <div class="report-approval-panel__actions">
        <NuxtLink
          class="report-approval-panel__link"
          :to="`/approval/requests/${request!.id}`"
        >
          View Full Approval Request →
        </NuxtLink>
      </div>
    </template>
  </AppCard>
</template>

<style scoped>
.report-approval-panel__empty {
  color: var(--text-secondary, #57606a);
  font-size: var(--font-size-sm, 0.875rem);
}
.report-approval-panel__row {
  display: flex;
  align-items: center;
  gap: var(--space-3, 12px);
  padding: var(--space-2, 8px) 0;
  border-bottom: 1px solid var(--border-muted, #f0f0f0);
}
.report-approval-panel__row:last-of-type {
  border-bottom: none;
}
.report-approval-panel__label {
  width: 100px;
  font-size: var(--font-size-xs, 0.75rem);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
  flex-shrink: 0;
}
.report-approval-panel__mono {
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-sm, 0.875rem);
}
.report-approval-panel__actions {
  padding-top: var(--space-3, 12px);
}
.report-approval-panel__link {
  font-size: var(--font-size-sm, 0.875rem);
  color: var(--color-primary-600, #2563eb);
  text-decoration: none;
}
.report-approval-panel__link:hover {
  text-decoration: underline;
}
</style>
