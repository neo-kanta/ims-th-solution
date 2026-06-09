import { computed, reactive, ref } from "vue";

import { chatSessionsApi, type SessionSummary } from "../services/chatSessionsApi";
import { firstUserText, mapHistory, sessionTitle } from "./chatHistory";
import type { ChatMessage } from "./useChatStream";

const SESSION_PAGE_LIMIT = 30;
const MESSAGE_PAGE_LIMIT = 200;

// useChatSessions manages the session-history rail: the list, lazily-resolved
// titles (from each session's first user message), and loading a session's
// transcript for resume.
export function useChatSessions() {
  const sessions = ref<SessionSummary[]>([]);
  const titles = reactive<Record<string, string>>({});
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchSessions(): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      const res = await chatSessionsApi.list({ page: 1, limit: SESSION_PAGE_LIMIT });
      sessions.value = res.items ?? [];
      void hydrateTitles(sessions.value);
    } catch (e) {
      error.value = e instanceof Error ? e.message : "Failed to load conversations";
    } finally {
      loading.value = false;
    }
  }

  // hydrateTitles resolves a human title for each session from its first user
  // message. Best-effort and non-blocking; failures fall back to time/model.
  async function hydrateTitles(list: SessionSummary[]): Promise<void> {
    await Promise.all(
      list.map(async (s) => {
        if (!s.id || titles[s.id]) {
          return;
        }
        try {
          const res = await chatSessionsApi.messages(s.id, { page: 1, limit: 1 });
          const title = sessionTitle(firstUserText(res.items), "");
          if (title) {
            titles[s.id] = title;
          }
        } catch {
          // ignore — UI falls back to relative time + model
        }
      }),
    );
  }

  // loadSessionMessages fetches and maps a session's transcript for resume.
  async function loadSessionMessages(sessionId: string): Promise<ChatMessage[]> {
    const res = await chatSessionsApi.messages(sessionId, { page: 1, limit: MESSAGE_PAGE_LIMIT });
    if (!titles[sessionId]) {
      const title = sessionTitle(firstUserText(res.items), "");
      if (title) {
        titles[sessionId] = title;
      }
    }
    return mapHistory(res.items);
  }

  // setTitleFromText seeds a title locally (e.g. the first message of a brand
  // new session) so the rail shows it immediately without a round-trip.
  function setTitleFromText(sessionId: string, text: string): void {
    if (sessionId && !titles[sessionId]) {
      const title = sessionTitle(text, "");
      if (title) {
        titles[sessionId] = title;
      }
    }
  }

  return {
    sessions: computed(() => sessions.value),
    titles,
    loading: computed(() => loading.value),
    error: computed(() => error.value),
    fetchSessions,
    loadSessionMessages,
    setTitleFromText,
  };
}
