<script setup lang="ts">
import type { DashboardOverviewWorkflowStage } from "../types";

defineProps<{
  stages: DashboardOverviewWorkflowStage[];
}>();
</script>

<template>
  <section class="workflow-rail" aria-label="Dashboard workflow">
    <div class="workflow-rail__scroll">
      <ol class="workflow-rail__list">
        <li
          v-for="(stage, index) in stages"
          :key="stage.id"
          class="workflow-rail__item"
        >
          <span
            v-if="index < stages.length - 1"
            class="workflow-rail__line"
            :class="{ 'is-complete': stage.status === 'complete' }"
          />

          <span
            class="workflow-rail__dot"
            :class="`is-${stage.status}`"
            aria-hidden="true"
          />

          <div class="workflow-rail__copy">
            <div class="workflow-rail__label">
              {{ stage.sequence }}. {{ stage.label }}
            </div>
            <div class="workflow-rail__time">{{ stage.timeLabel }}</div>
          </div>
        </li>
      </ol>
    </div>
  </section>
</template>

<style scoped>
.workflow-rail {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-xl);
  background: var(--bg-card);
  box-shadow: var(--shadow-sm);
}

.workflow-rail__scroll {
  overflow-x: auto;
  padding: var(--space-5) var(--space-6);
}

.workflow-rail__list {
  display: grid;
  grid-template-columns: repeat(7, minmax(7rem, 1fr));
  gap: var(--space-4);
  min-width: 52rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.workflow-rail__item {
  position: relative;
  display: grid;
  justify-items: center;
  gap: var(--space-3);
  min-width: 0;
}

.workflow-rail__line {
  position: absolute;
  top: 0.45rem;
  left: calc(50% + 0.6rem);
  width: calc(100% - 0.75rem);
  height: 2px;
  background: var(--border-subtle);
}

.workflow-rail__line.is-complete {
  background: var(--state-success);
}

.workflow-rail__dot {
  position: relative;
  z-index: 1;
  width: 14px;
  height: 14px;
  border-radius: 999px;
  background: var(--color-neutral-200);
}

.workflow-rail__dot.is-complete {
  background: var(--state-success);
}

.workflow-rail__dot.is-active {
  background: var(--action-primary);
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.12);
}

.workflow-rail__dot.is-upcoming {
  background: var(--color-neutral-200);
}

.workflow-rail__copy {
  display: grid;
  gap: 2px;
  text-align: center;
}

.workflow-rail__label {
  color: var(--text-primary);
  font-size: 11px;
  font-weight: var(--font-weight-medium);
  line-height: 1.3;
}

.workflow-rail__time {
  color: var(--text-tertiary);
  font-size: 10px;
  line-height: 1.2;
}

@media (max-width: 640px) {
  .workflow-rail__scroll {
    padding: var(--space-4);
  }

  .workflow-rail__list {
    min-width: 46rem;
    gap: var(--space-3);
  }
}
</style>
