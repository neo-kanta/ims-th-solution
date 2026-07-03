<script setup lang="ts">
import AppCard from "~/shared/ui/AppCard.vue";
import { useI18n } from "~/composables/useI18n";

interface Props {
  title: string;
  body: string;
  /** Optional explainer for the backend gap (capability matrix entry). */
  backendGap?: string;
  /** Optional CTA label + path (for cross-linking to existing pages). */
  ctaLabel?: string;
  ctaPath?: string;
}

withDefaults(defineProps<Props>(), {
  backendGap: "",
  ctaLabel: "",
  ctaPath: "",
});

const { t } = useI18n();
</script>

<template>
  <AppCard :title="title">
    <div class="empty-tab">
      <p class="empty-tab__body">{{ body }}</p>
      <p v-if="backendGap" class="empty-tab__gap">
        <span class="empty-tab__gap-tag">{{ t("myFunds.detail.empty.backendGap", "Backend gap") }}</span>
        <span>{{ backendGap }}</span>
      </p>
      <router-link
        v-if="ctaLabel && ctaPath"
        :to="ctaPath"
        class="empty-tab__cta"
      >
        {{ ctaLabel }}
      </router-link>
    </div>
  </AppCard>
</template>

<style scoped>
.empty-tab {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-4) 0;
  place-items: center;
  text-align: center;
  max-width: 36rem;
  margin: 0 auto;
}

.empty-tab__body {
  margin: 0;
  font-size: 13px;
  color: var(--text-secondary, #57606a);
}

.empty-tab__gap {
  margin: 0;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-tertiary, #6e7781);
  flex-wrap: wrap;
  justify-content: center;
}

.empty-tab__gap-tag {
  background: rgba(217, 119, 6, 0.12);
  color: var(--state-warning, #9a6700);
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.empty-tab__cta {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--state-info, #1f6feb);
  color: #ffffff;
  padding: 6px 12px;
  border-radius: 5px;
  font-weight: 600;
  font-size: 12px;
  text-decoration: none;
}

.empty-tab__cta:hover {
  filter: brightness(0.95);
}
</style>
