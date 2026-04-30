<script setup lang="ts">
import { computed, ref, watch } from "vue";

import type { AdminUser } from "../admin.types";
import type { BooleanFilterValue, UserDirectoryFilters } from "../ui.types";

const props = defineProps<{
  users: AdminUser[];
  total: number;
  offset: number;
  limit: number;
  selectedUserId: string | null;
  loading: boolean;
  error: string | null;
  filters: UserDirectoryFilters;
  formatDateTime: (value?: string | null) => string;
  statusLabel: (user: AdminUser) => string;
  statusClass: (user: AdminUser) => string;
}>();

const emit = defineEmits<{
  apply: [filters: UserDirectoryFilters];
  page: [offset: number];
  select: [userId: string];
}>();

const { t } = useI18n();

const search = ref(props.filters.search);
const active = ref<BooleanFilterValue>(props.filters.active);
const locked = ref<BooleanFilterValue>(props.filters.locked);

watch(
  () => props.filters,
  (filters) => {
    search.value = filters.search;
    active.value = filters.active;
    locked.value = filters.locked;
  },
  { deep: true },
);

const rangeLabel = computed(() => {
  if (props.total === 0) {
    return "0 of 0";
  }

  const start = props.offset + 1;
  const end = Math.min(props.offset + props.users.length, props.total);
  return `${start}-${end} of ${props.total}`;
});

const canPageBack = computed(() => props.offset > 0 && !props.loading);
const canPageForward = computed(
  () => props.offset + props.limit < props.total && !props.loading,
);

function applyFilters() {
  emit("apply", {
    search: search.value,
    active: active.value,
    locked: locked.value,
  });
}

function selectUser(userId: string) {
  emit("select", userId);
}
</script>

