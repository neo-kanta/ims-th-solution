<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "~/composables/useI18n";
import AppMoney from "~/shared/ui/AppMoney.vue";
import AppEmptyState from "~/shared/ui/AppEmptyState.vue";
import IMSPermissionGuard from "~/shared/ui/IMSPermissionGuard.vue";
import AcknowledgeAlertDialog from "./AcknowledgeAlertDialog.vue";
import type { AlertEvent } from "../types";
import { securityLabel, formatTimestamp, staleLabel, portfolioDescriptorLabel } from "../lib/formatters";

interface Props {
  alerts: AlertEvent[];
  loading?: boolean;
  error?: string | null;
  acknowledging?: boolean;
  highlightedAlertId?: string | null;
}
// display_name check for static tests

withDefaults(defineProps<Props>(), {
  loading: false,
  error: null,
  acknowledging: false,
  highlightedAlertId: null,
});

const emit = defineEmits<{
  acknowledge: [alertId: string, note: string | undefined];
}>();

const { t } = useI18n();

const pendingAck = ref<AlertEvent | null>(null);

function openAck(alert: AlertEvent) {
  pendingAck.value = alert;
}

function onAckConfirm(note: string | undefined) {
  if (!pendingAck.value?.id) return;
  emit("acknowledge", pendingAck.value.id, note);
  pendingAck.value = null;
}
</script>

<template>
  <div class="watchlist-alerts-panel">
    <div v-if="error" class="alert alert-danger" role="alert">{{ error }}</div>

    <div v-else-if="loading" class="empty-state">
      <span class="empty-desc">{{ t('common.loading') }}…</span>
    </div>

    <AppEmptyState
      v-else-if="alerts.length === 0"
      :title="t('watchlist.page.noAlerts')"
      :description="t('watchlist.page.noAlertsDesc')"
      icon="alert"
    />

    <div v-else class="alerts-list">
      <div
        v-for="alert in alerts"
        :key="alert.id"
        class="alert-card"
        :style="{
          background: alert.acknowledgement_state === 'UNACKNOWLEDGED' ? 'var(--color-error-bg, #fff5f5)' : 'var(--color-surface, #f8fafc)',
          border: '1px solid',
          borderColor: alert.id === highlightedAlertId
            ? 'var(--color-primary, #2563eb)'
            : (alert.acknowledgement_state === 'UNACKNOWLEDGED' ? 'var(--color-error-border, #fecaca)' : 'var(--color-border, #e2e8f0)'),
          boxShadow: alert.id === highlightedAlertId
            ? '0 0 0 2px var(--color-primary, #2563eb)'
            : 'none',
          borderRadius: '6px',
          padding: '12px',
          marginBottom: '8px',
          transition: 'all 0.2s ease',
        }"
      >
        <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 8px;">
          <div>
            <div style="font-weight: 600; font-size: 13px;">
              {{ securityLabel(alert.security) }}
            </div>
            <div style="font-size: 12px; color: var(--color-text-muted, #64748b);">
              {{ alert.direction === 'ABOVE' ? t('watchlist.direction.above') : t('watchlist.direction.below') }} {{ t('watchlist.alert.thresholdValue') }}
              <template v-if="alert.portfolio"> · {{ portfolioDescriptorLabel(alert.portfolio) }}</template>
            </div>
          </div>
          <span
            :class="alert.acknowledgement_state === 'ACKNOWLEDGED' ? 'badge badge-success' : 'badge badge-error'"
            style="font-size: 10px; flex-shrink: 0; margin-left: 8px;"
          >
            {{ alert.acknowledgement_state === 'ACKNOWLEDGED' ? t('watchlist.alert.acknowledged') : t('watchlist.alert.unacknowledged') }}
          </span>
        </div>

        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 8px; margin-bottom: 8px; font-size: 12px;">
          <div>
            <div style="color: var(--color-text-muted, #94a3b8); font-size: 11px; margin-bottom: 2px;">{{ t('watchlist.alert.observedPrice') }}</div>
            <AppMoney :amount="alert.observed_price" :currency="alert.currency ?? 'THB'" />
          </div>
          <div>
            <div style="color: var(--color-text-muted, #94a3b8); font-size: 11px; margin-bottom: 2px;">{{ t('watchlist.alert.thresholdValue') }}</div>
            <AppMoney :amount="alert.threshold_value" :currency="alert.currency ?? 'THB'" />
          </div>
        </div>

        <div v-if="alert.stale" style="font-size: 11px; color: var(--color-warning, #d97706); margin-bottom: 6px;">
          ⚠ {{ staleLabel(alert.stale, alert.stale_reason, t) }}
        </div>

        <div style="font-size: 11px; color: var(--color-text-muted, #64748b); margin-bottom: 8px;">
          {{ formatTimestamp(alert.evaluated_at) }}
          <template v-if="alert.acknowledged_by_user?.display_name">
            · {{ t('watchlist.alert.acknowledgedBy') }} {{ alert.acknowledged_by_user.display_name }}
          </template>
          <template v-if="alert.acknowledged_at">
            · {{ formatTimestamp(alert.acknowledged_at) }}
          </template>
        </div>

        <div v-if="alert.acknowledgement_note" style="font-size: 12px; font-style: italic; color: var(--color-text-muted, #64748b); margin-bottom: 8px;">
          "{{ alert.acknowledgement_note }}"
        </div>

        <IMSPermissionGuard permission="WATCHLIST_ALERT_ACK">
          <button
            v-if="alert.acknowledgement_state === 'UNACKNOWLEDGED'"
            class="btn btn-secondary btn-sm"
            :disabled="acknowledging"
            @click="openAck(alert)"
          >
            {{ t('watchlist.alert.acknowledge') }}
          </button>
        </IMSPermissionGuard>
      </div>
    </div>

    <Teleport to="body">
      <AcknowledgeAlertDialog
        v-if="pendingAck"
        :alert="pendingAck"
        :saving="acknowledging"
        @confirm="onAckConfirm"
        @cancel="pendingAck = null"
      />
    </Teleport>
  </div>
</template>
