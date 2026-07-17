import { enApprovalMessages } from "./approval";
import { enWatchlistMessages } from "./watchlist";
import { enChatMessages } from "./chat";
import { enCommonMessages } from "./common";
import { enComplianceMessages } from "./compliance";
import { enDashboardMessages } from "./dashboard";
import { enErdMessages } from "./erd";
import { enInvestmentResearchMessages } from "./investmentResearch";
import { enMarketDataMessages } from "./marketData";
import { enPlaceholderMessages } from "./placeholders";
import { enPortfolioMessages } from "./portfolio";
import { enOperatorMessages } from "./operator";
import { enSettingsMessages } from "./settings";

export const enMessages = {
  ...enApprovalMessages,
  ...enChatMessages,
  ...enCommonMessages,
  ...enDashboardMessages,
  ...enSettingsMessages,
  ...enErdMessages,
  ...enPlaceholderMessages,
  ...enInvestmentResearchMessages,
  ...enComplianceMessages,
  ...enMarketDataMessages,
  ...enWatchlistMessages,
  ...enPortfolioMessages,
  ...enOperatorMessages,
} as const;
