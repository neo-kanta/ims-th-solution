<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";

import type {
  ChangePersonalPasswordInput,
  PersonalAccountPayload,
  PersonalAccountSession,
  PersonalMfaStatus,
} from "../account.types";

const props = defineProps<{
  account: PersonalAccountPayload | null;
  mfaStatus: PersonalMfaStatus | null;
  sessions: PersonalAccountSession[];
  loading: boolean;
  error: string | null;
  mfaError: string | null;
  sessionsError: string | null;
  passwordLoading: boolean;
  passwordError: string | null;
  passwordSuccessNonce: number;
  revokingSessionId: string | null;
  formatDateTime: (value?: string | null) => string;
}>();

const emit = defineEmits<{
  refresh: [];
  changePassword: [payload: ChangePersonalPasswordInput];
  revokeSession: [session: PersonalAccountSession];
  startMfaEnroll: [];
  startMfaDisable: [];
}>();

const passwordForm = reactive({
  oldPassword: "",
  newPassword: "",
  confirmPassword: "",
});
const fieldError = ref<string | null>(null);
const showPasswords = ref(false);
const { t } = useI18n();

const functionGrantCount = computed(
  () => props.account?.permissions.functions.length ?? 0,
);
const contractGrantCount = computed(
  () => props.account?.permissions.contracts.length ?? 0,
);
const primaryGroups = computed(() => props.account?.user.groups ?? []);
const passwordToggleLabel = computed(() =>
  showPasswords.value
    ? t("settings.console.personal.hidePasswords")
    : t("settings.console.personal.showPasswords"),
);

watch(
  () => props.passwordSuccessNonce,
  () => {
    passwordForm.oldPassword = "";
    passwordForm.newPassword = "";
    passwordForm.confirmPassword = "";
    fieldError.value = null;
    showPasswords.value = false;
  },
);

function submitPasswordChange() {
  if (props.passwordLoading) {
    return;
  }

  if (!passwordForm.oldPassword) {
    fieldError.value = t("settings.console.personal.validationCurrentPassword");
    return;
  }

  const MIN_PASSWORD_LENGTH = 12;

  if (passwordForm.newPassword.length < MIN_PASSWORD_LENGTH) {
    fieldError.value = t("settings.console.personal.validationNewPassword");
    return;
  }

  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    fieldError.value = t("settings.console.personal.validationPasswordMismatch");
    return;
  }

  fieldError.value = null;
  emit("changePassword", {
    old_password: passwordForm.oldPassword,
    new_password: passwordForm.newPassword,
  });
}
</script>

