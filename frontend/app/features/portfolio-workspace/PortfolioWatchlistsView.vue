<script setup lang="ts">
/**
 * Portfolio-scoped Watchlists page. Reuses the existing user/portfolio
 * watchlist feature (frontend/app/features/watchlist) rather than
 * duplicating list/alert/drawer logic — see
 * frontend/app/features/portfolio-workspace/lib/watchlistScope.ts for the
 * pure query-locking helper this view delegates to. The existing feature
 * already supports scope_type=PORTFOLIO + portfolio_id (confirmed against
 * backend/internal/watchlist/domain/entity/watchlist_item.go's ScopePortfolio
 * and the generated ItemListQuery/AlertListQuery), so no backend change was
 * needed — only locking the existing free-choice UI to this workspace's
 * resolved portfolio.
 */
import { computed, onMounted, ref, watch } from "vue";
import { useState } from "#imports";

import AppCard from "~/shared/ui/AppCard.vue";
import AppConfirmDialog from "~/shared/ui/AppConfirmDialog.vue";
import { useI18n } from "~/composables/useI18n";
import { useAppToast } from "~/composables/useAppToast";
import { OpenApiRequestError } from "~/api/openapi";
import PortfolioWorkspaceHeader from "./components/PortfolioWorkspaceHeader.vue";
import { usePortfolioContext } from "./composables/usePortfolioContext";
import {
  buildLockedWatchlistAlertQuery,
  buildLockedWatchlistItemQuery,
} from "./lib/watchlistScope";

import WatchlistItemsTable from "~/features/watchlist/components/WatchlistItemsTable.vue";
import WatchlistAlertsPanel from "~/features/watchlist/components/WatchlistAlertsPanel.vue";
import WatchlistItemDrawer from "~/features/watchlist/components/WatchlistItemDrawer.vue";
import { useWatchlist } from "~/features/watchlist/composables/useWatchlist";
import { useWatchlistAlerts } from "~/features/watchlist/composables/useWatchlistAlerts";
import { useWatchlistSecuritySearch } from "~/features/watchlist/composables/useWatchlistSecuritySearch";
import type { WatchlistItem, CreateItemBody, UpdateItemBody } from "~/features/watchlist/types";

const props = defineProps<{ portfolioCode: string }>();
const { t } = useI18n();
const toast = useAppToast();

const pageTitle = useState<string>("page-title", () => "");
watch(
  () => t("portfolio.workspaceTabs.watchlists"),
  (newTitle) => {
    pageTitle.value = newTitle || "";
  },
  { immediate: true },
);

const ctx = usePortfolioContext(() => props.portfolioCode);
const watchlist = useWatchlist();
const alerts = useWatchlistAlerts();
const securitySearch = useWatchlistSecuritySearch();

const pageLoading = ref(true);
const pageError = ref<string | null>(null);

const drawerOpen = ref(false);
const drawerMode = ref<"create" | "edit">("create");
const editingItem = ref<WatchlistItem | null>(null);

const deleteConfirmOpen = ref(false);
const itemToDelete = ref<WatchlistItem | null>(null);
const deleting = ref(false);

const lockedPortfolioForDrawer = computed(() =>
  ctx.portfolio.value
    ? [ctx.portfolio.value]
    : [],
);

async function loadAll() {
  const itemQuery = buildLockedWatchlistItemQuery(ctx.portfolioId.value);
  const alertQuery = buildLockedWatchlistAlertQuery(ctx.portfolioId.value);
  if (!itemQuery || !alertQuery) return;
  pageLoading.value = true;
  pageError.value = null;
  try {
    await Promise.all([watchlist.load(itemQuery), alerts.load(alertQuery)]);
  } catch (err: unknown) {
    pageError.value =
      err instanceof Error ? err.message : t("portfolio.watchlistsPage.errorTitle");
  } finally {
    pageLoading.value = false;
  }
}

async function loadWorkspace() {
  await ctx.reload();
  if (!ctx.portfolio.value) return;
  await loadAll();
}

onMounted(() => {
  void loadWorkspace();
});
watch(
  () => props.portfolioCode,
  () => {
    void loadWorkspace();
  },
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
    return err instanceof Error ? err.message : t("portfolio.watchlistsPage.genericError");
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
    case "WATCHLIST_DUPLICATE_THRESHOLD":
      return t("watchlist.error.invalidThreshold");
    case "WATCHLIST_ITEM_NOT_FOUND":
      return t("watchlist.error.notFound");
    default:
      return err.message || t("portfolio.watchlistsPage.genericError");
  }
}

