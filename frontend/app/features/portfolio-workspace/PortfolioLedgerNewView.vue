<script setup lang="ts">
/**
 * Portfolio-code-routed "new ledger entry" page: simulate then post a
 * transaction directly to the ledger (no approval workflow — that is the
 * separate portfolio-decision feature at /portfolios/{code}/decisions/new).
 *
 * Reuses the existing simulate/post transaction workflow from
 * features/investment-ledger/composables/useOrderTicket.ts (validation
 * rules, request building, stage machine, stale-simulation guard, BLOCK
 * guard, duplicate-submit prevention) rather than rebuilding order-entry
 * logic. That composable previously called the legacy fund/UUID-scoped
 * routes only; it now accepts injectable simulate/post functions (see its
 * `OrderTicketDeps`), so this view injects the real, existing V2
 * portfolio-code routes (`POST /portfolios/{portfolioCode}/transactions
 * /simulate` and `POST /portfolios/{portfolioCode}/transactions` — both
 * confirmed mounted in backend/internal/investment/module.go and typed in
 * ims-api.d.ts) instead of duplicating the composable.
 *
 * MODEL portfolios are blocked from entering this page's form entirely
 * (docs/MANAGER/MEMORY.md: "MODEL has no real ledger or execution"). This is
 * a frontend-only affordance — see lib/ledgerGuard.ts's header comment: the
 * backend does not currently reject MODEL on this path either, so this is
 * UX, not enforcement.
 */
import { computed, onMounted, ref, watch } from "vue";
import { useState } from "#imports";

import AppCard from "~/shared/ui/AppCard.vue";
import AppConfirmDialog from "~/shared/ui/AppConfirmDialog.vue";
import { useI18n, type AppTranslationKey } from "~/composables/useI18n";
import PortfolioWorkspaceHeader from "./components/PortfolioWorkspaceHeader.vue";
import { usePortfolioContext } from "./composables/usePortfolioContext";
import { portfolioApi } from "./services/portfolioApi";
import { canEnterLedgerTransaction, ledgerIntentFor } from "./lib/ledgerGuard";

import { useOrderTicket } from "~/features/investment-ledger/composables/useOrderTicket";
import InstrumentCombobox from "~/features/portfolio-decision/components/InstrumentCombobox.vue";
import type { ApiInstrument } from "~/features/investment-ledger/services/investmentLedgerApi";

const props = defineProps<{ portfolioCode: string }>();
const { t } = useI18n();

const pageTitle = useState<string>("page-title", () => "");
watch(
  () => t("portfolio.ledgerNew.title"),
  (newTitle) => {
    pageTitle.value = newTitle || "";
  },
  { immediate: true },
);

const ctx = usePortfolioContext(() => props.portfolioCode);

const ticket = useOrderTicket({
  simulate: (code, payload) => portfolioApi.simulateTransaction(code, payload),
  post: (code, payload) => portfolioApi.postTransaction(code, payload),
});

const selectedInstrument = ref<ApiInstrument | null>(null);
const opened = ref(false);

function openTicket() {
  if (opened.value) return;
  ticket.open(props.portfolioCode, {
    currency: ctx.portfolio.value?.base_currency ?? "",
  });
  selectedInstrument.value = null;
  opened.value = true;
}

onMounted(() => {
  void ctx.reload();
});
watch(
  () => props.portfolioCode,
  () => {
    opened.value = false;
    void ctx.reload();
  },
);
watch(
  () => ctx.portfolio.value,
  (portfolio) => {
    if (portfolio && canEnterLedgerTransaction(ctx.portfolioType.value)) {
      openTicket();
    }
  },
);

const intent = computed(() => ledgerIntentFor(ctx.portfolioType.value));
const draft = ticket.draft;

function chooseInstrument(inst: ApiInstrument) {
  selectedInstrument.value = inst;
  ticket.patchDraft({
    instrument_id: inst.id ?? "",
    currency: draft.currency || inst.currency || "",
  });
}

function clearInstrument() {
  selectedInstrument.value = null;
  ticket.patchDraft({ instrument_id: "" });
}

const errorKeys = computed(() => Object.keys(ticket.validationErrors.value));

function fieldMessage(field: string): string | null {
  if (!errorKeys.value.includes(field)) return null;
  return t(`portfolio.ledgerNew.validation.${field}` as AppTranslationKey);
}

function onSimulate() {
  void ticket.simulate();
}

function onRequestPost() {
  ticket.requestConfirmation();
}

function onCancelPost() {
  ticket.cancelConfirmation();
}

async function onConfirmPost() {
  await ticket.post();
}

const verdict = computed(() => ticket.verdict.value);
</script>

