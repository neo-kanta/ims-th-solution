<script setup lang="ts">
import type { SettingsNavigationItem, SettingsSectionId } from "../ui.types";

defineProps<{
  items: SettingsNavigationItem[];
  activeSection: SettingsSectionId;
}>();

const emit = defineEmits<{
  select: [sectionId: SettingsSectionId];
}>();

const { t } = useI18n();
</script>

<template>
  <nav
    class="settings-section-nav"
    :aria-label="t('settings.console.nav.sectionsLabel')"
  >
    <button
      v-for="item in items"
      :key="item.id"
      class="settings-section-nav__item"
      :class="{ 'is-active': activeSection === item.id }"
      type="button"
      :disabled="item.disabled"
      :aria-current="activeSection === item.id ? 'page' : undefined"
      @click="emit('select', item.id)"
    >
      <span class="settings-section-nav__icon" aria-hidden="true">
        <AppIcon :name="item.icon" size="sm" />
      </span>
      <span class="settings-section-nav__copy">
        <span class="settings-section-nav__label">{{ item.label }}</span>
        <span class="settings-section-nav__description">
          {{ item.description }}
        </span>
      </span>
      <span class="settings-section-nav__meta">
        <span
          class="settings-section-nav__status"
          :class="`settings-section-nav__status--${item.status}`"
        >
          {{
            item.status === "live"
              ? t("settings.console.nav.live")
              : item.status === "read-only"
                ? t("settings.console.nav.readOnly")
                : t("settings.console.nav.apiPending")
          }}
        </span>
        <span v-if="item.count !== undefined" class="badge badge-neutral">
          {{ item.count }}
        </span>
      </span>
    </button>
  </nav>
</template>

<style scoped>
.settings-section-nav {
  position: sticky;
  top: calc(var(--header-height) + var(--space-5));
  display: grid;
  gap: var(--space-2);
  align-self: start;
  max-height: calc(100vh - var(--header-height) - var(--space-8));
  overflow-y: auto;
  padding: var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  box-shadow: var(--shadow-xs);
}

.settings-section-nav__item {
  width: 100%;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-3);
  padding: var(--space-4);
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--text-secondary);
  text-align: left;
  transition:
    background var(--transition-fast),
    border-color var(--transition-fast),
    color var(--transition-fast);
}

.settings-section-nav__item:hover {
  border-color: var(--border-subtle);
  background: var(--bg-card-hover);
  color: var(--text-primary);
}

.settings-section-nav__item.is-active {
  border-color: var(--border-focus);
  background: var(--bg-selected);
  color: var(--action-primary);
}

.settings-section-nav__item:disabled {
  cursor: not-allowed;
  opacity: 0.58;
}

.settings-section-nav__icon {
  width: 2rem;
  height: 2rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
  color: currentColor;
}

.settings-section-nav__copy {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.settings-section-nav__label {
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
}

.settings-section-nav__item.is-active .settings-section-nav__label {
  color: var(--action-primary);
}

.settings-section-nav__description {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-snug);
}

.settings-section-nav__meta {
  grid-column: 2;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.settings-section-nav__status {
  color: var(--text-tertiary);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.settings-section-nav__status--live {
  color: var(--state-success);
}

.settings-section-nav__status--pending {
  color: var(--state-warning);
}

@media (max-width: 1024px) {
  .settings-section-nav {
    position: static;
    display: flex;
    overflow-x: auto;
    overflow-y: hidden;
    max-height: none;
    padding-bottom: var(--space-2);
  }

  .settings-section-nav__item {
    width: min(18rem, 78vw);
    flex: 0 0 auto;
  }
}
</style>
