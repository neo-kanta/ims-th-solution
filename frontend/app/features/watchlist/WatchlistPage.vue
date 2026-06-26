<script setup lang="ts">
import { ref, reactive, onMounted, watch } from "vue";
import { useRoute } from "#imports";
import { useI18n } from "~/composables/useI18n";
import AppPage from "~/shared/ui/AppPage.vue";
import AppConfirmDialog from "~/shared/ui/AppConfirmDialog.vue";
import WatchlistItemsTable from "./components/WatchlistItemsTable.vue";
import WatchlistAlertsPanel from "./components/WatchlistAlertsPanel.vue";
import WatchlistFilters from "./components/WatchlistFilters.vue";
import WatchlistItemDrawer from "./components/WatchlistItemDrawer.vue";
import { useWatchlist } from "./composables/useWatchlist";
import { useWatchlistAlerts } from "./composables/useWatchlistAlerts";
import { useWatchlistPortfolios } from "./composables/useWatchlistPortfolios";
import { useWatchlistSecuritySearch } from "./composables/useWatchlistSecuritySearch";
import { useAppToast } from "~/composables/useAppToast";
import { OpenApiRequestError } from "~/api/openapi";
import type { WatchlistItem, CreateItemBody, UpdateItemBody } from "./types";
import type { WatchlistFilterState } from "./components/WatchlistFilters.vue";

const route = useRoute();
const toast = useAppToast();
const { t } = useI18n();

const watchlist = useWatchlist();
const alertsComposable = useWatchlistAlerts();
const portfolios = useWatchlistPortfolios();
const securitySearch = useWatchlistSecuritySearch();

const filters = reactive<WatchlistFilterState>({
  items: {},
  alerts: {},
});

const drawerOpen = ref(false);
const drawerMode = ref<"create" | "edit">("create");
const editingItem = ref<WatchlistItem | null>(null);

const pageLoading = ref(true);
const pageError = ref<string | null>(null);

const deleteConfirmOpen = ref(false);
const itemToDelete = ref<WatchlistItem | null>(null);
const deleting = ref(false);

async function loadAll() {
  pageLoading.value = true;
  pageError.value = null;
  try {
    await Promise.all([
      watchlist.load(filters.items),
      alertsComposable.load(filters.alerts),
      portfolios.load(),
    ]);

    const alertId = route.query.alert_id as string | undefined;
    if (alertId) {
      const found = alertsComposable.alerts.value.find((a) => a.id === alertId);
      if (!found) {
        toast.showWarning(t("watchlist.alert.notFound"));
      }
    }
  } catch (err: unknown) {
    pageError.value = err instanceof Error ? err.message : "Failed to load watchlist.";
  } finally {
    pageLoading.value = false;
  }
}

onMounted(loadAll);

watch(
  () => filters.items,
  () => watchlist.load(filters.items).catch(() => {}),
  { deep: true },
);

watch(
  () => filters.alerts,
  () => alertsComposable.load(filters.alerts).catch(() => {}),
  { deep: true },
);

function openCreate() {
  drawerMode.value = "create";
  editingItem.value = null;
  drawerOpen.value = true;
}

function openEdit(item: WatchlistItem) {
  drawerMode.value = "edit";
  editingItem.value = item;
  drawerOpen.value = true;
}

function closeDrawer() {
  drawerOpen.value = false;
  editingItem.value = null;
}

function mapErrorCode(err: unknown): string {
  if (!(err instanceof OpenApiRequestError)) {
    return err instanceof Error ? err.message : "An unexpected error occurred.";
  }
  const details = err.details as Record<string, unknown> | null | undefined;
  const code = details && typeof details === "object" ? (details["code"] as string | undefined) : undefined;
  switch (code) {
    case "WATCHLIST_DUPLICATE_ITEM":
      return t("watchlist.error.duplicateItem");
    case "WATCHLIST_FORBIDDEN_SCOPE":
      return t("watchlist.error.forbiddenScope");
    case "WATCHLIST_INVALID_SECURITY":
      return t("watchlist.error.invalidSecurity");
    case "WATCHLIST_INVALID_THRESHOLD":
      return t("watchlist.error.invalidThreshold");
    case "WATCHLIST_DUPLICATE_THRESHOLD":
      return t("watchlist.error.invalidThreshold");
    case "WATCHLIST_ITEM_NOT_FOUND":
      return t("watchlist.error.notFound");
    default:
      return err.message || "An unexpected error occurred.";
  }
}

async function onCreate(body: CreateItemBody) {
  try {
    await watchlist.createItem(body);
    toast.showSuccess("Security added to watchlist.");
    closeDrawer();
    await watchlist.load(filters.items);
  } catch (err: unknown) {
    toast.showError(mapErrorCode(err));
  }
}

