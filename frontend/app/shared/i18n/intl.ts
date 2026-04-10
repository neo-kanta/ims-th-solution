import type { Locale } from "./types";

const intlLocaleMap: Record<Locale, string> = {
  en: "en-US",
  th: "th-TH",
  zh: "zh-TW",
};

export function resolveIntlLocale(locale: Locale | string = "en"): string {
  return intlLocaleMap[locale as Locale] || locale;
}

export function createDateFormatter(
  locale: Locale | string,
  options: Intl.DateTimeFormatOptions,
  timeZone = "UTC",
) {
  return new Intl.DateTimeFormat(resolveIntlLocale(locale), {
    ...options,
    timeZone,
  });
}
