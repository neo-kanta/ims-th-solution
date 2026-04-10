<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";

import { auditApi } from "~/features/settings/services/auditApi";
import type { AuditListPayload } from "~/features/settings/audit.types";
import { adminApi } from "~/features/settings/services/adminApi";
import type {
  AdminSession,
  AdminUser,
  AdminUserListPayload,
  AdminUserStatusAction,
  CreateAdminUserInput,
} from "~/features/settings/admin.types";
import { createDateFormatter } from "~/shared/i18n/intl";

type BooleanFilterValue = "all" | "true" | "false";

const authStore = useAuthStore();
const { t, locale } = useI18n();
const canViewUsers = computed(() => authStore.hasPermission("IAM_USER_VIEW"));
const canCreateUsers = computed(() =>
  authStore.hasPermission("IAM_USER_CREATE"),
);
const canDeactivateUsers = computed(() =>
  authStore.hasPermission("IAM_USER_DEACTIVATE"),
);
const canUpdateUsers = computed(() =>
  authStore.hasPermission("IAM_USER_UPDATE"),
);
const canViewAudit = computed(() => authStore.hasPermission("IAM_AUDIT_VIEW"));
const hasAccess = computed(
  () =>
    canViewUsers.value ||
    canCreateUsers.value ||
    canDeactivateUsers.value ||
    canUpdateUsers.value ||
    canViewAudit.value,
);

if (!hasAccess.value) {
  await navigateTo("/403", { replace: true });
}

const bangkokFormatter = computed(() =>
  createDateFormatter(
    locale.value,
    {
      dateStyle: "medium",
      timeStyle: "short",
    },
    "Asia/Bangkok",
  ),
);

const userQuery = reactive({
  search: "",
  active: "all" as BooleanFilterValue,
  locked: "all" as BooleanFilterValue,
  offset: 0,
  limit: 12,
});
const auditQuery = reactive({
  actorId: "",
  eventType: "",
  targetType: "",
  targetId: "",
  offset: 0,
  limit: 12,
});
const createForm = reactive<CreateAdminUserInput>({
  username: "",
  display_name: "",
  email: "",
  password: "",
});
const resetPasswordForm = reactive({ newPassword: "" });

const usersState = ref<AdminUserListPayload>({
  users: [],
  total: 0,
  offset: 0,
  limit: 12,
});
const auditState = ref<AuditListPayload>({
  events: [],
  total: 0,
  offset: 0,
  limit: 12,
});
const sessions = ref<AdminSession[]>([]);
const selectedUserId = ref<string | null>(null);

const usersLoading = ref(false);
const auditLoading = ref(false);
const sessionsLoading = ref(false);
const createLoading = ref(false);
const resetLoading = ref(false);
const userAction = ref<AdminUserStatusAction | null>(null);
const revokingSessionId = ref<string | null>(null);

const usersError = ref<string | null>(null);
const auditError = ref<string | null>(null);
const formError = ref<string | null>(null);
const resetError = ref<string | null>(null);
const notice = ref<string | null>(null);

const users = computed(() => usersState.value.users);
const selectedUser = computed<AdminUser | null>(
  () => users.value.find((user) => user.id === selectedUserId.value) ?? null,
);
const totalUsers = computed(() => usersState.value.total);
const activeUsers = computed(
  () => users.value.filter((user) => user.is_active).length,
);
const lockedUsers = computed(
  () => users.value.filter((user) => user.is_locked).length,
);
const auditTotal = computed(() => auditState.value.total);

function toOptionalBoolean(value: BooleanFilterValue) {
  if (value === "true") return true;
  if (value === "false") return false;
  return undefined;
}

function formatDateTime(value?: string | null) {
  if (!value) return t("common.notAvailable");
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? value
    : bangkokFormatter.value.format(date);
}

function getErrorMessage(error: any, fallback: string) {
  return (
    error?.data?.error || error?.data?.message || error?.message || fallback
  );
}

function statusLabel(user: AdminUser) {
  if (!user.is_active) return t("settings.status.disabled");
  if (user.is_locked) return t("settings.status.locked");
  return t("settings.status.active");
}

