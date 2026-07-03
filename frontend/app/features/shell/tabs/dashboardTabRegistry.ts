import type { DashboardTabProvider } from "./types";
import { approvalDashboardTabs } from "~/features/approval/navigation/dashboardTabs";
import { complianceDashboardTabs } from "~/features/compliance/navigation/dashboardTabs";
import { investmentWorkspaceDashboardTabs } from "~/features/investment-workspace/navigation/dashboardTabs";
import { marketDataDashboardTabs } from "~/features/market-data/navigation/dashboardTabs";
import { workflowDashboardTabs } from "~/features/workflow/navigation/dashboardTabs";

// To add tabs for a new feature: create a DashboardTabProvider in your feature's
// navigation/ folder and append it here. dashboard.vue never needs to change.
export const dashboardTabRegistry: DashboardTabProvider[] = [
  approvalDashboardTabs,
  complianceDashboardTabs,
  investmentWorkspaceDashboardTabs,
  marketDataDashboardTabs,
  workflowDashboardTabs,
];

