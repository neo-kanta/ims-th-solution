<script setup lang="ts">
import MarketDataProviderCard from "./components/MarketDataProviderCard.vue";
import MarketDataSearchPanel from "./components/MarketDataSearchPanel.vue";
import MarketDataWatchlistTable from "./components/MarketDataWatchlistTable.vue";
import MarketDataSyncActivity from "./components/MarketDataSyncActivity.vue";
import MarketDataUploadPanel from "./components/MarketDataUploadPanel.vue";
import MarketDataBatchImportPanel from "./components/MarketDataBatchImportPanel.vue";
import RDSecuritiesTable from "~/features/reference-data/components/RDSecuritiesTable.vue";
import RDCreateSecurityForm from "~/features/reference-data/components/RDCreateSecurityForm.vue";
import RDCandidatesPanel from "~/features/reference-data/components/RDCandidatesPanel.vue";
import { useMarketData } from "./composables/useMarketData";
import { useReferenceData } from "~/features/reference-data/composables/useReferenceData";
import type {
  ApiCreateSecurityRequest,
} from "~/features/reference-data/services/referenceDataApi";
import type {
  WatchlistFilter,
  SearchFilter,
  BondUploadMetadata,
} from "./market-data.types";

type Tab = "watchlist" | "securities" | "unmapped";

const tabs: { id: Tab; label: string; description: string }[] = [
  {
    id: "watchlist",
    label: "Watchlist",
    description: "Live quotes for everything you track. Imports from Yahoo Finance and Alpha Vantage.",
  },
  {
    id: "securities",
    label: "Securities",
    description: "Canonical securities you've imported. Search the catalog, add a new instrument, or open one to view its detail.",
  },
  {
    id: "unmapped",
    label: "Unmapped",
    description: "Provider symbols the importer couldn't resolve. Map them to an existing security or reject them.",
  },
];

const route = useRoute();
const router = useRouter();

const activeTab = ref<Tab>(
  (typeof route.query.tab === "string" && tabs.some((t) => t.id === route.query.tab))
    ? (route.query.tab as Tab)
    : "watchlist",
);

watch(activeTab, (tab) => {
  if (route.query.tab !== tab) {
    void router.replace({ query: { ...route.query, tab } });
  }
});

const activeDescription = computed(() =>
  tabs.find((t) => t.id === activeTab.value)?.description ?? "",
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
  searchDebounce = setTimeout(() => { void search(q); }, 280);
});

const filteredSearch = computed(() => {
  return searchResults.value.filter((r) => {
    switch (searchFilter.value) {
      case "stocks": return r.assetClass === "Equity" || r.assetClass === "NVDR";
      case "bonds":  return r.assetClass === "Bond";
      case "etfs":   return r.assetClass === "ETF";
      case "fx":     return r.assetClass === "FX";
      default:       return true;
    }
  });
});

function togglePin(id: string) {
  // Pinning is a UI-only flag until /watchlist persistence ships.
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
  /* TODO: open native file picker once the bond-upload endpoint exists. */
  bondUpload.value = bondUpload.value;
}

function onUpload(_files: FileList) {
  /* Backend endpoint /market-data/bonds/upload not yet implemented.
     We surface this to the user through the footer banner. */
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
} = useReferenceData();

const catalogQuery = ref("");
const catalogAssetType = ref<string>("");
const catalogStatus = ref<string>("ACTIVE");
const showCreateForm = ref(false);

