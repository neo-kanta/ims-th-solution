/**
 * Pure helpers for the portfolio-first "Buy / Sell" entry point (OP-01).
 * Unlike the legacy fund-scoped flow, this queries the full portfolio list
 * directly — it never requires a fund to be pre-selected first.
 */

export interface PortfolioPick {
  code: string;
  name: string;
  status: string | null;
}

export function toPortfolioPicks(
  items: ReadonlyArray<{
    code?: string | null;
    name?: string | null;
    status?: string | null;
  }>,
): PortfolioPick[] {
  return items
    .filter((p): p is { code: string; name?: string | null; status?: string | null } =>
      Boolean(p.code),
    )
    .map((p) => ({
      code: p.code,
      name: p.name?.trim() || p.code,
      status: p.status ?? null,
    }));
}

export function filterPortfolioPicks(
  picks: ReadonlyArray<PortfolioPick>,
  query: string,
): PortfolioPick[] {
  const q = query.trim().toLowerCase();
  if (!q) return [...picks];
  return picks.filter(
    (p) => p.code.toLowerCase().includes(q) || p.name.toLowerCase().includes(q),
  );
}
