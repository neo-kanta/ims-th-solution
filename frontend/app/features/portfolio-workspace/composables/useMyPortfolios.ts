import { computed, ref } from "vue";

import { useOpenApiClient, unwrapOpenApiResponse } from "~/api/openapi";
import { useAuthStore } from "~/stores/useAuthStore";
import { investmentLedgerApi } from "~/features/investment-ledger/services/investmentLedgerApi";
import { dashboardApi } from "~/features/dashboard/services/dashboardApi";
import { collectOffsetPages } from "~/features/compliance/lib/pagination";
import { myFundsApi } from "~/features/my-funds/services/myFundsApi";
import { summarizeBreaches, todayBangkokIso } from "~/features/my-funds/lib/derive";
import { parseDecimal, parseDecimalOrNull } from "~/features/my-funds/lib/format";
import { buildPortfolioDirectoryKpis } from "../lib/directoryKpis";
import type { ValuationSummaryDTO } from "~/features/dashboard/types";

import type {
  ApiPortfolio,
  ApiValuation,
  ApiBreach,
  MyPortfolioCard,
  MyPortfoliosFilter,
  MyPortfoliosSort,
  MyPortfoliosKpiStrip,
} from "../types";

interface PortfolioDecoration {
  latestValuation: ApiValuation | null;
  breaches: ApiBreach[];
  complianceAvailable: boolean;
}

export function useMyPortfolios() {
  const authStore = useAuthStore();
  const cards = ref<MyPortfolioCard[]>([]);
  const loading = ref(false);
  const decorating = ref(false);
  const error = ref<string | null>(null);
  const valuationSummary = ref<ValuationSummaryDTO | null>(null);
  const valuationSummaryError = ref(false);
  const businessDate = ref<string>(todayBangkokIso());
  const filter = ref<MyPortfoliosFilter>("all");
  const sort = ref<MyPortfoliosSort>("updated");
  const search = ref("");

  const filtered = computed<MyPortfolioCard[]>(() => {
    const term = search.value.trim().toLowerCase();
    const base = cards.value.filter((card) => {
      if (term) {
        const haystack = `${card.code} ${card.name} ${card.portfolio_type} ${card.base_currency}`.toLowerCase();
        if (!haystack.includes(term)) return false;
      }
      switch (filter.value) {
        case "managed":
          return card.role === "MANAGER";
        case "breached":
          return card.compliance.open_count > 0;
        case "locked":
          return card.status === "CLOSED";
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
          return (a.code || a.name).localeCompare(b.code || b.name);
        case "breach":
          return (
            b.compliance.open_count - a.compliance.open_count ||
            (b.updated_at ?? "").localeCompare(a.updated_at ?? "")
          );
        case "updated":
          return (b.updated_at ?? "").localeCompare(a.updated_at ?? "");
        default:
          return (a.code || a.name).localeCompare(b.code || b.name);
      }
    });
  });

  const counts = computed(() => {
    const all = cards.value.length;
    const managed = cards.value.filter((c) => c.role === "MANAGER").length;
    const breached = cards.value.filter((c) => c.compliance.open_count > 0).length;
    const locked = cards.value.filter((c) => c.status === "CLOSED").length;
    const stale = cards.value.filter(
      (c) => c.valuation.has_stale_inputs || !c.valuation.available,
    ).length;
    return { all, managed, breached, locked, stale };
  });

  const kpis = computed<MyPortfoliosKpiStrip>(() =>
    buildPortfolioDirectoryKpis(
      cards.value,
      valuationSummary.value,
      valuationSummaryError.value,
    ),
  );

  async function decoratePortfolio(portfolio: ApiPortfolio): Promise<PortfolioDecoration> {
    const portfolioId = portfolio.id;
    if (!portfolioId) {
      return {
        latestValuation: null,
        breaches: [],
        complianceAvailable: false,
      };
    }

    const client = useOpenApiClient();
    const [latestValuation, complianceResult] = await Promise.all([
      myFundsApi.getLatestValuation(portfolioId).catch(() => null),
      collectOffsetPages<ApiBreach>(async (offset, limit) => {
        const res = await client.GET("/compliance/breaches", {
          params: {
            query: {
              portfolio_id: portfolioId,
              status: "OPEN",
              offset,
              limit,
            },
          },
        });
        const body = unwrapOpenApiResponse<{
          breaches?: ApiBreach[];
          total?: number;
          offset?: number;
        }>(res);
        return {
          items: body.breaches ?? [],
          total: body.total,
          offset: body.offset,
        };
      })
        .then(({ items }) => ({ available: true as const, breaches: items }))
        .catch(() => ({ available: false as const, breaches: [] as ApiBreach[] })),
    ]);

    return {
      latestValuation,
      breaches: complianceResult.breaches,
      complianceAvailable: complianceResult.available,
    };
  }

  function buildCard(portfolio: ApiPortfolio, dec: PortfolioDecoration): MyPortfolioCard {
    const valuation = buildPortfolioValuation(portfolio, dec.latestValuation);
    const compliance = dec.complianceAvailable
      ? summarizeBreaches(dec.breaches)
      : {
          available: false,
          open_count: 0,
          warning_count: 0,
          worst_severity: null,
          latest_breach_id: null,
          latest_message: null,
        };
    const role =
      authStore.user?.id && portfolio.manager_user_id === authStore.user.id
        ? "MANAGER"
        : "MEMBER";
    const status =
      compliance.available && compliance.open_count > 0
        ? "BREACH"
        : (portfolio.status ?? "").toUpperCase() === "ACTIVE"
          ? "ACTIVE"
          : "CLOSED";

    return {
      portfolio_id: portfolio.id ?? "",
      code: portfolio.code ?? "",
      name: portfolio.name ?? "",
      base_currency: portfolio.base_currency ?? "",
      valuation_currency: portfolio.valuation_currency ?? "",
      portfolio_type: portfolio.portfolio_type ?? "",
      risk_profile: portfolio.risk_profile ?? "",
      role,
      status,
      portfolio_status_raw: portfolio.status ?? "",
      manager_user_id: portfolio.manager_user_id ?? null,
      valuation,
      compliance,
      updated_at: portfolio.updated_at ?? "",
      fund_id: portfolio.fund_id ?? null,
    };
  }

  async function loadAll(portfolioLoadErrorMessage: string) {
    loading.value = true;
    error.value = null;
    valuationSummaryError.value = false;
    try {
      const valuationSummaryRequest = dashboardApi
        .valuationSummary("company")
        .then((summary) => ({ ok: true as const, summary }))
        .catch(() => ({ ok: false as const, summary: null }));
      const list = await investmentLedgerApi.listPortfolios({ limit: 200 });
      const items = list.items ?? [];

      // Render skeleton cards immediately
      cards.value = items.map((p) =>
        buildCard(p, {
          latestValuation: null,
          breaches: [],
          complianceAvailable: false,
        }),
      );

      decorating.value = true;
      const [decorated, valuationResult] = await Promise.all([
        Promise.all(
          items.map(async (p) => {
            try {
              const dec = await decoratePortfolio(p);
              return buildCard(p, dec);
            } catch {
              return buildCard(p, {
                latestValuation: null,
                breaches: [],
                complianceAvailable: false,
              });
            }
          }),
        ),
        valuationSummaryRequest,
      ]);
      cards.value = decorated;
      valuationSummary.value = valuationResult.summary;
      valuationSummaryError.value = !valuationResult.ok;
      if (valuationResult.summary?.businessDate) {
        businessDate.value = valuationResult.summary.businessDate;
      }
    } catch (err) {
      cards.value = [];
      valuationSummary.value = null;
      error.value = describeError(err, portfolioLoadErrorMessage);
    } finally {
      loading.value = false;
      decorating.value = false;
    }
  }

  function setFilter(next: MyPortfoliosFilter) {
    filter.value = next;
  }
  function setSort(next: MyPortfoliosSort) {
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
    kpis,
    loadAll,
    setFilter,
    setSort,
    setSearch,
    setBusinessDate,
  };
}

