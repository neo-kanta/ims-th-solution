import type { Locale, MessageCatalog, MessageTree, TranslationParams } from "./types";

export type { Locale, MessageCatalog, MessageTree, TranslationParams };

export interface TranslateOptions {
  fallback?: string;
  locale: Locale;
  messages: MessageCatalog;
  key: string;
  params?: TranslationParams;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

export function resolveMessage(
  localeMessages: MessageTree,
  key: string,
): string | undefined {
  const resolved = key.split(".").reduce<unknown>((value, segment) => {
    if (!isRecord(value)) {
      return undefined;
    }

    return value[segment];
  }, localeMessages);

  return typeof resolved === "string" ? resolved : undefined;
}

export function interpolateMessage(
  template: string,
  params: TranslationParams = {},
): string {
  return template.replace(/\{(\w+)\}/g, (match, key: string) => {
    const value = params[key];

    return value === undefined || value === null ? match : String(value);
  });
}

export function normalizeTranslateArgs(
  paramsOrFallback?: TranslationParams | string,
  fallback?: string,
): { fallback?: string; params?: TranslationParams } {
  if (typeof paramsOrFallback === "string") {
    return {
      fallback: paramsOrFallback,
      params: undefined,
    };
  }

  return {
    fallback,
    params: paramsOrFallback,
  };
}

export function translateMessage({
  fallback,
  key,
  locale,
  messages,
  params,
}: TranslateOptions): string {
  const localized = resolveMessage(messages[locale], key);
  const englishFallback =
    locale === "en" ? undefined : resolveMessage(messages.en, key);
  const template = localized ?? englishFallback ?? fallback ?? key;

  return interpolateMessage(template, params);
}

export function collectMessageKeys(
  localeMessages: MessageTree,
  prefix = "",
): string[] {
  return Object.entries(localeMessages).flatMap(([key, value]) => {
    const nextKey = prefix ? `${prefix}.${key}` : key;

    if (typeof value === "string") {
      return [nextKey];
    }

    if (isRecord(value)) {
      return collectMessageKeys(value as MessageTree, nextKey);
    }

    return [];
  });
}
