<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "~/composables/useI18n";
import type { AlertEvent } from "../types";
import { securityLabel } from "../lib/formatters";

interface Props {
  alert: AlertEvent;
  saving?: boolean;
}

withDefaults(defineProps<Props>(), {
  saving: false,
});

const emit = defineEmits<{
  confirm: [note: string | undefined];
  cancel: [];
}>();

const { t } = useI18n();
const note = ref("");

function onConfirm() {
  emit("confirm", note.value.trim() || undefined);
}
</script>

<template>
  <div class="modal-overlay" @click.self="emit('cancel')">
    <div class="modal" style="max-width: 480px; width: 100%;">
      <div class="modal-header">
        <div>
          <div class="modal-title">{{ t('watchlist.alert.acknowledgeConfirm') }}</div>
          <div class="modal-subtitle">
            {{ securityLabel(alert.security) }} —
            {{ alert.direction === 'ABOVE' ? t('watchlist.direction.above') : t('watchlist.direction.below') }}
            {{ t('watchlist.alert.thresholdValue') }}
          </div>
        </div>
        <button class="btn btn-ghost btn-icon-sm" @click="emit('cancel')">
          <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M3 3l10 10M13 3L3 13"/>
          </svg>
        </button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label class="form-label">{{ t('watchlist.alert.acknowledgeNote') }}</label>
          <textarea
            v-model="note"
            class="form-input"
            rows="3"
            :placeholder="t('watchlist.alert.acknowledgeNote') + '…'"
            maxlength="1000"
            style="resize: vertical;"
          />
          <div class="form-helper">{{ note.length }}/1000</div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary btn-sm" :disabled="saving" @click="emit('cancel')">
          {{ t('watchlist.drawer.cancel') }}
        </button>
        <button class="btn btn-primary btn-sm" :disabled="saving" @click="onConfirm">
          {{ saving ? t('watchlist.drawer.saving') : t('watchlist.alert.acknowledgeConfirm') }}
        </button>
      </div>
    </div>
  </div>
</template>
