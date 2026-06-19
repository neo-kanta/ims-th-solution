<script setup lang="ts">
/**
 * /investment/funds/{fundId}/operation
 *
 * The fund's Operation tab. Currently exposes one capability — a BUY/SELL
 * trade ticket against any portfolio under the fund. Future operations
 * (transfers, FX conversions, subscription/redemption) will live next to
 * the ticket on this same page.
 *
 * Flow:
 *   * If fund.require_pretrade_preview is true → user must click
 *     "Run pre-trade check" before the Post button activates.
 *   * Else → user can post directly (the server still enforces the
 *     pre-trade gates inside the post handler).
 */
import { computed, onMounted, ref, watch } from "vue";

import AppCard from "~/shared/ui/AppCard.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";
import { useI18n } from "~/composables/useI18n";
import { myFundsApi, useFundDetail } from "~/features/my-funds";
import type { ApiPortfolio } from "~/features/my-funds";
import type {
  ApiInstrument,
  ApiTransaction,
  ApiTransactionSimulation,
  PostTransactionPayload,
} from "~/features/my-funds/services/myFundsApi";
import { OpenApiRequestError } from "~/api/openapi";
import { useWorkflowStore } from "~/features/workflow/store/useWorkflowStore";
import { useAuthStore } from "~/stores/useAuthStore";
import { useOpenApiClient } from "~/api/openapi";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_FUND_VIEW",
});

const { t } = useI18n();
const route = useRoute();
const authStore = useAuthStore();
const workflowStore = useWorkflowStore();

const fundIdParam = computed(() => String(route.params.fundId ?? ""));
const fundDetail = useFundDetail(() => fundIdParam.value);

// ─── Ticket state ────────────────────────────────────────────────────────
const portfolioId = ref<string>("");
const side = ref<"BUY" | "SELL">("BUY");
const instrumentId = ref<string>("");
const quantity = ref<string>("");
const price = ref<string>("");
const fees = ref<string>("0");
const businessDate = ref<string>(new Date().toISOString().slice(0, 10));
const settlementDate = ref<string>("");
const reason = ref<string>("");

const instruments = ref<ApiInstrument[]>([]);
const instrumentSearch = ref<string>("");

const simulation = ref<ApiTransactionSimulation | null>(null);
const pretradeError = ref<string | null>(null);
const simulating = ref(false);

const posted = ref<ApiTransaction | null>(null);
const postError = ref<string | null>(null);
const posting = ref(false);

const portfolios = computed(() => fundDetail.portfolios.value);
const portfolio = computed<ApiPortfolio | undefined>(
  () => portfolios.value.find((p) => p.id === portfolioId.value),
);
const currency = computed(
  () => portfolio.value?.valuation_currency ?? portfolio.value?.base_currency ?? fundDetail.card.value?.base_currency ?? "THB",
);

const requirePretrade = computed(
  () => fundDetail.fund.value?.require_pretrade_preview === true,
);

const filteredInstruments = computed<ApiInstrument[]>(() => {
  const q = instrumentSearch.value.trim().toLowerCase();
  if (!q) return instruments.value.slice(0, 12);
  return instruments.value
    .filter(
      (i) =>
        (i.primary_ticker ?? "").toLowerCase().includes(q) ||
        (i.name ?? "").toLowerCase().includes(q),
    )
    .slice(0, 12);
});

const grossAmount = computed(() => {
  const q = Number(quantity.value);
  const p = Number(price.value);
  if (!Number.isFinite(q) || !Number.isFinite(p)) return 0;
  return q * p;
});

const netAmount = computed(() => {
  const f = Number(fees.value);
  const feesNum = Number.isFinite(f) ? f : 0;
  return side.value === "BUY" ? grossAmount.value + feesNum : grossAmount.value - feesNum;
});

const isLocked = computed(() => workflowStore.transactionsLocked);
const canSimulate = computed(() => authStore.hasPermission("INVESTMENT_LEDGER_SIMULATE"));
const canPostTx = computed(() => authStore.hasPermission("INVESTMENT_LEDGER_POST"));

