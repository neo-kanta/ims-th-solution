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
import type {
  ChangePersonalPasswordInput,
  PersonalAccountPayload,
  PersonalAccountSession,
  PersonalMfaEnrollResult,
  PersonalMfaStatus,
  PersonalMfaTotpInput,
} from "../account.types";
import type { AuditFilters, AuditListPayload } from "../audit.types";
import type {
  AuditLogFilters,
  BooleanFilterValue,
  DataPermissionGrant,
  FunctionPermissionRow,
  NotificationPreference,
  PermissionActionColumn,
  SettingsKpiMetric,
  SettingsGroupRole,
  SettingsConsoleMode,
  SettingsNavigationGroup,
  SettingsNavigationItem,
  SettingsOverviewSignal,
  SecurityPolicyModel,
  SecurityPolicySetting,
  SettingsSectionId,
  SettingsApiCapability,
  UserDirectoryFilters,
} from "../ui.types";
import { adminApi } from "../services/adminApi";
import { accountApi } from "../services/accountApi";
import { auditApi } from "../services/auditApi";
import { mfaApi } from "../services/mfaApi";
import SettingsDataPermissionsPanel from "./SettingsDataPermissionsPanel.vue";
import SettingsFunctionPermissionsPanel from "./SettingsFunctionPermissionsPanel.vue";
import SettingsGroupsRolesPanel from "./SettingsGroupsRolesPanel.vue";
import SettingsAuditLog from "./SettingsAuditLog.vue";
import SettingsConfirmDialog from "./SettingsConfirmDialog.vue";
import SettingsCreateUserForm from "./SettingsCreateUserForm.vue";
import SettingsMfaDisableDialog from "./SettingsMfaDisableDialog.vue";
import SettingsMfaEnrollDialog from "./SettingsMfaEnrollDialog.vue";
import SettingsNotificationsPanel from "./SettingsNotificationsPanel.vue";
import SettingsOverviewPanel from "./SettingsOverviewPanel.vue";
import SettingsPersonalAccountPanel from "./SettingsPersonalAccountPanel.vue";
import SettingsSecurityPolicyPanel from "./SettingsSecurityPolicyPanel.vue";
import SettingsSectionNav from "./SettingsSectionNav.vue";
import SettingsUserDetailPanel from "./SettingsUserDetailPanel.vue";
import SettingsUserDirectory from "./SettingsUserDirectory.vue";
import { useBangkokFormatter } from "~/shared/composables/useBangkokFormatter";
import { useSettingsActiveSection } from "../composables/useSettingsActiveSection";
import { useSettingsConfirm } from "../composables/useSettingsConfirm";
import { useSettingsToasts } from "../composables/useSettingsToasts";
import { useSettingsUserMetrics } from "../composables/useSettingsUserMetrics";
import { isRiskAuditEvent } from "../lib/audit";
import {
  SETTINGS_API_CAPABILITIES,
  SETTINGS_DEMO_DATA_GRANTS,
  SETTINGS_DEMO_GROUPS,
  SETTINGS_FUNCTION_PERMISSION_ROWS,
  SETTINGS_NOTIFICATION_PREFERENCES,
  SETTINGS_PERMISSION_ACTIONS,
  SETTINGS_SECURITY_POLICY,
} from "../lib/settingsCatalog";

const props = withDefaults(
  defineProps<{
    mode?: SettingsConsoleMode;
  }>(),
  {
    mode: "administration",
  },
);

const INDIVIDUAL_SETTINGS_SECTIONS = ["personal-account"] as const;
const ADMINISTRATION_SETTINGS_SECTIONS = [
  "overview",
  "users",
  "groups",
  "function-permissions",
  "data-permissions",
  "security-policy",
  "audit",
] as const;

const USER_PAGE_LIMIT = 12;
const AUDIT_PAGE_LIMIT = 12;
const PAGE_SIZE_OPTIONS = [12, 25, 50, 100] as const;

function clampPageSize(next: number): number {
  if (!Number.isFinite(next) || next <= 0) return USER_PAGE_LIMIT;
  return PAGE_SIZE_OPTIONS.includes(next as (typeof PAGE_SIZE_OPTIONS)[number])
    ? next
    : USER_PAGE_LIMIT;
}

// Toast types are owned by the useSettingsToasts composable.

// ConfirmAction union and reset-password storage are owned by useSettingsConfirm.

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

interface ServerFieldError {
  field: string;
  message: string;
}

