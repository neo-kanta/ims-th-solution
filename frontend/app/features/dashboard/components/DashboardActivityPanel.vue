<script setup lang="ts">
import type { DashboardOverviewActivityItem } from "../types";

const { t } = useI18n();

defineProps<{
  items: DashboardOverviewActivityItem[];
}>();
</script>

<template>
  <section class="dashboard-side-panel">
    <div class="dashboard-side-panel__header">
      <h2 class="dashboard-side-panel__title">
        {{ t("dashboardOverview.activityFeedTitle", "Activity Feed") }}
      </h2>
    </div>

    <ol class="activity-feed">
      <li v-for="item in items" :key="item.id" class="activity-feed__item">
        <span class="activity-feed__rail" aria-hidden="true" />
        <span class="activity-feed__dot" :class="`is-${item.tone}`" />

        <div class="activity-feed__content">
          <div class="activity-feed__topline">
            <div class="activity-feed__actor">{{ item.actor }}</div>
            <div class="activity-feed__time">{{ item.timeLabel }}</div>
          </div>
          <p class="activity-feed__message">{{ item.message }}</p>
        </div>
      </li>
    </ol>
  </section>
</template>

<style scoped>
.dashboard-side-panel {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-xl);
  background: var(--bg-card);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

.dashboard-side-panel__header {
  padding: var(--space-5) var(--space-5) var(--space-3);
}

.dashboard-side-panel__title {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  letter-spacing: -0.02em;
}

.activity-feed {
  margin: 0;
  padding: 0 var(--space-5) var(--space-5);
  list-style: none;
}

.activity-feed__item {
  position: relative;
  padding: var(--space-4) 0 var(--space-4) 1.15rem;
}

.activity-feed__item + .activity-feed__item {
  border-top: 1px solid var(--border-subtle);
}

.activity-feed__rail {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0.35rem;
  width: 1px;
  background: var(--border-subtle);
}

.activity-feed__dot {
  position: absolute;
  top: 1.15rem;
  left: 0.05rem;
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: var(--border-strong);
}

.activity-feed__dot.is-success {
  background: var(--state-success);
}

.activity-feed__dot.is-info {
  background: var(--action-primary);
}

.activity-feed__dot.is-warning {
  background: var(--state-warning);
}

.activity-feed__dot.is-teal {
  background: #0f766e;
}

:root[data-theme="dark"] .activity-feed__dot.is-teal {
  background: #14b8a6;
}

.activity-feed__dot.is-neutral {
  background: var(--border-strong);
}

.activity-feed__content {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
}

.activity-feed__topline {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.activity-feed__actor {
  color: var(--text-primary);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  line-height: var(--line-height-tight);
}

.activity-feed__message {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
  line-height: 1.45;
}

.activity-feed__time {
  color: var(--text-tertiary);
  font-size: var(--font-size-sm);
  line-height: var(--line-height-tight);
  white-space: nowrap;
}

@media (max-width: 640px) {
  .dashboard-side-panel__header {
    padding: var(--space-4) var(--space-4) var(--space-3);
  }

  .activity-feed {
    padding: 0 var(--space-4) var(--space-4);
  }

  .activity-feed__item {
    padding-left: 1rem;
  }

  .activity-feed__topline {
    gap: var(--space-3);
  }

  .activity-feed__actor,
  .activity-feed__time,
  .activity-feed__message {
    font-size: var(--font-size-sm);
  }
}
</style>