const canRunPretrade = computed(
  () =>
    !!portfolioId.value &&
    !!instrumentId.value &&
    Number(quantity.value) > 0 &&
    Number(price.value) > 0 &&
    !!businessDate.value &&
    !simulating.value &&
    canSimulate.value &&
    !isLocked.value,
);

const verdict = computed<"PASS" | "WARN" | "BLOCK" | null>(() => {
  const c = simulation.value?.compliance;
  if (!c) return null;
  const v = (c.verdict ?? "").toUpperCase();
  if (v === "BLOCK" || v === "WARN" || v === "PASS") return v;
  return null;
});

const canPost = computed(() => {
  if (isLocked.value) return false;
  if (!canPostTx.value) return false;

  // Basic validation checks (same as canRunPretrade but without checking canSimulate)
  const hasValidInputs =
    !!portfolioId.value &&
    !!instrumentId.value &&
    Number(quantity.value) > 0 &&
    Number(price.value) > 0 &&
    !!businessDate.value &&
    !posting.value;

  if (!hasValidInputs) return false;

  if (requirePretrade.value) {
    if (!simulation.value || verdict.value === null) return false;
    if (verdict.value === "BLOCK") return false;
  }
  return true;
});

onMounted(async () => {
  await fundDetail.load();
  const first = portfolios.value[0];
  if (!portfolioId.value && first?.id) {
    portfolioId.value = first.id;
  }
  // Pull the instrument master once — small list, cheap.
  try {
    instruments.value = await myFundsApi.listInstruments(500);
  } catch {
    instruments.value = [];
  }

  // Initialize and fetch daily workflow state
  if (workflowStore.client === null) {
    workflowStore.setClient(useOpenApiClient());
  }
  await workflowStore.fetchState();
});

// When the user changes any field, invalidate the prior simulation so they
// can't post stale results.
watch(
  () => [portfolioId.value, instrumentId.value, side.value, quantity.value, price.value, fees.value, businessDate.value],
  () => {
    simulation.value = null;
    pretradeError.value = null;
    posted.value = null;
    postError.value = null;
  },
);

function buildPayload(): PostTransactionPayload {
  // Intentionally do NOT send gross_amount / net_amount. The backend derives
  // them from qty × price and applies the BUY/SELL sign convention
  // (BUY = -(gross + fees), SELL = +(gross - fees)). Sending pre-computed
  // values would require the frontend to know the convention and would fail
  // validation if the signs disagree.
  return {
    transaction_type: side.value,
    side: side.value,
    instrument_id: instrumentId.value,
    quantity: quantity.value,
    price: price.value,
    currency: currency.value,
    business_date: businessDate.value,
    settlement_date: settlementDate.value || undefined,
    fees: fees.value || "0",
    reason: reason.value || undefined,
  };
}

async function onRunPretrade() {
  if (!canRunPretrade.value || !portfolioId.value) return;
  pretradeError.value = null;
  simulation.value = null;
  simulating.value = true;
  try {
    simulation.value = await myFundsApi.simulateTransaction(portfolioId.value, buildPayload());
  } catch (err) {
    if (err instanceof OpenApiRequestError) {
      pretradeError.value = `${err.status}: ${err.message}`;
    } else if (err instanceof Error) {
      pretradeError.value = err.message;
    } else {
      pretradeError.value = t("operation.errPretrade", "Pre-trade check failed");
    }
  } finally {
    simulating.value = false;
  }
}

async function onPost() {
  if (!canPost.value || !portfolioId.value) return;
  postError.value = null;
  posted.value = null;
  posting.value = true;
  try {
    posted.value = await myFundsApi.postTransaction(portfolioId.value, buildPayload());
    // Reset volatile inputs so the next ticket starts clean. Keep portfolio + side.
    instrumentId.value = "";
    quantity.value = "";
    price.value = "";
    simulation.value = null;
  } catch (err) {
    if (err instanceof OpenApiRequestError) {
      postError.value = `${err.status}: ${err.message}`;
    } else if (err instanceof Error) {
      postError.value = err.message;
    } else {
      postError.value = t("operation.errPost", "Post failed");
    }
  } finally {
    posting.value = false;
  }
}

