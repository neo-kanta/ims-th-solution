<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";

import type {
  ChangePersonalPasswordInput,
  PersonalAccountPayload,
  PersonalAccountSession,
  PersonalMfaEnrollResult,
  PersonalMfaStatus,
  PersonalMfaTotpInput,
} from "../account.types";
import type {
  PersonalActivityEvent,
  PersonalAppearancePreferences,
  PersonalLeaveRequest,
  PersonalNotificationPreference,
  PersonalSettingsSectionId,
  PersonalWorkPreferences,
} from "../personal.types";
import { accountApi } from "../services/accountApi";
import { mfaApi } from "../services/mfaApi";
import { personalSettingsApi } from "../services/personalSettingsApi";
import PersonalActivityPanel from "./PersonalActivityPanel.vue";
import PersonalAppearancePanel from "./PersonalAppearancePanel.vue";
import PersonalLeaveDelegationPanel from "./PersonalLeaveDelegationPanel.vue";
import PersonalNotificationsPanel from "./PersonalNotificationsPanel.vue";
import PersonalWorkPreferencesPanel from "./PersonalWorkPreferencesPanel.vue";
import SettingsConfirmDialog from "./SettingsConfirmDialog.vue";
import SettingsMfaDisableDialog from "./SettingsMfaDisableDialog.vue";
import SettingsMfaEnrollDialog from "./SettingsMfaEnrollDialog.vue";
import SettingsPersonalAccountPanel from "./SettingsPersonalAccountPanel.vue";
import SettingsPersonalSessionsPanel from "./SettingsPersonalSessionsPanel.vue";
import PersonalPublicProfilePanel from "./PersonalPublicProfilePanel.vue";
import PersonalPasswordAuthPanel from "./PersonalPasswordAuthPanel.vue";
import PersonalSessionsMfaPanel from "./PersonalSessionsMfaPanel.vue";
import PersonalLocaleTimezonePanel from "./PersonalLocaleTimezonePanel.vue";
import PersonalWorkstationPanel from "./PersonalWorkstationPanel.vue";
import PersonalCloseAccountPanel from "./PersonalCloseAccountPanel.vue";
import { usePersonalSettings } from "../composables/usePersonalSettings";
import { useBangkokFormatter } from "~/shared/composables/useBangkokFormatter";

interface PersonalNavItem {
  id: PersonalSettingsSectionId;
  label: string;
  description: string;
  icon: string;
}

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const { locale, setLocale } = useI18n();
const { formatDateTime } = useBangkokFormatter();
const {
  appearancePreferences,
  preferenceLoading,
  updateAppearancePreferences,
  loadAppearancePreferences,
} = usePersonalSettings();

interface PersonalNavGroup {
  id: string;
  label: string;
  items: PersonalNavItem[];
}

const navGroups: PersonalNavGroup[] = [
  {
    id: "account",
    label: "Account",
    items: [
      {
        id: "public-profile",
        label: "Public profile",
        description: "Manage your public profile",
        icon: "user",
      },
      {
        id: "account",
        label: "Account",
        description: "View account status and permissions",
        icon: "shield",
      },
      {
        id: "appearance",
        label: "Appearance",
        description: "Theme, density, and accessibility",
        icon: "dashboard",
      },
      {
        id: "notifications",
        label: "Notifications",
        description: "In-app and email channels",
        icon: "notifications",
      },
    ],
  },
  {
    id: "access",
    label: "Access",
    items: [
      {
        id: "password-auth",
        label: "Password & auth",
        description: "Update your password",
        icon: "lock",
      },
      {
        id: "sessions-mfa",
        label: "Sessions & MFA",
        description: "Manage 2FA and sessions",
        icon: "monitor",
      },
      {
        id: "delegation-leave",
        label: "Delegation & leave",
        description: "Manage out of office requests",
        icon: "clock",
      },
    ],
  },
  {
    id: "operational",
    label: "Operational",
    items: [
      {
        id: "locale-timezone",
        label: "Locale & timezone",
        description: "Language and time settings",
        icon: "globe",
      },
      {
        id: "workstation",
        label: "Workstation",
        description: "Table and dashboard defaults",
        icon: "list",
      },
      {
        id: "activity-log",
        label: "Activity log",
        description: "Security and approval history",
        icon: "activity",
      },
    ],
  },
  {
    id: "danger",
    label: "Danger zone",
    items: [
      {
        id: "close-account",
        label: "Close account",
        description: "Deactivate or close your account",
        icon: "trash",
      },
    ],
  },
];

