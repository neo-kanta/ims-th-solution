<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from "vue";

import AppButton from "~/shared/ui/AppButton.vue";
import AppConfirmDialog from "~/shared/ui/AppConfirmDialog.vue";

import { useInstrumentDirectory } from "../composables/useInstrumentDirectory";
import { type useOrderTicket } from "../composables/useOrderTicket";
import {
  formatMoney,
  formatQuantity,
  shortenId,
  verdictLabel,
} from "../lib/ledgerFormat";
import type {
  ApiInstrument,
  ApiPortfolio,
} from "../services/investmentLedgerApi";
import CashProjectionPanel from "./CashProjectionPanel.vue";
import CompliancePreviewPanel from "./CompliancePreviewPanel.vue";
import PositionProjectionPanel from "./PositionProjectionPanel.vue";

type Ticket = ReturnType<typeof useOrderTicket>;

interface Props {
  open: boolean;
  portfolio: ApiPortfolio | null;
  ticket: Ticket;
  canSimulate?: boolean;
  canPost?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  canSimulate: true,
  canPost: true,
});

const emit = defineEmits<{
  close: [];
  posted: [];
}>();

const instruments = useInstrumentDirectory();
const searchQuery = ref("");
const searchDebounce = ref<ReturnType<typeof setTimeout> | null>(null);

const draft = props.ticket.draft;

const selectedInstrument = computed<ApiInstrument | null>(() => {
  if (!draft.instrument_id) return null;
  return instruments.findById(draft.instrument_id);
});

const grossPreview = computed(() => {
  const qty = Number(draft.quantity);
  const px = Number(draft.price);
  if (!Number.isFinite(qty) || !Number.isFinite(px)) return null;
  return qty * px;
});

const netPreview = computed(() => {
  if (grossPreview.value === null) return null;
  const fees = Number(draft.fees || 0);
  const safeFees = Number.isFinite(fees) ? fees : 0;
  return draft.side === "BUY"
    ? grossPreview.value + safeFees
    : grossPreview.value - safeFees;
});

watch(
  () => props.open,
  async (isOpen) => {
    if (!isOpen) return;
    searchQuery.value = "";
    await instruments.ensureLoaded();
  },
);

watch(searchQuery, (value) => {
  if (searchDebounce.value) clearTimeout(searchDebounce.value);
  searchDebounce.value = setTimeout(() => {
    void instruments.search(value);
  }, 250);
});

watch(
  () => draft.instrument_id,
  (id) => {
    // Once an instrument is chosen, default the currency to the instrument's
    // own currency if the operator hasn't set one yet — saves a keystroke
    // for the common case (single-currency trades).
    const inst = instruments.findById(id);
    if (inst?.currency && !draft.currency) {
      props.ticket.patchDraft({ currency: inst.currency });
    }
  },
);

const firstFocusRef = ref<HTMLButtonElement | null>(null);

onMounted(async () => {
  if (props.open) {
    await instruments.ensureLoaded();
  }
});

watch(
  () => props.open,
  async (isOpen) => {
    if (!isOpen) return;
    await nextTick();
    firstFocusRef.value?.focus();
  },
);

function chooseInstrument(inst: ApiInstrument) {
  if (!inst.id) return;
  props.ticket.patchDraft({
    instrument_id: inst.id,
    currency: draft.currency || inst.currency || "",
  });
}

function clearInstrument() {
  props.ticket.patchDraft({ instrument_id: "" });
}

function onSimulate() {
  void props.ticket.simulate();
}

function onRequestPost() {
  props.ticket.requestConfirmation();
}

function onCancelPost() {
  props.ticket.cancelConfirmation();
}

async function onConfirmPost() {
  const result = await props.ticket.post();
  if (result) {
    emit("posted");
  }
}

function onClose() {
  if (props.ticket.posting.value) return;
  emit("close");
}

