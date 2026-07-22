<script setup lang="ts">
import { ref, watch, computed } from "vue";
import { useI18n } from "~/composables/useI18n";
import IMSPermissionGuard from "~/shared/ui/IMSPermissionGuard.vue";
import ThresholdRuleForm from "./ThresholdRuleForm.vue";
import WatchlistScopeSelector from "./WatchlistScopeSelector.vue";
import type { WatchlistItem, CreateItemBody, UpdateItemBody, ThresholdRuleInput } from "../types";
import type { WatchlistPortfolio } from "../composables/useWatchlistPortfolios";
import type { WatchlistSecurity } from "../composables/useWatchlistSecuritySearch";
import { isPositiveDecimal } from "../lib/formatters";

interface Props {
  mode: "create" | "edit";
  item?: WatchlistItem | null;
  portfolios: WatchlistPortfolio[];
  portfoliosLoading?: boolean;
  portfoliosError?: string | null;
  saving?: boolean;
  securitySearchResults?: WatchlistSecurity[];
  securitySearchLoading?: boolean;
  /**
   * Locks the scope selector to PORTFOLIO + a fixed portfolio id and hides
   * the ability to change either — used by the portfolio-workspace
   * Watchlists page so a user cannot silently create a PERSONAL item or an
   * item scoped to a different portfolio while "inside" this portfolio's
   * workspace. Undefined/false preserves the original free-choice behavior
   * for the general /watchlists page.
   */
  lockScope?: boolean;
  lockPortfolioId?: string | null;
}

const props = withDefaults(defineProps<Props>(), {
  item: null,
  portfoliosLoading: false,
  portfoliosError: null,
  saving: false,
  securitySearchResults: () => [],
  securitySearchLoading: false,
  lockScope: false,
  lockPortfolioId: null,
});

const emit = defineEmits<{
  close: [];
  create: [body: CreateItemBody];
  update: [id: string, body: UpdateItemBody];
  "security-search": [query: string];
}>();

const { t } = useI18n();

const scopeType = ref<"PERSONAL" | "PORTFOLIO">("PERSONAL");
const portfolioId = ref<string | null>(null);
const selectedSecurity = ref<WatchlistSecurity | null>(null);
const securityQuery = ref("");
const showSecurityDropdown = ref(false);
const pinned = ref(false);
const note = ref("");
const rules = ref<ThresholdRuleInput[]>([]);
const rulesModified = ref(false);

function blankRule(): ThresholdRuleInput {
  return { direction: "ABOVE", threshold_value: "", status: "ENABLED", cooldown_minutes: 60 };
}

watch(
  () => props.item,
  (item) => {
    if (item) {
      scopeType.value = (item.scope_type as "PERSONAL" | "PORTFOLIO") ?? "PERSONAL";
      portfolioId.value = item.portfolio_id ?? null;
      pinned.value = item.pinned ?? false;
      note.value = item.note ?? "";
      rules.value = (item.threshold_rules ?? []).map((r) => ({
        id: r.id,
        direction: r.direction ?? "ABOVE",
        threshold_value: r.threshold_value ?? "",
        currency: r.currency ?? undefined,
        cooldown_minutes: r.cooldown_minutes ?? 60,
        status: r.status ?? "ENABLED",
        metric_type: r.metric_type,
      }));
      rulesModified.value = false;
    } else if (props.lockScope) {
      scopeType.value = "PORTFOLIO";
      portfolioId.value = props.lockPortfolioId;
      selectedSecurity.value = null;
      securityQuery.value = "";
      pinned.value = false;
      note.value = "";
      rules.value = [blankRule()];
      rulesModified.value = false;
    } else {
      scopeType.value = "PERSONAL";
      portfolioId.value = null;
      selectedSecurity.value = null;
      securityQuery.value = "";
      pinned.value = false;
      note.value = "";
      rules.value = [blankRule()];
      rulesModified.value = false;
    }
  },
  { immediate: true },
);

function addRule() {
  rules.value = [...rules.value, blankRule()];
  rulesModified.value = true;
}

function removeRule(index: number) {
  rules.value = rules.value.filter((_, i) => i !== index);
  rulesModified.value = true;
}

function onRuleChange(index: number, val: ThresholdRuleInput) {
  const next = [...rules.value];
  next[index] = val;
  rules.value = next;
  rulesModified.value = true;
}

