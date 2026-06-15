import { computed } from "vue";
import type { DashboardTabProvider } from "~/features/shell/tabs/types";

export const workflowDashboardTabs: DashboardTabProvider = {
  id: "workflow",
  matches: (route) => route.path.startsWith("/workflow"),
  setup({ route, t }) {
    const items = computed(() => [
      {
        key: "overview",
        label: t("workflow.tabs.overview", "Overview"),
        to: "/workflow",
        icon: "list",
      },
      {
        key: "audit",
        label: t("workflow.tabs.audit", "Audit Trail"),
        to: "/workflow#audit",
        icon: "audit",
      },
      {
        key: "settings",
        label: t("workflow.tabs.settings", "Approval Settings"),
        to: "/workflow#settings",
        icon: "shield",
      },
    ]);

    const activeKey = computed(() => {
      if (route.hash === "#settings") return "settings";
      if (route.hash === "#audit") return "audit";
      return "overview";
    });

    return {
      items,
      activeKey,
      ariaLabel: "Workflow section tabs",
    };
  },
};
export default workflowDashboardTabs;
