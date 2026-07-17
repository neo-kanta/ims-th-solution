import type { AppTranslationKey } from "../../../shared/i18n/messages";
import type { TranslationParams } from "../../../shared/i18n/types";
import type { UserPermissions } from "../../auth/types";
import type {
  DashboardOverviewChangeTone,
  DashboardOverviewMetricTone,
  DashboardPayload,
  QuickAction,
  ValuationScope,
  ValuationSummaryDTO,
} from "../types";
import { createDateFormatter, resolveIntlLocale } from "../../../shared/i18n/intl";
import { formatMoneyCompact, formatPercent, parseDecimalOrNull } from "../../my-funds/lib/format";

type DashboardLocale = "en" | "th" | "zh";
type TranslateFn = (
  key: AppTranslationKey,
  paramsOrFallback?: TranslationParams | string,
  fallback?: string,
) => string;

export function filterQuickActionsByPermissions(
  actions: QuickAction[],
  permissions: UserPermissions,
): QuickAction[] {
  return actions.filter((action) => (
    !action.requiresPermission
    || permissions.functions.includes(action.requiresPermission)
  ));
}

export function formatDashboardBusinessDate(
  dateString: string,
  locale: DashboardLocale | string = "en",
): string {
  return createDateFormatter(locale, {
    weekday: "long",
    year: "numeric",
    month: "long",
    day: "numeric",
  }, "UTC").format(new Date(dateString));
}

export function formatDashboardHeadlineDate(
  dateString: string,
  locale: DashboardLocale | string = "en",
): string {
  const parts = createDateFormatter(locale, {
    weekday: "short",
    year: "numeric",
    month: "short",
    day: "2-digit",
  }, "UTC").formatToParts(new Date(dateString));
  const lookup = Object.fromEntries(
    parts
      .filter((part) => part.type !== "literal")
      .map((part) => [part.type, part.value]),
  );

  return `${lookup.weekday} ${lookup.day} ${lookup.month} ${lookup.year}`;
}


export function formatDashboardTime(
  timestamp: string,
  locale: DashboardLocale | string = "en",
  timeZone = "UTC",
): string {
  return createDateFormatter(locale, {
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  }, timeZone).format(new Date(timestamp));
}

export function formatDashboardRelativeTime(
  timestamp: string,
  referenceTimestamp: string,
  locale: DashboardLocale | string = "en",
): string {
  const eventTime = new Date(timestamp).getTime();
  const referenceTime = new Date(referenceTimestamp).getTime();
  const diff = Math.max(0, referenceTime - eventTime);
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(minutes / 60);
  const formatter = new Intl.RelativeTimeFormat(resolveIntlLocale(locale), {
    numeric: "auto",
    style: "short",
  });

  if (minutes < 1) {
    return formatter.format(0, "minute");
  }

  if (minutes < 60) {
    return formatter.format(-minutes, "minute");
  }

  if (hours < 24) {
    return formatter.format(-hours, "hour");
  }

  return createDateFormatter(locale, {
    month: "short",
    day: "numeric",
  }, "UTC").format(new Date(timestamp));
}

function canCreateContract(permissions: UserPermissions): boolean {
  if (permissions.functions.length === 0 && permissions.contracts.length === 0) {
    return true;
  }

  return (
    permissions.contracts.length > 0
    || permissions.functions.includes("INVESTMENT_VIEW")
    || permissions.functions.includes("INVESTMENT_CREATE")
  );
}

/**
 * Build a SAFE dashboard payload. No backend exposes the cockpit-style
 * AUM / contracts / approvals / activity / per-contract workflow stages yet,
 * so we return empty arrays — the screen renders honest empty/not-configured
 * states instead of fabricating figures.
 *
 * The real backend-backed views (workflow states, task feed) come from
 * `dashboardApi` via `useDashboardTasks` and are rendered by their own
 * components on the screen.
 */
export function buildDashboardSnapshot(
  now: number,
  permissions: UserPermissions,
  // Translator kept on the signature so callers don't have to rewire; we no
  // longer translate any payload field because every populated field came
  // from a fabricated string.
  _t: TranslateFn,
): DashboardPayload {
  const businessDate = new Date(now).toISOString().split("T")[0] || "";

  return {
    workflow: {
      businessDate,
      dayStatus: "",
      stages: [],
    },
    metrics: [],
    contracts: [],
    activityFeed: [],
    pendingApprovals: [],
    lastRefresh: new Date(now).toISOString(),
    canCreateContract: canCreateContract(permissions),
    createContractUrl: "/investment/operator",
  };
}

// ---------------------------------------------------------------------------
// AUM / P&L scope selector + metric derivation
//
// Pure functions so the scope default/visibility rule and the metric
// value/tone/no-data derivation can be unit tested without mounting the
// dashboard screen (this project's Vitest setup has no Nuxt/component
// runtime — see tests/*.test.ts, which all exercise pure lib functions).
// ---------------------------------------------------------------------------

export interface DashboardScopeOption {
  key: ValuationScope;
  label: string;
}

/**
 * The default AUM scope for a session. "company" is only ever the default
 * (or even offered — see buildScopeOptions) when the caller's data scope is
 * the "*" wildcard; a restricted user must not land on a "company" view
 * that is silently just their own subset mislabelled as company-wide.
 */
