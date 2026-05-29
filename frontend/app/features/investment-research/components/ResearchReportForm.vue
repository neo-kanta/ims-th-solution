<script setup lang="ts">
import { computed, reactive, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import {
  RECOMMENDATIONS,
  REPORT_STATUSES,
  type CreateResearchReportInput,
  type ResearchRecommendation,
  type ResearchReport,
  type ResearchReportStatus,
  type UpdateResearchReportInput,
} from "../types";
import {
  RESEARCH_FIELD_LIMITS,
  MIN_INVESTMENT_ANALYSIS_LENGTH,
  isCurrencyCode,
  isUuid,
  recommendationLabel,
  reportStatusLabel,
} from "../lib/researchReportFormat";

/**
 * ResearchReportForm
 * ------------------------------------------------------------------
 * Single-component form for both create and edit flows. Validation is
 * intentionally a duplicate of the backend's authoritative checks
 * (policy + DTO length limits) so the operator gets immediate feedback
 * instead of a 500 from a Postgres CHECK violation. Server-side remains
 * the source of truth — every backend rule is re-checked in
 * `policy.ValidateResearchReport` and the command handler.
 *
 * Identity defaulting:
 *   - On create, owner_user_id and author_user_id pre-fill from the
 *     currently authenticated user. The fields remain editable so an
 *     admin can override; the backend will still enforce identity
 *     trust separately (review fix #5).
 */

const props = defineProps<{
  mode: "create" | "edit";
  initial?: ResearchReport | null;
  /** Pre-filled author/owner for create mode. Comes from useAuthStore(). */
  defaultUserId?: string | null;
  saving: boolean;
  error: string | null;
}>();

const emit = defineEmits<{
  submit: [
    payload:
      | { mode: "create"; input: CreateResearchReportInput }
      | { mode: "edit"; input: UpdateResearchReportInput },
  ];
  cancel: [];
}>();

const { t } = useI18n();

interface FormState {
  report_no: string;
  report_date: string;
  effective_date: string;
  owner_user_id: string;
  author_user_id: string;
  applicable_contract_id: string;
  instrument_type: string;
  instrument_code: string;
  instrument_name: string;
  market: string;
  currency: string;
  recommendation: ResearchRecommendation | "";
  report_title: string;
  company_overview: string;
  company_outlook: string;
  esg_comment: string;
  financial_status: string;
  investment_analysis: string;
  rejection_reason: string;
  post_submission_note: string;
  report_status: ResearchReportStatus | "";
}

function emptyForm(): FormState {
  return {
    report_no: "",
    report_date: "",
    effective_date: "",
    owner_user_id: "",
    author_user_id: "",
    applicable_contract_id: "",
    instrument_type: "",
    instrument_code: "",
    instrument_name: "",
    market: "",
    currency: "",
    recommendation: "",
    report_title: "",
    company_overview: "",
    company_outlook: "",
    esg_comment: "",
    financial_status: "",
    investment_analysis: "",
    rejection_reason: "",
    post_submission_note: "",
    report_status: "",
  };
}

const form = reactive<FormState>(emptyForm());
const fieldErrors = reactive<Partial<Record<keyof FormState, string>>>({});

function hydrateFromReport(report: ResearchReport) {
  form.report_no = report.report_no;
  form.report_date = report.report_date;
  form.effective_date = report.effective_date ?? "";
  form.owner_user_id = report.owner_user_id;
  form.author_user_id = report.author_user_id;
  form.applicable_contract_id = report.applicable_contract_id ?? "";
  form.instrument_type = report.instrument_type;
  form.instrument_code = report.instrument_code;
  form.instrument_name = report.instrument_name;
  form.market = report.market;
  form.currency = report.currency;
  form.recommendation = report.recommendation;
  form.report_title = report.report_title;
  form.company_overview = report.company_overview;
  form.company_outlook = report.company_outlook;
  form.esg_comment = report.esg_comment;
  form.financial_status = report.financial_status;
  form.investment_analysis = report.investment_analysis;
  form.rejection_reason = report.rejection_reason;
  form.post_submission_note = report.post_submission_note;
  form.report_status = report.report_status;
}

// Identity defaults: in create mode, populate owner/author from the
// authenticated user. We don't overwrite values the user already typed,
// and we never apply this in edit mode (where we hydrate from `initial`).
watch(
  () => [props.mode, props.defaultUserId] as const,
  ([mode, uid]) => {
    if (mode !== "create" || !uid) return;
    if (!form.owner_user_id.trim()) form.owner_user_id = uid;
    if (!form.author_user_id.trim()) form.author_user_id = uid;
  },
  { immediate: true },
);

watch(
  () => props.initial,
  (next) => {
    if (next) {
      hydrateFromReport(next);
    }
  },
  { immediate: true },
);

const analysisLength = computed(
  () => form.investment_analysis.trim().length,
);

const reportNoLength = computed(() => form.report_no.length);
const instrumentCodeLength = computed(() => form.instrument_code.length);

function clearErrors() {
  for (const key of Object.keys(fieldErrors) as Array<keyof FormState>) {
    delete fieldErrors[key];
  }
}

function validate(): boolean {
  clearErrors();
  let ok = true;

  if (props.mode === "create") {
    const reportNo = form.report_no.trim();
    if (!reportNo) {
      fieldErrors.report_no = t("investmentResearch.required");
      ok = false;
    } else if (reportNo.length > RESEARCH_FIELD_LIMITS.reportNo) {
      fieldErrors.report_no = t("investmentResearch.fieldLimits.maxLength", { max: String(RESEARCH_FIELD_LIMITS.reportNo) });
      ok = false;
    }
  }

  if (!form.report_date) {
    fieldErrors.report_date = t("investmentResearch.required");
    ok = false;
  }

  const owner = form.owner_user_id.trim();
  if (!owner) {
    fieldErrors.owner_user_id = t("investmentResearch.required");
    ok = false;
  } else if (!isUuid(owner)) {
    fieldErrors.owner_user_id = t("investmentResearch.fieldLimits.uuid");
    ok = false;
  }

  const author = form.author_user_id.trim();
  if (!author) {
    fieldErrors.author_user_id = t("investmentResearch.required");
    ok = false;
  } else if (!isUuid(author)) {
    fieldErrors.author_user_id = t("investmentResearch.fieldLimits.uuid");
    ok = false;
  }

  const contractId = form.applicable_contract_id.trim();
  if (contractId && !isUuid(contractId)) {
    fieldErrors.applicable_contract_id = t("investmentResearch.fieldLimits.uuidOptional");
    ok = false;
  }

  const instrumentCode = form.instrument_code.trim();
  if (!instrumentCode) {
    fieldErrors.instrument_code = t("investmentResearch.required");
    ok = false;
  } else if (instrumentCode.length > RESEARCH_FIELD_LIMITS.instrumentCode) {
    fieldErrors.instrument_code = t("investmentResearch.fieldLimits.maxLength", { max: String(RESEARCH_FIELD_LIMITS.instrumentCode) });
    ok = false;
  }

  if (form.instrument_name.length > RESEARCH_FIELD_LIMITS.instrumentName) {
    fieldErrors.instrument_name = t("investmentResearch.fieldLimits.maxLength", { max: String(RESEARCH_FIELD_LIMITS.instrumentName) });
    ok = false;
  }
  if (form.instrument_type.length > RESEARCH_FIELD_LIMITS.instrumentType) {
    fieldErrors.instrument_type = t("investmentResearch.fieldLimits.maxLength", { max: String(RESEARCH_FIELD_LIMITS.instrumentType) });
    ok = false;
  }
  if (form.market.length > RESEARCH_FIELD_LIMITS.market) {
    fieldErrors.market = t("investmentResearch.fieldLimits.maxLength", { max: String(RESEARCH_FIELD_LIMITS.market) });
    ok = false;
  }
  if (form.report_title.length > RESEARCH_FIELD_LIMITS.reportTitle) {
    fieldErrors.report_title = t("investmentResearch.fieldLimits.maxLength", { max: String(RESEARCH_FIELD_LIMITS.reportTitle) });
    ok = false;
  }

  if (!isCurrencyCode(form.currency)) {
    fieldErrors.currency = t("investmentResearch.fieldLimits.currencyCode");
    ok = false;
  }

  if (!form.recommendation) {
    fieldErrors.recommendation = t("investmentResearch.required");
    ok = false;
  }

  if (analysisLength.value < MIN_INVESTMENT_ANALYSIS_LENGTH) {
    fieldErrors.investment_analysis = t("investmentResearch.fieldLimits.minLength", { min: String(MIN_INVESTMENT_ANALYSIS_LENGTH) });
    ok = false;
  }

  return ok;
}

/**
 * Focus the first field that failed validation. Cheap accessibility win:
 * keyboard-only users land directly on the broken control instead of
 * scrolling to find it.
 */
function focusFirstError() {
  if (typeof document === "undefined") return;
  const firstErroredKey = Object.keys(fieldErrors)[0] as
    | keyof FormState
    | undefined;
  if (!firstErroredKey) return;
  const el = document.querySelector<HTMLElement>(
    `[data-field="${firstErroredKey}"]`,
  );
  el?.focus();
}

function handleSubmit() {
  if (!validate()) {
    void Promise.resolve().then(focusFirstError);
    return;
  }

  if (props.mode === "create") {
    const input: CreateResearchReportInput = {
      report_no: form.report_no.trim(),
      report_date: form.report_date,
      effective_date: form.effective_date || undefined,
      owner_user_id: form.owner_user_id.trim(),
      author_user_id: form.author_user_id.trim(),
      applicable_contract_id: form.applicable_contract_id.trim() || undefined,
      instrument_type: form.instrument_type.trim(),
      instrument_code: form.instrument_code.trim().toUpperCase(),
      instrument_name: form.instrument_name.trim(),
      market: form.market.trim(),
      currency: form.currency.trim().toUpperCase(),
      recommendation: form.recommendation as ResearchRecommendation,
      report_title: form.report_title.trim(),
      company_overview: form.company_overview,
      company_outlook: form.company_outlook,
      esg_comment: form.esg_comment,
      financial_status: form.financial_status,
      investment_analysis: form.investment_analysis,
    };
    emit("submit", { mode: "create", input });
    return;
  }

  const input: UpdateResearchReportInput = {
    report_date: form.report_date,
    effective_date: form.effective_date || undefined,
    owner_user_id: form.owner_user_id.trim() || undefined,
    author_user_id: form.author_user_id.trim() || undefined,
    applicable_contract_id: form.applicable_contract_id.trim() || undefined,
    instrument_type: form.instrument_type,
    instrument_code: form.instrument_code.trim().toUpperCase(),
    instrument_name: form.instrument_name,
    market: form.market,
    currency: form.currency.trim().toUpperCase(),
    recommendation: (form.recommendation || undefined) as
      | ResearchRecommendation
      | undefined,
    report_title: form.report_title,
    company_overview: form.company_overview,
    company_outlook: form.company_outlook,
    esg_comment: form.esg_comment,
    financial_status: form.financial_status,
    investment_analysis: form.investment_analysis,
    rejection_reason: form.rejection_reason,
    post_submission_note: form.post_submission_note,
    report_status: (form.report_status || undefined) as
      | ResearchReportStatus
      | undefined,
  };
  emit("submit", { mode: "edit", input });
}

/**
 * Cmd/Ctrl+Enter submits the form from anywhere within it. This matches
 * common desktop financial-tool UX (Bloomberg, Reuters, etc.) where
 * analysts work primarily from the keyboard.
 */
function handleFormKeydown(event: KeyboardEvent) {
  if ((event.metaKey || event.ctrlKey) && event.key === "Enter") {
    event.preventDefault();
    handleSubmit();
  }
}
</script>

<template>
  <form
    class="research-form"
    novalidate
    @submit.prevent="handleSubmit"
    @keydown="handleFormKeydown"
  >
    <fieldset class="research-form__section">
      <legend class="research-form__legend">{{ t('investmentResearch.identification') }}</legend>
      <div class="research-form__row">
        <label class="research-form__field">
          <span class="research-form__label">
            {{ t('investmentResearch.reportNo') }}
            <span class="research-form__required" aria-hidden="true">*</span>
          </span>
          <input
            v-model="form.report_no"
            data-field="report_no"
            class="input"
            type="text"
            :disabled="mode === 'edit'"
            :maxlength="RESEARCH_FIELD_LIMITS.reportNo"
            :aria-invalid="Boolean(fieldErrors.report_no)"
            :aria-describedby="fieldErrors.report_no ? 'err-report-no' : 'hint-report-no'"
            placeholder="RR-2026-0001"
            autocomplete="off"
          />
          <small id="hint-report-no" class="research-form__hint">
            {{ t('investmentResearch.characterCounter', { count: String(reportNoLength), max: String(RESEARCH_FIELD_LIMITS.reportNo) }) }}
          </small>
          <small
            v-if="fieldErrors.report_no"
            id="err-report-no"
            class="research-form__error"
          >
            {{ fieldErrors.report_no }}
          </small>
        </label>
        <label class="research-form__field">
          <span class="research-form__label">
            {{ t('investmentResearch.reportDate') }}
            <span class="research-form__required" aria-hidden="true">*</span>
          </span>
          <input
            v-model="form.report_date"
            data-field="report_date"
            class="input"
            type="date"
            :aria-invalid="Boolean(fieldErrors.report_date)"
            :aria-describedby="fieldErrors.report_date ? 'err-report-date' : undefined"
          />
          <small
            v-if="fieldErrors.report_date"
            id="err-report-date"
            class="research-form__error"
          >
            {{ fieldErrors.report_date }}
          </small>
        </label>
        <label class="research-form__field">
          <span class="research-form__label">{{ t('investmentResearch.effectiveDate') }}</span>
          <input v-model="form.effective_date" class="input" type="date" />
        </label>
      </div>
      <div class="research-form__row">
        <label class="research-form__field">
          <span class="research-form__label">
            {{ t('investmentResearch.ownerUserId') }}
            <span class="research-form__required" aria-hidden="true">*</span>
          </span>
          <input
            v-model="form.owner_user_id"
            data-field="owner_user_id"
            class="input"
            type="text"
            placeholder="UUID"
            autocomplete="off"
            spellcheck="false"
            :aria-invalid="Boolean(fieldErrors.owner_user_id)"
            :aria-describedby="fieldErrors.owner_user_id ? 'err-owner' : undefined"
          />
          <small v-if="fieldErrors.owner_user_id" id="err-owner" class="research-form__error">
            {{ fieldErrors.owner_user_id }}
          </small>
        </label>
        <label class="research-form__field">
          <span class="research-form__label">
            {{ t('investmentResearch.authorUserId') }}
            <span class="research-form__required" aria-hidden="true">*</span>
          </span>
          <input
            v-model="form.author_user_id"
            data-field="author_user_id"
            class="input"
            type="text"
            placeholder="UUID"
            autocomplete="off"
            spellcheck="false"
            :aria-invalid="Boolean(fieldErrors.author_user_id)"
            :aria-describedby="fieldErrors.author_user_id ? 'err-author' : undefined"
          />
          <small v-if="fieldErrors.author_user_id" id="err-author" class="research-form__error">
            {{ fieldErrors.author_user_id }}
          </small>
        </label>
        <label class="research-form__field">
          <span class="research-form__label">{{ t('investmentResearch.applicableContractId') }}</span>
          <input
            v-model="form.applicable_contract_id"
            data-field="applicable_contract_id"
            class="input"
            type="text"
            :placeholder="t('investmentResearch.uuidOptionalPlaceholder')"
            autocomplete="off"
            spellcheck="false"
            :aria-invalid="Boolean(fieldErrors.applicable_contract_id)"
            :aria-describedby="fieldErrors.applicable_contract_id ? 'err-contract' : undefined"
          />
          <small
            v-if="fieldErrors.applicable_contract_id"
            id="err-contract"
            class="research-form__error"
          >
            {{ fieldErrors.applicable_contract_id }}
          </small>
        </label>
      </div>
    </fieldset>

    <fieldset class="research-form__section">
      <legend class="research-form__legend">{{ t('investmentResearch.instrument') }}</legend>
      <div class="research-form__row">
        <label class="research-form__field">
          <span class="research-form__label">
            {{ t('investmentResearch.instrumentCode') }}
            <span class="research-form__required" aria-hidden="true">*</span>
          </span>
          <input
            v-model="form.instrument_code"
            data-field="instrument_code"
            class="input"
            type="text"
            :maxlength="RESEARCH_FIELD_LIMITS.instrumentCode"
            placeholder="PTT"
            autocomplete="off"
            spellcheck="false"
            :aria-invalid="Boolean(fieldErrors.instrument_code)"
            :aria-describedby="
              fieldErrors.instrument_code
                ? 'err-instrument-code'
                : 'hint-instrument-code'
            "
          />
          <small id="hint-instrument-code" class="research-form__hint">
            {{ t('investmentResearch.characterCounter', { count: String(instrumentCodeLength), max: String(RESEARCH_FIELD_LIMITS.instrumentCode) }) }}
          </small>
          <small
            v-if="fieldErrors.instrument_code"
            id="err-instrument-code"
            class="research-form__error"
          >
            {{ fieldErrors.instrument_code }}
          </small>
        </label>
        <label class="research-form__field">
          <span class="research-form__label">{{ t('investmentResearch.instrumentName') }}</span>
          <input
            v-model="form.instrument_name"
            data-field="instrument_name"
            class="input"
            type="text"
            :maxlength="RESEARCH_FIELD_LIMITS.instrumentName"
            :aria-invalid="Boolean(fieldErrors.instrument_name)"
            :aria-describedby="fieldErrors.instrument_name ? 'err-instrument-name' : undefined"
          />
          <small
            v-if="fieldErrors.instrument_name"
            id="err-instrument-name"
            class="research-form__error"
          >
            {{ fieldErrors.instrument_name }}
          </small>
        </label>
        <label class="research-form__field">
          <span class="research-form__label">{{ t('investmentResearch.instrumentType') }}</span>
          <input
            v-model="form.instrument_type"
            data-field="instrument_type"
            class="input"
            type="text"
            :maxlength="RESEARCH_FIELD_LIMITS.instrumentType"
            :aria-invalid="Boolean(fieldErrors.instrument_type)"
            :aria-describedby="fieldErrors.instrument_type ? 'err-instrument-type' : undefined"
          />
          <small
            v-if="fieldErrors.instrument_type"
            id="err-instrument-type"
            class="research-form__error"
          >
            {{ fieldErrors.instrument_type }}
          </small>
        </label>
      </div>
      <div class="research-form__row">
        <label class="research-form__field">
          <span class="research-form__label">{{ t('investmentResearch.market') }}</span>
          <input
            v-model="form.market"
            data-field="market"
            class="input"
            type="text"
            :maxlength="RESEARCH_FIELD_LIMITS.market"
            placeholder="SET"
            :aria-invalid="Boolean(fieldErrors.market)"
            :aria-describedby="fieldErrors.market ? 'err-market' : undefined"
          />
          <small v-if="fieldErrors.market" id="err-market" class="research-form__error">
            {{ fieldErrors.market }}
          </small>
        </label>
        <label class="research-form__field">
          <span class="research-form__label">{{ t('investmentResearch.currency') }}</span>
          <input
            v-model="form.currency"
            data-field="currency"
            class="input"
            type="text"
            :maxlength="RESEARCH_FIELD_LIMITS.currency"
            placeholder="THB"
            autocomplete="off"
            spellcheck="false"
            :aria-invalid="Boolean(fieldErrors.currency)"
            :aria-describedby="fieldErrors.currency ? 'err-currency' : undefined"
          />
          <small v-if="fieldErrors.currency" id="err-currency" class="research-form__error">
            {{ fieldErrors.currency }}
          </small>
        </label>
        <label class="research-form__field">
          <span class="research-form__label">
            {{ t('investmentResearch.recommendation') }}
            <span class="research-form__required" aria-hidden="true">*</span>
          </span>
          <select
            v-model="form.recommendation"
            data-field="recommendation"
            class="select"
            :aria-invalid="Boolean(fieldErrors.recommendation)"
            :aria-describedby="
              fieldErrors.recommendation ? 'err-recommendation' : undefined
            "
          >
            <option value="">{{ t('investmentResearch.selectPlaceholder') }}</option>
            <option v-for="r in RECOMMENDATIONS" :key="r" :value="r">
              {{ recommendationLabel(r, t) }}
            </option>
          </select>
          <small
            v-if="fieldErrors.recommendation"
            id="err-recommendation"
            class="research-form__error"
          >
            {{ fieldErrors.recommendation }}
          </small>
        </label>
      </div>
    </fieldset>

    <fieldset class="research-form__section">
      <legend class="research-form__legend">{{ t('investmentResearch.reportContent') }}</legend>
      <label class="research-form__field">
        <span class="research-form__label">{{ t('investmentResearch.titleField') }}</span>
        <input
          v-model="form.report_title"
          data-field="report_title"
          class="input"
          type="text"
          :maxlength="RESEARCH_FIELD_LIMITS.reportTitle"
          :aria-invalid="Boolean(fieldErrors.report_title)"
          :aria-describedby="fieldErrors.report_title ? 'err-report-title' : undefined"
        />
        <small
          v-if="fieldErrors.report_title"
          id="err-report-title"
          class="research-form__error"
        >
          {{ fieldErrors.report_title }}
        </small>
      </label>
      <label class="research-form__field">
        <span class="research-form__label">{{ t('investmentResearch.companyOverview') }}</span>
        <textarea v-model="form.company_overview" class="textarea" rows="3" />
      </label>
      <label class="research-form__field">
        <span class="research-form__label">{{ t('investmentResearch.companyOutlook') }}</span>
        <textarea v-model="form.company_outlook" class="textarea" rows="3" />
      </label>
      <label class="research-form__field">
        <span class="research-form__label">{{ t('investmentResearch.esgComment') }}</span>
        <textarea v-model="form.esg_comment" class="textarea" rows="2" />
      </label>
      <label class="research-form__field">
        <span class="research-form__label">{{ t('investmentResearch.financialStatus') }}</span>
        <textarea v-model="form.financial_status" class="textarea" rows="3" />
      </label>
      <label class="research-form__field">
        <span class="research-form__label">
          {{ t('investmentResearch.investmentAnalysis') }}
          <span class="research-form__required" aria-hidden="true">*</span>
        </span>
        <textarea
          v-model="form.investment_analysis"
          data-field="investment_analysis"
          class="textarea"
          rows="6"
          :placeholder="t('investmentResearch.minimumAnalysisPlaceholder', { min: String(MIN_INVESTMENT_ANALYSIS_LENGTH) })"
          :aria-invalid="Boolean(fieldErrors.investment_analysis)"
          :aria-describedby="
            fieldErrors.investment_analysis
              ? 'err-analysis hint-analysis'
              : 'hint-analysis'
          "
        />
        <small
          id="hint-analysis"
          class="research-form__hint"
          :class="{
            'research-form__hint--warn':
              analysisLength < MIN_INVESTMENT_ANALYSIS_LENGTH,
          }"
        >
          {{ t('investmentResearch.characterCounterAnalysis', { count: String(analysisLength), min: String(MIN_INVESTMENT_ANALYSIS_LENGTH) }) }}
        </small>
        <small
          v-if="fieldErrors.investment_analysis"
          id="err-analysis"
          class="research-form__error"
        >
          {{ fieldErrors.investment_analysis }}
        </small>
      </label>
    </fieldset>

    <fieldset v-if="mode === 'edit'" class="research-form__section">
      <legend class="research-form__legend">{{ t('investmentResearch.lifecycleNotes') }}</legend>
      <div class="research-form__row">
        <label class="research-form__field">
          <span class="research-form__label">{{ t('investmentResearch.reportStatus') }}</span>
          <select v-model="form.report_status" class="select">
            <option value="">{{ t('investmentResearch.keepCurrent') }}</option>
            <option v-for="s in REPORT_STATUSES" :key="s" :value="s">
              {{ reportStatusLabel(s, t) }}
            </option>
          </select>
        </label>
      </div>
      <label class="research-form__field">
        <span class="research-form__label">{{ t('investmentResearch.rejectionReason') }}</span>
        <textarea v-model="form.rejection_reason" class="textarea" rows="2" />
      </label>
      <label class="research-form__field">
        <span class="research-form__label">{{ t('investmentResearch.postSubmissionNote') }}</span>
        <textarea v-model="form.post_submission_note" class="textarea" rows="2" />
      </label>
    </fieldset>

    <div v-if="error" class="research-form__alert" role="alert">{{ error }}</div>

    <div class="research-form__actions">
      <span class="research-form__shortcut-hint" aria-hidden="true">
        {{ mode === "create" ? t('investmentResearch.actions.shortcutHintCreate') : t('investmentResearch.actions.shortcutHintSave') }}
      </span>
      <AppButton
        type="button"
        variant="ghost"
        :disabled="saving"
        @click="emit('cancel')"
      >
        {{ t('investmentResearch.actions.cancel') }}
      </AppButton>
      <AppButton type="submit" variant="primary" :loading="saving">
        {{ mode === "create" ? t('investmentResearch.actions.createReport') : t('investmentResearch.actions.saveChanges') }}
      </AppButton>
    </div>
  </form>
