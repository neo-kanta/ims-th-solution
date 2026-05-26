<script setup lang="ts">
import MarketDataSecurityHeader from "./components/MarketDataSecurityHeader.vue";
import MarketDataSecurityEmptyState from "./components/MarketDataSecurityEmptyState.vue";
import MarketDataSecurityIdentityCard from "./components/MarketDataSecurityIdentityCard.vue";
import MarketDataProviderMappingsCard from "./components/MarketDataProviderMappingsCard.vue";
import MarketDataLatestQuoteCard from "./components/MarketDataLatestQuoteCard.vue";
import MarketDataPriceHistoryPanel from "./components/MarketDataPriceHistoryPanel.vue";
import MarketDataSecurityQualityPanel from "./components/MarketDataSecurityQualityPanel.vue";
import MarketDataSecuritySyncActivity from "./components/MarketDataSecuritySyncActivity.vue";
import RDSecurityEditForm from "~/features/reference-data/components/RDSecurityEditForm.vue";
import RDMappingsPanel from "~/features/reference-data/components/RDMappingsPanel.vue";
import { useMarketDataSecurityDetail } from "./composables/useMarketDataSecurityDetail";
import { useReferenceData } from "~/features/reference-data/composables/useReferenceData";
import type {
  ApiSecurity,
  ApiUpdateSecurityRequest,
  ApiAddMappingRequest,
} from "~/features/reference-data/services/referenceDataApi";

const props = defineProps<{ securityId: string }>();

const securityId = toRef(props, "securityId");
const router = useRouter();

const {
  detail,
  loading,
  loadingQuote,
  loadingHistory,
  error,
  notFound,
  lastSync,
  isRunning,
  load,
  refresh,
  syncQuote,
  importQuoteAndHistory,
  toggleWatching,
} = useMarketDataSecurityDetail(securityId);

// Identity + mappings editor (was the /reference-data/securities/[id] page).
// We use the existing reference-data composable for writes; reads still come
// from the market-data security-detail composable above, so we just refresh
// the detail page after a successful write to pick up the latest state.
const {
  updateSecurity,
  addMapping,
  deleteMapping,
} = useReferenceData();

const showManage = ref(false);
const manageRef = ref<HTMLElement | null>(null);
const editorSecurity = computed<ApiSecurity | null>(() => {
  if (!detail.value) return null;
  return {
    security_id: detail.value.securityId,
    ims_symbol: detail.value.imsSymbol,
    display_symbol: detail.value.displaySymbol,
    name: detail.value.name,
    asset_type: detail.value.assetType,
    currency: detail.value.currency,
    country_code: detail.value.countryCode,
    exchange_mic: detail.value.exchangeMic,
    isin: detail.value.isin,
    cusip: detail.value.cusip,
    figi: detail.value.figi,
    status: detail.value.status,
    provider_mappings: detail.value.providerMappings.map((m) => ({
      mapping_id: m.mappingId,
      security_id: detail.value!.securityId,
      provider_code: m.providerCode,
      provider_symbol: m.providerSymbol,
      provider_exchange: m.providerExchange,
      provider_asset_type: m.providerAssetType,
      provider_currency: m.providerCurrency,
      priority: m.priority,
      mapping_status: m.status,
      is_primary: m.isPrimary,
      confidence_score: "100",
    })),
  };
});

const saving = ref(false);
const saveError = ref<string | null>(null);

async function onSaveIdentity(req: ApiUpdateSecurityRequest) {
  if (!detail.value?.securityId) return;
  saving.value = true;
  saveError.value = null;
  try {
    await updateSecurity(detail.value.securityId, req);
    await load();
  } catch (err) {
    saveError.value = err instanceof Error ? err.message : "Failed to save";
  } finally {
    saving.value = false;
  }
}

async function onAddMapping(req: ApiAddMappingRequest) {
  if (!detail.value?.securityId) return;
  await addMapping(detail.value.securityId, req);
  await load();
}

async function onDeleteMapping(mappingId: string) {
  if (!detail.value?.securityId) return;
  await deleteMapping(detail.value.securityId, mappingId);
  await load();
}