export function defaultAumScope(canViewCompany: boolean): ValuationScope {
  return canViewCompany ? "company" : "mine";
}

/**
 * Builds the scope selector's options. "Entire company AUM" is present only
 * for callers with the "*" wildcard data scope; "My AUM" is always offered.
 */
export function buildScopeOptions(
  canViewCompany: boolean,
  t: TranslateFn,
): DashboardScopeOption[] {
  const items: DashboardScopeOption[] = [];
  if (canViewCompany) {
    items.push({
      key: "company",
      label: t("dashboardOverview.scopeCompany"),
    });
  }
  items.push({
    key: "mine",
    label: t("dashboardOverview.scopeMine"),
  });
  return items;
}

export interface DashboardValuationMetric {
  id: string;
  label: string;
  value: string;
  changeLabel: string;
  changeTone: DashboardOverviewChangeTone;
  helperText: string;
  icon: string;
  tone: DashboardOverviewMetricTone;
}

function unavailableValuationHelper(
  summary: ValuationSummaryDTO | null,
  t: TranslateFn,
): string {
  if (summary?.status === "INCOMPLETE") {
    const { includedPortfolioCount, totalPortfolioCount } = summary.coverage;
    if (totalPortfolioCount > 0) {
      return t("dashboardOverview.metricIncompleteCoverage", {
        included: includedPortfolioCount,
        total: totalPortfolioCount,
      });
    }
    return t("dashboardOverview.metricIncomplete");
  }
  if (summary?.status === "NO_DATA") {
    return t("dashboardOverview.metricNoData");
  }
  return t("dashboardOverview.metricNotAvailable");
}

/**
 * Builds the "AUM Today" metric card. Renders an explicit not-available /
 * load-error state (value "—") rather than a zero when no valuation
 * snapshot exists yet for the selected scope.
 */
export function buildAumMetric(
  summary: ValuationSummaryDTO | null,
  hasError: boolean,
  t: TranslateFn,
  locale: DashboardLocale | string = "en",
): DashboardValuationMetric {
  const base = {
    id: "aum",
    label: t("dashboardOverview.metricAumLabel"),
    icon: "portfolio",
    tone: "primary" as DashboardOverviewMetricTone,
  };
  if (hasError) {
    return {
      ...base,
      value: t("common.notAvailable"),
      changeLabel: "",
      changeTone: "neutral",
      helperText: t("dashboardOverview.metricLoadError"),
    };
  }
  if (
    !summary ||
    summary.status !== "AVAILABLE" ||
    !summary.dataAvailable ||
    summary.aumToday === null
  ) {
    return {
      ...base,
      value: t("common.notAvailable"),
      changeLabel: "",
      changeTone: "neutral",
      helperText: unavailableValuationHelper(summary, t),
    };
  }
  return {
    ...base,
    value: formatMoneyCompact(parseDecimalOrNull(summary.aumToday), summary.currency),
    changeLabel: "",
    changeTone: "neutral",
    helperText: summary.asOf
      ? t("dashboardOverview.metricAsOf", {
          time: formatDashboardTime(summary.asOf, locale, "Asia/Bangkok"),
        })
      : t("dashboardOverview.metricReportingCurrency", { currency: summary.currency }),
  };
}

/**
 * Builds the "Today's P&L" metric card. Positive P&L renders a success
 * tone, negative renders danger, and the no-data state stays visually
 * distinct ("—", neutral) from a genuine zero P&L.
 */
export function buildPnlMetric(
  summary: ValuationSummaryDTO | null,
  hasError: boolean,
  t: TranslateFn,
  locale: DashboardLocale | string = "en",
): DashboardValuationMetric {
  const base = {
    id: "pnl",
    label: t("dashboardOverview.metricPnlLabel"),
    icon: "analysis",
  };
  if (hasError) {
    return {
      ...base,
      value: t("common.notAvailable"),
      changeLabel: "",
      changeTone: "neutral",
      helperText: t("dashboardOverview.metricLoadError"),
      tone: "success",
    };
  }
  if (
    !summary ||
    summary.status !== "AVAILABLE" ||
    !summary.dataAvailable ||
    summary.todayPnl === null
  ) {
    return {
      ...base,
      value: t("common.notAvailable"),
      changeLabel: "",
      changeTone: "neutral",
      helperText: unavailableValuationHelper(summary, t),
      tone: "success",
    };
  }
  const pnl = parseDecimalOrNull(summary.todayPnl) ?? 0;
  const pct = parseDecimalOrNull(summary.todayPnlPercent);
  const changeTone: DashboardOverviewChangeTone =
    pnl > 0 ? "success" : pnl < 0 ? "danger" : "neutral";
  return {
    ...base,
    value: formatMoneyCompact(pnl, summary.currency, true),
    changeLabel: pct === null ? "" : formatPercent(pct, 2, true),
    changeTone,
    helperText: summary.asOf
      ? t("dashboardOverview.metricAsOf", {
          time: formatDashboardTime(summary.asOf, locale, "Asia/Bangkok"),
        })
      : t("dashboardOverview.metricReportingCurrency", { currency: summary.currency }),
    tone: pnl < 0 ? "danger" : "success",
  };
}
