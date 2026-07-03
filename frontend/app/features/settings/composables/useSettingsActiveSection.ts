import { computed } from "vue";

import type { SettingsSectionId } from "../ui.types";

const VALID_SECTIONS: readonly SettingsSectionId[] = [
  "overview",
  "personal-account",
  "users",
  "groups",
  "function-permissions",
  "data-permissions",
  "security-policy",
  "notifications",
  "audit",
];

const DEFAULT_SECTION: SettingsSectionId = "personal-account";

interface UseSettingsActiveSectionOptions {
  validSections?: readonly SettingsSectionId[];
  defaultSection?: SettingsSectionId;
}

/**
 * Two-way binds the active settings section to the URL `?section=...` query
 * parameter so admins can deep-link, bookmark, and use browser back/forward.
 */
export function useSettingsActiveSection(
  options: UseSettingsActiveSectionOptions = {},
) {
  const route = useRoute();
  const router = useRouter();
  const validSections = new Set(options.validSections ?? VALID_SECTIONS);
  const requestedDefault = options.defaultSection ?? DEFAULT_SECTION;
  const defaultSection = validSections.has(requestedDefault)
    ? requestedDefault
    : (options.validSections?.[0] ?? DEFAULT_SECTION);

  function isValidSection(value: unknown): value is SettingsSectionId {
    return typeof value === "string" && validSections.has(value as SettingsSectionId);
  }

  const activeSection = computed<SettingsSectionId>({
    get() {
      const raw = route.query.section;
      const value = Array.isArray(raw) ? raw[0] : raw;
      return isValidSection(value) ? value : defaultSection;
    },
    set(next) {
      if (!isValidSection(next)) return;
      if (route.query.section === next) return;
      void router.replace({ query: { ...route.query, section: next } });
    },
  });

  return { activeSection };
}
