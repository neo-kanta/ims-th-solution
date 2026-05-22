import { thCommonMessages } from "./common";
import { thComplianceMessages } from "./compliance";
import { thDashboardMessages } from "./dashboard";
import { thErdMessages } from "./erd";
import { thHoldingsMessages } from "./holdings";
import { thInvestmentResearchMessages } from "./investmentResearch";
import { thPlaceholderMessages } from "./placeholders";
import { thSettingsMessages } from "./settings";

export const thMessages = {
  ...thCommonMessages,
  ...thDashboardMessages,
  ...thSettingsMessages,
  ...thErdMessages,
  ...thPlaceholderMessages,
  ...thHoldingsMessages,
  ...thInvestmentResearchMessages,
  ...thComplianceMessages,
} as const;

