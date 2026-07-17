import { describe, expect, it } from "vitest";

import {
  collectNumberedPages,
  collectOffsetPages,
} from "../app/features/compliance/lib/pagination";

describe("collectOffsetPages", () => {
  it("collects every page before returning an exact directory", async () => {
    const calls: number[] = [];
    const result = await collectOffsetPages(async (offset, limit) => {
      calls.push(offset);
      const all = ["a", "b", "c", "d", "e"];
      return {
        items: all.slice(offset, offset + limit),
        total: all.length,
        offset,
      };
    }, 2);

    expect(calls).toEqual([0, 2, 4]);
    expect(result).toEqual({ items: ["a", "b", "c", "d", "e"], total: 5 });
  });

  it("fails closed when the reported total changes between pages", async () => {
    await expect(
      collectOffsetPages(
        async (offset) => ({
          items: [offset],
          total: offset === 0 ? 2 : 3,
          offset,
        }),
        1,
      ),
    ).rejects.toThrow("changed while pagination");
  });
});

describe("collectNumberedPages", () => {
  it("collects all page-numbered portfolio results", async () => {
    const result = await collectNumberedPages(async (page, limit) => {
      const all = [1, 2, 3];
      const start = (page - 1) * limit;
      return {
        items: all.slice(start, start + limit),
        total: all.length,
        page,
      };
    }, 2);

    expect(result).toEqual({ items: [1, 2, 3], total: 3 });
  });

  it("returns an empty directory without requesting a second page", async () => {
    let calls = 0;
    const result = await collectNumberedPages(async (page) => {
      calls += 1;
      return { items: [], total: 0, page };
    });

    expect(calls).toBe(1);
    expect(result).toEqual({ items: [], total: 0 });
  });
});
