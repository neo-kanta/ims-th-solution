import { computed } from "vue";
import type { DashboardTabProvider } from "~/features/shell/tabs/types";
import { useComplianceRuleDirectory } from "../composables/useComplianceRuleDirectory";
import { useComplianceBreachesList } from "../composables/useComplianceBreaches";

export const complianceDashboardTabs: DashboardTabProvider = {
  id: "compliance",
  matches: (route) => route.path.startsWith("/compliance"),
  setup({ route, t }) {
    const ruleDir = useComplianceRuleDirectory();
    const breaches = useComplianceBreachesList();

    const items = computed(() => {
      const rulesCount = ruleDir.loading.value || ruleDir.error.value ? null : ruleDir.total.value;
      const breachesCount = breaches.loading.value || breaches.error.value ? null : breaches.total.value;
      return [
        { key: "overview", label: t("compliance.dashboard.tabs.overview"), to: "/compliance", icon: "list" },
        { key: "library", label: t("compliance.dashboard.tabs.library"), to: "/compliance/rules", icon: "list", count: rulesCount },
        { key: "approvals", label: t("compliance.dashboard.tabs.approvals"), to: "/compliance", icon: "approval" },
        { key: "breaches", label: t("compliance.dashboard.tabs.breaches"), to: "/compliance/post-trade", icon: "warning", count: breachesCount },
        { key: "audit", label: t("compliance.dashboard.tabs.audit"), to: "/compliance/audit", icon: "audit" },
      ];
    });

    const activeKey = computed(() => {
      if (route.path === "/compliance") return "overview";
      if (route.path.startsWith("/compliance/rules")) return "library";
      if (route.path.startsWith("/compliance/post-trade")) return "breaches";
      if (route.path.startsWith("/compliance/audit")) return "audit";
      return "";
    });

    async function onRouteEnter() {
      void ruleDir.ensureLoaded();
      void breaches.fetchList({ limit: 1, status: "OPEN" });
    }

    return { items, activeKey, ariaLabel: "Compliance section tabs", onRouteEnter };
  },
};
