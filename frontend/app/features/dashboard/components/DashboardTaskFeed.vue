<script setup lang="ts">
import { computed } from "vue";

import type {
  DashboardSnapshotDTO,
  DashboardTodoAction,
  DashboardTodoFilter,
  TaskDTO,
} from "../types";
import DashboardTaskCard from "./DashboardTaskCard.vue";

const props = withDefaults(defineProps<{
  snapshot: DashboardSnapshotDTO | null;
  loading?: boolean;
  error?: string | null;
  activeFilter?: DashboardTodoFilter;
}>(), {
  loading: false,
  error: null,
  activeFilter: "my",
});

const emit = defineEmits<{
  retry: [];
  taskAction: [action: DashboardTodoAction, task: TaskDTO];
}>();

const { t } = useI18n();

const allTasks = computed(() => props.snapshot?.tasks ?? []);

const filteredTasks = computed(() => {
  switch (props.activeFilter) {
    case "approvals":
      return allTasks.value.filter((task) => task.type === "RESEARCH_REVIEW");
    case "workflow":
      return allTasks.value.filter((task) => task.type === "WORKFLOW_PENDING");
    case "alerts":
      return allTasks.value.filter((task) => task.type === "COMPLIANCE_BREACH");
    case "done":
      return allTasks.value.filter((task) => task.status === "COMPLETED");
    case "my":
    default:
      return allTasks.value.filter((task) => task.status !== "COMPLETED");
  }
});

const highPriorityCount = computed(
  () => props.snapshot?.summary.highPriority ?? 0,
);

const hasTasks = computed(() => filteredTasks.value.length > 0);

const activeFilterLabel = computed(() => {
  switch (props.activeFilter) {
    case "approvals":
      return t("dashboard.layers.approvalsLabel");
    case "workflow":
      return t("dashboard.layers.workflowLabel");
    case "alerts":
      return t("dashboard.layers.alertsLabel");
    case "done":
      return t("dashboard.layers.doneLabel");
    case "my":
    default:
      return t("dashboard.layers.myLabel");
  }
});

const scopeLabel = computed(() => {
  if (props.loading && !props.snapshot) return t("dashboard.taskFeed.loadingSelectedLayer");
  return t("dashboard.taskFeed.scopeLabel", {
    filter: activeFilterLabel.value,
    count: filteredTasks.value.length,
  });
});

const stateTitle = computed(() => {
  if (props.activeFilter === "done") return t("dashboard.taskFeed.noCompletedTasks");
  if (props.activeFilter === "approvals") return t("dashboard.taskFeed.noReviewTasks");
  if (props.activeFilter === "workflow") return t("dashboard.taskFeed.noWorkflowTasks");
  if (props.activeFilter === "alerts") return t("dashboard.taskFeed.noAlertTasks");
  return t("dashboard.taskFeed.noPending");
});

const stateCopy = computed(() => {
  if (props.activeFilter === "my") {
    return t("dashboard.taskFeed.empty");
  }
  return t("dashboard.taskFeed.emptyFilter");
});
</script>

<template>
  <section class="task-feed" aria-labelledby="dashboard-todos-title">
    <div class="task-feed__header">
      <div>
        <p class="task-feed__eyebrow">{{ t("dashboard.taskFeed.eyebrow") }}</p>
        <h2 id="dashboard-todos-title" class="task-feed__title">
          {{ t("dashboard.taskFeed.queueTitle") }}
        </h2>
        <p class="task-feed__scope">{{ scopeLabel }}</p>
      </div>
      <span v-if="highPriorityCount > 0" class="task-feed__urgent-chip">
        {{ highPriorityCount }}
        {{ t("dashboard.taskFeed.urgent") }}
      </span>
    </div>

    <template v-if="hasTasks">
      <div class="task-feed__list">
        <DashboardTaskCard
          v-for="task in filteredTasks"
          :key="task.taskId"
          :task="task"
          @action="(action) => emit('taskAction', action, task)"
        />
      </div>
    </template>

    <div v-else-if="loading" class="task-feed__skeleton" aria-live="polite">
      <div
        v-for="idx in 4"
        :key="idx"
        class="task-feed__skeleton-card"
      >
        <span class="task-feed__skeleton-dot" />
        <span class="task-feed__skeleton-line is-wide" />
        <span class="task-feed__skeleton-line" />
      </div>
    </div>

    <div v-else-if="error" class="task-feed__state task-feed__state--error">
      <AppIcon name="warning" size="sm" />
      <div>
        <strong>{{ t("dashboard.taskFeed.sourceUnavailable") }}</strong>
        <span>{{ error }}</span>
      </div>
      <AppButton variant="ghost" size="sm" @click="emit('retry')">
        {{ t("dashboard.taskFeed.retry") }}
      </AppButton>
    </div>

    <div v-else class="task-feed__state">
      <AppIcon name="check" size="sm" />
      <div>
        <strong>{{ stateTitle }}</strong>
        <span>{{ stateCopy }}</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.task-feed {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.task-feed__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.task-feed__eyebrow {
  margin: 0 0 var(--space-1);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.task-feed__title {
  color: var(--text-primary);
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  margin: 0;
}

.task-feed__scope {
  margin: var(--space-1) 0 0;
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.task-feed__urgent-chip {
  padding: 2px var(--space-3);
  color: var(--status-rejected-text);
  background: var(--status-rejected-bg);
  border: 1px solid var(--alert-danger-border);
  border-radius: var(--radius-pill);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
}

.task-feed__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.task-feed__state {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-5);
  color: var(--text-secondary);
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  font-size: var(--font-size-sm);
}

.task-feed__state strong,
.task-feed__state span {
  display: block;
}

.task-feed__state strong {
  margin-bottom: var(--space-1);
  color: var(--text-primary);
}

.task-feed__state--error {
  color: var(--state-danger);
}

.task-feed__skeleton {
  display: grid;
  gap: var(--space-2);
}

.task-feed__skeleton-card {
  display: grid;
  grid-template-columns: 30px minmax(0, 1fr) 8rem;
  gap: var(--space-3);
  align-items: center;
  padding: var(--space-4);
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
}

.task-feed__skeleton-dot,
.task-feed__skeleton-line {
  display: block;
  background:
    linear-gradient(
      90deg,
      var(--bg-card-muted),
      var(--bg-row-hover),
      var(--bg-card-muted)
    );
  background-size: 220% 100%;
  animation: task-feed-pulse 1.2s ease-in-out infinite;
}

.task-feed__skeleton-dot {
  width: 30px;
  height: 30px;
  border-radius: var(--radius-md);
}

.task-feed__skeleton-line {
  height: 12px;
  border-radius: var(--radius-pill);
}

.task-feed__skeleton-line.is-wide {
  height: 16px;
}

@keyframes task-feed-pulse {
  from {
    background-position: 120% 0;
  }
  to {
    background-position: -120% 0;
  }
}

@media (max-width: 720px) {
  .task-feed__header {
    align-items: flex-start;
    flex-direction: column;
  }

  .task-feed__state,
  .task-feed__skeleton-card {
    grid-template-columns: 1fr;
  }
}
</style>
