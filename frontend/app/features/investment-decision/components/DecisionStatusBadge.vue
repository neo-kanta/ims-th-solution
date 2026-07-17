<script setup lang="ts">
const props = defineProps<{ status?: string }>();

const labelMap: Record<string, string> = {
  DRAFT: "Draft",
  SUBMITTED: "Submitted",
  PENDING_APPROVAL: "Pending Approval",
  PENDING_COMPLIANCE_RELEASE: "Pending Compliance Release",
  APPROVED: "Approved",
  BLOCKED: "Blocked",
  REJECTED: "Rejected",
  CANCELLED: "Cancelled",
  READY_FOR_EXECUTION: "Ready for Execution",
  EXECUTED: "Executed",
};

const label = computed(() => labelMap[props.status ?? ""] ?? props.status ?? "—");
</script>

<template>
  <span class="status-badge" :data-status="status">{{ label }}</span>
</template>

<style scoped>
.status-badge {
  display: inline-block;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: 2px 8px;
  border-radius: 10px;
  background: var(--bg-card-hover);
  color: var(--text-secondary);
  white-space: nowrap;
}

.status-badge[data-status="DRAFT"] {
  background: rgba(110, 118, 129, 0.15);
  color: var(--text-secondary);
}

.status-badge[data-status="SUBMITTED"] {
  background: rgba(9, 105, 218, 0.15);
  color: var(--state-info, #0969da);
}

.status-badge[data-status="PENDING_APPROVAL"] {
  background: rgba(154, 103, 0, 0.15);
  color: var(--state-warning, #9a6700);
}

.status-badge[data-status="APPROVED"] {
  background: rgba(26, 127, 55, 0.15);
  color: var(--state-success, #1a7f37);
}

.status-badge[data-status="BLOCKED"] {
  background: rgba(207, 34, 46, 0.15);
  color: var(--state-danger, #cf222e);
}

.status-badge[data-status="CANCELLED"] {
  background: rgba(110, 118, 129, 0.15);
  color: var(--text-tertiary);
}

.status-badge[data-status="READY_FOR_EXECUTION"] {
  background: rgba(130, 80, 223, 0.15);
  color: #8250df;
}

.status-badge[data-status="PENDING_COMPLIANCE_RELEASE"] {
  background: rgba(154, 103, 0, 0.15);
  color: var(--state-warning, #9a6700);
}

.status-badge[data-status="REJECTED"] {
  background: rgba(207, 34, 46, 0.15);
  color: var(--state-danger, #cf222e);
}

.status-badge[data-status="EXECUTED"] {
  background: rgba(26, 127, 55, 0.15);
  color: var(--state-success, #1a7f37);
}
</style>