function statusClass(user: AdminUser) {
  if (!user.is_active) return "badge-neutral";
  if (user.is_locked) return "badge-warning";
  return "badge-success";
}

async function loadSessions(userId: string) {
  if (!canUpdateUsers.value) return;
  sessionsLoading.value = true;
  try {
    sessions.value = (await adminApi.listUserSessions(userId)).data;
  } catch {
    sessions.value = [];
  } finally {
    sessionsLoading.value = false;
  }
}

async function loadUsers(offset = userQuery.offset) {
  if (!canViewUsers.value) return;
  usersLoading.value = true;
  usersError.value = null;
  userQuery.offset = Math.max(0, offset);
  try {
    usersState.value = (
      await adminApi.listUsers({
        search: userQuery.search,
        is_active: toOptionalBoolean(userQuery.active),
        is_locked: toOptionalBoolean(userQuery.locked),
        offset: userQuery.offset,
        limit: userQuery.limit,
      })
    ).data;
    selectedUserId.value = usersState.value.users.some(
      (user) => user.id === selectedUserId.value,
    )
      ? selectedUserId.value
      : (usersState.value.users[0]?.id ?? null);
    sessions.value = [];
    if (selectedUserId.value) await loadSessions(selectedUserId.value);
  } catch (error: any) {
    usersError.value = getErrorMessage(error, t("settings.errors.loadUsers"));
    usersState.value = {
      users: [],
      total: 0,
      offset: userQuery.offset,
      limit: userQuery.limit,
    };
    selectedUserId.value = null;
    sessions.value = [];
  } finally {
    usersLoading.value = false;
  }
}

async function loadAudit(offset = auditQuery.offset) {
  if (!canViewAudit.value) return;
  auditLoading.value = true;
  auditError.value = null;
  auditQuery.offset = Math.max(0, offset);
  try {
    auditState.value = (
      await auditApi.listEvents({
        actor_id: auditQuery.actorId,
        event_type: auditQuery.eventType,
        target_type: auditQuery.targetType,
        target_id: auditQuery.targetId,
        offset: auditQuery.offset,
        limit: auditQuery.limit,
      })
    ).data;
  } catch (error: any) {
    auditError.value = getErrorMessage(error, t("settings.errors.loadAudit"));
    auditState.value = {
      events: [],
      total: 0,
      offset: auditQuery.offset,
      limit: auditQuery.limit,
    };
  } finally {
    auditLoading.value = false;
  }
}

async function createUser() {
  if (!canCreateUsers.value) return;
  createLoading.value = true;
  formError.value = null;
  notice.value = null;
  try {
    const created = await adminApi.createUser({
      username: createForm.username.trim(),
      display_name: createForm.display_name.trim(),
      email: createForm.email.trim(),
      password: createForm.password,
    });
    notice.value = t("settings.notices.userCreated");
    createForm.username = "";
    createForm.display_name = "";
    createForm.email = "";
    createForm.password = "";
    if (canViewUsers.value) {
      await loadUsers(0);
      selectedUserId.value = created.id;
      if (created.id) await loadSessions(created.id);
    }
    if (canViewAudit.value) await loadAudit(0);
  } catch (error: any) {
    formError.value = getErrorMessage(error, t("settings.errors.createUser"));
  } finally {
    createLoading.value = false;
  }
}

function actionLabel(action: AdminUserStatusAction) {
  switch (action) {
    case "disable":
      return t("settings.actions.disable");
    case "enable":
      return t("settings.actions.enable");
    case "lock":
      return t("settings.actions.lock");
    case "unlock":
      return t("settings.actions.unlock");
  }
}

async function changeUserStatus(action: AdminUserStatusAction) {
  if (!selectedUser.value) return;
  userAction.value = action;
  notice.value = null;
  try {
    await adminApi.setUserStatus(selectedUser.value.id, action);
    notice.value = t("settings.notices.userActionSuccess", {
      action: actionLabel(action),
    });
    await loadUsers(userQuery.offset);
    if (canViewAudit.value) await loadAudit(auditQuery.offset);
  } catch (error: any) {
    notice.value = getErrorMessage(error, t("settings.errors.changeUserStatus"));
  } finally {
    userAction.value = null;
  }
}

