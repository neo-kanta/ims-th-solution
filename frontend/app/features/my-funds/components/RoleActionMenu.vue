<script setup lang="ts">
import { computed } from "vue";

import { useAuthStore } from "~/stores/useAuthStore";
import { useI18n } from "~/composables/useI18n";

import type { MyFundCard } from "../types";

interface Props {
  card: MyFundCard;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "open", fundId: string): void;
  (e: "preTrade", fundId: string): void;
  (e: "viewBreaches", fundId: string): void;
  (e: "writeResearch", fundId: string): void;
}>();

const auth = useAuthStore();
const { t } = useI18n();

const canPreTrade = computed(() => auth.hasPermission("INVESTMENT_LEDGER_SIMULATE"));
const canExecute = computed(() => auth.hasPermission("INVESTMENT_LEDGER_POST"));
const canCreateResearch = computed(() => auth.hasPermission("INVESTMENT_RESEARCH_CREATE"));
const canViewBreaches = computed(() => auth.hasPermission("IRG_VIEW_BREACHES"));
const isLocked = computed(() => props.card.status === "LOCKED" || props.card.status === "CLOSED");

const showPreTrade = computed(
  () => canPreTrade.value && (props.card.role === "MANAGER" || canExecute.value) && !isLocked.value,
);
</script>

<template>
  <div class="role-actions">
    <button
      v-if="showPreTrade"
      type="button"
      class="role-actions__btn role-actions__btn--ghost"
      @click.stop="emit('preTrade', card.fund_id)"
    >
      {{ t("myFunds.actions.preTrade", "Run pre-trade") }}
    </button>
    <button
      v-if="canCreateResearch && card.role !== 'MEMBER'"
      type="button"
      class="role-actions__btn role-actions__btn--ghost"
      @click.stop="emit('writeResearch', card.fund_id)"
    >
      {{ t("myFunds.actions.writeResearch", "New research") }}
    </button>
    <button
      v-if="canViewBreaches && card.compliance.open_count > 0"
      type="button"
      class="role-actions__btn role-actions__btn--ghost"
      @click.stop="emit('viewBreaches', card.fund_id)"
    >
      {{ t("myFunds.actions.viewBreaches", "Breaches") }}
    </button>
    <button
      type="button"
      class="role-actions__btn role-actions__btn--primary"
      @click.stop="emit('open', card.fund_id)"
    >
      {{ t("myFunds.actions.open", "Open") }}
    </button>
  </div>
</template>

<style scoped>
.role-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.role-actions__btn {
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  border-radius: var(--radius-md, 5px);
  padding: 5px 10px;
  cursor: pointer;
  line-height: 1.2;
  transition: background-color 0.12s ease, border-color 0.12s ease;
}

.role-actions__btn--ghost {
  background: var(--bg-card, #ffffff);
  color: var(--text-primary, #1f2328);
  border: 1px solid var(--border-subtle, #d0d7de);
}

.role-actions__btn--ghost:hover {
  background: var(--bg-card-muted, #f6f8fa);
}

.role-actions__btn--primary {
  background: var(--state-info, #1f6feb);
  border: 1px solid var(--state-info, #1f6feb);
  color: #ffffff;
}

.role-actions__btn--primary:hover {
  filter: brightness(0.96);
}
</style>
