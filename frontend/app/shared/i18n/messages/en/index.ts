import { enCommonMessages } from "./common";
import { enDashboardMessages } from "./dashboard";
import { enPlaceholderMessages } from "./placeholders";
import { enSettingsMessages } from "./settings";

export const enMessages = {
  ...enCommonMessages,
  ...enDashboardMessages,
  ...enSettingsMessages,
  ...enPlaceholderMessages,
} as const;
