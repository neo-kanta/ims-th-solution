import { ref, shallowRef } from "vue";

import {
  investmentLedgerApi,
  type ApiInstrument,
} from "../services/investmentLedgerApi";

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

/**
 * Searchable instrument lookup used by the Order Ticket. Pre-loads a default
 * page on demand and supports debounced search by ticker/name; the caller is
 * responsible for triggering search() — we don't watch a v-model here so the
 * drawer can decide on its own debounce strategy.
 */
export function useInstrumentDirectory() {
  const items = shallowRef<ApiInstrument[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const lastQuery = ref<string>("");

  let loadedOnce = false;

  async function ensureLoaded() {
    if (loadedOnce) return;
    await search("");
  }

  async function search(query: string) {
    loading.value = true;
    error.value = null;
    try {
      const trimmed = (query ?? "").trim();
      lastQuery.value = trimmed;
      const list = await investmentLedgerApi.listInstruments({
        search: trimmed || undefined,
        limit: 50,
      });
      items.value = list.items ?? [];
      loadedOnce = true;
    } catch (err) {
      error.value = extractMessage(err, "Failed to load instruments.");
      items.value = [];
    } finally {
      loading.value = false;
    }
  }

  function findById(id: string | null | undefined): ApiInstrument | null {
    if (!id) return null;
    return items.value.find((i) => i.id === id) ?? null;
  }

  function reset() {
    items.value = [];
    error.value = null;
    loadedOnce = false;
  }

  return {
    items,
    loading,
    error,
    lastQuery,
    ensureLoaded,
    search,
    findById,
    reset,
  };
}
