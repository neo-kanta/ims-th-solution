<script setup lang="ts">
import { computed, ref, watch } from "vue";

import type {
  AdminSession,
  AdminUser,
  AdminUserStatusAction,
} from "../admin.types";

const props = defineProps<{
  user: AdminUser | null;
  sessions: AdminSession[];
  sessionsLoading: boolean;
  sessionsError: string | null;
  canDeactivateUsers: boolean;
  canUpdateUsers: boolean;
  statusAction: AdminUserStatusAction | null;
  resetLoading: boolean;
  revokingSessionId: string | null;
  resetError: string | null;
  resetSuccessNonce: number;
  formatDateTime: (value?: string | null) => string;
  statusLabel: (user: AdminUser) => string;
  statusClass: (user: AdminUser) => string;
}>();

const emit = defineEmits<{
  status: [action: AdminUserStatusAction];
  resetPassword: [newPassword: string];
  revokeSession: [session: AdminSession];
}>();

const { t } = useI18n();

const newPassword = ref("");
const resetFieldError = ref<string | null>(null);
const showPassword = ref(false);

const passwordToggleLabel = computed(() =>
  showPassword.value
    ? t("auth.hidePassword", "Hide password")
    : t("auth.showPassword", "Show password"),
);

const MIN_PASSWORD_LENGTH = 12;

const canSubmitReset = computed(
  () =>
    props.canUpdateUsers &&
    !props.resetLoading &&
    newPassword.value.length >= MIN_PASSWORD_LENGTH,
);

watch(
  () => props.user?.id,
  () => {
    newPassword.value = "";
    resetFieldError.value = null;
    showPassword.value = false;
  },
);

watch(
  () => props.resetSuccessNonce,
  () => {
    newPassword.value = "";
    resetFieldError.value = null;
    showPassword.value = false;
  },
);

const TEN_YEARS_MS = 10 * 365 * 24 * 60 * 60 * 1000;

function lockedUntilLabel(user: AdminUser): string {
  if (!user.locked_until) {
    return t("common.notAvailable");
  }
  const date = new Date(user.locked_until);
  if (Number.isNaN(date.getTime())) {
    return user.locked_until;
  }
  if (date.getTime() > Date.now() + TEN_YEARS_MS) {
    return t("settings.status.lockedIndefinitely");
  }
  return props.formatDateTime(user.locked_until);
}

function requestResetPassword() {
  if (!props.user || props.resetLoading) {
    return;
  }

  if (newPassword.value.length < MIN_PASSWORD_LENGTH) {
    resetFieldError.value = t("settings.console.otherAccounts.validationPassword");
    return;
  }

  resetFieldError.value = null;
  emit("resetPassword", newPassword.value);
}
</script>