async function resetPassword() {
  if (!selectedUser.value || !canUpdateUsers.value) return;
  resetLoading.value = true;
  resetError.value = null;
  notice.value = null;
  try {
    await adminApi.resetPassword(
      selectedUser.value.id,
      resetPasswordForm.newPassword,
    );
    notice.value = t("settings.notices.passwordReset");
    resetPasswordForm.newPassword = "";
    await loadUsers(userQuery.offset);
    if (canViewAudit.value) await loadAudit(auditQuery.offset);
  } catch (error: any) {
    resetError.value = getErrorMessage(error, t("settings.errors.resetPassword"));
  } finally {
    resetLoading.value = false;
  }
}

async function revokeSession(sessionId: string) {
  revokingSessionId.value = sessionId;
  notice.value = null;
  try {
    await adminApi.revokeSession(sessionId);
    notice.value = t("settings.notices.sessionRevoked");
    if (selectedUser.value) await loadSessions(selectedUser.value.id);
    if (canViewAudit.value) await loadAudit(auditQuery.offset);
  } catch (error: any) {
    notice.value = getErrorMessage(error, t("settings.errors.revokeSession"));
  } finally {
    revokingSessionId.value = null;
  }
}

onMounted(async () => {
  const tasks: Promise<unknown>[] = [];
  if (canViewUsers.value) tasks.push(loadUsers(0));
  if (canViewAudit.value) tasks.push(loadAudit(0));
  await Promise.all(tasks);
});
</script>

