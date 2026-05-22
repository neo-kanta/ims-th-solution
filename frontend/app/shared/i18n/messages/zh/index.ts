import { zhCommonMessages } from "./common";
import { zhComplianceMessages } from "./compliance";
import { zhDashboardMessages } from "./dashboard";
import { zhErdMessages } from "./erd";
import { zhHoldingsMessages } from "./holdings";
import { zhInvestmentResearchMessages } from "./investmentResearch";
import { zhPlaceholderMessages } from "./placeholders";
import { zhSettingsMessages } from "./settings";

export const zhMessages = {
  ...zhCommonMessages,
  ...zhDashboardMessages,
  ...zhSettingsMessages,
  ...zhErdMessages,
  ...zhPlaceholderMessages,
  ...zhHoldingsMessages,
  ...zhInvestmentResearchMessages,
  ...zhComplianceMessages,
} as const;

