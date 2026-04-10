import { describe, expect, it } from "vitest";
import {
  isPublicRouteMeta,
  resolveRequiredPermission,
} from "../app/shared/routing/routeAccess";

describe("route access helpers", () => {
  it("reads a direct permission from route meta", () => {
    expect(resolveRequiredPermission({ permission: "AUDIT_VIEW" })).toBe(
      "AUDIT_VIEW",
    );
  });

  it("reads a nested permission from legacy route meta", () => {
    expect(
      resolveRequiredPermission({
        meta: { permission: "PORTFOLIO_VIEW" },
      }),
    ).toBe("PORTFOLIO_VIEW");
  });

  it("detects direct public auth meta", () => {
    expect(isPublicRouteMeta({ auth: false })).toBe(true);
  });

  it("detects nested public auth meta", () => {
    expect(isPublicRouteMeta({ meta: { auth: false } })).toBe(true);
  });

  it("returns false when auth meta is absent", () => {
    expect(isPublicRouteMeta({})).toBe(false);
  });
});
