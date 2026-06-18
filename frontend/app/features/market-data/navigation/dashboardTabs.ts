import { computed } from "vue";
import type { DashboardTabProvider } from "~/features/shell/tabs/types";
import { useMarketDataCatalog } from "../composables/useMarketDataCatalog";

const VALID_TABS = ["market", "watchlist", "unmapped", "settings"] as const;
type MarketDataTab = (typeof VALID_TABS)[number];

export const marketDataDashboardTabs: DashboardTabProvider = {
  id: "market-data",
  matches: (route) => route.path.startsWith("/market-data"),
  setup({ route, t }) {
    const { candidates, refreshCandidates } = useMarketDataCatalog();

    const items = computed(() => {
      const count = candidates.value.length || null;
      return [
        { key: "market", label: t("marketData.tabs.market"), to: "/market-data?tab=market", icon: "briefcase" },
        { key: "watchlist", label: t("marketData.tabs.watchlist"), to: "/market-data?tab=watchlist", icon: "list" },
        { key: "unmapped", label: t("marketData.tabs.stage"), to: "/market-data?tab=unmapped", icon: "warning", count },
        { key: "settings", label: t("marketData.tabs.settings"), to: "/market-data?tab=settings", icon: "shield" },
      ];
    });

    const activeKey = computed(() => {
      const tab = typeof route.query.tab === "string" ? route.query.tab : "";
      return (VALID_TABS as readonly string[]).includes(tab) ? (tab as MarketDataTab) : "market";
    });

    async function onRouteEnter() {
      void refreshCandidates({ status: "REVIEW_REQUIRED", limit: 100 });
    }

    return { items, activeKey, ariaLabel: "Market Data section tabs", onRouteEnter };
  },
};
