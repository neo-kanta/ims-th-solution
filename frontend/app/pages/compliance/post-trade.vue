<script setup lang="ts">
import { onMounted, ref } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";

import ComplianceBreachFilters from "~/features/compliance/components/ComplianceBreachFilters.vue";
import ComplianceBreachInbox from "~/features/compliance/components/ComplianceBreachInbox.vue";
import ComplianceBreachOverrideDialog from "~/features/compliance/components/ComplianceBreachOverrideDialog.vue";
import CompliancePager from "~/features/compliance/components/CompliancePager.vue";
import {
  useComplianceBreachesList,
  useComplianceBreachOverride,
} from "~/features/compliance/composables/useComplianceBreaches";
import type {
  ComplianceBreach,
  ComplianceBreachListFilters,
} from "~/features/compliance/types";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IRG_VIEW_RULES",
});

const { t } = useI18n();
const router = useRouter();
const authStore = useAuthStore();

const filters = ref<ComplianceBreachListFilters>({});
const breaches = useComplianceBreachesList();
const overrideMutation = useComplianceBreachOverride();

const overrideTarget = ref<ComplianceBreach | null>(null);
const canOverride = computed(() =>
  authStore.hasPermission("IRG_OVERRIDE_BREACH"),
);

async function refresh() {
  await breaches.fetchList(filters.value);
}

function setOffset(next: number) {
  breaches.offset.value = next;
  void refresh();
}

function setLimit(next: number) {
  breaches.limit.value = next;
  breaches.offset.value = 0;
  void refresh();
}

function openOverride(breach: ComplianceBreach) {
  overrideMutation.reset();
  overrideTarget.value = breach;
}

function cancelOverride() {
  overrideTarget.value = null;
}

async function submitOverride(payload: {
  reason: string;
  approved_by?: string;
}) {
  if (!overrideTarget.value) return;
  try {
    await overrideMutation.override(overrideTarget.value.id, payload);
    overrideTarget.value = null;
    await refresh();
  } catch {
    // error surfaced via the composable
  }
}

function viewCheckGroup(groupId: string) {
  void router.push({ path: "/compliance/audit", query: { group: groupId } });
}

onMounted(() => {
  void refresh();
});
</script>

<template>
  <section class="post-trade">
    <AppPageHeader
      :title="t('compliance.postTrade.title')"
      :description="t('compliance.postTrade.description')"
    />

    <div class="post-trade__layout">
      <aside class="post-trade__rail">
        <ComplianceBreachFilters
          v-model="filters"
          :loading="breaches.loading.value"
          @apply="refresh"
          @reset="refresh"
        />
      </aside>
      <div class="post-trade__main">
        <ComplianceBreachInbox
          :items="breaches.items.value"
          :loading="breaches.loading.value"
          :error="breaches.error.value"
          :can-override="canOverride"
          @override="openOverride"
          @view-group="viewCheckGroup"
        />
        <CompliancePager
          :total="breaches.total.value"
          :offset="breaches.offset.value"
          :limit="breaches.limit.value"
          :loading="breaches.loading.value"
          @update:offset="setOffset"
          @update:limit="setLimit"
        />
      </div>
    </div>

    <ComplianceBreachOverrideDialog
      :breach="overrideTarget"
      :submitting="overrideMutation.submitting.value"
      :error="overrideMutation.error.value"
      @cancel="cancelOverride"
      @submit="submitOverride"
    />
  </section>
</template>

<style scoped>
.post-trade {
  display: grid;
  gap: var(--space-7);
}

.post-trade__layout {
  display: grid;
  grid-template-columns: minmax(260px, 300px) 1fr;
  gap: var(--space-6);
  align-items: start;
}

@media (max-width: 1024px) {
  .post-trade__layout {
    grid-template-columns: 1fr;
  }
}
</style>
