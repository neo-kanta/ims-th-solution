<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import AppCard from "~/shared/ui/AppCard.vue";

import { ruleLabel } from "../lib/ruleTypeCatalog";
import type { ComplianceRule } from "../types";

interface Props {
  items: ComplianceRule[];
  loading?: boolean;
}

defineProps<Props>();

const { t } = useI18n();
</script>

<template>
  <AppCard>
    <template #header-actions>
      <NuxtLink
        to="/compliance/rules?is_active=true"
        class="hr-list__seeall"
      >
        {{ t("compliance.dashboard.highRiskRules.seeAll") }}
      </NuxtLink>
    </template>

    <header class="hr-list__head">
      <h3 class="hr-list__title">
        {{ t("compliance.dashboard.highRiskRules.title") }}
        <span class="hr-list__count">{{ items.length }}</span>
      </h3>
    </header>

    <ul v-if="items.length > 0" class="hr-list">
      <li
        v-for="r in items.slice(0, 6)"
        :key="r.id"
        class="hr-item"
      >
        <span class="hr-pill">BLOCK</span>
        <NuxtLink
          :to="`/compliance/rules/${r.id}`"
          class="hr-item__link"
        >
          <code class="hr-item__code">{{ r.ruleTypeID }}</code>
          <span class="hr-item__name">· {{ r.name || ruleLabel(r.ruleTypeID, t) }}</span>
        </NuxtLink>
      </li>
    </ul>
    <p v-else class="hr-empty">
      {{ t("compliance.dashboard.highRiskRules.empty") }}
    </p>
  </AppCard>
</template>

<style scoped>
.hr-list__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: var(--space-3);
}

.hr-list__title {
  margin: 0;
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
}

.hr-list__count {
  margin-left: var(--space-2);
  display: inline-flex;
  align-items: center;
  height: 20px;
  min-width: 20px;
  padding: 0 var(--space-2);
  border-radius: var(--radius-pill);
  background: var(--bg-card-muted);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.hr-list__seeall {
  font-size: var(--font-size-xs);
  color: var(--text-link);
  text-decoration: none;
}

.hr-list__seeall:hover {
  text-decoration: underline;
}

.hr-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-3);
}

.hr-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--font-size-sm);
}

.hr-pill {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 var(--space-3);
  border-radius: var(--radius-pill);
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
  border: 1px solid var(--alert-danger-border);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-bold);
  letter-spacing: 0.05em;
}

.hr-item__link {
  display: inline-flex;
  align-items: baseline;
  gap: var(--space-2);
  flex-wrap: wrap;
  color: var(--text-primary);
  text-decoration: none;
  min-width: 0;
}

.hr-item__link:hover {
  text-decoration: underline;
}

.hr-item__code {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

.hr-item__name {
  color: var(--text-primary);
}

.hr-empty {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-tertiary);
}
</style>
