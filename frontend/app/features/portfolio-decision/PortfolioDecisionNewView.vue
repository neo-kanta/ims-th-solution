<script setup lang="ts">
import { computed, onMounted, watch } from "vue";
import { useRouter } from "#imports";

import AppButton from "~/shared/ui/AppButton.vue";
import AppDateField from "~/shared/ui/AppDateField.vue";
import AppTextarea from "~/shared/ui/AppTextarea.vue";
import IMSPermissionGuard from "~/shared/ui/IMSPermissionGuard.vue";
import { useI18n } from "~/composables/useI18n";
import { useAuthStore } from "~/stores/useAuthStore";
import { usePortfolioContext } from "~/features/portfolio-workspace/composables/usePortfolioContext";

import InstrumentCombobox from "./components/InstrumentCombobox.vue";
import OrderSideToggle from "./components/OrderSideToggle.vue";
import DecisionWorkflowPanel from "./components/DecisionWorkflowPanel.vue";
import PortfolioContextPanel from "./components/PortfolioContextPanel.vue";
import { useDecisionDraft } from "./composables/useDecisionDraft";
import { useDecisionLifecycle } from "./composables/useDecisionLifecycle";
import { usePortfolioHoldingsDirectory } from "./composables/usePortfolioHoldingsDirectory";
import { estimatedConsideration, formatMoney } from "./lib/decisionFormat";
import { decisionDetailPath } from "./lib/decisionRoutes";
import { resolveDecisionNewPageState } from "./lib/pageState";
import type { DecisionValidationError } from "./lib/decisionValidation";

const props = defineProps<{ portfolioCode: string }>();

const router = useRouter();
const auth = useAuthStore();
const { t } = useI18n();

const ctx = usePortfolioContext(() => props.portfolioCode);
const holdingsDir = usePortfolioHoldingsDirectory();
const draftCtl = useDecisionDraft();
const lifecycle = useDecisionLifecycle();

const canSubmit = computed(() => auth.hasPermission("INVESTMENT_DECISION_SUBMIT"));

function inputValue(event: Event): string {
  return event.target instanceof HTMLInputElement ? event.target.value : "";
}

function validationMessage(error: DecisionValidationError | undefined): string {
  return error ? t(error.key, error.params) : "";
}

// Pure state resolution (lib/pageState.ts) so loading/not-found/permission-
// denied/error/ready is testable without mounting this component, and so a
// stale response for a portfolio the user has already navigated away from is
// never rendered as if it were current (compares the loaded descriptor's own
// code against the route's portfolioCode rather than trusting call order).
// hasFund is intentionally omitted — decisions on a fund-less portfolio are
// fully supported (Phase 2), so fund status no longer gates this page.
const pageState = computed(() =>
  resolveDecisionNewPageState({
    portfolioCode: props.portfolioCode,
    loadedPortfolioCode: ctx.portfolio.value?.code ?? null,
    error: ctx.error.value,
  }),
);

const currentHolding = computed(() =>
  holdingsDir.holdingFor(draftCtl.draft.instrumentId),
);
const currentCash = computed(() => holdingsDir.cashFor(draftCtl.draft.currency));

watch(
  () => [draftCtl.draft.side, draftCtl.draft.instrumentId] as const,
  ([side, instrumentId]) => {
    if (side === "SELL") {
      const holding = instrumentId ? holdingsDir.holdingFor(instrumentId) : null;
      draftCtl.availableQuantity.value = holding?.quantityNumeric ?? 0;
      draftCtl.isOwnedInstrument.value = instrumentId ? holding !== null : null;
    } else {
      draftCtl.availableQuantity.value = null;
      draftCtl.isOwnedInstrument.value = null;
    }
  },
);

// Switching BUY -> SELL can leave a non-owned instrument (picked while in
// BUY mode) stranded in the draft. The combobox's restrictToOwned prop stops
// a new non-owned pick going forward, but an existing one must be cleared,
// not silently carried across the side switch.
watch(
  () => draftCtl.draft.side,
  (side, previousSide) => {
    if (side === previousSide) return;
    if (side === "SELL" && draftCtl.draft.instrumentId) {
      const owned = holdingsDir.holdingFor(draftCtl.draft.instrumentId) !== null;
      if (!owned) draftCtl.clearInstrument();
    }
  },
);