<template>
  <div class="settings-page">
    <div class="page-header">
      <div>
        <div class="breadcrumb">
          <span>{{ t("settings.breadcrumb") }}</span>
        </div>
        <h1 class="page-title">{{ t("settings.title") }}</h1>
        <p class="page-desc">{{ t("settings.description") }}</p>
      </div>
      <div class="page-actions">
        <button
          v-if="canViewUsers"
          class="btn btn-secondary btn-sm"
          type="button"
          :disabled="usersLoading"
          @click="loadUsers(userQuery.offset)"
        >
          {{ t("settings.refreshUsers") }}
        </button>
        <button
          v-if="canViewAudit"
          class="btn btn-secondary btn-sm"
          type="button"
          :disabled="auditLoading"
          @click="loadAudit(auditQuery.offset)"
        >
          {{ t("settings.refreshAudit") }}
        </button>
      </div>
    </div>

    <div v-if="notice" class="notice-banner">{{ notice }}</div>

    <div class="grid-4 metrics-row">
      <div class="stat-card">
        <div class="stat-card-label">{{ t("settings.metrics.usersLabel") }}</div>
        <div class="stat-card-value">{{ totalUsers }}</div>
        <div class="stat-card-meta">{{ t("settings.metrics.usersMeta") }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-card-label">
          {{ t("settings.metrics.activeOnPageLabel") }}
        </div>
        <div class="stat-card-value">{{ activeUsers }}</div>
        <div class="stat-card-meta">
          {{ t("settings.metrics.activeOnPageMeta") }}
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-card-label">
          {{ t("settings.metrics.lockedOnPageLabel") }}
        </div>
        <div class="stat-card-value">{{ lockedUsers }}</div>
        <div class="stat-card-meta">
          {{ t("settings.metrics.lockedOnPageMeta") }}
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-card-label">
          {{ t("settings.metrics.auditHitsLabel") }}
        </div>
        <div class="stat-card-value">{{ auditTotal }}</div>
        <div class="stat-card-meta">
          {{ t("settings.metrics.auditHitsMeta") }}
        </div>
      </div>
    </div>

    <div class="workspace-grid">
      <section v-if="canViewUsers" class="card panel">
        <div class="panel-head">
          <div class="panel-title">{{ t("settings.usersPanelTitle") }}</div>
          <span class="badge badge-neutral">
            {{ t("settings.totalCount", { count: totalUsers }) }}
          </span>
        </div>
        <div class="toolbar">
          <input
            v-model="userQuery.search"
            class="form-input"
            type="search"
            :placeholder="t('settings.searchUsersPlaceholder')"
            :aria-label="t('settings.searchUsersPlaceholder')"
            @keyup.enter="loadUsers(0)"
          />
          <select v-model="userQuery.active" class="form-select">
            <option value="all">{{ t("settings.allActivity") }}</option>
            <option value="true">{{ t("settings.active") }}</option>
            <option value="false">{{ t("settings.disabled") }}</option>
          </select>
          <select v-model="userQuery.locked" class="form-select">
            <option value="all">{{ t("settings.allLockStates") }}</option>
            <option value="true">{{ t("settings.locked") }}</option>
            <option value="false">{{ t("settings.unlocked") }}</option>
          </select>
          <button
            class="btn btn-secondary btn-sm"
            type="button"
            :disabled="usersLoading"
            @click="loadUsers(0)"
        >
            {{ t("common.apply") }}
          </button>
        </div>
        <div v-if="usersError" class="section-error">{{ usersError }}</div>
        <div class="table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>{{ t("settings.usersTable.user") }}</th>
                <th>{{ t("settings.usersTable.status") }}</th>
                <th>{{ t("settings.usersTable.groups") }}</th>
                <th>{{ t("settings.usersTable.lastLogin") }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="usersLoading">
                <td colspan="4" class="table-empty">
                  {{ t("settings.loadingUsers") }}
                </td>
              </tr>
              <tr v-else-if="users.length === 0">
                <td colspan="4" class="table-empty">
                  {{ t("settings.noUsers") }}
                </td>
              </tr>
              <tr
                v-for="user in users"
                v-else
                :key="user.id"
                class="clickable-row"
                :class="{ 'is-selected': selectedUserId === user.id }"
                @click="
                  selectedUserId = user.id;
                  loadSessions(user.id);
                "
              >
                <td>
                  <div class="row-primary">{{ user.display_name }}</div>
                  <div class="row-secondary">
                    {{ user.username }} / {{ user.email }}
                  </div>
                </td>
                <td>
                  <span class="badge" :class="statusClass(user)">{{
                    statusLabel(user)
                  }}</span>
                </td>
                <td>
                  {{
                    user.groups.length
                      ? user.groups.join(", ")
                      : t("common.notAvailable")
                  }}
                </td>
                <td>{{ formatDateTime(user.last_login_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="panel-foot">
          <span>
            {{
              t("settings.offsetLimit", {
                offset: usersState.offset,
                limit: usersState.limit,
              })
            }}
          </span>
          <div class="page-actions">
            <button
              class="btn btn-secondary btn-sm"
              type="button"
              :disabled="userQuery.offset === 0 || usersLoading"
              @click="loadUsers(userQuery.offset - userQuery.limit)"
            >
              {{ t("common.previous") }}
            </button>
            <button
              class="btn btn-secondary btn-sm"
              type="button"
              :disabled="
                usersState.offset + usersState.limit >= usersState.total ||
                usersLoading
              "
              @click="loadUsers(userQuery.offset + userQuery.limit)"
            >
              {{ t("common.next") }}
            </button>
          </div>
        </div>
      </section>

      <aside class="side-stack">
        <section v-if="canCreateUsers" class="card panel">
          <div class="panel-head">
            <div class="panel-title">{{ t("settings.createUserTitle") }}</div>
          </div>
          <div class="form-stack">
            <input
              v-model="createForm.username"
              class="form-input"
              type="text"
              :placeholder="t('settings.usernamePlaceholder')"
              :aria-label="t('settings.usernamePlaceholder')"
            />
            <input
              v-model="createForm.display_name"
              class="form-input"
              type="text"
              :placeholder="t('settings.displayNamePlaceholder')"
              :aria-label="t('settings.displayNamePlaceholder')"
            />
            <input
              v-model="createForm.email"
              class="form-input"
              type="email"
              :placeholder="t('settings.emailPlaceholder')"
              :aria-label="t('settings.emailPlaceholder')"
            />
            <input
              v-model="createForm.password"
              class="form-input"
              type="password"
              :placeholder="t('settings.temporaryPasswordPlaceholder')"
              :aria-label="t('settings.temporaryPasswordPlaceholder')"
            />
            <div v-if="formError" class="section-error">{{ formError }}</div>
            <button
              class="btn btn-primary"
              type="button"
              :disabled="createLoading"
              @click="createUser"
            >
              {{
                createLoading
                  ? t("settings.creating")
                  : t("settings.createAccount")
              }}
            </button>
          </div>
        </section>

        <section v-if="selectedUser" class="card panel">
          <div class="panel-head">
            <div>
              <div class="panel-title">{{ selectedUser.display_name }}</div>
              <div class="row-secondary">{{ selectedUser.username }}</div>
            </div>
            <span class="badge" :class="statusClass(selectedUser)">{{
              statusLabel(selectedUser)
            }}</span>
          </div>
          <div class="detail-list">
            <div>
              <span class="detail-label">{{ t("settings.details.email") }}</span
              ><span>{{ selectedUser.email }}</span>
            </div>
            <div>
              <span class="detail-label">{{ t("settings.details.groups") }}</span
              ><span>{{
                selectedUser.groups.length
                  ? selectedUser.groups.join(", ")
                  : t("common.notAvailable")
              }}</span>
            </div>
            <div>
              <span class="detail-label">
                {{ t("settings.details.failedLogins") }}</span
              ><span>{{ selectedUser.failed_login_attempts }}</span>
            </div>
            <div>
              <span class="detail-label">
                {{ t("settings.details.lockedUntil") }}</span
              ><span>{{ formatDateTime(selectedUser.locked_until) }}</span>
            </div>
          </div>
          <div class="action-grid">
            <button
              v-if="canDeactivateUsers"
              class="btn btn-secondary btn-sm"
              type="button"
              :disabled="userAction === 'disable' || !selectedUser.is_active"
              @click="changeUserStatus('disable')"
            >
              {{
                userAction === "disable"
                  ? t("settings.actions.disabling")
                  : t("settings.actions.disable")
              }}
            </button>
            <button
              v-if="canDeactivateUsers"
              class="btn btn-secondary btn-sm"
              type="button"
              :disabled="userAction === 'enable' || selectedUser.is_active"
              @click="changeUserStatus('enable')"
            >
              {{
                userAction === "enable"
                  ? t("settings.actions.enabling")
                  : t("settings.actions.enable")
              }}
            </button>
            <button
              v-if="canUpdateUsers"
              class="btn btn-secondary btn-sm"
              type="button"
              :disabled="userAction === 'lock' || selectedUser.is_locked"
              @click="changeUserStatus('lock')"
            >
              {{
                userAction === "lock"
                  ? t("settings.actions.locking")
                  : t("settings.actions.lock")
              }}
            </button>
            <button
              v-if="canUpdateUsers"
              class="btn btn-secondary btn-sm"
              type="button"
              :disabled="userAction === 'unlock' || !selectedUser.is_locked"
              @click="changeUserStatus('unlock')"
            >
              {{
                userAction === "unlock"
                  ? t("settings.actions.unlocking")
                  : t("settings.actions.unlock")
              }}
            </button>
          </div>
          <div v-if="canUpdateUsers" class="form-stack compact-section">
            <input
              v-model="resetPasswordForm.newPassword"
              class="form-input"
              type="password"
              :placeholder="t('settings.newTemporaryPasswordPlaceholder')"
              :aria-label="t('settings.newTemporaryPasswordPlaceholder')"
            />
            <div v-if="resetError" class="section-error">{{ resetError }}</div>
            <button
              class="btn btn-primary btn-sm"
              type="button"
              :disabled="resetLoading"
              @click="resetPassword"
            >
              {{
                resetLoading
                  ? t("settings.actions.submitting")
                  : t("settings.actions.resetPassword")
              }}
            </button>
          </div>
          <div v-if="canUpdateUsers" class="compact-section">
            <div class="panel-subtitle">{{ t("settings.sessionsTitle") }}</div>
            <div v-if="sessionsLoading" class="table-empty">
              {{ t("settings.loadingSessions") }}
            </div>
            <div v-else-if="sessions.length === 0" class="table-empty">
              {{ t("settings.noSessions") }}
            </div>
            <div v-else class="session-list">
              <div
                v-for="session in sessions"
                :key="session.id"
                class="session-item"
              >
                <div>
                  <div class="row-primary">
                    {{ session.ip_address || t("settings.unknownIp") }}
                  </div>
                  <div class="row-secondary">{{ session.user_agent }}</div>
                  <div class="row-secondary">
                    {{
                      t("settings.lastActivity", {
                        date: formatDateTime(session.last_activity_at),
                      })
                    }}
                  </div>
                </div>
                <button
                  class="btn btn-secondary btn-sm"
                  type="button"
                  :disabled="revokingSessionId === session.id"
                  @click="revokeSession(session.id)"
                >
                  {{
                    revokingSessionId === session.id
                      ? t("settings.actions.revoking")
                      : t("settings.actions.revoke")
                  }}
                </button>
              </div>
            </div>
          </div>
        </section>
      </aside>
    </div>

    <section v-if="canViewAudit" class="card panel">
      <div class="panel-head">
        <div class="panel-title">{{ t("settings.auditTitle") }}</div>
        <span class="badge badge-neutral">
          {{ t("settings.auditCount", { count: auditTotal }) }}
        </span>
      </div>
      <div class="toolbar">
        <input
          v-model="auditQuery.actorId"
          class="form-input"
          type="text"
          :placeholder="t('settings.actorIdPlaceholder')"
          :aria-label="t('settings.actorIdPlaceholder')"
          @keyup.enter="loadAudit(0)"
        />
        <input
          v-model="auditQuery.eventType"
          class="form-input"
          type="text"
          :placeholder="t('settings.eventTypePlaceholder')"
          :aria-label="t('settings.eventTypePlaceholder')"
          @keyup.enter="loadAudit(0)"
        />
        <input
          v-model="auditQuery.targetType"
          class="form-input"
          type="text"
          :placeholder="t('settings.targetTypePlaceholder')"
          :aria-label="t('settings.targetTypePlaceholder')"
          @keyup.enter="loadAudit(0)"
        />
        <input
          v-model="auditQuery.targetId"
          class="form-input"
          type="text"
          :placeholder="t('settings.targetIdPlaceholder')"
          :aria-label="t('settings.targetIdPlaceholder')"
          @keyup.enter="loadAudit(0)"
        />
        <button
          class="btn btn-secondary btn-sm"
          type="button"
          :disabled="auditLoading"
          @click="loadAudit(0)"
        >
          {{ t("common.apply") }}
        </button>
      </div>
      <div v-if="auditError" class="section-error">{{ auditError }}</div>
      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ t("settings.auditTable.event") }}</th>
              <th>{{ t("settings.auditTable.target") }}</th>
              <th>{{ t("settings.auditTable.actor") }}</th>
              <th>{{ t("settings.auditTable.ip") }}</th>
              <th>{{ t("settings.auditTable.created") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="auditLoading">
              <td colspan="5" class="table-empty">
                {{ t("settings.loadingAudit") }}
              </td>
            </tr>
            <tr v-else-if="auditState.events.length === 0">
              <td colspan="5" class="table-empty">
                {{ t("settings.noAudit") }}
              </td>
            </tr>
            <tr v-for="event in auditState.events" v-else :key="event.id">
              <td>
                <div class="row-primary">{{ event.event_type }}</div>
                <div class="row-secondary">{{ event.id }}</div>
              </td>
              <td>
                <div class="row-primary">{{ event.target_type }}</div>
                <div class="row-secondary">{{ event.target_id }}</div>
              </td>
              <td>{{ event.actor_id || t("settings.systemActor") }}</td>
              <td>{{ event.ip_address || t("common.notAvailable") }}</td>
              <td>{{ formatDateTime(event.created_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="panel-foot">
        <span>
          {{
            t("settings.offsetLimit", {
              offset: auditState.offset,
              limit: auditState.limit,
            })
          }}
        </span>
        <div class="page-actions">
          <button
            class="btn btn-secondary btn-sm"
            type="button"
            :disabled="auditQuery.offset === 0 || auditLoading"
            @click="loadAudit(auditQuery.offset - auditQuery.limit)"
          >
            {{ t("common.previous") }}
          </button>
          <button
            class="btn btn-secondary btn-sm"
            type="button"
            :disabled="
              auditState.offset + auditState.limit >= auditState.total ||
              auditLoading
            "
            @click="loadAudit(auditQuery.offset + auditQuery.limit)"
          >
            {{ t("common.next") }}
          </button>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.settings-page,
.side-stack,
.form-stack {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.page-actions,
.panel-foot,
.panel-head,
.action-grid,
.session-item {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}
.page-header,
.panel-head,
.panel-foot,
.session-item {
  justify-content: space-between;
}
.notice-banner {
  padding: 0.9rem 1rem;
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 0.9rem;
  background: #eff6ff;
  color: #1d4ed8;
}
.metrics-row {
  margin-bottom: 0.25rem;
}
.workspace-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.8fr) minmax(320px, 0.9fr);
  gap: 1rem;
  align-items: start;
}
.panel {
  padding: 0;
  overflow: hidden;
  border: 1px solid rgba(15, 23, 42, 0.08);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.04);
}
.panel-head,
.panel-foot {
  padding: 1rem 1.15rem;
  border-bottom: 1px solid rgba(15, 23, 42, 0.06);
}
.panel-foot {
  border-top: 1px solid rgba(15, 23, 42, 0.06);
  border-bottom: 0;
  color: #64748b;
}
.panel-title {
  font-size: 1rem;
  font-weight: 700;
  color: #0f172a;
}
.panel-subtitle,
.detail-label {
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #64748b;
}
.toolbar,
.form-stack,
.detail-list,
.compact-section,
.table-wrap {
  padding: 1rem 1.15rem;
}
.toolbar {
  display: grid;
  grid-template-columns: minmax(0, 1.6fr) repeat(2, minmax(160px, 0.8fr)) auto;
  gap: 0.75rem;
}
.table-wrap {
  overflow: auto;
}
.data-table {
  width: 100%;
  border-collapse: collapse;
}
.data-table th {
  padding: 0 0 0.8rem;
  text-align: left;
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: #64748b;
}
.data-table td {
  padding: 0.85rem 0;
  border-top: 1px solid rgba(15, 23, 42, 0.06);
  vertical-align: top;
  color: #334155;
}
.row-primary {
  font-weight: 600;
  color: #0f172a;
}
.row-secondary {
  margin-top: 0.2rem;
  color: #64748b;
  font-size: 0.92rem;
  line-height: 1.5;
}
.clickable-row {
  cursor: pointer;
  transition: background-color 0.14s ease;
}
.clickable-row:hover {
  background: rgba(248, 250, 252, 0.96);
}
.clickable-row.is-selected {
  background: rgba(239, 246, 255, 0.92);
}
.section-error {
  margin: 0 1.15rem 1rem;
  padding: 0.75rem 0.85rem;
  border-radius: 0.75rem;
  background: #fff7ed;
  border: 1px solid rgba(217, 119, 6, 0.18);
  color: #9a3412;
}
.table-empty {
  color: #64748b;
}
.detail-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.8rem 1rem;
}
.detail-list span:last-child {
  display: block;
  margin-top: 0.2rem;
  color: #0f172a;
}
.session-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.session-item {
  padding: 0.85rem 0;
  border-top: 1px solid rgba(15, 23, 42, 0.06);
}
.session-item:first-child {
  border-top: 0;
  padding-top: 0;
}
.badge-success {
  background: rgba(22, 163, 74, 0.12);
  color: #166534;
}
@media (max-width: 1080px) {
  .workspace-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
@media (max-width: 768px) {
  .toolbar,
  .detail-list {
    grid-template-columns: minmax(0, 1fr);
  }
  .page-header,
  .panel-head,
  .panel-foot,
  .session-item {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
