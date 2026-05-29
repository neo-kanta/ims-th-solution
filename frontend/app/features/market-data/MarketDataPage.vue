<script setup lang="ts">
import { computed, ref, watch, reactive, onMounted } from "vue";
import { useRoute, useRouter } from "#imports";
import AppIcon from "~/shared/ui/AppIcon.vue";
import MarketDataProviderCard from "./components/MarketDataProviderCard.vue";
import MarketDataSearchPanel from "./components/MarketDataSearchPanel.vue";
import MarketDataTable from "./components/MarketDataTable.vue";
import MarketDataSyncActivity from "./components/MarketDataSyncActivity.vue";
import MarketDataUploadPanel from "./components/MarketDataUploadPanel.vue";
import MarketDataBatchImportPanel from "./components/MarketDataBatchImportPanel.vue";
import MarketDataCreateSecurityForm from "./components/MarketDataCreateSecurityForm.vue";
import MarketDataUnmappedPanel from "./components/MarketDataUnmappedPanel.vue";
import { useMarketData } from "./composables/useMarketData";
import { useMarketDataCatalog } from "./composables/useMarketDataCatalog";
import type { ApiCreateSecurityRequest } from "./services/marketDataCatalogApi";
import type {
  WatchlistFilter,
  SearchFilter,
  BondUploadMetadata,
} from "./market-data.types";
import { useI18n } from "~/composables/useI18n";

type Tab = "market" | "watchlist" | "unmapped" | "settings";

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const activeTab = ref<Tab>(
  typeof route.query.tab === "string" &&
    ["market", "watchlist", "unmapped", "settings"].includes(route.query.tab)
    ? (route.query.tab as Tab)
    : "market",
);

watch(activeTab, (tab) => {
  if (route.query.tab !== tab) {
    void router.replace({ query: { ...route.query, tab } });
  }
});

watch(
  () => route.query.tab,
  (newTab) => {
    if (typeof newTab === "string" && ["market", "watchlist", "unmapped", "settings"].includes(newTab)) {
      activeTab.value = newTab as Tab;
    } else if (!newTab) {
      activeTab.value = "market";
    }
  },
);

// ---- live market data (Watchlist tab) -------------------------------------

const {
  providers,
  providersError,
  watchlist,
  watchlistLoading,
  watchlistError,
  searchResults,
  searchLoading,
  searchError,
  syncActivity,
  isSyncing,
  init,
  refreshWatchlist,
  search,
  addSymbol,
  syncAll,
} = useMarketData();

const searchQuery = ref("");
const searchFilter = ref<SearchFilter>("all");
const watchlistFilter = ref<WatchlistFilter>("all");
const watchlistQuery = ref("");
const bondUpload = ref<BondUploadMetadata | null>(null);

let searchDebounce: ReturnType<typeof setTimeout> | null = null;
watch(searchQuery, (q) => {
  if (searchDebounce) clearTimeout(searchDebounce);
  searchDebounce = setTimeout(() => {
    void search(q);
  }, 280);
});

const filteredSearch = computed(() => {
  return searchResults.value.filter((r) => {
    switch (searchFilter.value) {
      case "stocks":
        return r.assetClass === "Equity" || r.assetClass === "NVDR";
      case "bonds":
        return r.assetClass === "Bond";
      case "etfs":
        return r.assetClass === "ETF";
      case "fx":
        return r.assetClass === "FX";
      default:
        return true;
    }
  });
});

function togglePin(id: string) {
  const item = watchlist.value.find((w) => w.id === id);
  if (item) item.pinned = !item.pinned;
}

function onAddFromSearch(id: string) {
  const r = searchResults.value.find((s) => s.id === id);
  if (r) {
    addSymbol(r.symbol);
    r.watching = true;
  }
}

function onOpenFromSearch(id: string) {
  const r = searchResults.value.find((s) => s.id === id);
  const target = r?.id ?? id;
  void router.push(`/market-data/securities/${encodeURIComponent(target)}`);
}

