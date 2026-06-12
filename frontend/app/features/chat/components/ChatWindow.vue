<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "#app";
import { useI18n } from "~/composables/useI18n";

import { sessionTitle } from "../composables/chatHistory";
import { useChatSessions } from "../composables/useChatSessions";
import { useChatStream } from "../composables/useChatStream";
import ChatEmptyState from "./ChatEmptyState.vue";
import ChatSidebar from "./ChatSidebar.vue";
import InputBar from "./InputBar.vue";
import MessageList from "./MessageList.vue";

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const chat = useChatStream();
const sessions = useChatSessions();

const scrollerRef = ref<HTMLElement | null>(null);
const composerRef = ref<InstanceType<typeof InputBar> | null>(null);
const sidebarOpen = ref(false);

const headerTitle = computed(() => {
  if (chat.messages.value.length === 0) {
    return t("chat.title");
  }
  const sid = chat.sessionId.value;
  if (sid && sessions.titles[sid]) {
    return sessions.titles[sid];
  }
  const firstUser = chat.messages.value.find((m) => m.role === "user");
  return firstUser ? sessionTitle(firstUser.content, t("chat.title")) : t("chat.title");
});

function scrollToBottom() {
  const el = scrollerRef.value;
  if (el) {
    el.scrollTo({ top: el.scrollHeight, behavior: "smooth" });
  }
}

// Re-scroll as messages arrive and as the streaming answer grows.
watch(
  () => {
    const msgs = chat.messages.value;
    const last = msgs[msgs.length - 1];
    return `${msgs.length}:${last?.content.length ?? 0}`;
  },
  () => {
    void nextTick(scrollToBottom);
  },
);

async function onSubmit(text: string) {
  const wasNew = !chat.sessionId.value;
  await chat.send(text);
  const sid = chat.sessionId.value;
  if (sid) {
    sessions.setTitleFromText(sid, text);
    if (wasNew) {
      // A brand-new session was created — refresh the rail so it appears.
      await sessions.fetchSessions();
    }
  }
}

function onSuggest(text: string) {
  void onSubmit(text);
}

async function onSelectSession(id: string) {
  sidebarOpen.value = false;
  if (id === chat.sessionId.value) {
    return;
  }
  try {
    const history = await sessions.loadSessionMessages(id);
    chat.loadSession(id, history);
    void nextTick(scrollToBottom);
  } catch {
    // loadSessionMessages failure surfaces via the rail's own error path on
    // next fetch; keep the current conversation intact.
  }
}

function onNewChat() {
  chat.reset();
  sidebarOpen.value = false;
  void nextTick(() => composerRef.value?.focus());
}

function onStop() {
  chat.stop();
}

onMounted(async () => {
  await sessions.fetchSessions();

  const initialQuery = route.query.q as string | undefined;
  const initialModel = route.query.model as string | undefined;
  const initialProvider = route.query.provider as string | undefined;
  if (initialQuery) {
    await router.replace({ query: {} }); // clear so refresh doesn't resend
    const wasNew = !chat.sessionId.value;
    await chat.send(initialQuery, initialModel, initialProvider);
    const sid = chat.sessionId.value;
    if (sid) {
      sessions.setTitleFromText(sid, initialQuery);
      if (wasNew) {
        await sessions.fetchSessions();
      }
    }
  }
});
</script>

<template>
  <section class="chat">
    <div v-if="sidebarOpen" class="chat__backdrop" @click="sidebarOpen = false" />

    <ChatSidebar
      class="chat__rail"
      :class="{ 'is-open': sidebarOpen }"
      :sessions="sessions.sessions.value"
      :titles="sessions.titles"
      :loading="sessions.loading.value"
      :error="sessions.error.value"
      :active-id="chat.sessionId.value"
      @select="onSelectSession"
      @new-chat="onNewChat"
      @retry="sessions.fetchSessions()"
      @close="sidebarOpen = false"
    />

    <div class="chat__main">
      <header class="chat__header">
        <button
          type="button"
          class="chat__hamburger"
          :aria-label="t('chat.history.open')"
          @click="sidebarOpen = true"
        >
          ☰
        </button>
        <div class="chat__heading">
          <h1 class="chat__title">{{ headerTitle }}</h1>
          <p class="chat__subtitle">{{ t("chat.subtitle") }}</p>
        </div>
        <button type="button" class="chat__newbtn" @click="onNewChat">
          <span aria-hidden="true">＋</span>
          <span class="chat__newbtn-label">{{ t("chat.newChat") }}</span>
        </button>
      </header>

      <div ref="scrollerRef" class="chat__scroller">
        <ChatEmptyState v-if="chat.messages.value.length === 0" @suggest="onSuggest" />
        <MessageList v-else :messages="chat.messages.value" :streaming="chat.streaming.value" />
      </div>

      <div v-if="chat.error.value" class="chat__error" role="alert">
        <strong>{{ t("chat.errorPrefix") }}:</strong> {{ chat.error.value }}
      </div>

      <InputBar
        ref="composerRef"
        :disabled="chat.streaming.value"
        :streaming="chat.streaming.value"
        @submit="onSubmit"
        @stop="onStop"
      />
    </div>
  </section>
