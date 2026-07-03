import type { AppTranslationKey } from "../../../shared/i18n/messages";
import type { TranslationParams } from "../../../shared/i18n/types";
import type { UserPermissions } from "../../auth/types";
import type { DashboardPayload, QuickAction } from "../types";
import { createDateFormatter, resolveIntlLocale } from "../../../shared/i18n/intl";

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
): string {
  return createDateFormatter(locale, {
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  }, "UTC").format(new Date(timestamp));
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
    createContractUrl: "/investment/funds",
  };
}
