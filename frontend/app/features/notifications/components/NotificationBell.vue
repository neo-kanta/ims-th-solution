<script setup lang="ts">
import { ref, onMounted } from "vue";
import { onClickOutside } from "@vueuse/core";
import { useRouter } from "vue-router";
import { useNotifications } from "../composables/useNotifications";

const router = useRouter();
const { items, unreadCount, hasUnread, loading, load, markRead, markAllRead } = useNotifications();

const panelRef = ref<HTMLElement | null>(null);
const isOpen = ref(false);

function toggle() {
  isOpen.value = !isOpen.value;
  if (isOpen.value && items.value.length === 0) {
    void load({ limit: 10 });
  }
}

function close() {
  isOpen.value = false;
}

onClickOutside(panelRef, close);

async function handleMarkRead(id: string) {
  await markRead(id);
}

async function handleMarkAllRead() {
  await markAllRead();
}

function openCenter() {
  close();
  void router.push("/notifications");
}

onMounted(() => {
  void load({ limit: 10 });
});
</script>

<template>
  <div ref="panelRef" class="notification-bell">
    <button
      class="github-header-btn notification-bell__trigger"
      type="button"
      aria-label="Notifications"
      :aria-expanded="isOpen"
      @click="toggle"
    >
      <span class="notification-bell__icon-wrap">
        <AppIcon name="notifications" size="sm" />
        <span
          v-if="hasUnread"
          class="notification-bell__badge"
          aria-hidden="true"
        >{{ unreadCount }}</span>
      </span>
    </button>

    <div
      v-if="isOpen"
      class="notification-bell__panel"
      role="dialog"
      aria-label="Notifications panel"
    >
      <div class="notification-bell__panel-header">
        <span class="notification-bell__panel-title">Notifications</span>
        <button
          v-if="hasUnread"
          class="notification-bell__mark-all"
          type="button"
          @click="handleMarkAllRead"
        >
          Mark all read
        </button>
      </div>

      <div class="notification-bell__panel-body">
        <div v-if="loading" class="notification-bell__state">
          Loading…
        </div>
        <div v-else-if="!items.length" class="notification-bell__state notification-bell__state--empty">
          No notifications
        </div>
        <ul v-else class="notification-bell__list">
          <li
            v-for="n in items"
            :key="n.notification_id ?? n.id"
            class="notification-bell__item"
            :class="{ 'is-unread': !n.is_read }"
          >
            <div class="notification-bell__item-content">
              <div class="notification-bell__item-title">{{ n.title }}</div>
              <div v-if="n.body" class="notification-bell__item-body">{{ n.body }}</div>
              <div class="notification-bell__item-meta">
                <span v-if="n.event?.category" class="notification-bell__item-tag">{{ n.event.category }}</span>
                <NuxtLink
                  v-if="n.action?.url"
                  :to="n.action.url"
                  class="notification-bell__item-link"
                  @click="close"
                >
                  {{ n.action.label ?? "Open" }}
                </NuxtLink>
              </div>
            </div>
            <button
              v-if="!n.is_read"
              class="notification-bell__read-btn"
              type="button"
              aria-label="Mark as read"
              @click="handleMarkRead(n.notification_id ?? n.id ?? '')"
            >
              <AppIcon name="check" size="xs" />
            </button>
          </li>
        </ul>
      </div>

      <div class="notification-bell__panel-footer">
        <button class="notification-bell__view-all" type="button" @click="openCenter">
          View all notifications
        </button>
      </div>
    </div>

    <div class="github-tooltip">
      <AppIcon name="notifications" size="xs" />
      <span>Notifications<template v-if="hasUnread"> ({{ unreadCount }})</template></span>
    </div>
  </div>
</template>

<style scoped>
.notification-bell {
  position: relative;
  display: flex;
  align-items: center;
}

.notification-bell__trigger {
  position: relative;
}

.notification-bell__icon-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.notification-bell__badge {
  position: absolute;
  top: -6px;
  right: -7px;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  background: var(--color-danger, #e53e3e);
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  line-height: 16px;
  border-radius: 8px;
  text-align: center;
  pointer-events: none;
}

.notification-bell__panel {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 360px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  z-index: 200;
  overflow: hidden;
}

.notification-bell__panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.notification-bell__panel-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.notification-bell__mark-all {
  font-size: 12px;
  color: var(--text-link);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

.notification-bell__mark-all:hover {
  text-decoration: underline;
}

.notification-bell__panel-body {
  max-height: 360px;
  overflow-y: auto;
}

.notification-bell__state {
  padding: 32px 16px;
  text-align: center;
  font-size: 13px;
  color: var(--text-secondary);
}

.notification-bell__list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.notification-bell__item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-subtle);
  transition: background 0.1s;
}

.notification-bell__item:last-child {
  border-bottom: none;
}

.notification-bell__item.is-unread {
  background: var(--bg-subtle);
}

.notification-bell__item-content {
  flex: 1;
  min-width: 0;
}

.notification-bell__item-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.notification-bell__item-body {
  margin-top: 2px;
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.notification-bell__item-meta {
  margin-top: 4px;
  display: flex;
  gap: 8px;
  font-size: 11px;
  color: var(--text-tertiary);
  align-items: center;
}

.notification-bell__item-tag {
  background: var(--bg-canvas);
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
  padding: 0 4px;
  font-size: 11px;
}

.notification-bell__item-link {
  color: var(--text-link);
  font-size: 12px;
}

.notification-bell__read-btn {
  flex-shrink: 0;
  padding: 4px;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-tertiary);
  border-radius: 4px;
}

.notification-bell__read-btn:hover {
  background: var(--bg-subtle);
  color: var(--text-primary);
}

.notification-bell__panel-footer {
  padding: 10px 16px;
  border-top: 1px solid var(--border-subtle);
  text-align: center;
}

.notification-bell__view-all {
  font-size: 13px;
  color: var(--text-link);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  width: 100%;
}

.notification-bell__view-all:hover {
  text-decoration: underline;
}
</style>
