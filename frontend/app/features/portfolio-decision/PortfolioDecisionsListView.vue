<script setup lang="ts">
import { onMounted, watch } from "vue";
import { useRouter } from "#imports";

import AppButton from "~/shared/ui/AppButton.vue";
import IMSPermissionGuard from "~/shared/ui/IMSPermissionGuard.vue";
import { usePortfolioContext } from "~/features/portfolio-workspace/composables/usePortfolioContext";
import PortfolioWorkspaceHeader from "~/features/portfolio-workspace/components/PortfolioWorkspaceHeader.vue";

import DecisionListTable from "./components/DecisionListTable.vue";
import { usePortfolioDecisionsList } from "./composables/usePortfolioDecisionsList";
import type { ApiDecisionV2 } from "./services/portfolioDecisionApi";
import { decisionDetailPath, newDecisionPath } from "./lib/decisionRoutes";

const props = defineProps<{ portfolioCode: string }>();
const router = useRouter();

const ctx = usePortfolioContext(() => props.portfolioCode);
const list = usePortfolioDecisionsList();

function reload() {
  void ctx.reload();
  void list.load(props.portfolioCode);
}

onMounted(reload);
watch(() => props.portfolioCode, reload);

function openNew() {
  void router.push(newDecisionPath(props.portfolioCode));
}

function openDecision(decision: ApiDecisionV2) {
  if (!decision.id) return;
  void router.push(decisionDetailPath(props.portfolioCode, decision.id));
}
</script>

<template>
  <section class="decisions-list">
    <PortfolioWorkspaceHeader :portfolio="ctx.portfolio.value" />

    <div class="decisions-list__toolbar">
      <h2 class="decisions-list__title">Investment decisions</h2>
      <IMSPermissionGuard permission="INVESTMENT_DECISION_MANAGE">
        <AppButton variant="primary" size="sm" @click="openNew">+ New decision</AppButton>
      </IMSPermissionGuard>
    </div>

    <p v-if="list.error.value" class="decisions-list__error" role="alert">
      {{ list.error.value }}
    </p>

    <DecisionListTable
      :items="list.items.value"
      :loading="list.loading.value"
      @open="openDecision"
    />

    <div v-if="list.total.value > list.limit.value" class="decisions-list__pagination">
      <button
        type="button"
        :disabled="list.page.value <= 1"
        @click="list.setPage(list.page.value - 1, portfolioCode)"
      >
        ‹ Prev
      </button>
      <span>Page {{ list.page.value }} / {{ Math.ceil(list.total.value / list.limit.value) }}</span>
      <button
        type="button"
        :disabled="list.page.value * list.limit.value >= list.total.value"
        @click="list.setPage(list.page.value + 1, portfolioCode)"
      >
        Next ›
      </button>
    </div>
  </section>
</template>

<style scoped>
.decisions-list {
  display: grid;
  gap: 16px;
}

.decisions-list__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.decisions-list__title {
  margin: 0;
  font-size: 1rem;
  font-weight: 700;
  color: var(--text-primary);
}

.decisions-list__error {
  padding: 10px 14px;
  background: var(--alert-danger-bg, #ffebe9);
  border: 1px solid var(--alert-danger-border, #cf222e);
  border-radius: var(--radius-sm, 4px);
  color: var(--alert-danger-text, #cf222e);
  font-size: 13px;
}

.decisions-list__pagination {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  color: var(--text-secondary);
}

.decisions-list__pagination button {
  padding: 4px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-card);
  cursor: pointer;
  font-size: 13px;
}

.decisions-list__pagination button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>
