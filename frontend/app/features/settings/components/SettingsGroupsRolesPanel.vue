<script setup lang="ts">
import { computed, ref, watch } from "vue";

import type { SettingsGroupRole } from "../ui.types";

const props = defineProps<{
  groups: SettingsGroupRole[];
  directoryComplete: boolean;
  loading: boolean;
}>();

const selectedGroupId = ref<string | null>(props.groups[0]?.id ?? null);

watch(
  () => props.groups,
  (groups) => {
    if (!groups.some((group) => group.id === selectedGroupId.value)) {
      selectedGroupId.value = groups[0]?.id ?? null;
    }
  },
  { immediate: true },
);

const selectedGroup = computed(
  () => props.groups.find((group) => group.id === selectedGroupId.value) ?? null,
);

function riskBadgeClass(riskLevel: SettingsGroupRole["riskLevel"]) {
  if (riskLevel === "restricted") {
    return "badge-error";
  }

  if (riskLevel === "elevated") {
    return "badge-warning";
  }

  return "badge-neutral";
}

const { t } = useI18n();

function riskBadgeLabel(group: SettingsGroupRole): string {
  // Risk level is meaningful only when the backend supplied it (demo source
  // ships explicit values). Directory-derived groups have no risk metadata yet.
  if (group.source === "directory") {
    return t("settings.console.groupsRoles.notClassified");
  }
  return group.riskLevel;
}
</script>

<template>
  <section class="settings-panel settings-groups" aria-labelledby="settings-groups-title">
    <header class="settings-panel__header">
      <div>
        <h2 id="settings-groups-title" class="settings-panel__title">
          {{ t("settings.console.groupsRoles.title") }}
        </h2>
        <p class="settings-panel__subtitle">
          {{ t("settings.console.groupsRoles.subtitle") }}
        </p>
      </div>
      <span class="badge" :class="directoryComplete ? 'badge-info' : 'badge-warning'">
        {{
          directoryComplete
            ? t("settings.console.groupsRoles.directoryDerived")
            : t("settings.console.groupsRoles.partialDirectory")
        }}
      </span>
    </header>

    <div class="settings-groups__body" :aria-busy="loading">
      <div
        class="settings-groups__list"
        role="list"
        :aria-label="t('settings.console.nav.groupsRoles')"
      >
        <button
          v-for="group in groups"
          :key="group.id"
          class="settings-group-row"
          :class="{ 'is-selected': selectedGroupId === group.id }"
          type="button"
          role="listitem"
          @click="selectedGroupId = group.id"
        >
          <span class="settings-group-row__main">
            <span class="settings-group-row__name">{{ group.name }}</span>
            <span class="settings-group-row__description">{{ group.description }}</span>
          </span>
          <span class="settings-group-row__meta">
            <span
              class="badge badge-neutral"
              :title="t('settings.console.groupsRoles.membersCountHint')"
            >
              {{ t("settings.console.groupsRoles.membersCount", { count: group.membersCount }) }}
            </span>
            <span class="badge" :class="riskBadgeClass(group.riskLevel)">
              {{ riskBadgeLabel(group) }}
            </span>
          </span>
        </button>

        <div v-if="groups.length === 0" class="settings-card-state">
          {{ t("settings.console.groupsRoles.noGroups") }}
        </div>
      </div>

      <aside
        class="settings-groups__detail"
        :aria-label="t('settings.console.groupsRoles.selectedDetailLabel')"
      >
        <template v-if="selectedGroup">
          <div class="settings-groups__detail-header">
            <div>
              <h3>{{ selectedGroup.name }}</h3>
              <p>{{ selectedGroup.description }}</p>
            </div>
            <span
              class="badge"
              :class="selectedGroup.source === 'directory' ? 'badge-info' : 'badge-warning'"
            >
              {{
                selectedGroup.source === "directory"
                  ? t("settings.console.groupsRoles.fromUsers")
                  : t("settings.console.groupsRoles.demoScaffold")
              }}
            </span>
          </div>

          <dl class="settings-groups__stats">
            <div>
              <dt
                :title="t('settings.console.groupsRoles.membersCountHint')"
              >
                {{ t("settings.console.groupsRoles.membersOnPage") }}
              </dt>
              <dd>{{ selectedGroup.membersCount }}</dd>
            </div>
            <div>
              <dt>{{ t("settings.console.groupsRoles.risk") }}</dt>
              <dd>{{ riskBadgeLabel(selectedGroup) }}</dd>
            </div>
          </dl>

          <div class="settings-groups__section">
            <h4>{{ t("settings.console.groupsRoles.functionalResponsibility") }}</h4>
            <ul v-if="selectedGroup.responsibilities.length">
              <li
                v-for="responsibility in selectedGroup.responsibilities"
                :key="responsibility"
              >
                {{ responsibility }}
              </li>
            </ul>
            <p v-else class="settings-groups__empty">
              {{ t("settings.console.groupsRoles.notConfigured") }}
            </p>
          </div>

          <div class="settings-groups__section">
            <h4>{{ t("settings.console.groupsRoles.permissionFamilies") }}</h4>
            <div v-if="selectedGroup.permissionFamilies.length" class="settings-chip-list">
              <span
                v-for="permission in selectedGroup.permissionFamilies"
                :key="permission"
                class="badge badge-neutral"
              >
                {{ permission }}
              </span>
            </div>
            <p v-else class="settings-groups__empty">
              {{ t("settings.console.groupsRoles.notConfigured") }}
            </p>
          </div>

          <div class="settings-groups__section">
            <h4>{{ t("settings.console.groupsRoles.dataScopes") }}</h4>
            <div v-if="selectedGroup.dataScopes.length" class="settings-chip-list">
              <span
                v-for="scope in selectedGroup.dataScopes"
                :key="scope"
                class="badge badge-info"
              >
                {{ scope }}
              </span>
            </div>
            <p v-else class="settings-groups__empty">
              {{ t("settings.console.groupsRoles.notConfigured") }}
            </p>
          </div>
        </template>

        <div v-else class="settings-card-state">
          {{ t("settings.console.groupsRoles.selectGroup") }}
        </div>
      </aside>
    </div>
  </section>
