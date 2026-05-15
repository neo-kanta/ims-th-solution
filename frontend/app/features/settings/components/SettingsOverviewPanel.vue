<script setup lang="ts">
import type {
  SettingsApiCapability,
  SettingsKpiMetric,
  SettingsOverviewSignal,
} from "../ui.types";
import SettingsKpiGrid from "./SettingsKpiGrid.vue";

defineProps<{
  metrics: SettingsKpiMetric[];
  loading: boolean;
  signals: SettingsOverviewSignal[];
  capabilities: SettingsApiCapability[];
  appName: string;
  apiBaseUrl: string;
  timezone: string;
}>();

const { t } = useI18n();
</script>

<template>
  <section class="settings-overview" aria-labelledby="settings-overview-title">
    <div class="settings-overview__header">
      <div>
        <h2 id="settings-overview-title" class="settings-overview__title">
          {{ t("settings.console.overview.title") }}
        </h2>
        <p class="settings-overview__subtitle">
          {{ t("settings.console.overview.subtitle") }}
        </p>
      </div>
      <dl class="settings-overview__environment" aria-label="Current environment">
        <div>
          <dt>{{ t("settings.console.overview.application") }}</dt>
          <dd>{{ appName }}</dd>
        </div>
        <div>
          <dt>{{ t("settings.console.overview.apiBase") }}</dt>
          <dd>{{ apiBaseUrl }}</dd>
        </div>
        <div>
          <dt>{{ t("settings.console.overview.timezone") }}</dt>
          <dd>{{ timezone }}</dd>
        </div>
      </dl>
    </div>

    <SettingsKpiGrid :metrics="metrics" :loading="loading" />

    <div class="settings-overview__signals">
      <article
        v-for="signal in signals"
        :key="signal.id"
        class="settings-overview-signal"
        :class="`settings-overview-signal--${signal.tone}`"
      >
        <div class="settings-overview-signal__label">{{ signal.label }}</div>
        <div class="settings-overview-signal__value">{{ signal.value }}</div>
        <div class="settings-overview-signal__helper">{{ signal.helper }}</div>
      </article>
    </div>

    <div class="settings-overview__capability-panel">
      <div class="settings-overview__section-heading">
        <h3>{{ t("settings.console.overview.apiMapTitle") }}</h3>
        <p>{{ t("settings.console.overview.apiMapSubtitle") }}</p>
      </div>

      <div class="settings-overview__capabilities">
        <article
          v-for="capability in capabilities"
          :key="capability.id"
          class="settings-capability"
        >
          <div>
            <div class="settings-capability__label">{{ capability.label }}</div>
            <div class="settings-capability__copy">{{ capability.capability }}</div>
          </div>
          <span
            class="badge"
            :class="{
              'badge-success': capability.status === 'live',
              'badge-info': capability.status === 'read-only',
              'badge-warning': capability.status === 'not-exposed',
            }"
          >
            {{
              capability.status === "live"
                ? t("settings.console.overview.liveApi")
                : capability.status === "read-only"
                  ? t("settings.console.nav.readOnly")
                  : t("settings.console.nav.apiPending")
            }}
          </span>
        </article>
      </div>
    </div>
  </section>
</template>

<style scoped>
.settings-overview {
  display: grid;
  gap: var(--space-5);
}

.settings-overview__header {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  grid-template-rows: auto auto;
  gap: var(--space-5);
  align-items: start;
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  box-shadow: var(--shadow-xs);
}

.settings-overview__title {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
}

.settings-overview__subtitle {
  margin: var(--space-2) 0 0;
  color: var(--text-secondary);
}

.settings-overview__environment {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-3);
  margin: 0;
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
}

.settings-overview__environment div {
  min-width: 0;
}

.settings-overview__environment dt {
  color: var(--text-tertiary);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.settings-overview__environment dd {
  margin: var(--space-1) 0 0;
  color: var(--text-primary);
  overflow-wrap: anywhere;
}

.settings-overview__signals {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-4);
}

.settings-overview-signal {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.settings-overview-signal__label {
  color: var(--text-tertiary);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.settings-overview-signal__value {
  color: var(--text-primary);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
}

.settings-overview-signal__helper {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-relaxed);
}

.settings-overview-signal--success {
  border-left: 4px solid var(--state-success);
}

.settings-overview-signal--warning {
  border-left: 4px solid var(--state-warning);
}

.settings-overview-signal--danger {
  border-left: 4px solid var(--state-danger);
}

.settings-overview-signal--info {
  border-left: 4px solid var(--state-info);
}

.settings-overview-signal--neutral {
  border-left: 4px solid var(--border-strong);
}

.settings-overview__capability-panel {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  box-shadow: var(--shadow-xs);
}

.settings-overview__section-heading h3 {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-md);
}

.settings-overview__section-heading p {
  margin: var(--space-1) 0 0;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.settings-overview__capabilities {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3);
}

.settings-capability {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
}

.settings-capability__label {
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
}

.settings-capability__copy {
  margin-top: var(--space-1);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-relaxed);
}

@media (max-width: 1180px) {
  .settings-overview__signals,
  .settings-overview__capabilities {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .settings-overview__header,
  .settings-overview__environment,
  .settings-overview__signals,
  .settings-overview__capabilities {
    grid-template-columns: 1fr;
  }

  .settings-capability {
    flex-direction: column;
  }
}
</style>
