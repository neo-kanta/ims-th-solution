import { thApprovalMessages } from "./approval";
import { thWatchlistMessages } from "./watchlist";
import { thChatMessages } from "./chat";
import { thCommonMessages } from "./common";
import { thComplianceMessages } from "./compliance";
import { thDashboardMessages } from "./dashboard";
import { thErdMessages } from "./erd";
import { thInvestmentResearchMessages } from "./investmentResearch";
import { thMarketDataMessages } from "./marketData";
import { thPlaceholderMessages } from "./placeholders";
import { thPortfolioMessages } from "./portfolio";
import { thSettingsMessages } from "./settings";

export const thMessages = {
  ...thApprovalMessages,
  ...thChatMessages,
  ...thCommonMessages,
  ...thDashboardMessages,
  ...thSettingsMessages,
  ...thErdMessages,
  ...thPlaceholderMessages,
  ...thInvestmentResearchMessages,
  ...thComplianceMessages,
  ...thMarketDataMessages,
  ...thWatchlistMessages,
  ...thPortfolioMessages,
} as const;