function onOpenFromWatchlist(id: string) {
  void router.push(`/market-data/securities/${encodeURIComponent(id)}`);
}

function onBrowse() {
  bondUpload.value = bondUpload.value;
}

function onUpload(_files: FileList) {
  // Backend endpoint /market-data/bonds/upload not yet implemented.
}

// ---- securities catalog (Securities tab) ---------------------------------

const {
  securities,
  securitiesLoading,
  securitiesError,
  candidates,
  candidatesLoading,
  candidatesError,
  searchSecurities,
  createSecurity,
  refreshCandidates,
  mapCandidate,
  rejectCandidate,
} = useMarketDataCatalog();

const catalogQuery = ref("");
const catalogAssetType = ref<string>("");
const catalogStatus = ref<string>("ACTIVE");
const showCreateForm = ref(false);

let catalogDebounce: ReturnType<typeof setTimeout> | null = null;
watch([catalogQuery, catalogAssetType, catalogStatus, activeTab], () => {
  if (activeTab.value !== "market") return;
  if (catalogDebounce) clearTimeout(catalogDebounce);
  catalogDebounce = setTimeout(() => {
    void refreshCatalog();
  }, 240);
});

async function refreshCatalog() {
  await searchSecurities({
    query: catalogQuery.value || undefined,
    assetType: catalogAssetType.value || undefined,
    status: catalogStatus.value || undefined,
    limit: 200,
  });
}

async function onCreate(req: ApiCreateSecurityRequest) {
  await createSecurity(req);
  showCreateForm.value = false;
  await refreshCatalog();
}

// ---- unmapped candidates (Unmapped tab) ----------------------------------

async function refreshUnmapped() {
  await refreshCandidates({ status: "REVIEW_REQUIRED", limit: 100 });
}

async function onMap(candidateId: string, securityId: string) {
  await mapCandidate(candidateId, securityId);
  await refreshUnmapped();
}

async function onReject(candidateId: string, reason: string) {
  await rejectCandidate(candidateId, reason);
  await refreshUnmapped();
}

// ---- lazy loading per tab ------------------------------------------------

const alphaVantageKey = ref("");
const alphaVantageUrl = ref("https://www.alphavantage.co");
const alphaVantageStatus = ref("ACTIVE");

const yahooFinanceUrl = ref("https://query1.finance.yahoo.com");
const yahooFinanceStatus = ref("ACTIVE");

const showSaveToast = ref(false);
const toastMessage = ref("");
const toastTone = ref<"success" | "danger" | "info">("success");

const loaded = reactive<Record<Tab, boolean>>({
  market: false,
  watchlist: false,
  unmapped: false,
  settings: false,
});

async function loadTab(tab: Tab) {
  if (loaded[tab]) return;
  if (tab === "watchlist") {
    await init();
  } else if (tab === "market") {
    await refreshCatalog();
  } else if (tab === "unmapped") {
    await refreshUnmapped();
  }
  loaded[tab] = true;
}

watch(
  activeTab,
  (tab) => {
    void loadTab(tab);
  },
  { immediate: true },
);

onMounted(() => {
  void loadTab(activeTab.value);

  // Load from localStorage if present
  const avKey = localStorage.getItem("provider_av_key");
  if (avKey) alphaVantageKey.value = avKey;

  const avUrl = localStorage.getItem("provider_av_url");
  if (avUrl) alphaVantageUrl.value = avUrl;

  const avStatus = localStorage.getItem("provider_av_status");
  if (avStatus) alphaVantageStatus.value = avStatus;

  const yfUrl = localStorage.getItem("provider_yf_url");
  if (yfUrl) yahooFinanceUrl.value = yfUrl;

  const yfStatus = localStorage.getItem("provider_yf_status");
  if (yfStatus) yahooFinanceStatus.value = yfStatus;
});

