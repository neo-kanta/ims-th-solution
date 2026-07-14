<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useState } from "#imports";

import AppCard from "~/shared/ui/AppCard.vue";
import { useI18n } from "~/composables/useI18n";
import PortfolioWorkspaceHeader from "./components/PortfolioWorkspaceHeader.vue";
import { usePortfolioContext } from "./composables/usePortfolioContext";
import {
  buildBindRulePayload,
  deriveBindingState,
  isSubmissionInFlight,
  type BindDraft,
} from "./lib/complianceBindingState";
import {
  portfolioComplianceApi,
  type ApiPortfolioRuleCatalogEntry,
} from "./services/portfolioComplianceApi";

const props = defineProps<{ portfolioCode: string }>();
const { t } = useI18n();
const ctx = usePortfolioContext(() => props.portfolioCode);

const pageTitle = useState<string>("page-title", () => "");
watch(
  () => t("portfolio.workspaceTabs.compliance", "Compliance"),
  (newTitle) => {
    pageTitle.value = newTitle || "";
  },
  { immediate: true }
);

const rules = ref<ApiPortfolioRuleCatalogEntry[]>([]);
const loadingRules = ref(false);
const rulesError = ref<string | null>(null);

// Per-rule bind-form state, keyed by rule_instance_id. Opening the form for
// one rule does not affect any other row.
const openBindForm = ref<Record<string, boolean>>({});

const bindDrafts = ref<Record<string, BindDraft>>({});
const bindingBusy = ref<Record<string, boolean>>({});
const bindError = ref<Record<string, string | null>>({});

function extractMessage(err: unknown, fallback: string): string {
  if (err instanceof Error && err.message) return err.message;
  return fallback;
}

async function loadRules() {
  if (!props.portfolioCode) return;
  loadingRules.value = true;
  rulesError.value = null;
  try {
    rules.value = await portfolioComplianceApi.listRules(props.portfolioCode);
  } catch (err) {
    rulesError.value = extractMessage(err, "Failed to load compliance rules.");
  } finally {
    loadingRules.value = false;
  }
}

onMounted(() => {
  void ctx.reload();
  void loadRules();
});

watch(
  () => props.portfolioCode,
  () => {
    void ctx.reload();
    void loadRules();
  }
);

function formatDate(value?: string | null): string {
  if (!value) return "—";
  return value.slice(0, 10);
}

