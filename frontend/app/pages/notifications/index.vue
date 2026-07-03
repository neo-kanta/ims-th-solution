<script setup lang="ts">
import { onMounted } from "vue";
import { useNotifications } from "~/features/notifications/composables/useNotifications";
import { useBangkokFormatter } from "~/shared/composables/useBangkokFormatter";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth"],
});

const { items, total, unread, hasUnread, loading, error, load, markRead, markAllRead } = useNotifications();
const { formatDateTime } = useBangkokFormatter();

onMounted(() => load({ limit: 100 }));
</script>

<template>
  <div>
    <AppPageHeader
      title="My Notifications"
      description="System alerts and events routed to you based on your role and approval inbox"
    >
      <template #eyebrow>
        <div class="breadcrumb"><span>Notifications</span></div>
      </template>
      <template #actions>
        <button
          v-if="hasUnread"
          class="btn btn-secondary btn-sm"
          type="button"
          @click="markAllRead"
        >
          Mark all as read
        </button>
      </template>
    </AppPageHeader>

    <div class="card" style="margin-top: 16px;">
      <AppLoadingState v-if="loading" message="Loading notifications…" />

      <div v-else-if="error" class="notification-error">{{ error }}</div>

      <AppEmptyState
        v-else-if="!items.length"
        title="No notifications"
        description="When an approval task is assigned to you or one of your requests is acted on, a notification will appear here."
        icon="alert"
      />

      <div v-else>
        <div class="notification-summary">
          <span>{{ total }} total</span>
          <span v-if="unread > 0" class="notification-summary__unread">{{ unread }} unread</span>
        </div>

        <ul class="notification-list">
          <li
            v-for="n in items"
            :key="n.notification_id ?? n.id"
            class="notification-row"
            :class="{ 'is-read': n.is_read }"
          >
            <div class="notification-body">
              <div class="notification-title">{{ n.title }}</div>
              <div v-if="n.body" class="notification-detail">{{ n.body }}</div>
              <div class="notification-meta">
                <span v-if="n.event?.category" class="notification-tag">{{ n.event.category }}</span>
                <span v-if="n.event?.severity" class="notification-tag">{{ n.event.severity }}</span>
                <span class="notification-time">{{ formatDateTime(n.created_at) }}</span>
                <NuxtLink
                  v-if="n.action?.url"
                  :to="n.action.url"
                  class="notification-link"
                >
                  {{ n.action.label ?? "Open" }}
                </NuxtLink>
              </div>
            </div>
            <button
              v-if="!n.is_read"
              class="btn btn-tertiary btn-sm"
              type="button"
              @click="markRead(n.notification_id ?? n.id ?? '')"
            >
              Mark read
            </button>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<style scoped>
.notification-error {
  padding: 24px;
  color: var(--color-danger);
}

.notification-summary {
  display: flex;
  gap: 12px;
  padding: 10px 20px;
  font-size: 12px;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-subtle);
}

.notification-summary__unread {
  color: var(--color-warning, #d97706);
  font-weight: 600;
}

.notification-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.notification-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-elevated);
}

.notification-row:last-child {
  border-bottom: none;
}

.notification-row.is-read {
  opacity: 0.65;
}

.notification-body {
  flex: 1;
  min-width: 0;
}

.notification-title {
  font-weight: 600;
  color: var(--text-primary);
}

.notification-detail {
  margin-top: 4px;
  color: var(--text-secondary);
  font-size: 13px;
}

.notification-meta {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 12px;
  color: var(--text-tertiary);
  align-items: center;
}

.notification-tag {
  background: var(--bg-canvas);
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
  padding: 1px 6px;
  font-size: 11px;
}

.notification-time {
  font-size: 11px;
}

.notification-link {
  color: var(--text-link);
  font-size: 12px;
}
</style>
