import { computed, type ComputedRef, type Ref } from "vue";
import { useState } from "#imports";

import { complianceApi } from "../services/complianceApi";
import { deriveRuleStatus } from "../lib/formatters";
import type {
  ComplianceRule,
  ComplianceRuleCategory,
  ComplianceRuleDerivedStatus,
} from "../types";

interface DirectoryState {
  items: ComplianceRule[];
  total: number;
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
 * Shared compliance rule directory — fetched once per `useState` key.
 *
 * The dashboard's KPI grid, category panel, and high-risk list all need the
 * same underlying rule list; instead of issuing N separate paginated calls,
 * we pull the first 200 rows (the backend cap) into a shared store and
 * derive every aggregation from it.
 *
 * This is honest about its scope: if more than 200 rules exist, the derived
 * counts are an under-estimate of the true totals. The dashboard surfaces
 * a warning when that ceiling is hit so reviewers don't read it as "the
 * fund only has 200 rules" — they read it as "showing 200 of N."
 */
export function useComplianceRuleDirectory(): {
  items: ComputedRef<ComplianceRule[]>;
  total: ComputedRef<number>;
  loading: ComputedRef<boolean>;
  loaded: ComputedRef<boolean>;
  error: ComputedRef<string | null>;
  truncated: ComputedRef<boolean>;
  ensureLoaded: () => Promise<void>;
  refresh: () => Promise<void>;
  byDerivedStatus: ComputedRef<Map<ComplianceRuleDerivedStatus, ComplianceRule[]>>;
  byCategory: ComputedRef<Map<ComplianceRuleCategory | "UNKNOWN", ComplianceRule[]>>;
  highRiskRules: ComputedRef<ComplianceRule[]>;
} {
  const state: Ref<DirectoryState> = useState<DirectoryState>(
    "compliance-rule-directory",
    () => ({
      items: [],
      total: 0,
      loaded: false,
      loading: false,
      error: null,
    }),
  );

  async function load() {
    state.value = { ...state.value, loading: true, error: null };
    try {
      const payload = await complianceApi.listRules({ limit: 200, offset: 0 });
      state.value = {
        items: payload.instances ?? [],
        total: payload.total ?? 0,
        loaded: true,
        loading: false,
        error: null,
      };
    } catch (err) {
      state.value = {
        items: [],
        total: 0,
        loaded: true,
        loading: false,
        error: extractErrorMessage(err, "Failed to load compliance rules."),
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
  const total = computed(() => state.value.total);
  const loading = computed(() => state.value.loading);
  const loaded = computed(() => state.value.loaded);
  const error = computed(() => state.value.error);
  const truncated = computed(
    () => state.value.total > 0 && state.value.items.length < state.value.total,
  );

  const byDerivedStatus = computed(() => {
    const out = new Map<ComplianceRuleDerivedStatus, ComplianceRule[]>();
    for (const rule of state.value.items) {
      const status = deriveRuleStatus(rule);
      const bucket = out.get(status) ?? [];
      bucket.push(rule);
      out.set(status, bucket);
    }
    return out;
  });

  const byCategory = computed(() => {
    const out = new Map<ComplianceRuleCategory | "UNKNOWN", ComplianceRule[]>();
    for (const rule of state.value.items) {
      const cat = (rule.type_metadata?.category ?? "UNKNOWN") as
        | ComplianceRuleCategory
        | "UNKNOWN";
      const bucket = out.get(cat) ?? [];
      bucket.push(rule);
      out.set(cat, bucket);
    }
    return out;
  });

  const highRiskRules = computed(() =>
    state.value.items.filter(
      (r) => r.type_metadata?.default_severity === "BLOCK",
    ),
  );

  return {
    items,
    total,
    loading,
    loaded,
    error,
    truncated,
    ensureLoaded,
    refresh,
    byDerivedStatus,
    byCategory,
    highRiskRules,
  };
}