</template>

<style scoped>
.research-form {
  display: grid;
  gap: var(--space-5);
}

.research-form__section {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-4);
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-md, 8px);
  background: var(--bg-card, #fff);
}

.research-form__legend {
  padding: 0 var(--space-2);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-primary, #111827);
}

.research-form__row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: var(--space-3);
}

.research-form__field {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.research-form__label {
  font-size: var(--font-size-xs, 0.75rem);
  font-weight: var(--font-weight-semibold, 600);
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-secondary, #4b5563);
}

.research-form__required {
  color: var(--state-error, #dc2626);
}

.research-form__hint {
  color: var(--text-tertiary, #6b7280);
  font-size: var(--font-size-xs, 0.75rem);
}

.research-form__hint--warn {
  color: var(--state-error, #dc2626);
}

.research-form__error {
  color: var(--state-error, #dc2626);
  font-size: var(--font-size-xs, 0.75rem);
}

.research-form__alert {
  padding: var(--space-3);
  background: rgba(220, 38, 38, 0.08);
  border: 1px solid var(--state-error, #dc2626);
  border-radius: var(--radius-md, 8px);
  color: var(--state-error, #dc2626);
}

.research-form__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.research-form__shortcut-hint {
  flex: 1;
  color: var(--text-tertiary, #9ca3af);
  font-size: var(--font-size-xs, 0.75rem);
}

input[aria-invalid="true"],
select[aria-invalid="true"],
textarea[aria-invalid="true"] {
  border-color: var(--state-error, #dc2626) !important;
  box-shadow: 0 0 0 1px var(--state-error, #dc2626);
}
</style>