async function onCreate(body: CreateItemBody) {
  try {
    await watchlist.createItem(body);
    toast.showSuccess(t("portfolio.watchlistsPage.itemAdded"));
    closeDrawer();
    await loadAll();
  } catch (err: unknown) {
    toast.showError(mapErrorCode(err));
  }
}

async function onUpdate(id: string, body: UpdateItemBody) {
  try {
    await watchlist.updateItem(id, body);
    toast.showSuccess(t("portfolio.watchlistsPage.itemUpdated"));
    closeDrawer();
    await loadAll();
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
  if (!item?.id) return;
  deleting.value = true;
  try {
    await watchlist.deleteItem(item.id);
    toast.showSuccess(t("portfolio.watchlistsPage.itemRemoved"));
    deleteConfirmOpen.value = false;
    itemToDelete.value = null;
    await loadAll();
  } catch (err: unknown) {
    toast.showError(mapErrorCode(err));
  } finally {
    deleting.value = false;
  }
}

async function onAcknowledge(alertId: string, note: string | undefined) {
  try {
    await alerts.acknowledge(alertId, { note });
    toast.showSuccess(t("portfolio.watchlistsPage.alertAcknowledged"));
    const alertQuery = buildLockedWatchlistAlertQuery(ctx.portfolioId.value);
    if (alertQuery) await alerts.load(alertQuery);
  } catch (err: unknown) {
    toast.showError(mapErrorCode(err));
  }
}
</script>

<template>
  <section class="portfolio-watchlists">
    <PortfolioWorkspaceHeader :portfolio="ctx.portfolio.value" />

    <AppCard
      :title="t('portfolio.watchlistsPage.title')"
      :subtitle="t('portfolio.watchlistsPage.subtitle')"
    >
      <div v-if="pageLoading" class="portfolio-watchlists__notice" role="status">
        {{ t("portfolio.watchlistsPage.loading") }}
      </div>
      <div v-else-if="pageError" class="portfolio-watchlists__error" role="alert">
        <div>{{ t("portfolio.watchlistsPage.errorTitle") }}: {{ pageError }}</div>
        <button type="button" class="portfolio-watchlists__retry" @click="loadWorkspace">
          {{ t("portfolio.watchlistsPage.retry") }}
        </button>
      </div>
      <template v-else>
        <div class="portfolio-watchlists__banner">
          {{ t("portfolio.watchlistsPage.scopeBanner", { code: ctx.portfolio.value?.code ?? props.portfolioCode }) }}
        </div>

        <div class="portfolio-watchlists__actions">
          <IMSPermissionGuard permission="WATCHLIST_MANAGE">
            <button type="button" class="btn btn-primary btn-sm" @click="openCreate">
              + {{ t("portfolio.watchlistsPage.addItem") }}
            </button>
          </IMSPermissionGuard>
        </div>

        <div class="portfolio-watchlists__grid">
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
            <div class="portfolio-watchlists__alerts-title">
              {{ t("watchlist.alert.title") }}
            </div>
            <WatchlistAlertsPanel
              :alerts="alerts.alerts.value"
              :loading="alerts.loading.value"
              :error="alerts.error.value"
              :acknowledging="alerts.acknowledging.value"
              @acknowledge="onAcknowledge"
            />
          </div>
        </div>
      </template>
    </AppCard>

    <template v-if="drawerOpen">
      <WatchlistItemDrawer
        :mode="drawerMode"
        :item="editingItem"
        :portfolios="lockedPortfolioForDrawer"
        :lock-scope="true"
        :lock-portfolio-id="ctx.portfolioId.value"
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
  </section>
</template>

<style scoped>
.portfolio-watchlists {
  display: grid;
  gap: var(--space-4, 16px);
}

.portfolio-watchlists__notice,
.portfolio-watchlists__error {
  padding: var(--space-4, 16px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
}

.portfolio-watchlists__error {
  color: var(--alert-danger-text, #cf222e);
  display: grid;
  gap: 6px;
}

.portfolio-watchlists__retry {
  justify-self: start;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  border: 1px solid var(--border-subtle, #d0d7de);
  background: var(--bg-card, #ffffff);
  border-radius: var(--radius-md, 5px);
  padding: 4px 10px;
  cursor: pointer;
}

.portfolio-watchlists__banner {
  padding: 8px 12px;
  margin-bottom: 12px;
  font-size: 12px;
  color: var(--text-secondary, #57606a);
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
}

.portfolio-watchlists__actions {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 12px;
}

.portfolio-watchlists__grid {
  display: grid;
  grid-template-columns: 1fr 360px;
  gap: 20px;
  align-items: start;
}

.portfolio-watchlists__alerts-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
  color: var(--text-primary, #1e293b);
}

@media (max-width: 900px) {
  .portfolio-watchlists__grid {
    grid-template-columns: 1fr;
  }
}
</style>
