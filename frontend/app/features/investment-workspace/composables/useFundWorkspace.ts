import { ref } from "vue";

import { myFundsApi } from "~/features/my-funds";
import type { ApiFund } from "~/features/my-funds";

import type { FundCard, FundPrivacy } from "../types";

/**
 * Lists funds the caller can open and resolves the active fund by id.
 *
 * Sources every value from the live backend (`GET /investment/funds`). The
 * holdings workspace toolbar's "CONTRACT" dropdown is the only consumer of
 * the FundCard[] shape today; we adapt the backend FundResponse to that
 * shape here rather than reshaping the toolbar component.
 *
 * Fields that the backend does not expose yet (privacy, role, has_units) are
 * filled with safe defaults — never with mock values. The toolbar shows only
 * `contract_code` and `short_name`, so those defaults stay invisible.
 */
export function useFundWorkspace() {
  const funds = ref<FundCard[]>([]);
  const activeFund = ref<FundCard | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function loadFunds(activeId?: string) {
    loading.value = true;
    error.value = null;
    try {
      const apiFunds = await myFundsApi.listMyFunds(200);
      const list = apiFunds.map(toFundCard);
      funds.value = list;
      activeFund.value =
        (activeId && list.find((f) => f.fund_id === activeId)) || list[0] || null;
    } catch (err) {
      funds.value = [];
      activeFund.value = null;
      error.value = extractMessage(err, "Failed to load funds.");
    } finally {
      loading.value = false;
    }
  }

  function setActiveFund(id: string) {
    const next = funds.value.find((f) => f.fund_id === id);
    if (next) activeFund.value = next;
  }

  return { funds, activeFund, loading, error, loadFunds, setActiveFund };
}

function toFundCard(fund: ApiFund): FundCard {
  const id = fund.id ?? "";
  const code = fund.code ?? "";
  const shortName = fund.short_name ?? fund.name ?? code;
  return {
    fund_id: id,
    fund_uuid: id,
    code,
    short_name: shortName,
    // Backend does not yet expose a contract code on the Fund record.
    // The toolbar shows "<contract> · <short_name>" — fall back to the fund
    // code so the label still resolves to a real backend identifier.
    contract_code: code || shortName,
    privacy: "PRIVATE" satisfies FundPrivacy,
    // Per-fund role assignments are not yet exposed by the backend; the
    // toolbar does not read this field — picking the safest read-only
    // default so callers that *do* read it cannot accidentally enable
    // privileged UI.
    role: "RISK_VIEWER",
    base_currency: fund.base_currency ?? "",
    has_units: false,
    status: (fund.status as FundCard["status"]) ?? "ACTIVE",
  };
}

function extractMessage(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const msg = (err as { message?: unknown }).message;
  return typeof msg === "string" && msg.trim() ? msg : fallback;
}
