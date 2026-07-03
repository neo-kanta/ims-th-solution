<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppTabs, { type TabItem } from "~/shared/ui/AppTabs.vue";

interface TabCount {
  /** null means "not available" — UI renders nothing in the badge slot. */
  value: number | null;
  unavailableReason?: string;
}

interface Props {
  active:
    | "overview"
    | "library"
    | "approvals"
    | "breaches"
    | "exceptions"
    | "audit"
    | "settings";
  counts: {
    library: TabCount;
    approvals: TabCount;
    breaches: TabCount;
    exceptions: TabCount;
    audit: TabCount;
  };
}

const props = defineProps<Props>();

const { t } = useI18n();

interface TabDef {
  key: Props["active"];
  to: string;
  labelKey: string;
  icon: string;
  count?: TabCount;
}

const tabs = computed<TabDef[]>(() => [
  {
    key: "overview",
    to: "/compliance",
    labelKey: "compliance.dashboard.tabs.overview",
    icon: "list",
  },
  {
    key: "library",
    to: "/compliance/rules",
    labelKey: "compliance.dashboard.tabs.library",
    icon: "list",
    count: props.counts.library,
  },
  {
    key: "approvals",
    to: "/compliance",
    labelKey: "compliance.dashboard.tabs.approvals",
    icon: "approval",
    count: props.counts.approvals,
  },
  {
    key: "breaches",
    to: "/compliance/post-trade",
    labelKey: "compliance.dashboard.tabs.breaches",
    icon: "warning",
    count: props.counts.breaches,
  },
  {
    key: "exceptions",
    to: "/compliance/exceptions",
    labelKey: "compliance.dashboard.tabs.exceptions",
    icon: "approval",
    count: props.counts.exceptions,
  },
  {
    key: "audit",
    to: "/compliance/audit",
    labelKey: "compliance.dashboard.tabs.audit",
    icon: "audit",
    count: props.counts.audit,
  },
  {
    key: "settings",
    to: "/compliance/permissions",
    labelKey: "compliance.dashboard.tabs.settings",
    icon: "shield",
  },
]);

const tabItems = computed<TabItem[]>(() =>
  tabs.value.map((tab) => ({
    key: tab.key,
    label: t(tab.labelKey as any),
    to: tab.to,
    icon: tab.icon,
    count: tab.count
      ? tab.count.value === null
        ? "—"
        : tab.count.value
      : null,
  })),
);
</script>

<template>
  <AppTabs :items="tabItems" :model-value="active" aria-label="Compliance sections" />
</template>

<style scoped>
/* Scoped styles are no longer needed as AppTabs handles styling internally,
   but we keep a container wrapper/margin if needed. Currently, AppTabs handles borders. */
</style>
