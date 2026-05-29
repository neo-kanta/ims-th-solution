<script setup lang="ts">
import { computed } from "vue";
import type { ComplianceSeverity, TaskDTO } from "../types";
import type { AppTranslationKey } from "~/composables/useI18n";

interface Props {
  task: TaskDTO;
}

const props = defineProps<Props>();

const { t } = useI18n();

const priorityBadgeVariant: Record<
  TaskDTO["priority"],
  "error" | "warning" | "neutral" | "info"
> = {
  HIGH: "error",
  MEDIUM: "warning",
  LOW: "neutral",
  INFO: "info",
};

const typeIcon: Record<TaskDTO["type"], string> = {
  RESEARCH_REVIEW: "review",
  WORKFLOW_PENDING: "workflow",
  COMPLIANCE_BREACH: "warning",
};

// Known workflow state codes that have a localised label. Keep in sync with
// dashboard.workflowState.* in app/shared/i18n/messages/{en,th,zh}/dashboard.ts.
const WORKFLOW_STATE_KEYS = new Set<string>([
  "NOT_STARTED",
  "DAY_OPEN",
  "MANAGER_APPROVED",
  "TRANSACTION_CLOSED",
  "ACCOUNTING_CLOSED",
]);

// Known module codes that have a localised label.
const MODULE_KEYS = new Set<string>([
  "investment",
  "workflow",
  "compliance",
  "integration",
  "iam",
  "audit",
  "market_data",
]);

const icon = computed(() => typeIcon[props.task.type] ?? "info");
const badgeVariant = computed(
  () => priorityBadgeVariant[props.task.priority] ?? "neutral",
);

// Localised priority label. The priority union is closed so all four keys are
// guaranteed to exist in the message catalogue.
const priorityLabel = computed(() => {
  const key = `dashboard.priority.${props.task.priority}` as AppTranslationKey;
  return t(key);
});

// Localised module label. Falls back to the raw module code (defensive
// against new backend modules that don't yet have a translation).
const moduleLabel = computed(() => {
  const module = props.task.module;
  if (MODULE_KEYS.has(module)) {
    return t(`dashboard.module.${module}` as AppTranslationKey);
  }
  return module;
});

function translateSeverity(severity: ComplianceSeverity): string {
  if (!severity) {
    return "";
  }
  return t(`dashboard.severity.${severity}` as AppTranslationKey);
}

function translateWorkflowState(state: string): string | null {
  if (!WORKFLOW_STATE_KEYS.has(state)) {
    return null;
  }
  return t(`dashboard.workflowState.${state}` as AppTranslationKey);
}

// Localised title built from the task's structured data:
//   research   → "Review: <report_no>"
//   workflow   → "<localised state>"
//   compliance → "<rule_type_id> (<localised severity>)"
// Falls back to the backend's English title if any required field is missing.
const localizedTitle = computed(() => {
  const task = props.task;

  if (task.type === "RESEARCH_REVIEW" && task.subject) {
    return t("dashboard.taskCard.researchReviewTitle", {
      subject: task.subject,
    });
  }

  if (task.type === "WORKFLOW_PENDING" && task.subject) {
    const stateLabel = translateWorkflowState(task.subject);
    if (stateLabel) {
      return t("dashboard.taskCard.workflowTitle", { state: stateLabel });
    }
  }

  if (task.type === "COMPLIANCE_BREACH" && task.subject && task.severity) {
    const severityLabel = translateSeverity(task.severity) || task.severity;
    return t("dashboard.taskCard.complianceTitle", {
      rule: task.subject,
      severity: severityLabel,
    });
  }

  return task.title;
});

const actionLabel = computed(() => t("dashboard.taskCard.openAction"));
</script>

<template>
  <article class="task-card">
    <div class="task-card__icon-col">
      <AppIcon :name="icon" size="sm" class="task-card__icon" />
    </div>
    <div class="task-card__body">
      <div class="task-card__header">
        <span class="task-card__title">{{ localizedTitle }}</span>
        <AppBadge :variant="badgeVariant" size="sm" class="task-card__badge">
          {{ priorityLabel }}
        </AppBadge>
      </div>
      <p class="task-card__desc">{{ task.description }}</p>
      <div class="task-card__meta">
        <span class="task-card__date">{{ task.businessDate }}</span>
        <span class="task-card__module">{{ moduleLabel }}</span>
      </div>
    </div>
    <div v-if="task.canAct && task.actionUrl" class="task-card__action">
      <NuxtLink
        :to="task.actionUrl"
        class="task-card__action-link"
        :aria-label="actionLabel"
        :title="actionLabel"
      >
        <AppIcon name="chevron-right" size="xs" />
      </NuxtLink>
    </div>
  </article>
</template>

<style scoped>
.task-card {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--bg-card);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  transition: border-color 0.15s;
}

.task-card:hover {
  border-color: var(--border-default);
}

.task-card__icon-col {
  flex-shrink: 0;
  margin-top: 1px;
}

.task-card__icon {
  color: var(--text-tertiary);
}

.task-card__body {
  flex: 1;
  min-width: 0;
}

.task-card__header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
  margin-bottom: var(--space-1);
}

.task-card__title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-card__badge {
  flex-shrink: 0;
}

.task-card__desc {
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
  margin: 0 0 var(--space-2);
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.task-card__meta {
  display: flex;
  gap: var(--space-3);
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.task-card__module {
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.task-card__action {
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.task-card__action-link {
  display: flex;
  align-items: center;
  color: var(--text-tertiary);
  transition: color 0.15s;
}

.task-card__action-link:hover {
  color: var(--text-primary);
}
</style>