const verdictBadgeClass = computed(() => {
  const v = props.ticket.verdict.value;
  if (v === "PASS") return "order-ticket__verdict-pill order-ticket__verdict-pill--ok";
  if (v === "WARN") return "order-ticket__verdict-pill order-ticket__verdict-pill--warn";
  if (v === "BLOCK") return "order-ticket__verdict-pill order-ticket__verdict-pill--block";
  return "order-ticket__verdict-pill";
});

const showPostPermissionHint = computed(() =>
  !props.canPost &&
  Boolean(props.ticket.lastSimulation.value) &&
  props.ticket.isSimulationFresh.value &&
  props.ticket.stage.value !== "posted",
);

const errors = computed(() => props.ticket.validationErrors.value);
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="order-ticket-overlay" role="presentation" @click.self="onClose">
      <aside
        class="order-ticket"
        role="dialog"
        aria-modal="true"
        aria-labelledby="order-ticket-title"
      >
        <header class="order-ticket__header">
          <div class="order-ticket__header-text">
            <h2 id="order-ticket-title" class="order-ticket__title">Order ticket</h2>
            <p class="order-ticket__subtitle">
              <span v-if="portfolio">
                {{ portfolio.code ?? portfolio.id }}
                <span v-if="portfolio.name"> · {{ portfolio.name }}</span>
              </span>
              <span v-else>Select a portfolio first.</span>
            </p>
          </div>
          <AppButton variant="ghost" size="sm" :disabled="ticket.posting.value" @click="onClose">
            Close
          </AppButton>
        </header>

        <section class="order-ticket__body">
          <fieldset class="order-ticket__section">
            <legend class="order-ticket__legend">Order</legend>

            <div class="order-ticket__side">
              <button
                ref="firstFocusRef"
                type="button"
                class="order-ticket__side-btn"
                :class="{ 'order-ticket__side-btn--buy': draft.side === 'BUY' }"
                @click="ticket.setSide('BUY')"
              >
                BUY
              </button>
              <button
                type="button"
                class="order-ticket__side-btn"
                :class="{ 'order-ticket__side-btn--sell': draft.side === 'SELL' }"
                @click="ticket.setSide('SELL')"
              >
                SELL
              </button>
            </div>

            <div class="order-ticket__field">
              <label class="order-ticket__label" for="order-ticket-instrument">
                Instrument
              </label>
              <div v-if="selectedInstrument" class="order-ticket__instrument-picked">
                <div>
                  <div class="order-ticket__instrument-name">
                    {{ selectedInstrument.primary_ticker ?? selectedInstrument.name }}
                  </div>
                  <div class="order-ticket__instrument-meta">
                    {{ selectedInstrument.name }}
                    <span v-if="selectedInstrument.primary_exchange">
                      · {{ selectedInstrument.primary_exchange }}
                    </span>
                    <span v-if="selectedInstrument.currency">
                      · {{ selectedInstrument.currency }}
                    </span>
                  </div>
                </div>
                <AppButton variant="ghost" size="xs" @click="clearInstrument">Change</AppButton>
              </div>
              <template v-else>
                <input
                  id="order-ticket-instrument"
                  v-model="searchQuery"
                  type="search"
                  autocomplete="off"
                  class="order-ticket__input"
                  placeholder="Search by ticker, ISIN, or name"
                />
                <div class="order-ticket__instrument-list" role="listbox">
                  <p v-if="instruments.loading.value" class="order-ticket__instrument-hint">
                    Searching…
                  </p>
                  <p v-else-if="instruments.error.value" class="order-ticket__instrument-error">
                    {{ instruments.error.value }}
                  </p>
                  <p
                    v-else-if="instruments.items.value.length === 0"
                    class="order-ticket__instrument-hint"
                  >
                    No instruments match this search.
                  </p>
                  <button
                    v-for="inst in instruments.items.value.slice(0, 12)"
                    :key="inst.id"
                    type="button"
                    role="option"
                    class="order-ticket__instrument-option"
                    @click="chooseInstrument(inst)"
                  >
                    <span class="order-ticket__instrument-option-primary">
                      {{ inst.primary_ticker ?? inst.name ?? shortenId(inst.id) }}
                    </span>
                    <span class="order-ticket__instrument-option-secondary">
                      {{ inst.name ?? "" }}
                      <span v-if="inst.primary_exchange"> · {{ inst.primary_exchange }}</span>
                      <span v-if="inst.currency"> · {{ inst.currency }}</span>
                    </span>
                  </button>
                </div>
                <span v-if="errors.instrument_id" class="order-ticket__error">
                  {{ errors.instrument_id }}
                </span>
              </template>
            </div>

            <div class="order-ticket__grid">
              <div class="order-ticket__field">
                <label class="order-ticket__label" for="order-ticket-qty">Quantity</label>
                <input
                  id="order-ticket-qty"
                  :value="draft.quantity"
                  inputmode="decimal"
                  type="text"
                  class="order-ticket__input"
                  placeholder="0"
                  @input="ticket.patchDraft({ quantity: ($event.target as HTMLInputElement).value })"
                />
                <span v-if="errors.quantity" class="order-ticket__error">{{ errors.quantity }}</span>
              </div>
              <div class="order-ticket__field">
                <label class="order-ticket__label" for="order-ticket-price">Price</label>
                <input
                  id="order-ticket-price"
                  :value="draft.price"
                  inputmode="decimal"
                  type="text"
                  class="order-ticket__input"
                  placeholder="0.00"
                  @input="ticket.patchDraft({ price: ($event.target as HTMLInputElement).value })"
                />
                <span v-if="errors.price" class="order-ticket__error">{{ errors.price }}</span>
              </div>
              <div class="order-ticket__field">
                <label class="order-ticket__label" for="order-ticket-fees">Fees</label>
                <input
                  id="order-ticket-fees"
                  :value="draft.fees"
                  inputmode="decimal"
                  type="text"
                  class="order-ticket__input"
                  placeholder="0.00"
                  @input="ticket.patchDraft({ fees: ($event.target as HTMLInputElement).value })"
                />
                <span v-if="errors.fees" class="order-ticket__error">{{ errors.fees }}</span>
              </div>
              <div class="order-ticket__field">
                <label class="order-ticket__label" for="order-ticket-currency">Currency</label>
                <input
                  id="order-ticket-currency"
                  :value="draft.currency"
                  type="text"
                  maxlength="3"
                  class="order-ticket__input order-ticket__input--upper"
                  placeholder="THB"
                  @input="
                    ticket.patchDraft({
                      currency: ($event.target as HTMLInputElement).value.toUpperCase(),
                    })
                  "
                />
                <span v-if="errors.currency" class="order-ticket__error">{{ errors.currency }}</span>
              </div>
              <div class="order-ticket__field">
                <label class="order-ticket__label" for="order-ticket-business-date">
                  Business date
                </label>
                <input
                  id="order-ticket-business-date"
                  :value="draft.business_date"
                  type="date"
                  class="order-ticket__input"
                  @input="ticket.patchDraft({ business_date: ($event.target as HTMLInputElement).value })"
                />
                <span v-if="errors.business_date" class="order-ticket__error">
                  {{ errors.business_date }}
                </span>
              </div>
              <div class="order-ticket__field">
                <label class="order-ticket__label" for="order-ticket-settle-date">
                  Settlement date <span class="order-ticket__label-hint">(optional)</span>
                </label>
                <input
                  id="order-ticket-settle-date"
                  :value="draft.settlement_date"
                  type="date"
                  class="order-ticket__input"
                  @input="ticket.patchDraft({ settlement_date: ($event.target as HTMLInputElement).value })"
                />
                <span v-if="errors.settlement_date" class="order-ticket__error">
                  {{ errors.settlement_date }}
                </span>
              </div>
            </div>

            <div class="order-ticket__grid">
              <div class="order-ticket__field">
                <label class="order-ticket__label" for="order-ticket-ext-ref">
                  External reference <span class="order-ticket__label-hint">(optional)</span>
                </label>
                <input
                  id="order-ticket-ext-ref"
                  :value="draft.external_ref"
                  type="text"
                  class="order-ticket__input"
                  placeholder="e.g. broker order id"
                  @input="ticket.patchDraft({ external_ref: ($event.target as HTMLInputElement).value })"
                />
              </div>
              <div class="order-ticket__field">
                <label class="order-ticket__label" for="order-ticket-reason">
                  Reason / note <span class="order-ticket__label-hint">(optional)</span>
                </label>
                <input
                  id="order-ticket-reason"
                  :value="draft.reason"
                  type="text"
                  class="order-ticket__input"
                  placeholder="Notes captured with the transaction"
                  @input="ticket.patchDraft({ reason: ($event.target as HTMLInputElement).value })"
                />
              </div>
            </div>

            <div class="order-ticket__preview-strip">
              <div>
                <span class="order-ticket__preview-label">Gross (qty × price)</span>
                <span class="order-ticket__preview-value">
                  {{ formatMoney(grossPreview, draft.currency) }}
                </span>
              </div>
              <div>
                <span class="order-ticket__preview-label">Net (incl. fees)</span>
                <span class="order-ticket__preview-value">
                  {{ formatMoney(netPreview, draft.currency) }}
                </span>
              </div>
              <div>
                <span class="order-ticket__preview-label">Quantity</span>
                <span class="order-ticket__preview-value">
                  {{ formatQuantity(draft.quantity) }}
                </span>
              </div>
            </div>
          </fieldset>

          <fieldset class="order-ticket__section">
            <legend class="order-ticket__legend">Simulation preview</legend>

            <div class="order-ticket__simulate-row">
              <AppButton
                variant="primary"
                size="sm"
                :loading="ticket.simulating.value"
                :disabled="!ticket.isValid.value || ticket.simulating.value || !canSimulate"
                @click="onSimulate"
              >
                {{ ticket.lastSimulation.value ? "Re-simulate" : "Simulate" }}
              </AppButton>
              <span v-if="!canSimulate" class="order-ticket__hint">
                You don't have the INVESTMENT_LEDGER_SIMULATE permission.
              </span>
              <span v-else-if="!ticket.isValid.value" class="order-ticket__hint">
                Fix the highlighted fields to enable the simulation.
              </span>
              <span
                v-else-if="ticket.lastSimulation.value && !ticket.isSimulationFresh.value"
                class="order-ticket__hint order-ticket__hint--warn"
              >
                Form changed since the last simulation — re-run before posting.
              </span>
              <span
                v-else-if="ticket.verdict.value"
                :class="verdictBadgeClass"
                aria-live="polite"
              >
                {{ verdictLabel(ticket.verdict.value) }}
              </span>
            </div>

            <p
              v-if="ticket.simulationError.value"
              class="order-ticket__alert order-ticket__alert--danger"
              role="alert"
            >
              {{ ticket.simulationError.value }}
            </p>

            <div class="order-ticket__preview-grid">
              <div class="order-ticket__preview-card">
                <CashProjectionPanel
                  :projection="ticket.lastSimulation.value?.cash"
                  :loading="ticket.simulating.value"
                />
              </div>
              <div class="order-ticket__preview-card">
                <PositionProjectionPanel
                  :projection="ticket.lastSimulation.value?.position"
                  :instrument="selectedInstrument"
                  :currency="draft.currency"
                  :loading="ticket.simulating.value"
                />
              </div>
              <div class="order-ticket__preview-card order-ticket__preview-card--wide">
                <CompliancePreviewPanel
                  :compliance="ticket.lastSimulation.value?.compliance"
                  :loading="ticket.simulating.value"
                />
              </div>
            </div>
          </fieldset>
        </section>

        <footer class="order-ticket__footer">
          <div class="order-ticket__footer-status">
            <p v-if="ticket.postError.value" class="order-ticket__alert order-ticket__alert--danger" role="alert">
              {{ ticket.postError.value }}
            </p>
            <p v-else-if="ticket.successMessage.value" class="order-ticket__alert order-ticket__alert--success">
              {{ ticket.successMessage.value }}
            </p>
            <p v-else-if="ticket.isBlocked.value" class="order-ticket__alert order-ticket__alert--danger">
              Compliance returned BLOCK — the order cannot be posted as drafted.
            </p>
          </div>
          <div class="order-ticket__footer-actions">
            <AppButton variant="secondary" size="sm" :disabled="ticket.posting.value" @click="onClose">
              {{ ticket.stage.value === "posted" ? "Done" : "Cancel" }}
            </AppButton>
            <span v-if="showPostPermissionHint" class="order-ticket__post-hint">
              Posting requires <code>INVESTMENT_LEDGER_POST</code>.
            </span>
            <AppButton
              variant="primary"
              size="sm"
              :loading="ticket.posting.value"
              :disabled="!ticket.canPost.value || !canPost"
              @click="onRequestPost"
            >
              Post transaction
            </AppButton>
          </div>
        </footer>

        <AppConfirmDialog
          :open="ticket.stage.value === 'confirming'"
          :title="`Post ${draft.side} order?`"
          :description="`Posting will commit the ${draft.side} of ${formatQuantity(draft.quantity)} @ ${formatMoney(draft.price, draft.currency)} for ${selectedInstrument?.primary_ticker ?? 'the selected instrument'}. Cash impact: ${formatMoney(
            ticket.lastSimulation.value?.cash?.cash_impact ?? null,
            draft.currency,
          )}. This cannot be undone from this drawer — use the ledger's reverse action if needed.`"
          confirm-label="Post transaction"
          cancel-label="Back to ticket"
          tone="warning"
          :loading="ticket.posting.value"
          @cancel="onCancelPost"
          @confirm="onConfirmPost"
        />
      </aside>
    </div>
  </Teleport>
