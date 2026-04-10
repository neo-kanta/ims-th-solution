import { enMessages } from "./en";
import { thMessages } from "./th";
import { zhMessages } from "./zh";
import type { MessageCatalog, TranslationKey as TranslationKeyOf } from "../types";

export const messages = {
  en: enMessages,
  th: thMessages,
  zh: zhMessages,
} as const satisfies MessageCatalog<typeof enMessages>;

export type AppMessages = typeof enMessages;
export type AppTranslationKey = TranslationKeyOf<AppMessages>;
