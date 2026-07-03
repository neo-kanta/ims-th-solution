<script setup lang="ts">
import type {
  ApiSecurity,
  ApiUpdateSecurityRequest,
} from "../services/marketDataCatalogApi";
import { useI18n } from "~/composables/useI18n";

const props = defineProps<{
  security: ApiSecurity;
  saving?: boolean;
  error?: string | null;
}>();

const emit = defineEmits<{
  (e: "save", req: ApiUpdateSecurityRequest): void;
}>();
const { t } = useI18n();

const form = reactive({
  display_symbol: props.security.display_symbol ?? "",
  primary_identifier: props.security.primary_identifier ?? "",
  name: props.security.name ?? "",
  asset_type: props.security.asset_type ?? "EQUITY",
  currency: props.security.currency ?? "",
  country_code: props.security.country_code ?? "",
  exchange_mic: props.security.exchange_mic ?? "",
  isin: props.security.isin ?? "",
  cusip: props.security.cusip ?? "",
  figi: props.security.figi ?? "",
  status: props.security.status ?? "ACTIVE",
});

// Re-sync when the underlying security changes (e.g. after a save returns).
watch(() => props.security, (s) => {
  form.display_symbol = s.display_symbol ?? "";
  form.primary_identifier = s.primary_identifier ?? "";
  form.name = s.name ?? "";
  form.asset_type = s.asset_type ?? "EQUITY";
  form.currency = s.currency ?? "";
  form.country_code = s.country_code ?? "";
  form.exchange_mic = s.exchange_mic ?? "";
  form.isin = s.isin ?? "";
  form.cusip = s.cusip ?? "";
  form.figi = s.figi ?? "";
  form.status = s.status ?? "ACTIVE";
});

function onSubmit() {
  emit("save", {
    display_symbol: form.display_symbol,
    primary_identifier: form.primary_identifier,
    name: form.name,
    asset_type: form.asset_type,
    currency: form.currency || undefined,
    country_code: form.country_code || undefined,
    exchange_mic: form.exchange_mic || undefined,
    isin: form.isin || undefined,
    cusip: form.cusip || undefined,
    figi: form.figi || undefined,
    status: form.status,
  });
}
</script>

<template>
  <form class="mdc-edit" @submit.prevent="onSubmit">
    <div v-if="error" class="alert alert-danger">{{ error }}</div>

    <div class="mdc-edit-grid">
      <label class="mdc-field mdc-field--wide">
        <span class="mdc-label">{{ t("marketData.labels.name") }}</span>
        <input v-model="form.name" class="form-input" required />
      </label>
      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.assetType") }}</span>
        <select v-model="form.asset_type" class="form-input">
          <option value="EQUITY">EQUITY</option>
          <option value="ETF">ETF</option>
          <option value="MUTUAL_FUND">MUTUAL_FUND</option>
          <option value="BOND">BOND</option>
          <option value="FX">FX</option>
          <option value="INDEX">INDEX</option>
          <option value="CASH">CASH</option>
          <option value="DERIVATIVE">DERIVATIVE</option>
          <option value="OTHER">OTHER</option>
          <option value="UNKNOWN">UNKNOWN</option>
        </select>
      </label>
      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.status") }}</span>
        <select v-model="form.status" class="form-input">
          <option value="ACTIVE">ACTIVE</option>
          <option value="INACTIVE">INACTIVE</option>
          <option value="SUSPENDED">SUSPENDED</option>
        </select>
      </label>

      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.displaySymbol") }}</span>
        <input v-model="form.display_symbol" class="form-input" required />
      </label>
      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.primaryIdentifier") }}</span>
        <input v-model="form.primary_identifier" class="form-input" />
      </label>
      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.currency") }}</span>
        <input v-model="form.currency" class="form-input" maxlength="3" />
      </label>
      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.country") }}</span>
        <input v-model="form.country_code" class="form-input" maxlength="2" />
      </label>
      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.exchangeMic") }}</span>
        <input v-model="form.exchange_mic" class="form-input" />
      </label>
      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.isin") }}</span>
        <input v-model="form.isin" class="form-input" maxlength="12" />
      </label>
      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.cusip") }}</span>
        <input v-model="form.cusip" class="form-input" maxlength="16" />
      </label>
      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.figi") }}</span>
        <input v-model="form.figi" class="form-input" maxlength="32" />
      </label>
    </div>

    <div class="mdc-edit-actions">
      <button type="submit" class="btn btn-primary" :disabled="saving">
        {{ saving ? t("marketData.actions.saving") : t("marketData.actions.saveChanges") }}
      </button>
    </div>
  </form>
</template>

<style scoped>
.mdc-edit {
  display: grid;
  gap: 0.75rem;
}
.mdc-edit-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}
.mdc-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.mdc-field--wide {
  grid-column: span 3;
}
.mdc-label {
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--md-text-muted);
}
.mdc-edit-actions {
  display: flex;
  justify-content: flex-end;
}
@media (max-width: 768px) {
  .mdc-edit-grid {
    grid-template-columns: 1fr;
  }
  .mdc-field--wide {
    grid-column: auto;
  }
}
</style>