function getServerFieldErrors(error: unknown): ServerFieldError[] {
  if (!isRecord(error)) return [];
  const data = error.data;
  if (!isRecord(data)) return [];
  const details = data.details;
  if (!Array.isArray(details)) return [];

  const result: ServerFieldError[] = [];
  for (const item of details) {
    if (
      isRecord(item) &&
      typeof item.field === "string" &&
      typeof item.message === "string"
    ) {
      result.push({ field: item.field, message: item.message });
    }
  }
  return result;
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

function normalizeRoleId(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
}

function humanizeGroupName(value: string): string {
  return value
    .trim()
    .replace(/[_-]+/g, " ")
    .replace(/\s+/g, " ")
    .replace(/\b\w/g, (character) => character.toUpperCase());
}

// Group attributes (responsibilities, permission families, risk level, data scopes)
// require a real backend feed. Until the groups API exists, render them as
// "not configured" instead of inferring from substring matches on the group name —
// fabricated attributes are misleading in a compliance-facing IAM screen.

const authStore = useAuthStore();
const { t } = useI18n();
const runtimeConfig = useRuntimeConfig();
const isIndividualSettings = computed(() => props.mode === "individual");
const isAdministrationSettings = computed(() => props.mode === "administration");
const { activeSection } = useSettingsActiveSection({
  validSections: isIndividualSettings.value
    ? INDIVIDUAL_SETTINGS_SECTIONS
    : ADMINISTRATION_SETTINGS_SECTIONS,
  defaultSection: isIndividualSettings.value ? "personal-account" : "overview",
});

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
// Auth middleware on the /settings page handles unauthenticated redirects.
// Section visibility is gated by per-permission `v-if`/`v-show` below.

const { formatDateTime } = useBangkokFormatter();

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
// `userMetrics` is owned by useSettingsUserMetrics.
const auditState = ref<AuditListPayload>({
  events: [],
  total: 0,
  offset: 0,
  limit: AUDIT_PAGE_LIMIT,
});
const personalAccount = ref<PersonalAccountPayload | null>(null);
const personalMfaStatus = ref<PersonalMfaStatus | null>(null);
const personalSessions = ref<PersonalAccountSession[]>([]);
const sessions = ref<AdminSession[]>([]);
const selectedUserId = ref<string | null>(null);

const personalLoading = ref(false);
const personalPasswordLoading = ref(false);
const usersLoading = ref(false);
const auditLoading = ref(false);
const sessionsLoading = ref(false);
const createLoading = ref(false);
const resetLoading = ref(false);
const exportLoading = ref(false);
const statusAction = ref<AdminUserStatusAction | null>(null);
const revokingSessionId = ref<string | null>(null);

const usersError = ref<string | null>(null);
const personalError = ref<string | null>(null);
const personalMfaError = ref<string | null>(null);
const personalSessionsError = ref<string | null>(null);
const personalPasswordError = ref<string | null>(null);
const auditError = ref<string | null>(null);
const sessionsError = ref<string | null>(null);
const createError = ref<string | null>(null);
const resetError = ref<string | null>(null);

const createSuccessNonce = ref(0);
const resetSuccessNonce = ref(0);
const personalPasswordSuccessNonce = ref(0);
const revokingOwnSessionId = ref<string | null>(null);
const { toasts, show: showToast, dismiss: dismissToast, clear: clearToast } = useSettingsToasts();

const {
  confirmAction,
  confirmLoading,
  openStatusChange,
  openPasswordReset,
  openSessionRevoke,
  openOwnSessionRevoke,
  openAuditExport,
  cancelConfirm,
  confirm: runConfirm,
} = useSettingsConfirm();

const { userMetrics, userMetricsLoading, loadUserMetrics } = useSettingsUserMetrics({
  enabled: () => canViewUsers.value,
  onError: (error) =>
    showToast(
      getErrorMessage(error, t("settings.console.errors.refreshUserKpis")),
      "danger",
    ),
});

const mfaEnrollOpen = ref(false);
const mfaDisableOpen = ref(false);
const mfaEnrollment = ref<PersonalMfaEnrollResult | null>(null);
const mfaEnrollLoading = ref(false);
const mfaVerifyLoading = ref(false);
const mfaDisableLoading = ref(false);
const mfaEnrollError = ref<string | null>(null);
const mfaVerifyError = ref<string | null>(null);
const mfaDisableError = ref<string | null>(null);

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
const observedGroupCount = computed(() => {
  const groupNames = new Set<string>();

  for (const user of users.value) {
    for (const group of user.groups) {
      if (group.trim()) {
        groupNames.add(group.trim());
      }
    }
  }

  return groupNames.size;
});
const currentSessionFunctionCount = computed(
  () => authStore.permissions.functions.length,
);
const currentSessionContractCount = computed(
  () => authStore.permissions.contracts.length,
);
// `userMetrics.locked` is a backend total. `visibleRiskEvents` is page-local
// and is shown separately in the overview signal, not summed here.
const pendingReviewCount = computed(() => userMetrics.value.locked);

const kpiMetrics = computed<SettingsKpiMetric[]>(() => [
  {
    id: "total-users",
    label: t("settings.console.kpis.totalUsers"),
    value: canViewUsers.value ? userMetrics.value.total : "-",
    helper: t("settings.console.kpis.totalUsersHelper"),
    icon: "accounts",
    tone: "primary",
  },
  {
    id: "active-users",
    label: t("settings.console.kpis.activeUsers"),
    value: canViewUsers.value ? userMetrics.value.active : "-",
    helper: t("settings.console.kpis.activeUsersHelper"),
    icon: "check",
    tone: "success",
  },
  {
    id: "locked-users",
    label: t("settings.console.kpis.lockedUsers"),
    value: canViewUsers.value ? userMetrics.value.locked : "-",
    helper: t("settings.console.kpis.lockedUsersHelper"),
    icon: "lock",
    tone: userMetrics.value.locked > 0 ? "warning" : "neutral",
  },
  {
    id: "groups-roles",
    label: t("settings.console.kpis.groupsRoles"),
    value: canViewUsers.value ? observedGroupCount.value : "-",
    helper: t("settings.console.kpis.groupsRolesHelper"),
    icon: "groups",
    tone: "neutral",
  },
  {
    id: "pending-review",
    label: t("settings.console.kpis.pendingReview"),
    value:
      canViewUsers.value || canViewAudit.value ? pendingReviewCount.value : "-",
    helper: t("settings.console.kpis.pendingReviewHelper"),
    icon: "review",
    tone: pendingReviewCount.value > 0 ? "warning" : "success",
  },
  {
    id: "coverage",
    label: t("settings.console.kpis.permissionCoverage"),
    value:
      currentSessionFunctionCount.value + currentSessionContractCount.value,
    helper: t("settings.console.kpis.permissionCoverageHelper", {
      functionCount: currentSessionFunctionCount.value,
      dataScopeCount: currentSessionContractCount.value,
    }),
    icon: "shield",
    tone: "primary",
  },
]);

const metricLoading = computed(
  () =>
    userMetricsLoading.value ||
    (auditLoading.value && auditState.value.total === 0),
);
const settingsTitle = computed(() =>
  isIndividualSettings.value
    ? t("settings.individualTitle")
    : t("settings.administrationTitle"),
);
const settingsDescription = computed(() =>
  isIndividualSettings.value
    ? t("settings.individualDescription")
    : t("settings.administrationDescription"),
);
const settingsEyebrow = computed(() =>
  isIndividualSettings.value
    ? t("settings.individualBreadcrumb")
    : t("settings.administrationBreadcrumb"),
);
const appName = computed(() =>
  String(runtimeConfig.public.appName || "IMS Thailand"),
);
const apiBaseUrl = computed(() =>
  String(
    runtimeConfig.public.apiBaseUrl ||
      runtimeConfig.apiBaseUrl ||
      t("settings.console.common.unavailable"),
  ),
);
const clientTimezone = computed(() => {
  if (!import.meta.client) {
    return "Asia/Bangkok";
  }

  return Intl.DateTimeFormat().resolvedOptions().timeZone || "Asia/Bangkok";
});
const settingsAccountName = computed(
  () =>
    personalAccount.value?.user.display_name ||
    authStore.user?.displayName ||
    t("auth.welcome"),
);
const settingsAccountSubtitle = computed(
  () =>
    personalAccount.value?.user.username ||
    authStore.user?.username ||
    t("shell.activeSession"),
);
const settingsAccountInitials = computed(() => {
  const source = settingsAccountName.value || settingsAccountSubtitle.value;
  const parts = source.split(/\s+/).filter(Boolean).slice(0, 2);

  if (parts.length === 0) {
    return "IM";
  }

  return parts.map((part) => part[0]?.toUpperCase() || "").join("");
});
const settingsNavItems = computed<Record<SettingsSectionId, SettingsNavigationItem>>(() => ({
  overview: {
    id: "overview",
    label: t("settings.console.nav.overview"),
    description: t("settings.console.nav.overviewDesc"),
    icon: "dashboard",
    status: "live",
  },
  "personal-account": {
    id: "personal-account",
    label: t("settings.console.nav.personalAccount"),
    description: t("settings.console.nav.personalAccountDesc"),
    icon: "user",
    status: "live",
  },
  users: {
    id: "users",
    label: t("settings.console.nav.otherAccounts"),
    description: t("settings.console.nav.otherAccountsDesc"),
    icon: "accounts",
    status: canViewUsers.value || canCreateUsers.value ? "live" : "pending",
    count: canViewUsers.value ? usersState.value.total : undefined,
    disabled: !(canViewUsers.value || canCreateUsers.value),
  },
  groups: {
    id: "groups",
    label: t("settings.console.nav.groupsRoles"),
    description: t("settings.console.nav.groupsRolesDesc"),
    icon: "groups",
    status: "read-only",
    count: observedGroupCount.value,
    disabled: !canViewUsers.value,
  },
  "function-permissions": {
    id: "function-permissions",
    label: t("settings.console.nav.functionPermissions"),
    description: t("settings.console.nav.functionPermissionsDesc"),
    icon: "workflow",
    status: "read-only",
    count: currentSessionFunctionCount.value,
  },
  "data-permissions": {
    id: "data-permissions",
    label: t("settings.console.nav.dataPermissions"),
    description: t("settings.console.nav.dataPermissionsDesc"),
    icon: "portfolio",
    status: "read-only",
    count: currentSessionContractCount.value,
  },
  "security-policy": {
    id: "security-policy",
    label: t("settings.console.nav.securityPolicy"),
    description: t("settings.console.nav.securityPolicyDesc"),
    icon: "shield",
    status: "pending",
  },
  notifications: {
    id: "notifications",
    label: t("settings.console.nav.notifications"),
    description: t("settings.console.nav.notificationsDesc"),
    icon: "notifications",
    status: "pending",
  },
  audit: {
    id: "audit",
    label: t("settings.console.nav.auditLogs"),
    description: t("settings.console.nav.auditLogsDesc"),
    icon: "audit",
    status: canViewAudit.value ? "live" : "pending",
    count: canViewAudit.value ? auditState.value.total : undefined,
    disabled: !canViewAudit.value,
  },
}));
const settingsNavGroups = computed<SettingsNavigationGroup[]>(() => {
  const items = settingsNavItems.value;

  if (isIndividualSettings.value) {
    return [
      {
        id: "individual",
        label: t("settings.console.nav.individualGroup"),
        items: [items["personal-account"]],
      },
    ];
  }

  return [
    {
      id: "administration",
      label: t("settings.console.nav.administrationGroup"),
      items: [items.overview, items.users],
    },
    {
      id: "access",
      label: t("settings.console.nav.accessGroup"),
      items: [
        items.groups,
        items["function-permissions"],
        items["data-permissions"],
      ],
    },
    {
      id: "system",
      label: t("settings.console.nav.systemGroup"),
      items: [items["security-policy"], items.audit],
    },
  ];
});
const overviewSignals = computed<SettingsOverviewSignal[]>(() => [
  {
    id: "api-users",
    label: t("settings.console.overview.signals.userApi"),
    value: canViewUsers.value
      ? t("settings.console.common.available")
      : t("settings.console.common.restricted"),
    helper: canViewUsers.value
      ? t("settings.console.overview.signals.userApiAvailable")
      : t("settings.console.overview.signals.userApiRestricted"),
    tone: canViewUsers.value ? "success" : "warning",
  },
  {
    id: "locked-users",
    label: t("settings.console.overview.signals.lockedUsers"),
    value: canViewUsers.value ? String(userMetrics.value.locked) : "-",
    helper:
      userMetrics.value.locked > 0
        ? t("settings.console.overview.signals.lockedUsersReview")
        : t("settings.console.overview.signals.lockedUsersClear"),
    tone: userMetrics.value.locked > 0 ? "warning" : "success",
  },
  {
    id: "audit-risk",
    label: t("settings.console.overview.signals.auditRiskCurrentPage"),
    value: canViewAudit.value ? String(visibleRiskEvents.value) : "-",
    helper: t("settings.console.overview.signals.auditRiskCurrentPageHelper"),
    tone: visibleRiskEvents.value > 0 ? "danger" : "neutral",
  },
  {
    id: "data-access",
    label: t("settings.console.overview.signals.dataAccess"),
    value: String(currentSessionContractCount.value),
    helper: t("settings.console.overview.signals.dataAccessHelper"),
    tone: currentSessionContractCount.value > 0 ? "info" : "neutral",
  },
]);

function apiCapabilityCopy(
  capability: SettingsApiCapability,
): Pick<SettingsApiCapability, "label" | "capability"> {
  switch (capability.id) {
    case "personal-account":
      return {
        label: t("settings.console.overview.capabilities.personalAccount"),
        capability: t(
          "settings.console.overview.capabilities.personalAccountCapability",
        ),
      };
    case "users":
      return {
        label: t("settings.console.overview.capabilities.otherAccounts"),
        capability: t(
          "settings.console.overview.capabilities.otherAccountsCapability",
        ),
      };
    case "sessions":
      return {
        label: t("settings.console.overview.capabilities.userSessions"),
        capability: t(
          "settings.console.overview.capabilities.userSessionsCapability",
        ),
      };
    case "audit":
      return {
        label: t("settings.console.overview.capabilities.audit"),
        capability: t("settings.console.overview.capabilities.auditCapability"),
      };
    case "groups":
      return {
        label: t("settings.console.overview.capabilities.groups"),
        capability: t(
          "settings.console.overview.capabilities.groupsCapability",
        ),
      };
    case "permissions":
      return {
        label: t("settings.console.overview.capabilities.permissions"),
        capability: t(
          "settings.console.overview.capabilities.permissionsCapability",
        ),
      };
    case "policy":
      return {
        label: t("settings.console.overview.capabilities.policy"),
        capability: t(
          "settings.console.overview.capabilities.policyCapability",
        ),
      };
    default:
      return {
        label: capability.label,
        capability: capability.capability,
      };
  }
}

const apiCapabilities = computed<SettingsApiCapability[]>(() =>
  SETTINGS_API_CAPABILITIES.map((capability) => ({
    ...capability,
    ...apiCapabilityCopy(capability),
  })),
);

function permissionActionLabel(action: PermissionActionColumn["key"]) {
  switch (action) {
    case "view":
      return t("settings.console.functionPermissions.actions.view");
    case "search":
      return t("settings.console.functionPermissions.actions.search");
    case "add":
      return t("settings.console.functionPermissions.actions.add");
    case "edit":
      return t("settings.console.functionPermissions.actions.edit");
    case "delete":
      return t("settings.console.functionPermissions.actions.delete");
    case "approve":
      return t("settings.console.functionPermissions.actions.approve");
    case "revokeApproval":
      return t("settings.console.functionPermissions.actions.revokeApproval");
    case "export":
      return t("settings.console.functionPermissions.actions.export");
    case "settings":
      return t("settings.console.functionPermissions.actions.settings");
  }
}

const permissionActions = computed<PermissionActionColumn[]>(() =>
  SETTINGS_PERMISSION_ACTIONS.map((action) => ({
    ...action,
    label: permissionActionLabel(action.key),
  })),
);

function localizedPolicySetting(
  setting: SecurityPolicySetting,
): SecurityPolicySetting {
  switch (setting.id) {
    case "minimum-length":
      return {
        ...setting,
        label: t("settings.console.securityPolicy.settings.minimumLength"),
        helper: t(
          "settings.console.securityPolicy.settings.minimumLengthHelper",
        ),
        unit: t("settings.console.securityPolicy.settings.characters"),
      };
    case "expiry-days":
      return {
        ...setting,
        label: t("settings.console.securityPolicy.settings.expiryDays"),
        helper: t("settings.console.securityPolicy.settings.expiryDaysHelper"),
        unit: t("settings.console.securityPolicy.settings.days"),
      };
    case "reuse":
      return {
        ...setting,
        label: t("settings.console.securityPolicy.settings.reuse"),
        helper: t("settings.console.securityPolicy.settings.reuseHelper"),
        unit: t("settings.console.securityPolicy.settings.previousPasswords"),
      };
    case "failed-attempts":
      return {
        ...setting,
        label: t("settings.console.securityPolicy.settings.failedAttempts"),
        helper: t(
          "settings.console.securityPolicy.settings.failedAttemptsHelper",
        ),
      };
    case "change-interval":
      return {
        ...setting,
        label: t("settings.console.securityPolicy.settings.changeInterval"),
        helper: t(
          "settings.console.securityPolicy.settings.changeIntervalHelper",
        ),
        unit: t("settings.console.securityPolicy.settings.day"),
      };
    case "enforce-rule":
      return {
        ...setting,
        label: t("settings.console.securityPolicy.settings.enforceRule"),
        helper: t("settings.console.securityPolicy.settings.enforceRuleHelper"),
      };
    default:
      return setting;
  }
}

const securityPolicy = computed<SecurityPolicyModel>(() => ({
  ...SETTINGS_SECURITY_POLICY,
  settings: SETTINGS_SECURITY_POLICY.settings.map(localizedPolicySetting),
}));

function localizedNotification(
  preference: NotificationPreference,
): NotificationPreference {
  switch (preference.id) {
    case "account-lock":
      return {
        ...preference,
        event: t("settings.console.notifications.events.accountLock"),
        audience: t(
          "settings.console.notifications.audiences.iamAdministrators",
        ),
        channels: preference.channels.map(localizedNotificationChannel),
      };
    case "password-reset":
      return {
        ...preference,
        event: t("settings.console.notifications.events.passwordReset"),
        audience: t(
          "settings.console.notifications.audiences.securityAdministrators",
        ),
        channels: preference.channels.map(localizedNotificationChannel),
      };
    case "audit-export":
      return {
        ...preference,
        event: t("settings.console.notifications.events.auditExport"),
        audience: t("settings.console.notifications.audiences.auditors"),
        channels: preference.channels.map(localizedNotificationChannel),
      };
    case "permission-review":
      return {
        ...preference,
        event: t("settings.console.notifications.events.permissionReview"),
        audience: t("settings.console.notifications.audiences.groupOwners"),
        channels: preference.channels.map(localizedNotificationChannel),
      };
    default:
      return {
        ...preference,
        channels: preference.channels.map(localizedNotificationChannel),
      };
  }
}

function localizedNotificationChannel(channel: string) {
  if (channel === "In-app") {
    return t("settings.console.notifications.channels.inApp");
  }

  if (channel === "Email") {
    return t("settings.console.notifications.channels.email");
  }

  return channel;
}

const notificationPreferences = computed<NotificationPreference[]>(() =>
  SETTINGS_NOTIFICATION_PREFERENCES.map(localizedNotification),
);

const derivedGroups = computed<SettingsGroupRole[]>(() => {
  const groups = new Map<string, SettingsGroupRole>();

  for (const user of users.value) {
    for (const groupName of user.groups) {
      const trimmedName = groupName.trim();

      if (!trimmedName) {
        continue;
      }

      const id = normalizeRoleId(trimmedName);
      const existing = groups.get(id);

      if (existing) {
        existing.membersCount += 1;
      } else {
        groups.set(id, {
          id,
          name: humanizeGroupName(trimmedName),
          description: t(
            "settings.console.groupsRoles.directoryDerivedDescription",
            {
              group: trimmedName,
            },
          ),
          source: "directory",
          membersCount: 1,
          // Responsibilities, permissionFamilies, dataScopes, riskLevel are
          // intentionally empty: the groups API does not expose them yet.
          responsibilities: [],
          permissionFamilies: [],
          dataScopes: [],
          riskLevel: "standard",
        });
      }
    }
  }

  return groups.size > 0
    ? Array.from(groups.values()).sort((a, b) => a.name.localeCompare(b.name))
    : SETTINGS_DEMO_GROUPS;
});
const directoryCompleteForGroups = computed(
  () =>
    canViewUsers.value &&
    usersState.value.total > 0 &&
    usersState.value.users.length >= usersState.value.total,
);
const functionPermissionRows = computed<FunctionPermissionRow[]>(() => {
  const sessionFunctions = new Set(authStore.permissions.functions);

  return SETTINGS_FUNCTION_PERMISSION_ROWS.map((row) => ({
    ...row,
    status: sessionFunctions.has(`${row.id.toUpperCase()}_VIEW`)
      ? "live-session"
      : row.status,
    permissions: { ...row.permissions },
  }));
});
const dataPermissionGrants = computed<DataPermissionGrant[]>(() => {
  const currentUser = authStore.user;
  const sessionGrants = authStore.permissions.contracts.map((contractId) => ({
    id: `session-${contractId}`,
    user: currentUser?.username ?? "current.user",
    userLabel:
      currentUser?.displayName ?? t("settings.console.common.currentSession"),
    contractId,
    contractName: contractId,
    scope: "Read only" as const,
    source: "session" as const,
    updatedAt: t("settings.console.common.currentSession"),
  }));

  return [...sessionGrants, ...SETTINGS_DEMO_DATA_GRANTS];
});

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
    const isDestructive =
      action.action === "disable" || action.action === "lock";

    return {
      title: t("settings.console.confirm.statusTitle", {
        action: label,
        username: action.user.username,
      }),
      description:
        action.action === "disable"
          ? t("settings.console.confirm.statusDisable")
          : action.action === "lock"
            ? t("settings.console.confirm.statusLock")
            : t("settings.console.confirm.statusDefault"),
      confirmLabel: label,
      tone: isDestructive ? ("danger" as const) : ("warning" as const),
    };
  }

  if (action.type === "reset-password") {
    return {
      title: t("settings.console.confirm.resetPasswordTitle", {
        username: action.user.username,
      }),
      description: t("settings.console.confirm.resetPassword"),
      confirmLabel: t("settings.actions.resetPassword"),
      tone: "danger" as const,
    };
  }

  if (action.type === "revoke-session") {
    return {
      title: t("settings.console.confirm.revokeSessionTitle"),
      description: t("settings.console.confirm.revokeSession", {
        ip: action.session.ip_address || t("settings.console.common.unknownIp"),
      }),
      confirmLabel: t("settings.console.confirm.revokeSessionConfirmLabel"),
      tone: "warning" as const,
    };
  }

  if (action.type === "revoke-own-session") {
    return {
      title: t("settings.console.personal.revokeOwnSessionTitle"),
      description: t("settings.console.personal.revokeOwnSessionDescription", {
        ip: action.session.ip_address || t("settings.console.common.unknownIp"),
      }),
      confirmLabel: t("settings.console.confirm.revokeSessionConfirmLabel"),
      tone: "warning" as const,
    };
  }

  return {
    title: t("settings.console.audit.exportTitle"),
    description: t("settings.console.audit.exportDescription"),
    confirmLabel: t("settings.console.audit.exportCsv"),
    tone: "warning" as const,
  };
});