<template>
  <section class="settings-personal" aria-labelledby="settings-personal-title">
    <section class="settings-panel">
      <header class="settings-panel__header">
        <div>
          <h2 id="settings-personal-title" class="settings-panel__title">
            {{ t("settings.console.personal.title") }}
          </h2>
          <p class="settings-panel__subtitle">
            {{ t("settings.console.personal.subtitle") }}
          </p>
        </div>
        <AppButton
          variant="secondary"
          size="sm"
          :loading="loading"
          @click="emit('refresh')"
        >
          <AppIcon name="refresh" size="xs" />
          <span>{{ t("settings.console.personal.refresh") }}</span>
        </AppButton>
      </header>

      <div v-if="error" class="alert alert-danger settings-panel__alert" role="alert">
        {{ error }}
      </div>

      <div class="settings-personal__body" :aria-busy="loading">
        <div class="settings-personal__identity">
          <div class="settings-personal__avatar" aria-hidden="true">
            <AppIcon name="user" size="sm" />
          </div>
          <div class="settings-personal__identity-copy">
            <div class="settings-personal__name">
              {{ account?.user.display_name || t("settings.console.personal.currentUser") }}
            </div>
            <div class="settings-personal__meta">
              {{ account?.user.username || t("settings.console.personal.usernameUnavailable") }}
              <span aria-hidden="true">/</span>
              {{ account?.user.email || t("settings.console.personal.emailUnavailable") }}
            </div>
          </div>
        </div>

        <dl class="settings-personal__summary">
          <div>
            <dt>{{ t("settings.console.personal.functionGrants") }}</dt>
            <dd>{{ functionGrantCount }}</dd>
          </div>
          <div>
            <dt>{{ t("settings.console.personal.dataScopes") }}</dt>
            <dd>{{ contractGrantCount }}</dd>
          </div>
          <div>
            <dt>{{ t("settings.console.personal.mfaStatus") }}</dt>
            <dd>
              {{
                mfaStatus?.enabled
                  ? t("settings.console.personal.mfaEnabled")
                  : mfaStatus?.enrolled
                    ? t("settings.console.personal.mfaEnrolled")
                    : t("settings.console.personal.mfaNotEnabled")
              }}
            </dd>
          </div>
          <div>
            <dt>{{ t("settings.console.personal.sessions") }}</dt>
            <dd>{{ sessions.length }}</dd>
          </div>
        </dl>

        <div class="settings-personal__access-grid">
          <section class="settings-personal-card">
            <div class="settings-personal-card__header">
              <h3>{{ t("settings.console.personal.groups") }}</h3>
              <span class="badge badge-neutral">{{ primaryGroups.length }}</span>
            </div>
            <div v-if="primaryGroups.length" class="settings-chip-list">
              <span
                v-for="group in primaryGroups"
                :key="group"
                class="badge badge-info"
              >
                {{ group }}
              </span>
            </div>
            <p v-else class="settings-personal-card__empty">
              {{ t("settings.console.personal.noGroups") }}
            </p>
          </section>

          <section class="settings-personal-card">
            <div class="settings-personal-card__header">
              <h3>{{ t("settings.console.personal.mfaRecovery") }}</h3>
              <span
                class="badge"
                :class="mfaStatus?.enabled ? 'badge-success' : 'badge-warning'"
              >
                {{ mfaStatus?.enabled ? t("settings.console.common.protected") : t("settings.console.common.review") }}
              </span>
            </div>
            <div class="settings-personal-card__actions">
              <button
                v-if="!mfaStatus?.enabled"
                type="button"
                class="btn btn-primary btn-sm"
                @click="emit('startMfaEnroll')"
              >
                {{ t("settings.console.personal.mfaEnableCta") }}
              </button>
              <button
                v-else
                type="button"
                class="btn btn-danger btn-sm"
                @click="emit('startMfaDisable')"
              >
                {{ t("settings.console.personal.mfaDisableCta") }}
              </button>
            </div>
            <div v-if="mfaError" class="alert alert-danger" role="alert">
              {{ mfaError }}
            </div>
            <dl v-else class="settings-personal-card__facts">
              <div>
                <dt>{{ t("settings.console.personal.enrolled") }}</dt>
                <dd>{{ mfaStatus?.enrolled ? t("settings.console.common.yes") : t("settings.console.common.no") }}</dd>
              </div>
              <div>
                <dt>{{ t("settings.console.common.enabled") }}</dt>
                <dd>{{ mfaStatus?.enabled ? t("settings.console.common.yes") : t("settings.console.common.no") }}</dd>
              </div>
              <div>
                <dt>{{ t("settings.console.personal.recoveryCodes") }}</dt>
                <dd>{{ mfaStatus?.recovery_codes_left ?? "-" }}</dd>
              </div>
            </dl>
          </section>
        </div>
      </div>
    </section>

    <section class="settings-panel">
      <header class="settings-panel__header">
        <div>
          <h2 class="settings-panel__title">
            {{ t("settings.console.personal.changePasswordTitle") }}
          </h2>
          <p class="settings-panel__subtitle">
            {{ t("settings.console.personal.changePasswordSubtitle") }}
          </p>
        </div>
      </header>

      <form
        class="settings-personal-password"
        :aria-busy="passwordLoading"
        @submit.prevent="submitPasswordChange"
      >
        <div
          class="form-group"
          :class="{ 'field-error': fieldError || passwordError }"
        >
          <label for="settings-current-password" class="label label-required">
            {{ t("settings.console.personal.currentPassword") }}
          </label>
          <input
            id="settings-current-password"
            v-model="passwordForm.oldPassword"
            class="input"
            :type="showPasswords ? 'text' : 'password'"
            autocomplete="current-password"
            :aria-invalid="Boolean(fieldError || passwordError)"
          />
        </div>

        <div
          class="form-group"
          :class="{ 'field-error': fieldError || passwordError }"
        >
          <label for="settings-new-password" class="label label-required">
            {{ t("settings.console.personal.newPassword") }}
          </label>
          <input
            id="settings-new-password"
            v-model="passwordForm.newPassword"
            class="input"
            :type="showPasswords ? 'text' : 'password'"
            autocomplete="new-password"
            :aria-invalid="Boolean(fieldError || passwordError)"
          />
          <p class="help-text">{{ t("settings.console.personal.passwordHelp") }}</p>
        </div>

        <div
          class="form-group"
          :class="{ 'field-error': fieldError || passwordError }"
        >
          <label for="settings-confirm-password" class="label label-required">
            {{ t("settings.console.personal.confirmNewPassword") }}
          </label>
          <input
            id="settings-confirm-password"
            v-model="passwordForm.confirmPassword"
            class="input"
            :type="showPasswords ? 'text' : 'password'"
            autocomplete="new-password"
            :aria-invalid="Boolean(fieldError || passwordError)"
          />
        </div>

        <div v-if="fieldError" class="error-text">{{ fieldError }}</div>
        <div v-if="passwordError" class="error-text">{{ passwordError }}</div>

        <div class="settings-personal-password__actions">
          <AppButton
            type="button"
            variant="secondary"
            size="sm"
            :aria-pressed="showPasswords"
            @click="showPasswords = !showPasswords"
          >
            {{ passwordToggleLabel }}
          </AppButton>
          <AppButton
            type="submit"
            variant="primary"
            size="sm"
            :loading="passwordLoading"
          >
            {{ t("settings.console.personal.changePassword") }}
          </AppButton>
        </div>
      </form>
    </section>

    <section class="settings-panel">
      <header class="settings-panel__header">
        <div>
          <h2 class="settings-panel__title">
            {{ t("settings.console.personal.activeSessionsTitle") }}
          </h2>
          <p class="settings-panel__subtitle">
            {{ t("settings.console.personal.activeSessionsSubtitle") }}
          </p>
        </div>
        <span class="badge badge-neutral">
          {{ t("settings.console.personal.sessions") }}: {{ sessions.length }}
        </span>
      </header>

      <div v-if="sessionsError" class="alert alert-danger settings-panel__alert" role="alert">
        {{ sessionsError }}
      </div>

      <div class="settings-personal-sessions">
        <article
          v-for="session in sessions"
          :key="session.id"
          class="settings-personal-session"
        >
          <div class="settings-personal-session__copy">
            <div class="settings-record-primary">
              {{ session.ip_address || t("settings.console.common.unknownIp") }}
            </div>
            <div class="settings-record-secondary">
              {{ session.user_agent || t("settings.console.common.userAgentUnavailable") }}
            </div>
            <div class="settings-record-secondary">
              {{ t("settings.console.personal.lastActivity", { date: formatDateTime(session.last_activity_at) }) }}
            </div>
            <div class="settings-record-secondary">
              {{ t("settings.console.personal.expires", { date: formatDateTime(session.expires_at) }) }}
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

        <div v-if="!sessionsError && sessions.length === 0" class="settings-card-state">
          {{ t("settings.console.personal.noSessions") }}
        </div>
      </div>
    </section>
  </section>
