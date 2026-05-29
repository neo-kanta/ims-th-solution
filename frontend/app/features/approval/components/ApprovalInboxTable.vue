<script setup lang="ts">
import AppDataTable from "~/shared/ui/AppDataTable.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import type { TableColumn } from "~/shared/ui/AppDataTable.vue";

import ApprovalStatusBadge from "./ApprovalStatusBadge.vue";
import { prettify } from "../lib/approvalStatus";
import type { ApprovalInboxItem } from "../types";

defineProps<{
  items: ApprovalInboxItem[];
  loading?: boolean;
  error?: string | null;
}>();

const emit = defineEmits<{ open: [requestId: string] }>();

const columns: TableColumn[] = [
  { key: "request_number", label: "Request #" },
  { key: "process_type", label: "Process" },
  { key: "subject", label: "Subject" },
  { key: "submitter", label: "Submitter" },
  { key: "stage", label: "Stage", align: "center" },
  { key: "status", label: "Status", align: "center" },
  { key: "actions", label: "", align: "right" },
];

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
  <AppDataTable
    :columns="columns"
    :items="items"
    :loading="loading"
    :error="error"
    empty-text="No pending approvals in your inbox."
  >
    <template #[`cell(request_number)`]="{ item }">
      <span class="approval-inbox__mono">{{ (item as ApprovalInboxItem).request.request_number }}</span>
    </template>
    <template #[`cell(process_type)`]="{ item }">
      {{ prettify((item as ApprovalInboxItem).request.process_type ?? "") }}
    </template>
    <template #[`cell(subject)`]="{ item }">
      <div class="approval-inbox__subject">
        <span class="approval-inbox__subject-title">
          {{ (item as ApprovalInboxItem).request.subject_title || (item as ApprovalInboxItem).request.subject_reference || "—" }}
        </span>
        <span class="approval-inbox__subject-type">
          {{ prettify((item as ApprovalInboxItem).request.subject_type ?? "") }}
        </span>
      </div>
    </template>
    <template #[`cell(submitter)`]="{ item }">
      <div class="approval-inbox__subject">
        <span>{{ (item as ApprovalInboxItem).request.submitter_name || "—" }}</span>
        <span class="approval-inbox__subject-type">{{ fmtDate((item as ApprovalInboxItem).request.submitted_at) }}</span>
      </div>
    </template>
    <template #[`cell(stage)`]="{ item }">
      {{ (item as ApprovalInboxItem).task.stage_number }}
    </template>
    <template #[`cell(status)`]="{ item }">
      <ApprovalStatusBadge :status="(item as ApprovalInboxItem).request.status ?? ''" />
    </template>
    <template #[`cell(actions)`]="{ item }">
      <AppButton size="sm" variant="primary" @click="emit('open', (item as ApprovalInboxItem).request.id ?? '')">
        Review
      </AppButton>
    </template>
  </AppDataTable>
</template>

<style scoped>
.approval-inbox__mono {
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-sm, 0.875rem);
}
.approval-inbox__subject {
  display: flex;
  flex-direction: column;
}
.approval-inbox__subject-title {
  font-weight: var(--font-weight-semibold, 600);
}
.approval-inbox__subject-type {
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-tertiary, #6e7781);
}
</style>
