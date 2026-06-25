<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { useEmailOutboxDetail } from "~/features/notifications/composables/useEmailOutbox";
import { useAppToast } from "~/composables/useAppToast";
import { useBangkokFormatter } from "~/shared/composables/useBangkokFormatter";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "NOTIFICATION_VIEW",
});

const route = useRoute();
const outboxId = route.params.outbox_id as string;

const { detail, loading, error, retrying, retryResult, retryError, load, retry } = useEmailOutboxDetail(outboxId);
const toast = useAppToast();
const { formatDateTime } = useBangkokFormatter();

const confirmRetryOpen = ref(false);

// m2: FAILED → warning (retriable), DEAD → error (exhausted)
function statusVariant(status?: string): "success" | "warning" | "error" | "neutral" | "info" {
  switch (status) {
    case "SENT": return "success";
    case "FAILED": return "warning";
    case "DEAD": return "error";
    case "PENDING": return "warning";
    case "SENDING": return "info";
    default: return "neutral";
  }
}

function canRetry(status?: string) {
  return status === "FAILED" || status === "DEAD";
}

async function confirmRetry() {
  confirmRetryOpen.value = false;
  await retry();
  if (retryError.value) {
    toast.showError(retryError.value);
  } else {
    toast.showSuccess("Email queued for retry");
  }
}

onMounted(load);
</script>