</template>

<style scoped>
.chat {
  /* Alias the chat components' token names to the app design system so the
     whole feature follows light/dark mode. Custom properties inherit to every
     descendant component (ToolTrace, SourcesPanel, sidebar, etc.). */
  /* Point only at SEMANTIC tokens — the dark theme redefines these (it does
     not redefine the raw --color-* scales), so the chat adapts to light/dark. */
  --bg-elevated: var(--bg-card);
  --bg-canvas: var(--bg-page);
  --bg-hover: var(--bg-row-hover);
  --bg-accent: var(--action-primary);
  --bg-accent-subtle: var(--alert-info-bg);
  --bg-accent-subtle-hover: var(--bg-selected);
  --text-accent: var(--alert-info-text);
  --border-input: var(--border-default);
  --border-accent: var(--alert-info-border);
  --ring-focus: var(--focus-ring);
  --text-danger: var(--alert-danger-text);
  --bg-danger-subtle: var(--alert-danger-bg);
  --font-mono: var(--font-family-mono);
  --ok-text: var(--alert-success-text);
  --ok-border: var(--alert-success-border);
  --warn-text: var(--alert-warning-text);
  --warn-border: var(--alert-warning-border);
  --skeleton-base: var(--bg-row-hover);
  --skeleton-shine: var(--border-subtle);

  position: relative;
  display: grid;
  grid-template-columns: 288px minmax(0, 1fr);
  height: calc(100vh - 180px);
  min-height: 540px;
  border: 1px solid var(--border-subtle, #e2e8f0);
  border-radius: 14px;
  background: var(--bg-elevated, #ffffff);
  overflow: hidden;
}
.chat__rail {
  min-width: 0;
}
.chat__main {
  display: flex;
  flex-direction: column;
  min-width: 0;
  height: 100%;
  min-height: 0;
}
.chat__header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 18px;
  border-bottom: 1px solid var(--border-subtle, #e2e8f0);
}
.chat__heading {
  flex: 1;
  min-width: 0;
}
.chat__title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary, #0f172a);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.chat__subtitle {
  margin: 2px 0 0;
  font-size: 12px;
  color: var(--text-secondary, #475569);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.chat__newbtn {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 34px;
  padding: 0 12px;
  border-radius: 9px;
  border: 1px solid var(--border-subtle, #e2e8f0);
  background: var(--bg-canvas, #f8fafc);
  color: var(--text-primary, #0f172a);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}
.chat__newbtn:hover {
  border-color: var(--border-accent, #c7d2fe);
}
.chat__hamburger {
  display: none;
  width: 36px;
  height: 36px;
  border-radius: 9px;
  border: 1px solid var(--border-subtle, #e2e8f0);
  background: var(--bg-canvas, #f8fafc);
  color: var(--text-secondary, #475569);
  font-size: 16px;
  cursor: pointer;
}
.chat__scroller {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  background: var(--bg-canvas, #f8fafc);
}
.chat__error {
  padding: 10px 18px;
  font-size: 13px;
  color: var(--text-danger, #b91c1c);
  background: var(--bg-danger-subtle, #fef2f2);
  border-top: 1px solid var(--border-danger, #fecaca);
}
.chat__backdrop {
  display: none;
}

@media (max-width: 860px) {
  .chat {
    grid-template-columns: minmax(0, 1fr);
  }
  .chat__hamburger {
    display: inline-grid;
    place-items: center;
  }
  .chat__newbtn-label {
    display: none;
  }
  .chat__rail {
    position: absolute;
    z-index: 30;
    top: 0;
    bottom: 0;
    left: 0;
    width: 300px;
    max-width: 85%;
    transform: translateX(-100%);
    transition: transform 0.2s ease;
    box-shadow: 0 10px 30px rgba(15, 23, 42, 0.18);
  }
  .chat__rail.is-open {
    transform: translateX(0);
  }
  .chat__backdrop {
    display: block;
    position: absolute;
    inset: 0;
    z-index: 20;
    background: rgba(15, 23, 42, 0.35);
  }
}
</style>
