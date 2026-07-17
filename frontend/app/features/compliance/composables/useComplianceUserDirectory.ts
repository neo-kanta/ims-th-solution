import { computed, type ComputedRef, type Ref } from "vue";
import { useState } from "#imports";

import { complianceApi } from "../services/complianceApi";

export interface ComplianceUserOption {
  id: string;
  display_name?: string;
  username?: string;
}

interface DirectoryState {
  items: ComplianceUserOption[];
  loaded: boolean;
  loading: boolean;
  forbidden: boolean;
  error: string | null;
}

function statusOf(err: unknown): number | null {
  if (!err || typeof err !== "object") return null;
  const status = (err as { status?: unknown; statusCode?: unknown }).status;
  if (typeof status === "number") return status;
  const statusCode = (err as { statusCode?: unknown }).statusCode;
  return typeof statusCode === "number" ? statusCode : null;
}

/**
 * Shared user directory — resolves owner / actor UUIDs to display names via
 * `GET /admin/users`. The endpoint requires `IAM_USER_VIEW`; for users that
 * lack it (most compliance-only roles) we mark the directory `forbidden` and
 * every unresolved lookup returns an empty label. This is intentional:
 *
 *   - We never crash on 403.
 *   - We never invent a name.
 *   - We never re-issue the call on every page render — `loaded` short-circuits.
 *
 * Total backend load: at most one call per browser session, scoped to a
 * single page of 200 users (covers the demo team comfortably).
 */
export function useComplianceUserDirectory(): {
  items: ComputedRef<ComplianceUserOption[]>;
  loading: ComputedRef<boolean>;
  loaded: ComputedRef<boolean>;
  forbidden: ComputedRef<boolean>;
  error: ComputedRef<string | null>;
  ensureLoaded: () => Promise<void>;
  refresh: () => Promise<void>;
  labelFor: (id: string | null | undefined) => string;
} {
  const state: Ref<DirectoryState> = useState<DirectoryState>(
    "compliance-user-directory",
    () => ({
      items: [],
      loaded: false,
      loading: false,
      forbidden: false,
      error: null,
    }),
  );

  async function load() {
    state.value = { ...state.value, loading: true, error: null };
    try {
      const payload = await complianceApi.listUsers({ limit: 200, offset: 0 });
      state.value = {
        items: payload.users ?? [],
        loaded: true,
        loading: false,
        forbidden: false,
        error: null,
      };
    } catch (err) {
      const status = statusOf(err);
      // 403 is the expected "compliance user can't read users" case. Treat
      // as a soft degrade: mark loaded + forbidden, return UUIDs forever.
      state.value = {
        items: [],
        loaded: true,
        loading: false,
        forbidden: status === 403,
        error:
          status === 403
            ? null
            : (err instanceof Error ? err.message : "Failed to load user directory."),
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
  const forbidden = computed(() => state.value.forbidden);
  const error = computed(() => state.value.error);

  const byId = computed(
    () => new Map(state.value.items.map((u) => [u.id, u])),
  );

  function labelFor(id: string | null | undefined): string {
    if (!id) return "";
    const hit = byId.value.get(id);
    if (!hit) return "";
    return hit.display_name?.trim() || hit.username?.trim() || "";
  }

  return {
    items,
    loading,
    loaded,
    forbidden,
    error,
    ensureLoaded,
    refresh,
    labelFor,
  };
}
