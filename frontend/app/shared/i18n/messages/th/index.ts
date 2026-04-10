import { thCommonMessages } from "./common";
import { thDashboardMessages } from "./dashboard";
import { thPlaceholderMessages } from "./placeholders";
import { thSettingsMessages } from "./settings";

export const thMessages = {
  ...thCommonMessages,
  ...thDashboardMessages,
  ...thSettingsMessages,
  ...thPlaceholderMessages,
} as const;
