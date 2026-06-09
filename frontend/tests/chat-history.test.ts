import { describe, expect, it } from "vitest";

import {
  dayBucket,
  firstUserText,
  mapHistory,
  mapHistoryMessage,
  relativeTime,
  sessionTitle,
} from "../app/features/chat/composables/chatHistory";
import type { ChatMessageResponse } from "../app/features/chat/services/chatSessionsApi";

describe("history message mapping", () => {
  it("maps a user message", () => {
    const m: ChatMessageResponse = { id: "m1", role: "user", content: "hi" };
    const mapped = mapHistoryMessage(m);
    expect(mapped).toEqual({ role: "user", content: "hi", messageId: "m1", pending: false });
  });

  it("reconstructs assistant provenance into sources/bindings/validation", () => {
    const m: ChatMessageResponse = {
      id: "m2",
      role: "assistant",
      content: "The NAV is 10.25.",
      provenance: {
        validation: {
          decision: "allow",
          bindings: [{ figure: "10.25", source: "tool_raw", tool_name: "get_fund_nav" }],
        },
        tool_calls: [{ tool_name: "get_fund_nav", mcp_server: "ims", state: "succeeded" }],
      } as unknown as Record<string, never>,
    };
    const mapped = mapHistoryMessage(m);
    expect(mapped?.validation).toBe("passed");
    expect(mapped?.sources).toEqual([{ toolName: "get_fund_nav", server: "ims", state: "succeeded" }]);
    expect(mapped?.bindings).toEqual([{ figure: "10.25", source: "tool_raw", toolName: "get_fund_nav" }]);
  });

  it("maps a blocked decision to blocked validation", () => {
    const m: ChatMessageResponse = {
      id: "m3",
      role: "assistant",
      content: "blocked",
      provenance: { validation: { decision: "block" } } as unknown as Record<string, never>,
    };
    expect(mapHistoryMessage(m)?.validation).toBe("blocked");
  });

  it("drops non user/assistant roles", () => {
    expect(mapHistoryMessage({ id: "x", role: "tool", content: "{}" })).toBeNull();
  });

  it("mapHistory filters dropped messages", () => {
    const rows: ChatMessageResponse[] = [
      { id: "1", role: "user", content: "a" },
      { id: "2", role: "tool", content: "{}" },
      { id: "3", role: "assistant", content: "b" },
    ];
    expect(mapHistory(rows)).toHaveLength(2);
  });
});

describe("session title + first user text", () => {
  it("returns the first user message content", () => {
    const rows: ChatMessageResponse[] = [
      { id: "1", role: "assistant", content: "hello" },
      { id: "2", role: "user", content: "my question" },
    ];
    expect(firstUserText(rows)).toBe("my question");
  });

  it("truncates long titles and falls back when empty", () => {
    expect(sessionTitle("  short  ", "fallback")).toBe("short");
    expect(sessionTitle("", "fallback")).toBe("fallback");
    const long = "x".repeat(80);
    const title = sessionTitle(long, "fallback");
    expect(title.length).toBe(58); // 57 chars + ellipsis
    expect(title.endsWith("…")).toBe(true);
  });
});

describe("relative time + day bucket", () => {
  const now = new Date(2026, 5, 8, 12, 0, 0);

  it("formats relative times", () => {
    expect(relativeTime(new Date(2026, 5, 8, 11, 59, 40).toISOString(), now)).toBe("just now");
    expect(relativeTime(new Date(2026, 5, 8, 11, 55, 0).toISOString(), now)).toBe("5m ago");
    expect(relativeTime(new Date(2026, 5, 8, 9, 0, 0).toISOString(), now)).toBe("3h ago");
    expect(relativeTime(new Date(2026, 5, 6, 12, 0, 0).toISOString(), now)).toBe("2d ago");
    expect(relativeTime(undefined, now)).toBe("");
  });

  it("buckets timestamps by day", () => {
    expect(dayBucket(new Date(2026, 5, 8, 9, 0, 0).toISOString(), now)).toBe("today");
    expect(dayBucket(new Date(2026, 5, 7, 9, 0, 0).toISOString(), now)).toBe("yesterday");
    expect(dayBucket(new Date(2026, 5, 1, 9, 0, 0).toISOString(), now)).toBe("earlier");
  });
});
