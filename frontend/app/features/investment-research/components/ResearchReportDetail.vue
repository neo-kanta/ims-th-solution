<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import ResearchReportStatusBadge from "./ResearchReportStatusBadge.vue";
import type { ResearchReport } from "../types";
import {
  canCancelSubmitReport,
  canDeleteReport,
  canEditReport,
  canSubmitReport,
  formatBangkokDateTime,
} from "../lib/researchReportFormat";

const props = defineProps<{
  report: ResearchReport;
  saving: boolean;
  error: string | null;
}>();

const emit = defineEmits<{
  edit: [];
  delete: [];
  submit: [];
  cancelSubmit: [];
  back: [];
}>();

const { t } = useI18n();

function plainText(value: string | null | undefined) {
  if (!value || !value.trim()) return "—";
  return value;
}

/**
 * Surfaced in the audit footer. Backend stores UTC; project policy is to
 * display Asia/Bangkok (UTC+7). See AIREAD.md §7.
 */
const createdAtLocal = computed(() => formatBangkokDateTime(props.report.created_at));
const updatedAtLocal = computed(() => formatBangkokDateTime(props.report.updated_at));
</script>

<template>
  <div class="research-detail">
    <header class="research-detail__header">
      <div class="research-detail__heading">
        <span class="research-detail__report-no">{{ report.report_no }}</span>
        <h2 class="research-detail__title">{{ report.report_title || t('investmentResearch.untitledReport') }}</h2>
        <div class="research-detail__badges">
          <ResearchReportStatusBadge
            kind="recommendation"
            :value="report.recommendation"
          />
          <ResearchReportStatusBadge kind="report" :value="report.report_status" />
          <ResearchReportStatusBadge kind="review" :value="report.review_status" />
        </div>
      </div>
      <div class="research-detail__actions">
        <AppButton variant="ghost" size="sm" @click="emit('back')">
          {{ t('investmentResearch.actions.back') }}
        </AppButton>
        <AppButton
          v-if="canEditReport(report.review_status)"
          variant="secondary"
          size="sm"
          @click="emit('edit')"
        >
          {{ t('investmentResearch.actions.edit') }}
        </AppButton>
        <AppButton
          v-if="canSubmitReport(report.review_status)"
          variant="primary"
          size="sm"
          :loading="saving"
          @click="emit('submit')"
        >
          {{ t('investmentResearch.actions.submit') }}
        </AppButton>
        <AppButton
          v-if="canCancelSubmitReport(report.review_status)"
          variant="warning"
          size="sm"
          :loading="saving"
          @click="emit('cancelSubmit')"
        >
          {{ t('investmentResearch.actions.cancelSubmission') }}
        </AppButton>
        <AppButton
          v-if="canDeleteReport(report.review_status)"
          variant="danger"
          size="sm"
          :loading="saving"
          @click="emit('delete')"
        >
          {{ t('investmentResearch.actions.delete') }}
        </AppButton>
      </div>
    </header>

    <div v-if="error" class="research-detail__alert" role="alert">{{ error }}</div>

    <section class="research-detail__grid">
      <div class="research-detail__cell">
        <div class="research-detail__cell-label">{{ t('investmentResearch.instrument') }}</div>
        <div class="research-detail__cell-value">
          <strong>{{ report.instrument_code }}</strong>
          <span v-if="report.instrument_name"> · {{ report.instrument_name }}</span>
        </div>
      </div>
      <div class="research-detail__cell">
        <div class="research-detail__cell-label">{{ t('investmentResearch.typeMarketCurrency') }}</div>
        <div class="research-detail__cell-value">
          {{ plainText(report.instrument_type) }} · {{ plainText(report.market) }} ·
          {{ plainText(report.currency) }}
        </div>
      </div>
      <div class="research-detail__cell">
        <div class="research-detail__cell-label">{{ t('investmentResearch.reportDate') }}</div>
        <div class="research-detail__cell-value">{{ report.report_date }}</div>
      </div>
      <div class="research-detail__cell">
        <div class="research-detail__cell-label">{{ t('investmentResearch.effectiveDate') }}</div>
        <div class="research-detail__cell-value">
          {{ plainText(report.effective_date) }}
        </div>
      </div>
      <div class="research-detail__cell">
        <div class="research-detail__cell-label">{{ t('investmentResearch.ownerAuthor') }}</div>
        <div class="research-detail__cell-value">
          {{ report.owner_user_id }} <br />
          <small>{{ t('investmentResearch.by') }} {{ report.author_user_id }}</small>
        </div>
      </div>
      <div class="research-detail__cell">
        <div class="research-detail__cell-label">{{ t('investmentResearch.applicableContract') }}</div>
        <div class="research-detail__cell-value">
          {{ plainText(report.applicable_contract_id) }}
        </div>
      </div>
    </section>

    <section class="research-detail__section">
      <h3>{{ t('investmentResearch.companyOverview') }}</h3>
      <p>{{ plainText(report.company_overview) }}</p>
    </section>
    <section class="research-detail__section">
      <h3>{{ t('investmentResearch.companyOutlook') }}</h3>
      <p>{{ plainText(report.company_outlook) }}</p>
    </section>
    <section class="research-detail__section">
      <h3>{{ t('investmentResearch.financialStatus') }}</h3>
      <p>{{ plainText(report.financial_status) }}</p>
    </section>
    <section class="research-detail__section">
      <h3>{{ t('investmentResearch.esgComment') }}</h3>
      <p>{{ plainText(report.esg_comment) }}</p>
    </section>
    <section class="research-detail__section">
      <h3>{{ t('investmentResearch.investmentAnalysis') }}</h3>
      <p>{{ plainText(report.investment_analysis) }}</p>
    </section>
    <section
      v-if="report.rejection_reason || report.post_submission_note"
      class="research-detail__section"
    >
      <h3>{{ t('investmentResearch.lifecycleNotes') }}</h3>
      <p v-if="report.rejection_reason"><strong>{{ t('investmentResearch.rejectionReason') }}:</strong> {{ report.rejection_reason }}</p>
      <p v-if="report.post_submission_note">
        <strong>{{ t('investmentResearch.postSubmissionNote') }}:</strong> {{ report.post_submission_note }}
      </p>
    </section>

    <footer class="research-detail__footer">
      <div class="research-detail__footer-cell">
        <span class="research-detail__footer-label">{{ t('investmentResearch.created') }}</span>
        <span class="research-detail__footer-value">{{ createdAtLocal }}</span>
        <small v-if="report.created_by" class="research-detail__footer-meta">
          {{ t('investmentResearch.by') }} {{ report.created_by }}
        </small>
      </div>
      <div class="research-detail__footer-cell">
        <span class="research-detail__footer-label">{{ t('investmentResearch.lastUpdated') }}</span>
        <span class="research-detail__footer-value">{{ updatedAtLocal }}</span>
        <small v-if="report.updated_by" class="research-detail__footer-meta">
          {{ t('investmentResearch.by') }} {{ report.updated_by }}
        </small>
      </div>
      <div class="research-detail__footer-cell">
        <span class="research-detail__footer-label">{{ t('investmentResearch.timezone') }}</span>
        <span class="research-detail__footer-value">{{ t('investmentResearch.timezoneValue') }}</span>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.research-detail {
  display: grid;
  gap: var(--space-5);
}

