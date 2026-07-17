export interface AllocationDonutItem {
  key: string;
  label: string;
  description?: string;
  value: number;
}

export interface AllocationDonutSlice extends AllocationDonutItem {
  color: string;
  percentage: number;
}

export interface AllocationDonutArc {
  key: string;
  dashArray: string;
  dashOffset: number;
}

export const ALLOCATION_DONUT_PALETTE = [
  "var(--terminal-series-1, #2563eb)",
  "var(--terminal-series-2, #0d9488)",
  "var(--terminal-series-3, #d97706)",
  "var(--terminal-series-4, #7c3aed)",
  "var(--terminal-series-5, #db2777)",
  "var(--terminal-series-6, #4f46e5)",
] as const;

export const ALLOCATION_DONUT_OTHER_COLOR =
  "var(--terminal-series-other, #64748b)";

/**
 * Produces a stable, readable chart model from raw portfolio positions.
 * Invalid/zero values are excluded and the long tail is folded into a final
 * "Other" slice so the ring and legend remain legible in the overview rail.
 */
export function buildAllocationDonutSlices(
  items: AllocationDonutItem[],
  otherLabel: string,
  maxSlices = 7,
): AllocationDonutSlice[] {
  const normalized = items
    .filter((item) => Number.isFinite(item.value) && item.value > 0)
    .sort((a, b) => b.value - a.value);

  if (normalized.length === 0 || maxSlices <= 0) return [];

  const total = normalized.reduce((sum, item) => sum + item.value, 0);
  const visibleCount = Math.max(1, maxSlices - 1);
  const shouldGroupTail = normalized.length > maxSlices;
  const visible = shouldGroupTail
    ? normalized.slice(0, visibleCount)
    : normalized;

  const slices: AllocationDonutItem[] = [...visible];
  if (shouldGroupTail) {
    const tail = normalized.slice(visibleCount);
    slices.push({
      key: "__other__",
      label: otherLabel,
      description: tail.map((item) => item.label).join(", "),
      value: tail.reduce((sum, item) => sum + item.value, 0),
    });
  }

  return slices.map((slice, index) => ({
    ...slice,
    color:
      slice.key === "__other__"
        ? ALLOCATION_DONUT_OTHER_COLOR
        : (ALLOCATION_DONUT_PALETTE[index] ?? ALLOCATION_DONUT_OTHER_COLOR),
    percentage: (slice.value / total) * 100,
  }));
}

/**
 * Returns stacked-circle SVG geometry. Each arc starts where the previous one
 * ended; together the positive slices cover exactly one circumference.
 */
export function buildAllocationDonutArcs(
  slices: AllocationDonutSlice[],
  radius: number,
): AllocationDonutArc[] {
  const circumference = 2 * Math.PI * radius;
  let accumulated = 0;

  return slices.map((slice) => {
    const fraction = slice.percentage / 100;
    const arc = {
      key: slice.key,
      dashArray: `${fraction * circumference} ${circumference}`,
      dashOffset: -accumulated * circumference,
    };
    accumulated += fraction;
    return arc;
  });
}
