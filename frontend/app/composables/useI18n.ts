import { computed } from "vue";
import { useCookie, useState } from "#imports";
import {
  normalizeTranslateArgs,
  translateMessage,
  type Locale,
  type TranslationParams,
} from "../shared/i18n/core";
import { messages, type AppTranslationKey } from "../shared/i18n/messages";

export function useI18n() {
  const localeCookie = useCookie<Locale>("app_locale", {
    default: () => "en",
    sameSite: "lax",
    path: "/",
  });
  const localeState = useState<Locale>(
    "app-locale",
    () => localeCookie.value || "en",
  );

  const locale = computed({
    get: () => localeState.value,
    set: (value: Locale) => {
      localeState.value = value;
      localeCookie.value = value;
    },
  });

  function t(
    key: AppTranslationKey,
    paramsOrFallback?: TranslationParams | string,
    fallback?: string,
  ): string {
    const normalized = normalizeTranslateArgs(paramsOrFallback, fallback);

    return translateMessage({
      fallback: normalized.fallback,
      key,
      locale: locale.value,
      messages,
      params: normalized.params,
    });
  }

  return {
    t,
    locale,
    setLocale: (newLocale: Locale) => {
      locale.value = newLocale;
    },
    availableLocales: ["en", "th", "zh"] as const,
  };
}

export type { Locale, TranslationParams, AppTranslationKey };