const navItems = navGroups.flatMap((g) => g.items);

function normalizeSection(value: unknown): PersonalSettingsSectionId {
  return navItems.some((item) => item.id === value)
    ? (value as PersonalSettingsSectionId)
    : "account";
}

const activeSection = ref<PersonalSettingsSectionId>(
  normalizeSection(route.query.section),
);

watch(
  () => route.query.section,
  (section) => {
    activeSection.value = normalizeSection(section);
  },
);

watch(activeSection, (section) => {
  if (route.query.section === section) return;
  void router.replace({ query: { ...route.query, section } });
});

const personalAccount = ref<PersonalAccountPayload | null>(null);
const personalMfaStatus = ref<PersonalMfaStatus | null>(null);
const personalSessions = ref<PersonalAccountSession[]>([]);
const notificationPreferences = ref<PersonalNotificationPreference[]>([]);
const workPreferences = ref<PersonalWorkPreferences | null>(null);
const leaveRequests = ref<PersonalLeaveRequest[]>([]);
const activityEvents = ref<PersonalActivityEvent[]>([]);

const personalLoading = ref(false);
const personalPasswordLoading = ref(false);
const revokingOwnSessionId = ref<string | null>(null);
const leaveLoading = ref(false);
const activityLoading = ref(false);
const confirmSession = ref<PersonalAccountSession | null>(null);
const confirmLoading = ref(false);

const personalError = ref<string | null>(null);
const personalMfaError = ref<string | null>(null);
const personalSessionsError = ref<string | null>(null);
const personalPasswordError = ref<string | null>(null);
const preferenceError = ref<string | null>(null);
const leaveError = ref<string | null>(null);
const activityError = ref<string | null>(null);

