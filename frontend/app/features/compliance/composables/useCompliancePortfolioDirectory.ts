import { computed, type ComputedRef, type Ref } from "vue";
import { useState } from "#imports";

import { complianceApi } from "../services/complianceApi";
import type { CompliancePortfolioOption } from "../types";

interface DirectoryState {
  items: CompliancePortfolioOption[];
  loaded: boolean;
  loading: boolean;
  error: string | null;
}

function extractErrorMessage(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const data = (err as { data?: { error?: unknown; message?: unknown } }).data;
  if (data) {
    if (typeof data.error === "string" && data.error.trim()) return data.error;
    if (typeof data.message === "string" && data.message.trim()) return data.message;
  }
  const message = (err as { message?: unknown }).message;
  if (typeof message === "string" && message.trim()) return message;
  return fallback;
}

/**
 * Shared portfolio directory — fetched once per session and reused across
 * compliance screens (simulator, breach inbox, filters, etc.) so we never
 * show a raw UUID where a human-readable name is available.
 *
 * Backed by `useState` so SSR and CSR share the same payload and a second
 * subscriber doesn't re-fetch. `ensureLoaded()` is idempotent.
 */
export function useCompliancePortfolioDirectory(): {
  items: ComputedRef<CompliancePortfolioOption[]>;
  loading: ComputedRef<boolean>;
  loaded: ComputedRef<boolean>;
  error: ComputedRef<string | null>;
  ensureLoaded: () => Promise<void>;
  refresh: () => Promise<void>;
  byId: ComputedRef<Map<string, CompliancePortfolioOption>>;
  labelFor: (id: string | null | undefined) => string;
} {
  const state: Ref<DirectoryState> = useState<DirectoryState>(
    "compliance-portfolio-directory",
    () => ({ items: [], loaded: false, loading: false, error: null }),
  );

  async function load() {
    state.value = { ...state.value, loading: true, error: null };
    try {
      const payload = await complianceApi.listAllPortfolios();
      state.value = {
        items: payload.items ?? [],
        loaded: true,
        loading: false,
        error: null,
      };
    } catch (err) {
      state.value = {
        items: [],
        loaded: true, // mark loaded so we don't loop on errors
        loading: false,
        error: extractErrorMessage(err, "Failed to load portfolios."),
      };
    }
  }

  async function ensureLoaded() {
    if (state.value.loaded || state.value.loading) return;
    await load();
  }

  async function refresh() {
    await load();
  }

  const items = computed(() => state.value.items);
  const loading = computed(() => state.value.loading);
  const loaded = computed(() => state.value.loaded);
  const error = computed(() => state.value.error);

  const byId = computed(
    () => new Map(state.value.items.map((p) => [p.id, p])),
  );

  function labelFor(id: string | null | undefined): string {
    if (!id) return "—";
    const hit = byId.value.get(id);
    if (!hit) return id;
    const parts = [hit.code, hit.name].filter(Boolean);
    return parts.length > 0 ? parts.join(" — ") : id;
  }

  return {
    items,
    loading,
    loaded,
    error,
    ensureLoaded,
    refresh,
    byId,
    labelFor,
  };
}