async function onUpdate(id: string, body: UpdateItemBody) {
  try {
    await watchlist.updateItem(id, body);
    toast.showSuccess("Watchlist item updated.");
    closeDrawer();
    await watchlist.load(filters.items);
  } catch (err: unknown) {
    toast.showError(mapErrorCode(err));
  }
}

function triggerDelete(item: WatchlistItem) {
  itemToDelete.value = item;
  deleteConfirmOpen.value = true;
}

async function confirmDelete() {
  const item = itemToDelete.value;
  if (!item || !item.id) return;
  deleting.value = true;
  try {
    await watchlist.deleteItem(item.id);
    toast.showSuccess("Watchlist item removed.");
    deleteConfirmOpen.value = false;
    itemToDelete.value = null;
    await watchlist.load(filters.items);
  } catch (err: unknown) {
    toast.showError(mapErrorCode(err));
    if (err instanceof OpenApiRequestError) {
      const details = err.details as Record<string, unknown> | null | undefined;
      const code = details && typeof details === "object" ? (details["code"] as string | undefined) : undefined;
      if (code === "WATCHLIST_ITEM_NOT_FOUND") {
        await watchlist.load(filters.items);
      }
    }
  } finally {
    deleting.value = false;
  }
}

async function onAcknowledge(alertId: string, note: string | undefined) {
  try {
    await alertsComposable.acknowledge(alertId, { note });
    toast.showSuccess("Alert acknowledged.");
    await alertsComposable.load(filters.alerts);
  } catch (err: unknown) {
    if (err instanceof OpenApiRequestError) {
      const details = err.details as Record<string, unknown> | null | undefined;
      const code = details && typeof details === "object" ? (details["code"] as string | undefined) : undefined;
      if (code === "WATCHLIST_ALERT_ALREADY_ACKNOWLEDGED") {
        toast.showWarning(t("watchlist.error.alertAlreadyAcknowledged"));
        await alertsComposable.load(filters.alerts);
        return;
      }
    }
    toast.showError(mapErrorCode(err));
  }
}
</script>

<template>
  <AppPage
    :title="t('watchlist.page.title')"
    :subtitle="t('watchlist.page.subtitle')"
    :loading="pageLoading"
    :error="pageError"
    width="full"
    @retry="loadAll"
  >
    <template #actions>
      <IMSPermissionGuard permission="WATCHLIST_MANAGE">
        <button class="btn btn-primary btn-sm" @click="openCreate">
          + {{ t('watchlist.page.addItem') }}
        </button>
      </IMSPermissionGuard>
    </template>

    <div style="display: flex; flex-direction: column; gap: 20px;">
      <WatchlistFilters
        v-model="filters"
        :portfolios="portfolios.portfolios.value"
        :portfolios-loading="portfolios.loading.value"
        :portfolios-error="portfolios.error.value"
        :security-search-results="securitySearch.results.value"
        :security-search-loading="securitySearch.loading.value"
        @security-search="securitySearch.search"
      />

      <div style="display: grid; grid-template-columns: 1fr 360px; gap: 20px; align-items: start;">
        <div>
          <WatchlistItemsTable
            :items="watchlist.items.value"
            :loading="watchlist.loading.value"
            :error="watchlist.error.value"
            @edit="openEdit"
            @delete="triggerDelete"
          />
        </div>

        <div>
          <div style="font-size: 14px; font-weight: 600; margin-bottom: 12px; color: var(--color-text, #1e293b);">
            {{ t('watchlist.alert.title') }}
          </div>
          <WatchlistAlertsPanel
            :alerts="alertsComposable.alerts.value"
            :loading="alertsComposable.loading.value"
            :error="alertsComposable.error.value"
            :acknowledging="alertsComposable.acknowledging.value"
            :highlighted-alert-id="route.query.alert_id as string"
            @acknowledge="onAcknowledge"
          />
        </div>
      </div>
    </div>

    <template v-if="drawerOpen">
      <WatchlistItemDrawer
        :mode="drawerMode"
        :item="editingItem"
        :portfolios="portfolios.portfolios.value"
        :portfolios-loading="portfolios.loading.value"
        :portfolios-error="portfolios.error.value"
        :saving="watchlist.saving.value"
        :security-search-results="securitySearch.results.value"
        :security-search-loading="securitySearch.loading.value"
        @close="closeDrawer"
        @create="onCreate"
        @update="onUpdate"
        @security-search="securitySearch.search"
      />
    </template>

    <AppConfirmDialog
      :open="deleteConfirmOpen"
      :title="t('watchlist.item.confirmDelete')"
      :description="t('watchlist.item.confirmDeleteDesc')"
      :confirm-label="t('watchlist.item.deleteItem')"
      :cancel-label="t('watchlist.drawer.cancel')"
      tone="danger"
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="deleteConfirmOpen = false"
    />
  </AppPage>
</template>