function buildPortfolioValuation(
  portfolio: ApiPortfolio,
  val: ApiValuation | null,
): MyPortfolioCard["valuation"] {
  if (!val) {
    return {
      available: false,
      nav: "0",
      nav_numeric: 0,
      aum: "0",
      aum_numeric: 0,
      unrealised_pnl: "0",
      unrealised_pnl_numeric: 0,
      realised_pnl: "0",
      realised_pnl_numeric: 0,
      roi: null,
      cash_balance: "0",
      cash_buffer_pct: null,
      valuation_ccy: portfolio.valuation_currency || portfolio.base_currency || "",
      business_date: null,
      has_stale_inputs: false,
      is_indicative: false,
    };
  }

  const valuationCcy = val.valuation_ccy ?? portfolio.valuation_currency ?? portfolio.base_currency ?? "";
  const nav = parseDecimal(val.market_value);
  const aum = parseDecimal(val.aum);
  const unrealised = parseDecimal(val.unrealised_pnl);
  const realised = parseDecimal(val.realised_pnl);
  const cash = parseDecimal(val.cash_balance);
  const roi = parseDecimalOrNull(val.roi);

  // `ValuationResponse.cash_balance` is already normalized into the
  // valuation currency by the backend valuation snapshot. Never add raw cash
  // balances from different currencies in the client.
  const cashBufferPct = nav > 0 ? (cash / nav) * 100 : null;

  return {
    available: true,
    nav: val.market_value ?? "0",
    nav_numeric: nav,
    aum: val.aum ?? "0",
    aum_numeric: aum,
    unrealised_pnl: val.unrealised_pnl ?? "0",
    unrealised_pnl_numeric: unrealised,
    realised_pnl: val.realised_pnl ?? "0",
    realised_pnl_numeric: realised,
    roi: roi !== null ? roi.toString() : null,
    cash_balance: val.cash_balance ?? "0",
    cash_buffer_pct: cashBufferPct,
    valuation_ccy: valuationCcy,
    business_date: val.business_date ?? null,
    has_stale_inputs: !!val.has_stale_inputs,
    is_indicative: !!val.is_indicative,
  };
}

function describeError(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const msg = (err as { message?: unknown }).message;
  return typeof msg === "string" && msg.trim() ? msg : fallback;
}
