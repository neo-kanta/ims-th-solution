<script setup lang="ts">
import { onMounted } from "vue";
import { useEmailHealth } from "~/features/notifications/composables/useEmailHealth";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "NOTIFICATION_HEALTH",
});

const { health, loading, error, load } = useEmailHealth();

onMounted(load);
</script>

<template>
  <div>
    <AppPageHeader
      title="Email Operations"
      description="Email delivery configuration, queue health, and diagnostics"
    >
      <template #eyebrow>
        <div class="breadcrumb">
          <NuxtLink to="/admin/notifications/email">Email operations</NuxtLink>
        </div>
      </template>
      <template #actions>
        <IMSPermissionGuard permission="NOTIFICATION_TEST">
          <NuxtLink to="/admin/notifications/email/test" class="btn btn-secondary btn-sm">
            Send test email
          </NuxtLink>
        </IMSPermissionGuard>
        <IMSPermissionGuard permission="NOTIFICATION_VIEW">
          <NuxtLink to="/admin/notifications/email/outbox" class="btn btn-secondary btn-sm">
            View outbox
          </NuxtLink>
        </IMSPermissionGuard>
      </template>
    </AppPageHeader>

    <AppLoadingState v-if="loading" message="Loading email health…" />

    <div v-else-if="error" class="email-ops__error">{{ error }}</div>

    <template v-else-if="health">
      <!-- Status overview -->
      <div class="email-ops__grid">
        <div class="stat-card">
          <div class="stat-card__label">Worker status</div>
          <div class="stat-card__value">
            <AppBadge :variant="health.worker_enabled ? 'success' : 'neutral'">
              {{ health.worker_enabled ? "Enabled" : "Disabled" }}
            </AppBadge>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-card__label">Email sending</div>
          <div class="stat-card__value">
            <AppBadge :variant="health.enabled ? 'success' : 'warning'">
              {{ health.enabled ? "Active" : "Inactive" }}
            </AppBadge>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-card__label">Mode</div>
          <div class="stat-card__value">
            <AppBadge :variant="health.send_real_email ? 'info' : 'neutral'">
              {{ health.send_real_email ? "Live email" : "Sandbox" }}
            </AppBadge>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-card__label">Test endpoint</div>
          <div class="stat-card__value">
            <AppBadge :variant="health.test_endpoint_enabled ? 'success' : 'neutral'">
              {{ health.test_endpoint_enabled ? "Enabled" : "Disabled" }}
            </AppBadge>
          </div>
        </div>
      </div>

      <!-- Queue stats — m5: failed/dead counts link to filtered outbox -->
      <div class="card" style="margin-top: 16px;">
        <div class="card-header">Queue statistics</div>
        <table class="data-table">
          <tbody>
            <tr>
              <td class="data-table__label">Pending</td>
              <td>
                <NuxtLink to="/admin/notifications/email/outbox?status=PENDING" class="data-table__count-link">
                  {{ health.pending_count ?? 0 }}
                </NuxtLink>
              </td>
            </tr>
            <tr>
              <td class="data-table__label">Failed</td>
              <td>
                <NuxtLink
                  to="/admin/notifications/email/outbox?status=FAILED"
                  :class="['data-table__count-link', (health.failed_count ?? 0) > 0 ? 'text-warning' : '']"
                >
                  {{ health.failed_count ?? 0 }}
                </NuxtLink>
              </td>
            </tr>
            <tr>
              <td class="data-table__label">Dead</td>
              <td>
                <NuxtLink
                  to="/admin/notifications/email/outbox?status=DEAD"
                  :class="['data-table__count-link', (health.dead_count ?? 0) > 0 ? 'text-danger' : '']"
                >
                  {{ health.dead_count ?? 0 }}
                </NuxtLink>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- SMTP config (no secrets) -->
      <div class="card" style="margin-top: 16px;">
        <div class="card-header">SMTP configuration</div>
        <table class="data-table">
          <tbody>
            <tr>
              <td class="data-table__label">Host</td>
              <td>{{ health.smtp_host ?? "—" }}</td>
            </tr>
            <tr>
              <td class="data-table__label">Port</td>
              <td>{{ health.smtp_port ?? "—" }}</td>
            </tr>
            <tr>
              <td class="data-table__label">TLS mode</td>
              <td>{{ health.smtp_tls_mode ?? "—" }}</td>
            </tr>
            <tr>
              <td class="data-table__label">From address</td>
              <td>{{ health.from_address ?? "—" }}</td>
            </tr>
            <tr>
              <td class="data-table__label">From name</td>
              <td>{{ health.from_name ?? "—" }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Worker config -->
      <div class="card" style="margin-top: 16px;">
        <div class="card-header">Worker configuration</div>
        <table class="data-table">
          <tbody>
            <tr>
              <td class="data-table__label">Batch size</td>
              <td>{{ health.worker_batch_size ?? "—" }}</td>
            </tr>
            <tr>
              <td class="data-table__label">Worker interval</td>
              <td>{{ health.worker_interval ?? "—" }}</td>
            </tr>
            <tr>
              <td class="data-table__label">Stale sending timeout</td>
              <td>{{ health.stale_sending_timeout ?? "—" }}</td>
            </tr>
            <tr>
              <td class="data-table__label">Retry policy</td>
              <td>
                <span v-if="health.retry_policy?.length">
                  {{ health.retry_policy.join(", ") }}
                </span>
                <span v-else>—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<style scoped>
.email-ops__error {
  padding: 24px;
  color: var(--color-danger);
}

.email-ops__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 12px;
  margin-top: 16px;
}

.stat-card {
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  padding: 16px;
}

.stat-card__label {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.stat-card__value {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.data-table tr {
  border-bottom: 1px solid var(--border-subtle);
}

.data-table tr:last-child {
  border-bottom: none;
}

.data-table td {
  padding: 10px 20px;
  color: var(--text-primary);
}

.data-table__label {
  color: var(--text-secondary);
  width: 200px;
  font-weight: 500;
}

.data-table__count-link {
  color: inherit;
  text-decoration: none;
  font-weight: 500;
}

.data-table__count-link:hover {
  text-decoration: underline;
  color: var(--text-link);
}

.text-warning {
  color: var(--color-warning, #d97706);
}

.text-danger {
  color: var(--color-danger, #e53e3e);
}

.card-header {
  padding: 12px 20px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-subtle);
}
</style>
