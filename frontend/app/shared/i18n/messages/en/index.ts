import { enCommonMessages } from "./common";
import { enComplianceMessages } from "./compliance";
import { enDashboardMessages } from "./dashboard";
import { enErdMessages } from "./erd";
import { enFundsCreateMessages } from "./fundsCreate";
import { enHoldingsMessages } from "./holdings";
import { enInvestmentResearchMessages } from "./investmentResearch";
import { enMarketDataMessages } from "./marketData";
import { enMyFundsMessages } from "./myFunds";
import { enPlaceholderMessages } from "./placeholders";
import { enSettingsMessages } from "./settings";

export const enMessages = {
  ...enCommonMessages,
  ...enDashboardMessages,
  ...enSettingsMessages,
  ...enErdMessages,
  ...enPlaceholderMessages,
  ...enHoldingsMessages,
  ...enInvestmentResearchMessages,
  ...enComplianceMessages,
  ...enMarketDataMessages,
  ...enMyFundsMessages,
  ...enFundsCreateMessages,
} as const;