<template>
  <section class="portfolio-ledger-new">
    <PortfolioWorkspaceHeader :portfolio="ctx.portfolio.value" />

    <AppCard :title="t('portfolio.ledgerNew.title')" :subtitle="t('portfolio.ledgerNew.subtitle')">
      <div v-if="ctx.loading.value && !ctx.portfolio.value" class="portfolio-ledger-new__notice" role="status">
        {{ t("portfolio.ledgerNew.loading") }}
      </div>
      <div v-else-if="ctx.error.value" class="portfolio-ledger-new__error" role="alert">
        {{ t("portfolio.ledgerNew.loadErrorTitle") }}: {{ ctx.error.value }}
      </div>

      <div
        v-else-if="ctx.isModel.value"
        class="portfolio-ledger-new__blocked"
        role="alert"
      >
        <strong>{{ t("portfolio.ledgerNew.modelBlockedTitle") }}</strong>
        <p>{{ t("portfolio.ledgerNew.modelBlockedBody") }}</p>
      </div>

      <template v-else-if="ctx.portfolio.value">
        <div
          class="portfolio-ledger-new__intent-banner"
          :data-intent="intent"
          role="status"
        >
          {{ intent === "PAPER" ? t("portfolio.ledgerNew.intentPaper") : t("portfolio.ledgerNew.intentOfficial") }}
        </div>

        <form class="portfolio-ledger-new__form" @submit.prevent>
          <div class="portfolio-ledger-new__side">
            <button
              type="button"
              class="portfolio-ledger-new__side-btn"
              :class="{ 'is-active': draft.side === 'BUY' }"
              @click="ticket.setSide('BUY')"
            >
              {{ t("portfolio.ledgerNew.buy") }}
            </button>
            <button
              type="button"
              class="portfolio-ledger-new__side-btn"
              :class="{ 'is-active': draft.side === 'SELL' }"
              @click="ticket.setSide('SELL')"
            >
              {{ t("portfolio.ledgerNew.sell") }}
            </button>
          </div>

          <div class="portfolio-ledger-new__field">
            <label class="portfolio-ledger-new__label">{{ t("portfolio.ledgerNew.instrument") }}</label>
            <InstrumentCombobox
              :selected="selectedInstrument"
              :error-message="fieldMessage('instrument_id')"
              input-id="ledger-new-instrument"
              @select="chooseInstrument"
              @clear="clearInstrument"
            />
          </div>

          <div class="portfolio-ledger-new__grid">
            <div class="portfolio-ledger-new__field">
              <label class="portfolio-ledger-new__label" for="ledger-new-qty">{{ t("portfolio.ledgerNew.quantity") }}</label>
              <input
                id="ledger-new-qty"
                :value="draft.quantity"
                type="text"
                inputmode="decimal"
                class="portfolio-ledger-new__input"
                @input="ticket.patchDraft({ quantity: ($event.target as HTMLInputElement).value })"
              />
              <span v-if="fieldMessage('quantity')" class="portfolio-ledger-new__field-error">{{ fieldMessage('quantity') }}</span>
            </div>
            <div class="portfolio-ledger-new__field">
              <label class="portfolio-ledger-new__label" for="ledger-new-price">{{ t("portfolio.ledgerNew.price") }}</label>
              <input
                id="ledger-new-price"
                :value="draft.price"
                type="text"
                inputmode="decimal"
                class="portfolio-ledger-new__input"
                @input="ticket.patchDraft({ price: ($event.target as HTMLInputElement).value })"
              />
              <span v-if="fieldMessage('price')" class="portfolio-ledger-new__field-error">{{ fieldMessage('price') }}</span>
            </div>
            <div class="portfolio-ledger-new__field">
              <label class="portfolio-ledger-new__label" for="ledger-new-fees">{{ t("portfolio.ledgerNew.fees") }}</label>
              <input
                id="ledger-new-fees"
                :value="draft.fees"
                type="text"
                inputmode="decimal"
                class="portfolio-ledger-new__input"
                @input="ticket.patchDraft({ fees: ($event.target as HTMLInputElement).value })"
              />
              <span v-if="fieldMessage('fees')" class="portfolio-ledger-new__field-error">{{ fieldMessage('fees') }}</span>
            </div>
            <div class="portfolio-ledger-new__field">
              <label class="portfolio-ledger-new__label" for="ledger-new-currency">{{ t("portfolio.ledgerNew.currency") }}</label>
              <input
                id="ledger-new-currency"
                :value="draft.currency"
                type="text"
                maxlength="3"
                class="portfolio-ledger-new__input portfolio-ledger-new__input--upper"
                @input="ticket.patchDraft({ currency: ($event.target as HTMLInputElement).value.toUpperCase() })"
              />
              <span v-if="fieldMessage('currency')" class="portfolio-ledger-new__field-error">{{ fieldMessage('currency') }}</span>
            </div>
            <div class="portfolio-ledger-new__field">
              <label class="portfolio-ledger-new__label" for="ledger-new-business-date">{{ t("portfolio.ledgerNew.businessDate") }}</label>
              <input
                id="ledger-new-business-date"
                :value="draft.business_date"
                type="date"
                class="portfolio-ledger-new__input"
                @input="ticket.patchDraft({ business_date: ($event.target as HTMLInputElement).value })"
              />
              <span v-if="fieldMessage('business_date')" class="portfolio-ledger-new__field-error">{{ fieldMessage('business_date') }}</span>
            </div>
            <div class="portfolio-ledger-new__field">
              <label class="portfolio-ledger-new__label" for="ledger-new-settlement-date">
                {{ t("portfolio.ledgerNew.settlementDate") }}
              </label>
              <input
                id="ledger-new-settlement-date"
                :value="draft.settlement_date"
                type="date"
                class="portfolio-ledger-new__input"
                @input="ticket.patchDraft({ settlement_date: ($event.target as HTMLInputElement).value })"
              />
              <span v-if="fieldMessage('settlement_date')" class="portfolio-ledger-new__field-error">{{ fieldMessage('settlement_date') }}</span>
            </div>
          </div>

          <div class="portfolio-ledger-new__field">
            <label class="portfolio-ledger-new__label" for="ledger-new-reason">{{ t("portfolio.ledgerNew.reason") }}</label>
            <input
              id="ledger-new-reason"
              :value="draft.reason"
              type="text"
              class="portfolio-ledger-new__input"
              @input="ticket.patchDraft({ reason: ($event.target as HTMLInputElement).value })"
            />
          </div>

          <div class="portfolio-ledger-new__actions">
            <IMSPermissionGuard permission="INVESTMENT_LEDGER_SIMULATE">
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :disabled="!ticket.isValid.value || ticket.simulating.value"
                @click="onSimulate"
              >
                {{ ticket.lastSimulation.value ? t("portfolio.ledgerNew.resimulate") : t("portfolio.ledgerNew.simulate") }}
              </button>
            </IMSPermissionGuard>

            <IMSPermissionGuard permission="INVESTMENT_LEDGER_POST">
              <button
                type="button"
                class="btn btn-primary btn-sm"
                :disabled="!ticket.canPost.value"
                @click="onRequestPost"
              >
                {{ intent === "PAPER" ? t("portfolio.ledgerNew.postPaper") : t("portfolio.ledgerNew.postOfficial") }}
              </button>
            </IMSPermissionGuard>
          </div>

          <p v-if="ticket.simulationError.value" class="portfolio-ledger-new__alert" role="alert">
            {{ t("portfolio.ledgerNew.simulationFailedLabel") }}: {{ ticket.simulationError.value }}
          </p>
          <p v-else-if="ticket.postError.value" class="portfolio-ledger-new__alert" role="alert">
            {{ t("portfolio.ledgerNew.postFailedLabel") }}: {{ ticket.postError.value }}
          </p>
          <p v-else-if="ticket.successMessage.value" class="portfolio-ledger-new__success" role="status">
            {{ t("portfolio.ledgerNew.postedSuccess") }}
          </p>
          <p
            v-else-if="ticket.lastSimulation.value && !ticket.isSimulationFresh.value"
            class="portfolio-ledger-new__hint"
          >
            {{ t("portfolio.ledgerNew.staleSimulationHint") }}
          </p>

          <div v-if="ticket.lastSimulation.value" class="portfolio-ledger-new__preview">
            <div class="portfolio-ledger-new__preview-row">
              <span>{{ t("portfolio.ledgerNew.verdict") }}</span>
              <strong :data-verdict="verdict">{{ verdict ?? t("portfolio.terminal.unavailable") }}</strong>
            </div>
            <div class="portfolio-ledger-new__preview-row">
              <span>{{ t("portfolio.ledgerNew.cashImpact") }}</span>
              <strong>{{ ticket.lastSimulation.value.cash?.cash_impact ?? t("portfolio.terminal.unavailable") }}</strong>
            </div>
            <div class="portfolio-ledger-new__preview-row">
              <span>{{ t("portfolio.ledgerNew.projectedPosition") }}</span>
              <strong>{{ ticket.lastSimulation.value.position?.projected_quantity ?? t("portfolio.terminal.unavailable") }}</strong>
            </div>
          </div>
        </form>
      </template>
    </AppCard>

    <AppConfirmDialog
      :open="ticket.stage.value === 'confirming'"
      :title="t('portfolio.ledgerNew.confirmTitle')"
      :description="intent === 'PAPER' ? t('portfolio.ledgerNew.confirmDescriptionPaper') : t('portfolio.ledgerNew.confirmDescriptionOfficial')"
      :confirm-label="intent === 'PAPER' ? t('portfolio.ledgerNew.postPaper') : t('portfolio.ledgerNew.postOfficial')"
      :cancel-label="t('portfolio.ledgerNew.confirmCancel')"
      tone="warning"
      :loading="ticket.posting.value"
      @cancel="onCancelPost"
      @confirm="onConfirmPost"
    />
  </section>
