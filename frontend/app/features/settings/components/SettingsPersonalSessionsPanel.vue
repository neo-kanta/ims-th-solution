<script setup lang="ts">
import type { PersonalAccountSession } from "../account.types";

const props = defineProps<{
  sessions: PersonalAccountSession[];
  sessionsError: string | null;
  revokingSessionId: string | null;
  formatDateTime: (value?: string | null) => string;
}>();

const emit = defineEmits<{
  revokeSession: [session: PersonalAccountSession];
}>();

const { t } = useI18n();
</script>

<template>
  <section class="settings-personal-sessions-panel" aria-labelledby="settings-sessions-title">
    <header class="settings-panel__header">
      <div>
        <h2 id="settings-sessions-title" class="settings-panel__title">
          Web sessions
        </h2>
        <p class="settings-panel__subtitle">
          This is a list of devices that have logged into your account. Revoke any sessions that you do not recognize.
        </p>
      </div>
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
        <div class="settings-personal-session__icon">
          <AppIcon name="monitor" size="md" />
        </div>
        <div class="settings-personal-session__copy">
          <div class="settings-record-primary">
            {{ session.ip_address || t("settings.console.common.unknownIp") }}
          </div>
          <div class="settings-record-secondary settings-record-status">
            <span class="status-dot"></span> Active
          </div>
          <div class="settings-record-secondary">
            Your current session
          </div>
          <div class="settings-record-secondary">
            Seen in {{ session.user_agent || t("settings.console.common.userAgentUnavailable") }}
          </div>
        </div>
        <AppButton
          variant="secondary"
          size="sm"
          :loading="revokingSessionId === session.id"
          :disabled="Boolean(revokingSessionId)"
          @click="emit('revokeSession', session)"
        >
          Details
        </AppButton>
      </article>

      <div v-if="!sessionsError && sessions.length === 0" class="settings-card-state">
        {{ t("settings.console.personal.noSessions") }}
      </div>
    </div>
  </section>
</template>

<style scoped>
.settings-personal-sessions-panel {
  display: grid;
  gap: var(--space-5);
  padding: var(--space-5) 0;
}

.settings-panel__header {
  margin-bottom: var(--space-4);
}

.settings-panel__title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-medium);
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: var(--space-2);
  margin-bottom: var(--space-3);
  color: var(--text-primary);
}

.settings-panel__subtitle {
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.settings-personal-sessions {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.settings-personal-session {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4);
  border-bottom: 1px solid var(--border-subtle);
}
.settings-personal-session:last-child {
  border-bottom: none;
}

.settings-personal-session__icon {
  color: var(--text-tertiary);
  margin-top: var(--space-1);
}

.settings-personal-session__copy {
  flex: 1;
  min-width: 0;
}

.settings-record-primary {
  display: block;
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-sm);
}

.settings-record-secondary {
  display: block;
  margin-top: var(--space-1);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.settings-record-status {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  color: #1a7f37; /* GitHub green */
}
.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #1a7f37;
}

.settings-card-state {
  padding: var(--space-6);
  border: 1px dashed var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  text-align: center;
}
</style>