// `showToast` and `clearToast` are provided by useSettingsToasts() above.

// `formatDateTime` is provided by useBangkokFormatter() above.

interface UserStatusBadge {
  label: string;
  badgeClass: string;
}

function statusBadges(user: AdminUser): UserStatusBadge[] {
  const badges: UserStatusBadge[] = [];

  if (!user.is_active) {
    badges.push({
      label: t("settings.status.disabled"),
      badgeClass: "badge-neutral",
    });
  }

  if (user.is_locked) {
    badges.push({
      label: t("settings.status.locked"),
      badgeClass: "badge-warning",
    });
  }

  if (badges.length === 0) {
    badges.push({
      label: t("settings.status.active"),
      badgeClass: "badge-success",
    });
  }

  return badges;
}

// Kept for prop-drilling shape compatibility — first badge is the dominant one.
// statusBadges() never returns an empty array (it always pushes an "active"
// fallback), but the type system can't see that — the empty-string fallback
// keeps the signature honest without needing a non-null assertion.
function statusLabel(user: AdminUser) {
  return statusBadges(user)[0]?.label ?? "";
}

function statusClass(user: AdminUser) {
  return statusBadges(user)[0]?.badgeClass ?? "";
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

async function loadPersonalAccount() {
  personalLoading.value = true;
  personalError.value = null;
  personalMfaError.value = null;
  personalSessionsError.value = null;

  const [accountResult, mfaResult, sessionsResult] = await Promise.allSettled([
    accountApi.me(),
    accountApi.mfaStatus(),
    accountApi.listSessions(),
  ]);

  if (accountResult.status === "fulfilled") {
    personalAccount.value = accountResult.value;
  } else {
    personalAccount.value = null;
    personalError.value = getErrorMessage(
      accountResult.reason,
      t("settings.console.errors.loadPersonalAccount"),
    );
  }

  if (mfaResult.status === "fulfilled") {
    personalMfaStatus.value = mfaResult.value;
  } else {
    personalMfaStatus.value = null;
    personalMfaError.value = getErrorMessage(
      mfaResult.reason,
      t("settings.console.errors.loadMfaStatus"),
    );
  }

  if (sessionsResult.status === "fulfilled") {
    personalSessions.value = sessionsResult.value;
  } else {
    personalSessions.value = [];
    personalSessionsError.value = getErrorMessage(
      sessionsResult.reason,
      t("settings.console.errors.loadPersonalSessions"),
    );
  }

  personalLoading.value = false;
}

async function handlePersonalPasswordChange(
  payload: ChangePersonalPasswordInput,
) {
  personalPasswordLoading.value = true;
  personalPasswordError.value = null;
  clearToast();

  try {
    await accountApi.changePassword(payload);
    personalPasswordSuccessNonce.value += 1;
    showToast(t("settings.console.notices.personalPasswordChanged"), "success");
  } catch (error) {
    personalPasswordError.value = getErrorMessage(
      error,
      t("settings.console.errors.changePersonalPassword"),
    );
    showToast(personalPasswordError.value, "danger");
  } finally {
    personalPasswordLoading.value = false;
  }
}

// `loadUserMetrics` is provided by useSettingsUserMetrics() above.

async function loadSessions(userId: string) {
  if (!canUpdateUsers.value) {
    sessions.value = [];
    sessionsError.value = null;
    return;
  }

  sessionsLoading.value = true;
  sessionsError.value = null;

  try {
    sessions.value = await adminApi.listUserSessions(userId);
  } catch (error) {
    sessions.value = [];
    sessionsError.value = getErrorMessage(
      error,
      t("settings.console.errors.loadUserSessions"),
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
    usersState.value = await adminApi.listUsers(
      buildUserRequest(userQuery.offset),
    );

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
    auditState.value = await auditApi.listEvents(
      buildAuditRequest(auditQuery.offset),
    );
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

const createFieldErrors = ref<ServerFieldError[]>([]);

async function handleCreateUser(payload: CreateAdminUserInput) {
  if (!canCreateUsers.value) {
    return;
  }

  createLoading.value = true;
  createError.value = null;
  createFieldErrors.value = [];
  clearToast();

  try {
    const created = await adminApi.createUser(payload);
    createSuccessNonce.value += 1;
    showToast(t("settings.notices.userCreated"), "success");

    if (canViewUsers.value) {
      await loadUsers(0);
      await loadUserMetrics();

      // The OpenAPI schema marks `id` as optional, but the server always
      // populates it on a successful POST. Narrow it before propagating.
      const createdId = created.id;
      if (createdId && usersState.value.users.some((user) => user.id === createdId)) {
        selectedUserId.value = createdId;
        await loadSessions(createdId);
      }
    }

    if (canViewAudit.value) {
      await loadAudit(0);
    }
  } catch (error) {
    createError.value = getErrorMessage(error, t("settings.errors.createUser"));
    createFieldErrors.value = getServerFieldErrors(error);
    showToast(createError.value, "danger");
  } finally {
    createLoading.value = false;
  }
}

function requestStatusChange(action: AdminUserStatusAction) {
  if (!selectedUser.value) return;
  openStatusChange(action, selectedUser.value);
}

function requestStatusChangeForUser(user: AdminUser, action: AdminUserStatusAction) {
  selectedUserId.value = user.id;
  openStatusChange(action, user);
}

function requestPasswordReset(newPassword: string) {
  if (!selectedUser.value) return;
  openPasswordReset(selectedUser.value, newPassword);
}

function requestSessionRevoke(session: AdminSession) {
  if (!selectedUser.value) return;
  openSessionRevoke(selectedUser.value, session);
}

function requestOwnSessionRevoke(session: PersonalAccountSession) {
  openOwnSessionRevoke(session);
}

function requestAuditExport() {
  openAuditExport();
}

async function changeUserStatus(
  user: AdminUser,
  action: AdminUserStatusAction,
) {
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
    resetError.value = getErrorMessage(
      error,
      t("settings.errors.resetPassword"),
    );
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

function openMfaEnrollDialog() {
  mfaEnrollment.value = null;
  mfaEnrollError.value = null;
  mfaVerifyError.value = null;
  mfaEnrollOpen.value = true;
}

function closeMfaEnrollDialog() {
  if (mfaEnrollLoading.value || mfaVerifyLoading.value) return;
  mfaEnrollOpen.value = false;
  mfaEnrollment.value = null;
  mfaEnrollError.value = null;
  mfaVerifyError.value = null;
}

async function startMfaEnrollment() {
  mfaEnrollLoading.value = true;
  mfaEnrollError.value = null;

  try {
    mfaEnrollment.value = await mfaApi.enroll();
  } catch (error) {
    mfaEnrollError.value = getErrorMessage(
      error,
      t("settings.console.errors.mfaEnroll"),
    );
  } finally {
    mfaEnrollLoading.value = false;
  }
}

async function verifyMfaEnrollment(payload: PersonalMfaTotpInput) {
  mfaVerifyLoading.value = true;
  mfaVerifyError.value = null;

  try {
    await mfaApi.verify(payload);
    showToast(t("settings.console.notices.mfaEnabled"), "success");
    mfaEnrollment.value = null;
    mfaEnrollOpen.value = false;
    await loadPersonalAccount();
  } catch (error) {
    mfaVerifyError.value = getErrorMessage(
      error,
      t("settings.console.errors.mfaVerify"),
    );
  } finally {
    mfaVerifyLoading.value = false;
  }
}

function openMfaDisableDialog() {
  mfaDisableError.value = null;
  mfaDisableOpen.value = true;
}

function closeMfaDisableDialog() {
  if (mfaDisableLoading.value) return;
  mfaDisableOpen.value = false;
  mfaDisableError.value = null;
}

async function confirmMfaDisable(payload: PersonalMfaTotpInput) {
  mfaDisableLoading.value = true;
  mfaDisableError.value = null;

  try {
    await mfaApi.disable(payload);
    showToast(t("settings.console.notices.mfaDisabled"), "success");
    mfaDisableOpen.value = false;
    await loadPersonalAccount();
  } catch (error) {
    mfaDisableError.value = getErrorMessage(
      error,
      t("settings.console.errors.mfaDisable"),
    );
  } finally {
    mfaDisableLoading.value = false;
  }
}

async function revokeOwnSession(session: PersonalAccountSession) {
  revokingOwnSessionId.value = session.id;
  clearToast();

  try {
    await accountApi.revokeSession(session.id);
    showToast(t("settings.notices.sessionRevoked"), "success");
    personalSessions.value = await accountApi.listSessions();
  } catch (error) {
    showToast(
      getErrorMessage(error, t("settings.console.errors.revokeOwnSession")),
      "danger",
    );
  } finally {
    revokingOwnSessionId.value = null;
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
    showToast(t("settings.console.notices.auditExportStarted"), "success");

    if (canViewAudit.value) {
      await loadAudit(auditQuery.offset);
    }
  } catch (error) {
    showToast(
      getErrorMessage(error, t("settings.console.errors.exportAudit")),
      "danger",
    );
  } finally {
    exportLoading.value = false;
  }
}

async function confirmCurrentAction() {
  await runConfirm({
    onStatus: changeUserStatus,
    onResetPassword: resetPassword,
    onRevokeSession: revokeSession,
    onRevokeOwnSession: revokeOwnSession,
    onExportAudit: exportAudit,
  });
}
// `cancelConfirm` is provided by useSettingsConfirm() above.

function applyUserFilters(nextFilters: UserDirectoryFilters) {
  userFilters.search = nextFilters.search;
  userFilters.active = nextFilters.active;
  userFilters.locked = nextFilters.locked;
  void loadUsers(0);
}

function applyUserPageSize(next: number) {
  userQuery.limit = clampPageSize(next);
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

function applyAuditPageSize(next: number) {
  auditQuery.limit = clampPageSize(next);
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

  if (isIndividualSettings.value) {
    tasks.push(loadPersonalAccount());
  }

  if (isAdministrationSettings.value && canViewUsers.value) {
    tasks.push(loadUsers(0));
    tasks.push(loadUserMetrics());
  }

  if (isAdministrationSettings.value && canViewAudit.value) {
    tasks.push(loadAudit(0));
  }

  await Promise.all(tasks);
});
</script>

<template>
  <div class="settings-control-center">
    <AppPageHeader
      :title="settingsTitle"
      :description="settingsDescription"
    >
      <template #eyebrow>
        <span class="settings-eyebrow">{{ settingsEyebrow }}</span>
      </template>
      <template #actions>
        <AppButton
          v-if="isAdministrationSettings && canViewUsers"
          variant="secondary"
          size="sm"
          :loading="usersLoading || userMetricsLoading"
          @click="refreshUsers"
        >
          <AppIcon name="refresh" size="xs" />
          <span>{{ t("settings.refreshUsers") }}</span>
        </AppButton>
        <AppButton
          v-if="isAdministrationSettings && canViewAudit"
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

    <div v-if="toasts.length" class="settings-toast-stack" aria-label="Notifications">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="settings-toast"
        :class="`settings-toast--${toast.tone}`"
        :role="toast.tone === 'danger' ? 'alert' : 'status'"
        :aria-live="toast.tone === 'danger' ? 'assertive' : 'polite'"
      >
        <span>{{ toast.message }}</span>
        <button
          class="settings-toast__close"
          type="button"
          :aria-label="t('settings.console.common.dismissNotification')"
          @click="dismissToast(toast.id)"
        >
          <AppIcon name="close" size="xs" />
        </button>
      </div>
    </div>

    <div class="settings-console-shell">
      <SettingsSectionNav
        :groups="settingsNavGroups"
        :active-section="activeSection"
        :account-name="settingsAccountName"
        :account-subtitle="settingsAccountSubtitle"
        :account-initials="settingsAccountInitials"
        :show-identity="isIndividualSettings"
        @select="activeSection = $event"
      />

      <div class="settings-console-main">
        <SettingsOverviewPanel
          v-show="isAdministrationSettings && activeSection === 'overview'"
          :metrics="kpiMetrics"
          :loading="metricLoading"
          :signals="overviewSignals"
          :capabilities="apiCapabilities"
          :app-name="appName"
          :api-base-url="apiBaseUrl"
          :timezone="clientTimezone"
        />

        <SettingsPersonalAccountPanel
          v-show="isIndividualSettings && activeSection === 'personal-account'"
          :account="personalAccount"
          :mfa-status="personalMfaStatus"
          :sessions="personalSessions"
          :loading="personalLoading"
          :error="personalError"
          :mfa-error="personalMfaError"
          :sessions-error="personalSessionsError"
          :password-loading="personalPasswordLoading"
          :password-error="personalPasswordError"
          :password-success-nonce="personalPasswordSuccessNonce"
          :revoking-session-id="revokingOwnSessionId"
          :format-date-time="formatDateTime"
          @refresh="loadPersonalAccount"
          @change-password="handlePersonalPasswordChange"
          @revoke-session="requestOwnSessionRevoke"
          @start-mfa-enroll="openMfaEnrollDialog"
          @start-mfa-disable="openMfaDisableDialog"
        />

        <section
          v-show="isAdministrationSettings && activeSection === 'users'"
          class="settings-workspace"
        >
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
            :status-badges="statusBadges"
            :can-deactivate-users="canDeactivateUsers"
            :can-update-users="canUpdateUsers"
            :status-action="statusAction"
            @apply="applyUserFilters"
            @page="loadUsers"
            @page-size="applyUserPageSize"
            @select="selectUser"
            @status="requestStatusChangeForUser"
          />

          <div class="settings-side-stack">
            <SettingsCreateUserForm
              v-if="canCreateUsers"
              :loading="createLoading"
              :error="createError"
              :server-field-errors="createFieldErrors"
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
              :status-badges="statusBadges"
              @status="requestStatusChange"
              @reset-password="requestPasswordReset"
              @revoke-session="requestSessionRevoke"
            />
          </div>
        </section>

        <SettingsGroupsRolesPanel
          v-show="isAdministrationSettings && activeSection === 'groups'"
          :groups="derivedGroups"
          :directory-complete="directoryCompleteForGroups"
          :loading="usersLoading"
        />

        <SettingsFunctionPermissionsPanel
          v-show="isAdministrationSettings && activeSection === 'function-permissions'"
          :actions="permissionActions"
          :rows="functionPermissionRows"
          :session-permission-count="currentSessionFunctionCount"
        />

        <SettingsDataPermissionsPanel
          v-show="isAdministrationSettings && activeSection === 'data-permissions'"
          :grants="dataPermissionGrants"
          :format-date-time="formatDateTime"
        />

        <SettingsSecurityPolicyPanel
          v-show="isAdministrationSettings && activeSection === 'security-policy'"
          :policy="securityPolicy"
        />

        <SettingsNotificationsPanel
          v-show="isAdministrationSettings && activeSection === 'notifications'"
          :preferences="notificationPreferences"
        />

        <SettingsAuditLog
          v-if="isAdministrationSettings && canViewAudit"
          v-show="activeSection === 'audit'"
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
          @page-size="applyAuditPageSize"
          @refresh="loadAudit(auditQuery.offset)"
          @export="requestAuditExport"
        />
      </div>
    </div>

    <SettingsConfirmDialog
      :open="Boolean(confirmAction)"
      :title="confirmDialog.title"
      :description="confirmDialog.description"
      :confirm-label="confirmDialog.confirmLabel"
      :cancel-label="t('settings.console.common.cancel')"
      :tone="confirmDialog.tone"
      :loading="confirmLoading"
      @cancel="cancelConfirm"
      @confirm="confirmCurrentAction"
    />

    <SettingsMfaEnrollDialog
      :open="mfaEnrollOpen"
      :enrollment="mfaEnrollment"
      :enrolling="mfaEnrollLoading"
      :verifying="mfaVerifyLoading"
      :enroll-error="mfaEnrollError"
      :verify-error="mfaVerifyError"
      @cancel="closeMfaEnrollDialog"
      @start-enroll="startMfaEnrollment"
      @verify="verifyMfaEnrollment"
    />

    <SettingsMfaDisableDialog
      :open="mfaDisableOpen"
      :loading="mfaDisableLoading"
      :error="mfaDisableError"
      @cancel="closeMfaDisableDialog"
      @confirm="confirmMfaDisable"
    />
  </div>
</template>

<style scoped>
.settings-control-center {
  width: 100%;
  max-width: 1280px;
  margin: 0 auto;
  display: grid;
  gap: var(--space-4);
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

.settings-console-shell {
  display: grid;
  grid-template-columns: minmax(15rem, 17.5rem) minmax(0, 1fr);
  gap: var(--space-6);
  align-items: start;
}

.settings-console-main {
  display: grid;
  gap: var(--space-5);
  min-width: 0;
}

.settings-control-center :deep(.settings-panel) {
  overflow: hidden;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  box-shadow: none;
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
  min-width: 920px;
}

.settings-control-center :deep(.table th),
.settings-control-center :deep(.table td) {
  vertical-align: top;
}

.settings-control-center :deep(.table thead th) {
  position: sticky;
  top: 0;
  z-index: 1;
}

@media (max-width: 1180px) {
  .settings-workspace {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 1024px) {
  .settings-console-shell {
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

.settings-toast-stack {
  display: grid;
  gap: var(--space-3);
}
</style>
