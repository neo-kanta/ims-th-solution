<script setup lang="ts">
import { onMounted, ref } from "vue";
import DecisionCockpit from "~/features/investment-decision/components/DecisionCockpit.vue";
import DecisionPortfolioPanel from "~/features/investment-decision/components/DecisionPortfolioPanel.vue";
import DecisionFormPanel from "~/features/investment-decision/components/DecisionFormPanel.vue";
import { useDecisionMutations } from "~/features/investment-decision/composables/useDecisionMutations";
import type { ApiCreateDecisionRequest } from "~/features/investment-decision/services/decisionApi";
import { investmentLedgerApi } from "~/features/investment-ledger/services/investmentLedgerApi";
import type { components } from "~/api/ims-api";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_FUND_VIEW",
});

type ApiPortfolio = components["schemas"]["PortfolioResponse"];

const route = useRoute();
const router = useRouter();
const fundId = computed(() => route.params.fundId as string);

const mutations = useDecisionMutations();
const portfolios = ref<ApiPortfolio[]>([]);
const portfoliosLoading = ref(false);
const selectedPortfolio = ref<ApiPortfolio | null>(null);

const form = ref<Partial<ApiCreateDecisionRequest>>({});

function selectPortfolio(p: ApiPortfolio) {
  selectedPortfolio.value = p;
}

function validate(): string | null {
  if (!selectedPortfolio.value?.id) return "Select a portfolio";
  if (!form.value.instrument_code?.trim()) return "Instrument code is required";
  if (!form.value.side) return "Side is required";
  if (!form.value.currency?.trim() || form.value.currency.length !== 3)
    return "Currency must be 3 characters";
  if (!form.value.business_date) return "Business date is required";
  return null;
}

async function saveDraft() {
  const validErr = validate();
  if (validErr) {
    mutations.error.value = validErr;
    return;
  }
  const body = buildBody();
  if (!body) return;
  const result = await mutations.create(body);
  if (result?.id) {
    void router.push(`/investment/funds/${fundId.value}/operation/${result.id}`);
  }
}

async function submitDecision() {
  const validErr = validate();
  if (validErr) {
    mutations.error.value = validErr;
    return;
  }
  const body = buildBody();
  if (!body) return;
  const created = await mutations.create(body);
  if (!created?.id) return;
  const submitted = await mutations.submit(created.id);
  const targetId = submitted?.id ?? created.id;
  void router.push(`/investment/funds/${fundId.value}/operation/${targetId}`);
}

function buildBody(): ApiCreateDecisionRequest | null {
  const p = selectedPortfolio.value;
  if (!p?.id) return null;
  const f = form.value;
  return {
    fund_id: fundId.value,
    contract_id: fundId.value,
    portfolio_id: p.id,
    instrument_code: f.instrument_code ?? "",
    side: (f.side as "BUY" | "SELL") ?? "BUY",
    currency: f.currency ?? "",
    business_date: f.business_date ?? "",
    quantity: f.quantity || undefined,
    amount: f.amount || undefined,
    limit_price: f.limit_price || undefined,
    exchange: f.exchange || undefined,
    rationale: f.rationale || undefined,
  };
}

onMounted(async () => {
  portfoliosLoading.value = true;
  try {
    const result = await investmentLedgerApi.listPortfolios({
      fund_id: fundId.value,
    });
    portfolios.value = result.items ?? [];
    if (portfolios.value.length > 0) {
      selectedPortfolio.value = portfolios.value[0] ?? null;
    }
  } finally {
    portfoliosLoading.value = false;
  }
});
</script>

<template>
  <div>
    <DecisionCockpit>
      <template #title>New Investment Decision</template>
      <template #toolbar>
        <button
          class="toolbar-btn toolbar-btn--primary"
          :disabled="mutations.loading.value"
          @click="submitDecision"
        >
          Submit Decision
        </button>
        <button
          class="toolbar-btn"
          :disabled="mutations.loading.value"
          @click="saveDraft"
        >
          Save Draft
        </button>
        <button class="toolbar-btn" @click="router.back()">← Cancel</button>
      </template>

      <template #left>
        <DecisionPortfolioPanel
          :portfolios="portfolios"
          :selected-id="selectedPortfolio?.id"
          :loading="portfoliosLoading"
          @select="selectPortfolio"
        />
      </template>

      <template #middle>
        <DecisionFormPanel
          :model-value="form"
          @update:model-value="(v) => (form = v)"
        />
      </template>

      <template #right>
        <div class="panel-header">Instructions</div>
        <div class="instructions">
          <p>Select a portfolio from the left panel, then fill in the decision details.</p>
          <p><strong>Submit Decision</strong> creates the decision and immediately submits it for approval.</p>
          <p><strong>Save Draft</strong> creates the decision in DRAFT status so you can review it first.</p>
          <p>Fields marked <span class="required">*</span> are required.</p>
        </div>
      </template>
    </DecisionCockpit>

    <div v-if="mutations.error.value" class="error-banner mt">
      {{ mutations.error.value }}
    </div>
  </div>
</template>

<style scoped>
.toolbar-btn {
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 4px);
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: background-color 0.1s;
}

.toolbar-btn:hover {
  background: var(--bg-card-hover);
}

.toolbar-btn--primary {
  background: var(--state-success, #1a7f37);
  color: #ffffff;
  border-color: transparent;
}

.toolbar-btn--primary:hover {
  opacity: 0.9;
}

.toolbar-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.panel-header {
  background: var(--bg-card-hover);
  border-bottom: 1px solid var(--border-subtle);
  padding: 8px 12px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.instructions {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
}

.required {
  color: var(--state-danger, #cf222e);
}

.error-banner {
  padding: 10px 14px;
  background: rgba(207, 34, 46, 0.1);
  border: 1px solid var(--state-danger, #cf222e);
  border-radius: var(--radius-sm, 4px);
  color: var(--state-danger, #cf222e);
  font-size: 13px;
}

.mt {
  margin-top: 12px;
}
</style>
