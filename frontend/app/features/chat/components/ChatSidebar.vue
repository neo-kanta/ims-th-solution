<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "~/composables/useI18n";

import { dayBucket, relativeTime, type DayBucket } from "../composables/chatHistory";
import type { SessionSummary } from "../services/chatSessionsApi";

const props = defineProps<{
  sessions: readonly SessionSummary[];
  titles: Record<string, string>;
  loading: boolean;
  error: string | null;
  activeId: string | null;
}>();

const emit = defineEmits<{
  select: [id: string];
  newChat: [];
  retry: [];
  close: [];
}>();

const { t } = useI18n();

interface Group {
  key: DayBucket;
  label: string;
  items: SessionSummary[];
}

const groups = computed<Group[]>(() => {
  const buckets: Record<DayBucket, SessionSummary[]> = { today: [], yesterday: [], earlier: [] };
  for (const s of props.sessions) {
    buckets[dayBucket(s.updated_at)].push(s);
  }
  const labelFor: Record<DayBucket, string> = {
    today: t("chat.history.today"),
    yesterday: t("chat.history.yesterday"),
    earlier: t("chat.history.earlier"),
  };
  return (["today", "yesterday", "earlier"] as DayBucket[])
    .filter((k) => buckets[k].length > 0)
    .map((k) => ({ key: k, label: labelFor[k], items: buckets[k] }));
});

const isEmpty = computed(() => !props.loading && !props.error && props.sessions.length === 0);

function titleFor(s: SessionSummary): string {
  return (s.id && props.titles[s.id]) || t("chat.untitled");
}
</script>

<template>
  <aside class="chat-rail">
    <div class="chat-rail__top">
      <button type="button" class="chat-rail__new" @click="emit('newChat')">
        <span aria-hidden="true">＋</span> {{ t("chat.newChat") }}
      </button>
      <button
        type="button"
        class="chat-rail__close"
        :aria-label="t('chat.history.close')"
        @click="emit('close')"
      >
        ✕
      </button>
    </div>

    <div class="chat-rail__body">
      <!-- Loading skeletons -->
      <div v-if="loading" class="chat-rail__skeletons">
        <div v-for="n in 6" :key="n" class="chat-rail__skeleton" />
      </div>

      <!-- Error -->
      <div v-else-if="error" class="chat-rail__state">
        <p class="chat-rail__state-text">{{ error }}</p>
        <button type="button" class="chat-rail__retry" @click="emit('retry')">
          {{ t("chat.history.retry") }}
        </button>
      </div>

      <!-- Empty -->
      <div v-else-if="isEmpty" class="chat-rail__state">
        <p class="chat-rail__state-text">{{ t("chat.history.empty") }}</p>
      </div>

      <!-- Grouped sessions -->
      <template v-else>
        <section v-for="group in groups" :key="group.key" class="chat-rail__group">
          <h3 class="chat-rail__group-title">{{ group.label }}</h3>
          <ul class="chat-rail__list">
            <li v-for="s in group.items" :key="s.id">
              <button
                type="button"
                class="chat-rail__item"
                :class="{ 'is-active': s.id === activeId }"
                @click="s.id && emit('select', s.id)"
              >
                <span class="chat-rail__item-title">{{ titleFor(s) }}</span>
                <span class="chat-rail__item-meta">
                  <span>{{ relativeTime(s.updated_at) }}</span>
                  <span v-if="s.model" class="chat-rail__model">{{ s.model }}</span>
                </span>
              </button>
            </li>
          </ul>
        </section>
      </template>
    </div>

    <div class="chat-rail__footer">
      <span class="chat-rail__badge">{{ t("chat.sliceLabel") }}</span>
    </div>
  </aside>
</template>

<style scoped>
.chat-rail {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-canvas, #f8fafc);
  border-right: 1px solid var(--border-subtle, #e2e8f0);
}
.chat-rail__top {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  border-bottom: 1px solid var(--border-subtle, #e2e8f0);
}
.chat-rail__new {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 38px;
  border-radius: 10px;
  border: 1px solid var(--border-accent, #c7d2fe);
  background: var(--bg-accent-subtle, #eef2ff);
  color: var(--text-accent, #3730a3);
  font-weight: 600;
  font-size: 13px;
  cursor: pointer;
  transition: background 0.15s ease;
}
.chat-rail__new:hover {
  background: var(--bg-accent-subtle-hover, #e0e7ff);
}
.chat-rail__close {
  display: none;
  width: 38px;
  height: 38px;
  border-radius: 10px;
  border: 1px solid var(--border-subtle, #e2e8f0);
  background: var(--bg-elevated, #ffffff);
  color: var(--text-secondary, #475569);
  cursor: pointer;
}
.chat-rail__body {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}
.chat-rail__group {
  margin-bottom: 12px;
}
.chat-rail__group-title {
  margin: 8px 8px 6px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-tertiary, #64748b);
}
.chat-rail__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.chat-rail__item {
  width: 100%;
  text-align: left;
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: 9px 10px;
  border-radius: 9px;
  border: 1px solid transparent;
  background: transparent;
  cursor: pointer;
  transition: background 0.12s ease;
}
.chat-rail__item:hover {
  background: var(--bg-hover, #eef2f7);
}
.chat-rail__item.is-active {
  background: var(--bg-accent-subtle, #eef2ff);
  border-color: var(--border-accent, #c7d2fe);
}
.chat-rail__item-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary, #0f172a);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.chat-rail__item-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  color: var(--text-tertiary, #64748b);
}
.chat-rail__model {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  opacity: 0.8;
}
.chat-rail__skeletons {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px;
}
.chat-rail__skeleton {
  height: 44px;
  border-radius: 9px;
  background: linear-gradient(
    90deg,
    var(--skeleton-base, #eef2f7) 25%,
    var(--skeleton-shine, #e2e8f0) 37%,
    var(--skeleton-base, #eef2f7) 63%
  );
  background-size: 400% 100%;
  animation: chat-rail-shimmer 1.4s ease infinite;
}
@keyframes chat-rail-shimmer {
  0% {
    background-position: 100% 50%;
  }
  100% {
    background-position: 0 50%;
  }
}
.chat-rail__state {
  padding: 24px 16px;
  text-align: center;
}
.chat-rail__state-text {
  font-size: 13px;
  color: var(--text-secondary, #475569);
  margin: 0 0 10px;
}
.chat-rail__retry {
  padding: 6px 14px;
  border-radius: 8px;
  border: 1px solid var(--border-subtle, #e2e8f0);
  background: var(--bg-elevated, #ffffff);
  color: var(--text-primary, #0f172a);
  font-size: 12px;
  cursor: pointer;
}
.chat-rail__footer {
  padding: 10px 14px;
  border-top: 1px solid var(--border-subtle, #e2e8f0);
}
.chat-rail__badge {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--text-accent, #3730a3);
}

@media (max-width: 860px) {
  .chat-rail__close {
    display: inline-grid;
    place-items: center;
  }
}
</style>