const showEmptyState = computed(() => {
  if (!detail.value) return false;
  return detail.value.dataQuality.freshness === "NOT_IMPORTED" && !detail.value.latestQuote;
});

const showUnmappedBanner = computed(() => {
  if (!detail.value) return false;
  if (detail.value.dataQuality.mapping === "UNMAPPED" || detail.value.dataQuality.mapping === "REVIEW_REQUIRED") {
    return true;
  }
  return Boolean(lastSync.value?.unmapped || lastSync.value?.reviewRequired);
});

const showRateLimitedBanner = computed(() => {
  if (!detail.value) return false;
  return detail.value.dataQuality.rateLimited || Boolean(lastSync.value?.rateLimited);
});

function onBack() {
  void router.push("/market-data");
}

function onOpenMappings() {
  showManage.value = true;
  // Wait for the section to render, then scroll it into view.
  void nextTick(() => {
    manageRef.value?.scrollIntoView({ behavior: "smooth", block: "start" });
  });
}

function onOpenUnmappedTab() {
  void router.push("/market-data?tab=unmapped");
}

async function importHistory30() {
  await importQuoteAndHistory();
}

async function importHistory250() {
  await importQuoteAndHistory();
}

onMounted(() => { void load(); });
</script>

<template>
  <section class="md-sec-detail">
    <div v-if="loading && !detail" class="md-sec-detail__loading" role="status">
      Loading security…
    </div>

    <div v-else-if="notFound" class="md-sec-detail__notfound" role="alert">
      <h1>Security not found</h1>
      <p>The security <code>{{ securityId }}</code> isn’t registered in the IMS catalog.</p>
      <button class="btn btn-primary" type="button" @click="onBack">Back to Market Data</button>
    </div>

    <div v-else-if="error && !detail" class="md-sec-detail__error" role="alert">
      <h1>Couldn’t load this security</h1>
      <p>{{ error }}</p>
      <button class="btn btn-primary" type="button" @click="refresh">Retry</button>
    </div>

    <template v-else-if="detail">
      <MarketDataSecurityHeader
        :detail="detail"
        :is-running="isRunning"
        @back="onBack"
        @sync-quote="syncQuote"
        @import-history="importQuoteAndHistory"
        @open-mappings="onOpenMappings"
        @toggle-watching="toggleWatching"
      />

      <div
        v-if="showUnmappedBanner"
        class="alert alert-warning md-sec-detail__banner"
        role="status"
      >
        <span>
          Provider symbol is not mapped. Configure a mapping below before
          retrying the import.
        </span>
        <div class="md-sec-detail__banner-actions">
          <button class="btn btn-secondary btn-sm" type="button" @click="onOpenMappings">
            Configure mappings
          </button>
          <button class="btn btn-secondary btn-sm" type="button" @click="onOpenUnmappedTab">
            View Unmapped queue
          </button>
        </div>
      </div>

      <div
        v-if="showRateLimitedBanner"
        class="alert alert-warning md-sec-detail__banner"
        role="status"
      >
        <span>
          Provider rate-limited the last request. Wait a minute and try Sync
          quote again.
        </span>
      </div>

      <MarketDataSecurityEmptyState
        v-if="showEmptyState"
        :mapping-status="detail.dataQuality.mapping"
        :is-running="isRunning"
        @sync-quote="syncQuote"
        @import-history="importQuoteAndHistory"
        @open-mappings="onOpenMappings"
      />

      <div v-if="error" class="alert alert-danger md-sec-detail__banner" role="status">
        {{ error }}
        <button class="btn btn-secondary btn-sm" type="button" @click="refresh">Retry</button>
      </div>

      <div class="md-sec-detail__top-row">
        <MarketDataLatestQuoteCard
          class="md-sec-detail__quote"
          :quote="detail.latestQuote"
          :loading="loadingQuote"
        />
        <MarketDataSecurityQualityPanel
          class="md-sec-detail__quality"
          :quality="detail.dataQuality"
        />
        <MarketDataProviderMappingsCard
          class="md-sec-detail__mappings"
          :mappings="detail.providerMappings"
          :security-id="detail.securityId"
          @manage="onOpenMappings"
          @open-unmapped="onOpenUnmappedTab"
        />
      </div>

      <MarketDataPriceHistoryPanel
        :history="detail.history"
        :loading="loadingHistory"
        :is-running="isRunning"
        @import-history-30="importHistory30"
        @import-history-250="importHistory250"
      />

      <!-- Inline editor (formerly /reference-data/securities/[id]) -->
      <section ref="manageRef" class="md-sec-manage card">
        <header class="card-header md-sec-manage__head">
          <div>
            <span class="card-title">Manage security</span>
            <div class="card-subtitle">
              Edit identity fields and provider symbol mappings.
            </div>
          </div>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            :aria-expanded="showManage"
            @click="showManage = !showManage"
          >
            {{ showManage ? "Hide" : "Edit" }}
          </button>
        </header>
        <div v-if="showManage" class="md-sec-manage__body">
          <div class="md-sec-manage__col">
            <h3 class="md-sec-manage__section-title">Identity</h3>
            <RDSecurityEditForm
              v-if="editorSecurity"
              :security="editorSecurity"
              :saving="saving"
              :error="saveError"
              @save="onSaveIdentity"
            />
          </div>
          <div class="md-sec-manage__col">
            <h3 class="md-sec-manage__section-title">Provider mappings</h3>
            <RDMappingsPanel
              :mappings="editorSecurity?.provider_mappings ?? []"
              @add="onAddMapping"
              @delete="onDeleteMapping"
            />
          </div>
        </div>
      </section>

      <div class="md-sec-detail__bottom-row">
        <MarketDataSecurityIdentityCard :detail="detail" />
        <MarketDataSecuritySyncActivity :events="detail.syncActivity" />
      </div>

      <p class="md-sec-detail__footer">
        Security detail uses canonical IMS symbols. Provider symbols (Yahoo,
        Alpha Vantage, …) resolve through the mappings configured above.
      </p>
    </template>
  </section>
