<script setup lang="ts">
import { onMounted, watch } from "vue";

import { useI18n } from "~/composables/useI18n";

import { useFundDetail } from "../../composables/useFundDetail";

import FundDetailHeader from "./FundDetailHeader.vue";

interface Props {
  fundId: string;
  activeTab: string;
}

const props = defineProps<Props>();
const { t } = useI18n();
const router = useRouter();

const { card, loading, error, load } = useFundDetail(() => props.fundId);

onMounted(() => {
  void load();
});

watch(
  () => props.fundId,
  () => {
    void load();
  },
);

function back() {
  void router.push("/investment/funds");
}

function selectTab(tab: string) {
  if (tab === props.activeTab) return;
  void router.push(`/investment/funds/${props.fundId}/${tab}`);
}
</script>

<template>
  <section class="fund-detail-layout">
    <FundDetailHeader
      v-if="card"
      :card="card"
      :active-tab="activeTab"
      :loading="loading"
      @back="back"
      @select-tab="selectTab"
    />

    <div v-else-if="loading" class="fund-detail-layout__loading">
      {{ t("myFunds.detail.loading", "Loading fund…") }}
    </div>

    <div v-else-if="error" class="fund-detail-layout__error" role="alert">
      <div class="fund-detail-layout__error-title">
        {{ t("myFunds.detail.errorTitle", "Could not load fund") }}
      </div>
      <div class="fund-detail-layout__error-detail">{{ error }}</div>
      <button class="fund-detail-layout__retry" type="button" @click="back">
        {{ t("myFunds.detail.backToList", "Back to my funds") }}
      </button>
    </div>

    <div v-if="card" class="fund-detail-layout__body">
      <slot :card="card" />
    </div>
  </section>
</template>

<style scoped>
.fund-detail-layout {
  display: grid;
  gap: var(--space-4);
}

.fund-detail-layout__loading {
  padding: var(--space-4);
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-secondary, #57606a);
  text-align: center;
}

.fund-detail-layout__error {
  padding: var(--space-4);
  background: rgba(207, 34, 46, 0.06);
  border: 1px solid rgba(207, 34, 46, 0.3);
  border-radius: 6px;
  display: grid;
  gap: 6px;
}

.fund-detail-layout__error-title {
  font-weight: 600;
  color: var(--state-danger, #cf222e);
}

.fund-detail-layout__error-detail {
  font-size: 12px;
  color: var(--text-secondary, #57606a);
}

.fund-detail-layout__retry {
  justify-self: start;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  border: 1px solid var(--border-subtle, #d0d7de);
  background: var(--bg-card, #ffffff);
  border-radius: 5px;
  padding: 4px 10px;
  cursor: pointer;
}

.fund-detail-layout__body {
  display: grid;
  gap: var(--space-3);
}
</style>