const ownedHoldingsForCombobox = computed(() =>
  draftCtl.draft.side === "SELL" ? holdingsDir.holdings.value : [],
);

const considerationLabel = computed(() => {
  const value = estimatedConsideration(
    draftCtl.draft.quantity,
    draftCtl.draft.limitPrice,
    draftCtl.draft.amount,
  );
  return value === null ? "—" : formatMoney(value, draftCtl.draft.currency);
});

const estimateLabel = computed(() =>
  draftCtl.draft.side === "BUY"
    ? t("portfolio.decisionNew.estimateBuyLabel")
    : t("portfolio.decisionNew.estimateSellLabel"),
);

async function loadAll() {
  await Promise.all([ctx.reload(), holdingsDir.load(props.portfolioCode)]);
  if (!draftCtl.draft.currency) {
    draftCtl.patch({
      currency: ctx.portfolio.value?.valuation_currency || ctx.portfolio.value?.base_currency || "",
    });
  }
}

onMounted(loadAll);
watch(
  () => props.portfolioCode,
  () => {
    // The old portfolio's draft/lifecycle state (including any in-flight
    // retry-submit decisionId) must not leak into the newly selected
    // portfolio's ticket.
    draftCtl.reset();
    lifecycle.reset();
    void loadAll();
  },
);

function goToDetail(decisionId: string) {
  void router.push(decisionDetailPath(props.portfolioCode, decisionId));
}

async function onSaveDraft() {
  if (!draftCtl.isValid.value) return;
  const result = await lifecycle.saveDraft(props.portfolioCode, draftCtl.buildCreateBody());
  if (result?.id) goToDetail(result.id);
}

async function onSubmit() {
  if (!draftCtl.isValid.value) return;
  const result = await lifecycle.createThenSubmit(
    props.portfolioCode,
    draftCtl.buildCreateBody(),
    t("portfolio.decisionNew.errors.missingIdentifier"),
  );
  if (result?.id) goToDetail(result.id);
}

function onRetrySubmit() {
  void lifecycle.retrySubmit(props.portfolioCode).then((result) => {
    if (result?.id) goToDetail(result.id);
  });
}

function onCancel() {
  router.back();
}
</script>