<template>
  <section class="settings-panel settings-user-detail">
    <template v-if="user">
      <header class="settings-panel__header settings-user-detail__header">
        <div>
          <h2 class="settings-panel__title">{{ user.display_name }}</h2>
          <p class="settings-panel__subtitle">{{ user.username }}</p>
        </div>
        <span class="badge" :class="statusClass(user)">
          {{ statusLabel(user) }}
        </span>
      </header>

      <div class="settings-user-detail__body">
        <dl class="settings-detail-grid">
          <div>
            <dt>{{ t("settings.details.email") }}</dt>
            <dd>{{ user.email }}</dd>
          </div>
          <div>
            <dt>{{ t("settings.details.groups") }}</dt>
            <dd>
              {{
                user.groups.length
                  ? user.groups.join(", ")
                  : t("common.notAvailable")
              }}
            </dd>
          </div>
          <div>
            <dt>{{ t("settings.details.failedLogins") }}</dt>
            <dd>{{ user.failed_login_attempts }}</dd>
          </div>
          <div>
            <dt>{{ t("settings.details.lockedUntil") }}</dt>
            <dd>{{ lockedUntilLabel(user) }}</dd>
          </div>
          <div>
            <dt>{{ t("settings.console.accountDetail.passwordChange") }}</dt>
            <dd>
              {{
                user.force_password_change
                  ? t("settings.console.accountDetail.required")
                  : t("settings.console.accountDetail.current")
              }}
            </dd>
          </div>
          <div>
            <dt>{{ t("settings.console.accountDetail.updated") }}</dt>
            <dd>{{ formatDateTime(user.updated_at) }}</dd>
          </div>
        </dl>

        <div
          v-if="canDeactivateUsers || canUpdateUsers"
          class="settings-action-block"
        >
          <div class="settings-action-block__title">
            {{ t("settings.console.accountDetail.accountControls") }}
          </div>
          <div class="settings-action-grid">
            <AppButton
              v-if="canDeactivateUsers"
              variant="danger"
              size="sm"
              :loading="statusAction === 'disable'"
              :disabled="!user.is_active || Boolean(statusAction)"
              @click="emit('status', 'disable')"
            >
              {{ t("settings.actions.disable") }}
            </AppButton>
            <AppButton
              v-if="canDeactivateUsers"
              variant="secondary"
              size="sm"
              :loading="statusAction === 'enable'"
              :disabled="user.is_active || Boolean(statusAction)"
              @click="emit('status', 'enable')"
            >
              {{ t("settings.actions.enable") }}
            </AppButton>
            <AppButton
              v-if="canUpdateUsers"
              variant="secondary"
              size="sm"
              :loading="statusAction === 'lock'"
              :disabled="user.is_locked || Boolean(statusAction)"
              @click="emit('status', 'lock')"
            >
              {{ t("settings.actions.lock") }}
            </AppButton>
            <AppButton
              v-if="canUpdateUsers"
              variant="secondary"
              size="sm"
              :loading="statusAction === 'unlock'"
              :disabled="!user.is_locked || Boolean(statusAction)"
              @click="emit('status', 'unlock')"
            >
              {{ t("settings.actions.unlock") }}
            </AppButton>
          </div>
        </div>

        <form
          v-if="canUpdateUsers"
          class="settings-action-block"
          :aria-busy="resetLoading"
          @submit.prevent="requestResetPassword"
        >
          <div class="settings-action-block__title">
            {{ t("settings.actions.resetPassword") }}
          </div>
          <p class="settings-action-block__hint">
            {{ t("settings.console.accountDetail.resetPasswordHint") }}
          </p>
          <div
            class="form-group"
            :class="{ 'field-error': resetFieldError || resetError }"
          >
            <label for="settings-reset-password" class="label label-required">
              {{ t("settings.newTemporaryPasswordPlaceholder") }}
            </label>
            <div class="settings-password-field">
              <input
                id="settings-reset-password"
                v-model="newPassword"
                class="input"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="new-password"
                :aria-invalid="Boolean(resetFieldError || resetError)"
              />
              <button
                type="button"
                class="settings-password-field__toggle"
                :aria-label="passwordToggleLabel"
                :title="passwordToggleLabel"
                :aria-pressed="showPassword"
                @click="showPassword = !showPassword"
              >
                <svg
                  v-if="showPassword"
                  xmlns="http://www.w3.org/2000/svg"
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  aria-hidden="true"
                >
                  <path
                    d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"
                  />
                  <line x1="1" y1="1" x2="23" y2="23" />
                </svg>
                <svg
                  v-else
                  xmlns="http://www.w3.org/2000/svg"
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  aria-hidden="true"
                >
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                  <circle cx="12" cy="12" r="3" />
                </svg>
              </button>
            </div>
            <div v-if="resetFieldError" class="error-text">
              {{ resetFieldError }}
            </div>
            <div v-if="resetError" class="error-text">
              {{ resetError }}
            </div>
          </div>
          <AppButton
            type="submit"
            variant="primary"
            size="sm"
            :loading="resetLoading"
            :disabled="!canSubmitReset"
          >
            {{ t("settings.actions.resetPassword") }}
          </AppButton>
        </form>

        <div v-if="canUpdateUsers" class="settings-action-block">
          <div class="settings-action-block__header">
            <div class="settings-action-block__title">
              {{ t("settings.sessionsTitle") }}
            </div>
            <span class="badge badge-neutral">{{ sessions.length }}</span>
          </div>

          <div v-if="sessionsError" class="alert alert-danger" role="alert">
            {{ sessionsError }}
          </div>
          <div v-else-if="sessionsLoading" class="settings-card-state">
            {{ t("settings.loadingSessions") }}
          </div>
          <div v-else-if="sessions.length === 0" class="settings-card-state">
            {{ t("settings.noSessions") }}
          </div>
          <div v-else class="settings-session-list">
            <article
              v-for="session in sessions"
              :key="session.id"
              class="settings-session"
            >
              <div class="settings-session__copy">
                <div class="settings-record-primary">
                  {{ session.ip_address || t("settings.unknownIp") }}
                </div>
                <div class="settings-record-secondary">
                  {{ session.user_agent || t("common.notAvailable") }}
                </div>
                <div class="settings-record-secondary">
                  {{ t("settings.lastActivity", { date: formatDateTime(session.last_activity_at) }) }}
                </div>
                <div class="settings-record-secondary">
                  {{ t("settings.console.accountDetail.expires", { date: formatDateTime(session.expires_at) }) }}
                </div>
              </div>
              <AppButton
                variant="secondary"
                size="sm"
                :loading="revokingSessionId === session.id"
                :disabled="Boolean(revokingSessionId)"
                @click="emit('revokeSession', session)"
              >
                {{ t("settings.actions.revoke") }}
              </AppButton>
            </article>
          </div>
        </div>
      </div>
    </template>

    <div v-else class="settings-user-detail__empty">
      <AppIcon name="accounts" size="lg" />
      <div>
        <h2 class="settings-panel__title">
          {{ t("settings.console.accountDetail.noAccountSelected") }}
        </h2>
        <p class="settings-panel__subtitle">
          {{ t("settings.console.accountDetail.noAccountSelectedSubtitle") }}
        </p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.settings-user-detail {
  min-width: 0;
}

