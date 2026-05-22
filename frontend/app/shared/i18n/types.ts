export type Locale = "en" | "th" | "zh";

export type MessageTree = {
  [key: string]: string | MessageTree;
};

/**
 * Recursively widens literal string types to `string`, preserving the
 * tree shape. Used by MessageCatalog so non-English locales — which by
 * definition contain different literal values from the English reference
 * tree — still satisfy the catalog constraint as long as their keys and
 * tree shape match.
 */
type WidenStrings<T> = T extends string
  ? string
  : T extends MessageTree
    ? { [K in keyof T]: WidenStrings<T[K]> }
    : never;

export type MessageCatalog<TMessages extends MessageTree = MessageTree> = Record<
  Locale,
  WidenStrings<TMessages>
>;

export type TranslationParams = Record<
  string,
  string | number | boolean | null | undefined
>;

type Join<TKey extends string, TValue> = TValue extends string
  ? `${TKey}.${TValue}`
  : never;

export type TranslationKey<TMessages extends MessageTree> = {
  [TKey in Extract<keyof TMessages, string>]: TMessages[TKey] extends string
    ? TKey
    : TMessages[TKey] extends MessageTree
      ? Join<TKey, TranslationKey<TMessages[TKey]>>
      : never;
}[Extract<keyof TMessages, string>];
