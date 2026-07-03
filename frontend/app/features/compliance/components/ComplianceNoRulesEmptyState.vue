<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";

interface Props {
  /**
   * When true, this is the variant shown after a pre-trade check returned
   * `rules_evaluated === 0`. Otherwise it's the dashboard-level "no rules
   * exist yet" state.
   */
  postCheck?: boolean;
  showRulesLink?: boolean;
}

withDefaults(defineProps<Props>(), {
  postCheck: false,
  showRulesLink: true,
});

const { t } = useI18n();
</script>

<template>
  <section
    class="no-rules"
    :class="{ 'no-rules--post-check': postCheck }"
    role="status"
    aria-live="polite"
  >
    <div class="no-rules__icon" aria-hidden="true">
      <AppIcon name="info" size="lg" />
    </div>
    <h2 class="no-rules__title">
      {{
        postCheck
          ? t("compliance.emptyState.noRulesEvaluatedTitle")
          : t("compliance.emptyState.noRulesTitle")
      }}
    </h2>
    <p class="no-rules__subtitle">
      {{
        postCheck
          ? t("compliance.emptyState.noRulesEvaluatedSubtitle")
          : t("compliance.emptyState.noRulesSubtitle")
      }}
    </p>

    <p v-if="!postCheck" class="no-rules__intro">
      {{ t("compliance.emptyState.noRulesIntro") }}
    </p>

    <div v-if="!postCheck" class="no-rules__checklist-wrap">
      <h3 class="no-rules__checklist-title">
        {{ t("compliance.emptyState.setupChecklist") }}
      </h3>
      <ol class="no-rules__checklist">
        <li>{{ t("compliance.emptyState.step1") }}</li>
        <li>{{ t("compliance.emptyState.step2") }}</li>
        <li>{{ t("compliance.emptyState.step3") }}</li>
      </ol>
    </div>

    <div class="no-rules__actions">
      <NuxtLink v-if="showRulesLink" to="/compliance/rules">
        <AppButton variant="primary" size="sm">
          {{ t("compliance.dashboard.goToRules") }}
        </AppButton>
      </NuxtLink>
    </div>
  </section>
</template>

<style scoped>
.no-rules {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-7);
  border: 1px solid var(--alert-info-border);
  border-radius: var(--radius-lg);
  background: var(--alert-info-bg);
  color: var(--alert-info-text);
}

.no-rules--post-check {
  border-color: var(--alert-warning-border);
  background: var(--alert-warning-bg);
  color: var(--alert-warning-text);
}

.no-rules__icon {
  width: 40px;
  height: 40px;
  display: grid;
  place-items: center;
  border-radius: var(--radius-md);
  background: var(--bg-card);
  color: var(--state-info);
}

.no-rules--post-check .no-rules__icon {
  color: var(--state-warning);
}

.no-rules__title {
  margin: 0;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.no-rules__subtitle {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.no-rules__intro {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.no-rules__checklist-wrap {
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  color: var(--text-primary);
}

.no-rules__checklist-title {
  margin: 0 0 var(--space-3);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary);
}

.no-rules__checklist {
  margin: 0;
  padding-left: var(--space-6);
  display: grid;
  gap: var(--space-3);
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.no-rules__actions {
  display: flex;
  gap: var(--space-3);
  flex-wrap: wrap;
}
</style>
