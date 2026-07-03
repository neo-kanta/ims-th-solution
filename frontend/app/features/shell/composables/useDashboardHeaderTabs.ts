import { computed, watch } from "vue";
import { useRoute } from "#imports";
import { useI18n } from "~/composables/useI18n";
import { dashboardTabRegistry } from "../tabs/dashboardTabRegistry";
import type { DashboardTabSetup } from "../tabs/types";

type InitializedProvider = {
  matches: (route: ReturnType<typeof useRoute>) => boolean;
  setup: DashboardTabSetup;
};

export function useDashboardHeaderTabs() {
  const route = useRoute();
  const { t } = useI18n();

  // All providers are initialized eagerly during layout setup so composables
  // that rely on useState (compliance rule directory, etc.) are called while
  // the Nuxt component instance is active. Route-gated preloading happens
  // inside each provider's onRouteEnter.
  const providers: InitializedProvider[] = dashboardTabRegistry.map((provider) => ({
    matches: provider.matches,
    setup: provider.setup({ route, t }),
  }));

  const activeSetup = computed<DashboardTabSetup | null>(
    () => providers.find(({ matches }) => matches(route))?.setup ?? null,
  );

  const hasHeaderTabs = computed(() => activeSetup.value !== null);

  const activeTabItems = computed(() => {
    const s = activeSetup.value;
    return s ? s.items.value : [];
  });

  const activeTabValue = computed(() => {
    const s = activeSetup.value;
    return s ? s.activeKey.value : "";
  });

  const activeTabAriaLabel = computed(() => {
    const s = activeSetup.value;
    return s ? s.ariaLabel : "Tabs";
  });

  watch(
    () => route.path,
    () => {
      void activeSetup.value?.onRouteEnter?.();
    },
    { immediate: true },
  );

  return { hasHeaderTabs, activeTabItems, activeTabValue, activeTabAriaLabel };
}
