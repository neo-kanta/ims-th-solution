/**
 * Display helpers for the My Funds cards.
 *
 * Money values arrive as decimal strings on the wire (see investment DTOs).
 * Frontend MUST NOT recompute authoritative monetary values — we only format
 * for display. `formatMoneyCompact` is the cockpit-friendly variant
 * (₿2.57B, ₿543M) used on dense cards.
 */
const CURRENCY_SYMBOL: Record<string, string> = {
  THB: "฿",
  USD: "$",
  EUR: "€",
  JPY: "¥",
  GBP: "£",
};

export function currencySymbol(code: string | null | undefined): string {
  if (!code) return "";
  return CURRENCY_SYMBOL[code.toUpperCase()] ?? `${code} `;
}

const COMPACT_TIERS: Array<{ limit: number; suffix: string; divisor: number }> = [
  { limit: 1_000_000_000_000, suffix: "T", divisor: 1_000_000_000_000 },
  { limit: 1_000_000_000, suffix: "B", divisor: 1_000_000_000 },
  { limit: 1_000_000, suffix: "M", divisor: 1_000_000 },
  { limit: 1_000, suffix: "K", divisor: 1_000 },
];

export function formatMoneyCompact(
  value: number | null | undefined,
  currency: string | null | undefined,
  signed = false,
): string {
  if (value === null || value === undefined || !Number.isFinite(value)) {
    return "—";
  }
  const sym = currencySymbol(currency);
  const abs = Math.abs(value);
  const tier = COMPACT_TIERS.find((t) => abs >= t.limit);
  const sign = signed && value > 0 ? "+" : value < 0 ? "−" : "";

  if (!tier) {
    return `${sign}${sym}${formatPlain(abs)}`;
  }
  const scaled = abs / tier.divisor;
  const digits = scaled >= 100 ? 0 : scaled >= 10 ? 1 : 2;
  return `${sign}${sym}${scaled.toFixed(digits)}${tier.suffix}`;
}

export function formatMoneyExact(
  value: number | null | undefined,
  currency: string | null | undefined,
): string {
  if (value === null || value === undefined || !Number.isFinite(value)) {
    return "—";
  }
  return `${currencySymbol(currency)}${formatPlain(value)}`;
}

function formatPlain(value: number): string {
  return value.toLocaleString("en-US", {
    maximumFractionDigits: 0,
    minimumFractionDigits: 0,
  });
}

export function formatPercent(
  value: number | null | undefined,
  digits = 2,
  signed = false,
): string {
  if (value === null || value === undefined || !Number.isFinite(value)) {
    return "—";
  }
  const sign = signed && value > 0 ? "+" : value < 0 ? "−" : "";
  return `${sign}${Math.abs(value).toFixed(digits)}%`;
}

export function parseDecimal(value: string | null | undefined): number {
  if (!value) return 0;
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}

export function parseDecimalOrNull(value: string | null | undefined): number | null {
  if (!value) return null;
  const n = Number(value);
  return Number.isFinite(n) ? n : null;
}

/**
 * Best-effort relative time. Returns "—" when the timestamp is missing.
 * Kept ASCII so the same string renders in EN/TH/ZH layouts without surprises.
 */
export function relativeTime(iso: string | null | undefined, now = Date.now()): string {
  if (!iso) return "—";
  const t = Date.parse(iso);
  if (!Number.isFinite(t)) return "—";
  const diffSec = Math.round((now - t) / 1000);
  if (diffSec < 5) return "just now";
  if (diffSec < 60) return `${diffSec}s ago`;
  if (diffSec < 3600) return `${Math.round(diffSec / 60)}m ago`;
  if (diffSec < 86400) return `${Math.round(diffSec / 3600)}h ago`;
  return `${Math.round(diffSec / 86400)}d ago`;
}
