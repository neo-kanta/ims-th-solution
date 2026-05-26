/**
 * Pure-logic tests for the security-detail composable's data-quality
 * derivation. These don't touch the Nuxt runtime — we import the exported
 * deriveDataQuality helper directly.
 */
import { describe, expect, it } from "vitest";

import { deriveDataQuality } from "../app/features/market-data/composables/dataQuality";
import type {
  ProviderMappingView,
  MarketQuoteView,
} from "../app/features/market-data/market-data.types";

const activeMapping: ProviderMappingView = {
  mappingId: "m1",
  providerCode: "yahoo",
  providerSymbol: "KBANK.BK",
  status: "ACTIVE",
  isPrimary: true,
};

describe("deriveDataQuality", () => {
  it("returns NOT_IMPORTED + UNMAPPED when there is neither quote nor mapping", () => {
    const q = deriveDataQuality([], undefined);
    expect(q.freshness).toBe("NOT_IMPORTED");
    expect(q.mapping).toBe("UNMAPPED");
  });

  it("returns NOT_IMPORTED + MAPPED when mapping exists but no quote is loaded", () => {
    const q = deriveDataQuality([activeMapping], undefined);
    expect(q.freshness).toBe("NOT_IMPORTED");
    expect(q.mapping).toBe("MAPPED");
  });

  it("returns FRESH + MAPPED when both are present and quote is not stale", () => {
    const quote: MarketQuoteView = {
      provider: "yahoo",
      lastPrice: "142.50",
      stale: false,
    };
    const q = deriveDataQuality([activeMapping], quote);
    expect(q.freshness).toBe("FRESH");
    expect(q.mapping).toBe("MAPPED");
    expect(q.rateLimited).toBeFalsy();
  });

  it("returns RATE_LIMITED when staleReason mentions rate limiting", () => {
    const quote: MarketQuoteView = {
      provider: "yahoo",
      lastPrice: "142.50",
      stale: true,
      staleReason: "provider 429 rate limited",
    };
    const q = deriveDataQuality([activeMapping], quote);
    expect(q.freshness).toBe("RATE_LIMITED");
    expect(q.rateLimited).toBe(true);
    expect(q.errorCode).toBe("RATE_LIMITED");
  });

  it("returns FAILED when staleReason mentions failure", () => {
    const quote: MarketQuoteView = {
      provider: "yahoo",
      lastPrice: "142.50",
      stale: true,
      staleReason: "all providers failed",
    };
    const q = deriveDataQuality([activeMapping], quote);
    expect(q.freshness).toBe("FAILED");
  });

  it("returns STALE when staleReason is generic", () => {
    const quote: MarketQuoteView = {
      provider: "yahoo",
      lastPrice: "142.50",
      stale: true,
      staleReason: "returning cached snapshot",
    };
    const q = deriveDataQuality([activeMapping], quote);
    expect(q.freshness).toBe("STALE");
  });

  it("returns UNMAPPED mapping status when only inactive mappings exist", () => {
    const inactive: ProviderMappingView = { ...activeMapping, status: "INACTIVE" };
    const q = deriveDataQuality([inactive], undefined);
    expect(q.mapping).toBe("UNMAPPED");
  });

  it("returns REVIEW_REQUIRED mapping status when a candidate-style mapping exists", () => {
    const review: ProviderMappingView = { ...activeMapping, status: "REVIEW_REQUIRED" };
    const q = deriveDataQuality([review], undefined);
    expect(q.mapping).toBe("REVIEW_REQUIRED");
  });
});