</template>

<style scoped>
.settings-personal {
  display: grid;
  gap: var(--space-5);
}

.settings-personal__body {
  display: grid;
  gap: var(--space-5);
  padding: var(--space-5);
}

.settings-personal__identity {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  min-width: 0;
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
}

.settings-personal__avatar {
  width: 3rem;
  height: 3rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  border-radius: var(--radius-md);
  background: var(--bg-selected);
  color: var(--action-primary);
}

.settings-personal__identity-copy {
  min-width: 0;
}

.settings-personal__name {
  color: var(--text-primary);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
}

.settings-personal__meta {
  margin-top: var(--space-1);
  color: var(--text-secondary);
  overflow-wrap: anywhere;
}

.settings-personal__summary {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-3);
  margin: 0;
}

.settings-personal__summary div,
.settings-personal-card,
.settings-personal-session {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.settings-personal__summary div {
  padding: var(--space-4);
}

.settings-personal__summary dt,
.settings-personal-card__facts dt {
  color: var(--text-tertiary);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.settings-personal__summary dd,
.settings-personal-card__facts dd {
  margin: var(--space-1) 0 0;
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
}

.settings-personal__access-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-4);
}

.settings-personal-card {
  display: grid;
  align-content: start;
  gap: var(--space-4);
  padding: var(--space-5);
}

.settings-personal-card__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.settings-personal-card__header h3 {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-md);
}

.settings-personal-card__empty {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.settings-personal-card__facts {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-3);
  margin: 0;
}

.settings-chip-list {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.settings-personal-password {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-4);
  align-items: start;
  padding: var(--space-5);
}

.settings-personal-password .form-group {
  margin-bottom: 0;
}

.settings-personal-password__actions {
  grid-column: 1 / -1;
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.settings-personal-sessions {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-5);
}

.settings-personal-session {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4);
}

.settings-personal-session__copy {
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
  padding: var(--space-6);
  border: 1px dashed var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  text-align: center;
}

@media (max-width: 980px) {
  .settings-personal__summary,
  .settings-personal__access-grid,
  .settings-personal-password {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 680px) {
  .settings-personal__identity,
  .settings-personal-session {
    align-items: stretch;
    flex-direction: column;
  }

  .settings-personal__summary,
  .settings-personal__access-grid,
  .settings-personal-card__facts,
  .settings-personal-password {
    grid-template-columns: 1fr;
  }

  .settings-personal-password__actions {
    justify-content: stretch;
  }
}

.settings-personal-card__actions {
  display: flex;
  gap: var(--space-2);
}
</style>
