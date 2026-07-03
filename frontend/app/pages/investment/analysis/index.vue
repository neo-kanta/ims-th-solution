<script setup lang="ts">
import { onMounted, ref, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import ResearchReportFilters from "~/features/investment-research/components/ResearchReportFilters.vue";
import ResearchReportTable from "~/features/investment-research/components/ResearchReportTable.vue";
import { useResearchReportsList } from "~/features/investment-research/composables/useResearchReports";
import type {
  ResearchRecommendation,
  ResearchReportFilters as Filters,
  ResearchReportStatus,
  ResearchReviewStatus,
} from "~/features/investment-research/types";
import {
  RECOMMENDATIONS,
  REPORT_STATUSES,
  REVIEW_STATUSES,
} from "~/features/investment-research/types";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_RESEARCH_VIEW",
});

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const { items, total, page, limit, loading, error, fetchList } =
  useResearchReportsList();

const filters = ref<Filters>({});

/**
 * URL ⇄ filters sync. Query parameters are the single source of truth so
 * (a) the back/forward buttons restore state, and (b) operators can copy
 * the URL into a chat to share a filtered view.
 *
 * Only known/whitelisted keys are accepted from the URL — anything else
 * is dropped silently to keep the filter object well-typed.
 */
function pickEnum<T extends string>(
  raw: unknown,
  allowed: readonly T[],
): T | undefined {
  if (typeof raw !== "string") return undefined;
  return (allowed as readonly string[]).includes(raw) ? (raw as T) : undefined;
}

function firstString(raw: unknown): string | undefined {
  if (typeof raw === "string") return raw;
  if (Array.isArray(raw) && typeof raw[0] === "string") return raw[0];
  return undefined;
}

function parseIntInRange(raw: unknown, min: number, max: number): number | undefined {
  const s = firstString(raw);
  if (!s) return undefined;
  const n = Number.parseInt(s, 10);
  if (!Number.isFinite(n) || n < min || n > max) return undefined;
  return n;
}

function hydrateFromQuery() {
  const q = route.query;
  const next: Filters = {};
  const search = firstString(q.search);
  if (search) next.search = search;
  const status = pickEnum<ResearchReportStatus>(
    firstString(q.report_status),
    REPORT_STATUSES,
  );
  if (status) next.report_status = status;
  const reviewStatus = pickEnum<ResearchReviewStatus>(
    firstString(q.review_status),
    REVIEW_STATUSES,
  );
  if (reviewStatus) next.review_status = reviewStatus;
  const rec = pickEnum<ResearchRecommendation>(
    firstString(q.recommendation),
    RECOMMENDATIONS,
  );
  if (rec) next.recommendation = rec;
  const code = firstString(q.instrument_code);
  if (code) next.instrument_code = code;
  const owner = firstString(q.owner_user_id);
  if (owner) next.owner_user_id = owner;
  const from = firstString(q.report_date_from);
  if (from) next.report_date_from = from;
  const to = firstString(q.report_date_to);
  if (to) next.report_date_to = to;
  filters.value = next;

  const pageFromUrl = parseIntInRange(q.page, 1, 10_000);
  if (pageFromUrl) page.value = pageFromUrl;
  const limitFromUrl = parseIntInRange(q.limit, 1, 200);
  if (limitFromUrl) limit.value = limitFromUrl;
}

function writeQuery() {
  const q: Record<string, string> = {};
  for (const [key, value] of Object.entries(filters.value)) {
    if (value === undefined || value === null || value === "") continue;
    q[key] = String(value);
  }
  if (page.value > 1) q.page = String(page.value);
  // Only include limit when it diverges from the composable's default so
  // the URL stays clean for the common case.
  if (limit.value !== 20) q.limit = String(limit.value);
  void router.replace({ query: q });
}

async function refresh() {
  await fetchList(filters.value);
  writeQuery();
}

function applyFilters(next: Filters) {
  filters.value = next;
  page.value = 1;
  void refresh();
}

function resetFilters() {
  filters.value = {};
  page.value = 1;
  void refresh();
}

function openReport(id: string) {
  void router.push(`/investment/analysis/${id}`);
}

function goCreate() {
  void router.push("/investment/analysis/new");
}

function prevPage() {
  if (page.value <= 1 || loading.value) return;
  page.value -= 1;
  void refresh();
}

function nextPage() {
  if (page.value * limit.value >= total.value || loading.value) return;
  page.value += 1;
  void refresh();
}

// Re-hydrate from the URL when the operator presses back/forward — Nuxt's
// route object is reactive so we watch the entire query bag.
watch(
  () => route.query,
  () => {
    hydrateFromQuery();
    void fetchList(filters.value);
  },
);

onMounted(() => {
  hydrateFromQuery();
  void refresh();
});
</script>

<template>
  <section class="analysis-page">
    <AppPageHeader
      :title="t('investmentResearch.title')"
      :description="t('investmentResearch.descriptionList')"
    >
      <template #actions>
        <AppButton variant="primary" @click="goCreate">{{ t('investmentResearch.newReport') }}</AppButton>
      </template>
    </AppPageHeader>

    <ResearchReportFilters
      :filters="filters"
      :loading="loading"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <ResearchReportTable
      :items="items"
      :loading="loading"
      :error="error"
      @open="openReport"
    />

    <footer class="analysis-page__footer">
      <span class="analysis-page__range" aria-live="polite">
        {{ t('investmentResearch.pager.range', { page: String(page), count: String(items.length), total: String(total) }) }}
      </span>
      <div class="analysis-page__pager">
        <AppButton
          variant="secondary"
          size="sm"
          :disabled="page <= 1 || loading"
          @click="prevPage"
        >
          {{ t('investmentResearch.actions.prev') }}
        </AppButton>
        <AppButton
          variant="secondary"
          size="sm"
          :disabled="page * limit >= total || loading"
          @click="nextPage"
        >
          {{ t('investmentResearch.actions.next') }}
        </AppButton>
      </div>
    </footer>
  </section>
</template>

<style scoped>
.analysis-page {
  display: grid;
  gap: var(--space-5);
}

.analysis-page__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-3);
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-md, 8px);
}

.analysis-page__range {
  font-size: var(--font-size-sm, 0.875rem);
  color: var(--text-secondary, #4b5563);
}

.analysis-page__pager {
  display: flex;
  gap: var(--space-2);
}
</style>