</template>

<style scoped>
.order-ticket-overlay {
  position: fixed;
  inset: 0;
  background: var(--bg-overlay);
  display: flex;
  justify-content: flex-end;
  z-index: var(--z-modal);
}

.order-ticket {
  position: relative;
  width: min(720px, 100%);
  height: 100%;
  background: var(--bg-card);
  border-left: 1px solid var(--border-subtle);
  display: grid;
  grid-template-rows: auto 1fr auto;
  box-shadow: -2px 0 12px rgba(0, 0, 0, 0.08);
}

.order-ticket__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--border-subtle);
}

.order-ticket__title {
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.order-ticket__subtitle {
  margin: 4px 0 0;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.order-ticket__body {
  overflow-y: auto;
  padding: var(--space-4) var(--space-5);
  display: grid;
  gap: var(--space-5);
}

.order-ticket__section {
  border: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: var(--space-3);
}

.order-ticket__legend {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-tertiary);
  padding: 0;
}

.order-ticket__side {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-2);
  background: var(--bg-card-muted);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 4px;
}

.order-ticket__side-btn {
  border: none;
  background: transparent;
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: background var(--transition-fast), color var(--transition-fast);
}

.order-ticket__side-btn--buy {
  background: var(--state-success, #059669);
  color: var(--text-on-primary, #fff);
}

.order-ticket__side-btn--sell {
  background: var(--state-danger, #dc2626);
  color: var(--text-on-primary, #fff);
}

.order-ticket__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: var(--space-3);
}

.order-ticket__field {
  display: grid;
  gap: 4px;
}

.order-ticket__label {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.02em;
  text-transform: uppercase;
  color: var(--text-secondary);
}

.order-ticket__label-hint {
  font-weight: var(--font-weight-regular);
  text-transform: none;
  letter-spacing: 0;
  color: var(--text-tertiary);
}

.order-ticket__input {
  height: 36px;
  padding: 0 var(--space-3);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  font-size: var(--font-size-sm);
  font-family: inherit;
  color: var(--text-primary);
}

.order-ticket__input:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: var(--shadow-focus);
}

.order-ticket__input--upper {
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.order-ticket__instrument-picked {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
}

.order-ticket__instrument-name {
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.order-ticket__instrument-meta {
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
  margin-top: 2px;
}

.order-ticket__instrument-list {
  max-height: 220px;
  overflow-y: auto;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  display: grid;
  gap: 1px;
}

.order-ticket__instrument-hint,
.order-ticket__instrument-error {
  margin: 0;
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-sm);
  color: var(--text-tertiary);
}

.order-ticket__instrument-error {
  color: var(--state-danger, #dc2626);
}

.order-ticket__instrument-option {
  text-align: left;
  padding: var(--space-2) var(--space-3);
  border: none;
  background: transparent;
  cursor: pointer;
  display: grid;
  gap: 2px;
  transition: background var(--transition-fast);
}

.order-ticket__instrument-option:hover,
.order-ticket__instrument-option:focus {
  background: var(--bg-card-muted);
  outline: none;
}

.order-ticket__instrument-option-primary {
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}

.order-ticket__instrument-option-secondary {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.order-ticket__preview-strip {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  background: var(--bg-card-muted);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
}

.order-ticket__preview-strip > div {
  display: grid;
  gap: 2px;
}

.order-ticket__preview-label {
  font-size: var(--font-size-xs);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary);
}

.order-ticket__preview-value {
  font-variant-numeric: tabular-nums;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
}

.order-ticket__simulate-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.order-ticket__hint {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.order-ticket__hint--warn {
  color: var(--state-warning, #b45309);
}

.order-ticket__verdict-pill {
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: var(--font-weight-bold);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  border: 1px solid transparent;
  background: var(--bg-card-muted);
  color: var(--text-secondary);
}

.order-ticket__verdict-pill--ok {
  background: var(--alert-success-bg);
  color: var(--alert-success-text);
  border-color: var(--alert-success-border);
}

.order-ticket__verdict-pill--warn {
  background: var(--alert-warning-bg);
  color: var(--alert-warning-text);
  border-color: var(--alert-warning-border);
}

.order-ticket__verdict-pill--block {
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
  border-color: var(--alert-danger-border);
}

.order-ticket__preview-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-3);
}

.order-ticket__preview-card {
  padding: var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.order-ticket__preview-card--wide {
  grid-column: 1 / -1;
}

.order-ticket__error {
  font-size: var(--font-size-xs);
  color: var(--state-danger, #dc2626);
}

.order-ticket__alert {
  margin: 0;
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-sm);
  border: 1px solid transparent;
}

.order-ticket__alert--danger {
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
  border-color: var(--alert-danger-border);
}

.order-ticket__alert--success {
  background: var(--alert-success-bg);
  color: var(--alert-success-text);
  border-color: var(--alert-success-border);
}

.order-ticket__footer {
  border-top: 1px solid var(--border-subtle);
  padding: var(--space-3) var(--space-5);
  background: var(--bg-card-muted);
  display: grid;
  gap: var(--space-2);
}

.order-ticket__footer-status {
  display: grid;
  gap: var(--space-2);
}

.order-ticket__footer-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.order-ticket__post-hint {
  max-width: 260px;
  font-size: var(--font-size-xs);
  line-height: 1.4;
  color: var(--text-tertiary);
}

.order-ticket__post-hint code {
  font-size: 0.92em;
  color: var(--text-secondary);
}

@media (max-width: 720px) {
  .order-ticket__preview-grid {
    grid-template-columns: 1fr;
  }
  .order-ticket__preview-strip {
    grid-template-columns: 1fr;
  }
}
</style>
