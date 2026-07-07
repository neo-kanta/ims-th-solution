import { computed, ref } from "vue";

import { useOpenApiClient, unwrapOpenApiResponse } from "~/api/openapi";
import { useAuthStore } from "~/stores/useAuthStore";
import { investmentLedgerApi } from "~/features/investment-ledger/services/investmentLedgerApi";
import { myFundsApi } from "~/features/my-funds/services/myFundsApi";
import {
  deriveRole,
  deriveStatus,
  summarizeBreaches,
  todayBangkokIso,
} from "~/features/my-funds/lib/derive";
import { parseDecimal, parseDecimalOrNull } from "~/features/my-funds/lib/format";

import type {
  ApiCashBalance,
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
  cashBalances: ApiCashBalance[];
  breaches: ApiBreach[];
}

export function useMyPortfolios() {
  const authStore = useAuthStore();
  const cards = ref<MyPortfolioCard[]>([]);
  const loading = ref(false);
  const decorating = ref(false);
  const error = ref<string | null>(null);
  const businessDate = ref<string>(todayBangkokIso());
  const filter = ref<MyPortfoliosFilter>("all");
  const sort = ref<MyPortfoliosSort>("aum");
  const search = ref("");

  const filtered = computed<MyPortfolioCard[]>(() => {
    const term = search.value.trim().toLowerCase();
    const base = cards.value.filter((card) => {
      if (term) {
        const haystack = `${card.code} ${card.name} ${card.portfolio_id} ${card.portfolio_type} ${card.base_currency}`.toLowerCase();
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
            b.valuation.aum_numeric - a.valuation.aum_numeric
          );
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
    const locked = cards.value.filter((c) => c.status === "CLOSED").length;
    const stale = cards.value.filter(
      (c) => c.valuation.has_stale_inputs || !c.valuation.available,
    ).length;
    return { all, managed, breached, locked, stale };
  });

  const kpis = computed<MyPortfoliosKpiStrip>(() => {
    let aum = 0;
    let unrealised = 0;
    let active = 0;
    let breachCount = 0;
    let staleCount = 0;
    let worstSeverity: MyPortfoliosKpiStrip["worst_breach_severity"] = null;
    const ccyTally: Record<string, number> = {};

    for (const card of cards.value) {
      aum += card.valuation.aum_numeric;
      unrealised += card.valuation.unrealised_pnl_numeric;
      breachCount += card.compliance.open_count;
      if (card.status === "ACTIVE" || card.status === "BREACH") {
        active += 1;
      }
      if (card.valuation.has_stale_inputs || !card.valuation.available) {
        staleCount += 1;
      }
      if (card.compliance.worst_severity) {
        if (
          card.compliance.worst_severity === "BLOCK" ||
          (card.compliance.worst_severity === "WARN" && worstSeverity !== "BLOCK") ||
          (card.compliance.worst_severity === "INFO" && worstSeverity === null)
        ) {
          worstSeverity = card.compliance.worst_severity;
        }
      }
      const ccy = card.valuation.valuation_ccy || card.base_currency || "";
      if (ccy) ccyTally[ccy] = (ccyTally[ccy] ?? 0) + 1;
    }

    const ccy = pickDominantCurrency(ccyTally) || cards.value[0]?.base_currency || "";
    const trend: MyPortfoliosKpiStrip["unrealised_pnl_trend"] =
      unrealised > 0 ? "up" : unrealised < 0 ? "down" : "flat";

    return {
      total_aum: aum.toString(),
      total_aum_numeric: aum,
      total_unrealised_pnl: unrealised.toString(),
      total_unrealised_pnl_numeric: unrealised,
      unrealised_pnl_trend: trend,
      active_count: active,
      total_count: cards.value.length,
      open_breach_count: breachCount,
      worst_breach_severity: worstSeverity,
      stale_count: staleCount,
      valuation_ccy: ccy,
    };
  });

  async function decoratePortfolio(portfolio: ApiPortfolio): Promise<PortfolioDecoration> {
    const portfolioId = portfolio.id;
    if (!portfolioId) {
      return { latestValuation: null, cashBalances: [], breaches: [] };
    }

    const client = useOpenApiClient();
    const [latestValuation, cashBalances, breaches] = await Promise.all([
      myFundsApi.getLatestValuation(portfolioId).catch(() => null),
      myFundsApi.listCash(portfolioId).catch(() => [] as ApiCashBalance[]),
      client
        .GET("/compliance/breaches", {
          params: { query: { portfolio_id: portfolioId, status: "OPEN", limit: 50 } },
        })
        .then((res) => {
          const body = unwrapOpenApiResponse<{ breaches?: ApiBreach[] }>(res);
          return body.breaches ?? [];
        })
        .catch(() => [] as ApiBreach[]),
    ]);

    return { latestValuation, cashBalances, breaches };
  }

  function buildCard(portfolio: ApiPortfolio, dec: PortfolioDecoration): MyPortfolioCard {
    const valuation = buildPortfolioValuation(portfolio, dec.latestValuation, dec.cashBalances);
    const compliance = summarizeBreaches(dec.breaches);
    const role = deriveRole(
      { manager_user_id: portfolio.manager_user_id } as any,
      authStore.user?.id ?? null,
    );

    // Mock workflow summary since portfolios do not contain workflow
    const mockWorkflow = {
      available: false,
      contract_id: "",
      business_date: "",
      current_state: "NOT_STARTED",
      allowed_actions: [],
      opened_at: null,
      manager_approved_at: null,
      transaction_closed_at: null,
      accounting_closed_at: null,
    };

    const status = deriveStatus(
      { status: portfolio.status } as any,
      mockWorkflow,
      compliance,
    ) as any;

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

  async function loadAll() {
    loading.value = true;
    error.value = null;
    try {
      const list = await investmentLedgerApi.listPortfolios({ limit: 200 });
      const items = list.items ?? [];

      // Render skeleton cards immediately
      cards.value = items.map((p) =>
        buildCard(p, { latestValuation: null, cashBalances: [], breaches: [] }),
      );

      decorating.value = true;
      const decorated = await Promise.all(
        items.map(async (p) => {
          try {
            const dec = await decoratePortfolio(p);
            return buildCard(p, dec);
          } catch {
            return buildCard(p, { latestValuation: null, cashBalances: [], breaches: [] });
          }
        }),
      );
      cards.value = decorated;
    } catch (err) {
      cards.value = [];
      error.value = describeError(err, "Failed to load portfolios.");
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
  balances: ApiCashBalance[],
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

  let liveCash = 0;
  let haveLiveCash = false;
  for (const b of balances) {
    const amount = parseDecimal(b.balance);
    if (Number.isFinite(amount)) {
      liveCash += amount;
      haveLiveCash = true;
    }
  }

  const effectiveCash = haveLiveCash ? liveCash : cash;
  const cashBufferPct = nav > 0 ? (effectiveCash / nav) * 100 : null;

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
    cash_balance: effectiveCash.toString(),
    cash_buffer_pct: cashBufferPct,
    valuation_ccy: valuationCcy,
    business_date: val.business_date ?? null,
    has_stale_inputs: !!val.has_stale_inputs,
    is_indicative: !!val.is_indicative,
  };
}

function pickDominantCurrency(tally: Record<string, number>): string {
  let best = "";
  let bestCount = -1;
  for (const [ccy, count] of Object.entries(tally)) {
    if (count > bestCount) {
      best = ccy;
      bestCount = count;
    }
  }
  return best;
}

function describeError(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const msg = (err as { message?: unknown }).message;
  return typeof msg === "string" && msg.trim() ? msg : fallback;
}
