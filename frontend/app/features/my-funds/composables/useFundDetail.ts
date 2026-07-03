/**
 * useFundDetail — reactive controller for one fund's detail workspace.
 *
 * Loads the fund record, its portfolios, latest valuations, live cash,
 * workflow state for today's Thai business date, and open compliance
 * breaches. Each piece fails independently so a missing breach feed
 * never blocks the fund header from rendering.
 */
import { computed, ref } from "vue";

import { useAuthStore } from "~/stores/useAuthStore";

import {
  aggregateValuations,
  deriveRole,
  deriveStatus,
  mapWorkflowSummary,
  summarizeBreaches,
  todayBangkokIso,
} from "../lib/derive";
import { myFundsApi } from "../services/myFundsApi";
import type {
  ApiBreach,
  ApiCashBalance,
  ApiFund,
  ApiPortfolio,
  ApiValuation,
  ApiWorkflowState,
  MyFundCard,
} from "../types";

export function useFundDetail(fundIdGetter: () => string) {
  const authStore = useAuthStore();
  const fund = ref<ApiFund | null>(null);
  const portfolios = ref<ApiPortfolio[]>([]);
  const latestByPortfolio = ref(new Map<string, ApiValuation | null>());
  const cashByPortfolio = ref(new Map<string, ApiCashBalance[]>());
  const workflow = ref<ApiWorkflowState | null>(null);
  const breaches = ref<ApiBreach[]>([]);

  const loading = ref(false);
  const error = ref<string | null>(null);
  const businessDate = ref(todayBangkokIso());

  const card = computed<MyFundCard | null>(() => {
    if (!fund.value) return null;
    const valuation = aggregateValuations(
      portfolios.value,
      latestByPortfolio.value,
      cashByPortfolio.value,
    );
    const wf = mapWorkflowSummary(
      workflow.value,
      fund.value.id ?? "",
      businessDate.value,
    );
    const compliance = summarizeBreaches(breaches.value);
    const role = deriveRole(fund.value, authStore.user?.id ?? null);
    const status = deriveStatus(fund.value, wf, compliance);

    return {
      fund_id: fund.value.id ?? "",
      code: fund.value.code ?? "",
      short_name: fund.value.short_name ?? "",
      name: fund.value.name ?? "",
      base_currency: fund.value.base_currency ?? "",
      role,
      status,
      fund_status_raw: fund.value.status ?? "",
      manager_user_id: fund.value.manager_user_id ?? null,
      valuation,
      workflow: wf,
      compliance,
      updated_at: fund.value.updated_at ?? "",
    };
  });

  async function load() {
    const id = fundIdGetter();
    if (!id) return;
    loading.value = true;
    error.value = null;
    try {
      // The fund detail endpoint also enforces data-scope permission on the
      // server — a 403 will surface as an OpenApiRequestError.
      const fundResult = await fetchOptional(() => loadFund(id));
      fund.value = fundResult ?? null;
      if (!fund.value) {
        error.value = "Fund is not accessible or does not exist.";
        portfolios.value = [];
        latestByPortfolio.value = new Map();
        cashByPortfolio.value = new Map();
        workflow.value = null;
        breaches.value = [];
        return;
      }

      const [pf, wf, bx] = await Promise.all([
        myFundsApi.listPortfoliosForFund(id).catch(() => [] as ApiPortfolio[]),
        myFundsApi.getWorkflowState(id, businessDate.value).catch(() => null),
        myFundsApi.listOpenBreaches(id).catch(() => [] as ApiBreach[]),
      ]);
      portfolios.value = pf;
      workflow.value = wf;
      breaches.value = bx;

      const nextLatest = new Map<string, ApiValuation | null>();
      const nextCash = new Map<string, ApiCashBalance[]>();
      await Promise.all(
        pf.flatMap((p) => {
          const pid = p.id;
          if (!pid) return [];
          return [
            myFundsApi.getLatestValuation(pid)
              .then((v) => nextLatest.set(pid, v))
              .catch(() => nextLatest.set(pid, null)),
            myFundsApi.listCash(pid)
              .then((c) => nextCash.set(pid, c))
              .catch(() => nextCash.set(pid, [])),
          ];
        }),
      );
      latestByPortfolio.value = nextLatest;
      cashByPortfolio.value = nextCash;
    } catch (err) {
      error.value = describeError(err, "Failed to load fund detail.");
    } finally {
      loading.value = false;
    }
  }

  function setBusinessDate(next: string) {
    if (next === businessDate.value) return;
    businessDate.value = next;
  }

  return {
    fund,
    portfolios,
    breaches,
    workflow,
    businessDate,
    card,
    loading,
    error,
    load,
    setBusinessDate,
  };
}

async function loadFund(id: string): Promise<ApiFund | null> {
  const funds = await myFundsApi.listMyFunds(200);
  return funds.find((f) => f.id === id) ?? null;
}

async function fetchOptional<T>(loader: () => Promise<T>): Promise<T | null> {
  try {
    return await loader();
  } catch {
    return null;
  }
}

function describeError(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const msg = (err as { message?: unknown }).message;
  return typeof msg === "string" && msg.trim() ? msg : fallback;
}
