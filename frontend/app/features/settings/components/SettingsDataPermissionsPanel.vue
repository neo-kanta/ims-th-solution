<script setup lang="ts">
import { computed, ref } from "vue";

import type { DataPermissionGrant } from "../ui.types";

const props = defineProps<{
  grants: DataPermissionGrant[];
  formatDateTime: (value?: string | null) => string;
}>();

type DataPermissionMode = "user" | "contract";

const mode = ref<DataPermissionMode>("user");
const search = ref("");
const { t } = useI18n();

const filteredGrants = computed(() => {
  const query = search.value.trim().toLowerCase();

  if (!query) {
    return props.grants;
  }

  return props.grants.filter((grant) => {
    const haystack = [
      grant.user,
      grant.userLabel,
      grant.contractId,
      grant.contractName,
      grant.scope,
    ].join(" ").toLowerCase();

    return haystack.includes(query);
  });
});

const groupedRows = computed(() => {
  const groups = new Map<string, DataPermissionGrant[]>();

  for (const grant of filteredGrants.value) {
    const key = mode.value === "user" ? grant.user : grant.contractId;
    groups.set(key, [...(groups.get(key) ?? []), grant]);
  }

  return Array.from(groups.entries()).map(([key, grants]) => {
    // Each grants array is populated by the loop above before being
    // returned through the Map, so it always has at least one entry;
    // `head` narrows the type for the label/sublabel lookups below.
    const head = grants[0]!;
    return {
      key,
      label: mode.value === "user" ? head.userLabel : head.contractName,
      sublabel: head.scope,
      grants,
    };
  });
});
</script>

<template>
  <section class="settings-panel settings-data-permissions" aria-labelledby="settings-data-title">
    <header class="settings-panel__header">
      <div>
        <h2 id="settings-data-title" class="settings-panel__title">
          {{ t("settings.console.dataPermissions.title") }}
        </h2>
        <p class="settings-panel__subtitle">
          {{ t("settings.console.dataPermissions.subtitle") }}
        </p>
      </div>
    </header>

    <div class="settings-data-permissions__toolbar">
      <div
        class="settings-mode-switch"
        role="group"
        :aria-label="t('settings.console.dataPermissions.viewModeLabel')"
      >
        <button
          type="button"
          :class="{ 'is-active': mode === 'user' }"
          @click="mode = 'user'"
        >
          {{ t("settings.console.dataPermissions.userToContracts") }}
        </button>
        <button
          type="button"
          :class="{ 'is-active': mode === 'contract' }"
          @click="mode = 'contract'"
        >
          {{ t("settings.console.dataPermissions.contractToUsers") }}
        </button>
      </div>

      <label class="settings-data-permissions__search">
        <span class="label">{{ t("settings.console.dataPermissions.searchLabel") }}</span>
        <input
          v-model="search"
          class="input"
          type="search"
          :placeholder="t('settings.console.dataPermissions.searchPlaceholder')"
          autocomplete="off"
        />
      </label>
    </div>

    <div class="settings-data-permissions__body">
      <article
        v-for="group in groupedRows"
        :key="group.key"
        class="settings-data-group"
      >
        <header class="settings-data-group__header">
          <div>
            <h3>{{ group.label }}</h3>
            <p>{{ group.sublabel }}</p>
          </div>
          <span class="badge badge-neutral">
            {{ t("settings.console.dataPermissions.grantsCount", { count: group.grants.length }) }}
          </span>
        </header>

        <div class="settings-data-group__grants">
          <div
            v-for="grant in group.grants"
            :key="grant.id"
            class="settings-data-grant"
          >
            <div>
              <div class="settings-record-primary">
                {{ mode === "user" ? grant.contractName : grant.userLabel }}
              </div>
              <div class="settings-record-secondary">
                {{ mode === "user" ? grant.scope : grant.userLabel }}
              </div>
            </div>
            <span class="badge" :class="grant.source === 'session' ? 'badge-success' : 'badge-warning'">
              {{
                grant.source === "session"
                  ? t("settings.console.common.session")
                  : t("settings.console.common.demo")
              }}
            </span>
            <span class="badge badge-info">{{ grant.scope }}</span>
            <span class="settings-data-grant__updated">
              {{ formatDateTime(grant.updatedAt) }}
            </span>
          </div>
        </div>
      </article>

      <div v-if="groupedRows.length === 0" class="settings-card-state">
        {{ t("settings.console.dataPermissions.noMatches") }}
      </div>
    </div>
  </section>
</template>

<style scoped>
.settings-data-permissions__toolbar {
  display: grid;
  grid-template-columns: auto minmax(18rem, 1fr);
  gap: var(--space-4);
  align-items: end;
  padding: var(--space-5);
  border-bottom: 1px solid var(--border-subtle);
}

.settings-mode-switch {
  display: inline-flex;
  padding: var(--space-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
}

.settings-mode-switch button {
  min-height: var(--size-control-md);
  padding: 0 var(--space-4);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-weight: var(--font-weight-medium);
}

.settings-mode-switch button:hover {
  color: var(--text-primary);
}

.settings-mode-switch button.is-active {
  background: var(--bg-card);
  color: var(--action-primary);
  box-shadow: var(--shadow-xs);
}

.settings-data-permissions__search {
  display: grid;
  gap: var(--space-3);
}

.settings-data-permissions__body {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-5);
}

.settings-data-group {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.settings-data-group__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.settings-data-group__header h3 {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-md);
}

.settings-data-group__header p {
  margin: var(--space-1) 0 0;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.settings-data-group__grants {
  display: grid;
  gap: var(--space-2);
}

.settings-data-grant {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto auto;
  gap: var(--space-3);
  align-items: center;
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
}

.settings-data-grant__updated {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  white-space: nowrap;
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

.settings-card-state {
  padding: var(--space-6);
  border: 1px dashed var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  text-align: center;
}

@media (max-width: 820px) {
  .settings-data-permissions__toolbar,
  .settings-data-grant {
    grid-template-columns: 1fr;
  }

  .settings-mode-switch {
    width: 100%;
  }

  .settings-mode-switch button {
    flex: 1;
  }
}
</style>
