<script setup lang="ts">
import { onMounted } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppEmptyState from "~/shared/ui/AppEmptyState.vue";

import { useCompliancePortfolioDirectory } from "../composables/useCompliancePortfolioDirectory";
import { formatIsoDateTime } from "../lib/formatters";
import { ruleLabel } from "../lib/ruleTypeCatalog";
import type { ComplianceBreach } from "../types";

interface Props {
  items: ComplianceBreach[];
  loading: boolean;
  error: string | null;
}

defineProps<Props>();

const { t } = useI18n();
const portfolios = useCompliancePortfolioDirectory();

onMounted(() => {
  void portfolios.ensureLoaded();
});

function shortenId(id: string | null | undefined, len = 6): string {
  if (!id) return "—";
  return id.length <= len ? id : `${id.slice(0, len)}…`;
}

function relativeTime(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  const diff = Date.now() - d.getTime();
  const minutes = Math.round(diff / 60_000);
  if (minutes < 1) return "just now";
  if (minutes < 60) return `${minutes} min ago`;
  const hours = Math.round(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.round(hours / 24);
  return `${days}d ago`;
}

function tonePill(verdict: ComplianceBreach["verdict"]): string {
  if (verdict === "BLOCK") return "fail-pill fail-pill--block";
  if (verdict === "WARN") return "fail-pill fail-pill--warn";
  return "fail-pill";
}
</script>

<template>
  <AppCard>
    <template #header-actions>
      <NuxtLink to="/compliance/post-trade" class="recent__inbox">
        {{ t("compliance.dashboard.recentFailures.inboxLink") }}
      </NuxtLink>
    </template>

    <header class="recent__head">
      <h3 class="recent__title">
        {{ t("compliance.dashboard.recentFailures.title") }}
        <span class="recent__count">{{ items.length }}</span>
      </h3>
    </header>

    <div v-if="loading" class="recent__state">
      {{ t("compliance.common.loading") }}
    </div>
    <div
      v-else-if="error"
      class="recent__state recent__state--error"
      role="alert"
    >
      {{ error }}
    </div>
    <AppEmptyState
      v-else-if="items.length === 0"
      icon="check"
      :title="t('compliance.dashboard.recentFailures.empty')"
    />
    <ul v-else class="recent__list">
      <li v-for="b in items.slice(0, 4)" :key="b.id" class="recent__row">
        <span :class="tonePill(b.verdict)">{{ b.verdict }}</span>
        <div class="recent__copy">
          <div class="recent__rule">
            {{ ruleLabel(b.ruleTypeID) }} ·
            <span class="recent__msg">{{ b.message || "—" }}</span>
          </div>
          <div class="recent__meta">
            <span :title="b.portfolioID">{{ portfolios.labelFor(b.portfolioID) }}</span>
            <span>·</span>
            <span>{{ t("compliance.dashboard.recentFailures.postTrade") }}</span>
            <span>·</span>
            <span>breach <code>{{ shortenId(b.id, 8) }}</code></span>
            <span>·</span>
            <span :title="formatIsoDateTime(b.createdAt)">
              {{ relativeTime(b.createdAt) }}
            </span>
          </div>
        </div>
        <NuxtLink :to="`/compliance/audit?group=${b.checkGroupID}`">
          <AppButton variant="secondary" size="sm">
            {{ t("compliance.dashboard.recentFailures.reviewCta") }}
          </AppButton>
        </NuxtLink>
      </li>
    </ul>
  </AppCard>
</template>

<style scoped>
.recent__head {
  margin-bottom: var(--space-3);
}

.recent__title {
  margin: 0;
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
}

.recent__count {
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

.recent__inbox {
  font-size: var(--font-size-xs);
  color: var(--text-link);
  text-decoration: none;
}

.recent__inbox:hover { text-decoration: underline; }

.recent__state {
  padding: var(--space-3);
  color: var(--text-secondary);
}

.recent__state--error {
  color: var(--state-danger);
}

.recent__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-3);
}

.recent__row {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.fail-pill {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 var(--space-3);
  border-radius: var(--radius-pill);
  background: var(--bg-card-muted);
  color: var(--text-secondary);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-bold);
  letter-spacing: 0.05em;
  border: 1px solid var(--border-subtle);
}

.fail-pill--block {
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
  border-color: var(--alert-danger-border);
}

.fail-pill--warn {
  background: var(--alert-warning-bg);
  color: var(--alert-warning-text);
  border-color: var(--alert-warning-border);
}

.recent__copy {
  display: grid;
  gap: 2px;
  min-width: 0;
}

.recent__rule {
  font-weight: var(--font-weight-medium);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.recent__msg {
  color: var(--text-secondary);
  font-weight: var(--font-weight-regular);
}

.recent__meta {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  display: inline-flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.recent__meta code {
  font-family: var(--font-family-mono);
}
</style>
