<script setup lang="ts">
import type { ApiCreateSecurityRequest } from "../services/marketDataCatalogApi";
import { useI18n } from "~/composables/useI18n";

const emit = defineEmits<{
  (e: "submit", req: ApiCreateSecurityRequest): void;
}>();
const { t } = useI18n();

const form = reactive<ApiCreateSecurityRequest>({
  ims_symbol: "",
  display_symbol: "",
  name: "",
  asset_type: "EQUITY",
  currency: "",
  country_code: "",
  exchange_mic: "",
  isin: "",
  cusip: "",
  figi: "",
  status: "ACTIVE",
  primary_identifier: "",
  auto_build_ims_symbol: true,
});

const submitting = ref(false);
const errorText = ref<string | null>(null);

async function onSubmit() {
  errorText.value = null;
  submitting.value = true;
  try {
    const payload: ApiCreateSecurityRequest = {
      ...form,
      ims_symbol: form.ims_symbol || undefined,
      currency: form.currency || undefined,
      country_code: form.country_code || undefined,
      exchange_mic: form.exchange_mic || undefined,
      isin: form.isin || undefined,
      cusip: form.cusip || undefined,
      figi: form.figi || undefined,
      primary_identifier: form.primary_identifier || undefined,
    };
    emit("submit", payload);
  } catch (err) {
    errorText.value = err instanceof Error ? err.message : t("marketData.messages.failedCreateSecurity");
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <form class="mdc-form" @submit.prevent="onSubmit">
    <div v-if="errorText" class="alert alert-danger">{{ errorText }}</div>

    <div class="mdc-form-grid">
      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.assetTypeRequired") }}</span>
        <select v-model="form.asset_type" class="form-input" required>
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
        <span class="mdc-label">{{ t("marketData.labels.displaySymbolRequired") }}</span>
        <input v-model="form.display_symbol" class="form-input" required placeholder="KBANK.BK" />
      </label>

      <label class="mdc-field mdc-field--wide">
        <span class="mdc-label">{{ t("marketData.labels.nameRequired") }}</span>
        <input v-model="form.name" class="form-input" required placeholder="Kasikornbank PCL" />
      </label>

      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.imsSymbol") }}</span>
        <input v-model="form.ims_symbol" class="form-input" placeholder="TH_EQ_XBKK_KBANK" />
      </label>

      <label class="mdc-field mdc-field--checkbox">
        <input v-model="form.auto_build_ims_symbol" type="checkbox" />
        <span>{{ t("marketData.labels.autoBuildImsSymbol") }}</span>
      </label>

      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.currencyChar") }}</span>
        <input v-model="form.currency" class="form-input" maxlength="3" placeholder="THB" />
      </label>

      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.countryChar") }}</span>
        <input v-model="form.country_code" class="form-input" maxlength="2" placeholder="TH" />
      </label>

      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.exchangeMic") }}</span>
        <input v-model="form.exchange_mic" class="form-input" placeholder="XBKK" />
      </label>

      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.isinChar") }}</span>
        <input v-model="form.isin" class="form-input" maxlength="12" placeholder="TH0016010R14" />
      </label>

      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.cusip") }}</span>
        <input v-model="form.cusip" class="form-input" maxlength="16" />
      </label>

      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.figi") }}</span>
        <input v-model="form.figi" class="form-input" maxlength="32" />
      </label>

      <label class="mdc-field">
        <span class="mdc-label">{{ t("marketData.labels.status") }}</span>
        <select v-model="form.status" class="form-input">
          <option value="ACTIVE">ACTIVE</option>
          <option value="INACTIVE">INACTIVE</option>
          <option value="SUSPENDED">SUSPENDED</option>
        </select>
      </label>
    </div>

    <div class="mdc-form-actions">
      <button type="submit" class="btn btn-primary" :disabled="submitting">
        {{ submitting ? t("marketData.actions.creating") : t("marketData.actions.create") }}
      </button>
    </div>
  </form>
</template>

<style scoped>
.mdc-form {
  display: grid;
  gap: 1rem;
}
.mdc-form-grid {
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
  grid-column: span 2;
}
.mdc-field--checkbox {
  flex-direction: row;
  align-items: center;
  gap: 0.5rem;
  margin-top: 1.4rem;
}
.mdc-label {
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--md-text-muted);
}
.mdc-form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
@media (max-width: 768px) {
  .mdc-form-grid {
    grid-template-columns: 1fr;
  }
  .mdc-field--wide {
    grid-column: auto;
  }
}
</style>
