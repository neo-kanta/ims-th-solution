<script setup lang="ts">
const { t } = useI18n();

definePageMeta({
  layout: "default",
});
</script>

<template>
  <div class="forbidden-page">
    <div class="forbidden-card">
      <p class="forbidden-code">403</p>
      <h1 class="forbidden-title">{{ t("forbidden.title") }}</h1>
      <p class="forbidden-message">
        {{ t("forbidden.message") }}
      </p>
      <NuxtLink to="/" class="forbidden-link">
        {{ t("forbidden.backToDashboard") }}
      </NuxtLink>
    </div>
  </div>
</template>

<style scoped>
.forbidden-page {
  min-height: 70vh;
  display: grid;
  place-items: center;
}

.forbidden-card {
  width: min(100%, 420px);
  padding: var(--space-8);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  box-shadow: var(--shadow-sm);
  text-align: center;
}

.forbidden-code {
  margin: 0 0 var(--space-2);
  color: var(--state-danger);
  font-size: 0.875rem;
  font-weight: var(--font-weight-bold);
  letter-spacing: 0.08em;
}

.forbidden-title {
  margin: 0 0 var(--space-3);
  color: var(--text-primary);
  font-size: var(--font-size-2xl);
}

.forbidden-message {
  margin: 0 0 var(--space-6);
  color: var(--text-secondary);
}

.forbidden-link {
  color: var(--text-link);
  font-weight: var(--font-weight-semibold);
  text-decoration: none;
}
</style>
