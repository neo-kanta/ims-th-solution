<script setup lang="ts">
import { computed } from "vue";
import type { ComplianceSeverity, DashboardTodoAction, TaskDTO } from "../types";
import type { AppTranslationKey } from "~/composables/useI18n";

interface Props {
  task: TaskDTO;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  action: [action: DashboardTodoAction, task: TaskDTO];
}>();

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

const taskTypeBadge: Record<
  TaskDTO["type"],
  { label: string; variant: "purple" | "info" | "error" }
> = {
  RESEARCH_REVIEW: { label: "Review", variant: "purple" },
  WORKFLOW_PENDING: { label: "Workflow", variant: "info" },
  COMPLIANCE_BREACH: { label: "Alert", variant: "error" },
};

const statusBadgeVariant: Record<
  TaskDTO["status"],
  "warning" | "info" | "success"
> = {
  PENDING: "warning",
  IN_PROGRESS: "info",
  COMPLETED: "success",
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

const typeBadge = computed(() => taskTypeBadge[props.task.type]);

const statusVariant = computed(() => statusBadgeVariant[props.task.status]);

const statusLabel = computed(() => {
  switch (props.task.status) {
    case "COMPLETED":
      return "Done";
    case "IN_PROGRESS":
      return t("dashboard.status.IN_PROGRESS");
    case "PENDING":
    default:
      return t("dashboard.status.PENDING");
  }
});

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

const relatedLabel = computed(() => {
  if (props.task.contractId) return props.task.contractId;
  if (props.task.subject) return props.task.subject;
  return moduleLabel.value;
});

function emitPreviewAction(action: DashboardTodoAction) {
  emit("action", action, props.task);
}
</script>

<template>
  <article class="todo-card" :class="`todo-card--${task.priority.toLowerCase()}`">
    <div class="todo-card__accent" aria-hidden="true" />

    <div class="todo-card__icon-col">
      <AppIcon :name="icon" size="sm" class="todo-card__icon" />
    </div>

    <div class="todo-card__body">
      <div class="todo-card__header">
        <div class="todo-card__title-block">
          <span class="todo-card__title">{{ localizedTitle }}</span>
          <span class="todo-card__description">{{ task.description }}</span>
        </div>

        <div class="todo-card__badges">
          <AppBadge :variant="typeBadge.variant" size="sm" class="todo-card__badge">
            {{ typeBadge.label }}
          </AppBadge>
          <AppBadge :variant="badgeVariant" size="sm" class="todo-card__badge">
            {{ priorityLabel }}
          </AppBadge>
          <AppBadge :variant="statusVariant" dot size="sm" class="todo-card__badge">
            {{ statusLabel }}
          </AppBadge>
        </div>
      </div>

      <div class="todo-card__meta" aria-label="Task metadata">
        <span class="todo-card__meta-item">
          <AppIcon name="clock" size="xs" />
          <span>{{ task.businessDate || "No business date" }}</span>
        </span>
        <span class="todo-card__meta-item">
          <AppIcon name="portfolio" size="xs" />
          <span>{{ relatedLabel }}</span>
        </span>
        <span class="todo-card__meta-item">
          <AppIcon name="dashboard" size="xs" />
          <span>{{ moduleLabel }}</span>
        </span>
      </div>
    </div>

    <div class="todo-card__actions" aria-label="Task quick actions">
      <NuxtLink
        v-if="task.canAct && task.actionUrl"
        :to="task.actionUrl"
        class="todo-card__action todo-card__action--primary"
        :aria-label="actionLabel"
        :title="actionLabel"
      >
        <AppIcon name="chevron-right" size="xs" />
        <span>Open</span>
      </NuxtLink>
      <button
        class="todo-card__action"
        type="button"
        :disabled="task.status === 'COMPLETED'"
        @click="emitPreviewAction('mark-done')"
      >
        <AppIcon name="check" size="xs" />
        <span>Mark done</span>
      </button>
      <button
        class="todo-card__action"
        type="button"
        @click="emitPreviewAction('snooze')"
      >
        <AppIcon name="clock" size="xs" />
        <span>Snooze</span>
      </button>
      <button
        class="todo-card__action todo-card__action--icon"
        type="button"
        aria-label="More task actions"
        @click="emitPreviewAction('more')"
      >
        <span aria-hidden="true">...</span>
      </button>
    </div>
  </article>
</template>

<style scoped>
.todo-card {
  position: relative;
  display: grid;
  grid-template-columns: 30px minmax(0, 1fr) auto;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-4);
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  transition:
    background var(--transition-fast),
    border-color var(--transition-fast),
    box-shadow var(--transition-fast);
  overflow: hidden;
}

.todo-card:hover {
  background: var(--bg-card-hover);
  border-color: var(--border-default);
  box-shadow: var(--shadow-xs);
}

.todo-card__accent {
  position: absolute;
  inset: 0 auto 0 0;
  width: 3px;
  background: var(--state-info);
}

.todo-card--high .todo-card__accent {
  background: var(--state-danger);
}

.todo-card--medium .todo-card__accent {
  background: var(--state-warning);
}

.todo-card--low .todo-card__accent {
  background: var(--text-tertiary);
}

.todo-card__icon-col {
  flex-shrink: 0;
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  color: var(--text-secondary);
  background: var(--bg-card-muted);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.todo-card__icon {
  color: var(--text-tertiary);
}

.todo-card__body {
  min-width: 0;
  display: grid;
  gap: var(--space-3);
}

.todo-card__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.todo-card__title-block {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.todo-card__title {
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
}

.todo-card__description {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-normal);
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.todo-card__badges {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.todo-card__badge {
  white-space: nowrap;
}

.todo-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.todo-card__meta-item {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.todo-card__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
  flex-wrap: wrap;
  min-width: 15rem;
}

.todo-card__action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 30px;
  padding: 0 var(--space-3);
  color: var(--text-secondary);
  background: var(--action-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  text-decoration: none;
  transition:
    background var(--transition-fast),
    border-color var(--transition-fast),
    color var(--transition-fast);
}

.todo-card__action:hover:not(:disabled) {
  color: var(--text-primary);
  background: var(--bg-row-hover);
  border-color: var(--border-default);
  text-decoration: none;
}

.todo-card__action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.todo-card__action--primary {
  color: var(--text-link);
}

.todo-card__action--icon {
  width: 30px;
  padding: 0;
  font-family: var(--font-family-mono);
  font-weight: var(--font-weight-bold);
}

@media (max-width: 980px) {
  .todo-card {
    grid-template-columns: 30px minmax(0, 1fr);
  }

  .todo-card__actions {
    grid-column: 2;
    justify-content: flex-start;
    min-width: 0;
  }
}

@media (max-width: 640px) {
  .todo-card {
    grid-template-columns: 1fr;
  }

  .todo-card__icon-col {
    display: none;
  }

  .todo-card__header {
    flex-direction: column;
    gap: var(--space-3);
  }

  .todo-card__badges,
  .todo-card__actions {
    justify-content: flex-start;
  }

  .todo-card__actions {
    grid-column: auto;
  }

  .todo-card__action:not(.todo-card__action--icon) {
    flex: 1 1 7rem;
  }
}
</style>