function instrumentLabel(i: ApiInstrument): string {
  const ticker = i.primary_ticker ?? "—";
  const name = i.name ?? "";
  return name ? `${ticker} · ${name}` : ticker;
}

function fmtMoney(v: number, digits = 2): string {
  if (!Number.isFinite(v)) return "—";
  return v.toLocaleString("en-US", {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  });
}
</script>

<template>
  <section class="op-page">
    <AppPageHeader
      :title="t('operation.title', 'Operation')"
      :description="
        t(
          'operation.subtitle',
          'BUY / SELL trade ticket against any portfolio under this fund. Pre-trade compliance is enforced server-side; the UI can also force an explicit preview based on the fund’s setting.',
        )
      "
    />

    <AppCard
      :title="t('operation.ticketTitle', 'Trade ticket')"
      :subtitle="
        requirePretrade
          ? t('operation.subtitlePretrade', 'This fund requires an explicit pre-trade preview before posting.')
          : t('operation.subtitleDirect', 'This fund posts directly — the server enforces pre-trade gates.')
      "
    >
      <div class="op-ticket">
        <!-- Workflow Lock Ribbon -->
        <div v-if="isLocked" class="op-ticket__lock-banner" role="alert">
          <AppIcon name="lock" size="sm" class="op-ticket__lock-icon" />
          <div class="op-ticket__lock-content">
            <strong>Ledger Locked</strong>
            <span>
              The investment workflow for business date <strong>{{ workflowStore.businessDate }}</strong> is in state <strong>{{ workflowStore.currentStateCode }}</strong>. Transactions are locked and no new operations can be posted.
            </span>
          </div>
        </div>

        <!-- Permission Warning Banners -->
        <div v-if="!canSimulate && !isLocked" class="op-ticket__permission-warning" role="alert">
          <AppIcon name="info" size="xs" />
          <span>You do not have the <code>INVESTMENT_LEDGER_SIMULATE</code> permission. The pre-trade simulator is disabled.</span>
        </div>
        <div v-if="!canPostTx && !isLocked" class="op-ticket__permission-warning" role="alert">
          <AppIcon name="info" size="xs" />
          <span>You do not have the <code>INVESTMENT_LEDGER_POST</code> permission. You will not be able to record transactions.</span>
        </div>

        <!-- BUY / SELL toggle ------------------------------------------- -->
        <div class="op-ticket__row">
          <div class="op-ticket__sides">
            <button
              type="button"
              class="op-ticket__side"
              :class="{ 'is-active': side === 'BUY', 'is-buy': side === 'BUY' }"
              :disabled="isLocked"
              @click="side = 'BUY'"
            >{{ t("operation.buy", "BUY") }}</button>
            <button
              type="button"
              class="op-ticket__side"
              :class="{ 'is-active': side === 'SELL', 'is-sell': side === 'SELL' }"
              :disabled="isLocked"
              @click="side = 'SELL'"
            >{{ t("operation.sell", "SELL") }}</button>
          </div>

          <div class="op-ticket__field">
            <label>{{ t("operation.portfolio", "Portfolio") }}</label>
            <select v-model="portfolioId" class="cf-input" :disabled="isLocked">
              <option value="" disabled>
                {{ t("operation.pickPortfolio", "Pick a portfolio") }}
              </option>
              <option v-for="p in portfolios" :key="p.id" :value="p.id">
                {{ p.code }} — {{ p.name }} ({{ p.valuation_currency }})
              </option>
            </select>
          </div>
        </div>

        <!-- Instrument picker ------------------------------------------- -->
        <div class="op-ticket__field">
          <label>{{ t("operation.instrument", "Instrument") }}</label>
          <input
            v-model="instrumentSearch"
            class="cf-input"
            :placeholder="t('operation.instrumentSearch', 'Search ticker or name…')"
            :disabled="isLocked"
          />
          <select v-model="instrumentId" class="cf-input op-ticket__instrument-select" size="6" :disabled="isLocked">
            <option value="" disabled>{{ t("operation.pickInstrument", "Pick an instrument") }}</option>
            <option v-for="i in filteredInstruments" :key="i.id" :value="i.id">
              {{ instrumentLabel(i) }}
            </option>
          </select>
        </div>

        <!-- Numbers ---------------------------------------------------- -->
        <div class="op-ticket__row">
          <div class="op-ticket__field">
            <label>{{ t("operation.quantity", "Quantity") }}</label>
            <input
              v-model="quantity"
              type="number"
              min="0"
              step="any"
              class="cf-input"
              placeholder="0"
              :disabled="isLocked"
            />
          </div>
          <div class="op-ticket__field">
            <label>{{ t("operation.price", "Price") }}</label>
            <input
              v-model="price"
              type="number"
              min="0"
              step="any"
              class="cf-input"
              placeholder="0.00"
              :disabled="isLocked"
            />
          </div>
          <div class="op-ticket__field">
            <label>{{ t("operation.fees", "Fees") }}</label>
            <input
              v-model="fees"
              type="number"
              min="0"
              step="any"
              class="cf-input"
              placeholder="0.00"
              :disabled="isLocked"
            />
          </div>
        </div>

        <div class="op-ticket__row">
          <div class="op-ticket__field">
            <label>{{ t("operation.businessDate", "Business date") }}</label>
            <input v-model="businessDate" type="date" class="cf-input" :disabled="isLocked" />
          </div>
          <div class="op-ticket__field">
            <label>{{ t("operation.settlementDate", "Settlement date") }}</label>
            <input v-model="settlementDate" type="date" class="cf-input" :disabled="isLocked" />
          </div>
          <div class="op-ticket__field">
            <label>{{ t("operation.currency", "Currency") }}</label>
            <input v-model="currency" class="cf-input" disabled />
          </div>
        </div>

        <div class="op-ticket__field">
          <label>{{ t("operation.reason", "Reason / Note") }}</label>
          <input
            v-model="reason"
            class="cf-input"
            :placeholder="t('operation.reasonHint', 'Free-text note attached to the audit trail')"
            :disabled="isLocked"
          />
        </div>

        <!-- Computed totals -------------------------------------------- -->
        <div class="op-ticket__totals" :class="`tone-${side === 'BUY' ? 'buy' : 'sell'}`">
          <div>
            <span class="op-ticket__totals-label">{{ t("operation.gross", "Gross") }}</span>
            <span class="op-ticket__totals-value">{{ fmtMoney(grossAmount) }} {{ currency }}</span>
          </div>
          <div>
            <span class="op-ticket__totals-label">
              {{ side === "BUY" ? t("operation.netDebit", "Net (debit)") : t("operation.netCredit", "Net (credit)") }}
            </span>
            <span class="op-ticket__totals-value">{{ fmtMoney(netAmount) }} {{ currency }}</span>
          </div>
          <div>
            <span class="op-ticket__totals-label">{{ t("operation.side", "Side") }}</span>
            <span class="op-ticket__totals-value">{{ side }}</span>
          </div>
        </div>

        <!-- Pre-trade results ------------------------------------------ -->
        <div v-if="pretradeError" class="op-ticket__alert" role="alert">{{ pretradeError }}</div>

        <div v-if="simulation" class="op-ticket__verdict" :class="`tone-${verdict?.toLowerCase() ?? 'neutral'}`">
          <header>
            <strong>{{ t("operation.verdict", "Pre-trade verdict") }}: {{ verdict ?? "—" }}</strong>
            <small>{{ simulation.compliance?.rules_evaluated ?? 0 }} {{ t("operation.rulesEvaluated", "rules evaluated") }}</small>
          </header>
          <ul v-if="simulation.compliance?.breaches?.length" class="op-ticket__breaches">
            <li
              v-for="b in simulation.compliance.breaches"
              :key="(b.breach_id ?? '') + (b.rule_type_id ?? '')"
              :class="`tone-${(b.severity ?? '').toLowerCase() || 'warn'}`"
            >
              <span class="op-ticket__breach-sev">{{ b.severity }}</span>
              <span class="op-ticket__breach-msg">
                <strong>{{ b.rule_type_id }}</strong>
                — {{ b.message ?? t("operation.noMessage", "No message") }}
              </span>
            </li>
          </ul>
          <p v-else class="op-ticket__breaches-empty">
            {{ t("operation.allGreen", "All pre-trade rules passed.") }}
          </p>
        </div>

        <div v-if="postError" class="op-ticket__alert" role="alert">{{ postError }}</div>
        <div v-if="posted" class="op-ticket__success" role="status">
          {{ t("operation.posted", "Transaction posted") }} — id {{ posted.id ?? "—" }}
        </div>

        <!-- Actions ---------------------------------------------------- -->
        <div class="op-ticket__actions">
          <button
            type="button"
            class="cf-btn cf-btn--secondary"
            :disabled="!canRunPretrade"
            @click="onRunPretrade"
          >
            {{
              simulating
                ? t("operation.pretrading", "Running pre-trade…")
                : t("operation.runPretrade", "Run pre-trade check")
            }}
          </button>
          <button
            type="button"
            class="cf-btn cf-btn--primary"
            :class="{ 'is-sell': side === 'SELL' }"
            :disabled="!canPost"
            @click="onPost"
          >
            {{
              posting
                ? t("operation.posting", "Posting…")
                : side === "BUY"
                  ? t("operation.postBuy", "Post BUY")
                  : t("operation.postSell", "Post SELL")
            }}
          </button>
        </div>
      </div>
    </AppCard>
  </section>
