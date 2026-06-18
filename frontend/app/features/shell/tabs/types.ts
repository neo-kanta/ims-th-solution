import type { TabItem } from "~/shared/ui/AppTabs.vue";
import type { RouteLocationNormalizedLoaded } from "vue-router";
import type { ComputedRef } from "vue";
import type { AppTranslationKey, TranslationParams } from "~/composables/useI18n";

export type { TabItem };

export type DashboardTabContext = {
  route: RouteLocationNormalizedLoaded;
  t: (key: AppTranslationKey, paramsOrFallback?: TranslationParams | string, fallback?: string) => string;
};

export type DashboardTabSetup = {
  items: ComputedRef<TabItem[]>;
  activeKey: ComputedRef<string>;
  ariaLabel: string;
  onRouteEnter?: () => void | Promise<void>;
};

export type DashboardTabProvider = {
  id: string;
  matches: (route: RouteLocationNormalizedLoaded) => boolean;
  setup: (ctx: DashboardTabContext) => DashboardTabSetup;
};
