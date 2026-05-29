<script setup lang="ts">
import type { PersonalAccountSession, PersonalMfaStatus } from "../account.types";
import SettingsPersonalSessionsPanel from "./SettingsPersonalSessionsPanel.vue";

defineProps<{
  mfaStatus: PersonalMfaStatus | null;
  sessions: PersonalAccountSession[];
  loading: boolean;
  mfaError: string | null;
  sessionsError: string | null;
  revokingSessionId: string | null;
  formatDateTime: (value?: string | null | undefined) => string;
}>();

const emit = defineEmits<{
  refresh: [];
  "revoke-session": [session: PersonalAccountSession];
  "start-mfa-enroll": [];
  "start-mfa-disable": [];
}>();
</script>

<template>
  <div style="display: grid; gap: var(--space-5);">
    <section class="settings-panel">
      <header class="settings-panel__header">
        <h2 class="settings-panel__title">Two-factor authentication</h2>
      </header>
      <div class="settings-panel__body" style="padding: var(--space-5);">
        <div v-if="mfaStatus?.enabled">
          <p>MFA is currently <strong>enabled</strong>.</p>
          <AppButton variant="danger" @click="emit('start-mfa-disable')">Disable MFA</AppButton>
        </div>
        <div v-else>
          <p>MFA is <strong>disabled</strong>.</p>
          <AppButton @click="emit('start-mfa-enroll')">Enable MFA</AppButton>
        </div>
      </div>
    </section>

    <SettingsPersonalSessionsPanel
      :sessions="sessions"
      :sessions-error="sessionsError"
      :revoking-session-id="revokingSessionId"
      :format-date-time="formatDateTime"
      @revoke-session="emit('revoke-session', $event)"
    />
  </div>
</template>