<template>
  <div class="decision-new">
    <div v-if="pageState.kind === 'loading'" class="decision-new__state" role="status">
      {{ t("portfolio.decisionNew.loading") }}
    </div>

    <div v-else-if="pageState.kind === 'not-found'" class="decision-new__state">
      <div class="decision-new__state-title">{{ t("portfolio.decisionNew.notFoundTitle") }}</div>
      <p>{{ t("portfolio.decisionNew.notFoundBody") }}</p>
      <NuxtLink to="/portfolios" class="decision-new__state-link">
        {{ t("portfolio.decisionNew.notFoundLink") }}
      </NuxtLink>
    </div>

    <div v-else-if="pageState.kind === 'permission-denied'" class="decision-new__state decision-new__state--danger" role="alert">
      <div class="decision-new__state-title">{{ t("portfolio.decisionNew.permissionDeniedTitle") }}</div>
      <p>{{ pageState.message }}</p>
    </div>

    <div v-else-if="pageState.kind === 'error'" class="decision-new__state decision-new__state--danger" role="alert">
      <div class="decision-new__state-title">{{ t("portfolio.decisionNew.errorTitle") }}</div>
      <p>{{ pageState.message }}</p>
      <button type="button" class="decision-new__retry-btn" @click="loadAll">
        {{ t("portfolio.decisionNew.retry") }}
      </button>
    </div>

    <template v-else>
      <header class="decision-new__header">
        <div class="decision-new__title-block">
          <h1 class="decision-new__title">{{ t("portfolio.decisionNew.title") }}</h1>
          <p v-if="ctx.portfolio.value" class="decision-new__subtitle">
            {{ ctx.portfolio.value.code }} · {{ ctx.portfolio.value.name }}
            <span class="decision-new__draft-badge">{{ t("portfolio.decisionNew.draftBadge") }}</span>
          </p>
        </div>
        <div class="decision-new__actions">
          <AppButton variant="ghost" size="sm" :disabled="lifecycle.busy.value" @click="onCancel">
            ← {{ t("portfolio.decisionNew.cancel") }}
          </AppButton>
          <IMSPermissionGuard permission="INVESTMENT_DECISION_MANAGE" mode="disable">
            <AppButton
              variant="secondary"
              size="sm"
              :loading="lifecycle.busy.value"
              :disabled="!draftCtl.isValid.value || lifecycle.busy.value"
              @click="onSaveDraft"
            >
              {{ t("portfolio.decisionNew.saveDraft") }}
            </AppButton>
          </IMSPermissionGuard>
          <IMSPermissionGuard permission="INVESTMENT_DECISION_SUBMIT" mode="disable">
            <AppButton
              variant="primary"
              size="sm"
              :loading="lifecycle.busy.value"
              :disabled="!draftCtl.isValid.value || lifecycle.busy.value"
              @click="onSubmit"
            >
              {{ t("portfolio.decisionNew.submitForApproval") }}
            </AppButton>
          </IMSPermissionGuard>
        </div>
      </header>

      <p v-if="!canSubmit" class="decision-new__permission-hint">
        {{ t("portfolio.decisionNew.submitPermissionHint", { permission: "INVESTMENT_DECISION_SUBMIT" }) }}
      </p>

      <div v-if="lifecycle.error.value" class="decision-new__banner decision-new__banner--danger" role="alert">
        <span>{{ lifecycle.error.value }}</span>
        <button
          v-if="lifecycle.phase.value === 'submit-failed'"
          type="button"
          class="decision-new__retry-btn"
          :disabled="lifecycle.busy.value"
          @click="onRetrySubmit"
        >
          {{ t("portfolio.decisionNew.retrySubmit") }}
        </button>
      </div>

      <div class="decision-new__grid">
        <aside class="decision-new__panel">
          <div class="decision-new__panel-header">{{ t("portfolio.decisionNew.portfolioPanelTitle") }}</div>
          <div class="decision-new__panel-body">
            <PortfolioContextPanel
              :portfolio="ctx.portfolio.value"
              :side="draftCtl.draft.side"
              :currency="draftCtl.draft.currency"
              :cash-balance="currentCash?.balance ?? null"
              :holding="currentHolding"
            />
          </div>
        </aside>

        <section class="decision-new__panel decision-new__ticket">
          <div class="decision-new__panel-header">{{ t("portfolio.decisionNew.ticketPanelTitle") }}</div>
          <div class="decision-new__panel-body decision-new__ticket-body">
            <OrderSideToggle
              :model-value="draftCtl.draft.side"
              :disabled="lifecycle.busy.value"
              @update:model-value="draftCtl.setSide"
            />

            <div class="decision-new__field">
              <label class="decision-new__label" for="decision-instrument">
                {{ t("portfolio.decisionNew.instrumentLabel") }} <span class="req">*</span>
              </label>
              <InstrumentCombobox
                :selected="draftCtl.selectedInstrument.value"
                :owned-holdings="ownedHoldingsForCombobox"
                :restrict-to-owned="draftCtl.draft.side === 'SELL'"
                :disabled="lifecycle.busy.value"
                :error-message="validationMessage(draftCtl.errors.value.instrument) || null"
                input-id="decision-instrument"
                @select="draftCtl.selectInstrument"
                @clear="draftCtl.clearInstrument"
              />
            </div>

            <div class="decision-new__row">
              <div class="decision-new__field">
                <label class="decision-new__label" for="decision-quantity">
                  {{ t("portfolio.decisionNew.quantityLabel") }}
                </label>
                <input
                  id="decision-quantity"
                  class="input"
                  type="text"
                  inputmode="decimal"
                  placeholder="0"
                  :disabled="lifecycle.busy.value"
                  :value="draftCtl.draft.quantity"
                  @input="draftCtl.patch({ quantity: inputValue($event) })"
                />
                <span v-if="draftCtl.errors.value.quantity" class="decision-new__error">
                  {{ validationMessage(draftCtl.errors.value.quantity) }}
                </span>
              </div>
              <div class="decision-new__field">
                <label class="decision-new__label" for="decision-amount">
                  {{ t("portfolio.decisionNew.amountLabel") }}
                </label>
                <input
                  id="decision-amount"
                  class="input"
                  type="text"
                  inputmode="decimal"
                  placeholder="0.00"
                  :disabled="lifecycle.busy.value"
                  :value="draftCtl.draft.amount"
                  @input="draftCtl.patch({ amount: inputValue($event) })"
                />
                <span v-if="draftCtl.errors.value.amount" class="decision-new__error">
                  {{ validationMessage(draftCtl.errors.value.amount) }}
                </span>
              </div>
              <div class="decision-new__field">
                <label class="decision-new__label" for="decision-limit-price">
                  {{ t("portfolio.decisionNew.limitPriceLabel") }}
                </label>
                <input
                  id="decision-limit-price"
                  class="input"
                  type="text"
                  inputmode="decimal"
                  placeholder="0.00"
                  :disabled="lifecycle.busy.value"
                  :value="draftCtl.draft.limitPrice"
                  @input="draftCtl.patch({ limitPrice: inputValue($event) })"
                />
                <span v-if="draftCtl.errors.value.limitPrice" class="decision-new__error">
                  {{ validationMessage(draftCtl.errors.value.limitPrice) }}
                </span>
              </div>
            </div>

            <div class="decision-new__estimate">
              <span class="decision-new__estimate-label">{{ estimateLabel }}</span>
              <span class="decision-new__estimate-value">{{ considerationLabel }}</span>
            </div>

            <div class="decision-new__row">
              <div class="decision-new__field">
                <label class="decision-new__label" for="decision-currency">
                  {{ t("portfolio.decisionNew.currencyLabel") }} <span class="req">*</span>
                </label>
                <input
                  id="decision-currency"
                  class="input"
                  type="text"
                  maxlength="3"
                  :placeholder="t('portfolio.decisionNew.currencyPlaceholder')"
                  :disabled="lifecycle.busy.value"
                  :value="draftCtl.draft.currency"
                  @input="draftCtl.patch({ currency: inputValue($event).toUpperCase() })"
                />
                <span v-if="draftCtl.errors.value.currency" class="decision-new__error">
                  {{ validationMessage(draftCtl.errors.value.currency) }}
                </span>
              </div>
              <div class="decision-new__field">
                <label class="decision-new__label" for="decision-exchange">
                  {{ t("portfolio.decisionNew.exchangeLabel") }}
                </label>
                <input
                  id="decision-exchange"
                  class="input"
                  type="text"
                  :placeholder="t('portfolio.decisionNew.exchangePlaceholder')"
                  :disabled="lifecycle.busy.value"
                  :value="draftCtl.draft.exchange"
                  @input="draftCtl.patch({ exchange: inputValue($event) })"
                />
              </div>
              <div class="decision-new__field">
                <label class="decision-new__label" for="decision-business-date">
                  {{ t("portfolio.decisionNew.businessDateLabel") }} <span class="req">*</span>
                </label>
                <AppDateField
                  id="decision-business-date"
                  :model-value="draftCtl.draft.businessDate"
                  :disabled="lifecycle.busy.value"
                  :error="!!draftCtl.errors.value.businessDate"
                  @update:model-value="(v: string) => draftCtl.patch({ businessDate: v })"
                />
                <span v-if="draftCtl.errors.value.businessDate" class="decision-new__error">
                  {{ validationMessage(draftCtl.errors.value.businessDate) }}
                </span>
              </div>
            </div>

            <div class="decision-new__field">
              <label class="decision-new__label" for="decision-rationale">
                {{ t("portfolio.decisionNew.rationaleLabel") }}
              </label>
              <AppTextarea
                id="decision-rationale"
                :model-value="draftCtl.draft.rationale"
                :placeholder="t('portfolio.decisionNew.rationalePlaceholder')"
                :disabled="lifecycle.busy.value"
                :rows="4"
                @update:model-value="(v: string) => draftCtl.patch({ rationale: v })"
              />
            </div>
          </div>
        </section>

        <aside class="decision-new__panel">
          <div class="decision-new__panel-header">{{ t("portfolio.decisionNew.nextPanelTitle") }}</div>
          <div class="decision-new__panel-body decision-new__instructions">
            <p>
              <strong>{{ t("portfolio.decisionNew.saveDraft") }}</strong>
              {{ " " }}{{ t("portfolio.decisionNew.saveDraftHelp") }}
            </p>
            <p>
              <strong>{{ t("portfolio.decisionNew.submitForApproval") }}</strong>
              {{ " " }}{{ t("portfolio.decisionNew.submitHelp") }}
            </p>
            <DecisionWorkflowPanel :status="lifecycle.decision.value?.status ?? 'DRAFT'" />
          </div>
        </aside>
      </div>
    </template>
  </div>
