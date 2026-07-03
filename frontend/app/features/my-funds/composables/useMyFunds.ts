/**
 * useMyFunds — orchestrates the My Funds list.
 *
 * Strategy: load the user's accessible funds (single round-trip), then in
 * parallel decorate each fund with its portfolio set, latest valuation,
 * live cash, workflow state, and open compliance breaches. Each decoration
 * is best-effort: a failure on one fund's workflow or breach call must
 * never empty the whole list.
 */
import { computed, ref } from "vue";

import { useAuthStore } from "~/stores/useAuthStore";

import {
  activeStageIndex,
  aggregateValuations,
  deriveRole,
  deriveStatus,
  emptyValuationSummary,
  mapWorkflowSummary,
  summarizeBreaches,
  todayBangkokIso,
  WORKFLOW_STAGES,
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
  MyFundsFilter,
  MyFundsSort,
} from "../types";

interface FundDecoration {
  portfolios: ApiPortfolio[];
  latestByPortfolio: Map<string, ApiValuation | null>;
  cashByPortfolio: Map<string, ApiCashBalance[]>;
  workflow: ApiWorkflowState | null;
  breaches: ApiBreach[];
}

export function useMyFunds() {
  const authStore = useAuthStore();
  const cards = ref<MyFundCard[]>([]);
  const loading = ref(false);
  const decorating = ref(false);
  const error = ref<string | null>(null);
  const businessDate = ref<string>(todayBangkokIso());
  const filter = ref<MyFundsFilter>("all");
  const sort = ref<MyFundsSort>("aum");
  const search = ref("");

  const filtered = computed<MyFundCard[]>(() => {
    const term = search.value.trim().toLowerCase();
    const base = cards.value.filter((card) => {
      if (term) {
        const haystack = `${card.code} ${card.short_name} ${card.name} ${card.fund_id}`.toLowerCase();
        if (!haystack.includes(term)) return false;
      }
      switch (filter.value) {
        case "managed":
          return card.role === "MANAGER";
        case "breached":
          return card.compliance.open_count > 0;
        case "locked":
          return card.status === "LOCKED" || card.status === "CLOSED";
        case "stale":
          return card.valuation.has_stale_inputs || !card.valuation.available;
        case "all":
        default:
          return true;
      }
    });

    return [...base].sort((a, b) => {
      switch (sort.value) {
        case "name":
          return (a.code || a.short_name).localeCompare(b.code || b.short_name);
        case "breach":
          return (b.compliance.open_count - a.compliance.open_count)
            || (b.valuation.aum_numeric - a.valuation.aum_numeric);
        case "updated":
          return (b.updated_at ?? "").localeCompare(a.updated_at ?? "");
        case "aum":
        default:
          return b.valuation.aum_numeric - a.valuation.aum_numeric;
      }
    });
  });

  const counts = computed(() => {
    const all = cards.value.length;
    const managed = cards.value.filter((c) => c.role === "MANAGER").length;
    const breached = cards.value.filter((c) => c.compliance.open_count > 0).length;
    const locked = cards.value.filter(
      (c) => c.status === "LOCKED" || c.status === "CLOSED",
    ).length;
    const stale = cards.value.filter(
      (c) => c.valuation.has_stale_inputs || !c.valuation.available,
    ).length;
    return { all, managed, breached, locked, stale };
  });

  async function decorateFund(fund: ApiFund): Promise<FundDecoration> {
    const fundId = fund.id;
    if (!fundId) {
      return {
        portfolios: [],
        latestByPortfolio: new Map(),
        cashByPortfolio: new Map(),
        workflow: null,
        breaches: [],
      };
    }

    const [portfolios, workflow, breaches] = await Promise.all([
      myFundsApi.listPortfoliosForFund(fundId).catch(() => [] as ApiPortfolio[]),
      myFundsApi.getWorkflowState(fundId, businessDate.value).catch(() => null),
      myFundsApi.listOpenBreaches(fundId).catch(() => [] as ApiBreach[]),
    ]);

    const latestByPortfolio = new Map<string, ApiValuation | null>();
    const cashByPortfolio = new Map<string, ApiCashBalance[]>();
    await Promise.all(
      portfolios.flatMap((portfolio) => {
        const id = portfolio.id;
        if (!id) return [];
        return [
          myFundsApi
            .getLatestValuation(id)
            .then((v) => latestByPortfolio.set(id, v))
            .catch(() => latestByPortfolio.set(id, null)),
          myFundsApi
            .listCash(id)
            .then((c) => cashByPortfolio.set(id, c))
            .catch(() => cashByPortfolio.set(id, [])),
        ];
      }),
    );

    return { portfolios, latestByPortfolio, cashByPortfolio, workflow, breaches };
  }

  function buildCard(fund: ApiFund, dec: FundDecoration): MyFundCard {
    const valuation = aggregateValuations(
      dec.portfolios,
      dec.latestByPortfolio,
      dec.cashByPortfolio,
    );
    const workflowSummary = mapWorkflowSummary(
      dec.workflow,
      fund.id ?? "",
      businessDate.value,
    );
    const compliance = summarizeBreaches(dec.breaches);
    const role = deriveRole(fund, authStore.user?.id ?? null);
    const status = deriveStatus(fund, workflowSummary, compliance);

    return {
      fund_id: fund.id ?? "",
      code: fund.code ?? "",
      short_name: fund.short_name ?? "",
      name: fund.name ?? "",
      base_currency: fund.base_currency ?? "",
      role,
      status,
      fund_status_raw: fund.status ?? "",
      manager_user_id: fund.manager_user_id ?? null,
      valuation,
      workflow: workflowSummary,
      compliance,
      updated_at: fund.updated_at ?? "",
    };
  }

  async function loadAll() {
    loading.value = true;
    error.value = null;
    try {
      const funds = await myFundsApi.listMyFunds();
      // Render skeleton cards immediately so the user sees structure, then
      // decorate with valuation/workflow/compliance in the background.
      cards.value = funds.map((f) => buildCard(f, {
        portfolios: [],
        latestByPortfolio: new Map(),
        cashByPortfolio: new Map(),
        workflow: null,
        breaches: [],
      }));

      decorating.value = true;
      const decorated = await Promise.all(
        funds.map(async (fund) => {
          try {
            const dec = await decorateFund(fund);
            return buildCard(fund, dec);
          } catch {
            return buildCard(fund, {
              portfolios: [],
              latestByPortfolio: new Map(),
              cashByPortfolio: new Map(),
              workflow: null,
              breaches: [],
            });
          }
        }),
      );
      cards.value = decorated;
    } catch (err) {
      cards.value = [];
      error.value = describeError(err, "Failed to load funds.");
    } finally {
      loading.value = false;
      decorating.value = false;
    }
  }

  function setFilter(next: MyFundsFilter) {
    filter.value = next;
  }
  function setSort(next: MyFundsSort) {
    sort.value = next;
  }
  function setSearch(next: string) {
    search.value = next;
  }
  function setBusinessDate(next: string) {
    if (next === businessDate.value) return;
    businessDate.value = next;
  }

  return {
    cards,
    filtered,
    counts,
    filter,
    sort,
    search,
    businessDate,
    loading,
    decorating,
    error,
    workflowStages: WORKFLOW_STAGES,
    activeStageIndex,
    loadAll,
    setFilter,
    setSort,
    setSearch,
    setBusinessDate,
  };
}

function describeError(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const msg = (err as { message?: unknown }).message;
  return typeof msg === "string" && msg.trim() ? msg : fallback;
}