</template>

<style scoped>
.settings-groups__body {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(20rem, 0.72fr);
  gap: var(--space-5);
  padding: var(--space-5);
}

.settings-groups__list {
  display: grid;
  gap: var(--space-3);
}

.settings-group-row {
  width: 100%;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-4);
  align-items: center;
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  text-align: left;
  transition:
    background var(--transition-fast),
    border-color var(--transition-fast);
}

.settings-group-row:hover {
  border-color: var(--border-default);
  background: var(--bg-card-hover);
}

.settings-group-row.is-selected {
  border-color: var(--border-focus);
  background: var(--bg-selected);
}

.settings-group-row__main {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.settings-group-row__name {
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
}

.settings-group-row__description {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-relaxed);
}

.settings-group-row__meta {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.settings-groups__detail {
  display: grid;
  align-content: start;
  gap: var(--space-5);
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
}

.settings-groups__detail-header {
  display: flex;
  justify-content: space-between;
  gap: var(--space-4);
  align-items: flex-start;
}

.settings-groups__detail h3,
.settings-groups__section h4 {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-md);
}

.settings-groups__detail p {
  margin: var(--space-1) 0 0;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-relaxed);
}

.settings-groups__stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3);
  margin: 0;
}

.settings-groups__stats div {
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.settings-groups__stats dt {
  color: var(--text-tertiary);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
}

.settings-groups__stats dd {
  margin: var(--space-1) 0 0;
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
  text-transform: capitalize;
}

.settings-groups__section {
  display: grid;
  gap: var(--space-3);
}

.settings-groups__section ul {
  display: grid;
  gap: var(--space-2);
  margin: 0;
  padding-left: var(--space-5);
  color: var(--text-secondary);
}

.settings-chip-list {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.settings-card-state {
  padding: var(--space-6);
  border: 1px dashed var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  text-align: center;
}

@media (max-width: 1024px) {
  .settings-groups__body {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .settings-group-row {
    grid-template-columns: 1fr;
  }

  .settings-groups__detail-header {
    flex-direction: column;
  }

  .settings-group-row__meta {
    justify-content: flex-start;
  }
}

.settings-groups__empty {
  margin: 0;
  padding: var(--space-4);
  border: 1px dashed var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  text-align: center;
}
</style>