<template>
  <section class="settings-panel settings-user-directory">
    <header class="settings-panel__header">
      <div>
        <h2 class="settings-panel__title">
          {{ t("settings.usersPanelTitle") }}
        </h2>
        <p class="settings-panel__subtitle">
          Search, filter, and inspect accounts under IAM controls.
        </p>
      </div>
      <span class="badge badge-neutral">
        {{ t("settings.totalCount", { count: total }) }}
      </span>
    </header>

    <form class="settings-user-directory__filters" @submit.prevent="applyFilters">
      <div class="form-group">
        <label for="settings-user-search" class="label">Search users</label>
        <input
          id="settings-user-search"
          v-model="search"
          class="input"
          type="search"
          :placeholder="t('settings.searchUsersPlaceholder')"
          autocomplete="off"
        />
      </div>

      <div class="form-group">
        <label for="settings-user-active" class="label">Activity</label>
        <select id="settings-user-active" v-model="active" class="select">
          <option value="all">{{ t("settings.allActivity") }}</option>
          <option value="true">{{ t("settings.active") }}</option>
          <option value="false">{{ t("settings.disabled") }}</option>
        </select>
      </div>

      <div class="form-group">
        <label for="settings-user-locked" class="label">Lock state</label>
        <select id="settings-user-locked" v-model="locked" class="select">
          <option value="all">{{ t("settings.allLockStates") }}</option>
          <option value="true">{{ t("settings.locked") }}</option>
          <option value="false">{{ t("settings.unlocked") }}</option>
        </select>
      </div>

      <AppButton
        class="settings-user-directory__apply"
        type="submit"
        variant="secondary"
        size="sm"
        :loading="loading"
      >
        {{ t("common.apply") }}
      </AppButton>
    </form>

    <div v-if="error" class="alert alert-danger settings-panel__alert" role="alert">
      {{ error }}
    </div>

    <div class="settings-user-directory__table-shell">
      <div class="table-wrap settings-user-directory__table">
        <table class="table">
          <thead>
            <tr>
              <th>{{ t("settings.usersTable.user") }}</th>
              <th>{{ t("settings.usersTable.status") }}</th>
              <th>{{ t("settings.usersTable.groups") }}</th>
              <th>Failed</th>
              <th>{{ t("settings.usersTable.lastLogin") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="5" class="settings-table-state">
                {{ t("settings.loadingUsers") }}
              </td>
            </tr>
            <tr v-else-if="users.length === 0">
              <td colspan="5" class="settings-table-state">
                {{ t("settings.noUsers") }}
              </td>
            </tr>
            <template v-else>
              <tr
                v-for="user in users"
                :key="user.id"
                class="settings-user-row"
                :class="{ 'is-selected': selectedUserId === user.id }"
                tabindex="0"
                role="button"
                :aria-selected="selectedUserId === user.id"
                @click="selectUser(user.id)"
                @keydown.enter.prevent="selectUser(user.id)"
                @keydown.space.prevent="selectUser(user.id)"
              >
                <td>
                  <div class="settings-record-primary">{{ user.display_name }}</div>
                  <div class="settings-record-secondary">
                    {{ user.username }} / {{ user.email }}
                  </div>
                </td>
                <td>
                  <span class="badge" :class="statusClass(user)">
                    {{ statusLabel(user) }}
                  </span>
                </td>
                <td>
                  {{
                    user.groups.length
                      ? user.groups.join(", ")
                      : t("common.notAvailable")
                  }}
                </td>
                <td>{{ user.failed_login_attempts }}</td>
                <td>{{ formatDateTime(user.last_login_at) }}</td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <div class="settings-user-directory__mobile-list">
        <div v-if="loading" class="settings-card-state">
          {{ t("settings.loadingUsers") }}
        </div>
        <div v-else-if="users.length === 0" class="settings-card-state">
          {{ t("settings.noUsers") }}
        </div>
        <template v-else>
          <button
            v-for="user in users"
            :key="user.id"
            class="settings-user-card"
            :class="{ 'is-selected': selectedUserId === user.id }"
            type="button"
            @click="selectUser(user.id)"
          >
            <span class="settings-user-card__head">
              <span>
                <span class="settings-record-primary">{{ user.display_name }}</span>
                <span class="settings-record-secondary">{{ user.username }}</span>
              </span>
              <span class="badge" :class="statusClass(user)">
                {{ statusLabel(user) }}
              </span>
            </span>
            <span class="settings-user-card__meta">
              <span>{{ user.email }}</span>
              <span>{{ formatDateTime(user.last_login_at) }}</span>
            </span>
          </button>
        </template>
      </div>
    </div>

    <footer class="settings-panel__footer">
      <span>{{ rangeLabel }}</span>
      <div class="settings-pagination">
        <AppButton
          variant="secondary"
          size="sm"
          :disabled="!canPageBack"
          @click="emit('page', Math.max(0, offset - limit))"
        >
          {{ t("common.previous") }}
        </AppButton>
        <AppButton
          variant="secondary"
          size="sm"
          :disabled="!canPageForward"
          @click="emit('page', offset + limit)"
        >
          {{ t("common.next") }}
        </AppButton>
      </div>
    </footer>
  </section>
</template>

<style scoped>
.settings-user-directory {
  min-width: 0;
}

.settings-user-directory__filters {
  display: grid;
  grid-template-columns: minmax(16rem, 1.4fr) minmax(10rem, 0.7fr) minmax(10rem, 0.7fr) auto;
  gap: var(--space-4);
  align-items: end;
  padding: var(--space-5);
  border-bottom: 1px solid var(--border-subtle);
}

.settings-user-directory__filters .form-group {
  margin-bottom: 0;
}

.settings-user-directory__apply {
  align-self: end;
}

.settings-user-directory__table-shell {
  padding: var(--space-5);
}

.settings-user-directory__table {
  border-radius: var(--radius-md);
}

.settings-user-directory__mobile-list {
  display: none;
}

.settings-user-row {
  cursor: pointer;
}

.settings-user-row:focus-visible {
  outline: none;
  box-shadow: inset 0 0 0 2px var(--border-focus);
}

.settings-user-row.is-selected {
  background: var(--bg-selected);
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
  line-height: var(--line-height-relaxed);
  overflow-wrap: anywhere;
}

.settings-table-state,
.settings-card-state {
  color: var(--text-secondary);
  text-align: center;
}

.settings-card-state {
  padding: var(--space-6);
}

.settings-user-card {
  width: 100%;
  display: grid;
  gap: var(--space-4);
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  text-align: left;
}

.settings-user-card.is-selected {
  border-color: var(--border-focus);
  background: var(--bg-selected);
}

.settings-user-card__head {
  display: flex;
  justify-content: space-between;
  gap: var(--space-4);
  align-items: flex-start;
}

.settings-user-card__meta {
  display: grid;
  gap: var(--space-2);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  overflow-wrap: anywhere;
}

@media (max-width: 1120px) {
  .settings-user-directory__filters {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .settings-user-directory__filters {
    grid-template-columns: 1fr;
  }

  .settings-user-directory__table {
    display: none;
  }

  .settings-user-directory__mobile-list {
    display: grid;
    gap: var(--space-3);
  }

  .settings-user-card__head {
    flex-direction: column;
  }
}
</style>
