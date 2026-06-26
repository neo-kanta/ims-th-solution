import { ref } from "vue";
import type { WatchlistItem, ListItemsResult, ItemListQuery, CreateItemBody, UpdateItemBody } from "../types";
import { watchlistApi } from "../services/watchlistApi";

export function useWatchlist() {
  const items = ref<WatchlistItem[]>([]);
  const pagination = ref<ListItemsResult["pagination"]>(undefined);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const saving = ref(false);

  async function load(query: ItemListQuery = {}) {
    loading.value = true;
    error.value = null;
    try {
      const result = await watchlistApi.listItems(query);
      items.value = result.items ?? [];
      pagination.value = result.pagination;
    } catch (err: unknown) {
      error.value = err instanceof Error ? err.message : "Failed to load watchlist.";
      items.value = [];
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function createItem(body: CreateItemBody): Promise<WatchlistItem> {
    saving.value = true;
    try {
      return await watchlistApi.createItem(body);
    } finally {
      saving.value = false;
    }
  }

  async function updateItem(id: string, body: UpdateItemBody): Promise<WatchlistItem> {
    saving.value = true;
    try {
      return await watchlistApi.updateItem(id, body);
    } finally {
      saving.value = false;
    }
  }

  async function deleteItem(id: string): Promise<void> {
    saving.value = true;
    try {
      await watchlistApi.deleteItem(id);
    } finally {
      saving.value = false;
    }
  }

  return {
    items,
    pagination,
    loading,
    error,
    saving,
    load,
    createItem,
    updateItem,
    deleteItem,
  };
}
