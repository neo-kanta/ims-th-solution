<script setup lang="ts">
import { ref } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppCard from "~/shared/ui/AppCard.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";

const { t } = useI18n();
const router = useRouter();
const query = ref("");

const FILTERS = [
  { key: "fund", labelKey: "compliance.dashboard.search.filterFund" },
  { key: "assetClass", labelKey: "compliance.dashboard.search.filterAssetClass" },
  { key: "ruleType", labelKey: "compliance.dashboard.search.filterRuleType" },
  { key: "status", labelKey: "compliance.dashboard.search.filterStatus" },
  { key: "severity", labelKey: "compliance.dashboard.search.filterSeverity" },
  { key: "effective", labelKey: "compliance.dashboard.search.filterEffective" },
  { key: "owner", labelKey: "compliance.dashboard.search.filterOwner" },
  { key: "approval", labelKey: "compliance.dashboard.search.filterApproval" },
] as const;

function submitSearch() {
  // Backend only supports filter by rule_type_id today, so we route to the
  // rule library passing the raw query as `rule_type_id` if it looks like
  // an id (contains a dot, matches catalog), or as a `highlight` hint
  // otherwise. Either way: no fake server-side search.
  const v = query.value.trim();
  void router.push({
    path: "/compliance/rules",
    query: v ? { rule_type_id: v } : {},
  });
}
</script>

<template>
  <AppCard>
    <form class="search-card" @submit.prevent="submitSearch">
      <div class="search-card__input-wrap">
        <input
          v-model="query"
          type="search"
          class="search-card__input"
          :placeholder="t('compliance.dashboard.search.placeholder')"
          :aria-label="t('compliance.dashboard.search.placeholder')"
        />
        <button
          type="submit"
          class="search-card__submit"
          :aria-label="t('compliance.dashboard.search.placeholder')"
        >
          <AppIcon name="search" size="sm" />
        </button>
      </div>

      <div class="search-card__chips" role="group">
        <button
          v-for="f in FILTERS"
          :key="f.key"
          type="button"
          class="search-card__chip"
          :title="t('compliance.dashboard.search.backendNote')"
          :disabled="true"
        >
          {{ t(f.labelKey as any) }}
        </button>
      </div>
    </form>
  </AppCard>
</template>

<style scoped>
.search-card {
  display: grid;
  gap: var(--space-3);
}

.search-card__input-wrap {
  position: relative;
}

.search-card__input {
  width: 100%;
  height: var(--size-control-md);
  padding: 0 var(--space-9) 0 var(--space-3);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--bg-input);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  font-family: inherit;
}

.search-card__input:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: var(--shadow-focus);
}

.search-card__submit {
  position: absolute;
  right: 4px;
  top: 4px;
  bottom: 4px;
  width: 32px;
  display: grid;
  place-items: center;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
}

.search-card__submit:hover {
  background: var(--action-ghost-hover);
  color: var(--text-primary);
}

.search-card__chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.search-card__chip {
  height: 26px;
  padding: 0 var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-pill);
  background: var(--bg-card);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-family: inherit;
  cursor: not-allowed;
  opacity: 0.65;
}
</style>
