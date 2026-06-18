import { thApprovalMessages } from "./approval";
import { thChatMessages } from "./chat";
import { thCommonMessages } from "./common";
import { thComplianceMessages } from "./compliance";
import { thDashboardMessages } from "./dashboard";
import { thErdMessages } from "./erd";
import { thFundsCreateMessages } from "./fundsCreate";
import { thHoldingsMessages } from "./holdings";
import { thInvestmentResearchMessages } from "./investmentResearch";
import { thMarketDataMessages } from "./marketData";
import { thMyFundsMessages } from "./myFunds";
import { thPlaceholderMessages } from "./placeholders";
import { thSettingsMessages } from "./settings";

export const thMessages = {
  ...thApprovalMessages,
  ...thChatMessages,
  ...thCommonMessages,
  ...thDashboardMessages,
  ...thSettingsMessages,
  ...thErdMessages,
  ...thPlaceholderMessages,
  ...thHoldingsMessages,
  ...thInvestmentResearchMessages,
  ...thComplianceMessages,
  ...thMarketDataMessages,
  ...thMyFundsMessages,
  ...thFundsCreateMessages,
} as const;
