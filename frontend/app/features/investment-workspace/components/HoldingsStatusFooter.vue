<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import { formatBangkokTime } from "../lib/holdingsFormat";
import type { FundFooter } from "../types";

defineProps<{
  footer: FundFooter;
}>();

const { t } = useI18n();
</script>

<template>
  <footer class="ht-footer">
    <span class="ht-footer__item">
      <span
        class="ht-footer__dot"
        :class="footer.is_active ? 'is-on' : 'is-off'"
        aria-hidden="true"
      />
      {{ t("holdings.footer.active", "Active") }}
    </span>
    <span class="ht-footer__item">
      {{ t("holdings.footer.lastPriced", { time: formatBangkokTime(footer.last_priced_at) }, `last priced ${formatBangkokTime(footer.last_priced_at)}`) }}
    </span>
    <span class="ht-footer__item">
      {{ t("holdings.footer.irgRule", { version: footer.irg_rule_version }, `IRG ruleset ${footer.irg_rule_version}`) }}
    </span>
    <span class="ht-footer__item">
      {{ t("holdings.footer.auditHash", { hash: footer.audit_hash }, `audit hash ${footer.audit_hash}`) }}
    </span>
    <span class="ht-footer__spacer" />
    <span class="ht-footer__item ht-footer__item--right">
      {{ t("holdings.footer.businessDate", { date: footer.reference_date }, `business date ${footer.reference_date}`) }}
    </span>
    <span class="ht-footer__item ht-footer__item--right">
      {{ t("holdings.footer.language", "Asia/Bangkok") }}
    </span>
  </footer>
</template>

<style scoped>
.ht-footer {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: 8px 12px;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  font-size: 11px;
  color: var(--text-tertiary);
  flex-wrap: wrap;
}

.ht-footer__item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.ht-footer__spacer {
  flex: 1 1 auto;
}

.ht-footer__item--right {
  color: var(--text-secondary);
}

.ht-footer__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.ht-footer__dot.is-on {
  background: var(--color-success-500, #12b76a);
}

.ht-footer__dot.is-off {
  background: var(--border-strong);
}
</style>
