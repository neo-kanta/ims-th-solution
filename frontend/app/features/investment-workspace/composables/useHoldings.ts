import { ref } from "vue";

import { downloadBlob } from "../lib/holdingsFormat";
import type {
  AllocationDimension,
  HoldingsSubTab,
  NavHistoryRange,
} from "../types";

/**
 * Holdings workspace UI-state controller.
 *
 * Owns only the local controls the page renders (as-of date, sub-tab,
 * allocation dimension, NAV range, error banner, export spinner). All
 * financial data is fetched by the page itself through the backend-backed
 * `myFundsApi`; this composable no longer issues any data calls of its own.
 *
 * Export reuses the rows already loaded for the table so we never invent
 * holdings data — the CSV is exactly what the user is looking at.
 */
export function useHoldings() {
  const asOf = ref<string>(todayIsoUTC());
  const subTab = ref<HoldingsSubTab>("overview");
  const allocationDimension = ref<AllocationDimension>("country");
  const navRange = ref<NavHistoryRange>("3M");

  const loading = ref(false);
  const error = ref<string | null>(null);
  const exportError = ref<string | null>(null);
  const exporting = ref(false);

  async function setAsOf(next: string) {
    asOf.value = next;
  }

  async function setSubTab(next: HoldingsSubTab) {
    subTab.value = next;
  }

  function setAllocationDimension(next: AllocationDimension) {
    allocationDimension.value = next;
  }

  async function setNavRange(next: NavHistoryRange) {
    navRange.value = next;
  }

  /**
   * Generate a CSV from rows the page has already loaded from the backend.
   * Caller passes a flat list of [section, ...cells] rows so this composable
   * stays oblivious to the row shape — it only knows how to escape and emit.
   */
  async function exportRows(
    rows: string[][],
    filename: string,
  ): Promise<void> {
    exporting.value = true;
    exportError.value = null;
    try {
      if (!rows.length) {
        throw new Error("No rows to export.");
      }
      const bom = "﻿";
      const body = rows.map((cols) => cols.map(csvCell).join(",")).join("\r\n");
      const blob = new Blob([bom + body], {
        type: "text/csv;charset=utf-8",
      });
      downloadBlob(blob, filename);
    } catch (err) {
      exportError.value = extractMessage(err, "Export failed.");
    } finally {
      exporting.value = false;
    }
  }

  return {
    // controls
    asOf,
    subTab,
    allocationDimension,
    navRange,
    // state
    loading,
    error,
    exporting,
    exportError,
    // actions
    setAsOf,
    setSubTab,
    setAllocationDimension,
    setNavRange,
    exportRows,
  };
}

function todayIsoUTC(): string {
  return new Date().toISOString().slice(0, 10);
}

function csvCell(value: string): string {
  if (/[",\r\n]/.test(value)) {
    return `"${value.replace(/"/g, '""')}"`;
  }
  return value;
}

function extractMessage(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const msg = (err as { message?: unknown }).message;
  return typeof msg === "string" && msg.trim() ? msg : fallback;
}