</template>

<style scoped>
.md-sec-detail {
  display: grid;
  gap: 1rem;
  color: var(--md-text);
}
.md-sec-detail__loading,
.md-sec-detail__notfound,
.md-sec-detail__error {
  padding: 2rem;
  text-align: center;
  background: var(--md-surface);
  border: 1px solid var(--md-border);
  border-radius: 0.75rem;
  display: grid;
  justify-items: center;
  gap: 0.5rem;
}
.md-sec-detail__notfound h1,
.md-sec-detail__error h1 {
  margin: 0;
  font-size: 1.2rem;
}
.md-sec-detail__notfound p,
.md-sec-detail__error p {
  margin: 0;
  color: var(--md-text-muted);
}
.md-sec-detail__banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.md-sec-detail__banner-actions {
  display: flex;
  gap: 0.4rem;
  flex-wrap: wrap;
}
.md-sec-manage {
  background: var(--md-surface);
  border: 1px solid var(--md-border);
}
.md-sec-manage__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.5rem;
}
.md-sec-manage__body {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr);
  gap: 1rem;
  padding: 1rem;
}
.md-sec-manage__col {
  display: grid;
  gap: 0.5rem;
}
.md-sec-manage__section-title {
  margin: 0;
  font-size: 0.85rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--md-text-muted);
}
@media (max-width: 960px) {
  .md-sec-manage__body {
    grid-template-columns: 1fr;
  }
}
.md-sec-detail__top-row {
  display: grid;
  grid-template-columns: 1.2fr 1fr 1.2fr;
  gap: 1rem;
  align-items: stretch;
}
.md-sec-detail__top-row > * { min-width: 0; }
.md-sec-detail__bottom-row {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr);
  gap: 1rem;
}
.md-sec-detail__footer {
  margin: 0;
  font-size: 0.78rem;
  color: var(--md-text-muted);
  text-align: center;
  padding-top: 0.5rem;
}
@media (max-width: 1100px) {
  .md-sec-detail__top-row { grid-template-columns: 1fr 1fr; }
}
@media (max-width: 768px) {
  .md-sec-detail__top-row,
  .md-sec-detail__bottom-row { grid-template-columns: 1fr; }
}
</style>
