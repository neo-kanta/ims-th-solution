// Pure helpers for rendering persisted chat history. Maps the backend's
// ChatMessageResponse (with its provenance map) back into the rich view-model
// the chat components render, and derives session titles / relative times.
// Framework-free so it is unit-testable (see frontend/tests/chat-history.test.ts).

import type { ChatMessageResponse } from "../services/chatSessionsApi";
import type { BindingView, SourceView, ValidationStatus } from "./chatStreamReducer";
import type { ChatMessage } from "./useChatStream";

// Shape of the provenance map persisted by the backend (safe subset).
interface ProvenanceShape {
  validation?: {
    decision?: string;
    bindings?: { figure?: string; source?: string; tool_name?: string }[];
    violations?: { figure?: string }[];
  };
  tool_calls?: { tool_name?: string; mcp_server?: string; state?: string }[];
}

// mapHistoryMessage converts one persisted message into a ChatMessage view-model.
// Assistant turns reconstruct their tool sources, figure bindings and validation
// status from the stored provenance map. Returns null for non-user/assistant.
export function mapHistoryMessage(m: ChatMessageResponse): ChatMessage | null {
  const role = m.role === "user" || m.role === "assistant" ? m.role : null;
  if (!role) {
    return null;
  }
  const base: ChatMessage = {
    role,
    content: m.content ?? "",
    messageId: m.id,
    pending: false,
  };
  if (role !== "assistant") {
    return base;
  }

  const prov = (m.provenance ?? {}) as ProvenanceShape;
  base.validation = decisionToValidation(prov.validation?.decision);
  base.sources = (prov.tool_calls ?? []).map(
    (tc): SourceView => ({
      toolName: tc.tool_name ?? "",
      server: tc.mcp_server,
      state: tc.state ?? "",
    }),
  );
  base.bindings = (prov.validation?.bindings ?? []).map(
    (b): BindingView => ({
      figure: b.figure ?? "",
      source: b.source ?? "",
      toolName: b.tool_name,
    }),
  );
  base.unverified = (prov.validation?.violations ?? [])
    .map((v) => v.figure ?? "")
    .filter((f) => f.length > 0);
  return base;
}

export function mapHistory(messages: ChatMessageResponse[] | undefined): ChatMessage[] {
  if (!messages) {
    return [];
  }
  const out: ChatMessage[] = [];
  for (const m of messages) {
    const mapped = mapHistoryMessage(m);
    if (mapped) {
      out.push(mapped);
    }
  }
  return out;
}

function decisionToValidation(decision?: string): ValidationStatus {
  if (decision === "allow") {
    return "passed";
  }
  if (decision === "block") {
    return "blocked";
  }
  return "idle";
}

// firstUserText returns the first user message's content (for a session title).
export function firstUserText(messages: ChatMessageResponse[] | undefined): string | undefined {
  return messages?.find((m) => m.role === "user")?.content ?? undefined;
}

// sessionTitle derives a clean, truncated title from the first user message.
export function sessionTitle(firstUserMessage: string | undefined, fallback: string): string {
  const trimmed = (firstUserMessage ?? "").trim().replace(/\s+/g, " ");
  if (!trimmed) {
    return fallback;
  }
  return trimmed.length > 60 ? `${trimmed.slice(0, 57)}…` : trimmed;
}

// relativeTime formats an ISO timestamp as a compact relative label.
export function relativeTime(iso: string | undefined, now: Date = new Date()): string {
  if (!iso) {
    return "";
  }
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) {
    return "";
  }
  const sec = Math.round((now.getTime() - then) / 1000);
  if (sec < 45) {
    return "just now";
  }
  const min = Math.round(sec / 60);
  if (min < 60) {
    return `${min}m ago`;
  }
  const hr = Math.round(min / 60);
  if (hr < 24) {
    return `${hr}h ago`;
  }
  const day = Math.round(hr / 24);
  if (day < 7) {
    return `${day}d ago`;
  }
  const wk = Math.round(day / 7);
  if (wk < 5) {
    return `${wk}w ago`;
  }
  return new Date(iso).toLocaleDateString();
}

// dayBucket groups a timestamp for the sidebar (today / yesterday / earlier).
export type DayBucket = "today" | "yesterday" | "earlier";

export function dayBucket(iso: string | undefined, now: Date = new Date()): DayBucket {
  if (!iso) {
    return "earlier";
  }
  const d = new Date(iso).getTime();
  if (Number.isNaN(d)) {
    return "earlier";
  }
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
  if (d >= startOfToday) {
    return "today";
  }
  if (d >= startOfToday - 86_400_000) {
    return "yesterday";
  }
  return "earlier";
}
