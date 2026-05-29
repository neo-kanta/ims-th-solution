<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useDashboardTasks } from "../composables/useDashboardTasks";
import DashboardTaskCard from "./DashboardTaskCard.vue";

const { t } = useI18n();
const { snapshot, loading, error, fetchTasks } = useDashboardTasks();

onMounted(() => {
  void fetchTasks();
});

const hasTasks = computed(
  () => snapshot.value !== null && snapshot.value.tasks.length > 0,
);

const highPriorityCount = computed(
  () => snapshot.value?.summary.highPriority ?? 0,
);
</script>

<template>
  <section class="task-feed">
    <div class="task-feed__header">
      <h2 class="task-feed__title">
        {{ t("dashboard.taskFeed.title", "My Tasks") }}
      </h2>
      <span v-if="highPriorityCount > 0" class="task-feed__urgent-chip">
        {{ highPriorityCount }}
        {{ t("dashboard.taskFeed.urgent", "urgent") }}
      </span>
    </div>

    <template v-if="hasTasks && snapshot">
      <div class="task-feed__list">
        <DashboardTaskCard
          v-for="task in snapshot.tasks"
          :key="task.taskId"
          :task="task"
        />
      </div>
    </template>

    <div v-else-if="loading" class="task-feed__state">
      <AppIcon class="is-spinning" name="refresh" size="sm" />
      <span>{{ t("dashboard.taskFeed.loading", "Loading tasks…") }}</span>
    </div>

    <div v-else-if="error" class="task-feed__state task-feed__state--error">
      <AppIcon name="warning" size="sm" />
      <span>{{ error }}</span>
      <AppButton variant="ghost" size="sm" @click="fetchTasks">
        {{ t("dashboard.taskFeed.retry", "Retry") }}
      </AppButton>
    </div>

    <div v-else class="task-feed__state">
      <AppIcon name="check" size="sm" />
      <span>{{
        t("dashboard.taskFeed.empty", "No pending tasks — all clear.")
      }}</span>
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
  gap: var(--space-3);
}

.task-feed__title {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  margin: 0;
}

.task-feed__urgent-chip {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--state-danger);
  background: color-mix(in srgb, var(--state-danger) 10%, transparent);
  padding: 2px var(--space-2);
  border-radius: var(--radius-pill);
}

.task-feed__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.task-feed__state {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-4);
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  background: var(--bg-card);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
}

.task-feed__state--error {
  color: var(--state-danger);
}

.is-spinning {
  animation: task-feed-spin 0.9s linear infinite;
}

@keyframes task-feed-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
