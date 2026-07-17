import { describe, expect, it } from "vitest";

import {
  buildAllocationDonutArcs,
  buildAllocationDonutSlices,
} from "../app/features/portfolio-workspace/lib/allocationDonut";

describe("buildAllocationDonutSlices", () => {
  it("sorts positive positions and calculates their share of the portfolio", () => {
    const slices = buildAllocationDonutSlices(
      [
        { key: "cash", label: "Cash", value: 20 },
        { key: "equity", label: "Equity", value: 70 },
        { key: "bond", label: "Bond", value: 10 },
      ],
      "Other",
    );

    expect(slices.map((slice) => slice.key)).toEqual([
      "equity",
      "cash",
      "bond",
    ]);
    expect(slices.map((slice) => slice.percentage)).toEqual([70, 20, 10]);
  });

  it("excludes invalid and non-positive values rather than drawing misleading arcs", () => {
    const slices = buildAllocationDonutSlices(
      [
        { key: "valid", label: "Valid", value: 125 },
        { key: "zero", label: "Zero", value: 0 },
        { key: "negative", label: "Negative", value: -10 },
        { key: "invalid", label: "Invalid", value: Number.NaN },
      ],
      "Other",
    );

    expect(slices).toHaveLength(1);
    expect(slices[0]?.key).toBe("valid");
    expect(slices[0]?.percentage).toBe(100);
  });

  it("folds a long tail into Other while preserving the full value", () => {
    const items = Array.from({ length: 9 }, (_, index) => ({
      key: `position-${index + 1}`,
      label: `Position ${index + 1}`,
      value: 90 - index * 5,
    }));
    const slices = buildAllocationDonutSlices(items, "Other", 7);

    expect(slices).toHaveLength(7);
    expect(slices.at(-1)?.key).toBe("__other__");
    expect(slices.reduce((sum, slice) => sum + slice.value, 0)).toBe(
      items.reduce((sum, item) => sum + item.value, 0),
    );
    expect(
      slices.reduce((sum, slice) => sum + slice.percentage, 0),
    ).toBeCloseTo(100);
  });
});

describe("buildAllocationDonutArcs", () => {
  it("places each arc after the preceding arc", () => {
    const slices = buildAllocationDonutSlices(
      [
        { key: "a", label: "A", value: 50 },
        { key: "b", label: "B", value: 30 },
        { key: "c", label: "C", value: 20 },
      ],
      "Other",
    );
    const arcs = buildAllocationDonutArcs(slices, 48);
    const circumference = 2 * Math.PI * 48;

    expect(arcs[0]?.dashOffset).toBeCloseTo(0);
    expect(arcs[1]?.dashOffset).toBeCloseTo(-0.5 * circumference);
    expect(arcs[2]?.dashOffset).toBeCloseTo(-0.8 * circumference);
  });
});