</template>

<style scoped>
.op-page {
  display: grid;
  gap: var(--space-3);
}

.op-ticket {
  display: grid;
  gap: 12px;
}

.op-ticket__row {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.op-ticket__field {
  display: grid;
  gap: 4px;
}

.op-ticket__field label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary);
  font-weight: 600;
}

.cf-input {
  height: 32px;
  padding: 0 10px;
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-primary);
}
.cf-input:focus {
  outline: none;
  border-color: var(--border-focus, var(--color-primary-500, #1f6feb));
  box-shadow: 0 0 0 3px rgba(31, 111, 235, 0.18);
}

.op-ticket__instrument-select {
  height: auto;
  padding: 6px;
}

.op-ticket__sides {
  display: inline-flex;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  padding: 3px;
  align-self: end;
}

.op-ticket__side {
  padding: 4px 16px;
  font-size: 12px;
  font-weight: 700;
  background: transparent;
  border: 0;
  border-radius: 4px;
  color: var(--text-secondary);
  cursor: pointer;
  letter-spacing: 0.04em;
}
.op-ticket__side.is-active.is-buy {
  background: var(--state-success, #1a7f37);
  color: #fff;
}
.op-ticket__side.is-active.is-sell {
  background: var(--state-danger, #cf222e);
  color: #fff;
}

.op-ticket__totals {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  padding: 12px;
  border-radius: 6px;
  background: var(--surface-1);
  border-left: 3px solid transparent;
}
.op-ticket__totals.tone-buy  { border-left-color: var(--state-success, #1a7f37); }
.op-ticket__totals.tone-sell { border-left-color: var(--state-danger, #cf222e); }

.op-ticket__totals-label {
  display: block;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary);
  font-weight: 600;
}
.op-ticket__totals-value {
  display: block;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
}

.op-ticket__alert {
  padding: 10px 12px;
  background: var(--alert-danger-bg);
  border: 1px solid var(--alert-danger-border);
  border-radius: 6px;
  color: var(--alert-danger-text);
  font-size: 12px;
}

.op-ticket__success {
  padding: 10px 12px;
  background: var(--alert-success-bg, rgba(26, 127, 55, 0.1));
  border: 1px solid var(--alert-success-border, rgba(26, 127, 55, 0.4));
  border-radius: 6px;
  color: var(--alert-success-text, #1a7f37);
  font-size: 12px;
  font-weight: 500;
}

.op-ticket__verdict {
  padding: 12px;
  border-radius: 6px;
  border: 1px solid var(--border-default);
}
.op-ticket__verdict header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 6px;
}
.op-ticket__verdict header strong { font-size: 13px; }
.op-ticket__verdict header small { font-size: 11px; color: var(--text-tertiary); }
.op-ticket__verdict.tone-pass  { background: rgba(26, 127, 55, 0.08); border-color: rgba(26, 127, 55, 0.35); }
.op-ticket__verdict.tone-warn  { background: rgba(214, 152, 0, 0.08); border-color: rgba(214, 152, 0, 0.35); }
.op-ticket__verdict.tone-block { background: rgba(207, 34, 46, 0.08); border-color: rgba(207, 34, 46, 0.35); }

.op-ticket__breaches {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 6px;
}

.op-ticket__breaches li {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 8px;
  align-items: start;
  font-size: 12px;
  color: var(--text-primary);
}

.op-ticket__breach-sev {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 10px;
  letter-spacing: 0.04em;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
}
.op-ticket__breaches li.tone-block .op-ticket__breach-sev {
  background: rgba(207, 34, 46, 0.15);
  border-color: rgba(207, 34, 46, 0.5);
  color: var(--state-danger, #cf222e);
}
.op-ticket__breaches li.tone-warn .op-ticket__breach-sev {
  background: rgba(214, 152, 0, 0.15);
  border-color: rgba(214, 152, 0, 0.5);
  color: var(--state-warning, #b06800);
}

.op-ticket__breaches-empty {
  margin: 0;
  font-size: 12px;
  color: var(--state-success, #1a7f37);
}

.op-ticket__actions {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  justify-content: flex-end;
  border-top: 1px solid var(--border-subtle);
  padding-top: 12px;
}

.cf-btn {
  display: inline-flex;
  align-items: center;
  height: 34px;
  padding: 0 18px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid var(--border-default);
}
.cf-btn--secondary {
  background: var(--action-secondary);
  color: var(--text-primary);
}
.cf-btn--secondary:hover { background: var(--action-secondary-hover, var(--surface-1)); }
.cf-btn--secondary:disabled { opacity: 0.6; cursor: not-allowed; }

.cf-btn--primary {
  background: var(--state-success, #1a7f37);
  color: #fff;
  border-color: transparent;
}
.cf-btn--primary:hover { background: #1f8e3f; }
.cf-btn--primary:disabled { opacity: 0.6; cursor: not-allowed; }
.cf-btn--primary.is-sell { background: var(--state-danger, #cf222e); }
.cf-btn--primary.is-sell:hover { background: #d6303d; }

@media (max-width: 720px) {
  .op-ticket__row,
  .op-ticket__totals {
    grid-template-columns: 1fr;
  }
}

.op-ticket__lock-banner {
  display: flex;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--alert-danger-bg, rgba(207, 34, 46, 0.08));
  border: 1px solid var(--alert-danger-border, rgba(207, 34, 46, 0.35));
  border-radius: var(--radius-lg, 6px);
  color: var(--alert-danger-text, #cf222e);
  align-items: center;
  font-size: var(--font-size-sm, 14px);
  margin-bottom: var(--space-2);
}

.op-ticket__lock-icon {
  flex-shrink: 0;
  color: var(--state-danger, #cf222e);
}

.op-ticket__lock-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
  text-align: left;
}

.op-ticket__permission-warning {
  display: flex;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  background: var(--alert-warning-bg, rgba(214, 152, 0, 0.08));
  border: 1px solid var(--alert-warning-border, rgba(214, 152, 0, 0.35));
  border-radius: var(--radius-md, 4px);
  color: var(--alert-warning-text, #b06800);
  align-items: center;
  font-size: var(--font-size-xs, 12px);
  margin-bottom: var(--space-2);
  text-align: left;
}
</style>
