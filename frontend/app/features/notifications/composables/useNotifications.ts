import { ref, computed } from "vue";
import { notificationApi, notificationErrorMessage } from "../services/notificationApi";
import type { NotificationItem } from "../types";

export function useNotifications() {
  const items = ref<NotificationItem[]>([]);
  const total = ref(0);
  const unread = ref(0);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const hasUnread = computed(() => unread.value > 0);
  const unreadCount = computed(() => Math.min(unread.value, 99));

  async function load(params: { unread_only?: boolean; limit?: number; offset?: number } = {}) {
    loading.value = true;
    error.value = null;
    try {
      const res = await notificationApi.list({ limit: 50, ...params });
      items.value = res.items ?? [];
      total.value = res.total ?? 0;
      unread.value = res.unread ?? 0;
    } catch (e) {
      error.value = notificationErrorMessage(e, "Unable to load notifications");
    } finally {
      loading.value = false;
    }
  }

  async function markRead(id: string) {
    try {
      await notificationApi.markRead(id);
      const idx = items.value.findIndex((n) => (n.notification_id ?? n.id) === id);
      if (idx !== -1) {
        items.value[idx] = { ...items.value[idx], is_read: true };
        if (unread.value > 0) unread.value--;
      }
    } catch (e) {
      error.value = notificationErrorMessage(e, "Unable to mark notification as read");
    }
  }

  async function markAllRead() {
    try {
      await notificationApi.markAllRead();
      items.value = items.value.map((n) => ({ ...n, is_read: true }));
      unread.value = 0;
    } catch (e) {
      error.value = notificationErrorMessage(e, "Unable to mark all notifications as read");
    }
  }

  return {
    items,
    total,
    unread,
    unreadCount,
    hasUnread,
    loading,
    error,
    load,
    markRead,
    markAllRead,
  };
}
