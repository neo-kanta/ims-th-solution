<script setup lang="ts">
import type { NotificationPreference } from "../ui.types";

defineProps<{
  preferences: NotificationPreference[];
}>();

const { t } = useI18n();

function severityClass(severity: NotificationPreference["severity"]) {
  if (severity === "Critical" || severity === "High") {
    return "badge-warning";
  }

  if (severity === "Medium") {
    return "badge-info";
  }

  return "badge-neutral";
}
</script>

<template>
  <section class="settings-panel settings-notifications" aria-labelledby="settings-notifications-title">
    <header class="settings-panel__header">
      <div>
        <h2 id="settings-notifications-title" class="settings-panel__title">
          {{ t("settings.console.notifications.title") }}
        </h2>
        <p class="settings-panel__subtitle">
          {{ t("settings.console.notifications.subtitle") }}
        </p>
      </div>
      <span class="badge badge-warning">
        {{ t("settings.console.common.readOnlyScaffold") }}
      </span>
    </header>

    <div class="settings-notifications__body">
      <article
        v-for="preference in preferences"
        :key="preference.id"
        class="settings-notification-rule"
      >
        <div class="settings-notification-rule__main">
          <div class="settings-record-primary">{{ preference.event }}</div>
          <div class="settings-record-secondary">{{ preference.audience }}</div>
        </div>

        <div
          class="settings-chip-list"
          :aria-label="t('settings.console.notifications.channelsLabel')"
        >
          <span
            v-for="channel in preference.channels"
            :key="`${preference.id}-${channel}`"
            class="badge badge-neutral"
          >
            {{ channel }}
          </span>
        </div>

        <span class="badge" :class="severityClass(preference.severity)">
          {{ preference.severity }}
        </span>

        <span
          class="badge"
          :class="preference.enabled ? 'badge-success' : 'badge-neutral'"
        >
          {{
            preference.enabled
              ? t("settings.console.common.enabled")
              : t("settings.console.common.disabled")
          }}
        </span>
      </article>
    </div>
  </section>
</template>

<style scoped>
.settings-notifications__body {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-5);
}

.settings-notification-rule {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(10rem, auto) auto auto;
  gap: var(--space-4);
  align-items: center;
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.settings-notification-rule__main {
  min-width: 0;
}

.settings-chip-list {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.settings-record-primary {
  display: block;
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
}

.settings-record-secondary {
  display: block;
  margin-top: var(--space-1);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

@media (max-width: 900px) {
  .settings-notification-rule {
    grid-template-columns: 1fr;
    align-items: start;
  }
}
</style>
