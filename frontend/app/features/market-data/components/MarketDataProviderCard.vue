<script setup lang="ts">
import type { MarketDataProvider } from "../market-data.types";
import { useI18n } from "~/composables/useI18n";

const props = defineProps<{ provider: MarketDataProvider }>();
defineEmits<{ (e: "configure", id: string): void; (e: "upload", id: string): void }>();
const { t } = useI18n();

const statusLabel = computed(() => {
  switch (props.provider.status) {
    case "connected":   return t("marketData.statuses.connected");
    case "rate-limited": return t("marketData.statuses.rateLimited");
    case "idle":        return t("marketData.statuses.idle");
    case "error":       return t("marketData.statuses.error");
  }
});

const statusClass = computed(() => `md-status md-status--${props.provider.status}`);

const usagePct = computed(() => Math.min(100, Math.round((props.provider.usageRatio ?? 0) * 100)));
const usageClass = computed(() => {
  const pct = usagePct.value;
  if (pct >= 90) return "md-usage-bar md-usage-bar--danger";
  if (pct >= 75) return "md-usage-bar md-usage-bar--warning";
  return "md-usage-bar md-usage-bar--ok";
});
</script>

<template>
  <div class="md-provider-card">
    <div class="md-provider-card__header">
      <div class="md-provider-icon" :data-provider="provider.id">{{ provider.initials }}</div>
      <div class="md-provider-card__title">
        <div class="md-provider-card__name">{{ provider.name }}</div>
        <div class="md-provider-card__endpoint">{{ provider.endpoint }}</div>
      </div>
      <span :class="statusClass">{{ statusLabel }}</span>
    </div>

    <div v-if="provider.usageLabel" class="md-provider-card__usage">
      <div class="md-usage-track">
        <div :class="usageClass" :style="{ width: usagePct + '%' }" />
      </div>
      <div class="md-provider-card__usage-label">{{ provider.usageLabel }}</div>
    </div>

    <div v-if="provider.description" class="md-provider-card__desc">{{ provider.description }}</div>

    <div class="md-provider-card__meta">
      <div v-if="provider.symbolsCount != null">
        <div class="md-meta-label">{{ t("marketData.labels.symbols") }}</div>
        <div class="md-meta-value">{{ provider.symbolsCount }}</div>
      </div>
      <div v-if="provider.recordsCount != null">
        <div class="md-meta-label">{{ t("marketData.labels.records") }}</div>
        <div class="md-meta-value">{{ provider.recordsCount }}</div>
      </div>
      <div v-if="provider.lastSync">
        <div class="md-meta-label">{{ t("marketData.labels.lastSync") }}</div>
        <div class="md-meta-value">{{ provider.lastSync }}</div>
      </div>
      <div v-if="provider.lastUpload">
        <div class="md-meta-label">{{ t("marketData.labels.lastUpload") }}</div>
        <div class="md-meta-value">{{ provider.lastUpload }}</div>
      </div>
      <div v-if="provider.errors24h != null">
        <div class="md-meta-label">{{ t("marketData.labels.errors24h") }}</div>
        <div class="md-meta-value" :class="{ 'md-text-danger': provider.errors24h > 0 }">
          {{ provider.errors24h }}
        </div>
      </div>
      <div v-if="provider.supportedFiles">
        <div class="md-meta-label">{{ t("marketData.labels.files") }}</div>
        <div class="md-meta-value">{{ provider.supportedFiles }}</div>
      </div>
    </div>

    <div class="md-provider-card__footer">
      <button
        v-if="provider.id === 'manual'"
        class="btn btn-secondary btn-sm"
        @click="$emit('upload', provider.id)"
      >
        {{ t("marketData.actions.uploadFile") }}
      </button>
      <button v-else class="btn btn-secondary btn-sm" @click="$emit('configure', provider.id)">
        {{ t("marketData.actions.configure") }}
      </button>
    </div>
  </div>
</template>
