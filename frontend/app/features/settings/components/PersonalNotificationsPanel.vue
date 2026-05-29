<script setup lang="ts">
import type { PersonalNotificationPreference } from "../personal.types";

defineProps<{
  preferences: PersonalNotificationPreference[];
  loading: boolean;
  error: string | null;
}>();

const emit = defineEmits<{
  toggle: [id: string, channel: "inApp" | "email", value: boolean];
  refresh: [];
}>();
</script>

<template>
  <section class="settings-panel">
    <header class="settings-panel__header">
      <div>
        <h2 class="settings-panel__title">Notifications</h2>
        <p class="settings-panel__subtitle">
          Personal channel preferences only. Administrative notification rules and templates remain controlled by administrators.
        </p>
      </div>
      <AppButton variant="secondary" size="sm" :loading="loading" @click="emit('refresh')">
        <AppIcon name="refresh" size="xs" />
        <span>Refresh</span>
      </AppButton>
    </header>

    <div v-if="error" class="alert alert-danger settings-panel__alert" role="alert">
      {{ error }}
    </div>

    <div class="personal-settings-table-wrap" :aria-busy="loading">
      <table class="personal-settings-table">
        <thead>
          <tr>
            <th>Event</th>
            <th>Severity</th>
            <th>In-app</th>
            <th>Email</th>
            <th>Policy</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="preference in preferences" :key="preference.id">
            <td>
              <div class="personal-settings-primary">{{ preference.label }}</div>
              <div class="personal-settings-secondary">{{ preference.description }}</div>
            </td>
            <td>
              <span class="badge" :class="preference.severity === 'Critical' || preference.severity === 'High' ? 'badge-warning' : 'badge-neutral'">
                {{ preference.severity }}
              </span>
            </td>
            <td>
              <label class="personal-toggle">
                <input
                  type="checkbox"
                  :checked="preference.inApp"
                  :disabled="preference.mandatory"
                  @change="emit('toggle', preference.id, 'inApp', ($event.target as HTMLInputElement).checked)"
                />
                <span>In-app</span>
              </label>
            </td>
            <td>
              <label class="personal-toggle">
                <input
                  type="checkbox"
                  :checked="preference.email"
                  :disabled="preference.mandatory"
                  @change="emit('toggle', preference.id, 'email', ($event.target as HTMLInputElement).checked)"
                />
                <span>Email</span>
              </label>
            </td>
            <td>
              <span v-if="preference.mandatory" class="personal-settings-secondary">
                {{ preference.mandatoryReason }}
              </span>
              <span v-else class="badge badge-info">User configurable</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <footer class="settings-panel__footer">
      <span>Stored locally until the user notification preference API is available.</span>
    </footer>
  </section>
</template>

<style scoped>
.personal-settings-table-wrap {
  overflow-x: auto;
}

.personal-settings-table {
  width: 100%;
  min-width: 760px;
  border-collapse: collapse;
}

.personal-settings-table th,
.personal-settings-table td {
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--border-subtle);
  text-align: left;
  vertical-align: top;
}

.personal-settings-table th {
  color: var(--text-tertiary);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.personal-settings-table tbody tr:last-child td {
  border-bottom: 0;
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

.personal-toggle {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  white-space: nowrap;
}
</style>