.settings-user-detail__header {
  align-items: flex-start;
}

.settings-user-detail__body {
  display: grid;
  gap: var(--space-5);
  padding: var(--space-5);
}

.settings-user-detail__empty {
  display: grid;
  justify-items: center;
  gap: var(--space-4);
  padding: var(--space-8) var(--space-5);
  text-align: center;
  color: var(--text-tertiary);
}

.settings-detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-5);
  margin: 0;
}

.settings-detail-grid dt {
  color: var(--text-tertiary);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.settings-detail-grid dd {
  margin: var(--space-1) 0 0;
  color: var(--text-primary);
  overflow-wrap: anywhere;
}

.settings-action-block {
  display: grid;
  gap: var(--space-4);
  padding-top: var(--space-5);
  border-top: 1px solid var(--border-subtle);
}

.settings-action-block .form-group {
  margin-bottom: 0;
}

.settings-action-block__header {
  display: flex;
  justify-content: space-between;
  gap: var(--space-4);
  align-items: center;
}

.settings-action-block__title {
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
}

.settings-action-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3);
}

.settings-password-field {
  position: relative;
}

.settings-password-field .input {
  padding-right: var(--space-10);
}

.settings-password-field__toggle {
  position: absolute;
  top: 50%;
  right: var(--space-2);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border-radius: var(--radius-sm);
  color: var(--text-tertiary);
  transform: translateY(-50%);
}

.settings-password-field__toggle:hover {
  color: var(--text-primary);
  background: var(--action-ghost-hover);
}

.settings-session-list {
  display: grid;
  gap: var(--space-3);
}

.settings-session {
  display: flex;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
}

.settings-session__copy {
  min-width: 0;
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

.settings-card-state {
  padding: var(--space-5);
  border: 1px dashed var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  text-align: center;
}

@media (max-width: 640px) {
  .settings-detail-grid,
  .settings-action-grid {
    grid-template-columns: 1fr;
  }

  .settings-session {
    flex-direction: column;
  }
}

.settings-action-block__hint {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-relaxed);
}
</style>
