import { ref } from "vue";

import { portfolioApi, type ApiPortfolioV2, type ApiPatchPortfolioV2Request } from "../services/portfolioApi";
import {
  classifyPortfolioUpdateError,
  type ClassifiedPortfolioUpdateError,
} from "../lib/portfolioUpdateErrors";

export interface PortfolioSettingsFormValues {
  name: string;
  description: string;
  benchmark: string;
  risk_profile: string;
  strategy_code: string;
}

export function valuesFromPortfolio(p: ApiPortfolioV2 | null): PortfolioSettingsFormValues {
  return {
    name: p?.name ?? "",
    description: p?.description ?? "",
    benchmark: p?.benchmark ?? "",
    risk_profile: p?.risk_profile ?? "",
    strategy_code: p?.strategy_code ?? "",
  };
}

/**
 * Orchestrates the Portfolio Settings metadata-edit form: duplicate-
 * submission guard (mirrors `usePortfolioCreateForm`'s `busy` pattern),
 * optimistic `expected_version` concurrency (a 409 means the portfolio
 * changed since it was loaded — the caller must reload rather than retry
 * blindly, this composable never re-sends the same stale version), and
 * status-classified error surfacing. Never mutates fund association, code,
 * portfolio_type, or lifecycle status — those are not fields on this form.
 */
export function usePortfolioSettingsForm() {
  const values = ref<PortfolioSettingsFormValues>(valuesFromPortfolio(null));
  const busy = ref(false);
  const submitError = ref<ClassifiedPortfolioUpdateError | null>(null);
  const updated = ref<ApiPortfolioV2 | null>(null);

  function setValues(next: Partial<PortfolioSettingsFormValues>) {
    values.value = { ...values.value, ...next };
  }

  function hydrate(portfolio: ApiPortfolioV2 | null) {
    values.value = valuesFromPortfolio(portfolio);
    submitError.value = null;
    updated.value = null;
  }

  async function submit(
    portfolioCode: string,
    expectedVersion: number | undefined,
    fallbackErrorMessage: string,
  ): Promise<ApiPortfolioV2 | null> {
    if (busy.value) return null;
    if (!expectedVersion) {
      submitError.value = {
        kind: "unexpected",
        status: null,
        message: fallbackErrorMessage,
      };
      return null;
    }

    busy.value = true;
    submitError.value = null;
    try {
      const body: ApiPatchPortfolioV2Request = {
        expected_version: expectedVersion,
        name: values.value.name.trim() || undefined,
        description: values.value.description.trim() || undefined,
        benchmark: values.value.benchmark.trim() || undefined,
        risk_profile: values.value.risk_profile.trim() || undefined,
        strategy_code: values.value.strategy_code.trim() || undefined,
      };
      const portfolio = await portfolioApi.update(portfolioCode, body);
      updated.value = portfolio;
      return portfolio;
    } catch (err) {
      // Preserve entered values on a recoverable error — including a 409
      // version conflict, where the caller re-reads the fresh portfolio and
      // decides whether to reload or retry, rather than this composable
      // silently overwriting the user's edits.
      submitError.value = classifyPortfolioUpdateError(err, fallbackErrorMessage);
      return null;
    } finally {
      busy.value = false;
    }
  }

  return { values, busy, submitError, updated, setValues, hydrate, submit };
}
