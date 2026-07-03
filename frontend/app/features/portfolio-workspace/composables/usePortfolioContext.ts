/**
 * Loads a Portfolio V2 workspace's portfolio by business code
 * (docs/frontend/portfolio-v2-frontend-ddd.md section 6).
 *
 * `portfolioId` is internal state resolved from the loaded DTO for API
 * calls that still need it — it must never be used to build a
 * user-facing route. Every portfolio-scoped V2 route param is
 * `portfolioCode`.
 */
import { computed, ref } from "vue";

import { portfolioApi, type ApiPortfolioV2 } from "../services/portfolioApi";

function extractMessage(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const data = (err as { data?: { error?: unknown; message?: unknown } }).data;
  if (data && typeof data === "object") {
    if (typeof data.error === "string" && data.error.trim()) return data.error;
    if (typeof data.message === "string" && data.message.trim())
      return data.message;
  }
  const msg = (err as { message?: unknown }).message;
  if (typeof msg === "string" && msg.trim()) return msg;
  return fallback;
}

export function usePortfolioContext(portfolioCode: () => string) {
  const portfolio = ref<ApiPortfolioV2 | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const portfolioId = computed(() => portfolio.value?.id ?? null);
  const displayName = computed(
    () => portfolio.value?.name || portfolio.value?.code || portfolioCode(),
  );
  const portfolioType = computed(() => portfolio.value?.portfolio_type ?? "LIVE");
  const isLive = computed(() => portfolioType.value === "LIVE");
  const isSimulation = computed(() => portfolioType.value === "SIMULATION");
  const isModel = computed(() => portfolioType.value === "MODEL");

  async function reload() {
    const code = portfolioCode();
    if (!code) {
      portfolio.value = null;
      return;
    }
    loading.value = true;
    error.value = null;
    try {
      portfolio.value = await portfolioApi.getByCode(code);
    } catch (err) {
      portfolio.value = null;
      error.value = extractMessage(err, "Failed to load portfolio.");
    } finally {
      loading.value = false;
    }
  }

  return {
    portfolio,
    portfolioId,
    displayName,
    portfolioType,
    isLive,
    isSimulation,
    isModel,
    loading,
    error,
    reload,
  };
}