</template>

<style scoped>
.portfolio-ledger-new {
  display: grid;
  gap: var(--space-4, 16px);
}

.portfolio-ledger-new__notice,
.portfolio-ledger-new__error {
  padding: var(--space-4, 16px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
}

.portfolio-ledger-new__error {
  color: var(--alert-danger-text, #cf222e);
}

.portfolio-ledger-new__blocked {
  padding: 14px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  background: var(--bg-card-muted, #f6f8fa);
  color: var(--text-secondary, #57606a);
  display: grid;
  gap: 6px;
}

.portfolio-ledger-new__intent-banner {
  padding: 8px 12px;
  margin-bottom: 14px;
  font-size: 12px;
  font-weight: 600;
  border-radius: var(--radius-md, 6px);
  border: 1px solid var(--border-subtle, #d0d7de);
  background: var(--bg-card-muted, #f6f8fa);
  color: var(--text-secondary, #57606a);
}

.portfolio-ledger-new__intent-banner[data-intent="PAPER"] {
  background: var(--alert-warning-bg, #fff8c5);
  border-color: var(--alert-warning-border, #9a6700);
  color: var(--alert-warning-text, #9a6700);
}

.portfolio-ledger-new__intent-banner[data-intent="OFFICIAL"] {
  background: var(--alert-success-bg, #dafbe1);
  border-color: var(--alert-success-border, #1a7f37);
  color: var(--alert-success-text, #1a7f37);
}

.portfolio-ledger-new__form {
  display: grid;
  gap: 14px;
}

.portfolio-ledger-new__side {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px;
  max-width: 260px;
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  padding: 4px;
}

.portfolio-ledger-new__side-btn {
  border: none;
  background: transparent;
  padding: 8px;
  font-size: 12px;
  font-weight: 700;
  color: var(--text-secondary, #57606a);
  cursor: pointer;
  border-radius: var(--radius-sm, 4px);
}

.portfolio-ledger-new__side-btn.is-active {
  background: var(--bg-card, #ffffff);
  color: var(--text-primary, #1f2328);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
}

.portfolio-ledger-new__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px;
}

.portfolio-ledger-new__field {
  display: grid;
  gap: 4px;
}

.portfolio-ledger-new__label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.02em;
  color: var(--text-secondary, #57606a);
}

.portfolio-ledger-new__input {
  height: 34px;
  padding: 0 10px;
  border: 1px solid var(--border-default, #d0d7de);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-input, #ffffff);
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary, #1f2328);
}

.portfolio-ledger-new__input--upper {
  text-transform: uppercase;
}

.portfolio-ledger-new__field-error {
  font-size: 11px;
  color: var(--state-danger, #cf222e);
}

.portfolio-ledger-new__actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.portfolio-ledger-new__alert {
  margin: 0;
  padding: 8px 12px;
  border-radius: var(--radius-sm, 4px);
  font-size: 12px;
  color: var(--alert-danger-text, #cf222e);
  background: var(--alert-danger-bg, #ffebe9);
  border: 1px solid var(--alert-danger-border, #cf222e);
}

.portfolio-ledger-new__success {
  margin: 0;
  padding: 8px 12px;
  border-radius: var(--radius-sm, 4px);
  font-size: 12px;
  color: var(--alert-success-text, #1a7f37);
  background: var(--alert-success-bg, #dafbe1);
  border: 1px solid var(--alert-success-border, #1a7f37);
}

.portfolio-ledger-new__hint {
  margin: 0;
  font-size: 12px;
  color: var(--text-tertiary, #6e7781);
}

.portfolio-ledger-new__preview {
  display: grid;
  gap: 8px;
  padding: 12px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  background: var(--bg-card-muted, #f6f8fa);
}

.portfolio-ledger-new__preview-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
}

.portfolio-ledger-new__preview-row strong {
  font-variant-numeric: tabular-nums;
}

.portfolio-ledger-new__preview-row strong[data-verdict="PASS"] {
  color: var(--alert-success-text, #1a7f37);
}

.portfolio-ledger-new__preview-row strong[data-verdict="WARN"] {
  color: var(--alert-warning-text, #9a6700);
}

.portfolio-ledger-new__preview-row strong[data-verdict="BLOCK"] {
  color: var(--alert-danger-text, #cf222e);
}
</style>
