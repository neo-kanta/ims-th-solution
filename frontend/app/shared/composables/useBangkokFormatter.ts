import { computed } from "vue";

import { createDateFormatter } from "~/shared/i18n/intl";

/**
 * Returns a stable Asia/Bangkok medium+short formatter and a `formatDateTime`
 * helper that returns "—" for null/undefined and falls back to the raw value
 * if the input is not a parseable date. Replaces the per-component pattern
 * of building a fresh formatter in every Vue file.
 */
export function useBangkokFormatter(notAvailableLabel?: () => string) {
  const { t, locale } = useI18n();

  const bangkokFormatter = computed(() =>
    createDateFormatter(
      locale.value,
      {
        dateStyle: "medium",
        timeStyle: "short",
      },
      "Asia/Bangkok",
    ),
  );

  function formatDateTime(value?: string | null): string {
    if (!value) {
      return notAvailableLabel ? notAvailableLabel() : t("common.notAvailable");
    }

    const date = new Date(value);
    return Number.isNaN(date.getTime())
      ? value
      : bangkokFormatter.value.format(date);
  }

  return { bangkokFormatter, formatDateTime };
}
