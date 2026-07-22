import { ref } from "vue";

import { investmentLedgerApi } from "../../investment-ledger/services/investmentLedgerApi";

/**
 * Resolves a portfolio UUID (never shown to the user) to its business code
 * for display/routing. Module-level cache: every caller in this feature
 * shares one in-memory map, so opening several decision rows for the same
 * portfolio only calls the API once. Each lookup still goes through
 * `investmentLedgerApi.getPortfolio`, which enforces the caller's own data
 * scope — an id the caller cannot access simply fails to resolve.
 */
const codeCache = new Map<string, string>();
const pending = new Map<string, Promise<string | null>>();

export function usePortfolioCodeLookup() {
  const codesById = ref<Record<string, string>>({ ...Object.fromEntries(codeCache) });

  async function resolve(portfolioId: string | undefined | null): Promise<string | null> {
    if (!portfolioId) return null;
    const cached = codeCache.get(portfolioId);
    if (cached) {
      codesById.value = { ...codesById.value, [portfolioId]: cached };
      return cached;
    }

    let promise = pending.get(portfolioId);
    if (!promise) {
      promise = investmentLedgerApi
        .getPortfolio(portfolioId)
        .then((portfolio) => portfolio.code ?? null)
        .catch(() => null)
        .finally(() => {
          pending.delete(portfolioId);
        });
      pending.set(portfolioId, promise);
    }

    const code = await promise;
    if (code) {
      codeCache.set(portfolioId, code);
      codesById.value = { ...codesById.value, [portfolioId]: code };
    }
    return code;
  }

  async function resolveMany(
    portfolioIds: ReadonlyArray<string | undefined | null>,
  ): Promise<void> {
    const unique = Array.from(new Set(portfolioIds.filter((id): id is string => Boolean(id))));
    await Promise.all(unique.map((id) => resolve(id)));
  }

  return { codesById, resolve, resolveMany };
}
