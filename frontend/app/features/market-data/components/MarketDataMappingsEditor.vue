<script setup lang="ts">
import type {
  ApiProviderMapping,
  ApiAddMappingRequest,
} from "../services/marketDataCatalogApi";
import { useI18n } from "~/composables/useI18n";

defineProps<{
  mappings: ApiProviderMapping[];
}>();

const emit = defineEmits<{
  (e: "add", req: ApiAddMappingRequest): void;
  (e: "delete", mappingId: string): void;
}>();
const { t } = useI18n();

const form = reactive<ApiAddMappingRequest>({
  provider_code: "",
  provider_symbol: "",
  provider_exchange: "",
  provider_asset_type: "",
  provider_currency: "",
  priority: 100,
  is_primary: false,
});

const submitting = ref(false);
async function onSubmit() {
  if (!form.provider_code?.trim() || !form.provider_symbol?.trim()) return;
  submitting.value = true;
  try {
    emit("add", {
      provider_code: form.provider_code.trim(),
      provider_symbol: form.provider_symbol.trim(),
      provider_exchange: form.provider_exchange || undefined,
      provider_asset_type: form.provider_asset_type || undefined,
      provider_currency: form.provider_currency || undefined,
      priority: form.priority,
      is_primary: form.is_primary,
    });
    // reset
    form.provider_symbol = "";
    form.provider_exchange = "";
    form.provider_asset_type = "";
    form.provider_currency = "";
    form.is_primary = false;
  } finally {
    submitting.value = false;
  }
}

function statusVariant(status?: string): string {
  switch (status) {
    case "ACTIVE":   return "badge-success";
    case "INACTIVE": return "badge-neutral";
    default:         return "badge-warning";
  }
}
</script>

<template>
  <div class="card mdc-mappings">
    <div class="card-header">
      <span class="card-title">{{ t("marketData.labels.providerMappings") }}</span>
    </div>

    <div class="mdc-mapping-list">
      <div v-if="mappings.length === 0" class="mdc-empty">
        {{ t("marketData.messages.noMappingsYet") }}
      </div>
      <div
        v-for="m in mappings"
        :key="m.mapping_id"
        class="mdc-mapping"
        :class="{ 'mdc-mapping--inactive': m.mapping_status !== 'ACTIVE' }"
      >
        <div class="mdc-mapping__head">
          <div>
            <span class="mdc-mapping__provider">{{ m.provider_code }}</span>
            <span class="mdc-mapping__symbol">{{ m.provider_symbol }}</span>
            <span v-if="m.is_primary" class="badge badge-success mdc-mapping__primary">{{ t("marketData.labels.primary") }}</span>
          </div>
          <div class="mdc-mapping__head-right">
            <span class="badge" :class="statusVariant(m.mapping_status)">{{ m.mapping_status }}</span>
            <button
              v-if="m.mapping_status === 'ACTIVE' && m.mapping_id"
              class="btn btn-secondary btn-xs"
              type="button"
              @click="emit('delete', m.mapping_id!)"
            >{{ t("marketData.actions.deactivate") }}</button>
          </div>
        </div>
        <div class="mdc-mapping__meta">
          <span v-if="m.provider_asset_type">{{ m.provider_asset_type }}</span>
          <span v-if="m.provider_currency">· {{ m.provider_currency }}</span>
          <span v-if="m.provider_exchange">· {{ m.provider_exchange }}</span>
          <span v-if="m.priority != null">· priority {{ m.priority }}</span>
        </div>
      </div>
    </div>

    <form class="mdc-mapping-add" @submit.prevent="onSubmit">
      <div class="mdc-mapping-add__row">
        <label class="mdc-field">
          <span class="mdc-label">{{ t("marketData.labels.provider") }}</span>
          <input
            v-model="form.provider_code"
            class="form-input"
            list="mdc-provider-suggestions"
            placeholder="yahoo, alpha_vantage, …"
            required
          />
          <datalist id="mdc-provider-suggestions">
            <option value="yahoo" />
            <option value="alpha_vantage" />
          </datalist>
        </label>
        <label class="mdc-field">
          <span class="mdc-label">{{ t("marketData.labels.providerSymbol") }}</span>
          <input v-model="form.provider_symbol" class="form-input" required />
        </label>
      </div>
      <div class="mdc-mapping-add__row">
        <label class="mdc-field">
          <span class="mdc-label">{{ t("marketData.labels.providerAssetType") }}</span>
          <input v-model="form.provider_asset_type" class="form-input" />
        </label>
        <label class="mdc-field">
          <span class="mdc-label">{{ t("marketData.labels.ccy") }}</span>
          <input v-model="form.provider_currency" class="form-input" maxlength="3" />
        </label>
        <label class="mdc-field">
          <span class="mdc-label">{{ t("marketData.labels.exchange") }}</span>
          <input v-model="form.provider_exchange" class="form-input" />
        </label>
        <label class="mdc-field">
          <span class="mdc-label">{{ t("marketData.labels.priority") }}</span>
          <input v-model.number="form.priority" type="number" class="form-input" min="1" />
        </label>
      </div>
      <div class="mdc-mapping-add__footer">
        <label class="mdc-field--checkbox">
          <input v-model="form.is_primary" type="checkbox" />
          <span>{{ t("marketData.actions.markPrimaryMapping") }}</span>
        </label>
        <button type="submit" class="btn btn-primary btn-sm" :disabled="submitting">
          {{ submitting ? t("marketData.actions.adding") : t("marketData.actions.addMapping") }}
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.mdc-mappings .card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.mdc-mapping-list {
  display: grid;
  gap: 0.5rem;
  padding: 0.75rem;
}
.mdc-mapping {
  border: 1px solid var(--md-border);
  border-radius: 0.5rem;
  padding: 0.75rem;
  display: grid;
  gap: 0.25rem;
}
.mdc-mapping--inactive {
  opacity: 0.6;
}
.mdc-mapping__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}
.mdc-mapping__head-right {
  display: flex;
  gap: 0.4rem;
  align-items: center;
}
.mdc-mapping__provider {
  display: inline-block;
  padding: 0.05rem 0.4rem;
  border-radius: 0.25rem;
  background: var(--md-surface-muted);
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-right: 0.4rem;
}
.mdc-mapping__symbol {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-weight: 600;
}
.mdc-mapping__primary {
  margin-left: 0.5rem;
}
.mdc-mapping__meta {
  font-size: 0.78rem;
  color: var(--md-text-muted);
}
.mdc-mapping-add {
  display: grid;
  gap: 0.5rem;
  padding: 0.75rem;
  border-top: 1px solid var(--md-border);
}
.mdc-mapping-add__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
}
.mdc-mapping-add__row:nth-child(2) {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}
.mdc-mapping-add__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}
.mdc-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.mdc-field--checkbox {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.85rem;
}
.mdc-label {
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--md-text-muted);
}
.mdc-empty {
  text-align: center;
  padding: 1.5rem;
  color: var(--md-text-muted);
}
.btn-xs {
  padding: 0.2rem 0.5rem;
  font-size: 0.75rem;
}
@media (max-width: 768px) {
  .mdc-mapping-add__row,
  .mdc-mapping-add__row:nth-child(2) {
    grid-template-columns: 1fr;
  }
}
</style>
