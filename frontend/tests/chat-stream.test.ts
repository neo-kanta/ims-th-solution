import { describe, expect, it } from "vitest";

import {
  applyChatEvent,
  emptyAssistantTurn,
  findFrameEnd,
  frameTerminatorLength,
  mapBindings,
  mapSources,
  parseFrame,
  type ChatStreamEvent,
} from "../app/features/chat/composables/chatStreamReducer";

describe("chat SSE reducer", () => {
  it("appends streamed text", () => {
    const s = emptyAssistantTurn();
    applyChatEvent(s, { kind: "text", text: "Hello " } as ChatStreamEvent);
    applyChatEvent(s, { kind: "text", text: "world" } as ChatStreamEvent);
    expect(s.content).toBe("Hello world");
  });

  it("captures session id and final message id", () => {
    const s = emptyAssistantTurn();
    applyChatEvent(s, { kind: "session_started", session_id: "sess-1" } as ChatStreamEvent);
    applyChatEvent(s, { kind: "done", message_id: "msg-1" } as ChatStreamEvent);
    expect(s.sessionId).toBe("sess-1");
    expect(s.messageId).toBe("msg-1");
    expect(s.done).toBe(true);
  });

  it("reduces a tool_call running -> ok into one trace entry", () => {
    const s = emptyAssistantTurn();
    applyChatEvent(s, { kind: "tool_call", tool_name: "get_fund_nav", tool_status: "running" } as ChatStreamEvent);
    applyChatEvent(s, { kind: "tool_call", tool_name: "get_fund_nav", tool_status: "ok" } as ChatStreamEvent);
    expect(s.toolCalls).toHaveLength(1);
    expect(s.toolCalls[0]).toEqual({ name: "get_fund_nav", status: "ok" });
  });

  it("tracks multiple tools including a denied one", () => {
    const s = emptyAssistantTurn();
    applyChatEvent(s, { kind: "tool_call", tool_name: "list_portfolios", tool_status: "running" } as ChatStreamEvent);
    applyChatEvent(s, { kind: "tool_call", tool_name: "list_portfolios", tool_status: "ok" } as ChatStreamEvent);
    applyChatEvent(s, { kind: "tool_call", tool_name: "get_portfolio_holdings", tool_status: "running" } as ChatStreamEvent);
    applyChatEvent(s, { kind: "tool_call", tool_name: "get_portfolio_holdings", tool_status: "denied" } as ChatStreamEvent);
    expect(s.toolCalls).toEqual([
      { name: "list_portfolios", status: "ok" },
      { name: "get_portfolio_holdings", status: "denied" },
    ]);
  });

  it("records validation status", () => {
    const s = emptyAssistantTurn();
    applyChatEvent(s, { kind: "validation", tool_status: "running" } as ChatStreamEvent);
    expect(s.validation).toBe("running");
    applyChatEvent(s, { kind: "validation", tool_status: "blocked" } as ChatStreamEvent);
    expect(s.validation).toBe("blocked");
  });

  it("captures unverified figures on a blocked validation", () => {
    const s = emptyAssistantTurn();
    applyChatEvent(s, {
      kind: "validation",
      tool_status: "blocked",
      unverified: ["10.99", "1,234.50"],
    } as ChatStreamEvent);
    expect(s.validation).toBe("blocked");
    expect(s.unverified).toEqual(["10.99", "1,234.50"]);
  });

  it("maps sources and figure bindings from a sources event", () => {
    const s = emptyAssistantTurn();
    applyChatEvent(s, {
      kind: "sources",
      sources: [{ tool_name: "get_fund_nav", server: "ims", state: "succeeded", as_of: "2026-06-07T00:00:00Z" }],
      bindings: [{ figure: "10.25", source: "tool_raw", tool_name: "get_fund_nav" }],
    } as ChatStreamEvent);
    expect(s.sources).toEqual([
      { toolName: "get_fund_nav", server: "ims", state: "succeeded", asOf: "2026-06-07T00:00:00Z" },
    ]);
    expect(s.bindings).toEqual([{ figure: "10.25", source: "tool_raw", toolName: "get_fund_nav" }]);
  });

  it("records a stream error", () => {
    const s = emptyAssistantTurn();
    applyChatEvent(s, { kind: "error", error: "the assistant could not complete this turn" } as ChatStreamEvent);
    expect(s.error).toBe("the assistant could not complete this turn");
  });

  it("ignores unknown event kinds gracefully", () => {
    const s = emptyAssistantTurn();
    const before = JSON.stringify(s);
    applyChatEvent(s, { kind: "future_kind" } as unknown as ChatStreamEvent);
    expect(JSON.stringify(s)).toBe(before);
  });
});

describe("source/binding mappers", () => {
  it("return empty arrays for undefined input", () => {
    expect(mapSources(undefined)).toEqual([]);
    expect(mapBindings(undefined)).toEqual([]);
  });
});

describe("SSE frame parsing", () => {
  it("parses a data line into a ChatStreamEvent", () => {
    const frame = "event: text\ndata: {\"kind\":\"text\",\"text\":\"hi\"}";
    const ev = parseFrame(frame);
    expect(ev?.kind).toBe("text");
    expect(ev?.text).toBe("hi");
  });

  it("ignores comment lines and returns null when no data", () => {
    expect(parseFrame(": keep-alive")).toBeNull();
  });

  it("returns null on malformed JSON", () => {
    expect(parseFrame("data: {not json")).toBeNull();
  });

  it("finds LF and CRLF frame terminators with correct lengths", () => {
    const lf = "data: a\n\ndata: b";
    expect(findFrameEnd(lf)).toBe(7);
    expect(frameTerminatorLength(lf, findFrameEnd(lf))).toBe(2);

    const crlf = "data: a\r\n\r\ndata: b";
    const idx = findFrameEnd(crlf);
    expect(crlf.startsWith("\r\n\r\n", idx)).toBe(true);
    expect(frameTerminatorLength(crlf, idx)).toBe(4);
  });
});