function onSecuritySearch(e: Event) {
  const q = (e.target as HTMLInputElement).value;
  securityQuery.value = q;
  showSecurityDropdown.value = true;
  emit("security-search", q);
}

function selectSecurity(sec: WatchlistSecurity) {
  selectedSecurity.value = sec;
  securityQuery.value = sec.display_symbol ?? sec.ims_symbol ?? "";
  showSecurityDropdown.value = false;
}

const validationError = computed<string | null>(() => {
  if (props.mode === "create" && !selectedSecurity.value?.security_id) {
    return t("watchlist.drawer.validationSelectSecurity");
  }
  if (scopeType.value === "PORTFOLIO" && !portfolioId.value) {
    return t("watchlist.drawer.validationSelectPortfolio");
  }
  for (const r of rules.value) {
    if (!r.threshold_value || !isPositiveDecimal(r.threshold_value)) {
      return t("watchlist.drawer.validationPositiveThreshold");
    }
  }
  return null;
});

function onSave() {
  if (validationError.value) return;

  if (props.mode === "create") {
    const body: CreateItemBody = {
      scope_type: scopeType.value,
      security_id: selectedSecurity.value!.security_id!,
      pinned: pinned.value || undefined,
      note: note.value || undefined,
      threshold_rules: rules.value.map((r) => ({
        direction: r.direction,
        threshold_value: r.threshold_value,
        currency: r.currency,
        cooldown_minutes: r.cooldown_minutes,
        status: r.status,
        metric_type: "MARKET_PRICE",
      })),
    };
    if (scopeType.value === "PORTFOLIO" && portfolioId.value) {
      body.portfolio_id = portfolioId.value;
    }
    emit("create", body);
  } else if (props.item?.id) {
    const body: UpdateItemBody = {
      pinned: pinned.value,
      status: props.item.status as "ACTIVE" | "DISABLED" | undefined,
    };
    if (note.value !== (props.item.note ?? "")) {
      body.note = note.value;
    }
    if (rulesModified.value) {
      body.threshold_rules = rules.value.map((r) => ({
        id: r.id,
        direction: r.direction,
        threshold_value: r.threshold_value,
        currency: r.currency,
        cooldown_minutes: r.cooldown_minutes,
        status: r.status,
        metric_type: "MARKET_PRICE",
      }));
    }
    emit("update", props.item.id, body);
  }
}

const defaultCurrency = computed(() => selectedSecurity.value?.currency ?? props.item?.security?.currency ?? "");
</script>

