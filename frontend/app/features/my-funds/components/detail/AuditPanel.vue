<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AppCard from "~/shared/ui/AppCard.vue";
import { OpenApiRequestError, unwrapOpenApiResponse, useOpenApiClient } from "~/api/openapi";
import { useI18n } from "~/composables/useI18n";
import type { components } from "~/api/ims-api";

import { todayBangkokIso } from "../../lib/derive";
import { relativeTime } from "../../lib/format";

interface Props {
  fundId: string;
}

const props = defineProps<Props>();
const { t } = useI18n();
const client = useOpenApiClient();

type HistoryResponse = components["schemas"]["HistoryResponse"];
type TransitionEntry = components["schemas"]["TransitionEntry"];

const history = ref<TransitionEntry[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);
const businessDate = ref(todayBangkokIso());

const sorted = computed(() =>
  [...history.value].sort((a, b) => (b.occurredAt ?? "").localeCompare(a.occurredAt ?? "")),
);

async function load() {
  loading.value = true;
  error.value = null;
  try {
    const response = await client.GET(
      "/workflow/day-states/{contractId}/history",
      {
        params: {
          path: { contractId: props.fundId },
          query: { businessDate: businessDate.value },
        },
      },
    );
    if ("error" in response && response.error !== undefined) {
      throw new OpenApiRequestError(response.response, response.error);
    }
    const body = unwrapOpenApiResponse<HistoryResponse>(response);
    history.value = body.transitions ?? [];
  } catch (err) {
    history.value = [];
    if (err instanceof OpenApiRequestError && err.status === 404) {
      // 404 is acceptable — means no rows yet for today's business date.
      return;
    }
    error.value = describeError(err, "Failed to load audit history.");
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void load();
});

function describeError(err: unknown, fallback: string): string {
  if (err instanceof Error) return err.message;
  return fallback;
}
</script>

<template>
  <div class="audit-panel">
    <AppCard
      :title="t('myFunds.detail.auditPanel.title', 'Activity feed')"
      :subtitle="
        t(
          'myFunds.detail.auditPanel.subtitle',
          { date: businessDate },
          `Workflow transitions for ${businessDate}`,
        )
      "
    >
      <template #header-actions>
        <button type="button" class="audit-panel__btn" @click="load" :disabled="loading">
          {{ loading ? t("myFunds.detail.auditPanel.refreshing", "Refreshing…") : t("myFunds.detail.auditPanel.refresh", "Refresh") }}
        </button>
      </template>

      <div v-if="loading && history.length === 0" class="audit-panel__empty">
        {{ t("myFunds.detail.auditPanel.loading", "Loading activity…") }}
      </div>

      <div v-else-if="error" class="audit-panel__error" role="alert">
        {{ error }}
      </div>

      <div v-else-if="sorted.length === 0" class="audit-panel__empty">
        {{ t("myFunds.detail.auditPanel.empty", "No workflow activity recorded for this business date.") }}
      </div>

      <ol v-else class="audit-panel__list">
        <li
          v-for="entry in sorted"
          :key="entry.id ?? `${entry.occurredAt}-${entry.action}`"
          class="audit-panel__item"
        >
          <div class="audit-panel__pip" aria-hidden="true" />
          <div class="audit-panel__main">
            <div class="audit-panel__line">
              <span class="audit-panel__action">{{ entry.action ?? "—" }}</span>
              <span class="audit-panel__states">
                {{ entry.fromState ?? "—" }}
                <span aria-hidden="true">→</span>
                {{ entry.toState ?? "—" }}
              </span>
            </div>
            <div class="audit-panel__sub">
              <span>{{ entry.actorUsername || entry.actorId || t("myFunds.detail.auditPanel.unknownActor", "unknown actor") }}</span>
              <span>·</span>
              <span>{{ relativeTime(entry.occurredAt) }}</span>
              <span v-if="entry.reason">· {{ entry.reason }}</span>
            </div>
          </div>
        </li>
      </ol>

      <template #footer>
        <p class="audit-panel__footer">
          {{
            t(
              "myFunds.detail.auditPanel.scopeHint",
              "Showing the per-fund workflow transition log. Full audit search is in Admin → Audit.",
            )
          }}
        </p>
      </template>
    </AppCard>
  </div>
</template>

<style scoped>
.audit-panel {
  display: grid;
  gap: var(--space-3);
}

.audit-panel__btn {
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  background: var(--bg-card, #ffffff);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: 4px;
  padding: 4px 10px;
  cursor: pointer;
}

.audit-panel__btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.audit-panel__loading,
.audit-panel__empty {
  font-size: 12px;
  color: var(--text-tertiary, #6e7781);
}

.audit-panel__error {
  font-size: 12px;
  color: var(--state-danger, #cf222e);
}

.audit-panel__list {
  list-style: none;
  padding: 0;
  margin: 0;
  position: relative;
}

.audit-panel__list::before {
  content: "";
  position: absolute;
  left: 7px;
  top: 8px;
  bottom: 8px;
  width: 2px;
  background: var(--border-subtle, #d0d7de);
}

.audit-panel__item {
  position: relative;
  padding: 6px 0 6px 28px;
}

.audit-panel__pip {
  position: absolute;
  left: 1px;
  top: 12px;
  width: 14px;
  height: 14px;
  background: var(--bg-card, #ffffff);
  border: 2px solid var(--state-info, #1f6feb);
  border-radius: 50%;
}

.audit-panel__line {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: 12px;
}

.audit-panel__action {
  font-weight: 700;
  color: var(--state-info, #1f6feb);
}

.audit-panel__states {
  font-family: var(--font-mono, ui-monospace, SFMono-Regular, monospace);
  font-size: 11px;
  color: var(--text-secondary, #57606a);
}

.audit-panel__sub {
  display: flex;
  gap: 6px;
  font-size: 11px;
  color: var(--text-tertiary, #6e7781);
  margin-top: 2px;
  flex-wrap: wrap;
}

.audit-panel__footer {
  margin: 0;
  font-size: 11px;
  color: var(--text-tertiary, #6e7781);
}
</style>
