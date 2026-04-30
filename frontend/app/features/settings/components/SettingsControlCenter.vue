<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";

import type {
  AdminSession,
  AdminUser,
  AdminUserFilters,
  AdminUserListPayload,
  AdminUserStatusAction,
  CreateAdminUserInput,
} from "../admin.types";
import type { AuditFilters, AuditListPayload } from "../audit.types";
import type {
  AuditLogFilters,
  BooleanFilterValue,
  SettingsKpiMetric,
  UserDirectoryFilters,
} from "../ui.types";
import { adminApi } from "../services/adminApi";
import { auditApi } from "../services/auditApi";
import SettingsAuditLog from "./SettingsAuditLog.vue";
import SettingsConfirmDialog from "./SettingsConfirmDialog.vue";
import SettingsCreateUserForm from "./SettingsCreateUserForm.vue";
import SettingsKpiGrid from "./SettingsKpiGrid.vue";
import SettingsUserDetailPanel from "./SettingsUserDetailPanel.vue";
import SettingsUserDirectory from "./SettingsUserDirectory.vue";
import { createDateFormatter } from "~/shared/i18n/intl";
import { isRiskAuditEvent } from "../lib/audit";

const USER_PAGE_LIMIT = 12;
const AUDIT_PAGE_LIMIT = 12;

type ToastTone = "success" | "danger" | "info";

interface ToastState {
  tone: ToastTone;
  message: string;
}

type ConfirmAction =
  | {
      type: "status";
      action: AdminUserStatusAction;
      user: AdminUser;
    }
  | {
      type: "reset-password";
      user: AdminUser;
      newPassword: string;
    }
  | {
      type: "revoke-session";
      user: AdminUser;
      session: AdminSession;
    }
  | {
      type: "export-audit";
    };

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function getErrorMessage(error: unknown, fallback: string): string {
  if (isRecord(error)) {
    const data = error.data;

    if (isRecord(data)) {
      if (typeof data.error === "string") {
        return data.error;
      }

      if (typeof data.message === "string") {
        return data.message;
      }
    }

    if (typeof error.message === "string") {
      return error.message;
    }
  }

  return fallback;
}

function toOptionalBoolean(value: BooleanFilterValue): boolean | undefined {
  if (value === "true") {
    return true;
  }

  if (value === "false") {
    return false;
  }

  return undefined;
}

function toAuditTimestamp(value: string): string | undefined {
  if (!value.trim()) {
    return undefined;
  }

  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? value : parsed.toISOString();
}

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

const userFilters = reactive<UserDirectoryFilters>({
  search: "",
  active: "all",
  locked: "all",
});

const userQuery = reactive({
  offset: 0,
  limit: USER_PAGE_LIMIT,
});

const auditFilters = reactive<AuditLogFilters>({
  actorId: "",
  eventType: "",
  targetType: "",
  targetId: "",
  since: "",
  until: "",
});

const auditQuery = reactive({
  offset: 0,
  limit: AUDIT_PAGE_LIMIT,
});

const usersState = ref<AdminUserListPayload>({
  users: [],
  total: 0,
  offset: 0,
  limit: USER_PAGE_LIMIT,
});
const userMetrics = ref({
  total: 0,
  active: 0,
  locked: 0,
});
const auditState = ref<AuditListPayload>({
  events: [],
  total: 0,
  offset: 0,
  limit: AUDIT_PAGE_LIMIT,
});
const sessions = ref<AdminSession[]>([]);
const selectedUserId = ref<string | null>(null);

const usersLoading = ref(false);
const userMetricsLoading = ref(false);
const auditLoading = ref(false);
const sessionsLoading = ref(false);
const createLoading = ref(false);
const resetLoading = ref(false);
const exportLoading = ref(false);
const confirmLoading = ref(false);
const statusAction = ref<AdminUserStatusAction | null>(null);
const revokingSessionId = ref<string | null>(null);

const usersError = ref<string | null>(null);
const auditError = ref<string | null>(null);
const sessionsError = ref<string | null>(null);
const createError = ref<string | null>(null);
const resetError = ref<string | null>(null);

