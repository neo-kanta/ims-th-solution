import { ref } from "vue";

import { downloadBlob } from "../lib/holdingsFormat";
import { holdingsMockApi } from "../services/holdingsMockApi";
import type {
  AllocationDimension,
  AllocationPayload,
  FundFooter,
  HoldingsAssets,
  HoldingsSubTab,
  HoldingsSummary,
  NavHistoryPayload,
  NavHistoryRange,
  RatiosPayload,
} from "../types";

/**
 * Orchestrates the five-section Holdings page. Loaders run in parallel so the
 * NAV chart can finish while the asset table is still resolving — matches the
 * UX we'll see once the real API is wired.
 *
 * MOCK REPLACEMENT POINT — swap holdingsMockApi for holdingsApi to switch.
 */
export function useHoldings(fundId: string, viewLanguage: "en" | "th" | "zh" = "en") {
  const summary = ref<HoldingsSummary | null>(null);
  const assets = ref<HoldingsAssets | null>(null);
  const allocation = ref<AllocationPayload | null>(null);
  const ratios = ref<RatiosPayload | null>(null);
  const navHistory = ref<NavHistoryPayload | null>(null);
  const footer = ref<FundFooter | null>(null);

  const asOf = ref<string>("2026-05-21");
  const subTab = ref<HoldingsSubTab>("overview");
  const allocationDimension = ref<AllocationDimension>("country");
  const navRange = ref<NavHistoryRange>("3M");

  const loading = ref(false);
  const loadingSection = ref<Record<string, boolean>>({});
  const error = ref<string | null>(null);
  const exportError = ref<string | null>(null);
  const exporting = ref(false);

  async function refreshAll() {
    loading.value = true;
    error.value = null;
    try {
      const results = await Promise.allSettled([
        load("summary", () => holdingsMockApi.getSummary(fundId, asOf.value)),
        load("assets", () => holdingsMockApi.getAssets(fundId, asOf.value, subTab.value)),
        load("allocation", () => holdingsMockApi.getAllocation(fundId, asOf.value)),
        load("ratios", () => holdingsMockApi.getRatios(fundId, asOf.value)),
        load("navHistory", () => holdingsMockApi.getNavHistory(fundId, navRange.value)),
        load("footer", () => holdingsMockApi.getFooter(fundId, asOf.value, viewLanguage)),
      ]);

      const [s, a, al, r, nh, f] = results;
      if (s.status === "fulfilled") summary.value = s.value as HoldingsSummary;
      if (a.status === "fulfilled") assets.value = a.value as HoldingsAssets;
      if (al.status === "fulfilled") allocation.value = al.value as AllocationPayload;
      if (r.status === "fulfilled") ratios.value = r.value as RatiosPayload;
      if (nh.status === "fulfilled") navHistory.value = nh.value as NavHistoryPayload;
      if (f.status === "fulfilled") footer.value = f.value as FundFooter;

      const firstFailure = results.find((x) => x.status === "rejected");
      if (firstFailure && firstFailure.status === "rejected") {
        error.value = extractMessage(firstFailure.reason, "Some sections failed to load.");
      }
    } finally {
      loading.value = false;
    }
  }

  async function load<T>(key: string, fn: () => Promise<T>): Promise<T | undefined> {
    loadingSection.value = { ...loadingSection.value, [key]: true };
    try {
      return await fn();
    } finally {
      loadingSection.value = { ...loadingSection.value, [key]: false };
    }
  }

  async function setAsOf(next: string) {
    asOf.value = next;
    await refreshAll();
  }

  async function setSubTab(next: HoldingsSubTab) {
    subTab.value = next;
    // Sub-tab change re-loads only the assets section in real life; here we
    // re-call assets so the loading state pings, matching production UX.
    assets.value = await load("assets", () =>
      holdingsMockApi.getAssets(fundId, asOf.value, next),
    ) ?? assets.value;
  }

  function setAllocationDimension(next: AllocationDimension) {
    // No refetch — the API returns all four dimensions in one payload.
    allocationDimension.value = next;
  }

  async function setNavRange(next: NavHistoryRange) {
    navRange.value = next;
    navHistory.value = await load("navHistory", () =>
      holdingsMockApi.getNavHistory(fundId, next),
    ) ?? navHistory.value;
  }

  async function download() {
    exporting.value = true;
    exportError.value = null;
    try {
      const { blob, filename } = await holdingsMockApi.exportCsv(fundId, asOf.value);
      downloadBlob(blob, filename);
    } catch (err) {
      exportError.value = extractMessage(err, "Export failed.");
    } finally {
      exporting.value = false;
    }
  }

  return {
    // data
    summary,
    assets,
    allocation,
    ratios,
    navHistory,
    footer,
    // controls
    asOf,
    subTab,
    allocationDimension,
    navRange,
    // state
    loading,
    loadingSection,
    error,
    exporting,
    exportError,
    // actions
    refreshAll,
    setAsOf,
    setSubTab,
    setAllocationDimension,
    setNavRange,
    download,
  };
}

function extractMessage(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const msg = (err as { message?: unknown }).message;
  return typeof msg === "string" && msg.trim() ? msg : fallback;
}