// Safe boundary normalizer: PortfolioRuleCatalogEntry.parameters is typed
// Record<string, never> by the generated client (backend/pkg/contract/
// contracts.go's `swaggertype:"object"` gives swaggo no shape info — see
// docs/MANAGER/TASKS.md P2 follow-up). We accept `unknown` here and validate
// a plain, non-array object at the boundary instead of casting the drifted
// generated type further.
function asParameterRecord(value: unknown): Record<string, unknown> {
  if (value && typeof value === "object" && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  return {};
}

function formatParameters(entry: ApiPortfolioRuleCatalogEntry): string {
  const params = asParameterRecord(entry.parameters);
  if (Object.keys(params).length === 0) return "—";
  return Object.entries(params)
    .map(([key, value]) => `${key}: ${String(value)}`)
    .join(", ");
}

// The workspace exposes no authoritative backend business date, so "today"
// is the browser's local date, always shown next to an explicit "as of"
// label rather than implied as the backend's business date.
const todayIso = computed(() => new Date().toISOString().slice(0, 10));

function bindingStateLabel(state: ReturnType<typeof deriveBindingState>): string {
  switch (state) {
    case "SCHEDULED":
      return t("portfolio.compliance.state.scheduled", "Scheduled");
    case "EFFECTIVE":
      return t("portfolio.compliance.state.effective", "Currently effective");
    case "EXPIRED":
      return t("portfolio.compliance.state.expired", "Expired");
    case "DEACTIVATED":
      return t("portfolio.compliance.state.deactivated", "Deactivated");
    default:
      return "—";
  }
}

function toggleBindForm(entry: ApiPortfolioRuleCatalogEntry) {
  const id = entry.rule_instance_id ?? "";
  if (!id) return;
  const next = !openBindForm.value[id];
  openBindForm.value = { ...openBindForm.value, [id]: next };
  if (next && !bindDrafts.value[id]) {
    bindDrafts.value = {
      ...bindDrafts.value,
      [id]: {
        severity: "BLOCK",
        effectiveFrom: new Date().toISOString().slice(0, 10),
        effectiveTo: "",
      },
    };
  }
}

function updateBindDraft(
  entry: ApiPortfolioRuleCatalogEntry,
  field: keyof BindDraft,
  event: Event,
) {
  const id = entry.rule_instance_id ?? "";
  const draft = bindDrafts.value[id];
  const target = event.target as HTMLInputElement | HTMLSelectElement | null;
  if (!id || !draft || !target) return;
  bindDrafts.value = {
    ...bindDrafts.value,
    [id]: { ...draft, [field]: target.value },
  };
}

async function submitBind(entry: ApiPortfolioRuleCatalogEntry) {
  const id = entry.rule_instance_id ?? "";
  if (!id) return;
  const draft = bindDrafts.value[id];
  if (!draft) return;
  if (isSubmissionInFlight(bindingBusy.value, id)) return;

  bindingBusy.value = { ...bindingBusy.value, [id]: true };
  bindError.value = { ...bindError.value, [id]: null };
  try {
    await portfolioComplianceApi.bindRule(
      props.portfolioCode,
      id,
      buildBindRulePayload(draft),
    );
    openBindForm.value = { ...openBindForm.value, [id]: false };
    await loadRules();
  } catch (err) {
    bindError.value = {
      ...bindError.value,
      [id]: extractMessage(err, "Failed to bind rule to this portfolio."),
    };
  } finally {
    bindingBusy.value = { ...bindingBusy.value, [id]: false };
  }
}

async function deactivate(entry: ApiPortfolioRuleCatalogEntry) {
  const id = entry.rule_instance_id ?? "";
  const bindingId = entry.binding?.binding_id;
  if (!id || !bindingId) return;
  if (isSubmissionInFlight(bindingBusy.value, id)) return;
  bindingBusy.value = { ...bindingBusy.value, [id]: true };
  try {
    await portfolioComplianceApi.deactivateBinding(props.portfolioCode, id, bindingId);
    await loadRules();
  } catch (err) {
    bindError.value = {
      ...bindError.value,
      [id]: extractMessage(err, "Failed to deactivate this binding."),
    };
  } finally {
    bindingBusy.value = { ...bindingBusy.value, [id]: false };
  }
}

const severityOptions = ["BLOCK", "WARN", "REQUIRE_APPROVAL", "MONITOR"];

const boundRules = computed(() => rules.value.filter((r) => r.binding?.is_active));
const unboundRules = computed(() => rules.value.filter((r) => !r.binding?.is_active));
</script>

<template>
  <section class="portfolio-compliance">
    <div v-if="ctx.loading.value && !ctx.portfolio.value" class="portfolio-compliance__notice" role="status">
      {{ t("portfolio.overview.loading") }}
    </div>

    <template v-else-if="ctx.portfolio.value">
      <PortfolioWorkspaceHeader :portfolio="ctx.portfolio.value" />

      <AppCard
        :title="t('portfolio.compliance.boundTitle', 'Bound Compliance Rules')"
        :subtitle="t('portfolio.compliance.boundSubtitle', 'Rules bound to this portfolio. Effective status is evaluated against the as-of date below, not just the active flag.')"
      >
        <div v-if="loadingRules" class="portfolio-compliance__card-notice">
          {{ t("portfolio.compliance.loading", "Loading rules…") }}
        </div>
        <div v-else-if="rulesError" class="portfolio-compliance__error" role="alert">{{ rulesError }}</div>
        <div v-else-if="boundRules.length === 0" class="portfolio-compliance__card-notice">
          {{ t("portfolio.compliance.noneBound", "No compliance rules are bound to this portfolio yet.") }}
        </div>
        <template v-else>
          <p class="portfolio-compliance__as-of">
            {{ t("portfolio.compliance.asOfLabel", { date: todayIso }, "As of {date} (your device's local date)") }}
          </p>
          <table class="portfolio-compliance__table">
          <thead>
            <tr>
              <th>{{ t("portfolio.compliance.columns.rule", "Rule") }}</th>
              <th>{{ t("portfolio.compliance.columns.parameters", "Parameters") }}</th>
              <th>{{ t("portfolio.compliance.columns.severity", "Severity") }}</th>
              <th>{{ t("portfolio.compliance.columns.effective", "Effective") }}</th>
              <th>{{ t("portfolio.compliance.columns.status", "Status") }}</th>
              <th class="text-right">{{ t("portfolio.compliance.columns.actions", "Actions") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="entry in boundRules" :key="entry.rule_instance_id">
              <td>
                <div class="portfolio-compliance__rule-name">{{ entry.name }}</div>
                <div class="portfolio-compliance__rule-type">{{ entry.rule_type_id }}</div>
              </td>
              <td class="portfolio-compliance__params">{{ formatParameters(entry) }}</td>
              <td>
                <span class="portfolio-compliance__severity-pill" :data-severity="entry.binding?.severity">
                  {{ entry.binding?.severity }}
                </span>
              </td>
              <td class="portfolio-compliance__mono">
                {{ formatDate(entry.binding?.effective_from) }} –
                {{ entry.binding?.effective_to ? formatDate(entry.binding.effective_to) : t("portfolio.compliance.indefinite", "indefinite") }}
              </td>
              <td>
                <span
                  class="portfolio-compliance__state-pill"
                  :data-state="deriveBindingState(entry.binding, todayIso)"
                >
                  {{ bindingStateLabel(deriveBindingState(entry.binding, todayIso)) }}
                </span>
              </td>
              <td class="text-right">
                <button
                  type="button"
                  class="portfolio-compliance__btn portfolio-compliance__btn--danger"
                  :disabled="bindingBusy[entry.rule_instance_id ?? '']"
                  @click="deactivate(entry)"
                >
                  {{ t("portfolio.compliance.deactivate", "Deactivate") }}
                </button>
              </td>
            </tr>
          </tbody>
          </table>
        </template>
      </AppCard>

      <AppCard
        :title="t('portfolio.compliance.availableTitle', 'Available Compliance Rules')"
        :subtitle="t('portfolio.compliance.availableSubtitle', 'Bind an existing rule instance to this portfolio.')"
      >
        <div v-if="!loadingRules && unboundRules.length === 0" class="portfolio-compliance__card-notice">
          {{ t("portfolio.compliance.noneAvailable", "Every active rule instance is already bound to this portfolio.") }}
        </div>
        <ul v-else class="portfolio-compliance__available-list">
          <li v-for="entry in unboundRules" :key="entry.rule_instance_id" class="portfolio-compliance__available-item">
            <div class="portfolio-compliance__available-row">
              <div>
                <div class="portfolio-compliance__rule-name">{{ entry.name }}</div>
                <div class="portfolio-compliance__rule-type">{{ entry.rule_type_id }}</div>
                <div class="portfolio-compliance__params">{{ formatParameters(entry) }}</div>
              </div>
              <button
                type="button"
                class="portfolio-compliance__btn"
                @click="toggleBindForm(entry)"
              >
                {{ openBindForm[entry.rule_instance_id ?? ''] ? t("portfolio.compliance.cancel", "Cancel") : t("portfolio.compliance.bind", "Bind") }}
              </button>
            </div>

            <div v-if="openBindForm[entry.rule_instance_id ?? '']" class="portfolio-compliance__bind-form">
              <label class="portfolio-compliance__field">
                <span>{{ t("portfolio.compliance.columns.severity", "Severity") }}</span>
                <select
                  :value="bindDrafts[entry.rule_instance_id ?? '']?.severity ?? 'BLOCK'"
                  @change="updateBindDraft(entry, 'severity', $event)"
                >
                  <option v-for="s in severityOptions" :key="s" :value="s">{{ s }}</option>
                </select>
              </label>
              <label class="portfolio-compliance__field">
                <span>{{ t("portfolio.compliance.effectiveFrom", "Effective from") }}</span>
                <input
                  type="date"
                  :value="bindDrafts[entry.rule_instance_id ?? '']?.effectiveFrom ?? ''"
                  @change="updateBindDraft(entry, 'effectiveFrom', $event)"
                />
              </label>
              <label class="portfolio-compliance__field">
                <span>{{ t("portfolio.compliance.effectiveTo", "Effective to (optional)") }}</span>
                <input
                  type="date"
                  :value="bindDrafts[entry.rule_instance_id ?? '']?.effectiveTo ?? ''"
                  @change="updateBindDraft(entry, 'effectiveTo', $event)"
                />
              </label>
              <button
                type="button"
                class="portfolio-compliance__btn portfolio-compliance__btn--primary"
                :disabled="bindingBusy[entry.rule_instance_id ?? '']"
                @click="submitBind(entry)"
              >
                {{ t("portfolio.compliance.confirmBind", "Confirm bind") }}
              </button>
              <div v-if="bindError[entry.rule_instance_id ?? '']" class="portfolio-compliance__error">
                {{ bindError[entry.rule_instance_id ?? ''] }}
              </div>
            </div>
          </li>
        </ul>
      </AppCard>
    </template>

    <div v-else class="portfolio-compliance__notice">
      {{ t("portfolio.overview.notFound") }}
    </div>
  </section>
</template>

<style scoped>
.portfolio-compliance {
  display: grid;
  gap: var(--space-4, 16px);
}

.portfolio-compliance__notice,
.portfolio-compliance__card-notice {
  padding: var(--space-4, 16px);
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
  text-align: center;
}

.portfolio-compliance__error {
  padding: 10px 12px;
  background: var(--alert-danger-bg);
  border: 1px solid var(--alert-danger-border);
  border-radius: 6px;
  color: var(--alert-danger-text, #cf222e);
  font-size: 12px;
}

.portfolio-compliance__as-of {
  margin: 0 0 10px;
  font-size: 11px;
  color: var(--text-tertiary, #6e7781);
}

.portfolio-compliance__state-pill {
  display: inline-block;
  font-size: 10px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--bg-card-muted, #f1f3f4);
  border: 1px solid var(--border-subtle, #d0d7de);
  color: var(--text-secondary, #5f6368);
}

.portfolio-compliance__state-pill[data-state="EFFECTIVE"] {
  background: #e6f4ea;
  color: #1e7e34;
  border-color: #b7dfc0;
}

.portfolio-compliance__state-pill[data-state="SCHEDULED"] {
  background: #e8f0fe;
  color: #1a56db;
  border-color: #c3d7fb;
}

.portfolio-compliance__state-pill[data-state="EXPIRED"],
.portfolio-compliance__state-pill[data-state="DEACTIVATED"] {
  background: #fef7e0;
  color: #b25e00;
  border-color: #fce1a6;
}

.portfolio-compliance__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.portfolio-compliance__table th {
  text-align: left;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-secondary, #6e7781);
  padding: 10px 12px;
  border-bottom: 2px solid var(--border-subtle, #d0d7de);
}

.portfolio-compliance__table td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
  vertical-align: top;
}

.portfolio-compliance__rule-name {
  font-weight: 600;
  color: var(--text-primary, #1f2328);
}

.portfolio-compliance__rule-type {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 11px;
  color: var(--text-secondary, #6e7781);
}

.portfolio-compliance__params {
  font-size: 12px;
  color: var(--text-secondary, #57606a);
  max-width: 320px;
}

.portfolio-compliance__mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}

.portfolio-compliance__severity-pill {
  display: inline-block;
  font-size: 10px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--bg-card-muted, #f1f3f4);
  border: 1px solid var(--border-subtle, #d0d7de);
  color: var(--text-secondary, #5f6368);
}

.portfolio-compliance__severity-pill[data-severity="BLOCK"] {
  background: #fce8e6;
  color: #c5221f;
  border-color: #f5b4ad;
}

.portfolio-compliance__severity-pill[data-severity="WARN"] {
  background: #fef7e0;
  color: #b25e00;
  border-color: #fce1a6;
}

.portfolio-compliance__btn {
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  border: 1px solid var(--border-subtle, #d0d7de);
  background: var(--bg-card, #ffffff);
  border-radius: var(--radius-md, 5px);
  padding: 5px 12px;
  cursor: pointer;
}

.portfolio-compliance__btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.portfolio-compliance__btn--primary {
  background: var(--state-primary, #1a73e8);
  color: #fff;
  border-color: transparent;
}

.portfolio-compliance__btn--danger {
  color: var(--state-danger, #cf222e);
  border-color: var(--alert-danger-border, #f5b4ad);
}

.portfolio-compliance__available-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 12px;
}

.portfolio-compliance__available-item {
  border: 1px solid var(--border-subtle, #e1e8ed);
  border-radius: 8px;
  padding: 12px;
}

.portfolio-compliance__available-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.portfolio-compliance__bind-form {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  gap: 12px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-subtle, #e1e8ed);
}

.portfolio-compliance__field {
  display: grid;
  gap: 4px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary, #6e7781);
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.portfolio-compliance__field select,
.portfolio-compliance__field input {
  font-family: inherit;
  font-size: 13px;
  font-weight: 400;
  text-transform: none;
  padding: 6px 8px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: 6px;
  background: var(--bg-card, #ffffff);
  color: var(--text-primary);
}

.text-right {
  text-align: right !important;
}
</style>
