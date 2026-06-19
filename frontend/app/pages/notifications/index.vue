<script setup lang="ts">
import { onMounted, ref } from "vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";

definePageMeta({
  layout: 'dashboard',
  middleware: ['auth'],
});

// Notification center.
//
// Backed by the notification module baseline: GET /notifications,
// POST /notifications/{id}/read, POST /notifications/read-all.

interface NotificationItem {
  id: string;
  category: string;
  title: string;
  body?: string;
  link?: string;
  source_module?: string;
  source_type?: string;
  source_id?: string;
  is_read: boolean;
  read_at?: string | null;
  created_at: string;
}

interface NotificationListResponse {
  items: NotificationItem[];
  total: number;
  unread: number;
}

const { apiFetch } = useApi();
const items = ref<NotificationItem[]>([]);
const total = ref(0);
const unread = ref(0);
const loading = ref(false);
const error = ref<string | null>(null);

async function load() {
  loading.value = true;
  error.value = null;
  try {
    const res = await apiFetch<NotificationListResponse>('/notifications?limit=100');
    items.value = res.items ?? [];
    total.value = res.total ?? 0;
    unread.value = res.unread ?? 0;
  } catch (e: any) {
    error.value = e?.message || 'Unable to load notifications';
  } finally {
    loading.value = false;
  }
}

async function markRead(id: string) {
  try {
    await apiFetch(`/notifications/${id}/read`, { method: 'POST' });
    await load();
  } catch (e: any) {
    error.value = e?.message || 'Unable to mark as read';
  }
}

async function markAllRead() {
  try {
    await apiFetch('/notifications/read-all', { method: 'POST' });
    await load();
  } catch (e: any) {
    error.value = e?.message || 'Unable to mark all as read';
  }
}

onMounted(load);
</script>

<template>
  <div>
    <AppPageHeader
      title="My Notifications"
      description="System alerts and events routed to you based on your role and approval inbox"
    >
      <template #eyebrow>
        <div class="breadcrumb" style="margin-bottom:4px;"><span>Notifications</span></div>
      </template>
      <template #actions>
        <button
          v-if="unread > 0"
          class="btn btn-secondary btn-sm"
          @click="markAllRead"
        >
          Mark all as read
        </button>
      </template>
    </AppPageHeader>

    <div class="card" style="margin-top:16px;">
      <AppLoadingState v-if="loading" message="Loading notifications..." />
      <div v-else-if="error" style="padding:24px;color:#b91c1c;">{{ error }}</div>
      <div v-else-if="!items.length" class="empty-state" style="padding:48px 24px;text-align:center;">
        <div class="empty-title" style="font-size:16px;font-weight:600;color:#1e293b;">
          No notifications
        </div>
        <div class="empty-desc" style="margin-top:8px;color:#64748b;line-height:1.55;font-size:13px;">
          When an approval task is assigned to you or one of your requests is approved or rejected,
          a notification will appear here.
        </div>
      </div>
      <ul v-else class="notification-list">
        <li v-for="n in items" :key="n.id" :class="['notification-row', { 'is-read': n.is_read }]">
          <div class="notification-body">
            <div class="notification-title">{{ n.title }}</div>
            <div v-if="n.body" class="notification-detail">{{ n.body }}</div>
            <div class="notification-meta">
              <span class="notification-category">{{ n.category }}</span>
              <span class="notification-time">{{ n.created_at }}</span>
              <NuxtLink v-if="n.link" :to="n.link" class="notification-link">Open</NuxtLink>
            </div>
          </div>
          <button
            v-if="!n.is_read"
            class="btn btn-tertiary btn-sm"
            @click="markRead(n.id)"
          >
            Mark read
          </button>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
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
.notification-row.is-read {
  opacity: 0.7;
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
  gap: 12px;
  font-size: 12px;
  color: var(--text-tertiary);
}
.notification-link {
  color: var(--text-link);
}
</style>
