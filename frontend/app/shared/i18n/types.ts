export type Locale = "en" | "th" | "zh";

export type MessageTree = {
  [key: string]: string | MessageTree;
};

export type MessageCatalog<TMessages extends MessageTree = MessageTree> = Record<
  Locale,
  TMessages
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