<template>
  <div class="drawer-overlay" @click.self="emit('close')" />
  <div class="drawer">
    <div class="drawer-header">
      <div>
        <div class="drawer-title">{{ mode === 'create' ? t('watchlist.drawer.createTitle') : t('watchlist.drawer.editTitle') }}</div>
        <div class="drawer-subtitle">{{ mode === 'create' ? t('watchlist.drawer.createSubtitle') : t('watchlist.drawer.editSubtitle') }}</div>
      </div>
      <button class="btn btn-ghost btn-icon-sm" @click="emit('close')">
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M3 3l10 10M13 3L3 13"/>
        </svg>
      </button>
    </div>

    <div class="drawer-body" style="display: flex; flex-direction: column; gap: 16px;">
      <WatchlistScopeSelector
        :scope-type="scopeType"
        :portfolio-id="portfolioId"
        :portfolios="portfolios"
        :portfolios-loading="portfoliosLoading"
        :portfolios-error="portfoliosError"
        :disabled="mode === 'edit' || lockScope"
        @update:scope-type="scopeType = $event"
        @update:portfolio-id="portfolioId = $event"
      />

      <div v-if="mode === 'create'" class="form-group" style="position: relative;">
        <label class="form-label form-label-req">{{ t('watchlist.drawer.security') }}</label>
        <input
          class="form-input"
          type="text"
          :placeholder="t('watchlist.drawer.securitySearch')"
          :value="securityQuery"
          :disabled="securitySearchLoading"
          @input="onSecuritySearch"
          @focus="showSecurityDropdown = true"
        />
        <div
          v-if="showSecurityDropdown && securitySearchResults.length > 0"
          style="position: absolute; top: 100%; left: 0; right: 0; z-index: 100; background: white; border: 1px solid var(--color-border, #e2e8f0); border-radius: 4px; max-height: 200px; overflow-y: auto; box-shadow: 0 4px 12px rgba(0,0,0,0.1);"
        >
          <button
            v-for="sec in securitySearchResults"
            :key="sec.security_id"
            type="button"
            style="display: block; width: 100%; text-align: left; padding: 8px 12px; border: none; background: none; cursor: pointer; font-size: 13px;"
            @mousedown.prevent="selectSecurity(sec)"
          >
            <span style="font-weight: 600;">{{ sec.display_symbol ?? sec.ims_symbol }}</span>
            <span style="color: var(--color-text-muted, #64748b); margin-left: 8px;">{{ sec.name }}</span>
            <span v-if="sec.asset_type" style="font-size: 11px; color: var(--color-text-muted, #94a3b8); margin-left: 4px;">{{ sec.asset_type }}</span>
          </button>
        </div>
        <div v-if="selectedSecurity" style="margin-top: 4px; font-size: 12px; color: var(--color-success, #16a34a);">
          ✓ {{ selectedSecurity.display_symbol }} – {{ selectedSecurity.name }}
        </div>
      </div>

      <div v-else class="form-group">
        <label class="form-label">{{ t('watchlist.drawer.security') }}</label>
        <div style="font-weight: 600; font-size: 13px;">
          {{ item?.security?.display_symbol ?? item?.security?.ims_symbol ?? '—' }}
        </div>
        <div v-if="item?.security?.name" style="font-size: 12px; color: var(--color-text-muted, #64748b);">
          {{ item.security.name }}
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">{{ t('watchlist.drawer.note') }}</label>
        <textarea
          v-model="note"
          class="form-input"
          rows="2"
          :placeholder="t('watchlist.drawer.notePlaceholder')"
          maxlength="1000"
          style="resize: vertical;"
        />
      </div>

      <div class="form-group">
        <label style="display: flex; align-items: center; gap: 8px; cursor: pointer; font-size: 13px;">
          <input type="checkbox" v-model="pinned" />
          {{ t('watchlist.drawer.pinned') }}
        </label>
      </div>

      <div>
        <div style="font-size: 13px; font-weight: 600; margin-bottom: 8px;">
          {{ t('watchlist.drawer.thresholdRules') }}
          <span style="font-size: 11px; color: var(--color-text-muted, #64748b); font-weight: 400; margin-left: 6px;">MARKET_PRICE</span>
        </div>

        <div
          v-for="(rule, index) in rules"
          :key="index"
          style="border: 1px solid var(--color-border, #e2e8f0); border-radius: 6px; padding: 12px; margin-bottom: 10px;"
        >
          <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;">
            <span style="font-size: 12px; color: var(--color-text-muted, #64748b);">{{ t('watchlist.drawer.ruleNumber', { number: index + 1 }) }}</span>
            <button
              type="button"
              class="btn btn-ghost btn-sm"
              style="font-size: 12px;"
              @click="removeRule(index)"
            >
              {{ t('watchlist.drawer.removeRule') }}
            </button>
          </div>
          <ThresholdRuleForm
            :model-value="rule"
            :default-currency="defaultCurrency"
            :index="index"
            @update:model-value="onRuleChange(index, $event)"
          />
        </div>

        <button
          type="button"
          class="btn btn-secondary btn-sm"
          @click="addRule"
        >
          + {{ t('watchlist.drawer.addRule') }}
        </button>
      </div>

      <div v-if="validationError" class="alert alert-danger" role="alert" style="font-size: 13px;">
        {{ validationError }}
      </div>
    </div>

    <div class="drawer-footer">
      <button class="btn btn-secondary btn-sm" :disabled="saving" @click="emit('close')">{{ t('watchlist.drawer.cancel') }}</button>
      <IMSPermissionGuard permission="WATCHLIST_MANAGE">
        <button
          class="btn btn-primary btn-sm"
          :disabled="saving || !!validationError"
          @click="onSave"
        >
          {{ saving ? t('watchlist.drawer.saving') : (mode === 'create' ? t('watchlist.page.addItem') : t('watchlist.drawer.save')) }}
        </button>
      </IMSPermissionGuard>
    </div>
  </div>
</template>
