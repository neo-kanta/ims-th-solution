/**
 * Neither /compliance nor the Portfolio Compliance V2 workspace has an
 * authoritative backend business date, so both surfaces label "today" as the
 * device's local date rather than implying a backend business date.
 */
export function deviceLocalIsoDate(now: Date = new Date()): string {
  const localTime = now.getTime() - now.getTimezoneOffset() * 60_000;
  return new Date(localTime).toISOString().slice(0, 10);
}

/** `null` while loading or on error — renders as "—" rather than a misleading value. */
export function countOrNull(
  value: number,
  loading: boolean,
  hasError: boolean,
): number | null {
  if (loading || hasError) return null;
  return value;
}
