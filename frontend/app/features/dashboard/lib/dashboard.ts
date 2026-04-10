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

export function formatDashboardWorkflowDate(
  dateString: string,
  locale: DashboardLocale | string = "en",
): string {
  return createDateFormatter(locale, {
    month: "short",
    day: "numeric",
    year: "numeric",
  }, "UTC").format(new Date(dateString));
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

export function buildDashboardSnapshot(
  now: number,
  permissions: UserPermissions,
  t: TranslateFn,
): DashboardPayload {
  const businessDate = new Date(now).toISOString().split("T")[0] || "";
  const duePrefix = t("dashboardOverview.duePrefix", "Due");

  const workflow = {
    businessDate,
    dayStatus: t("dashboardOverview.dayStatusOpen", "Open"),
    stages: [
      {
        id: "stage-1",
        sequence: 1,
        label: t("dashboardOverview.workflowDayStart", "Day Start"),
        timeLabel: "08:30",
        status: "complete" as const,
      },
      {
        id: "stage-2",
        sequence: 2,
        label: t("dashboardOverview.workflowAnalysis", "Analysis"),
        timeLabel: "09:15",
        status: "complete" as const,
      },
      {
        id: "stage-3",
        sequence: 3,
        label: t("dashboardOverview.workflowDecision", "Decision"),
        timeLabel: "--",
        status: "active" as const,
      },
      {
        id: "stage-4",
        sequence: 4,
        label: t(
          "dashboardOverview.workflowManagerApproval",
          "Mgr Approval",
        ),
        timeLabel: "--",
        status: "upcoming" as const,
      },
      {
        id: "stage-5",
        sequence: 5,
        label: t("dashboardOverview.workflowExecution", "Execution"),
        timeLabel: "--",
        status: "upcoming" as const,
      },
      {
        id: "stage-6",
        sequence: 6,
        label: t("dashboardOverview.workflowTxClosing", "Tx Closing"),
        timeLabel: "--",
        status: "upcoming" as const,
      },
      {
        id: "stage-7",
        sequence: 7,
        label: t(
          "dashboardOverview.workflowAccountingClosing",
          "Acctg Closing",
        ),
        timeLabel: "--",
        status: "upcoming" as const,
      },
    ],
  };

  const contracts = [
    {
      id: "contract-1",
      code: "KBANK-EQ-001",
      assetType: t("dashboardOverview.assetEquity", "Equity"),
      valueLabel: "B 12,500,000",
      statusLabel: t("dashboardOverview.statusDecision", "Decision"),
      statusTone: "info" as const,
      manager: "Somchai T.",
      updatedAt: "14:30",
    },
    {
      id: "contract-2",
      code: "PTT-FI-002",
      assetType: t("dashboardOverview.assetFixedIncome", "Fixed Inc."),
      valueLabel: "B 8,200,000",
      statusLabel: t(
        "dashboardOverview.statusPendingApproval",
        "Pending App.",
      ),
      statusTone: "warning" as const,
      manager: "Wichit P.",
      updatedAt: "13:15",
    },
    {
      id: "contract-3",
      code: "ADVANC-EQ-003",
      assetType: t("dashboardOverview.assetEquity", "Equity"),
      valueLabel: "B 5,750,000",
      statusLabel: t("dashboardOverview.statusExecution", "Execution"),
      statusTone: "teal" as const,
      manager: "Naree S.",
      updatedAt: "11:45",
    },
    {
      id: "contract-4",
      code: "SCC-EQ-004",
      assetType: t("dashboardOverview.assetEquity", "Equity"),
      valueLabel: "B 22,000,000",
      statusLabel: t("dashboardOverview.statusApproved", "Approved"),
      statusTone: "success" as const,
      manager: "Somchai T.",
      updatedAt: "09:00",
    },
    {
      id: "contract-5",
      code: "TMB-FI-005",
      assetType: t("dashboardOverview.assetBond", "Bond"),
      valueLabel: "B 6,800,000",
      statusLabel: t("dashboardOverview.statusAnalysis", "Analysis"),
      statusTone: "purple" as const,
      manager: "Wichit P.",
      updatedAt: "08:45",
    },
  ];

  const pendingApprovals = [
    {
      id: "approval-1",
      contractCode: "KBANK-EQ-001",
      valueLabel: "B12.5M",
      dueLabel: `${duePrefix} 16:00`,
      isUrgent: true,
    },
    {
      id: "approval-2",
      contractCode: "PTT-FI-002",
      valueLabel: "B8.2M",
      dueLabel: `${duePrefix} 16:30`,
      isUrgent: false,
    },
    {
      id: "approval-3",
      contractCode: "SCC-EQ-004",
      valueLabel: "B22M",
      dueLabel: `${duePrefix} 15:45`,
      isUrgent: true,
    },
    {
      id: "approval-4",
      contractCode: "TMB-FI-005",
      valueLabel: "B6.8M",
      dueLabel: `${duePrefix} 17:00`,
      isUrgent: false,
    },
  ];

  const metrics = [
    {
      id: "metric-aum",
      label: t("dashboardOverview.metricAumLabel", "AUM Today"),
      value: "B 842.5M",
      icon: "briefcase",
      tone: "primary" as const,
      changeLabel: t("dashboardOverview.metricAumChange", "+2.4%"),
      changeTone: "success" as const,
      helperText: t(
        "dashboardOverview.metricAumHelper",
        "Across 12 contracts",
      ),
    },
    {
      id: "metric-contracts",
      label: t("dashboardOverview.metricContractsLabel", "Active Contracts"),
      value: "12",
      icon: "dashboard",
      tone: "info" as const,
      changeLabel: t("dashboardOverview.metricContractsChange", "+1"),
      changeTone: "success" as const,
      helperText: t(
        "dashboardOverview.metricContractsHelper",
        "3 pending decisions",
      ),
    },
    {
      id: "metric-approvals",
      label: t("dashboardOverview.metricApprovalsLabel", "Pending Approvals"),
      value: String(pendingApprovals.length),
      icon: "approval",
      tone: "danger" as const,
      changeLabel: t(
        "dashboardOverview.metricApprovalsChange",
        "Due before 16:30",
      ),
      changeTone: "danger" as const,
      helperText: t(
        "dashboardOverview.metricApprovalsHelper",
        "2 urgent approvals",
      ),
    },
    {
      id: "metric-pnl",
      label: t("dashboardOverview.metricPnlLabel", "Today's P&L"),
      value: "B +3.2M",
      icon: "trend-up",
      tone: "success" as const,
      changeLabel: t("dashboardOverview.metricPnlChange", "+0.38%"),
      changeTone: "success" as const,
      helperText: t("dashboardOverview.metricPnlHelper", "vs prior close"),
    },
  ];

  const activityFeed = [
    {
      id: "activity-1",
      actor: "Somchai T.",
      message: t(
        "dashboardOverview.activityApprovedAnalysis",
        "approved analysis for KBANK-EQ-001",
      ),
      timeLabel: "14:32",
      tone: "success" as const,
    },
    {
      id: "activity-2",
      actor: t("dashboardOverview.activitySystemActor", "System"),
      message: t(
        "dashboardOverview.activityAutoEscalated",
        "auto-escalated PTT-FI-002",
      ),
      timeLabel: "13:20",
      tone: "info" as const,
    },
    {
      id: "activity-3",
      actor: "Wichit P.",
      message: t(
        "dashboardOverview.activityRequestedDecision",
        "requested decision on ADVANC-EQ-003",
      ),
      timeLabel: "11:50",
      tone: "warning" as const,
    },
    {
      id: "activity-4",
      actor: "Naree S.",
      message: t(
        "dashboardOverview.activityExecutedTrade",
        "executed trade for SCC-EQ-004",
      ),
      timeLabel: "09:05",
      tone: "teal" as const,
    },
    {
      id: "activity-5",
      actor: t("dashboardOverview.activityIrgOpsActor", "IRG Ops"),
      message: t(
        "dashboardOverview.activityClearedOvernight",
        "cleared overnight controls",
      ),
      timeLabel: "08:40",
      tone: "neutral" as const,
    },
  ];

  return {
    workflow,
    metrics,
    contracts,
    activityFeed,
    pendingApprovals,
    lastRefresh: new Date(now).toISOString(),
    canCreateContract: canCreateContract(permissions),
    createContractUrl: "/investment/analysis",
  };
}
