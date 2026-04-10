import { zhCommonMessages } from "./common";
import { zhDashboardMessages } from "./dashboard";
import { zhPlaceholderMessages } from "./placeholders";
import { zhSettingsMessages } from "./settings";

export const zhMessages = {
  ...zhCommonMessages,
  ...zhDashboardMessages,
  ...zhSettingsMessages,
  ...zhPlaceholderMessages,
} as const;
