<script setup lang="ts">
import { computed, ref } from "vue";

const { t } = useI18n();

import type {
  DashboardSnapshotDTO,
  DashboardTodoFilter,
  TaskDTO,
} from "../types";

const props = withDefaults(defineProps<{
  snapshot: DashboardSnapshotDTO | null;
  activeFilter: DashboardTodoFilter;
  loading?: boolean;
  error?: string | null;
}>(), {
  loading: false,
  error: null,
});

const emit = defineEmits<{
  "update:activeFilter": [filter: DashboardTodoFilter];
}>();

interface LayerItem {
  value: DashboardTodoFilter;
  label: string;
  description: string;
  icon: string;
  count: number;
}

const searchQuery = ref("");
const tasks = computed(() => props.snapshot?.tasks ?? []);

function countBy(predicate: (task: TaskDTO) => boolean): number {
  return tasks.value.filter(predicate).length;
}

const layers = computed<LayerItem[]>(() => [
  {
    value: "my",
    label: t("dashboard.layers.myLabel", "My task queue"),
    description: t("dashboard.layers.myDesc", "Open items assigned to this session"),
    icon: "list",
    count: countBy((task) => task.status !== "COMPLETED"),
  },
  {
    value: "approvals",
    label: t("dashboard.layers.approvalsLabel", "Review approvals"),
    description: t("dashboard.layers.approvalsDesc", "Research and manager review work"),
    icon: "approval",
    count: countBy((task) => task.type === "RESEARCH_REVIEW"),
  },
  {
    value: "workflow",
    label: t("dashboard.layers.workflowLabel", "Business-day flow"),
    description: t("dashboard.layers.workflowDesc", "Investment day and closing steps"),
    icon: "workflow",
    count: countBy((task) => task.type === "WORKFLOW_PENDING"),
  },
  {
    value: "alerts",
    label: t("dashboard.layers.alertsLabel", "Compliance alerts"),
    description: t("dashboard.layers.alertsDesc", "Exceptions requiring attention"),
    icon: "warning",
    count: countBy((task) => task.type === "COMPLIANCE_BREACH"),
  },
  {
    value: "done",
    label: t("dashboard.layers.doneLabel", "Completed archive"),
    description: t("dashboard.layers.doneDesc", "Resolved items from the task source"),
    icon: "check",
    count: countBy((task) => task.status === "COMPLETED"),
  },
]);

const filteredLayers = computed(() => {
  const term = searchQuery.value.trim().toLowerCase();
  if (!term) return layers.value;

  return layers.value.filter((layer) =>
    [layer.label, layer.description].some((value) =>
      value.toLowerCase().includes(term),
    ),
  );
});

const sourceLabel = computed(() => {
  if (props.loading) return t("dashboard.layers.loadingSource", "Loading task source");
  if (props.error) return t("dashboard.layers.sourceUnavailable", "Task source unavailable");
  if (props.snapshot) return t("dashboard.layers.integrationApi", "Integration task API");
  return t("dashboard.layers.waitingSource", "Waiting for task source");
});

function selectLayer(layer: DashboardTodoFilter) {
  emit("update:activeFilter", layer);
}

function resetLayers() {
  searchQuery.value = "";
  emit("update:activeFilter", "my");
}
</script>

<template>
  <aside class="layer-sidebar" aria-labelledby="dashboard-layer-sidebar-title">
    <div class="layer-sidebar__header">
      <span class="layer-sidebar__mark" aria-hidden="true">
        <AppIcon name="portfolio" size="sm" />
      </span>
      <div class="layer-sidebar__heading">
        <p class="layer-sidebar__eyebrow">{{ t("dashboard.layers.fundManagerHome", "Fund manager home") }}</p>
        <h2 id="dashboard-layer-sidebar-title" class="layer-sidebar__title">
          {{ t("dashboard.layers.taskLayers", "Task layers") }}
        </h2>
      </div>
    </div>

    <label class="layer-sidebar__search-label" for="dashboard-layer-search">
      {{ t("dashboard.layers.findLayer", "Find a layer") }}
    </label>
    <div class="layer-sidebar__search">
      <AppIcon name="search" size="xs" />
      <input
        id="dashboard-layer-search"
        v-model="searchQuery"
        type="search"
        :placeholder="t('dashboard.layers.findLayerPlaceholder', 'Find a task layer...')"
        autocomplete="off"
      />
    </div>

    <nav class="layer-sidebar__nav" aria-label="Dashboard task layers">
      <button
        v-for="layer in filteredLayers"
        :key="layer.value"
        class="layer-sidebar__item"
        :class="{ 'is-active': activeFilter === layer.value }"
        type="button"
        :aria-current="activeFilter === layer.value ? 'page' : undefined"
        @click="selectLayer(layer.value)"
      >
        <span class="layer-sidebar__item-icon">
          <AppIcon :name="layer.icon" size="xs" />
        </span>
        <span class="layer-sidebar__item-copy">
          <span class="layer-sidebar__item-title">{{ layer.label }}</span>
          <span class="layer-sidebar__item-desc">{{ layer.description }}</span>
        </span>
        <span class="layer-sidebar__count">
          <template v-if="loading">...</template>
          <template v-else>{{ layer.count }}</template>
        </span>
      </button>

      <div v-if="filteredLayers.length === 0" class="layer-sidebar__empty">
        {{ t("dashboard.layers.noMatchingLayers", "No matching layers") }}
      </div>
    </nav>

    <button class="layer-sidebar__reset" type="button" @click="resetLayers">
      {{ t("dashboard.layers.showPrimaryQueue", "Show primary queue") }}
    </button>

    <hr class="layer-sidebar__separator" />

    <div class="layer-sidebar__source">
      <span class="layer-sidebar__source-dot" :class="{ 'is-error': error }" />
      <span>{{ sourceLabel }}</span>
    </div>
  </aside>
