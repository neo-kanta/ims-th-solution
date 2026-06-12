// Pure, framework-free logic for the chat SSE stream: frame parsing, the
// per-turn event reducer, and source/binding mapping. Keeping this Vue-free
// makes it unit-testable without a DOM (see frontend/tests/chat-stream.test.ts).
//
// Transport types come from the generated OpenAPI client — never hand-written.

import type { components } from "~/api/ims-api";

export type ChatStreamEvent = components["schemas"]["ChatStreamEvent"];

export type ToolStatus = "running" | "ok" | "error" | "denied";
export type ValidationStatus = "idle" | "running" | "passed" | "blocked";

export interface ToolCallView {
  name: string;
  status: ToolStatus;
}

export interface SourceView {
  toolName: string;
  server?: string;
  state: string;
  asOf?: string;
}

export interface BindingView {
  figure: string;
  source: string;
  toolName?: string;
}

// AssistantTurnState is the accumulated state of one streamed assistant turn.
export interface AssistantTurnState {
  role: "assistant";
  content: string;
  pending: boolean;
  toolCalls: ToolCallView[];
  validation: ValidationStatus;
  /** Figures that failed numeric validation (present when validation = blocked). */
  unverified: string[];
  sources: SourceView[];
  bindings: BindingView[];
  sessionId?: string;
  messageId?: string;
  error?: string;
  done: boolean;
}

export function emptyAssistantTurn(): AssistantTurnState {
  return {
    role: "assistant",
    content: "",
    pending: true,
    toolCalls: [],
    validation: "idle",
    unverified: [],
    sources: [],
    bindings: [],
    done: false,
  };
}

// applyChatEvent folds one SSE event into the turn state. Pure w.r.t. inputs
// (mutates and returns the passed state for convenience/reactivity).
export function applyChatEvent(state: AssistantTurnState, ev: ChatStreamEvent): AssistantTurnState {
  switch (ev.kind) {
    case "session_started":
      if (ev.session_id) {
        state.sessionId = ev.session_id;
      }
      break;
    case "text":
      if (ev.text) {
        state.content += ev.text;
      }
      break;
    case "tool_call":
      applyToolCall(state, ev.tool_name, ev.tool_status);
      break;
    case "validation":
      state.validation = normaliseValidation(ev.tool_status);
      state.unverified = ev.unverified ?? [];
      break;
    case "sources":
      state.sources = mapSources(ev.sources);
      state.bindings = mapBindings(ev.bindings);
      break;
    case "done":
      if (ev.message_id) {
        state.messageId = ev.message_id;
      }
      state.done = true;
      break;
    case "error":
      state.error = ev.error || "Unknown error";
      break;
    default:
      // Unknown/future kinds: ignore gracefully.
      break;
  }
  return state;
}

function applyToolCall(state: AssistantTurnState, name?: string, status?: string): void {
  if (!name || !status) {
    return;
  }
  if (status === "running") {
    state.toolCalls.push({ name, status: "running" });
    return;
  }
  const terminal = normaliseToolStatus(status);
  // Update the most recent still-running call for this tool.
  for (let i = state.toolCalls.length - 1; i >= 0; i--) {
    const tc = state.toolCalls[i];
    if (tc && tc.name === name && tc.status === "running") {
      tc.status = terminal;
      return;
    }
  }
  state.toolCalls.push({ name, status: terminal });
}

function normaliseToolStatus(s?: string): ToolStatus {
  if (s === "ok" || s === "error" || s === "denied" || s === "running") {
    return s;
  }
  return "error";
}

function normaliseValidation(s?: string): ValidationStatus {
  if (s === "running" || s === "passed" || s === "blocked") {
    return s;
  }
  return "idle";
}

export function mapSources(sources?: components["schemas"]["ChatSourceRef"][]): SourceView[] {
  if (!sources) {
    return [];
  }
  return sources.map((s) => ({
    toolName: s.tool_name ?? "",
    server: s.server,
    state: s.state ?? "",
    asOf: s.as_of,
  }));
}

export function mapBindings(bindings?: components["schemas"]["ChatFigureBinding"][]): BindingView[] {
  if (!bindings) {
    return [];
  }
  return bindings.map((b) => ({
    figure: b.figure ?? "",
    source: b.source ?? "",
    toolName: b.tool_name,
  }));
}

// ── SSE frame parsing ───────────────────────────────────────────────────────

// findFrameEnd returns the index of the first SSE frame terminator ("\n\n" or
// "\r\n\r\n") in buf, or -1 if none.
export function findFrameEnd(buf: string): number {
  const lf = buf.indexOf("\n\n");
  const crlf = buf.indexOf("\r\n\r\n");
  if (lf < 0) {
    return crlf;
  }
  if (crlf < 0) {
    return lf;
  }
  return Math.min(lf, crlf);
}

// frameTerminatorLength returns the length of the terminator found at index
// (2 for "\n\n", 4 for "\r\n\r\n") so the caller can advance the buffer.
export function frameTerminatorLength(buf: string, frameEnd: number): number {
  return buf.startsWith("\r\n\r\n", frameEnd) ? 4 : 2;
}

// parseFrame extracts the JSON ChatStreamEvent from one SSE frame's data lines.
export function parseFrame(frame: string): ChatStreamEvent | null {
  const dataLines: string[] = [];
  for (const rawLine of frame.split(/\r?\n/)) {
    if (rawLine.startsWith(":")) {
      continue; // comment
    }
    if (rawLine.startsWith("data:")) {
      const payload = rawLine.slice("data:".length);
      dataLines.push(payload.startsWith(" ") ? payload.slice(1) : payload);
    }
    // event: lines are informational; the trusted discriminator is `kind`.
  }
  if (dataLines.length === 0) {
    return null;
  }
  try {
    return JSON.parse(dataLines.join("\n")) as ChatStreamEvent;
  } catch {
    return null;
  }
}
