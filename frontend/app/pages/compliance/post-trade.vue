<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import { useBangkokFormatter } from "~/shared/composables/useBangkokFormatter";
import AppButton from "~/shared/ui/AppButton.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";

import ComplianceBreachDetailDrawer from "~/features/compliance/components/ComplianceBreachDetailDrawer.vue";
import ComplianceBreachFilters from "~/features/compliance/components/ComplianceBreachFilters.vue";
import ComplianceBreachOverrideDialog from "~/features/compliance/components/ComplianceBreachOverrideDialog.vue";
import ComplianceBreachTable from "~/features/compliance/components/ComplianceBreachTable.vue";
import CompliancePager from "~/features/compliance/components/CompliancePager.vue";
import {
  useComplianceBreachesList,
  useComplianceBreachOverride,
} from "~/features/compliance/composables/useComplianceBreaches";
import {
  apiFiltersFromState,
  breachFilterStatesEqual,
  defaultBreachFilterState,
  filterStateFromQuery,
  hasNonDefaultFilters,
  queryFromFilterState,
} from "~/features/compliance/lib/breachFilters";
import type { ComplianceBreach } from "~/features/compliance/types";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IRG_VIEW_RULES",
});

const { t } = useI18n();
const { formatDateTime } = useBangkokFormatter();
const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const filters = ref(filterStateFromQuery(route.query));
const breaches = useComplianceBreachesList();
const overrideMutation = useComplianceBreachOverride();

const lastRefreshedAt = ref<string | null>(null);
const overrideTarget = ref<ComplianceBreach | null>(null);
const detailTarget = ref<ComplianceBreach | null>(null);

const canOverride = computed(() => authStore.hasPermission("IRG_OVERRIDE_BREACH"));
const activeFilters = computed(() => hasNonDefaultFilters(filters.value));
const lastRefreshedLabel = computed(() =>
  lastRefreshedAt.value
    ? t("compliance.postTrade.lastRefreshed", { time: formatDateTime(lastRefreshedAt.value) })
    : t("compliance.postTrade.lastRefreshedNever"),
);

const overrideDisplayError = computed(() => {
  if (!overrideMutation.error.value) return null;
  return overrideMutation.conflict.value
    ? t("compliance.postTrade.override.conflict")
    : overrideMutation.error.value;
});

let syncingRouteFromFilters = false;

async function refresh() {
  await breaches.fetchList(apiFiltersFromState(filters.value));
  lastRefreshedAt.value = new Date().toISOString();
}

watch(
  filters,
  (next, prev) => {
    if (prev && breachFilterStatesEqual(next, prev)) return;
    breaches.offset.value = 0;
    syncingRouteFromFilters = true;
    void router.replace({ query: queryFromFilterState(next) });
    void refresh();
  },
  { deep: true },
);

watch(
  () => route.query,
  (query) => {
    if (syncingRouteFromFilters) {
      syncingRouteFromFilters = false;
      return;
    }
    const next = filterStateFromQuery(query);
    if (breachFilterStatesEqual(next, filters.value)) return;
    filters.value = next;
    breaches.offset.value = 0;
    void refresh();
  },
);

function setOffset(next: number) {
  breaches.offset.value = next;
  void refresh();
}

function setLimit(next: number) {
  breaches.limit.value = next;
  breaches.offset.value = 0;
  void refresh();
}

function clearFilters() {
  filters.value = defaultBreachFilterState();
}

function openDetail(breach: ComplianceBreach) {
  detailTarget.value = breach;
}

function closeDetail() {
  detailTarget.value = null;
}

function openOverride(breach: ComplianceBreach) {
  overrideMutation.reset();
  overrideTarget.value = breach;
}

function cancelOverride() {
  if (overrideMutation.submitting.value) return;
  overrideTarget.value = null;
}

async function submitOverride(payload: { reason: string }) {
  if (!overrideTarget.value) return;
  try {
    await overrideMutation.override(overrideTarget.value.id, payload);
    overrideTarget.value = null;
    detailTarget.value = null;
    await refresh();
  } catch {
    // error surfaced via the composable; dialog stays open for retry
  }
}

onMounted(() => {
  void router.replace({ query: queryFromFilterState(filters.value) });
  void refresh();
});
</script>

<template>
  <section class="post-trade">
    <AppPageHeader
      :title="t('compliance.postTrade.title')"
      :description="t('compliance.postTrade.description')"
    >
      <template #actions>
        <span class="post-trade__last-refreshed">{{ lastRefreshedLabel }}</span>
        <AppButton
          variant="secondary"
          size="sm"
          :loading="breaches.loading.value"
          @click="refresh"
        >
          {{ t("compliance.postTrade.refresh") }}
        </AppButton>
      </template>
    </AppPageHeader>

    <p class="post-trade__limitation">{{ t("compliance.postTrade.limitationNotice") }}</p>

    <ComplianceBreachFilters v-model="filters" :loading="breaches.loading.value" />

    <ComplianceBreachTable
      :items="breaches.items.value"
      :loading="breaches.loading.value"
      :error="breaches.error.value"
      :can-override="canOverride"
      :has-active-filters="activeFilters"
      @view-details="openDetail"
      @override="openOverride"
      @retry="refresh"
      @clear-filters="clearFilters"
    />

    <CompliancePager
      :total="breaches.total.value"
      :offset="breaches.offset.value"
      :limit="breaches.limit.value"
      :loading="breaches.loading.value"
      @update:offset="setOffset"
      @update:limit="setLimit"
    />

    <ComplianceBreachDetailDrawer
      :breach="detailTarget"
      :can-override="canOverride"
      @close="closeDetail"
      @override="openOverride"
    />

    <ComplianceBreachOverrideDialog
      :breach="overrideTarget"
      :submitting="overrideMutation.submitting.value"
      :error="overrideDisplayError"
      @cancel="cancelOverride"
      @submit="submitOverride"
    />
  </section>
</template>

<style scoped>
.post-trade {
  display: grid;
  gap: var(--space-5);
}

.post-trade__last-refreshed {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  white-space: nowrap;
}

.post-trade__limitation {
  margin: 0;
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}
</style>
