import { computed } from "vue";
import type { DashboardTabProvider } from "~/features/shell/tabs/types";

export const investmentDecisionDashboardTabs: DashboardTabProvider = {
  id: "investment-decision",
  matches: (route) => route.path.startsWith("/investment/funds"),
  setup() {
    const items = computed(() => [
      {
        key: "operation",
        label: "Operation",
        to: "#",
        icon: "decision",
      },
    ]);

    const activeKey = computed(() => "operation");

    return { items, activeKey, ariaLabel: "Investment decision tabs" };
  },
};