function saveSettings() {
  localStorage.setItem("provider_av_key", alphaVantageKey.value);
  localStorage.setItem("provider_av_url", alphaVantageUrl.value);
  localStorage.setItem("provider_av_status", alphaVantageStatus.value);
  localStorage.setItem("provider_yf_url", yahooFinanceUrl.value);
  localStorage.setItem("provider_yf_status", yahooFinanceStatus.value);

  // Trigger toast
  toastMessage.value = t("marketData.messages.saveSuccess");
  toastTone.value = "success";
  showSaveToast.value = true;
}
</script>

<template>
  <div class="md-page">
    <!-- ----------------- MARKET CATALOG TAB ----------------- -->
    <section v-if="activeTab === 'market'" class="md-tab-panel">
      <div v-if="showCreateForm" class="card md-catalog-card">
        <div class="card-header">
          <span class="card-title">{{ t("marketData.labels.createSecurity") }}</span>
        </div>
        <div class="card-body">
          <MarketDataCreateSecurityForm @submit="onCreate" />
        </div>
      </div>

      <div class="card md-filter-bar">
        <div class="card-body md-filter-bar__body">
          <div class="md-filter-group">
            <label class="md-filter-label">{{ t("marketData.labels.search") }}</label>
            <input
              v-model="catalogQuery"
              class="form-input"
              :placeholder="t('marketData.labels.searchPlaceholder')"
            />
          </div>
          <div class="md-filter-group">
              <label class="md-filter-label">{{ t("marketData.labels.assetType") }}</label>
            <select v-model="catalogAssetType" class="form-input">
              <option value="">{{ t("marketData.filters.any") }}</option>
              <option value="EQUITY">EQUITY</option>
              <option value="ETF">ETF</option>
              <option value="MUTUAL_FUND">MUTUAL_FUND</option>
              <option value="BOND">BOND</option>
              <option value="FX">FX</option>
              <option value="INDEX">INDEX</option>
              <option value="CASH">CASH</option>
              <option value="DERIVATIVE">DERIVATIVE</option>
              <option value="OTHER">OTHER</option>
              <option value="UNKNOWN">UNKNOWN</option>
            </select>
          </div>
          <div class="md-filter-group">
              <label class="md-filter-label">{{ t("marketData.labels.status") }}</label>
            <select v-model="catalogStatus" class="form-input">
              <option value="">{{ t("marketData.filters.any") }}</option>
              <option value="ACTIVE">ACTIVE</option>
              <option value="INACTIVE">INACTIVE</option>
              <option value="SUSPENDED">SUSPENDED</option>
            </select>
          </div>
          <div class="md-filter-group md-filter-group--action">
            <button
              class="btn btn-primary btn-sm md-btn-primary"
              type="button"
              @click="showCreateForm = true"
            >
              <AppIcon name="plus" size="xs" />
              <span>{{ t("marketData.actions.addSecurity") }}</span>
            </button>
          </div>
        </div>
      </div>

      <div v-if="securitiesError" class="alert alert-danger" role="status">
        {{ t("marketData.messages.failedLoadSecurities", { error: securitiesError }) }}
      </div>

      <MarketDataTable
        type="securities"
        :items="securities"
        :loading="securitiesLoading"
      />
    </section>

    <!-- ----------------- WATCHLIST TAB ----------------- -->
    <section v-if="activeTab === 'watchlist'" class="md-tab-panel">
      <div class="md-actions-row">
        <div class="md-action-item">
          <button
            class="btn btn-secondary btn-sm md-btn"
            :disabled="isSyncing"
            @click="syncAll"
          >
            <AppIcon name="refresh" size="xs" />
            <span>{{ isSyncing ? t("marketData.actions.syncing") : t("marketData.actions.syncAll") }}</span>
          </button>
          <span class="md-badge">{{ t("marketData.actions.live") }}</span>
        </div>

        <button class="btn btn-secondary btn-sm md-btn">
          <AppIcon name="shield" size="xs" />
          <span>{{ t("marketData.actions.apiKeys") }}</span>
        </button>
      </div>

      <div v-if="providersError" class="alert alert-warning" role="status">
        {{ t("marketData.messages.providerHealthError", { error: providersError }) }}
      </div>

      <div class="md-provider-grid">
        <MarketDataProviderCard
          v-for="p in providers"
          :key="p.id"
          :provider="p"
          @configure="activeTab = 'settings'"
        />
      </div>

      <MarketDataSearchPanel
        v-model:query="searchQuery"
        v-model:filter="searchFilter"
        :results="filteredSearch"
        :loading="searchLoading"
        :error="searchError"
        @add="onAddFromSearch"
        @open="onOpenFromSearch"
      />

      <div v-if="watchlistError" class="alert alert-danger" role="status">
        {{ t("marketData.messages.failedLoadWatchlist", { error: watchlistError }) }}
        <button class="md-link" @click="refreshWatchlist">{{ t("marketData.actions.retry") }}</button>
      </div>
      <MarketDataTable
        type="watchlist"
        v-model:filter="watchlistFilter"
        v-model:query="watchlistQuery"
        :items="watchlist"
        :loading="watchlistLoading"
        @toggle-pin="togglePin"
        @open="onOpenFromWatchlist"
        @sync="syncAll"
      />

      <MarketDataBatchImportPanel @completed="refreshWatchlist" />

      <div class="md-bottom-grid">
        <MarketDataSyncActivity :events="syncActivity" />
        <MarketDataUploadPanel
          :last-upload="bondUpload"
          @browse="onBrowse"
          @upload="onUpload"
        />
      </div>
    </section>

    <!-- ----------------- UNMAPPED / STAGE TAB ----------------- -->
    <section v-if="activeTab === 'unmapped'" class="md-tab-panel">
      <div v-if="candidatesError" class="alert alert-warning" role="status">
        {{ t("marketData.messages.failedLoadCandidates", { error: candidatesError }) }}
      </div>

      <MarketDataUnmappedPanel
        :items="candidates"
        :loading="candidatesLoading"
        :securities="securities"
        @map="onMap"
        @reject="onReject"
      />
    </section>

    <!-- ----------------- SETTINGS TAB ----------------- -->
    <section v-if="activeTab === 'settings'" class="md-tab-panel">
      <div class="card settings-card">
        <div class="card-header">
          <span class="card-title">{{ t("marketData.labels.providerConfig") }}</span>
        </div>
        <div class="card-body">
          <form @submit.prevent="saveSettings">
            <div class="settings-provider-section">
              <h3 class="provider-title">
                <span class="provider-initials av-bg">AV</span> Alpha Vantage
              </h3>
              <div class="form-grid">
                <div class="form-group">
                  <label class="form-label">{{ t("marketData.labels.apiKey") }}</label>
                  <input
                    v-model="alphaVantageKey"
                    type="password"
                    class="form-input"
                    :placeholder="t('marketData.placeholders.apiKey')"
                  />
                  <span class="form-help">{{ t("marketData.labels.apiKeyHelp") }}</span>
                </div>
                <div class="form-group">
                  <label class="form-label">{{ t("marketData.labels.baseUrl") }}</label>
                  <input
                    v-model="alphaVantageUrl"
                    type="text"
                    class="form-input"
                    placeholder="https://www.alphavantage.co"
                  />
                </div>
                <div class="form-group">
                  <label class="form-label">{{ t("marketData.labels.providerStatus") }}</label>
                  <select v-model="alphaVantageStatus" class="form-input">
                    <option value="ACTIVE">{{ t("marketData.labels.enabled") }}</option>
                    <option value="INACTIVE">{{ t("marketData.labels.disabled") }}</option>
                  </select>
                </div>
              </div>
            </div>

            <hr class="settings-divider" />

            <div class="settings-provider-section">
              <h3 class="provider-title">
                <span class="provider-initials yf-bg">YF</span> Yahoo Finance
              </h3>
              <div class="form-grid">
                <div class="form-group">
                  <label class="form-label">{{ t("marketData.labels.baseUrl") }}</label>
                  <input
                    v-model="yahooFinanceUrl"
                    type="text"
                    class="form-input"
                    placeholder="https://query1.finance.yahoo.com"
                  />
                </div>
                <div class="form-group">
                  <label class="form-label">{{ t("marketData.labels.providerStatus") }}</label>
                  <select v-model="yahooFinanceStatus" class="form-input">
                    <option value="ACTIVE">{{ t("marketData.labels.enabled") }}</option>
                    <option value="INACTIVE">{{ t("marketData.labels.disabled") }}</option>
                  </select>
                </div>
              </div>
            </div>

            <div class="settings-form-actions">
              <button class="btn btn-primary" type="submit">
                {{ t("marketData.actions.saveConfigurations") }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </section>

    <!-- Footer info banner — shared across tabs -->
    <div class="alert alert-info md-footer-banner" role="status">
      <svg
        width="14"
        height="14"
        viewBox="0 0 16 16"
        fill="currentColor"
        aria-hidden="true"
      >
        <path
          d="M8 1a7 7 0 100 14A7 7 0 008 1zm0 3a1 1 0 110 2 1 1 0 010-2zm1 9H7V7h2v6z"
        />
      </svg>
      <span>{{ t("marketData.messages.footerBanner") }}</span>
    </div>
  </div>
</template>

<style scoped>
.md-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

/* Custom Action Row (formerly GitHub Button Grouping) */
.md-actions-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-2);
}

