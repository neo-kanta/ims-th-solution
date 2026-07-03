/**
 * Display helpers for the investment ledger UI. Numbers come from the
 * backend as strings (decimal precision matters for cash and quantities),
 * so we centralise parsing and formatting here.
 */

const NBSP = " ";

function parseDecimal(value: string | number | null | undefined): number | null {
  if (value === null || value === undefined) return null;
  if (typeof value === "number") return Number.isFinite(value) ? value : null;
  const trimmed = value.trim();
  if (!trimmed) return null;
  const n = Number(trimmed);
  return Number.isFinite(n) ? n : null;
}

export interface FormatNumberOptions {
  minimumFractionDigits?: number;
  maximumFractionDigits?: number;
  signDisplay?: "auto" | "always" | "never";
}

export function formatNumber(
  value: string | number | null | undefined,
  options: FormatNumberOptions = {},
): string {
  const parsed = parseDecimal(value);
  if (parsed === null) return "—";
  const min = options.minimumFractionDigits ?? 2;
  const max = options.maximumFractionDigits ?? Math.max(min, 4);
  return parsed.toLocaleString("en-US", {
    minimumFractionDigits: min,
    maximumFractionDigits: max,
    signDisplay: options.signDisplay ?? "auto",
  });
}

export function formatQuantity(value: string | number | null | undefined): string {
  const parsed = parseDecimal(value);
  if (parsed === null) return "—";
  const digits = Math.abs(parsed) >= 1 ? 0 : 4;
  return parsed.toLocaleString("en-US", {
    minimumFractionDigits: digits,
    maximumFractionDigits: 4,
  });
}

export function formatMoney(
  value: string | number | null | undefined,
  currency?: string | null,
): string {
  const parsed = parseDecimal(value);
  if (parsed === null) return "—";
  const body = parsed.toLocaleString("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
  if (!currency) return body;
  return `${currency}${NBSP}${body}`;
}

export function formatMoneySigned(
  value: string | number | null | undefined,
  currency?: string | null,
): string {
  const parsed = parseDecimal(value);
  if (parsed === null) return "—";
  const sign = parsed > 0 ? "+" : "";
  return `${sign}${formatMoney(parsed, currency)}`;
}

export function compareDecimal(
  a: string | number | null | undefined,
  b: string | number | null | undefined,
): number {
  const pa = parseDecimal(a);
  const pb = parseDecimal(b);
  if (pa === null && pb === null) return 0;
  if (pa === null) return 1;
  if (pb === null) return -1;
  return pa - pb;
}

export function isPositiveNumeric(value: string): boolean {
  if (!value) return false;
  const n = Number(value);
  return Number.isFinite(n) && n > 0;
}

export function isNonNegativeNumeric(value: string): boolean {
  if (!value) return true;
  const n = Number(value);
  return Number.isFinite(n) && n >= 0;
}

export function todayIso(): string {
  return new Date().toISOString().slice(0, 10);
}

export function shortenId(id: string | null | undefined, head = 8, tail = 4): string {
  if (!id) return "";
  if (id.length <= head + tail + 1) return id;
  return `${id.slice(0, head)}…${id.slice(-tail)}`;
}

const VERDICT_LABEL: Record<string, string> = {
  PASS: "Pass",
  WARN: "Warning",
  BLOCK: "Block",
};

export function verdictLabel(verdict: string | null | undefined): string {
  if (!verdict) return "Unknown";
  return VERDICT_LABEL[verdict] ?? verdict;
}

export function verdictTone(
  verdict: string | null | undefined,
): "success" | "warning" | "error" | "neutral" {
  switch (verdict) {
    case "PASS":
      return "success";
    case "WARN":
      return "warning";
    case "BLOCK":
      return "error";
    default:
      return "neutral";
  }
}

export function severityTone(
  severity: string | null | undefined,
): "success" | "warning" | "error" | "neutral" {
  switch ((severity ?? "").toUpperCase()) {
    case "BLOCK":
    case "HIGH":
    case "CRITICAL":
      return "error";
    case "WARN":
    case "WARNING":
    case "MEDIUM":
      return "warning";
    case "INFO":
    case "LOW":
      return "neutral";
    default:
      return "neutral";
  }
}

export function statusBadgeVariant(
  status: string | null | undefined,
): "success" | "warning" | "error" | "neutral" | "info" {
  switch ((status ?? "").toUpperCase()) {
    case "POSTED":
    case "SETTLED":
    case "COMPLETE":
      return "success";
    case "PENDING":
    case "DRAFT":
      return "warning";
    case "REVERSED":
    case "FAILED":
    case "REJECTED":
      return "error";
    case "SIMULATED":
      return "info";
    default:
      return "neutral";
  }
}

export const ORDER_SIDES = ["BUY", "SELL"] as const;
export type OrderSide = (typeof ORDER_SIDES)[number];

export function transactionTypeForSide(side: OrderSide): string {
  return side === "BUY" ? "BUY" : "SELL";
}
