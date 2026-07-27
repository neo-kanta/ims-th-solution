/**
 * Tests for canCancelCashRequest (G2 item 6): only the submitter of a still-
 * PENDING cash request may cancel it. This is a pure function so it needs no
 * mocking of the Pinia auth store or Nuxt runtime.
 */
import { describe, expect, it } from "vitest";

import { canCancelCashRequest } from "../app/features/portfolio-workspace/lib/cashRequestAuthz";

describe("canCancelCashRequest", () => {
  const submitterId = "11111111-1111-1111-1111-111111111111";
  const otherUserId = "22222222-2222-2222-2222-222222222222";

  it("allows the submitter to cancel their own PENDING request", () => {
    expect(canCancelCashRequest({ status: "PENDING", submitted_by: submitterId }, submitterId)).toBe(true);
  });

  it("refuses a different authenticated user viewing the same PENDING request", () => {
    expect(canCancelCashRequest({ status: "PENDING", submitted_by: submitterId }, otherUserId)).toBe(false);
  });

  it("refuses the submitter once the request is no longer PENDING", () => {
    for (const status of ["APPROVED", "REJECTED", "CANCELLED"]) {
      expect(canCancelCashRequest({ status, submitted_by: submitterId }, submitterId)).toBe(false);
    }
  });

  it("refuses when the current user id is unknown (not yet authenticated / restoring session)", () => {
    expect(canCancelCashRequest({ status: "PENDING", submitted_by: submitterId }, undefined)).toBe(false);
    expect(canCancelCashRequest({ status: "PENDING", submitted_by: submitterId }, null)).toBe(false);
  });
});