.md-action-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.md-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-md);
  transition: all 0.15s ease;
  background: var(--bg-card-muted);
  color: var(--text-primary);
  border: 1px solid var(--border-default);
}

.md-btn:hover {
  background: var(--bg-row-hover);
}

.md-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-1) var(--space-3);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-bold);
  border-radius: var(--radius-pill);
  background: var(--bg-selected);
  color: var(--text-primary);
  border: 1px solid var(--border-default);
}

.md-btn-primary {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-md);
  background: var(--action-primary);
  color: #fff;
  border: 1px solid rgba(27, 31, 36, 0.15);
  box-shadow: 0 1px 0 rgba(27, 31, 35, 0.1);
  transition: background-color 0.15s ease;
}

.md-btn-primary:hover {
  background: var(--action-primary-hover);
}

.md-tab-panel {
  display: grid;
  gap: 1rem;
}

.md-catalog-card {
  background: var(--md-surface);
  border: 1px solid var(--md-border);
}

.md-filter-bar__body {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr auto;
  gap: 1rem;
  align-items: end;
}

.md-filter-group {
  display: grid;
  gap: 0.25rem;
}

.md-filter-label {
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--md-text-muted);
}

@media (max-width: 768px) {
  .md-filter-bar__body {
    grid-template-columns: 1fr;
  }
}

.settings-card {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
}

.settings-provider-section {
  padding: 1.5rem;
}

.provider-title {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  margin-bottom: 1.5rem;
  color: var(--text-primary);
}

.provider-initials {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 700;
  color: #fff;
}

.av-bg {
  background: #1f883d;
}

.yf-bg {
  background: #7e57c2;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.form-help {
  font-size: 11px;
  color: var(--text-secondary);
}

.settings-divider {
  border: 0;
  border-top: 1px solid var(--border-subtle);
  margin: 0;
}

.settings-form-actions {
  padding: 1.5rem;
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-card-muted);
  display: flex;
  justify-content: flex-end;
  border-bottom-left-radius: 8px;
  border-bottom-right-radius: 8px;
}
</style>
