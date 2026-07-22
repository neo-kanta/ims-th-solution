import { computed } from "vue";
import type { DashboardTabProvider } from "~/features/shell/tabs/types";

export const portfolioWorkspaceDashboardTabs: DashboardTabProvider = {
  id: "portfolio-workspace",
  matches: (route) =>
    route.path.startsWith("/portfolios") && !!route.params.portfolioCode,
  setup({ route, t }) {
    const items = computed(() => {
      const code = encodeURIComponent(String(route.params.portfolioCode || ""));
      return [
        {
          key: "overview",
          label: t("portfolio.workspaceTabs.overview"),
          to: `/portfolios/${code}/overview`,
          icon: "globe",
        },
        {
          key: "holdings",
          label: t("portfolio.workspaceTabs.holdings"),
          to: `/portfolios/${code}/holdings`,
          icon: "portfolio",
        },
        {
          key: "cash",
          label: t("portfolio.workspaceTabs.cash"),
          to: `/portfolios/${code}/cash`,
          icon: "compliance",
        },
        {
          key: "ledger",
          label: t("portfolio.workspaceTabs.ledger"),
          to: `/portfolios/${code}/ledger`,
          icon: "decision",
        },
        {
          key: "decisions",
          label: t("portfolio.workspaceTabs.decisions"),
          to: `/portfolios/${code}/decisions`,
          icon: "decision",
        },
        {
          key: "executions",
          label: t("portfolio.workspaceTabs.executions"),
          to: `/portfolios/${code}/executions`,
          icon: "execution",
        },
        {
          key: "confirmations",
          label: t("portfolio.workspaceTabs.confirmations"),
          to: `/portfolios/${code}/confirmations`,
          icon: "execution",
        },
        {
          key: "compliance",
          label: t("portfolio.workspaceTabs.compliance"),
          to: `/portfolios/${code}/compliance`,
          icon: "compliance",
        },
        {
          key: "valuations",
          label: t("portfolio.workspaceTabs.valuations"),
          to: `/portfolios/${code}/valuations`,
          icon: "trend-up",
        },
        {
          key: "watchlists",
          label: t("portfolio.workspaceTabs.watchlists"),
          to: `/portfolios/${code}/watchlists`,
          icon: "list",
        },
        {
          key: "settings",
          label: t("portfolio.workspaceTabs.settings"),
          to: `/portfolios/${code}/settings`,
          icon: "system",
        },
      ];
    });

    const activeKey = computed(() => {
      const parts = route.path.split("/");
      return parts[3] || "overview";
    });

    return { items, activeKey, ariaLabel: "Portfolio workspace tabs" };
  },
};
