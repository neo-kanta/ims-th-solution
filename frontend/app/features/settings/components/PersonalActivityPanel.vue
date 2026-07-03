<script setup lang="ts">
import type { PersonalActivityEvent } from "../personal.types";

defineProps<{
  events: PersonalActivityEvent[];
  loading: boolean;
  error: string | null;
  formatDateTime: (value?: string | null) => string;
}>();
</script>

<template>
  <section class="settings-panel">
    <header class="settings-panel__header">
      <div>
        <h2 class="settings-panel__title">My Activity</h2>
        <p class="settings-panel__subtitle">
          Read-only personal security, settings, workflow, and approval activity.
        </p>
      </div>
      <span class="badge badge-warning">API pending</span>
    </header>

    <div v-if="error" class="alert alert-danger settings-panel__alert" role="alert">
      {{ error }}
    </div>

    <div class="personal-activity" :aria-busy="loading">
      <article v-for="event in events" :key="event.id" class="personal-activity-row">
        <span class="personal-activity-row__marker" aria-hidden="true" />
        <div>
          <div class="personal-settings-primary">{{ event.title }}</div>
          <div class="personal-settings-secondary">{{ event.description }}</div>
          <div class="personal-settings-secondary">
            {{ formatDateTime(event.createdAt) }}
            <span v-if="event.ipAddress"> / {{ event.ipAddress }}</span>
          </div>
        </div>
        <span class="badge badge-neutral">{{ event.category }}</span>
      </article>

      <AppEmptyState
        v-if="!events.length"
        title="No personal activity feed"
        description="A self-scoped /me/activity API is not available yet. Admin audit logs are intentionally not used here."
        icon="lock"
      />
    </div>
  </section>
</template>

<style scoped>
.personal-activity {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-5);
}

.personal-activity-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: var(--space-4);
  align-items: start;
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.personal-activity-row__marker {
  width: 0.625rem;
  height: 0.625rem;
  margin-top: 0.375rem;
  border-radius: var(--radius-pill);
  background: var(--action-primary);
}

.personal-settings-primary {
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
}

.personal-settings-secondary {
  margin-top: var(--space-1);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-relaxed);
}

@media (max-width: 640px) {
  .personal-activity-row {
    grid-template-columns: auto minmax(0, 1fr);
  }
}
</style>