</template>

<style scoped>
.layer-sidebar {
  display: grid;
  gap: var(--space-4);
  min-width: 0;
  padding: var(--space-8) var(--space-6);
  border-right: 1px solid var(--border-subtle);
  background: var(--bg-sidebar);
  height: 100%;
}

.layer-sidebar__header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.layer-sidebar__mark {
  display: grid;
  place-items: center;
  width: 34px;
  height: 34px;
  color: var(--text-link);
  background: var(--bg-selected);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.layer-sidebar__heading {
  min-width: 0;
}

.layer-sidebar__eyebrow {
  margin: 0 0 var(--space-1);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.layer-sidebar__title {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
}

.layer-sidebar__search-label {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.layer-sidebar__search {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  height: 34px;
  padding: 0 var(--space-3);
  color: var(--text-tertiary);
  background: var(--bg-input);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  transition:
    border-color var(--transition-fast),
    box-shadow var(--transition-fast);
}

.layer-sidebar__search:focus-within {
  border-color: var(--border-focus);
  box-shadow: var(--shadow-focus);
}

.layer-sidebar__search input {
  min-width: 0;
  width: 100%;
  color: var(--text-primary);
  background: transparent;
  border: 0;
  outline: none;
  font: inherit;
  font-size: var(--font-size-xs);
}

.layer-sidebar__search input::placeholder {
  color: var(--text-placeholder);
}

.layer-sidebar__nav {
  display: grid;
  gap: var(--space-2);
}

.layer-sidebar__item {
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-3);
  min-height: 48px;
  padding: var(--space-2);
  color: var(--text-secondary);
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  text-align: left;
  transition:
    background var(--transition-fast),
    border-color var(--transition-fast),
    color var(--transition-fast);
}

.layer-sidebar__item:hover {
  color: var(--text-primary);
  background: var(--bg-row-hover);
}

.layer-sidebar__item.is-active {
  color: var(--text-primary);
  background: var(--bg-selected);
  border-color: var(--border-default);
}

.layer-sidebar__item-icon {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  color: var(--text-tertiary);
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.layer-sidebar__item-copy {
  display: grid;
  gap: 1px;
  min-width: 0;
}

.layer-sidebar__item-title {
  overflow: hidden;
  color: inherit;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.layer-sidebar__item-desc {
  overflow: hidden;
  color: var(--text-tertiary);
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.layer-sidebar__count {
  min-width: 1.75rem;
  padding: 1px var(--space-2);
  color: var(--text-tertiary);
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-pill);
  font-size: var(--font-size-xs);
  font-variant-numeric: tabular-nums;
  text-align: center;
}

.layer-sidebar__empty {
  padding: var(--space-4);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  text-align: center;
}

.layer-sidebar__reset {
  justify-self: start;
  padding: 0;
  color: var(--text-tertiary);
  background: transparent;
  border: 0;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  text-align: left;
}

.layer-sidebar__reset:hover {
  color: var(--text-link);
  text-decoration: underline;
}

.layer-sidebar__source {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.layer-sidebar__source-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--state-success);
}

.layer-sidebar__source-dot.is-error {
  background: var(--state-danger);
}

@media (max-width: 900px) {
  .layer-sidebar {
    padding: var(--space-6) var(--space-4);
    border-right: 0;
    border-bottom: 0;
    height: auto;
    align-content: start;
    overflow-y: auto;
  }

  .layer-sidebar__nav {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 640px) {
  .layer-sidebar {
    gap: var(--space-3);
    padding: var(--space-5) var(--space-3);
  }
}

.layer-sidebar__separator {
  border: 0;
  border-top: 1px solid var(--border-subtle);
  margin: var(--space-4) 0;
}
</style>
