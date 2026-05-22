<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AppCard from "~/shared/ui/AppCard.vue";
import { useI18n } from "~/composables/useI18n";
import FundWorkspaceHeader from "./FundWorkspaceHeader.vue";
import FundWorkspaceTabs from "./FundWorkspaceTabs.vue";
import { useFundWorkspace } from "../composables/useFundWorkspace";
import type { WorkspaceTab } from "../types";

/**
 * Shared shell for the 6 non-Holdings workspace tabs. Renders the same
 * fund header + tab bar so the user can navigate, then a single empty
 * "coming next" card. Replace this with the real implementation per tab
 * in future milestones.
 */

const props = defineProps<{
  fundId: string;
  tab: WorkspaceTab;
}>();

const { t } = useI18n();
const { funds, activeFund, loadFunds } = useFundWorkspace();

const tabLabel = computed(() =>
  t(`holdings.workspaceTabs.${props.tab}` as any, props.tab),
);

const scaffoldTitle = computed(() =>
  t("holdings.scaffold.title", { tab: tabLabel.value }, `${tabLabel.value} for fund-alpha`),
);

const headerMissing = ref<string | null>(null);

onMounted(async () => {
  try {
    await loadFunds(props.fundId);
  } catch (err) {
    headerMissing.value = (err as Error)?.message ?? "Header unavailable";
  }
});
</script>

<template>
  <section class="ws-scaffold">
    <FundWorkspaceHeader v-if="activeFund" :header="{
      code: activeFund.code,
      short_name: activeFund.short_name,
      contract_code: activeFund.contract_code,
      privacy: activeFund.privacy,
      watch_count: 14,
      star_count: 38,
      subscribed: false,
      tab_counts: { stages: 7, decisions: 23, compliance: 2 },
    }" />

    <FundWorkspaceTabs
      v-if="activeFund"
      :fund-id="activeFund.fund_id"
      :active="tab"
      :counts="{ stages: 7, decisions: 23, compliance: 2 }"
    />

    <AppCard :title="scaffoldTitle">
      <div class="ws-scaffold__body">
        <p class="ws-scaffold__copy">{{ t("holdings.scaffold.description", "This workspace tab is scaffolded.") }}</p>
        <span class="ws-scaffold__chip">{{ t("holdings.mockDisclaimer", "Showing mock data — backend integration pending.") }}</span>
      </div>
    </AppCard>
  </section>
</template>

<style scoped>
.ws-scaffold {
  display: grid;
  gap: var(--space-3);
}

.ws-scaffold__body {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-6) 0;
  place-items: center;
  text-align: center;
}

.ws-scaffold__copy {
  margin: 0;
  color: var(--text-secondary);
  max-width: 36rem;
}

.ws-scaffold__chip {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  background: var(--bg-row-hover);
  border: 1px solid var(--border-subtle);
  border-radius: 999px;
  font-size: 11px;
  color: var(--text-tertiary);
}
</style>
