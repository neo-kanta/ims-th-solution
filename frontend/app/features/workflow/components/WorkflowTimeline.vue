<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "~/composables/useI18n";
import { ACTION_CATALOG } from "../permissions";
import type { WorkflowAction } from "../types";
import type { components } from "~/api/ims-api";
import AppBadge from "~/shared/ui/AppBadge.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";

const props = defineProps<{
  timeline?: components["schemas"]["DailyTimelineEntry"][];
}>();

const { t } = useI18n();

const sortedTimeline = computed(() => {
  if (!props.timeline) return [];
  // Sort in reverse chronological order
  return [...props.timeline].sort((a, b) => {
    const timeA = a.executedAt ? Date.parse(a.executedAt) : 0;
    const timeB = b.executedAt ? Date.parse(b.executedAt) : 0;
    return timeB - timeA;
  });
});

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
  <div class="workflow-timeline">
    <h3 class="workflow-section-title">
      {{ t("workflow.overview.timeline", "Timeline Summary") }}
    </h3>
    <div v-if="sortedTimeline.length === 0" class="workflow-timeline__empty">
      {{ t("dashboard.noData", "No workflow data is available yet.") }}
    </div>
    <div v-else class="workflow-timeline__list">
      <div
        v-for="(entry, index) in sortedTimeline"
        :key="entry.transitionId || index"
        class="workflow-timeline-item"
      >
        <div class="workflow-timeline-item__badge-connector">
          <div class="workflow-timeline-item__dot"></div>
          <div v-if="index < sortedTimeline.length - 1" class="workflow-timeline-item__line"></div>
        </div>

        <div class="workflow-timeline-item__content">
          <div class="workflow-timeline-item__top-row">
            <span class="workflow-timeline-item__action">
              {{ getActionLabel(entry.operationType) }}
            </span>
            <span class="workflow-timeline-item__time">
              {{ formatTime(entry.executedAt) }}
            </span>
          </div>

          <div class="workflow-timeline-item__meta-row">
            <div class="workflow-timeline-item__path">
              <span class="state-code">{{ entry.fromState || "NOT_STARTED" }}</span>
              <AppIcon name="chevron-down" size="xs" class="path-arrow" />
              <span class="state-code">{{ entry.toState }}</span>
            </div>

            <div class="workflow-timeline-item__personnel">
              <AppIcon name="user" size="xs" class="meta-icon" />
              <span>{{ getActorDisplay(entry) }}</span>
            </div>

            <AppBadge v-if="entry.isAdminOverride" variant="warning" size="sm">
              Admin Override
            </AppBadge>
          </div>

          <p v-if="entry.reason" class="workflow-timeline-item__remark">
            <span class="remark-label">Reason:</span> {{ entry.reason }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.workflow-timeline {
  display: grid;
  gap: var(--space-4);
}

.workflow-section-title {
  margin: 0;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.workflow-timeline__empty {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  background: var(--bg-card-muted, #f9fafb);
  text-align: center;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.workflow-timeline__list {
  display: grid;
  gap: 0;
  padding-left: var(--space-2);
}

.workflow-timeline-item {
  display: grid;
  grid-template-columns: 24px 1fr;
  gap: var(--space-3);
  position: relative;
}

.workflow-timeline-item__badge-connector {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.workflow-timeline-item__dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--color-primary-500);
  border: 2px solid var(--bg-card);
  box-shadow: 0 0 0 1px var(--color-primary-200);
  z-index: 2;
  margin-top: var(--space-1);
}

.workflow-timeline-item__line {
  flex-grow: 1;
  width: 2px;
  background: var(--border-subtle);
  margin-top: var(--space-1);
  margin-bottom: -1px;
}

.workflow-timeline-item__content {
  padding-bottom: var(--space-5);
  display: grid;
  gap: var(--space-2);
}

.workflow-timeline-item__top-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
}

.workflow-timeline-item__action {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.workflow-timeline-item__time {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.workflow-timeline-item__meta-row {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.workflow-timeline-item__path {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--font-size-2xs);
  font-family: var(--font-mono, ui-monospace, monospace);
  color: var(--text-secondary);
}

.path-arrow {
  transform: rotate(-90deg);
  color: var(--text-placeholder);
}

.workflow-timeline-item__personnel {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

.meta-icon {
  color: var(--text-placeholder);
}

.workflow-timeline-item__remark {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
  background: var(--color-neutral-50);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: var(--space-2) var(--space-3);
  line-height: 1.4;
}

.remark-label {
  font-weight: var(--font-weight-semibold);
  color: var(--text-tertiary);
}
</style>
