import { computed } from "vue";
import type { DashboardTabProvider } from "~/features/shell/tabs/types";

export const investmentDecisionDashboardTabs: DashboardTabProvider = {
  id: "investment-decision",
  matches: (route) => route.path.startsWith("/investment/decision"),
  setup({ t }) {
    const items = computed(() => [
      {
        key: "approval",
        label: t("investmentDecision.tabs.approval", "Batch Approval"),
        to: "/investment/decision",
        icon: "decision",
      },
    ]);

    const activeKey = computed(() => "approval");

    return { items, activeKey, ariaLabel: "Investment decision tabs" };
  },
};
