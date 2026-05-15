<script setup lang="ts">
import { computed, ref, watch } from "vue";

import type { AdminUser, AdminUserStatusAction } from "../admin.types";
import type { BooleanFilterValue, UserDirectoryFilters } from "../ui.types";
import { buildCsv, downloadBlob } from "../lib/csv";
import SettingsPaginationFooter from "./SettingsPaginationFooter.vue";

interface UserStatusBadge {
  label: string;
  badgeClass: string;
}

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
  statusBadges: (user: AdminUser) => UserStatusBadge[];
  canDeactivateUsers: boolean;
  canUpdateUsers: boolean;
  statusAction: AdminUserStatusAction | null;
}>();

const emit = defineEmits<{
  apply: [filters: UserDirectoryFilters];
  page: [offset: number];
  pageSize: [limit: number];
  select: [userId: string];
  status: [user: AdminUser, action: AdminUserStatusAction];
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

function nextActivationAction(user: AdminUser): AdminUserStatusAction {
  return user.is_active ? "disable" : "enable";
}

function nextLockAction(user: AdminUser): AdminUserStatusAction {
  return user.is_locked ? "unlock" : "lock";
}

function exportVisibleUsersCsv() {
  if (props.users.length === 0) return;

  const headers = [
    "id",
    "username",
    "display_name",
    "email",
    "is_active",
    "is_locked",
    "failed_login_attempts",
    "last_login_at",
    "groups",
  ];

  const rows = props.users.map((user) => [
    user.id ?? "",
    user.username ?? "",
    user.display_name ?? "",
    user.email ?? "",
    String(Boolean(user.is_active)),
    String(Boolean(user.is_locked)),
    String(user.failed_login_attempts ?? 0),
    user.last_login_at ?? "",
    user.groups?.join("|") ?? "",
  ]);

  const blob = buildCsv(headers, rows);
  const stamp = new Date().toISOString().replace(/[:.]/g, "-");
  downloadBlob(blob, `iam_users_visible_${stamp}.csv`);
}
</script>

<template>
  <section class="settings-panel settings-user-directory">
    <header class="settings-panel__header">
      <div>
        <h2 class="settings-panel__title">
          {{ t("settings.console.otherAccounts.title") }}
        </h2>
        <p class="settings-panel__subtitle">
          {{ t("settings.console.otherAccounts.subtitle") }}
        </p>
      </div>
      <div class="settings-user-directory__header-actions">
        <span class="badge badge-neutral">
          {{ t("settings.totalCount", { count: total }) }}
        </span>
        <AppButton
          variant="secondary"
          size="sm"
          :disabled="loading || users.length === 0"
          :title="t('settings.console.otherAccounts.exportCsvHint')"
          @click="exportVisibleUsersCsv"
        >
          {{ t("settings.console.otherAccounts.exportCsv") }}
        </AppButton>
      </div>
    </header>

    <div class="settings-user-directory__status-legend" aria-label="Account statuses">
      <span class="badge badge-success">
        {{ t("settings.console.otherAccounts.activeStatus") }}
      </span>
      <span class="badge badge-neutral">
        {{ t("settings.console.otherAccounts.disabledStatus") }}
      </span>
      <span class="badge badge-warning">
        {{ t("settings.console.otherAccounts.lockedStatus") }}
      </span>
      <span
        class="badge badge-closed"
        :title="t('settings.console.otherAccounts.resignedUnavailable')"
      >
        {{ t("settings.console.otherAccounts.resignedNotExposed") }}
      </span>
    </div>

    <form class="settings-user-directory__filters" @submit.prevent="applyFilters">
      <div class="form-group">
        <label for="settings-user-search" class="label">
          {{ t("settings.console.otherAccounts.searchUsers") }}
        </label>
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
        <label for="settings-user-active" class="label">
          {{ t("settings.console.otherAccounts.activity") }}
        </label>
        <select id="settings-user-active" v-model="active" class="select">
          <option value="all">{{ t("settings.allActivity") }}</option>
          <option value="true">{{ t("settings.active") }}</option>
          <option value="false">{{ t("settings.disabled") }}</option>
        </select>
      </div>

      <div class="form-group">
        <label for="settings-user-locked" class="label">
          {{ t("settings.console.otherAccounts.lockState") }}
        </label>
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
              <th>{{ t("settings.console.otherAccounts.failed") }}</th>
              <th>{{ t("settings.usersTable.lastLogin") }}</th>
              <th>{{ t("settings.console.otherAccounts.actions") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="6" class="settings-table-state">
                {{ t("settings.loadingUsers") }}
              </td>
            </tr>
            <tr v-else-if="users.length === 0">
              <td colspan="6" class="settings-table-state">
                {{ t("settings.noUsers") }}
              </td>
            </tr>
            <template v-else>
              <tr
                v-for="user in users"
                :key="user.id"
                class="settings-user-row"
                :class="{ 'is-selected': selectedUserId === user.id }"
                :aria-selected="selectedUserId === user.id"
              >
                <td>
                  <button
                    type="button"
                    class="settings-user-row__select"
                    :aria-pressed="selectedUserId === user.id"
                    @click="selectUser(user.id)"
                  >
                    <span class="settings-record-primary">{{ user.display_name }}</span>
                    <span class="settings-record-secondary">
                      {{ user.username }} / {{ user.email }}
                    </span>
                  </button>
                </td>
                <td>
                  <div class="settings-status-stack">
                    <span
                      v-for="badge in statusBadges(user)"
                      :key="badge.label"
                      class="badge"
                      :class="badge.badgeClass"
                    >
                      {{ badge.label }}
                    </span>
                  </div>
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
                <td>
                  <div class="settings-user-actions">
                    <AppButton
                      variant="secondary"
                      size="sm"
                      @click="selectUser(user.id)"
                    >
                      {{ t("settings.console.otherAccounts.view") }}
                    </AppButton>
                    <AppButton
                      variant="ghost"
                      size="sm"
                      disabled
                      :title="t('settings.console.otherAccounts.editUnavailable')"
                    >
                      {{ t("settings.console.otherAccounts.edit") }}
                    </AppButton>
                    <AppButton
                      v-if="canDeactivateUsers"
                      :variant="user.is_active ? 'danger' : 'secondary'"
                      size="sm"
                      :disabled="Boolean(statusAction)"
                      :loading="statusAction === nextActivationAction(user)"
                      @click="emit('status', user, nextActivationAction(user))"
                    >
                      {{ user.is_active ? t("settings.actions.disable") : t("settings.actions.enable") }}
                    </AppButton>
                    <AppButton
                      v-if="canUpdateUsers"
                      variant="secondary"
                      size="sm"
                      :disabled="Boolean(statusAction)"
                      :loading="statusAction === nextLockAction(user)"
                      @click="emit('status', user, nextLockAction(user))"
                    >
                      {{ user.is_locked ? t("settings.actions.unlock") : t("settings.actions.lock") }}
                    </AppButton>
                  </div>
                </td>
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
          <article
            v-for="user in users"
            :key="user.id"
            class="settings-user-card"
            :class="{ 'is-selected': selectedUserId === user.id }"
            :aria-selected="selectedUserId === user.id"
          >
            <button
              type="button"
              class="settings-user-card__select"
              :aria-pressed="selectedUserId === user.id"
              @click="selectUser(user.id)"
            >
              <span class="settings-user-card__head">
                <span>
                  <span class="settings-record-primary">{{ user.display_name }}</span>
                  <span class="settings-record-secondary">{{ user.username }}</span>
                </span>
                <span class="settings-status-stack">
                  <span
                    v-for="badge in statusBadges(user)"
                    :key="badge.label"
                    class="badge"
                    :class="badge.badgeClass"
                  >
                    {{ badge.label }}
                  </span>
                </span>
              </span>
              <span class="settings-user-card__meta">
                <span>{{ user.email }}</span>
                <span>{{ formatDateTime(user.last_login_at) }}</span>
              </span>
            </button>
            <span class="settings-user-card__actions">
              <AppButton
                variant="secondary"
                size="sm"
                @click="selectUser(user.id)"
              >
                {{ t("settings.console.otherAccounts.view") }}
              </AppButton>
              <AppButton
                variant="ghost"
                size="sm"
                disabled
                :title="t('settings.console.otherAccounts.editUnavailable')"
              >
                {{ t("settings.console.otherAccounts.edit") }}
              </AppButton>
              <AppButton
                v-if="canDeactivateUsers"
                :variant="user.is_active ? 'danger' : 'secondary'"
                size="sm"
                :disabled="Boolean(statusAction)"
                @click="emit('status', user, nextActivationAction(user))"
              >
                {{ user.is_active ? t("settings.actions.disable") : t("settings.actions.enable") }}
              </AppButton>
              <AppButton
                v-if="canUpdateUsers"
                variant="secondary"
                size="sm"
                :disabled="Boolean(statusAction)"
                @click="emit('status', user, nextLockAction(user))"
              >
                {{ user.is_locked ? t("settings.actions.unlock") : t("settings.actions.lock") }}
              </AppButton>
            </span>
          </article>
        </template>
      </div>
    </div>

    <SettingsPaginationFooter
      :total="total"
      :offset="offset"
      :limit="limit"
      :page-count="users.length"
      :loading="loading"
      @page="(next) => emit('page', next)"
      @page-size="(next) => emit('pageSize', next)"
    />
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

.settings-user-directory__status-legend {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-card-muted);
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

.settings-user-actions,
.settings-user-card__actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.settings-user-row__select {
  width: 100%;
  display: block;
  padding: 0;
  border: 0;
  background: transparent;
  text-align: left;
  color: inherit;
  cursor: pointer;
}

.settings-user-row__select:focus-visible {
  outline: 2px solid var(--border-focus);
  outline-offset: 2px;
  border-radius: var(--radius-sm);
}

.settings-user-card__select {
  width: 100%;
  display: grid;
  gap: var(--space-3);
  padding: 0;
  border: 0;
  background: transparent;
  text-align: left;
  color: inherit;
  cursor: pointer;
}

.settings-user-card__select:focus-visible {
  outline: 2px solid var(--border-focus);
  outline-offset: 2px;
  border-radius: var(--radius-sm);
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

.settings-user-card__actions {
  margin-top: var(--space-1);
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

.settings-status-stack {
  display: inline-flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.settings-user-directory__header-actions {
  display: inline-flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}
</style>