const personalPasswordSuccessNonce = ref(0);
const notice = ref<{ tone: "success" | "danger"; message: string } | null>(
  null,
);
const showToast = ref(false);
watch(notice, (newNotice) => {
  if (newNotice) {
    showToast.value = true;
  }
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

const accountName = computed(
  () =>
    personalAccount.value?.user.display_name ||
    authStore.user?.displayName ||
    "Personal account",
);
const accountSubtitle = computed(
  () =>
    personalAccount.value?.user.username ||
    authStore.user?.username ||
    "Signed-in user",
);
const accountInitials = computed(() => {
  const source = accountName.value || accountSubtitle.value;
  const parts = source.split(/\s+/).filter(Boolean).slice(0, 2);
  return parts.length
    ? parts.map((part) => part[0]?.toUpperCase() || "").join("")
    : "IM";
});

function getErrorMessage(error: unknown, fallback: string): string {
  if (typeof error === "object" && error !== null && "message" in error) {
    const message = (error as { message?: unknown }).message;
    if (typeof message === "string" && message.trim()) return message;
  }

  return fallback;
}

function showNotice(message: string, tone: "success" | "danger" = "success") {
  notice.value = { message, tone };
}

async function loadPersonalAccount() {
  personalLoading.value = true;
  personalError.value = null;
  personalMfaError.value = null;
  personalSessionsError.value = null;

  try {
    const [account, mfaStatus, sessions] = await Promise.all([
      accountApi.me(),
      accountApi.mfaStatus(),
      accountApi.listSessions(),
    ]);
    personalAccount.value = account;
    personalMfaStatus.value = mfaStatus;
    personalSessions.value = sessions;
  } catch (error) {
    personalError.value = getErrorMessage(
      error,
      "Unable to load personal account.",
    );
  } finally {
    personalLoading.value = false;
  }
}

async function handlePersonalPasswordChange(
  payload: ChangePersonalPasswordInput,
) {
  personalPasswordLoading.value = true;
  personalPasswordError.value = null;

  try {
    await accountApi.changePassword(payload);
    personalPasswordSuccessNonce.value += 1;
    showNotice("Password changed successfully.");
    await loadPersonalAccount();
  } catch (error) {
    personalPasswordError.value = getErrorMessage(
      error,
      "Unable to change password.",
    );
  } finally {
    personalPasswordLoading.value = false;
  }
}

function requestOwnSessionRevoke(session: PersonalAccountSession) {
  confirmSession.value = session;
}

async function confirmOwnSessionRevoke() {
  if (!confirmSession.value) return;
  confirmLoading.value = true;
  revokingOwnSessionId.value = confirmSession.value.id;

  try {
    await accountApi.revokeSession(confirmSession.value.id);
    showNotice("Session revoked.");
    confirmSession.value = null;
    personalSessions.value = await accountApi.listSessions();
  } catch (error) {
    showNotice(getErrorMessage(error, "Unable to revoke session."), "danger");
  } finally {
    confirmLoading.value = false;
    revokingOwnSessionId.value = null;
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
}

async function startMfaEnrollment() {
  mfaEnrollLoading.value = true;
  mfaEnrollError.value = null;

  try {
    mfaEnrollment.value = await mfaApi.enroll();
  } catch (error) {
    mfaEnrollError.value = getErrorMessage(
      error,
      "Unable to start MFA enrollment.",
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
    showNotice("MFA enabled.");
    mfaEnrollment.value = null;
    mfaEnrollOpen.value = false;
    await loadPersonalAccount();
  } catch (error) {
    mfaVerifyError.value = getErrorMessage(error, "Unable to verify MFA code.");
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
}

async function confirmMfaDisable(payload: PersonalMfaTotpInput) {
  mfaDisableLoading.value = true;
  mfaDisableError.value = null;

  try {
    await mfaApi.disable(payload);
    showNotice("MFA disabled.");
    mfaDisableOpen.value = false;
    await loadPersonalAccount();
  } catch (error) {
    mfaDisableError.value = getErrorMessage(error, "Unable to disable MFA.");
  } finally {
    mfaDisableLoading.value = false;
  }
}

async function loadPreferences() {
  preferenceLoading.value = true;
  preferenceError.value = null;

  try {
    const [notifications, work] = await Promise.all([
      personalSettingsApi.listNotifications(),
      personalSettingsApi.getWorkPreferences(),
    ]);
    notificationPreferences.value = notifications;
    workPreferences.value = {
      ...work,
      language: (locale.value as "en" | "th" | "zh") || work.language,
    };
  } catch (error) {
    preferenceError.value = getErrorMessage(
      error,
      "Unable to load preferences.",
    );
  } finally {
    preferenceLoading.value = false;
  }
}

async function toggleNotification(
  id: string,
  channel: "inApp" | "email",
  value: boolean,
) {
  const next = notificationPreferences.value.map((preference) =>
    preference.id === id ? { ...preference, [channel]: value } : preference,
  );

  notificationPreferences.value =
    await personalSettingsApi.updateNotifications(next);
  showNotice("Notification preferences saved.");
}

async function updateAppearance(preferences: PersonalAppearancePreferences) {
  await updateAppearancePreferences(preferences);
  showNotice("Appearance preferences saved.");
}

async function saveWorkPreferences(preferences: PersonalWorkPreferences) {
  preferenceLoading.value = true;
  const globalProgress = useGlobalProgress();
  globalProgress.value = true;

  try {
    workPreferences.value =
      await personalSettingsApi.updateWorkPreferences(preferences);
    await setLocale(workPreferences.value.language);
    showNotice("Work preferences saved.");
  } finally {
    preferenceLoading.value = false;
    globalProgress.value = false;
  }
}

async function loadLeaveRequests() {
  leaveLoading.value = true;
  leaveError.value = null;

  try {
    leaveRequests.value = await personalSettingsApi.listLeaveRequests();
  } catch (error) {
    leaveError.value = getErrorMessage(error, "Unable to load leave requests.");
  } finally {
    leaveLoading.value = false;
  }
}

async function loadActivity() {
  activityLoading.value = true;
  activityError.value = null;

  try {
    activityEvents.value = await personalSettingsApi.listActivity();
  } catch (error) {
    activityError.value = getErrorMessage(
      error,
      "Unable to load personal activity.",
    );
  } finally {
    activityLoading.value = false;
  }
}

onMounted(async () => {
  await Promise.all([
    loadPersonalAccount(),
    loadPreferences(),
    loadAppearancePreferences(),
    loadLeaveRequests(),
    loadActivity(),
  ]);
});
</script>

<template>
  <div class="personal-settings-console">
    <AppToast
      v-if="notice"
      v-model="showToast"
      :message="notice.message"
      :tone="notice.tone"
      @dismiss="notice = null"
    />

    <header class="personal-settings-header" aria-labelledby="settings-header">
      <div class="personal-settings-header__profile">
        <span class="personal-settings-header__avatar" aria-hidden="true">
          {{ accountInitials }}
        </span>
        <div class="personal-settings-header__info">
          <h1 id="settings-header" class="personal-settings-header__title">
            {{ accountName }}
            <span class="personal-settings-header__username">({{ accountSubtitle }})</span>
          </h1>
          <div class="personal-settings-header__meta">
            <p class="personal-settings-header__subtitle">
              Your personal account
            </p>
          </div>
        </div>
      </div>
    </header>

    <div class="personal-settings-shell">
      <AppSettingsNav
        v-model="activeSection"
        :groups="navGroups"
      />

      <main class="personal-settings-main">
        <PersonalPublicProfilePanel
          v-if="activeSection === 'public-profile'"
          :account="personalAccount"
          :loading="personalLoading"
        />

        <SettingsPersonalAccountPanel
          v-if="activeSection === 'account'"
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

        <PersonalAppearancePanel
          v-if="activeSection === 'appearance'"
          :preferences="appearancePreferences"
          :loading="preferenceLoading"
          @update="updateAppearance"
        />

        <PersonalNotificationsPanel
          v-if="activeSection === 'notifications'"
          :preferences="notificationPreferences"
          :loading="preferenceLoading"
          :error="preferenceError"
          @toggle="toggleNotification"
          @refresh="loadPreferences"
        />

        <PersonalPasswordAuthPanel
          v-if="activeSection === 'password-auth'"
          :password-loading="personalPasswordLoading"
          :password-error="personalPasswordError"
          :password-success-nonce="personalPasswordSuccessNonce"
          @change-password="handlePersonalPasswordChange"
        />

        <PersonalSessionsMfaPanel
          v-if="activeSection === 'sessions-mfa'"
          :mfa-status="personalMfaStatus"
          :sessions="personalSessions"
          :loading="personalLoading"
          :mfa-error="personalMfaError"
          :sessions-error="personalSessionsError"
          :revoking-session-id="revokingOwnSessionId"
          :format-date-time="formatDateTime"
          @refresh="loadPersonalAccount"
          @revoke-session="requestOwnSessionRevoke"
          @start-mfa-enroll="openMfaEnrollDialog"
          @start-mfa-disable="openMfaDisableDialog"
        />

        <PersonalLocaleTimezonePanel
          v-if="activeSection === 'locale-timezone'"
          :preferences="workPreferences"
          :appearance-preferences="appearancePreferences"
          :loading="preferenceLoading"
          @save-work="saveWorkPreferences"
          @save-appearance="updateAppearance"
        />

        <PersonalWorkstationPanel
          v-if="activeSection === 'workstation'"
          :preferences="workPreferences"
          :appearance-preferences="appearancePreferences"
          :loading="preferenceLoading"
          @save-work="saveWorkPreferences"
          @save-appearance="updateAppearance"
        />

        <PersonalLeaveDelegationPanel
          v-if="activeSection === 'delegation-leave'"
          :requests="leaveRequests"
          :loading="leaveLoading"
          :error="leaveError"
        />

        <PersonalActivityPanel
          v-if="activeSection === 'activity-log'"
          :events="activityEvents"
          :loading="activityLoading"
          :error="activityError"
          :format-date-time="formatDateTime"
        />

        <PersonalCloseAccountPanel v-if="activeSection === 'close-account'" />
      </main>
    </div>

    <SettingsConfirmDialog
      :open="Boolean(confirmSession)"
      title="Revoke your active session"
      :description="`This ends the selected session from ${confirmSession?.ip_address || 'unknown IP'}. If this is your current browser session, you may need to sign in again.`"
      confirm-label="Revoke session"
      cancel-label="Cancel"
      tone="danger"
      :loading="confirmLoading"
      @cancel="confirmSession = null"
      @confirm="confirmOwnSessionRevoke"
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
.personal-settings-console {
  width: 100%;
  max-width: 1280px;
  margin: 0 auto;
  display: grid;
  gap: var(--space-4);
  padding-bottom: var(--space-8);
}

.personal-settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding-bottom: var(--space-5);
  border-bottom: 1px solid var(--border-subtle);
  margin-top: var(--space-1);
  margin-bottom: var(--space-2);
}

.personal-settings-header__profile {
  display: flex;
  align-items: center;
  gap: var(--space-4);
}

.personal-settings-header__avatar {
  width: 48px;
  height: 48px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border-radius: var(--radius-pill);
  border: 1px solid var(--border-subtle);
  background: var(--bg-selected);
  color: var(--action-primary);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-sm);
}

.personal-settings-header__info {
  display: flex;
  flex-direction: column;
}

.personal-settings-header__title {
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  display: flex;
  align-items: baseline;
}

.personal-settings-header__username {
  font-size: var(--font-size-sm);
  color: var(--text-tertiary);
  font-weight: var(--font-weight-normal);
  margin-left: var(--space-2);
}

.personal-settings-header__subtitle {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

@media (max-width: 768px) {
  .personal-settings-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-3);
  }
}

.personal-settings-eyebrow {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.personal-settings-shell {
  display: grid;
  grid-template-columns: minmax(15rem, 17.5rem) minmax(0, 1fr);
  gap: var(--space-6);
  align-items: start;
}

.personal-settings-main {
  min-width: 0;
  display: grid;
  gap: var(--space-5);
}

.personal-settings-console :deep(.settings-panel) {
  overflow: hidden;
  border: none;
  background: transparent;
  box-shadow: none;
  border-radius: 0;
  margin-bottom: var(--space-6);
}

.personal-settings-console :deep(.settings-panel__header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: 0 0 var(--space-4) 0;
  border-bottom: 1px solid var(--border-subtle);
}

.personal-settings-console :deep(.settings-panel__footer) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-5) 0;
  border-top: 1px solid var(--border-subtle);
  background: transparent;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.personal-settings-console :deep(.settings-panel__title) {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
}

.personal-settings-console :deep(.settings-panel__subtitle) {
  margin: var(--space-1) 0 0;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-relaxed);
}

.personal-settings-console :deep(.settings-panel__body) {
  padding-left: 0 !important;
  padding-right: 0 !important;
}

.personal-settings-console :deep(.settings-personal__body),
.personal-settings-console :deep(.settings-personal-password) {
  padding-left: 0 !important;
  padding-right: 0 !important;
}

.personal-settings-console :deep(.settings-panel__alert) {
  margin: var(--space-5) var(--space-5) 0;
}

@media (max-width: 1024px) {
  .personal-settings-shell {
    grid-template-columns: 1fr;
  }

  .personal-settings-console :deep(.settings-panel__header),
  .personal-settings-console :deep(.settings-panel__footer) {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