<template>
  <div>
    <AppPageHeader
      :title="detail ? `Outbox: ${detail.subject ?? outboxId}` : 'Email Outbox Detail'"
      description="Full delivery record including email body and error details"
    >
      <template #eyebrow>
        <div class="breadcrumb">
          <NuxtLink to="/admin/notifications/email">Email operations</NuxtLink>
          <span> / </span>
          <NuxtLink to="/admin/notifications/email/outbox">Outbox</NuxtLink>
          <span> / </span>
          <span>{{ outboxId }}</span>
        </div>
      </template>
      <template #actions>
        <IMSPermissionGuard permission="NOTIFICATION_RETRY">
          <button
            v-if="detail && canRetry(detail.status)"
            class="btn btn-warning btn-sm"
            type="button"
            :disabled="retrying"
            @click="confirmRetryOpen = true"
          >
            {{ retrying ? "Retrying…" : "Retry delivery" }}
          </button>
        </IMSPermissionGuard>
      </template>
    </AppPageHeader>

    <AppLoadingState v-if="loading" message="Loading outbox record…" />

    <div v-else-if="error" class="detail-error">{{ error }}</div>

    <template v-else-if="detail">
      <!-- Status header -->
      <div class="detail-status-bar">
        <AppBadge :variant="statusVariant(detail.status)">
          {{ detail.status ?? "—" }}
        </AppBadge>
        <span class="detail-status-bar__attempts">
          Attempt {{ detail.attempts ?? 0 }} of {{ detail.max_attempts ?? "—" }}
        </span>
        <span v-if="detail.sent_at" class="detail-status-bar__sent">
          Sent {{ formatDateTime(detail.sent_at) }}
        </span>
        <span v-if="detail.next_attempt_at && canRetry(detail.status)" class="detail-status-bar__next">
          Next attempt: {{ formatDateTime(detail.next_attempt_at) }}
        </span>
      </div>

      <!-- Recipient -->
      <div class="card" style="margin-top: 16px;">
        <div class="card-header">Recipient</div>
        <table class="detail-table">
          <tbody>
            <tr>
              <td class="detail-table__label">Username</td>
              <td>{{ detail.recipient?.username ?? "—" }}</td>
            </tr>
            <tr>
              <td class="detail-table__label">Display name</td>
              <td>{{ detail.recipient?.display_name ?? detail.to_name ?? "—" }}</td>
            </tr>
            <tr>
              <td class="detail-table__label">Email</td>
              <td>{{ detail.to_email ?? detail.recipient?.email ?? "—" }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Message -->
      <div class="card" style="margin-top: 16px;">
        <div class="card-header">Message</div>
        <table class="detail-table">
          <tbody>
            <tr>
              <td class="detail-table__label">Subject</td>
              <td>{{ detail.subject ?? "—" }}</td>
            </tr>
            <tr>
              <td class="detail-table__label">Provider message ID</td>
              <td>{{ detail.provider_message_id ?? "—" }}</td>
            </tr>
          </tbody>
        </table>

        <!-- m3: renamed "HTML body" → "HTML source" to match what's shown (escaped source text) -->
        <div v-if="detail.body_html" class="detail-body">
          <div class="detail-body__label">HTML source</div>
          <pre class="detail-body__pre">{{ detail.body_html }}</pre>
        </div>

        <div v-else-if="detail.body_text" class="detail-body">
          <div class="detail-body__label">Text body</div>
          <pre class="detail-body__pre">{{ detail.body_text }}</pre>
        </div>
      </div>

      <!-- Event context -->
      <div class="card" style="margin-top: 16px;">
        <div class="card-header">Event context</div>
        <table class="detail-table">
          <tbody>
            <tr>
              <td class="detail-table__label">Event type</td>
              <td>{{ detail.event?.type ?? "—" }}</td>
            </tr>
            <tr>
              <td class="detail-table__label">Category</td>
              <td>{{ detail.event?.category ?? "—" }}</td>
            </tr>
            <tr>
              <td class="detail-table__label">Severity</td>
              <td>{{ detail.event?.severity ?? "—" }}</td>
            </tr>
            <tr>
              <td class="detail-table__label">Business type</td>
              <td>{{ detail.context?.business_type ?? "—" }}</td>
            </tr>
            <tr>
              <td class="detail-table__label">Business reference</td>
              <td>{{ detail.context?.business_reference ?? "—" }}</td>
            </tr>
            <tr>
              <td class="detail-table__label">Business title</td>
              <td>{{ detail.context?.business_title ?? "—" }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Error details -->
      <div v-if="detail.last_error" class="card card--error" style="margin-top: 16px;">
        <div class="card-header card-header--error">Last error</div>
        <pre class="detail-error-pre">{{ detail.last_error }}</pre>
      </div>

      <!-- Timestamps -->
      <div class="card" style="margin-top: 16px;">
        <div class="card-header">Timeline</div>
        <table class="detail-table">
          <tbody>
            <tr>
              <td class="detail-table__label">Created</td>
              <td>{{ formatDateTime(detail.created_at) }}</td>
            </tr>
            <tr>
              <td class="detail-table__label">Updated</td>
              <td>{{ formatDateTime(detail.updated_at) }}</td>
            </tr>
            <tr v-if="detail.sent_at">
              <td class="detail-table__label">Sent</td>
              <td>{{ formatDateTime(detail.sent_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>

    <!-- Retry confirmation -->
    <AppConfirmDialog
      :open="confirmRetryOpen"
      title="Retry email delivery"
      :description="`Re-queue this message for delivery to ${detail?.to_email ?? detail?.recipient?.email ?? 'recipient'}. The worker will pick it up on the next cycle.`"
      confirm-label="Retry"
      tone="warning"
      :loading="retrying"
      @cancel="confirmRetryOpen = false"
      @confirm="confirmRetry"
    />
  </div>
</template>

<style scoped>
.detail-error {
  padding: 24px;
  color: var(--color-danger);
}

.detail-status-bar {
  display: flex;
  gap: 16px;
  align-items: center;
  margin-top: 16px;
  padding: 12px 0;
  flex-wrap: wrap;
}

.detail-status-bar__attempts,
.detail-status-bar__sent,
.detail-status-bar__next {
  font-size: 13px;
  color: var(--text-secondary);
}

.detail-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.detail-table tr {
  border-bottom: 1px solid var(--border-subtle);
}

.detail-table tr:last-child {
  border-bottom: none;
}

.detail-table td {
  padding: 10px 20px;
  color: var(--text-primary);
}

.detail-table__label {
  color: var(--text-secondary);
  width: 200px;
  font-weight: 500;
}

.detail-body {
  padding: 16px 20px;
  border-top: 1px solid var(--border-subtle);
}

.detail-body__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.detail-body__pre {
  font-size: 12px;
  font-family: var(--font-mono, monospace);
  background: var(--bg-canvas);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  padding: 12px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--text-primary);
  max-height: 300px;
  overflow-y: auto;
}

.card--error {
  border-color: var(--color-danger, #e53e3e);
}

.card-header--error {
  color: var(--color-danger, #e53e3e);
}

.detail-error-pre {
  padding: 12px 20px 16px;
  font-size: 12px;
  font-family: var(--font-mono, monospace);
  color: var(--color-danger, #e53e3e);
  white-space: pre-wrap;
  word-break: break-word;
}

.card-header {
  padding: 12px 20px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-subtle);
}
</style>
