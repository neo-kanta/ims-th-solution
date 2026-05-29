<script setup lang="ts">
import type { PersonalLeaveRequest } from "../personal.types";

defineProps<{
  requests: PersonalLeaveRequest[];
  loading: boolean;
  error: string | null;
}>();
</script>

<template>
  <section class="settings-panel">
    <header class="settings-panel__header">
      <div>
        <h2 class="settings-panel__title">Leave & Delegation</h2>
        <p class="settings-panel__subtitle">
          Self-service leave, temporary leave, cancellation, and delegation visibility for the signed-in user.
        </p>
      </div>
      <span class="badge badge-warning">API pending</span>
    </header>

    <div v-if="error" class="alert alert-danger settings-panel__alert" role="alert">
      {{ error }}
    </div>

    <div class="personal-leave-grid" :aria-busy="loading">
      <section class="personal-subpanel">
        <h3>Current requests</h3>
        <div v-if="requests.length" class="personal-request-list">
          <article v-for="request in requests" :key="request.id" class="personal-request">
            <div>
              <div class="personal-settings-primary">{{ request.type }}</div>
              <div class="personal-settings-secondary">
                {{ request.startDate }} to {{ request.endDate }}
              </div>
            </div>
            <span class="badge badge-neutral">{{ request.status }}</span>
          </article>
        </div>
        <AppEmptyState
          v-else
          title="No leave requests"
          description="No personal leave requests were returned. The backend self-service leave API is not available yet."
          icon="folder"
        />
      </section>

      <section class="personal-subpanel">
        <h3>Create request</h3>
        <form class="personal-leave-form">
          <label class="form-group">
            <span class="label">Request type</span>
            <select class="input" disabled>
              <option>Leave</option>
              <option>Temporary leave</option>
              <option>Cancellation</option>
            </select>
          </label>
          <label class="form-group">
            <span class="label">Date range</span>
            <input class="input" disabled placeholder="Select dates" />
          </label>
          <label class="form-group">
            <span class="label">Reason</span>
            <textarea class="input" disabled rows="4" placeholder="Required by leave policy" />
          </label>
          <AppButton variant="secondary" size="sm" disabled>
            Submission waits for /me/leave-requests API
          </AppButton>
        </form>
      </section>
    </div>
  </section>
</template>

<style scoped>
.personal-leave-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(20rem, 0.8fr);
  gap: var(--space-4);
  padding: var(--space-5);
}

.personal-subpanel {
  display: grid;
  gap: var(--space-4);
  align-content: start;
  min-width: 0;
}

.personal-subpanel h3 {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-md);
}

.personal-request-list,
.personal-leave-form {
  display: grid;
  gap: var(--space-3);
}

.personal-request {
  display: flex;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.personal-settings-primary {
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
}

.personal-settings-secondary {
  margin-top: var(--space-1);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

@media (max-width: 860px) {
  .personal-leave-grid {
    grid-template-columns: 1fr;
  }
}
</style>