const createSuccessNonce = ref(0);
const resetSuccessNonce = ref(0);
const toast = ref<ToastState | null>(null);
const confirmAction = ref<ConfirmAction | null>(null);

const users = computed(() => usersState.value.users);
const selectedUser = computed<AdminUser | null>(
  () => users.value.find((user) => user.id === selectedUserId.value) ?? null,
);
const visibleRiskEvents = computed(
  () =>
    auditState.value.events.filter((event) =>
      isRiskAuditEvent(event.event_type),
    ).length,
);

const kpiMetrics = computed<SettingsKpiMetric[]>(() => [
  {
    id: "total-users",
    label: "Total users",
    value: canViewUsers.value ? userMetrics.value.total : "-",
    helper: "Directory accounts under IAM",
    icon: "accounts",
    tone: "primary",
  },
  {
    id: "active-users",
    label: "Active users",
    value: canViewUsers.value ? userMetrics.value.active : "-",
    helper: "Enabled accounts",
    icon: "check",
    tone: "success",
  },
  {
    id: "locked-users",
    label: "Locked users",
    value: canViewUsers.value ? userMetrics.value.locked : "-",
    helper: "Immediate access review",
    icon: "lock",
    tone: userMetrics.value.locked > 0 ? "warning" : "neutral",
  },
  {
    id: "audit-risk",
    label: "Audit hits",
    value: canViewAudit.value ? auditState.value.total : "-",
    helper: `${visibleRiskEvents.value} risk events visible`,
    icon: "audit",
    tone: visibleRiskEvents.value > 0 ? "danger" : "neutral",
  },
]);

const metricLoading = computed(
  () => userMetricsLoading.value || (auditLoading.value && auditState.value.total === 0),
);

const confirmDialog = computed(() => {
  const action = confirmAction.value;

  if (!action) {
    return {
      title: "",
      description: "",
      confirmLabel: "",
      tone: "neutral" as const,
    };
  }

  if (action.type === "status") {
    const label = actionLabel(action.action);
    const isDestructive = action.action === "disable" || action.action === "lock";

    return {
      title: `${label} ${action.user.username}`,
      description:
        action.action === "disable"
          ? "This prevents future sign-ins for the selected account. Existing audit evidence remains available."
          : action.action === "lock"
            ? "This immediately blocks account access until an administrator unlocks it."
            : "This updates account access state and records the administrative action in the audit trail.",
      confirmLabel: label,
      tone: isDestructive ? ("danger" as const) : ("warning" as const),
    };
  }

  if (action.type === "reset-password") {
    return {
      title: `Reset password for ${action.user.username}`,
      description:
        "The account will be forced to change this temporary password at next sign-in. Do not share it through unsecured channels.",
      confirmLabel: "Reset password",
      tone: "danger" as const,
    };
  }

  if (action.type === "revoke-session") {
    return {
      title: "Revoke active session",
      description: `This ends the session from ${
        action.session.ip_address || "an unknown IP"
      } and records the administrative revocation.`,
      confirmLabel: "Revoke session",
      tone: "warning" as const,
    };
  }

  return {
    title: "Export audit records",
    description:
      "The export contains security event metadata for the current filters. Keep the CSV in an approved evidence location.",
    confirmLabel: "Export CSV",
    tone: "warning" as const,
  };
});

function showToast(message: string, tone: ToastTone = "success") {
  toast.value = { message, tone };
}

function clearToast() {
  toast.value = null;
}

function formatDateTime(value?: string | null) {
  if (!value) {
    return t("common.notAvailable");
  }

  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? value
    : bangkokFormatter.value.format(date);
}

function statusLabel(user: AdminUser) {
  if (!user.is_active) {
    return t("settings.status.disabled");
  }

  if (user.is_locked) {
    return t("settings.status.locked");
  }

  return t("settings.status.active");
}

