<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import type { FundHeader, WorkspaceTab } from "../types";

const props = defineProps<{
  fundId: string;
  active: WorkspaceTab;
  counts: FundHeader["tab_counts"];
}>();

const { t } = useI18n();

interface TabSpec {
  key: WorkspaceTab;
  label: string;
  icon: string;
  to: string;
  count?: number;
}

const tabs = computed<TabSpec[]>(() => [
  {
    key: "holdings",
    label: t("holdings.workspaceTabs.holdings", "Holdings"),
    icon: "📦",
    to: `/investment/funds/${props.fundId}/holdings`,
    count: props.counts.holdings,
  },
  {
    key: "stages",
    label: t("holdings.workspaceTabs.stages", "Stages"),
    icon: "🧭",
    to: `/investment/funds/${props.fundId}/stages`,
    count: props.counts.stages,
  },
  {
    key: "decisions",
    label: t("holdings.workspaceTabs.decisions", "Decisions"),
    icon: "🧠",
    to: `/investment/funds/${props.fundId}/decisions`,
    count: props.counts.decisions,
  },
  {
    key: "compliance",
    label: t("holdings.workspaceTabs.compliance", "Compliance"),
    icon: "🛡️",
    to: `/investment/funds/${props.fundId}/compliance`,
    count: props.counts.compliance,
  },
  {
    key: "audit",
    label: t("holdings.workspaceTabs.audit", "Audit"),
    icon: "📝",
    to: `/investment/funds/${props.fundId}/audit`,
    count: props.counts.audit,
  },
  {
    key: "reviewers",
    label: t("holdings.workspaceTabs.reviewers", "Reviewers"),
    icon: "👥",
    to: `/investment/funds/${props.fundId}/reviewers`,
    count: props.counts.reviewers,
  },
  {
    key: "settings",
    label: t("holdings.workspaceTabs.settings", "Settings"),
    icon: "⚙️",
    to: `/investment/funds/${props.fundId}/settings`,
    count: props.counts.settings,
  },
]);
</script>

<template>
  <nav class="fw-tabs" :aria-label="t('holdings.workspaceTabs.holdings', 'Workspace tabs')">
    <NuxtLink
      v-for="tab in tabs"
      :key="tab.key"
      :to="tab.to"
      class="fw-tabs__item"
      :class="{ 'is-active': active === tab.key }"
    >
      <span class="fw-tabs__icon" aria-hidden="true">{{ tab.icon }}</span>
      <span class="fw-tabs__label">{{ tab.label }}</span>
      <span v-if="tab.count !== undefined" class="fw-tabs__count">{{ tab.count }}</span>
    </NuxtLink>
  </nav>
</template>

<style scoped>
.fw-tabs {
  display: flex;
  align-items: stretch;
  gap: 2px;
  margin-top: var(--space-3);
  padding-bottom: 0;
  overflow-x: auto;
  scrollbar-width: thin;
}

.fw-tabs__item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border: 1px solid transparent;
  border-bottom: 2px solid transparent;
  border-radius: 6px 6px 0 0;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  text-decoration: none;
  white-space: nowrap;
  transition: background-color 0.15s ease, border-color 0.15s ease;
}

.fw-tabs__item:hover {
  background: var(--bg-row-hover);
}

.fw-tabs__item.is-active {
  background: var(--bg-card);
  border-color: var(--border-subtle);
  border-bottom-color: var(--color-primary-600, #0969da);
  color: var(--text-link);
}

.fw-tabs__icon {
  font-size: 13px;
}

.fw-tabs__count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 6px;
  background: var(--border-subtle);
  border-radius: 10px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary);
}

.fw-tabs__item.is-active .fw-tabs__count {
  background: var(--status-executed-bg);
  color: var(--status-executed-text);
}
</style>
