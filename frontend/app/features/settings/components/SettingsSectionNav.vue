<script setup lang="ts">
import type {
  SettingsNavigationGroup,
  SettingsNavigationItem,
  SettingsSectionId,
} from "../ui.types";

withDefaults(defineProps<{
  groups: SettingsNavigationGroup[];
  activeSection: SettingsSectionId;
  accountName: string;
  accountSubtitle: string;
  accountInitials: string;
  showIdentity?: boolean;
}>(), {
  showIdentity: false,
});

const emit = defineEmits<{
  select: [sectionId: SettingsSectionId];
}>();

const { t } = useI18n();

function statusLabel(item: SettingsNavigationItem) {
  if (item.status === "live") {
    return t("settings.console.nav.live");
  }

  if (item.status === "read-only") {
    return t("settings.console.nav.readOnly");
  }

  return t("settings.console.nav.apiPending");
}

function itemAriaLabel(item: SettingsNavigationItem) {
  const count =
    item.count === undefined
      ? ""
      : `, ${item.count} ${t("settings.console.common.recorded")}`;

  return `${item.label}, ${item.description}, ${statusLabel(item)}${count}`;
}

function selectItem(item: SettingsNavigationItem) {
  if (!item.disabled) {
    emit("select", item.id);
  }
}
</script>

<template>
  <aside
    class="settings-section-nav"
    :class="{ 'settings-section-nav--menu-only': !showIdentity }"
  >
    <div v-if="showIdentity" class="settings-section-nav__identity">
      <span class="settings-section-nav__avatar" aria-hidden="true">
        {{ accountInitials }}
      </span>
      <span class="settings-section-nav__account">
        <span class="settings-section-nav__account-name">
          {{ accountName }}
        </span>
        <span class="settings-section-nav__account-subtitle">
          {{ accountSubtitle }}
        </span>
      </span>
    </div>

    <nav
      class="settings-section-nav__menu"
      :aria-label="t('settings.console.nav.sectionsLabel')"
    >
      <section
        v-for="group in groups"
        :key="group.id"
        class="settings-section-nav__group"
      >
        <h2 class="settings-section-nav__group-label">
          {{ group.label }}
        </h2>

        <div class="settings-section-nav__items">
          <button
            v-for="item in group.items"
            :key="item.id"
            class="settings-section-nav__item"
            :class="{ 'is-active': activeSection === item.id }"
            type="button"
            :disabled="item.disabled"
            :title="item.description"
            :aria-label="itemAriaLabel(item)"
            :aria-current="activeSection === item.id ? 'page' : undefined"
            @click="selectItem(item)"
          >
            <span class="settings-section-nav__icon" aria-hidden="true">
              <AppIcon :name="item.icon" size="sm" />
            </span>
            <span class="settings-section-nav__label">
              {{ item.label }}
            </span>
            <span v-if="item.count !== undefined" class="badge badge-neutral">
              {{ item.count }}
            </span>
          </button>
        </div>
      </section>
    </nav>
  </aside>
</template>

<style scoped>
.settings-section-nav {
  position: sticky;
  top: calc(var(--header-height) + var(--space-5));
  display: grid;
  gap: var(--space-4);
  align-self: start;
  max-height: calc(100vh - var(--header-height) - var(--space-8));
  overflow-y: auto;
  padding-right: var(--space-2);
}

.settings-section-nav--menu-only {
  gap: var(--space-3);
}

.settings-section-nav__identity {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-width: 0;
  padding: 0 0 var(--space-4);
  border-bottom: 1px solid var(--border-subtle);
}

.settings-section-nav__avatar {
  width: 3rem;
  height: 3rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-pill);
  background: var(--bg-selected);
  color: var(--action-primary);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
}

.settings-section-nav__account {
  display: grid;
  gap: 2px;
  min-width: 0;
}

.settings-section-nav__account-name {
  overflow: hidden;
  color: var(--text-primary);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.settings-section-nav__account-subtitle {
  overflow: hidden;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.settings-section-nav__menu {
  display: grid;
  gap: var(--space-4);
}

.settings-section-nav__group {
  display: grid;
  gap: var(--space-2);
}

.settings-section-nav__group + .settings-section-nav__group {
  padding-top: var(--space-4);
  border-top: 1px solid var(--border-subtle);
}

.settings-section-nav__group-label {
  margin: 0;
  padding: 0 var(--space-3);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
}

.settings-section-nav__items {
  display: grid;
  gap: 2px;
}

.settings-section-nav__item {
  position: relative;
  width: 100%;
  min-height: 2.25rem;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: 0 var(--space-3);
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-secondary);
  text-align: left;
  transition:
    background var(--transition-fast),
    border-color var(--transition-fast),
    color var(--transition-fast);
}

.settings-section-nav__item:hover {
  background: var(--bg-card-hover);
  color: var(--text-primary);
}

.settings-section-nav__item.is-active {
  background: var(--bg-card-hover);
  color: var(--action-primary);
  font-weight: var(--font-weight-semibold);
}

.settings-section-nav__item.is-active::before {
  content: "";
  position: absolute;
  left: 0;
  top: 50%;
  width: 3px;
  height: 1.25rem;
  border-radius: var(--radius-pill);
  background: currentColor;
  transform: translateY(-50%);
}

.settings-section-nav__item:disabled {
  cursor: not-allowed;
  opacity: 0.58;
}

.settings-section-nav__icon {
  width: 1rem;
  height: 1rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: currentColor;
}

.settings-section-nav__label {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  color: inherit;
  font-size: var(--font-size-sm);
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 1024px) {
  .settings-section-nav {
    position: static;
    grid-template-columns: minmax(14rem, 0.8fr) minmax(0, 1.2fr);
    align-items: start;
    overflow: visible;
    max-height: none;
    padding-right: 0;
  }

  .settings-section-nav--menu-only {
    grid-template-columns: 1fr;
  }

  .settings-section-nav__identity {
    padding-bottom: 0;
    border-bottom: 0;
  }

  .settings-section-nav__menu {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: var(--space-3);
  }

  .settings-section-nav__group + .settings-section-nav__group {
    padding-top: 0;
    border-top: 0;
  }
}

@media (max-width: 760px) {
  .settings-section-nav,
  .settings-section-nav__menu {
    grid-template-columns: 1fr;
  }
}
</style>
