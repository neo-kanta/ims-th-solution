/**
 * Display helpers for the Holdings workspace. Backend numbers are
 * already pre-formatted decimal strings — these helpers only handle
 * UI-side concerns (sign coloring, percent rounding, date display in
 * Asia/Bangkok per project policy).
 */

/** Currency code → symbol prefix. Falls back to bare code. */
export function currencyPrefix(code: string): string {
  switch (code.toUpperCase()) {
    case "THB": return "฿ ";
    case "USD": return "$ ";
    case "EUR": return "€ ";
    case "JPY": return "¥ ";
    case "CNY": return "¥ ";
    default:    return `${code} `;
  }
}

/** Render a percentage with sensible defaults. */
export function formatPercent(value: number, digits = 3): string {
  if (!Number.isFinite(value)) return "—";
  return `${value.toFixed(digits)}%`;
}

/** Signed percent change with explicit sign. */
export function formatSignedPercent(value: number, digits = 2): string {
  if (!Number.isFinite(value) || value === 0) return "0.00%";
  const sign = value > 0 ? "+" : "";
  return `${sign}${value.toFixed(digits)}%`;
}

/** P&L semantic for coloring. */
export type PnlTone = "positive" | "negative" | "neutral";

export function pnlTone(value: number | null | undefined): PnlTone {
  if (value === null || value === undefined || value === 0) return "neutral";
  return value > 0 ? "positive" : "negative";
}

/**
 * Bangkok-formatted timestamp. Project policy is UTC in storage, display
 * in Asia/Bangkok — matches features/investment-research/lib/researchReportFormat.ts.
 */
export function formatBangkokDateTime(iso: string): string {
  if (!iso) return "—";
  try {
    const d = new Date(iso);
    return new Intl.DateTimeFormat("en-GB", {
      timeZone: "Asia/Bangkok",
      day: "2-digit",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      hour12: false,
    }).format(d);
  } catch {
    return iso;
  }
}

export function formatBangkokTime(iso: string): string {
  if (!iso) return "—";
  try {
    const d = new Date(iso);
    return new Intl.DateTimeFormat("en-GB", {
      timeZone: "Asia/Bangkok",
      hour: "2-digit",
      minute: "2-digit",
      hour12: false,
    }).format(d);
  } catch {
    return iso;
  }
}

export function formatShortDate(iso: string): string {
  if (!iso) return "—";
  return iso;
}

/**
 * Trigger a browser download for a Blob without leaving artifacts.
 * Used by the Export button while we're still in mock mode.
 */
export function downloadBlob(blob: Blob, filename: string): void {
  if (typeof window === "undefined") return;
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  document.body.removeChild(anchor);
  // Defer revoke so the click handler can fire across browsers.
  setTimeout(() => URL.revokeObjectURL(url), 0);
}

/** Clamp a 0..100 bar value safely for CSS width. */
export function clampPct(value: number): number {
  if (!Number.isFinite(value) || value < 0) return 0;
  if (value > 100) return 100;
  return value;
}
