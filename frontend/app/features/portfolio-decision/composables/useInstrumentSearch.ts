import { computed, ref, shallowRef } from "vue";

import {
  investmentLedgerApi,
  type ApiInstrument,
} from "../../investment-ledger/services/investmentLedgerApi";
import { extractMessage } from "../lib/decisionErrors";

const DEBOUNCE_MS = 280;

/**
 * Debounced, keyboard-navigable instrument search for the decision order
 * ticket's combobox.
 *
 * Two problems in the pre-existing investment-ledger equivalent
 * (features/investment-ledger/composables/useInstrumentDirectory.ts) are
 * fixed here rather than inherited:
 *   - stale responses: a slow early request could overwrite the results of
 *     a faster later one. Every search() call stamps a sequence number and
 *     only the highest-seen sequence is allowed to commit results.
 *   - keyboard navigation: this composable owns `activeIndex` so the
 *     combobox component can implement Arrow/Enter/Escape without local
 *     index bookkeeping of its own.
 *
 * The backend `search` filter only matches `primary_ticker` and `name`
 * (backend/internal/investment/infrastructure/persistence/instrument_repository.go) —
 * there is no ISIN field anywhere in InstrumentResponse or the query, so this
 * composable does not attempt ISIN matching.
 */
export function useInstrumentSearch() {
  const query = ref("");
  const items = shallowRef<ApiInstrument[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const activeIndex = ref(-1);
  const isOpen = ref(false);

  let debounceHandle: ReturnType<typeof setTimeout> | null = null;
  let requestSeq = 0;
  let latestAppliedSeq = 0;

  async function runSearch(term: string) {
    const seq = ++requestSeq;
    loading.value = true;
    error.value = null;
    try {
      const result = await investmentLedgerApi.listInstruments({
        search: term || undefined,
        limit: 20,
      });
      // A newer request may have started (and even finished) while this one
      // was in flight — never let an older response clobber newer results.
      if (seq < latestAppliedSeq) return;
      latestAppliedSeq = seq;
      items.value = result.items ?? [];
      activeIndex.value = items.value.length > 0 ? 0 : -1;
    } catch (err) {
      if (seq < latestAppliedSeq) return;
      latestAppliedSeq = seq;
      error.value = extractMessage(err, "Failed to search instruments.");
      items.value = [];
      activeIndex.value = -1;
    } finally {
      if (seq === requestSeq) loading.value = false;
    }
  }

  function setQuery(value: string) {
    query.value = value;
    isOpen.value = true;
    if (debounceHandle) clearTimeout(debounceHandle);
    debounceHandle = setTimeout(() => {
      void runSearch(value.trim());
    }, DEBOUNCE_MS);
  }

  function open() {
    isOpen.value = true;
    if (items.value.length === 0 && !loading.value) {
      void runSearch(query.value.trim());
    }
  }

  function close() {
    isOpen.value = false;
    activeIndex.value = -1;
  }

  function moveNext() {
    if (items.value.length === 0) return;
    activeIndex.value = (activeIndex.value + 1) % items.value.length;
  }

  function movePrev() {
    if (items.value.length === 0) return;
    activeIndex.value =
      activeIndex.value <= 0 ? items.value.length - 1 : activeIndex.value - 1;
  }

  function reset() {
    if (debounceHandle) clearTimeout(debounceHandle);
    query.value = "";
    items.value = [];
    error.value = null;
    loading.value = false;
    activeIndex.value = -1;
    isOpen.value = false;
  }

  const activeItem = computed<ApiInstrument | null>(
    () => items.value[activeIndex.value] ?? null,
  );

  return {
    query,
    items,
    loading,
    error,
    isOpen,
    activeIndex,
    activeItem,
    setQuery,
    open,
    close,
    moveNext,
    movePrev,
    reset,
  };
}
