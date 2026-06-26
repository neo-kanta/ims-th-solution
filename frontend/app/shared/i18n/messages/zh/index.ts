import { zhApprovalMessages } from "./approval";
import { zhWatchlistMessages } from "./watchlist";
import { zhChatMessages } from "./chat";
import { zhCommonMessages } from "./common";
import { zhComplianceMessages } from "./compliance";
import { zhDashboardMessages } from "./dashboard";
import { zhErdMessages } from "./erd";
import { zhFundsCreateMessages } from "./fundsCreate";
import { zhHoldingsMessages } from "./holdings";
import { zhInvestmentResearchMessages } from "./investmentResearch";
import { zhMarketDataMessages } from "./marketData";
import { zhMyFundsMessages } from "./myFunds";
import { zhPlaceholderMessages } from "./placeholders";
import { zhSettingsMessages } from "./settings";

export const zhMessages = {
  ...zhApprovalMessages,
  ...zhChatMessages,
  ...zhCommonMessages,
  ...zhDashboardMessages,
  ...zhSettingsMessages,
  ...zhErdMessages,
  ...zhPlaceholderMessages,
  ...zhHoldingsMessages,
  ...zhInvestmentResearchMessages,
  ...zhComplianceMessages,
  ...zhMarketDataMessages,
  ...zhMyFundsMessages,
  ...zhFundsCreateMessages,
  ...zhWatchlistMessages,
} as const;
