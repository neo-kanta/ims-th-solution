<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useEmailOutbox } from "~/features/notifications/composables/useEmailOutbox";
import { notificationApi, notificationErrorMessage } from "~/features/notifications/services/notificationApi";
import { useAppToast } from "~/composables/useAppToast";
import { useBangkokFormatter } from "~/shared/composables/useBangkokFormatter";
import type { OutboxFilter } from "~/features/notifications/types";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "NOTIFICATION_VIEW",
});

const route = useRoute();
const router = useRouter();
const toast = useAppToast();
const { formatDateTime } = useBangkokFormatter();

const { items, total, loading, error, load, applyFilter } = useEmailOutbox();

// M2: CANCELLED added as a valid status
const statusOptions = ["PENDING", "SENDING", "SENT", "FAILED", "DEAD", "CANCELLED"];

const filterStatus = ref((route.query.status as string) ?? "");
const filterRecipient = ref("");
const filterEventType = ref("");

// M1: per-row retry state
const retryingId = ref<string | null>(null);
const confirmRetryItem = ref<{ outbox_id: string; to_email?: string; recipient_email?: string } | null>(null);

function search() {
  const f: OutboxFilter = { limit: 50, offset: 0 };
  if (filterStatus.value) f.status = filterStatus.value;
  if (filterRecipient.value.trim()) f.recipient_username = filterRecipient.value.trim();
  if (filterEventType.value.trim()) f.event_type = filterEventType.value.trim();
  void router.replace({ query: filterStatus.value ? { status: filterStatus.value } : {} });
  void applyFilter(f);
}

function reset() {
  filterStatus.value = "";
  filterRecipient.value = "";
  filterEventType.value = "";
  void router.replace({ query: {} });
  void applyFilter({});
}

// m2: FAILED → warning (retriable), DEAD → error (exhausted)
function statusVariant(status?: string): "success" | "warning" | "error" | "neutral" | "info" {
  switch (status) {
    case "SENT": return "success";
    case "FAILED": return "warning";
    case "DEAD": return "error";
    case "PENDING": return "info";
    case "SENDING": return "info";
    default: return "neutral";
  }
}

function canRetry(status?: string) {
  return status === "FAILED" || status === "DEAD";
}

// M1: row-level retry flow
function openRetryConfirm(item: typeof items.value[0]) {
  confirmRetryItem.value = {
    outbox_id: item.outbox_id ?? "",
    to_email: item.to_email ?? item.recipient?.email,
  };
}

async function confirmRowRetry() {
  if (!confirmRetryItem.value?.outbox_id) return;
  const id = confirmRetryItem.value.outbox_id;
  confirmRetryItem.value = null;
  retryingId.value = id;
  try {
    await notificationApi.retryOutbox(id);
    toast.showSuccess("Email queued for retry");
    void load();
  } catch (e) {
    toast.showError(notificationErrorMessage(e, "Unable to retry email delivery"));
  } finally {
    retryingId.value = null;
  }
}

onMounted(() => {
  const initialFilter: OutboxFilter = { limit: 50, offset: 0 };
  if (filterStatus.value) initialFilter.status = filterStatus.value;
  void applyFilter(initialFilter);
});
</script>

