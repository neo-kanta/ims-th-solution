import { computed } from "vue";
import type { DashboardTabProvider } from "~/features/shell/tabs/types";

export const approvalDashboardTabs: DashboardTabProvider = {
  id: "approval",
  matches: (route) => route.path.startsWith("/approval"),
  setup({ route, t }) {
    const items = computed(() => [
      {
        key: "inbox",
        label: t("approval.tabs.inbox", "Inbox"),
        to: "/approval",
        icon: "approval",
      },
      {
        key: "processes",
        label: t("approval.tabs.processes", "Processes"),
        to: "/approval/config/processes",
        icon: "workflow",
      },
      {
        key: "groups",
        label: t("approval.tabs.groups", "Groups"),
        to: "/approval/config/groups",
        icon: "groups",
      },
      {
        key: "teams",
        label: t("approval.tabs.teams", "Teams"),
        to: "/approval/config/teams",
        icon: "accounts",
      },
    ]);

    const activeKey = computed(() => {
      const path = route.path;
      if (path === "/approval" || path.startsWith("/approval/requests/")) return "inbox";
      if (path.startsWith("/approval/config/processes")) return "processes";
      if (path.startsWith("/approval/config/groups")) return "groups";
      if (path.startsWith("/approval/config/teams")) return "teams";
      return "inbox";
    });

    return {
      items,
      activeKey,
      ariaLabel: "Approval section tabs",
    };
  },
};

export default approvalDashboardTabs;
