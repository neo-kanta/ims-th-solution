export function formatDecimalOrDash(value: string | null | undefined): string {
  if (!value) return "—";
  return value;
}

export function formatTimestamp(iso: string | null | undefined): string {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleString("th-TH", {
      timeZone: "Asia/Bangkok",
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return iso;
  }
}

export function formatDateOnly(iso: string | null | undefined): string {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleDateString("th-TH", {
      timeZone: "Asia/Bangkok",
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
    });
  } catch {
    return iso;
  }
}

export function scopeLabel(scope: string | null | undefined): string {
  if (scope === "PERSONAL") return "Personal";
  if (scope === "PORTFOLIO") return "Portfolio";
  return scope ?? "—";
}

export function directionLabel(direction: string | null | undefined): string {
  if (direction === "ABOVE") return "Above";
  if (direction === "BELOW") return "Below";
  return direction ?? "—";
}

export function itemStatusLabel(status: string | null | undefined): string {
  if (status === "ACTIVE") return "Active";
  if (status === "DISABLED") return "Disabled";
  return status ?? "—";
}

export function ruleStatusLabel(status: string | null | undefined): string {
  if (status === "ENABLED") return "Enabled";
  if (status === "DISABLED") return "Disabled";
  return status ?? "—";
}

export function ruleStateLabel(state: string | null | undefined): string {
  if (state === "NON_BREACHED") return "Non-breached";
  if (state === "BREACHED") return "Breached";
  if (state === "UNKNOWN") return "Unknown";
  return state ?? "—";
}

export function ackStateLabel(state: string | null | undefined): string {
  if (state === "ACKNOWLEDGED") return "Acknowledged";
  if (state === "UNACKNOWLEDGED") return "Unacknowledged";
  return state ?? "—";
}

export function notificationStatusLabel(status: string | null | undefined): string {
  if (status === "PENDING") return "Pending";
  if (status === "CREATED") return "Sent";
  if (status === "SUPPRESSED") return "Suppressed";
  if (status === "FAILED") return "Failed";
  if (status === "SKIPPED") return "Skipped";
  return status ?? "—";
}

export function thresholdSummary(
  direction: string | undefined,
  value: string | undefined,
  currency: string | undefined | null,
): string {
  if (!direction || !value) return "—";
  const dir = direction === "ABOVE" ? "↑" : "↓";
  const curr = currency ? ` ${currency}` : "";
  return `${dir} ${value}${curr}`;
}

export function securityLabel(
  security: { display_symbol?: string; name?: string } | null | undefined,
): string {
  if (!security) return "—";
  return security.display_symbol ?? security.name ?? "—";
}

export function isUuid(value: string): boolean {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(value.trim());
}

export function isPositiveDecimal(value: string): boolean {
  if (!value || !value.trim()) return false;
  const n = Number(value);
  return !Number.isNaN(n) && n > 0;
}

export function staleLabel(stale: boolean, staleReason: string | null | undefined, t?: (key: string) => string): string {
  if (!stale) return "";
  if (staleReason) {
    const prefix = t ? t("watchlist.alert.stale") : "Stale";
    return `${prefix}: ${staleReason}`;
  }
  return t ? t("watchlist.alert.stale") : "Market data is stale";
}

export function portfolioLabel(
  portfolio: { code?: string; name?: string; id?: string } | null | undefined,
): string {
  if (!portfolio) return "—";
  const code = portfolio.code && !isUuid(portfolio.code) ? portfolio.code : "";
  const name = portfolio.name && !isUuid(portfolio.name) ? portfolio.name : "";
  if (code && name) return `${code} – ${name}`;
  if (code) return code;
  if (name) return name;
  return "Portfolio";
}

export function portfolioDescriptorLabel(
  portfolio: { portfolio_code?: string; portfolio_name?: string; display_name?: string; portfolio_id?: string } | null | undefined,
): string {
  if (!portfolio) return "—";
  const code = portfolio.portfolio_code && !isUuid(portfolio.portfolio_code) ? portfolio.portfolio_code : "";
  let name = portfolio.display_name && !isUuid(portfolio.display_name) ? portfolio.display_name : "";
  if (!name) {
    name = portfolio.portfolio_name && !isUuid(portfolio.portfolio_name) ? portfolio.portfolio_name : "";
  }
  if (code && name && code !== name) {
    return name.includes(code) ? name : `${code} – ${name}`;
  }
  if (code) return code;
  if (name) return name;
  return "Portfolio";
}
