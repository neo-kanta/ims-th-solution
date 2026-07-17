<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRouter } from "#imports";

import AppPage from "~/shared/ui/AppPage.vue";
import AppFormField from "~/shared/ui/AppFormField.vue";
import AppInput from "~/shared/ui/AppInput.vue";
import AppDataTable, { type TableColumn } from "~/shared/ui/AppDataTable.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppBadge from "~/shared/ui/AppBadge.vue";
import { useI18n } from "~/composables/useI18n";

import { usePortfolioPicker } from "./composables/usePortfolioPicker";
import { newDecisionPath } from "./lib/decisionRoutes";
import type { PortfolioPick } from "./lib/portfolioPicker";

const router = useRouter();
const { t } = useI18n();

const { portfolios, filtered, search, loading, error, load } = usePortfolioPicker();

const columns = computed<TableColumn[]>(() => [
  { key: "code", label: t("portfolio.directory.columns.code") },
  { key: "name", label: t("portfolio.directory.columns.name") },
  { key: "status", label: t("portfolio.directory.columns.status") },
  { key: "actions", label: "", align: "right" },
]);

const rows = computed(() => filtered.value.map((p) => ({ ...p, id: p.code })));

const emptyText = computed(() =>
  search.value.trim() ? t("portfolio.decisionPicker.noMatches") : t("portfolio.directory.empty"),
);

function openDecision(portfolio: PortfolioPick) {
  void router.push(newDecisionPath(portfolio.code));
}

onMounted(() => {
  void load();
});
</script>

<template>
  <AppPage
    :title="t('portfolio.decisionPicker.title')"
    :subtitle="t('portfolio.decisionPicker.description')"
    :loading="loading && portfolios.length === 0"
    :error="portfolios.length === 0 ? error : null"
    @retry="load"
  >
    <AppFormField :label="t('portfolio.decisionPicker.searchLabel')" id="portfolio-decision-picker-search">
      <AppInput
        id="portfolio-decision-picker-search"
        v-model="search"
        type="search"
        :placeholder="t('portfolio.decisionPicker.searchPlaceholder')"
      />
    </AppFormField>

    <AppDataTable
      :columns="columns"
      :items="rows"
      :loading="loading && portfolios.length > 0"
      :empty-text="emptyText"
      row-clickable
      @row-click="openDecision"
    >
      <template #cell(status)="{ item }">
        <AppBadge :variant="item.status === 'ACTIVE' ? 'success' : 'neutral'" size="sm">
          {{ item.status ?? "—" }}
        </AppBadge>
      </template>
      <template #cell(actions)="{ item }">
        <AppButton variant="primary" size="xs" @click.stop="openDecision(item)">
          {{ t("portfolio.decisionPicker.action") }}
        </AppButton>
      </template>
    </AppDataTable>
  </AppPage>
</template>
