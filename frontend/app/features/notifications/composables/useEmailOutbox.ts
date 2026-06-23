import { ref } from "vue";
import { notificationApi, notificationErrorMessage } from "../services/notificationApi";
import type { OutboxItem, OutboxDetail, RetryResult, OutboxFilter } from "../types";

export function useEmailOutbox() {
  const items = ref<OutboxItem[]>([]);
  const total = ref(0);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const filter = ref<OutboxFilter>({ limit: 50, offset: 0 });

  async function load(overrides: OutboxFilter = {}) {
    loading.value = true;
    error.value = null;
    try {
      const merged = { ...filter.value, ...overrides };
      const res = await notificationApi.listOutbox(merged);
      items.value = res.items ?? [];
      total.value = res.total ?? 0;
    } catch (e) {
      error.value = notificationErrorMessage(e, "Unable to load email outbox");
    } finally {
      loading.value = false;
    }
  }

  function applyFilter(newFilter: OutboxFilter) {
    filter.value = { ...newFilter, limit: newFilter.limit ?? 50, offset: 0 };
    return load();
  }

  return { items, total, loading, error, filter, load, applyFilter };
}

export function useEmailOutboxDetail(outboxId: string) {
  const detail = ref<OutboxDetail | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const retrying = ref(false);
  const retryResult = ref<RetryResult | null>(null);
  const retryError = ref<string | null>(null);

  async function load() {
    loading.value = true;
    error.value = null;
    try {
      detail.value = await notificationApi.getOutboxDetail(outboxId);
    } catch (e) {
      error.value = notificationErrorMessage(e, "Unable to load outbox record");
    } finally {
      loading.value = false;
    }
  }

  async function retry() {
    retrying.value = true;
    retryError.value = null;
    try {
      retryResult.value = await notificationApi.retryOutbox(outboxId);
      await load();
    } catch (e) {
      retryError.value = notificationErrorMessage(e, "Unable to retry email delivery");
    } finally {
      retrying.value = false;
    }
  }

  return { detail, loading, error, retrying, retryResult, retryError, load, retry };
}