<template>
  <div>
    <AppPageHeader
      title="Email Outbox"
      description="All email delivery records with status and retry controls"
    >
      <template #eyebrow>
        <div class="breadcrumb">
          <NuxtLink to="/admin/notifications/email">Email operations</NuxtLink>
          <span> / </span>
          <span>Outbox</span>
        </div>
      </template>
    </AppPageHeader>

    <!-- Filters -->
    <div class="outbox-filters card" style="margin-top: 16px;">
      <div class="outbox-filters__row">
        <div class="outbox-filters__field">
          <label class="outbox-filters__label">Status</label>
          <select v-model="filterStatus" class="outbox-filters__select">
            <option value="">All statuses</option>
            <option v-for="s in statusOptions" :key="s" :value="s">{{ s }}</option>
          </select>
        </div>
        <div class="outbox-filters__field">
          <label class="outbox-filters__label">Recipient username</label>
          <input
            v-model="filterRecipient"
            class="outbox-filters__input"
            placeholder="username…"
            @keyup.enter="search"
          />
        </div>
        <div class="outbox-filters__field">
          <label class="outbox-filters__label">Event type</label>
          <input
            v-model="filterEventType"
            class="outbox-filters__input"
            placeholder="event.type…"
            @keyup.enter="search"
          />
        </div>
        <div class="outbox-filters__actions">
          <button class="btn btn-primary btn-sm" type="button" @click="search">Search</button>
          <button class="btn btn-tertiary btn-sm" type="button" @click="reset">Reset</button>
        </div>
      </div>
    </div>

    <AppLoadingState v-if="loading" message="Loading outbox…" />

    <div v-else-if="error" class="outbox-error">{{ error }}</div>

    <div v-else class="card" style="margin-top: 16px;">
      <div class="outbox-meta">{{ total }} records</div>

      <AppEmptyState
        v-if="!items.length"
        title="No outbox records"
        description="No email delivery records match the current filter."
        icon="search"
      />

      <table v-else class="outbox-table">
        <thead>
          <tr>
            <th>Status</th>
            <th>Recipient</th>
            <th>Subject</th>
            <th>Event</th>
            <th>Attempts</th>
            <th>Created</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.outbox_id">
            <td>
              <AppBadge :variant="statusVariant(item.status)" size="sm">
                {{ item.status ?? "—" }}
              </AppBadge>
            </td>
            <td class="outbox-table__recipient">
              <span>{{ item.recipient?.display_name ?? item.to_name ?? "—" }}</span>
              <span class="outbox-table__email">{{ item.to_email ?? item.recipient?.email ?? "" }}</span>
            </td>
            <td class="outbox-table__subject">{{ item.subject ?? "—" }}</td>
            <td class="outbox-table__event">{{ item.event?.type ?? "—" }}</td>
            <td class="outbox-table__attempts">
              {{ item.attempts ?? 0 }}/{{ item.max_attempts ?? "—" }}
            </td>
            <td class="outbox-table__date">{{ formatDateTime(item.created_at) }}</td>
            <td class="outbox-table__actions">
              <!-- M1: per-row retry button for FAILED/DEAD rows -->
              <IMSPermissionGuard v-if="canRetry(item.status)" permission="NOTIFICATION_RETRY">
                <button
                  class="btn btn-warning btn-xs"
                  type="button"
                  :disabled="retryingId === item.outbox_id"
                  style="margin-right: 4px;"
                  @click="openRetryConfirm(item)"
                >
                  {{ retryingId === item.outbox_id ? "…" : "Retry" }}
                </button>
              </IMSPermissionGuard>
              <NuxtLink
                :to="`/admin/notifications/email/outbox/${item.outbox_id}`"
                class="btn btn-tertiary btn-xs"
              >
                Details
              </NuxtLink>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- M1: row retry confirm dialog -->
    <AppConfirmDialog
      :open="!!confirmRetryItem"
      title="Retry email delivery"
      :description="`Re-queue this message for delivery to ${confirmRetryItem?.to_email ?? 'recipient'}. The worker will pick it up on the next cycle.`"
      confirm-label="Retry"
      tone="warning"
      @cancel="confirmRetryItem = null"
      @confirm="confirmRowRetry"
    />
  </div>
</template>

<style scoped>
.outbox-error {
  padding: 24px;
  color: var(--color-danger);
}

.outbox-filters {
  padding: 16px 20px;
}

.outbox-filters__row {
  display: flex;
  gap: 12px;
  align-items: flex-end;
  flex-wrap: wrap;
}

.outbox-filters__field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.outbox-filters__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}

.outbox-filters__select,
.outbox-filters__input {
  height: 32px;
  padding: 0 10px;
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  background: var(--bg-canvas);
  color: var(--text-primary);
  font-size: 13px;
  min-width: 140px;
}

.outbox-filters__actions {
  display: flex;
  gap: 8px;
}

.outbox-meta {
  padding: 10px 20px;
  font-size: 12px;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-subtle);
}

.outbox-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.outbox-table th {
  padding: 10px 16px;
  text-align: left;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}

.outbox-table td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-subtle);
  color: var(--text-primary);
  vertical-align: top;
}

.outbox-table tbody tr:last-child td {
  border-bottom: none;
}

.outbox-table__recipient {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.outbox-table__email {
  font-size: 11px;
  color: var(--text-tertiary);
}

.outbox-table__subject {
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.outbox-table__event {
  font-size: 12px;
  color: var(--text-secondary);
}

.outbox-table__attempts,
.outbox-table__date {
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
}

.outbox-table__actions {
  white-space: nowrap;
}
</style>
