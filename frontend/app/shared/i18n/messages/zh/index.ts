import { zhApprovalMessages } from "./approval";
import { zhWatchlistMessages } from "./watchlist";
import { zhChatMessages } from "./chat";
import { zhCommonMessages } from "./common";
import { zhComplianceMessages } from "./compliance";
import { zhDashboardMessages } from "./dashboard";
import { zhErdMessages } from "./erd";
import { zhInvestmentResearchMessages } from "./investmentResearch";
import { zhMarketDataMessages } from "./marketData";
import { zhPlaceholderMessages } from "./placeholders";
import { zhPortfolioMessages } from "./portfolio";
import { zhOperatorMessages } from "./operator";
import { zhSettingsMessages } from "./settings";

export const zhMessages = {
  ...zhApprovalMessages,
  ...zhChatMessages,
  ...zhCommonMessages,
  ...zhDashboardMessages,
  ...zhSettingsMessages,
  ...zhErdMessages,
  ...zhPlaceholderMessages,
  ...zhInvestmentResearchMessages,
  ...zhComplianceMessages,
  ...zhMarketDataMessages,
  ...zhWatchlistMessages,
  ...zhPortfolioMessages,
  ...zhOperatorMessages,
} as const;
