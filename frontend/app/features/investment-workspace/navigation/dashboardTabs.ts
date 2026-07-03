import { computed } from "vue";
import type { DashboardTabProvider } from "~/features/shell/tabs/types";

export const investmentWorkspaceDashboardTabs: DashboardTabProvider = {
  id: "investment-workspace",
  matches: (route) => route.path.startsWith("/investment") && !!route.params.fundId,
  setup({ route, t }) {
    const items = computed(() => {
      const fId = String(route.params.fundId || "");
      return [
        { key: "holdings", label: t("holdings.workspaceTabs.holdings"), to: `/investment/funds/${fId}/holdings`, icon: "portfolio" },
        { key: "operation", label: t("holdings.workspaceTabs.operation", "Operation"), to: `/investment/funds/${fId}/operation`, icon: "decision" },
        { key: "stages", label: t("holdings.workspaceTabs.stages"), to: `/investment/funds/${fId}/stages`, icon: "globe" },
        { key: "decisions", label: t("holdings.workspaceTabs.decisions"), to: `/investment/funds/${fId}/decisions`, icon: "decision" },
        { key: "compliance", label: t("holdings.workspaceTabs.compliance"), to: `/investment/funds/${fId}/compliance`, icon: "compliance" },
        { key: "audit", label: t("holdings.workspaceTabs.audit"), to: `/investment/funds/${fId}/audit`, icon: "audit" },
        { key: "reviewers", label: t("holdings.workspaceTabs.reviewers"), to: `/investment/funds/${fId}/reviewers`, icon: "groups" },
        { key: "settings", label: t("holdings.workspaceTabs.settings"), to: `/investment/funds/${fId}/settings`, icon: "shield" },
      ];
    });

    const activeKey = computed(() => {
      const parts = route.path.split("/");
      return parts[4] || "holdings";
    });

    return { items, activeKey, ariaLabel: "Investment workspace tabs" };
  },
};
