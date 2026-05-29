<script setup lang="ts">
/**
 * History timeline — most recent first, with an action filter and basic
 * pagination ("show more"). The timeline reads transitions verbatim from
 * the backend; do not reorder server-side fields here.
 */
import { computed, ref } from "vue";

import { ACTION_CATALOG } from "../permissions";
import { WORKFLOW_ACTIONS } from "../types";
import type {
  WorkflowAction,
  WorkflowTransitionEntry,
} from "../types";

const PAGE_SIZE = 10;

const props = defineProps<{
  entries: readonly WorkflowTransitionEntry[];
  loading: boolean;
  labels: {
    title: string;
    filterAll: string;
    loading: string;
    empty: string;
    showMore: string;
  };
}>();

const actionFilter = ref<WorkflowAction | "">("");
const showAll = ref(false);

const filtered = computed(() => {
  const all = [...props.entries];
  all.sort((a, b) => {
    const ta = a.occurredAt ? Date.parse(a.occurredAt) : 0;
    const tb = b.occurredAt ? Date.parse(b.occurredAt) : 0;
    return tb - ta;
  });
  if (!actionFilter.value) return all;
  return all.filter((e) => e.action === actionFilter.value);
});

const visible = computed(() =>
  showAll.value ? filtered.value : filtered.value.slice(0, PAGE_SIZE),
);

const canShowMore = computed(
  () => !showAll.value && filtered.value.length > PAGE_SIZE,
);

function actionLabel(action: string | undefined): string {
  if (!action) return "";
  const opt = (ACTION_CATALOG as Record<string, { labelFallback: string }>)[
    action
  ];
  return opt?.labelFallback ?? action;
}

function formatTimestamp(value: string | null | undefined): string {
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
  <section class="workflow-history" data-testid="workflow-history">
    <header class="workflow-history__header">
      <h3 class="workflow-history__title">{{ labels.title }}</h3>
      <select
        v-model="actionFilter"
        class="workflow-history__filter"
        data-testid="workflow-history-filter"
      >
        <option value="">{{ labels.filterAll }}</option>
        <option v-for="action in WORKFLOW_ACTIONS" :key="action" :value="action">
          {{ actionLabel(action) }}
        </option>
      </select>
      <span v-if="loading" class="workflow-history__loading">
        {{ labels.loading }}
      </span>
    </header>

    <ol v-if="visible.length" class="workflow-history__timeline">
      <li
        v-for="entry in visible"
        :key="entry.id"
        class="workflow-history__item"
      >
        <div class="workflow-history__time">
          {{ formatTimestamp(entry.occurredAt) }}
        </div>
        <div class="workflow-history__body">
          <div class="workflow-history__headline">
            <strong>{{ actionLabel(entry.action) }}</strong>
            <span class="workflow-history__arrow">
              {{ entry.fromState }} → {{ entry.toState }}
            </span>
          </div>
          <div class="workflow-history__meta">
            {{ entry.actorUsername || entry.actorType || "—" }}
            <span v-if="entry.reason"> · {{ entry.reason }}</span>
          </div>
        </div>
      </li>
    </ol>

    <p
      v-else-if="!loading"
      class="workflow-history__empty"
      data-testid="workflow-history-empty"
    >
      {{ labels.empty }}
    </p>

    <button
      v-if="canShowMore"
      type="button"
      class="workflow-history__more"
      @click="showAll = true"
    >
      {{ labels.showMore }}
    </button>
  </section>
</template>

<style scoped>
.workflow-history {
  display: grid;
  gap: var(--space-2);
}

.workflow-history__header {
  display: flex;
  align-items: baseline;
  gap: var(--space-3);
}

.workflow-history__title {
  margin: 0;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.workflow-history__filter {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--space-1) var(--space-2);
  font-size: 12px;
}

.workflow-history__loading,
.workflow-history__empty {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.workflow-history__timeline {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-2);
}

.workflow-history__item {
  display: grid;
  grid-template-columns: 10rem 1fr;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  border-left: 2px solid var(--border-subtle);
}

.workflow-history__time {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 12px;
  color: var(--text-tertiary);
}

.workflow-history__headline {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
  font-size: var(--font-size-sm);
}

.workflow-history__arrow {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 12px;
  color: var(--text-tertiary);
}

.workflow-history__meta {
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

.workflow-history__more {
  background: none;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--space-1) var(--space-3);
  font-size: var(--font-size-xs);
  color: var(--action-primary);
  cursor: pointer;
}

@media (max-width: 640px) {
  .workflow-history__item {
    grid-template-columns: 1fr;
  }
}
</style>