</template>

<style scoped>
.decision-new {
  display: grid;
  gap: 16px;
}

.decision-new__state {
  padding: 32px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-lg, 8px);
  background: var(--bg-card, #fff);
  display: grid;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary);
  max-width: 520px;
}

.decision-new__state--danger {
  background: var(--alert-danger-bg, #ffebe9);
  border-color: var(--alert-danger-border, #cf222e);
  color: var(--alert-danger-text, #cf222e);
}

.decision-new__state-title {
  font-weight: 700;
  font-size: 14px;
  color: inherit;
}

.decision-new__state-link {
  justify-self: start;
  color: var(--action-primary, #0969da);
  font-weight: 600;
  text-decoration: none;
}

.decision-new__state-link:hover {
  text-decoration: underline;
}

.decision-new__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.decision-new__title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
}

.decision-new__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.decision-new__draft-badge {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  color: var(--text-tertiary);
}

.decision-new__actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.decision-new__permission-hint {
  margin: -8px 0 0;
  font-size: 12px;
  color: var(--text-tertiary);
}

.decision-new__banner {
  padding: 10px 14px;
  border-radius: var(--radius-sm, 4px);
  font-size: 13px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.decision-new__banner--danger {
  background: var(--alert-danger-bg, #ffebe9);
  color: var(--alert-danger-text, #cf222e);
  border: 1px solid var(--alert-danger-border, #cf222e);
}

.decision-new__retry-btn {
  flex-shrink: 0;
  border: 1px solid currentColor;
  background: transparent;
  color: inherit;
  border-radius: var(--radius-sm, 4px);
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}

.decision-new__grid {
  display: grid;
  grid-template-columns: 240px 1fr 280px;
  gap: 16px;
  align-items: start;
}

.decision-new__panel {
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-lg, 8px);
  background: var(--bg-card, #fff);
  overflow: hidden;
}

.decision-new__panel-header {
  background: var(--bg-card-muted, #f6f8fa);
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
  padding: 8px 14px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
}

.decision-new__panel-body {
  padding: 14px;
}

.decision-new__ticket-body {
  display: grid;
  gap: 16px;
}

.decision-new__field {
  display: grid;
  gap: 4px;
}

.decision-new__label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--text-secondary);
}

.req {
  color: var(--state-danger, #cf222e);
}

.decision-new__row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px;
}

.decision-new__error {
  font-size: 12px;
  color: var(--state-danger, #cf222e);
}

.decision-new__estimate {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  flex-wrap: wrap;
}

.decision-new__estimate-label {
  font-size: 11px;
  color: var(--text-tertiary);
}

.decision-new__estimate-value {
  font-size: 16px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary);
}

.decision-new__instructions {
  display: grid;
  gap: 14px;
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
}

@media (max-width: 1080px) {
  .decision-new__grid {
    grid-template-columns: 1fr;
  }
}
</style>
