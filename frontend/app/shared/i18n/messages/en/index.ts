import { enCommonMessages } from "./common";
import { enComplianceMessages } from "./compliance";
import { enDashboardMessages } from "./dashboard";
import { enErdMessages } from "./erd";
import { enHoldingsMessages } from "./holdings";
import { enInvestmentResearchMessages } from "./investmentResearch";
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
} as const;

