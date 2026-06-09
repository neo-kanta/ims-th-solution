import { computed, ref } from "vue";
import { useRuntimeConfig } from "#imports";

import type { components } from "~/api/ims-api";
import { useAuthStore } from "~/stores/useAuthStore";

import {
  applyChatEvent,
  emptyAssistantTurn,
  findFrameEnd,
  frameTerminatorLength,
  parseFrame,
  type AssistantTurnState,
  type BindingView,
  type SourceView,
  type ToolCallView,
  type ValidationStatus,
} from "./chatStreamReducer";

// Generated transport types — the contract with the Go backend. Hand-writing
// them is forbidden per CLAUDE.md; the OpenAPI generator is the source of truth.
type SendMessageRequest = components["schemas"]["SendMessageRequest"];

// ChatMessage is the local view-model the chat components render. User turns use
// only role/content; assistant turns additionally carry the streamed tool
// trace, validation status, and provenance (sources + figure bindings).
export interface ChatMessage {
  role: "user" | "assistant";
  content: string;
  messageId?: string;
  pending?: boolean;
  toolCalls?: ToolCallView[];
  validation?: ValidationStatus;
  unverified?: string[];
  sources?: SourceView[];
  bindings?: BindingView[];
}

// useChatStream streams an assistant reply for one user turn over SSE.
//
// Native fetch() + ReadableStream is used instead of EventSource because
// EventSource is GET-only and cannot carry an Authorization header.
export function useChatStream() {
  const config = useRuntimeConfig();
  const authStore = useAuthStore();

  const sessionId = ref<string | null>(null);
  const messages = ref<ChatMessage[]>([]);
  const streaming = ref(false);
  const error = ref<string | null>(null);
  const controller = ref<AbortController | null>(null);

  async function send(content: string, modelId?: string, providerId?: string): Promise<void> {
    if (streaming.value) {
      return;
    }
    const trimmed = content.trim();
    if (!trimmed) {
      return;
    }

    error.value = null;
    streaming.value = true;
    const abort = new AbortController();
    controller.value = abort;

    messages.value.push({ role: "user", content: trimmed });
    // The assistant turn state IS the rendered message — applyChatEvent mutates
    // it in place and Vue reactivity picks up the updates.
    const assistant = emptyAssistantTurn() as AssistantTurnState & ChatMessage;
    const assistantIdx = messages.value.push(assistant) - 1;

    const baseURL = config.public.apiBaseUrl as string;
    const body: SendMessageRequest = {
      content: trimmed,
      ...(sessionId.value ? { session_id: sessionId.value } : {}),
      ...(modelId ? { model: modelId } : {}),
      ...(providerId ? { provider: providerId } : {}),
    };

    try {
      const resp = await fetch(`${baseURL}/chat`, {
        method: "POST",
        headers: {
          "content-type": "application/json",
          accept: "text/event-stream",
          ...(authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {}),
        },
        body: JSON.stringify(body),
        credentials: "omit",
        signal: abort.signal,
      });

      if (!resp.ok || !resp.body) {
        const text = await resp.text().catch(() => "");
        throw new Error(buildErrorMessage(resp.status, text));
      }

      const reader = resp.body.getReader();
      const decoder = new TextDecoder("utf-8");
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) {
          break;
        }
        buffer += decoder.decode(value, { stream: true });

        let frameEnd = findFrameEnd(buffer);
        while (frameEnd >= 0) {
          const frame = buffer.slice(0, frameEnd);
          buffer = buffer.slice(frameEnd + frameTerminatorLength(buffer, frameEnd));
          const ev = parseFrame(frame);
          if (ev) {
            applyChatEvent(assistant, ev);
            // Mirror turn-scoped signals up to the composable refs.
            if (assistant.sessionId) {
              sessionId.value = assistant.sessionId;
            }
            if (assistant.error) {
              error.value = assistant.error;
            }
          }
          frameEnd = findFrameEnd(buffer);
        }
      }
    } catch (err) {
      // A user-initiated stop (AbortController) is not an error — keep whatever
      // partial text streamed and exit quietly.
      if (err instanceof DOMException && err.name === "AbortError") {
        return;
      }
      const msg = err instanceof Error ? err.message : "Streaming error";
      error.value = msg;
      const placeholder = messages.value[assistantIdx];
      if (placeholder && !placeholder.content) {
        placeholder.content = "[error]";
      }
    } finally {
      const placeholder = messages.value[assistantIdx];
      if (placeholder) {
        placeholder.pending = false;
      }
      streaming.value = false;
      controller.value = null;
    }
  }

  // stop aborts the in-flight stream. The backend sees the client disconnect
  // (ctx cancelled) and stops generating.
  function stop(): void {
    controller.value?.abort();
  }

  // loadSession replaces the conversation with a previously-persisted session's
  // messages so the user can resume it; further sends append to that session.
  function loadSession(id: string, history: ChatMessage[]): void {
    stop();
    sessionId.value = id;
    messages.value = history;
    error.value = null;
  }

  function reset(): void {
    stop();
    sessionId.value = null;
    messages.value = [];
    error.value = null;
  }

  return {
    sessionId: computed(() => sessionId.value),
    messages: computed(() => messages.value),
    streaming: computed(() => streaming.value),
    error: computed(() => error.value),
    send,
    stop,
    loadSession,
    reset,
  };
}

function buildErrorMessage(status: number, body: string): string {
  if (status === 401) {
    return "Your session has expired. Please sign in again.";
  }
  if (status === 400 && body) {
    try {
      const parsed = JSON.parse(body) as { error?: string };
      if (parsed.error) {
        return parsed.error;
      }
    } catch {
      // fallthrough
    }
  }
  return `Chat request failed (HTTP ${status})`;
}
