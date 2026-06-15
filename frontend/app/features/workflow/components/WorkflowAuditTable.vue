<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import { ACTION_CATALOG } from "../permissions";
import type { WorkflowAction } from "../types";
import type { components } from "~/api/ims-api";
import AppBadge from "~/shared/ui/AppBadge.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";

const props = defineProps<{
  transitions?: components["schemas"]["DailyTimelineEntry"][];
}>();

const { t } = useI18n();

function formatTime(isoString?: string): string {
  if (!isoString) return "—";
  try {
    return new Intl.DateTimeFormat("en-CA", {
      timeZone: "Asia/Bangkok",
      year: "numeric",
      month: "short",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hourCycle: "h23",
    }).format(new Date(isoString));
  } catch {
    return isoString;
  }
}

function getActionLabel(action?: string): string {
  if (!action) return "";
  const opt = ACTION_CATALOG[action as WorkflowAction];
  return opt ? t(opt.labelKey, opt.labelFallback) : action;
}

function getActorDisplay(entry: components["schemas"]["DailyTimelineEntry"]): string {
  return entry.executedByUsername || entry.executedByAccountCode || "system";
}
</script>

<template>
  <div class="workflow-audit-container">
    <!-- Empty State -->
    <div v-if="!transitions || transitions.length === 0" class="workflow-audit-empty-state">
      <AppIcon name="audit" size="lg" class="empty-icon" />
      <span class="empty-title">No Transitions Found</span>
      <span class="empty-desc">There are no workflow state transitions logged for this day.</span>
    </div>

    <template v-else>
      <!-- Desktop Table View -->
      <div class="workflow-desktop-table-wrapper">
        <table class="workflow-audit-table">
          <thead>
            <tr>
              <th>Time</th>
              <th>Operation</th>
              <th>From → To State</th>
              <th>Actor</th>
              <th>Admin Override</th>
              <th>Remark</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(entry, index) in transitions" :key="entry.transitionId || index">
              <td class="cell-time">{{ formatTime(entry.executedAt) }}</td>
              <td class="cell-operation">
                <span class="operation-name">{{ getActionLabel(entry.operationType) }}</span>
              </td>
              <td class="cell-states">
                <div class="states-flow">
                  <span class="state-chip">{{ entry.fromState || "NOT_STARTED" }}</span>
                  <AppIcon name="chevron-down" size="xs" class="states-arrow" />
                  <span class="state-chip">{{ entry.toState }}</span>
                </div>
              </td>
              <td class="cell-actor">
                <div class="actor-info">
                  <span>{{ getActorDisplay(entry) }}</span>
                  <span v-if="entry.executedByAccountCode" class="actor-code">({{ entry.executedByAccountCode }})</span>
                </div>
              </td>
              <td class="cell-override">
                <AppBadge v-if="entry.isAdminOverride" variant="warning" size="sm">
                  Yes
                </AppBadge>
                <span v-else class="text-muted">No</span>
              </td>
              <td class="cell-remark">
                <span v-if="entry.reason" class="remark-text" :title="entry.reason">
                  {{ entry.reason }}
                </span>
                <span v-else class="text-muted">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Mobile Cards View -->
      <div class="workflow-mobile-cards">
        <div
          v-for="(entry, index) in transitions"
          :key="entry.transitionId || index"
          class="workflow-audit-card"
        >
          <div class="workflow-audit-card__header">
            <span class="workflow-audit-card__operation">
              {{ getActionLabel(entry.operationType) }}
            </span>
            <span class="workflow-audit-card__time">
              {{ formatTime(entry.executedAt) }}
            </span>
          </div>

          <div class="workflow-audit-card__states">
            <span class="state-chip">{{ entry.fromState || "NOT_STARTED" }}</span>
            <AppIcon name="chevron-down" size="xs" class="states-arrow" />
            <span class="state-chip">{{ entry.toState }}</span>
          </div>

          <div class="workflow-audit-card__details">
            <div class="workflow-detail-row">
              <span class="workflow-detail-row__label">Actor:</span>
              <span class="workflow-detail-row__val">
                {{ getActorDisplay(entry) }}
                <span v-if="entry.executedByAccountCode" class="actor-code">({{ entry.executedByAccountCode }})</span>
              </span>
            </div>
            <div class="workflow-detail-row">
              <span class="workflow-detail-row__label">Admin Override:</span>
              <span class="workflow-detail-row__val">
                <AppBadge v-if="entry.isAdminOverride" variant="warning" size="sm">Yes</AppBadge>
                <span v-else>No</span>
              </span>
            </div>
            <div v-if="entry.reason" class="workflow-detail-row is-remark">
              <span class="workflow-detail-row__label">Remark:</span>
              <p class="workflow-detail-row__remark-body">{{ entry.reason }}</p>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.workflow-audit-container {
  width: 100%;
}

