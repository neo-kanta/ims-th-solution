import { computed } from "vue";

import type { SettingsSectionId } from "../ui.types";

const VALID_SECTIONS: ReadonlySet<SettingsSectionId> = new Set([
  "overview",
  "personal-account",
  "users",
  "groups",
  "function-permissions",
  "data-permissions",
  "security-policy",
  "notifications",
  "audit",
]);

const DEFAULT_SECTION: SettingsSectionId = "personal-account";

function isValidSection(value: unknown): value is SettingsSectionId {
  return typeof value === "string" && VALID_SECTIONS.has(value as SettingsSectionId);
}

/**
 * Two-way binds the active settings section to the URL `?section=…` query
 * parameter so admins can deep-link, bookmark, and use browser back/forward.
 * Falls back to "personal-account" if the query is missing or invalid.
 */
export function useSettingsActiveSection() {
  const route = useRoute();
  const router = useRouter();

  const activeSection = computed<SettingsSectionId>({
    get() {
      const raw = route.query.section;
      const value = Array.isArray(raw) ? raw[0] : raw;
      return isValidSection(value) ? value : DEFAULT_SECTION;
    },
    set(next) {
      if (!isValidSection(next)) return;
      if (route.query.section === next) return;
      void router.replace({ query: { ...route.query, section: next } });
    },
  });

  return { activeSection };
}
