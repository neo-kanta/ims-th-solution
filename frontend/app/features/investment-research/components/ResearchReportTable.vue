<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import AppEmptyState from "~/shared/ui/AppEmptyState.vue";
import ResearchReportStatusBadge from "./ResearchReportStatusBadge.vue";
import type { ResearchReport } from "../types";

defineProps<{
  items: ResearchReport[];
  loading: boolean;
  error: string | null;
}>();

const emit = defineEmits<{
  open: [id: string];
}>();

const { t } = useI18n();

function openReport(id: string) {
  emit("open", id);
}
</script>

<template>
  <div class="research-table-wrap">
    <div v-if="loading" class="research-table-state">{{ t('investmentResearch.loadingList') }}</div>

    <div v-else-if="error" class="research-table-state research-table-state--error">
      {{ error }}
    </div>

    <AppEmptyState
      v-else-if="items.length === 0"
      :title="t('investmentResearch.empty.title')"
      :description="t('investmentResearch.empty.description')"
      icon="table"
    />

    <table v-else class="research-table">
      <thead>
        <tr>
          <th>{{ t('investmentResearch.reportNo') }}</th>
          <th>{{ t('investmentResearch.reportDate') }}</th>
          <th>{{ t('investmentResearch.instrument') }}</th>
          <th>{{ t('investmentResearch.recommendation') }}</th>
          <th>{{ t('investmentResearch.reportStatus') }}</th>
          <th>{{ t('investmentResearch.review') }}</th>
          <th class="research-table__title-col">{{ t('investmentResearch.titleField') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="report in items"
          :key="report.id"
          class="research-table__row"
          @click="openReport(report.id)"
        >
          <td class="research-table__report-no">{{ report.report_no }}</td>
          <td>{{ report.report_date }}</td>
          <td>
            <div class="research-table__instrument">
              <span class="research-table__instrument-code">
                {{ report.instrument_code }}
              </span>
              <span
                v-if="report.instrument_name"
                class="research-table__instrument-name"
              >
                {{ report.instrument_name }}
              </span>
            </div>
          </td>
          <td>
            <ResearchReportStatusBadge
              kind="recommendation"
              :value="report.recommendation"
            />
          </td>
          <td>
            <ResearchReportStatusBadge kind="report" :value="report.report_status" />
          </td>
          <td>
            <ResearchReportStatusBadge kind="review" :value="report.review_status" />
          </td>
          <td class="research-table__title">{{ report.report_title || "—" }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.research-table-wrap {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-md, 8px);
  overflow: hidden;
}

.research-table-state {
  padding: var(--space-6);
  text-align: center;
  color: var(--text-secondary, #4b5563);
}

.research-table-state--error {
  color: var(--state-error, #dc2626);
}

.research-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm, 0.875rem);
}

.research-table thead th {
  background: var(--bg-table-header, #f9fafb);
  color: var(--text-secondary, #4b5563);
  font-weight: var(--font-weight-semibold, 600);
  font-size: var(--font-size-xs, 0.75rem);
  text-align: left;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
}

.research-table tbody td {
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
  vertical-align: middle;
}

.research-table tbody tr:last-child td {
  border-bottom: none;
}

.research-table__row {
  cursor: pointer;
  transition: background-color 0.1s ease;
}

.research-table__row:hover {
  background: var(--bg-row-hover, #f3f4f6);
}

.research-table__report-no {
  font-weight: var(--font-weight-semibold, 600);
}

.research-table__instrument {
  display: grid;
  gap: 2px;
}

.research-table__instrument-code {
  font-weight: var(--font-weight-semibold, 600);
}

.research-table__instrument-name {
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-secondary, #6b7280);
}

.research-table__title {
  color: var(--text-secondary, #4b5563);
  max-width: 320px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
