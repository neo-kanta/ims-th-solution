/**
 * Market Data — catalog composable.
 *
 * Wraps the typed marketDataCatalogApi service with loading/error state for
 * use inside Vue components. Drives the Securities and Unmapped tabs of the
 * Market Data page and the inline editor on the security detail page.
 */
import {
  marketDataCatalogApi,
  type ApiSecurity,
  type ApiUnmappedCandidate,
  type ApiProviderMapping,
  type ApiCreateSecurityRequest,
  type ApiUpdateSecurityRequest,
  type ApiAddMappingRequest,
  type SecuritySearchParams,
  type UnmappedCandidateParams,
} from "../services/marketDataCatalogApi";

const candidates = ref<ApiUnmappedCandidate[]>([]);
const candidatesLoading = ref(false);
const candidatesError = ref<string | null>(null);

export function useMarketDataCatalog() {
  const securities = ref<ApiSecurity[]>([]);
  const securitiesLoading = ref(false);
  const securitiesError = ref<string | null>(null);

  const detail = ref<ApiSecurity | null>(null);
  const detailLoading = ref(false);
  const detailError = ref<string | null>(null);

  function clearDetail() {
    detail.value = null;
    detailError.value = null;
  }

  async function searchSecurities(params: SecuritySearchParams = {}) {
    securitiesLoading.value = true;
    securitiesError.value = null;
    try {
      const resp = await marketDataCatalogApi.searchSecurities(params);
      securities.value = resp.items ?? [];
    } catch (err) {
      securitiesError.value = err instanceof Error ? err.message : "Failed to load securities";
      securities.value = [];
    } finally {
      securitiesLoading.value = false;
    }
  }

  async function loadSecurity(securityId: string) {
    detailLoading.value = true;
    detailError.value = null;
    try {
      detail.value = await marketDataCatalogApi.getSecurity(securityId);
    } catch (err) {
      detailError.value = err instanceof Error ? err.message : "Failed to load security";
      detail.value = null;
    } finally {
      detailLoading.value = false;
    }
  }

  async function createSecurity(req: ApiCreateSecurityRequest): Promise<ApiSecurity> {
    return marketDataCatalogApi.createSecurity(req);
  }

  async function updateSecurity(securityId: string, req: ApiUpdateSecurityRequest): Promise<ApiSecurity> {
    const updated = await marketDataCatalogApi.updateSecurity(securityId, req);
    if (detail.value && detail.value.security_id === securityId) {
      detail.value = updated;
    }
    return updated;
  }

  async function addMapping(securityId: string, req: ApiAddMappingRequest): Promise<ApiProviderMapping> {
    const created = await marketDataCatalogApi.addMapping(securityId, req);
    if (detail.value && detail.value.security_id === securityId) {
      // Optimistically merge into detail.
      const merged = [...(detail.value.provider_mappings ?? [])];
      const idx = merged.findIndex((m) => m.mapping_id === created.mapping_id);
      if (idx >= 0) merged[idx] = created;
      else merged.push(created);
      detail.value = { ...detail.value, provider_mappings: merged };
    }
    return created;
  }

  async function deleteMapping(securityId: string, mappingId: string) {
    await marketDataCatalogApi.deleteMapping(securityId, mappingId);
    if (detail.value && detail.value.security_id === securityId) {
      const merged = (detail.value.provider_mappings ?? []).map((m) => (
        m.mapping_id === mappingId ? { ...m, mapping_status: "INACTIVE" } : m
      ));
      detail.value = { ...detail.value, provider_mappings: merged };
    }
  }

  async function refreshCandidates(params: UnmappedCandidateParams = {}) {
    candidatesLoading.value = true;
    candidatesError.value = null;
    try {
      const resp = await marketDataCatalogApi.listUnmappedCandidates(params);
      candidates.value = resp.items ?? [];
    } catch (err) {
      candidatesError.value = err instanceof Error ? err.message : "Failed to load candidates";
      candidates.value = [];
    } finally {
      candidatesLoading.value = false;
    }
  }

  async function mapCandidate(candidateId: string, securityId: string) {
    await marketDataCatalogApi.mapCandidate(candidateId, securityId);
    candidates.value = candidates.value.filter((c) => c.candidate_id !== candidateId);
  }

  async function rejectCandidate(candidateId: string, reason: string) {
    await marketDataCatalogApi.rejectCandidate(candidateId, reason);
    candidates.value = candidates.value.filter((c) => c.candidate_id !== candidateId);
  }

  return {
    // state
    securities,
    securitiesLoading,
    securitiesError,

    detail,
    detailLoading,
    detailError,

    candidates,
    candidatesLoading,
    candidatesError,

    // actions
    searchSecurities,
    loadSecurity,
    clearDetail,
    createSecurity,
    updateSecurity,
    addMapping,
    deleteMapping,
    refreshCandidates,
    mapCandidate,
    rejectCandidate,
  };
}
