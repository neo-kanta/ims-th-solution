import type { AppTranslationKey } from "../../shared/i18n/messages";
import type { TranslationParams } from "../../shared/i18n/types";

export interface NavigationItem {
  label: string;
  icon: string;
  to: string;
  requiredPermissions?: string[];
}

export interface NavigationSection {
  label: string;
  icon: string;
  items: NavigationItem[];
}

export function buildDashboardNavigation(
  t: (
    key: AppTranslationKey,
    paramsOrFallback?: TranslationParams | string,
    fallback?: string,
  ) => string,
  hasPermission: (code: string) => boolean,
): NavigationSection[] {
  const sections: NavigationSection[] = [
    {
      label: t("dashboard.overview"),
      icon: "dashboard",
      items: [
        { label: t("navigation.dashboard"), to: "/", icon: "dashboard" },
        {
          label: t("navigation.workflow", "Workflow"),
          to: "/workflow",
          icon: "list",
          requiredPermissions: ["WORKFLOW_VIEW"],
        },
        { label: t("navigation.chat", "Chat"), to: "/chat", icon: "chat" },
      ],
    },
    {
      label: t("navigation.investment"),
      icon: "analysis",
      items: [
        {
          label: t("holdings.page.myFunds", "My funds"),
          to: "/investment/funds",
          icon: "analysis",
          requiredPermissions: ["INVESTMENT_FUND_VIEW"],
        },
        {
          label: t("navigation.investmentResearch"),
          to: "/investment/analysis",
          icon: "analysis",
          requiredPermissions: ["INVESTMENT_RESEARCH_VIEW"],
        },
        {
          label: t("navigation.marketData", "Market data"),
          to: "/market-data",
          icon: "globe",
        },
        {
          label: t("navigation.watchlist", "Watchlist"),
          to: "/watchlist",
          icon: "notifications",
          requiredPermissions: ["WATCHLIST_VIEW"],
        },
      ],
    },
    {
      label: t("navigation.operations", "Operations"),
      icon: "decision",
      items: [
        {
          label: t("navigation.operatorPage", "Operator page"),
          to: "/investment/operator",
          icon: "decision",
          requiredPermissions: ["INVESTMENT_VIEW"],
        },
      ],
    },
    {
      label: t("compliance.nav.section", "Compliance"),
      icon: "compliance",
      items: [
        {
          label: t("compliance.nav.dashboard", "Compliance dashboard"),
          to: "/compliance",
          icon: "compliance",
          requiredPermissions: ["IRG_VIEW_RULES"],
        },
        {
          label: t("compliance.nav.rules", "Rule library"),
          to: "/compliance/rules",
          icon: "list",
          requiredPermissions: ["IRG_VIEW_RULES"],
        },
        {
          label: t("compliance.nav.preTrade", "Pre-trade simulator"),
          to: "/compliance/pre-trade",
          icon: "shield",
          requiredPermissions: ["IRG_VIEW_RULES"],
        },
        {
          label: t("compliance.nav.postTrade", "Post-trade breaches"),
          to: "/compliance/post-trade",
          icon: "warning",
          requiredPermissions: ["IRG_VIEW_RULES"],
        },
        {
          label: t("compliance.nav.exceptions", "Pre-trade exceptions"),
          to: "/compliance/exceptions",
          icon: "approval",
          requiredPermissions: ["IRG_VIEW_RULES"],
        },
        {
          label: t("compliance.nav.audit", "Audit trail"),
          to: "/compliance/audit",
          icon: "audit",
          requiredPermissions: ["IRG_VIEW_RULES"],
        },
        {
          label: t("compliance.nav.permissions", "Permission matrix"),
          to: "/compliance/permissions",
          icon: "shield",
          requiredPermissions: ["IRG_VIEW_RULES"],
        },
      ],
    },
    {
      label: t("approval.nav.section", "Approval"),
      icon: "approval",
      items: [
        {
          label: t("approval.nav.inbox", "Approval inbox"),
          to: "/approval",
          icon: "approval",
          requiredPermissions: ["APPROVAL_VIEW_INBOX"],
        },
        {
          label: t("approval.nav.processes", "Approval processes"),
          to: "/approval/config/processes",
          icon: "list",
          requiredPermissions: ["APPROVAL_CONFIG_VIEW"],
        },
        {
          label: t("approval.nav.groups", "Approval groups"),
          to: "/approval/config/groups",
          icon: "groups",
          requiredPermissions: ["APPROVAL_CONFIG_VIEW"],
        },
        {
          label: t("approval.nav.teams", "Approval teams"),
          to: "/approval/config/teams",
          icon: "groups",
          requiredPermissions: ["APPROVAL_CONFIG_VIEW"],
        },
      ],
    },
    {
      label: t("navigation.notifications", "Notifications"),
      icon: "notifications",
      items: [
        {
          label: t("navigation.notificationCenter", "My notifications"),
          to: "/notifications",
          icon: "notifications",
        },
        {
          label: t("navigation.emailOperations", "Email operations"),
          to: "/admin/notifications/email",
          icon: "send",
          requiredPermissions: ["NOTIFICATION_VIEW", "NOTIFICATION_HEALTH"],
        },
        {
          label: t("navigation.emailOutbox", "Email outbox"),
          to: "/admin/notifications/email/outbox",
          icon: "list",
          requiredPermissions: ["NOTIFICATION_VIEW"],
        },
      ],
    },
    {
      label: t("navigation.administration"),
      icon: "shield",
      items: [
        {
          label: "Permission requests",
          to: "/permissions/change-requests",
          icon: "approval",
          requiredPermissions: ["permission.change_request.review"],
        },
        {
          label: "Accounts",
          to: "/permissions/accounts",
          icon: "accounts",
          requiredPermissions: ["permission.users.view"],
        },
        {
          label: "Groups",
          to: "/permissions/groups",
          icon: "groups",
          requiredPermissions: ["permission.groups.view"],
        },
        {
          label: "Role hierarchy",
          to: "/permissions/roles",
          icon: "groups",
          requiredPermissions: ["permission.roles.view"],
        },
        {
          label: "Effective permissions",
          to: "/permissions/effective",
          icon: "shield",
          requiredPermissions: ["permission.users.view"],
        },
        {
          label: t("navigation.administrationSettings"),
          to: "/administration/settings",
          icon: "shield",
          requiredPermissions: [
            "IAM_USER_VIEW",
            "IAM_USER_CREATE",
            "IAM_USER_UPDATE",
            "IAM_USER_DEACTIVATE",
            "IAM_AUDIT_VIEW",
          ],
        },
      ],
    },
  ];

  return sections
    .map((section) => ({
      ...section,
      items: section.items.filter((item) => (
        !item.requiredPermissions?.length
        || item.requiredPermissions.some((code) => hasPermission(code))
      )),
    }))
    .filter((section) => section.items.length > 0);
}
