import type { AppTranslationKey } from "../../shared/i18n/messages";
import type { TranslationParams } from "../../shared/i18n/types";

export interface NavigationItem {
  label: string;
  icon: string;
  to: string;
  requiredPermissions?: string[];
}

export interface NavigationSection {
  label: string;
  items: NavigationItem[];
}

export function buildDashboardNavigation(
  t: (
    key: AppTranslationKey,
    paramsOrFallback?: TranslationParams | string,
    fallback?: string,
  ) => string,
  hasPermission: (code: string) => boolean,
): NavigationSection[] {
  const sections: NavigationSection[] = [
    {
      label: t("dashboard.overview"),
      items: [
        { label: t("navigation.dashboard"), to: "/", icon: "dashboard" },
      ],
    },
    {
      label: t("navigation.settings"),
      items: [
        {
          label: t("navigation.settings"),
          to: "/settings",
          icon: "shield",
          requiredPermissions: [
            "IAM_USER_VIEW",
            "IAM_USER_CREATE",
            "IAM_USER_UPDATE",
            "IAM_USER_DEACTIVATE",
            "IAM_AUDIT_VIEW",
          ],
        },
      ],
    },
  ];

  return sections
    .map((section) => ({
      ...section,
      items: section.items.filter((item) => (
        !item.requiredPermissions?.length
        || item.requiredPermissions.some((code) => hasPermission(code))
      )),
    }))
    .filter((section) => section.items.length > 0);
}