let catalogDebounce: ReturnType<typeof setTimeout> | null = null;
watch([catalogQuery, catalogAssetType, catalogStatus, activeTab], () => {
  if (activeTab.value !== "securities") return;
  if (catalogDebounce) clearTimeout(catalogDebounce);
  catalogDebounce = setTimeout(() => { void refreshCatalog(); }, 240);
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
//
// Tabs only fetch their data the first time they're opened, so switching
// tabs after the initial Watchlist load doesn't fan out unnecessary calls.

const loaded = reactive<Record<Tab, boolean>>({
  watchlist: false,
  securities: false,
  unmapped: false,
});

async function loadTab(tab: Tab) {
  if (loaded[tab]) return;
  if (tab === "watchlist") {
    await init();
  } else if (tab === "securities") {
    await refreshCatalog();
  } else if (tab === "unmapped") {
    await refreshUnmapped();
  }
  loaded[tab] = true;
}

watch(activeTab, (tab) => { void loadTab(tab); }, { immediate: true });

onMounted(() => { void loadTab(activeTab.value); });
</script>

<template>
  <div class="md-page">
    <!-- Header -->
    <div class="page-header md-page-header">
      <div>
        <h1 class="page-title">Market data</h1>
        <p class="page-desc">{{ activeDescription }}</p>
      </div>
      <div v-if="activeTab === 'watchlist'" class="md-header-actions">
        <button class="btn btn-secondary btn-sm">API keys</button>
        <button
          class="btn btn-secondary btn-sm"
          :disabled="isSyncing"
          @click="syncAll"
        >
          {{ isSyncing ? "Syncing…" : "Sync all" }}
        </button>
        <button
          class="btn btn-primary btn-sm"
          type="button"
          @click="activeTab = 'securities'; showCreateForm = true"
        >
          <svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
            <path d="M8 3v10M3 8h10" />
          </svg>
          Add security
        </button>
      </div>
      <div v-else-if="activeTab === 'securities'" class="md-header-actions">
        <button class="btn btn-secondary btn-sm" @click="refreshCatalog">Refresh</button>
        <button
          class="btn btn-primary btn-sm"
          type="button"
          @click="showCreateForm = !showCreateForm"
        >
          {{ showCreateForm ? "Cancel" : "New security" }}
        </button>
      </div>
      <div v-else-if="activeTab === 'unmapped'" class="md-header-actions">
        <button class="btn btn-secondary btn-sm" @click="refreshUnmapped">Refresh</button>
      </div>
    </div>

    <!-- Tab bar -->
    <nav class="md-tabbar" role="tablist" aria-label="Market data sections">
      <button
        v-for="t in tabs"
        :key="t.id"
        type="button"
        role="tab"
        :aria-selected="activeTab === t.id"
        :tabindex="activeTab === t.id ? 0 : -1"
        class="md-tabbar__tab"
        :class="{ 'md-tabbar__tab--active': activeTab === t.id }"
        @click="activeTab = t.id"
      >
        <span>{{ t.label }}</span>
        <span
          v-if="t.id === 'unmapped' && candidates.length > 0"
          class="md-tabbar__count"
        >{{ candidates.length }}</span>
      </button>
    </nav>

    <!-- ----------------- WATCHLIST TAB ----------------- -->
    <section v-if="activeTab === 'watchlist'" class="md-tab-panel">
      <div v-if="providersError" class="alert alert-warning" role="status">
        Provider health endpoint unavailable: {{ providersError }}
      </div>

      <div class="md-provider-grid">
        <MarketDataProviderCard
          v-for="p in providers"
          :key="p.id"
          :provider="p"
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
        Failed to load watchlist: {{ watchlistError }}
        <button class="md-link" @click="refreshWatchlist">Retry</button>
      </div>
      <MarketDataWatchlistTable
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

    <!-- ----------------- SECURITIES TAB ----------------- -->
    <section v-if="activeTab === 'securities'" class="md-tab-panel">
      <div v-if="showCreateForm" class="card md-catalog-card">
        <div class="card-header">
          <span class="card-title">Create canonical security</span>
        </div>
        <div class="card-body">
          <RDCreateSecurityForm @submit="onCreate" />
        </div>
      </div>

      <div class="card md-filter-bar">
        <div class="card-body md-filter-bar__body">
          <div class="md-filter-group">
            <label class="md-filter-label">Search</label>
            <input
              v-model="catalogQuery"
              class="form-input"
              placeholder="ims_symbol, display_symbol, name, ISIN"
            />
          </div>
          <div class="md-filter-group">
            <label class="md-filter-label">Asset type</label>
            <select v-model="catalogAssetType" class="form-input">
              <option value="">Any</option>
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
            <label class="md-filter-label">Status</label>
            <select v-model="catalogStatus" class="form-input">
              <option value="">Any</option>
              <option value="ACTIVE">ACTIVE</option>
              <option value="INACTIVE">INACTIVE</option>
              <option value="SUSPENDED">SUSPENDED</option>
            </select>
          </div>
        </div>
      </div>

      <div v-if="securitiesError" class="alert alert-danger" role="status">
        Failed to load securities: {{ securitiesError }}
      </div>

      <RDSecuritiesTable
        :items="securities"
        :loading="securitiesLoading"
      />
    </section>

    <!-- ----------------- UNMAPPED TAB ----------------- -->
    <section v-if="activeTab === 'unmapped'" class="md-tab-panel">
      <div v-if="candidatesError" class="alert alert-warning" role="status">
        Failed to load candidates: {{ candidatesError }}
      </div>

      <RDCandidatesPanel
        :items="candidates"
        :loading="candidatesLoading"
        :securities="securities"
        @map="onMap"
        @reject="onReject"
      />
    </section>

    <!-- Footer info banner — shared across tabs -->
    <div class="alert alert-info md-footer-banner" role="status">
      <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
        <path d="M8 1a7 7 0 100 14A7 7 0 008 1zm0 3a1 1 0 110 2 1 1 0 010-2zm1 9H7V7h2v6z" />
      </svg>
      <span>
        Quotes, history, and chunk-based bulk import are live. Imports flow
        through Yahoo Finance and Alpha Vantage; unresolved symbols land in
        the Unmapped tab for review. ThaiBMA bond upload is UI only —
        backend endpoint pending.
      </span>
    </div>
  </div>
</template>

<style scoped>
.md-tabbar {
  display: flex;
  gap: 0.25rem;
  border-bottom: 1px solid var(--md-border);
  margin-bottom: 0.25rem;
}
.md-tabbar__tab {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 1rem;
  background: transparent;
  color: var(--md-text-muted);
  border: 0;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 600;
  transition: color 120ms ease, border-color 120ms ease;
}
.md-tabbar__tab:hover { color: var(--md-text); }
.md-tabbar__tab:focus-visible {
  outline: 2px solid var(--md-accent);
  outline-offset: -2px;
  border-radius: 2px;
}
.md-tabbar__tab--active {
  color: var(--md-text);
  border-bottom-color: var(--md-accent);
}
.md-tabbar__count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.5rem;
  height: 1.1rem;
  padding: 0 0.35rem;
  font-size: 0.7rem;
  font-weight: 700;
  border-radius: 999px;
  background: var(--md-warning-bg);
  color: var(--md-warning-text);
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
  grid-template-columns: 2fr 1fr 1fr;
  gap: 1rem;
  align-items: end;
}
.md-filter-group { display: grid; gap: 0.25rem; }
.md-filter-label {
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--md-text-muted);
}
@media (max-width: 768px) {
  .md-filter-bar__body { grid-template-columns: 1fr; }
}
</style>
