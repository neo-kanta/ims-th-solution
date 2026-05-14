<script setup lang="ts">
import type { SecurityPolicyModel, SecurityPolicySetting } from "../ui.types";

defineProps<{
  policy: SecurityPolicyModel;
}>();

const { t } = useI18n();

function displayValue(setting: SecurityPolicySetting) {
  if (typeof setting.value === "boolean") {
    return setting.value
      ? t("settings.console.common.enabled")
      : t("settings.console.common.disabled");
  }

  return setting.unit
    ? `${setting.value} ${setting.unit}`
    : String(setting.value);
}
</script>

<template>
  <section
    class="settings-panel settings-security"
    aria-labelledby="settings-security-title"
  >
    <header class="settings-panel__header">
      <div>
        <h2 id="settings-security-title" class="settings-panel__title">
          {{ t("settings.console.securityPolicy.title") }}
        </h2>
        <p class="settings-panel__subtitle">
          {{ t("settings.console.securityPolicy.subtitle") }}
        </p>
      </div>
      <span class="badge badge-warning">
        {{ t("settings.console.common.apiPending") }}
      </span>
    </header>

    <div class="alert alert-warning settings-panel__alert" role="note">
      <strong>{{ t("settings.console.common.sampleDataTitle") }}</strong>
      <span>{{ t("settings.console.securityPolicy.sampleNotice") }}</span>
    </div>

    <div class="settings-security__body">
      <div class="settings-security__mode">
        <div>
          <div class="settings-security__mode-label">
            {{ t("settings.console.securityPolicy.authMode") }}
          </div>
          <div class="settings-security__mode-value">{{ policy.authMode }}</div>
          <p>
            {{ t("settings.console.securityPolicy.authModeHelper") }}
          </p>
        </div>
        <span
          class="badge"
          :class="
            policy.authMode === 'AD / SSO' ? 'badge-info' : 'badge-neutral'
          "
        >
          {{
            policy.authMode === "AD / SSO"
              ? t("settings.console.securityPolicy.externalAuthority")
              : t("settings.console.securityPolicy.internalAuthority")
          }}
        </span>
      </div>

      <div class="settings-security__grid">
        <article
          v-for="setting in policy.settings"
          :key="setting.id"
          class="settings-security-setting"
        >
          <label :for="`settings-security-${setting.id}`" class="label">
            {{ setting.label }}
          </label>

          <div
            v-if="typeof setting.value === 'boolean'"
            class="settings-security-toggle"
          >
            <input
              :id="`settings-security-${setting.id}`"
              type="checkbox"
              :checked="setting.value"
              disabled
            />
            <span>{{ displayValue(setting) }}</span>
          </div>

          <input
            v-else
            :id="`settings-security-${setting.id}`"
            class="input"
            :value="displayValue(setting)"
            disabled
          />

          <p class="help-text">{{ setting.helper }}</p>
        </article>
      </div>
    </div>
  </section>
</template>

<style scoped>
.settings-security__body {
  display: grid;
  gap: var(--space-5);
  padding: var(--space-5);
}

.settings-security__mode {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
}

.settings-security__mode-label {
  color: var(--text-tertiary);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.settings-security__mode-value {
  margin-top: var(--space-1);
  color: var(--text-primary);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
}

.settings-security__mode p {
  margin: var(--space-2) 0 0;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.settings-security__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-4);
}

.settings-security-setting {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.settings-security-toggle {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-height: var(--size-control-md);
  padding: 0 var(--space-4);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--bg-disabled);
  color: var(--text-disabled);
}

.settings-security-toggle input {
  width: 1rem;
  height: 1rem;
}

@media (max-width: 760px) {
  .settings-security__mode {
    flex-direction: column;
  }

  .settings-security__grid {
    grid-template-columns: 1fr;
  }
}
</style>
