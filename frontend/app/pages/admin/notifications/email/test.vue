<script setup lang="ts">
import { useTestEmail } from "~/features/notifications/composables/useTestEmail";
import { useAppToast } from "~/composables/useAppToast";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "NOTIFICATION_TEST",
});

const { form, sending, result, error, send, reset } = useTestEmail();
const toast = useAppToast();

async function handleSend() {
  await send();
  if (error.value) {
    toast.showError(error.value);
  } else if (result.value) {
    toast.showSuccess(
      `Test email queued (outbox: ${result.value.outbox_id ?? "—"}, status: ${result.value.status ?? "—"})`,
    );
  }
}
</script>

<template>
  <div>
    <AppPageHeader
      title="Send Test Email"
      description="Queue a test email to verify SMTP delivery. The background worker will send it."
    >
      <template #eyebrow>
        <div class="breadcrumb">
          <NuxtLink to="/admin/notifications/email">Email operations</NuxtLink>
          <span> / </span>
          <span>Send test email</span>
        </div>
      </template>
    </AppPageHeader>

    <div class="card test-email-card" style="margin-top: 16px;">
      <div class="test-email__notice">
        <AppIcon name="info" size="sm" />
        <span>
          The test email is delivered by the background worker, not immediately.
          Check the <NuxtLink to="/admin/notifications/email/outbox" class="notice-link">outbox</NuxtLink>
          for delivery status.
        </span>
      </div>

      <form class="test-email-form" @submit.prevent="handleSend">
        <div class="form-field">
          <label class="form-label" for="te-username">Recipient username <span class="required">*</span></label>
          <input
            id="te-username"
            v-model="form.to_username"
            class="form-input"
            type="text"
            placeholder="username"
            required
            autocomplete="off"
          />
          <div class="form-hint">The IMS username to receive the test email. Their registered address is used.</div>
        </div>

        <div class="form-field">
          <label class="form-label" for="te-subject">Subject</label>
          <input
            id="te-subject"
            v-model="form.subject"
            class="form-input"
            type="text"
            placeholder="Subject…"
          />
        </div>

        <div class="form-field">
          <label class="form-label" for="te-body">Body</label>
          <textarea
            id="te-body"
            v-model="form.body"
            class="form-textarea"
            rows="6"
            placeholder="Message body…"
          />
        </div>

        <div class="test-email-form__actions">
          <button
            class="btn btn-primary"
            type="submit"
            :disabled="sending || !form.to_username?.trim()"
          >
            {{ sending ? "Sending…" : "Send test email" }}
          </button>
          <button
            class="btn btn-tertiary"
            type="button"
            :disabled="sending"
            @click="reset"
          >
            Reset
          </button>
        </div>
      </form>

      <!-- Result panel -->
      <div v-if="result" class="test-email-result test-email-result--success">
        <div class="test-email-result__title">
          <AppIcon name="check" size="sm" />
          Test email queued
        </div>
        <table class="detail-mini-table">
          <tbody>
            <tr>
              <td>Outbox ID</td>
              <td>
                <NuxtLink
                  v-if="result.outbox_id"
                  :to="`/admin/notifications/email/outbox/${result.outbox_id}`"
                  class="notice-link"
                >
                  {{ result.outbox_id }}
                </NuxtLink>
                <span v-else>—</span>
              </td>
            </tr>
            <tr>
              <td>Status</td>
              <td>{{ result.status ?? "—" }}</td>
            </tr>
            <tr>
              <td>Recipient</td>
              <td>{{ result.recipient?.display_name ?? result.recipient?.username ?? "—" }}</td>
            </tr>
            <tr>
              <td>Event type</td>
              <td>{{ result.event?.type ?? "—" }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.test-email-card {
  max-width: 600px;
  padding: 0;
  overflow: hidden;
}

.test-email__notice {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px 20px;
  background: var(--bg-subtle);
  border-bottom: 1px solid var(--border-subtle);
  font-size: 13px;
  color: var(--text-secondary);
}

.notice-link {
  color: var(--text-link);
}

.notice-link:hover {
  text-decoration: underline;
}

.test-email-form {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

.required {
  color: var(--color-danger, #e53e3e);
}

.form-hint {
  font-size: 12px;
  color: var(--text-secondary);
}

.form-input {
  height: 36px;
  padding: 0 10px;
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  background: var(--bg-canvas);
  color: var(--text-primary);
  font-size: 13px;
  width: 100%;
}

.form-textarea {
  padding: 8px 10px;
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  background: var(--bg-canvas);
  color: var(--text-primary);
  font-size: 13px;
  width: 100%;
  resize: vertical;
  font-family: inherit;
}

.test-email-form__actions {
  display: flex;
  gap: 8px;
}

.test-email-result {
  margin: 0 20px 20px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  overflow: hidden;
}

.test-email-result--success {
  border-color: var(--color-success, #22c55e);
}

.test-email-result__title {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  background: var(--bg-subtle);
  font-size: 13px;
  font-weight: 600;
  color: var(--color-success, #22c55e);
  border-bottom: 1px solid var(--border-subtle);
}

.detail-mini-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.detail-mini-table td {
  padding: 8px 16px;
  border-bottom: 1px solid var(--border-subtle);
  color: var(--text-primary);
}

.detail-mini-table td:first-child {
  color: var(--text-secondary);
  font-weight: 500;
  width: 140px;
}

.detail-mini-table tr:last-child td {
  border-bottom: none;
}
</style>