.research-detail__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.research-detail__heading {
  display: grid;
  gap: var(--space-1);
}

.research-detail__report-no {
  font-size: var(--font-size-xs, 0.75rem);
  font-weight: var(--font-weight-semibold, 600);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text-secondary, #6b7280);
}

.research-detail__title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: var(--font-weight-semibold, 600);
}

.research-detail__badges {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
  margin-top: var(--space-1);
}

.research-detail__actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.research-detail__alert {
  padding: var(--space-3);
  border: 1px solid var(--state-error, #dc2626);
  background: rgba(220, 38, 38, 0.08);
  color: var(--state-error, #dc2626);
  border-radius: var(--radius-md, 8px);
}

.research-detail__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: var(--space-3);
  padding: var(--space-4);
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-md, 8px);
}

.research-detail__cell-label {
  font-size: var(--font-size-xs, 0.75rem);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-secondary, #6b7280);
}

.research-detail__cell-value {
  font-size: var(--font-size-sm, 0.875rem);
  word-break: break-word;
}

.research-detail__section {
  padding: var(--space-4);
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-md, 8px);
}

.research-detail__section h3 {
  margin: 0 0 var(--space-2);
  font-size: var(--font-size-sm, 0.875rem);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-secondary, #6b7280);
}

.research-detail__section p {
  margin: 0;
  white-space: pre-wrap;
  color: var(--text-primary, #111827);
}

.research-detail__footer {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--bg-card-muted, #f9fafb);
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-md, 8px);
}

.research-detail__footer-cell {
  display: grid;
  gap: 2px;
}

.research-detail__footer-label {
  font-size: var(--font-size-xs, 0.75rem);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-secondary, #6b7280);
}

.research-detail__footer-value {
  font-size: var(--font-size-sm, 0.875rem);
  color: var(--text-primary, #111827);
}

.research-detail__footer-meta {
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-tertiary, #9ca3af);
  word-break: break-all;
}
</style>
