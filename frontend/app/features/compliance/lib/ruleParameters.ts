/**
 * Safe boundary normalizer + human-readable formatting for Portfolio
 * Compliance V2 rule parameters. `PortfolioRuleCatalogEntry.parameters` is
 * typed `Record<string, never>` by the generated client (backend/pkg/contract/
 * contracts.go's `swaggertype:"object"` gives swaggo no shape info — see
 * docs/MANAGER/TASKS.md P2 follow-up). We accept `unknown` here and validate
 * a plain, non-array object at the boundary instead of casting the drifted
 * generated type further.
 */

export function asParameterRecord(value: unknown): Record<string, unknown> {
  if (value && typeof value === "object" && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  return {};
}

export interface ParameterEntry {
  key: string;
  label: string;
  value: string;
}

/** Splits a snake_case/camelCase parameter key into readable words, e.g. "max_percent_nav" -> "Max Percent Nav". */
function humanizeKey(key: string): string {
  const spaced = key
    .replace(/[_-]+/g, " ")
    .replace(/([a-z0-9])([A-Z])/g, "$1 $2");
  return spaced
    .split(" ")
    .filter(Boolean)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(" ");
}

/** Converts a rule instance's raw parameter set into a stable, human-readable list for display. */
export function formatParameterEntries(value: unknown): ParameterEntry[] {
  const record = asParameterRecord(value);
  return Object.entries(record)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([key, val]) => ({
      key,
      label: humanizeKey(key),
      value: String(val),
    }));
}
