<script setup lang="ts">
import { onBeforeUnmount, reactive, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import {
  RECOMMENDATIONS,
  REPORT_STATUSES,
  REVIEW_STATUSES,
  type ResearchRecommendation,
  type ResearchReportFilters,
  type ResearchReportStatus,
  type ResearchReviewStatus,
} from "../types";
import {
  recommendationLabel,
  reportStatusLabel,
  reviewStatusLabel,
} from "../lib/researchReportFormat";

const props = defineProps<{
  filters: ResearchReportFilters;
  loading: boolean;
}>();

const emit = defineEmits<{
  apply: [filters: ResearchReportFilters];
  reset: [];
}>();

const { t } = useI18n();

/**
 * Search-input debounce. 350ms matches the rest of the app's text-search
 * inputs and is short enough that keystrokes feel snappy without flooding
 * the backend.
 */
const SEARCH_DEBOUNCE_MS = 350;
let searchTimer: ReturnType<typeof setTimeout> | null = null;

function clearSearchTimer() {
  if (searchTimer !== null) {
    clearTimeout(searchTimer);
    searchTimer = null;
  }
}

onBeforeUnmount(clearSearchTimer);

interface LocalFilters {
  search: string;
  report_status: ResearchReportStatus | "";
  review_status: ResearchReviewStatus | "";
  recommendation: ResearchRecommendation | "";
  instrument_code: string;
  report_date_from: string;
  report_date_to: string;
}

const local = reactive<LocalFilters>({
  search: props.filters.search ?? "",
  report_status: (props.filters.report_status ?? "") as ResearchReportStatus | "",
  review_status: (props.filters.review_status ?? "") as ResearchReviewStatus | "",
  recommendation: (props.filters.recommendation ?? "") as ResearchRecommendation | "",
  instrument_code: props.filters.instrument_code ?? "",
  report_date_from: props.filters.report_date_from ?? "",
  report_date_to: props.filters.report_date_to ?? "",
});

watch(
  () => props.filters,
  (next) => {
    local.search = next.search ?? "";
    local.report_status = (next.report_status ?? "") as ResearchReportStatus | "";
    local.review_status = (next.review_status ?? "") as ResearchReviewStatus | "";
    local.recommendation = (next.recommendation ?? "") as ResearchRecommendation | "";
    local.instrument_code = next.instrument_code ?? "";
    local.report_date_from = next.report_date_from ?? "";
    local.report_date_to = next.report_date_to ?? "";
  },
  { deep: true },
);

function toFilters(): ResearchReportFilters {
  const out: ResearchReportFilters = {};
  if (local.search.trim()) out.search = local.search.trim();
  if (local.report_status) out.report_status = local.report_status;
  if (local.review_status) out.review_status = local.review_status;
  if (local.recommendation) out.recommendation = local.recommendation;
  if (local.instrument_code.trim()) out.instrument_code = local.instrument_code.trim().toUpperCase();
  if (local.report_date_from) out.report_date_from = local.report_date_from;
  if (local.report_date_to) out.report_date_to = local.report_date_to;
  return out;
}

function apply() {
  clearSearchTimer();
  emit("apply", toFilters());
}

function reset() {
  clearSearchTimer();
  local.search = "";
  local.report_status = "";
  local.review_status = "";
  local.recommendation = "";
  local.instrument_code = "";
  local.report_date_from = "";
  local.report_date_to = "";
  emit("reset");
}

/**
 * Trigger a debounced apply when the operator types in the free-text
 * search box. Other filters are explicit (selects/dates) and fire on
 * change immediately via the Apply button. Treating only `search` as
 * debounced keeps the keyboard experience snappy without spamming the
 * backend with one request per keystroke.
 */
function handleSearchInput() {
  clearSearchTimer();
  searchTimer = setTimeout(() => {
    emit("apply", toFilters());
  }, SEARCH_DEBOUNCE_MS);
}
</script>

<template>
  <form class="research-filters" @submit.prevent="apply">
    <div class="research-filters__row">
      <label class="research-filters__field">
        <span class="research-filters__label">{{ t('investmentResearch.search') }}</span>
        <input
          v-model="local.search"
          class="input"
          type="search"
          :placeholder="t('investmentResearch.searchPlaceholder')"
          autocomplete="off"
          @input="handleSearchInput"
        />
      </label>

      <label class="research-filters__field">
        <span class="research-filters__label">{{ t('investmentResearch.instrumentCode') }}</span>
        <input
          v-model="local.instrument_code"
          class="input"
          type="text"
          placeholder="PTT"
          autocomplete="off"
        />
      </label>

      <label class="research-filters__field">
        <span class="research-filters__label">{{ t('investmentResearch.recommendation') }}</span>
        <select v-model="local.recommendation" class="select">
          <option value="">{{ t('investmentResearch.all') }}</option>
          <option v-for="r in RECOMMENDATIONS" :key="r" :value="r">
            {{ recommendationLabel(r, t) }}
          </option>
        </select>
      </label>

      <label class="research-filters__field">
        <span class="research-filters__label">{{ t('investmentResearch.reportStatus') }}</span>
        <select v-model="local.report_status" class="select">
          <option value="">{{ t('investmentResearch.all') }}</option>
          <option v-for="s in REPORT_STATUSES" :key="s" :value="s">
            {{ reportStatusLabel(s, t) }}
          </option>
        </select>
      </label>

      <label class="research-filters__field">
        <span class="research-filters__label">{{ t('investmentResearch.reviewStatus') }}</span>
        <select v-model="local.review_status" class="select">
          <option value="">{{ t('investmentResearch.all') }}</option>
          <option v-for="s in REVIEW_STATUSES" :key="s" :value="s">
            {{ reviewStatusLabel(s, t) }}
          </option>
        </select>
      </label>

      <label class="research-filters__field">
        <span class="research-filters__label">{{ t('investmentResearch.reportDateFrom') }}</span>
        <input v-model="local.report_date_from" class="input" type="date" />
      </label>

      <label class="research-filters__field">
        <span class="research-filters__label">{{ t('investmentResearch.reportDateTo') }}</span>
        <input v-model="local.report_date_to" class="input" type="date" />
      </label>
    </div>

    <div class="research-filters__actions">
      <AppButton
        type="submit"
        variant="primary"
        size="sm"
        :loading="loading"
      >
        {{ t('investmentResearch.actions.apply') }}
      </AppButton>
      <AppButton type="button" variant="ghost" size="sm" @click="reset">
        {{ t('investmentResearch.actions.reset') }}
      </AppButton>
    </div>
  </form>
</template>

<style scoped>
.research-filters {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-4);
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-md, 8px);
}

.research-filters__row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: var(--space-3);
}

.research-filters__field {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.research-filters__label {
  font-size: var(--font-size-xs, 0.75rem);
  font-weight: var(--font-weight-semibold, 600);
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-secondary, #4b5563);
}

.research-filters__actions {
  display: flex;
  gap: var(--space-2);
  justify-content: flex-end;
}
</style>
