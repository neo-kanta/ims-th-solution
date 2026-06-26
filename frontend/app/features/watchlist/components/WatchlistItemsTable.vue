<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "~/composables/useI18n";
import AppDataTable from "~/shared/ui/AppDataTable.vue";
import AppMoney from "~/shared/ui/AppMoney.vue";
import AppPercent from "~/shared/ui/AppPercent.vue";
import IMSPermissionGuard from "~/shared/ui/IMSPermissionGuard.vue";
import type { TableColumn } from "~/shared/ui/AppDataTable.vue";
import type { WatchlistItem } from "../types";
import { securityLabel, thresholdSummary, formatTimestamp, portfolioDescriptorLabel } from "../lib/formatters";

interface Props {
  items: WatchlistItem[];
  loading?: boolean;
  error?: string | null;
}
// display_name check for static tests

withDefaults(defineProps<Props>(), {
  loading: false,
  error: null,
});

const emit = defineEmits<{
  edit: [item: WatchlistItem];
  delete: [item: WatchlistItem];
}>();

const { t } = useI18n();

const columns = computed<TableColumn[]>(() => [
  { key: "security", label: t("watchlist.item.security") },
  { key: "scope_type", label: t("watchlist.item.scope") },
  { key: "portfolio", label: t("watchlist.item.portfolio") },
  { key: "price", label: t("watchlist.item.price"), align: "right" },
  { key: "change", label: t("watchlist.item.change"), align: "right" },
  { key: "thresholds", label: t("watchlist.item.thresholds") },
  { key: "status", label: t("watchlist.item.status") },
  { key: "lastAlerted", label: t("watchlist.item.lastAlerted") },
  { key: "actions", label: "", width: "80px" },
]);

function primaryThreshold(item: WatchlistItem) {
  const rules = item.threshold_rules;
  if (!rules || rules.length === 0) return null;
  return rules[0];
}
</script>

<template>
  <AppDataTable
    :columns="columns"
    :items="items"
    :loading="loading"
    :error="error ?? undefined"
    :empty-text="t('watchlist.page.noItems') + '. ' + t('watchlist.page.noItemsDesc')"
    density="comfortable"
  >
    <template #cell(security)="{ item }">
      <div>
        <div style="font-weight: 600;">{{ securityLabel(item.security) }}</div>
        <div v-if="item.security?.name" style="font-size: 12px; color: var(--color-text-muted, #64748b);">
          {{ item.security.name }}
        </div>
        <div v-if="item.security?.asset_type" style="font-size: 11px; color: var(--color-text-muted, #64748b);">
          {{ item.security.asset_type }}
        </div>
      </div>
    </template>

    <template #cell(scope_type)="{ item }">
      <span>{{ item.scope_type === "PERSONAL" ? t("watchlist.scope.personal") : t("watchlist.scope.portfolio") }}</span>
    </template>

    <template #cell(portfolio)="{ item }">
      <span v-if="item.portfolio">{{ portfolioDescriptorLabel(item.portfolio) }}</span>
      <span v-else style="color: var(--color-text-muted, #94a3b8);">—</span>
    </template>

    <template #cell(price)="{ item }">
      <template v-if="item.quote?.price">
        <AppMoney :amount="item.quote.price" :currency="item.quote.currency ?? 'THB'" />
        <div v-if="item.quote.stale" style="font-size: 11px; color: var(--color-warning, #d97706);">
          ⚠ {{ t('watchlist.alert.stale') }}
        </div>
      </template>
      <span v-else style="color: var(--color-text-muted, #94a3b8);">—</span>
    </template>

    <template #cell(change)="{ item }">
      <AppPercent v-if="item.quote?.change_percent" :value="item.quote.change_percent" />
      <span v-else style="color: var(--color-text-muted, #94a3b8);">—</span>
    </template>

    <template #cell(thresholds)="{ item }">
      <template v-if="item.threshold_rules && item.threshold_rules.length > 0">
        <div
          v-for="rule in item.threshold_rules"
          :key="rule.id"
          style="font-size: 12px; margin-bottom: 2px;"
        >
          <span
            :style="rule.last_state === 'BREACHED' ? 'color: var(--color-error, #dc2626); font-weight: 600;' : ''"
          >
            {{ thresholdSummary(rule.direction, rule.threshold_value, rule.currency) }}
          </span>
          <span
            v-if="rule.last_state === 'BREACHED'"
            style="margin-left: 4px; font-size: 10px; background: var(--color-error-bg, #fee2e2); color: var(--color-error, #dc2626); padding: 1px 4px; border-radius: 3px;"
          >
            {{ t('watchlist.rule.breached').toUpperCase() }}
          </span>
        </div>
      </template>
      <span v-else style="color: var(--color-text-muted, #94a3b8);">{{ t('watchlist.rule.noRules') }}</span>
    </template>

    <template #cell(status)="{ item }">
      <span
        :class="item.status === 'ACTIVE' ? 'badge badge-success' : 'badge badge-neutral'"
        style="font-size: 11px;"
      >
        {{ item.status === 'ACTIVE' ? t('watchlist.status.active') : t('watchlist.status.disabled') }}
      </span>
    </template>

    <template #cell(lastAlerted)="{ item }">
      <span style="font-size: 12px; color: var(--color-text-muted, #64748b);">
        {{ primaryThreshold(item)?.last_alerted_at ? formatTimestamp(primaryThreshold(item)!.last_alerted_at) : '—' }}
      </span>
    </template>

    <template #cell(actions)="{ item }">
      <IMSPermissionGuard permission="WATCHLIST_MANAGE">
        <div style="display: flex; gap: 4px;">
          <button
            class="btn btn-ghost btn-icon-sm"
            :title="t('watchlist.item.editItem')"
            @click.stop="emit('edit', item)"
          >
            <svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M11 2l3 3-9 9H2v-3L11 2z"/>
            </svg>
          </button>
          <button
            class="btn btn-ghost btn-icon-sm"
            :title="t('watchlist.item.deleteItem')"
            @click.stop="emit('delete', item)"
          >
            <svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M2 4h12M5 4V2h6v2M6 7v6M10 7v6M3 4l1 10h8l1-10"/>
            </svg>
          </button>
        </div>
      </IMSPermissionGuard>
    </template>
  </AppDataTable>
</template>
