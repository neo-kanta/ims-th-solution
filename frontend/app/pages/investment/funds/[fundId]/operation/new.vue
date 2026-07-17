<script setup lang="ts">
/**
 * Legacy fund-scoped "new decision" route. Portfolio V2
 * (docs/frontend/portfolio-v2-frontend-ddd.md) replaces this screen with a
 * portfolio-scoped decision form at /portfolios/{portfolioCode}/decisions/new
 * so there is exactly one decision form, not two divergent ones. This shell
 * resolves the fund's portfolio(s) and forwards the user there instead of
 * rendering its own form.
 */
import { computed, onMounted, ref } from "vue";
import { useI18n } from "~/composables/useI18n";
import { investmentLedgerApi } from "~/features/investment-ledger/services/investmentLedgerApi";
import { resolveLegacyFundRedirect, type LegacyRedirectPortfolio } from "~/features/portfolio-decision/lib/legacyEntry";
import { newDecisionPath } from "~/features/portfolio-decision/lib/decisionRoutes";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_FUND_VIEW",
});

const route = useRoute();
const router = useRouter();
const fundId = computed(() => route.params.fundId as string);
const { t } = useI18n();

const loading = ref(true);
const error = ref<string | null>(null);
const choices = ref<LegacyRedirectPortfolio[]>([]);
const isEmpty = ref(false);

onMounted(async () => {
  loading.value = true;
  error.value = null;
  choices.value = [];
  isEmpty.value = false;
  try {
    const result = await investmentLedgerApi.listPortfolios({ fund_id: fundId.value });
    const resolved = resolveLegacyFundRedirect(result.items ?? []);
    if (resolved.kind === "redirect") {
      void router.replace(resolved.path);
      return;
    }
    if (resolved.kind === "empty") {
      isEmpty.value = true;
    } else {
      choices.value = resolved.portfolios;
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : t("portfolio.decisionNew.legacyRedirect.genericError");
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div class="legacy-redirect">
    <div v-if="loading" class="legacy-redirect__notice" role="status">
      {{ t("portfolio.decisionNew.legacyRedirect.loading") }}
    </div>

    <div v-else-if="error" class="legacy-redirect__error" role="alert">
      {{ error }}
    </div>

    <div v-else-if="isEmpty" class="legacy-redirect__notice">
      <p>{{ t("portfolio.decisionNew.legacyRedirect.empty") }}</p>
      <NuxtLink to="/portfolios" class="legacy-redirect__link">
        {{ t("portfolio.decisionNew.legacyRedirect.emptyLink") }}
      </NuxtLink>
    </div>

    <div v-else class="legacy-redirect__notice">
      <p>{{ t("portfolio.decisionNew.legacyRedirect.multiple") }}</p>
      <ul class="legacy-redirect__list">
        <li v-for="p in choices" :key="p.code">
          <NuxtLink :to="newDecisionPath(p.code)" class="legacy-redirect__link">
            {{ p.code }} · {{ p.name }}
          </NuxtLink>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
.legacy-redirect {
  padding: 32px;
  max-width: 520px;
}

.legacy-redirect__notice {
  font-size: 13px;
  color: var(--text-secondary);
  display: grid;
  gap: 12px;
}

.legacy-redirect__error {
  padding: 10px 14px;
  background: var(--alert-danger-bg, #ffebe9);
  border: 1px solid var(--alert-danger-border, #cf222e);
  border-radius: var(--radius-sm, 4px);
  color: var(--alert-danger-text, #cf222e);
  font-size: 13px;
}

.legacy-redirect__list {
  margin: 0;
  padding-left: 18px;
  display: grid;
  gap: 6px;
}

.legacy-redirect__link {
  color: var(--action-primary, #0969da);
  font-weight: 600;
  text-decoration: none;
}

.legacy-redirect__link:hover {
  text-decoration: underline;
}
</style>