.workflow-desktop-table-wrapper {
  overflow-x: auto;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
}

.workflow-audit-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm);
  text-align: left;
  background: var(--bg-card);
}

.workflow-audit-table th,
.workflow-audit-table td {
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--border-subtle);
  vertical-align: middle;
}

.workflow-audit-table th {
  background: var(--bg-row-hover);
  color: var(--text-secondary);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-xs);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.workflow-audit-table tbody tr:last-child td {
  border-bottom: none;
}

.workflow-audit-table tbody tr:hover {
  background: var(--bg-row-hover);
}

.cell-time {
  white-space: nowrap;
  color: var(--text-secondary);
}

.operation-name {
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.states-flow {
  display: flex;
  align-items: center;
  gap: var(--space-1);
}

.state-chip {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: var(--font-size-2xs);
  background: var(--color-neutral-100);
  padding: 2px var(--space-2);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
}

.states-arrow {
  transform: rotate(-90deg);
  color: var(--text-placeholder);
}

.actor-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.actor-code {
  font-size: var(--font-size-2xs);
  color: var(--text-tertiary);
}

.remark-text {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 280px;
  line-height: 1.4;
  color: var(--text-secondary);
}

.text-muted {
  color: var(--text-placeholder);
}

/* Empty State */
.workflow-audit-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-10) var(--space-6);
  text-align: center;
  gap: var(--space-2);
}

.empty-icon {
  color: var(--text-placeholder);
  margin-bottom: var(--space-2);
}

.empty-title {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-bold);
  color: var(--text-primary);
}

.empty-desc {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

/* Mobile View Styles */
.workflow-mobile-cards {
  display: none;
  flex-direction: column;
  gap: var(--space-4);
}

.workflow-audit-card {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--space-4);
  background: var(--bg-card);
  display: grid;
  gap: var(--space-3);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02);
}

.workflow-audit-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.workflow-audit-card__operation {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
  color: var(--text-primary);
}

.workflow-audit-card__time {
  font-size: var(--font-size-2xs);
  color: var(--text-tertiary);
}

.workflow-audit-card__states {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.workflow-audit-card__states .states-arrow {
  transform: rotate(-90deg);
}

.workflow-audit-card__details {
  display: grid;
  gap: var(--space-2);
  border-top: 1px solid var(--border-subtle);
  padding-top: var(--space-3);
}

.workflow-detail-row {
  display: flex;
  justify-content: space-between;
  font-size: var(--font-size-xs);
}

.workflow-detail-row__label {
  color: var(--text-tertiary);
  font-weight: var(--font-weight-medium);
}

.workflow-detail-row__val {
  color: var(--text-secondary);
}

.workflow-detail-row.is-remark {
  flex-direction: column;
  gap: var(--space-1);
  margin-top: var(--space-1);
}

.workflow-detail-row__remark-body {
  margin: 0;
  background: var(--color-neutral-50);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: var(--space-2);
  color: var(--text-secondary);
  line-height: 1.4;
}

@media (max-width: 768px) {
  .workflow-desktop-table-wrapper {
    display: none;
  }
  .workflow-mobile-cards {
    display: flex;
  }
}
</style>
