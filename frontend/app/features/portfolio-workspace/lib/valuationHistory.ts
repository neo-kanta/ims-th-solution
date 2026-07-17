export interface ValuationHistoryEntry {
  business_date?: string;
  aum?: string;
}

export interface ValuationHistoryPoint {
  date: string;
  value: number;
}

export interface ValuationSparkline {
  areaPath: string;
  coordinates: Array<ValuationHistoryPoint & { x: number; y: number }>;
  linePath: string;
}

const DAY_MS = 86_400_000;

/**
 * Normalise official valuation snapshots for charting. Duplicate business
 * dates are collapsed to the last snapshot returned for that date and the
 * range is anchored to the newest available snapshot, not the device clock.
 */
export function buildValuationHistorySeries(
  entries: ValuationHistoryEntry[],
  rangeDays: number,
): ValuationHistoryPoint[] {
  const byDate = new Map<string, number>();

  for (const entry of entries) {
    const date = entry.business_date?.trim();
    const value = Number(entry.aum);
    if (
      !date ||
      !Number.isFinite(Date.parse(date)) ||
      !Number.isFinite(value) ||
      value <= 0
    ) {
      continue;
    }
    byDate.set(date, value);
  }

  const sorted = [...byDate.entries()]
    .map(([date, value]) => ({ date, value }))
    .sort((a, b) => a.date.localeCompare(b.date));

  if (sorted.length === 0 || !Number.isFinite(rangeDays) || rangeDays <= 0) {
    return sorted;
  }

  const newest = Date.parse(sorted[sorted.length - 1]!.date);
  const cutoff = newest - rangeDays * DAY_MS;
  return sorted.filter((point) => Date.parse(point.date) >= cutoff);
}

export function buildValuationSparkline(
  points: ValuationHistoryPoint[],
  width: number,
  height: number,
  padding = 6,
): ValuationSparkline {
  if (points.length < 2 || width <= 0 || height <= 0) {
    return { areaPath: "", coordinates: [], linePath: "" };
  }

  const values = points.map((point) => point.value);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const spread = max - min || Math.max(Math.abs(max) * 0.01, 1);
  const drawableWidth = Math.max(width - padding * 2, 1);
  const drawableHeight = Math.max(height - padding * 2, 1);

  const coordinates = points.map((point, index) => {
    const x = padding + (index / (points.length - 1)) * drawableWidth;
    const y = padding + (1 - (point.value - min) / spread) * drawableHeight;
    return { ...point, x, y };
  });

  const linePath = coordinates
    .map(
      (point, index) =>
        `${index === 0 ? "M" : "L"}${point.x.toFixed(2)},${point.y.toFixed(2)}`,
    )
    .join(" ");
  const first = coordinates[0]!;
  const last = coordinates[coordinates.length - 1]!;
  const baseline = height - padding;

  return {
    linePath,
    coordinates,
    areaPath: `${linePath} L${last.x.toFixed(2)},${baseline.toFixed(2)} L${first.x.toFixed(2)},${baseline.toFixed(2)} Z`,
  };
}

export function valuationHistoryChange(
  points: ValuationHistoryPoint[],
): number | null {
  if (points.length < 2) return null;
  const first = points[0]!.value;
  const latest = points[points.length - 1]!.value;
  if (first === 0) return null;
  return ((latest - first) / first) * 100;
}