function statusClass(user: AdminUser) {
  if (!user.is_active) {
    return "badge-neutral";
  }

  if (user.is_locked) {
    return "badge-warning";
  }

  return "badge-success";
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

function buildUserRequest(offset = userQuery.offset): AdminUserFilters {
  return {
    search: userFilters.search,
    is_active: toOptionalBoolean(userFilters.active),
    is_locked: toOptionalBoolean(userFilters.locked),
    offset,
    limit: userQuery.limit,
  };
}

function buildAuditRequest(offset = auditQuery.offset): AuditFilters {
  return {
    actor_id: auditFilters.actorId,
    event_type: auditFilters.eventType,
    target_type: auditFilters.targetType,
    target_id: auditFilters.targetId,
    since: toAuditTimestamp(auditFilters.since),
    until: toAuditTimestamp(auditFilters.until),
    offset,
    limit: auditQuery.limit,
  };
}

function buildAuditExportRequest(): AuditFilters {
  return {
    actor_id: auditFilters.actorId,
    event_type: auditFilters.eventType,
    target_type: auditFilters.targetType,
    target_id: auditFilters.targetId,
    since: toAuditTimestamp(auditFilters.since),
    until: toAuditTimestamp(auditFilters.until),
  };
}

async function loadUserMetrics() {
  if (!canViewUsers.value) {
    return;
  }

  userMetricsLoading.value = true;

  try {
    const [total, active, locked] = await Promise.all([
      adminApi.listUsers({ offset: 0, limit: 1 }),
      adminApi.listUsers({ is_active: true, offset: 0, limit: 1 }),
      adminApi.listUsers({ is_locked: true, offset: 0, limit: 1 }),
    ]);

    userMetrics.value = {
      total: total.data.total,
      active: active.data.total,
      locked: locked.data.total,
    };
  } catch (error) {
    showToast(getErrorMessage(error, "Failed to refresh user KPIs."), "danger");
  } finally {
    userMetricsLoading.value = false;
  }
}

async function loadSessions(userId: string) {
  if (!canUpdateUsers.value) {
    sessions.value = [];
    sessionsError.value = null;
    return;
  }

  sessionsLoading.value = true;
  sessionsError.value = null;

  try {
    sessions.value = (await adminApi.listUserSessions(userId)).data;
  } catch (error) {
    sessions.value = [];
    sessionsError.value = getErrorMessage(
      error,
      "Failed to load active sessions.",
    );
  } finally {
    sessionsLoading.value = false;
  }
}

async function loadUsers(offset = userQuery.offset) {
  if (!canViewUsers.value) {
    return;
  }

  usersLoading.value = true;
  usersError.value = null;
  userQuery.offset = Math.max(0, offset);

  try {
    usersState.value = (await adminApi.listUsers(buildUserRequest(userQuery.offset))).data;

    const selectedStillVisible = usersState.value.users.some(
      (user) => user.id === selectedUserId.value,
    );
    selectedUserId.value = selectedStillVisible
      ? selectedUserId.value
      : (usersState.value.users[0]?.id ?? null);

    sessions.value = [];
    sessionsError.value = null;
    if (selectedUserId.value) {
      await loadSessions(selectedUserId.value);
    }
  } catch (error) {
    usersError.value = getErrorMessage(error, t("settings.errors.loadUsers"));
    usersState.value = {
      users: [],
      total: 0,
      offset: userQuery.offset,
      limit: userQuery.limit,
    };
    selectedUserId.value = null;
    sessions.value = [];
    sessionsError.value = null;
  } finally {
    usersLoading.value = false;
  }
}

async function loadAudit(offset = auditQuery.offset) {
  if (!canViewAudit.value) {
    return;
  }

  auditLoading.value = true;
  auditError.value = null;
  auditQuery.offset = Math.max(0, offset);

  try {
    auditState.value = (await auditApi.listEvents(buildAuditRequest(auditQuery.offset))).data;
  } catch (error) {
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

async function refreshAfterMutation() {
  const tasks: Promise<unknown>[] = [];

  if (canViewUsers.value) {
    tasks.push(loadUsers(userQuery.offset));
    tasks.push(loadUserMetrics());
  }

  if (canViewAudit.value) {
    tasks.push(loadAudit(auditQuery.offset));
  }

  await Promise.all(tasks);
}

async function handleCreateUser(payload: CreateAdminUserInput) {
  if (!canCreateUsers.value) {
    return;
  }

  createLoading.value = true;
  createError.value = null;
  clearToast();

  try {
    const created = await adminApi.createUser(payload);
    createSuccessNonce.value += 1;
    showToast(t("settings.notices.userCreated"), "success");

    if (canViewUsers.value) {
      await loadUsers(0);
      await loadUserMetrics();

      if (usersState.value.users.some((user) => user.id === created.id)) {
        selectedUserId.value = created.id;
        await loadSessions(created.id);
      }
    }

    if (canViewAudit.value) {
      await loadAudit(0);
    }
  } catch (error) {
    createError.value = getErrorMessage(error, t("settings.errors.createUser"));
    showToast(createError.value, "danger");
  } finally {
    createLoading.value = false;
  }
}

function requestStatusChange(action: AdminUserStatusAction) {
  if (!selectedUser.value) {
    return;
  }

  confirmAction.value = {
    type: "status",
    action,
    user: selectedUser.value,
  };
}

function requestPasswordReset(newPassword: string) {
  if (!selectedUser.value) {
    return;
  }

  confirmAction.value = {
    type: "reset-password",
    user: selectedUser.value,
    newPassword,
  };
}

function requestSessionRevoke(session: AdminSession) {
  if (!selectedUser.value) {
    return;
  }

  confirmAction.value = {
    type: "revoke-session",
    user: selectedUser.value,
    session,
  };
}

function requestAuditExport() {
  confirmAction.value = {
    type: "export-audit",
  };
}

async function changeUserStatus(user: AdminUser, action: AdminUserStatusAction) {
  statusAction.value = action;
  clearToast();

  try {
    await adminApi.setUserStatus(user.id, action);
    showToast(
      t("settings.notices.userActionSuccess", { action: actionLabel(action) }),
      "success",
    );
    await refreshAfterMutation();
  } catch (error) {
    showToast(
      getErrorMessage(error, t("settings.errors.changeUserStatus")),
      "danger",
    );
  } finally {
    statusAction.value = null;
  }
}

async function resetPassword(user: AdminUser, newPassword: string) {
  resetLoading.value = true;
  resetError.value = null;
  clearToast();

  try {
    await adminApi.resetPassword(user.id, newPassword);
    resetSuccessNonce.value += 1;
    showToast(t("settings.notices.passwordReset"), "success");
    await refreshAfterMutation();
  } catch (error) {
    resetError.value = getErrorMessage(error, t("settings.errors.resetPassword"));
    showToast(resetError.value, "danger");
  } finally {
    resetLoading.value = false;
  }
}

async function revokeSession(session: AdminSession) {
  revokingSessionId.value = session.id;
  clearToast();

  try {
    await adminApi.revokeSession(session.id);
    showToast(t("settings.notices.sessionRevoked"), "success");

    if (selectedUser.value) {
      await loadSessions(selectedUser.value.id);
    }

    if (canViewAudit.value) {
      await loadAudit(auditQuery.offset);
    }
  } catch (error) {
    showToast(
      getErrorMessage(error, t("settings.errors.revokeSession")),
      "danger",
    );
  } finally {
    revokingSessionId.value = null;
  }
}

function downloadAuditExport(blob: Blob) {
  if (!import.meta.client) {
    return;
  }

  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = `iam_audit_${new Date().toISOString().replace(/[:.]/g, "-")}.csv`;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}

async function exportAudit() {
  exportLoading.value = true;
  clearToast();

  try {
    const blob = await auditApi.exportEvents(buildAuditExportRequest());
    downloadAuditExport(blob);
    showToast("Audit export started.", "success");

    if (canViewAudit.value) {
      await loadAudit(auditQuery.offset);
    }
  } catch (error) {
    showToast(getErrorMessage(error, "Failed to export audit events."), "danger");
  } finally {
    exportLoading.value = false;
  }
}

async function confirmCurrentAction() {
  const action = confirmAction.value;

  if (!action || confirmLoading.value) {
    return;
  }

  confirmLoading.value = true;

  try {
    if (action.type === "status") {
      await changeUserStatus(action.user, action.action);
    } else if (action.type === "reset-password") {
      await resetPassword(action.user, action.newPassword);
    } else if (action.type === "revoke-session") {
      await revokeSession(action.session);
    } else {
      await exportAudit();
    }

    confirmAction.value = null;
  } finally {
    confirmLoading.value = false;
  }
}

function cancelConfirm() {
  if (!confirmLoading.value) {
    confirmAction.value = null;
  }
}

function applyUserFilters(nextFilters: UserDirectoryFilters) {
  userFilters.search = nextFilters.search;
  userFilters.active = nextFilters.active;
  userFilters.locked = nextFilters.locked;
  void loadUsers(0);
}

function applyAuditFilters(nextFilters: AuditLogFilters) {
  auditFilters.actorId = nextFilters.actorId;
  auditFilters.eventType = nextFilters.eventType;
  auditFilters.targetType = nextFilters.targetType;
  auditFilters.targetId = nextFilters.targetId;
  auditFilters.since = nextFilters.since;
  auditFilters.until = nextFilters.until;
  void loadAudit(0);
}

async function selectUser(userId: string) {
  selectedUserId.value = userId;
  resetError.value = null;
  await loadSessions(userId);
}

async function refreshUsers() {
  await Promise.all([loadUsers(userQuery.offset), loadUserMetrics()]);
}

onMounted(async () => {
  const tasks: Promise<unknown>[] = [];

  if (canViewUsers.value) {
    tasks.push(loadUsers(0));
    tasks.push(loadUserMetrics());
  }

  if (canViewAudit.value) {
    tasks.push(loadAudit(0));
  }

  await Promise.all(tasks);
});
</script>

<template>
  <div class="settings-control-center">
    <AppPageHeader
      :title="t('settings.title')"
      :description="t('settings.description')"
    >
      <template #eyebrow>
        <span class="settings-eyebrow">{{ t("settings.breadcrumb") }}</span>
      </template>
      <template #actions>
        <AppButton
          v-if="canViewUsers"
          variant="secondary"
          size="sm"
          :loading="usersLoading || userMetricsLoading"
          @click="refreshUsers"
        >
          <AppIcon name="refresh" size="xs" />
          <span>{{ t("settings.refreshUsers") }}</span>
        </AppButton>
        <AppButton
          v-if="canViewAudit"
          variant="secondary"
          size="sm"
          :loading="auditLoading"
          @click="loadAudit(auditQuery.offset)"
        >
          <AppIcon name="refresh" size="xs" />
          <span>{{ t("settings.refreshAudit") }}</span>
        </AppButton>
      </template>
    </AppPageHeader>

    <div
      v-if="toast"
      class="settings-toast"
      :class="`settings-toast--${toast.tone}`"
      role="status"
      aria-live="polite"
    >
      <span>{{ toast.message }}</span>
      <button
        class="settings-toast__close"
        type="button"
        aria-label="Dismiss notification"
        @click="clearToast"
      >
        <AppIcon name="close" size="xs" />
      </button>
    </div>

    <SettingsKpiGrid :metrics="kpiMetrics" :loading="metricLoading" />

    <section class="settings-workspace">
      <SettingsUserDirectory
        v-if="canViewUsers"
        :users="users"
        :total="usersState.total"
        :offset="usersState.offset"
        :limit="usersState.limit"
        :selected-user-id="selectedUserId"
        :loading="usersLoading"
        :error="usersError"
        :filters="userFilters"
        :format-date-time="formatDateTime"
        :status-label="statusLabel"
        :status-class="statusClass"
        @apply="applyUserFilters"
        @page="loadUsers"
        @select="selectUser"
      />

      <div class="settings-side-stack">
        <SettingsCreateUserForm
          v-if="canCreateUsers"
          :loading="createLoading"
          :error="createError"
          :success-nonce="createSuccessNonce"
          @submit="handleCreateUser"
        />

        <SettingsUserDetailPanel
          v-if="canViewUsers"
          :user="selectedUser"
          :sessions="sessions"
          :sessions-loading="sessionsLoading"
          :sessions-error="sessionsError"
          :can-deactivate-users="canDeactivateUsers"
          :can-update-users="canUpdateUsers"
          :status-action="statusAction"
          :reset-loading="resetLoading"
          :revoking-session-id="revokingSessionId"
          :reset-error="resetError"
          :reset-success-nonce="resetSuccessNonce"
          :format-date-time="formatDateTime"
          :status-label="statusLabel"
          :status-class="statusClass"
          @status="requestStatusChange"
          @reset-password="requestPasswordReset"
          @revoke-session="requestSessionRevoke"
        />
      </div>
    </section>

    <SettingsAuditLog
      v-if="canViewAudit"
      :events="auditState.events"
      :total="auditState.total"
      :offset="auditState.offset"
      :limit="auditState.limit"
      :loading="auditLoading"
      :exporting="exportLoading"
      :error="auditError"
      :filters="auditFilters"
      :format-date-time="formatDateTime"
      @apply="applyAuditFilters"
      @page="loadAudit"
      @refresh="loadAudit(auditQuery.offset)"
      @export="requestAuditExport"
    />

    <SettingsConfirmDialog
      :open="Boolean(confirmAction)"
      :title="confirmDialog.title"
      :description="confirmDialog.description"
      :confirm-label="confirmDialog.confirmLabel"
      :tone="confirmDialog.tone"
      :loading="confirmLoading"
      @cancel="cancelConfirm"
      @confirm="confirmCurrentAction"
    />
  </div>
</template>

<style scoped>
.settings-control-center {
  display: grid;
  gap: var(--space-5);
  padding-bottom: var(--space-8);
}

.settings-eyebrow {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.settings-workspace {
  display: grid;
  grid-template-columns: minmax(0, 1.65fr) minmax(20rem, 0.85fr);
  gap: var(--space-4);
  align-items: start;
}

.settings-side-stack {
  display: grid;
  gap: var(--space-4);
  min-width: 0;
}

.settings-toast {
  display: flex;
  justify-content: space-between;
  gap: var(--space-4);
  align-items: center;
  padding: var(--space-4) var(--space-5);
  border: 1px solid var(--alert-info-border);
  border-radius: var(--radius-md);
  background: var(--alert-info-bg);
  color: var(--alert-info-text);
}

.settings-toast--success {
  border-color: var(--alert-success-border);
  background: var(--alert-success-bg);
  color: var(--alert-success-text);
}

.settings-toast--danger {
  border-color: var(--alert-danger-border);
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
}

.settings-toast__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  flex: 0 0 auto;
  border-radius: var(--radius-sm);
  color: currentColor;
}

.settings-toast__close:hover {
  background: var(--action-ghost-hover);
}

.settings-control-center :deep(.settings-panel) {
  overflow: hidden;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  box-shadow: var(--shadow-xs);
}

.settings-control-center :deep(.settings-panel__header),
.settings-control-center :deep(.settings-panel__footer) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-5);
}

.settings-control-center :deep(.settings-panel__header) {
  border-bottom: 1px solid var(--border-subtle);
}

.settings-control-center :deep(.settings-panel__footer) {
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-card-muted);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.settings-control-center :deep(.settings-panel__title) {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
}

.settings-control-center :deep(.settings-panel__subtitle) {
  margin: var(--space-1) 0 0;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-relaxed);
}

.settings-control-center :deep(.settings-panel__alert) {
  margin: var(--space-5) var(--space-5) 0;
}

.settings-control-center :deep(.settings-pagination) {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-3);
}

.settings-control-center :deep(.table) {
  min-width: 760px;
}

.settings-control-center :deep(.table th),
.settings-control-center :deep(.table td) {
  vertical-align: top;
}

@media (max-width: 1180px) {
  .settings-workspace {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .settings-control-center :deep(.settings-panel__header),
  .settings-control-center :deep(.settings-panel__footer),
  .settings-toast {
    align-items: stretch;
    flex-direction: column;
  }

  .settings-control-center :deep(.settings-pagination) {
    width: 100%;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
