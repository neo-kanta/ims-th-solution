<script setup lang="ts">
import { computed, ref, watch } from "vue";

import type { AuditEvent } from "../audit.types";
import type { AuditLogFilters } from "../ui.types";
import {
  getAuditSeverity,
  getAuditSeverityClass,
  getAuditSeverityLabel,
} from "../lib/audit";

const props = defineProps<{
  events: AuditEvent[];
  total: number;
  offset: number;
  limit: number;
  loading: boolean;
  exporting: boolean;
  error: string | null;
  filters: AuditLogFilters;
  formatDateTime: (value?: string | null) => string;
}>();

const emit = defineEmits<{
  apply: [filters: AuditLogFilters];
  page: [offset: number];
  refresh: [];
  export: [];
}>();

const { t } = useI18n();

const actorId = ref(props.filters.actorId);
const eventType = ref(props.filters.eventType);
const targetType = ref(props.filters.targetType);
const targetId = ref(props.filters.targetId);
const since = ref(props.filters.since);
const until = ref(props.filters.until);

watch(
  () => props.filters,
  (filters) => {
    actorId.value = filters.actorId;
    eventType.value = filters.eventType;
    targetType.value = filters.targetType;
    targetId.value = filters.targetId;
    since.value = filters.since;
    until.value = filters.until;
  },
  { deep: true },
);

const rangeLabel = computed(() => {
  if (props.total === 0) {
    return "0 of 0";
  }

  const start = props.offset + 1;
  const end = Math.min(props.offset + props.events.length, props.total);
  return `${start}-${end} of ${props.total}`;
});

const canPageBack = computed(() => props.offset > 0 && !props.loading);
const canPageForward = computed(
  () => props.offset + props.limit < props.total && !props.loading,
);

function applyFilters() {
  emit("apply", {
    actorId: actorId.value,
    eventType: eventType.value,
    targetType: targetType.value,
    targetId: targetId.value,
    since: since.value,
    until: until.value,
  });
}

function severityLabel(eventTypeValue: string) {
  return getAuditSeverityLabel(getAuditSeverity(eventTypeValue));
}

function severityClass(eventTypeValue: string) {
  return getAuditSeverityClass(getAuditSeverity(eventTypeValue));
}
</script>

<template>
  <section class="settings-panel settings-audit-log">
    <header class="settings-panel__header">
      <div>
        <h2 class="settings-panel__title">{{ t("settings.auditTitle") }}</h2>
        <p class="settings-panel__subtitle">
          Immutable IAM security events with actor, target, and network context.
        </p>
      </div>
      <div class="settings-audit-log__header-actions">
        <span class="badge badge-neutral">
          {{ t("settings.auditCount", { count: total }) }}
        </span>
        <AppButton
          variant="secondary"
          size="sm"
          :loading="loading"
          @click="emit('refresh')"
        >
          <AppIcon name="refresh" size="xs" />
          <span>{{ t("settings.refreshAudit") }}</span>
        </AppButton>
        <AppButton
          variant="secondary"
          size="sm"
          :loading="exporting"
          :disabled="loading || total === 0"
          @click="emit('export')"
        >
          Export CSV
        </AppButton>
      </div>
    </header>

    <form class="settings-audit-log__filters" @submit.prevent="applyFilters">
      <div class="form-group">
        <label for="settings-audit-event-type" class="label">Event type</label>
        <input
          id="settings-audit-event-type"
          v-model="eventType"
          class="input"
          type="text"
          :placeholder="t('settings.eventTypePlaceholder')"
        />
      </div>
      <div class="form-group">
        <label for="settings-audit-actor-id" class="label">Actor ID</label>
        <input
          id="settings-audit-actor-id"
          v-model="actorId"
          class="input"
          type="text"
          :placeholder="t('settings.actorIdPlaceholder')"
        />
      </div>
      <div class="form-group">
        <label for="settings-audit-target-type" class="label">Target type</label>
        <input
          id="settings-audit-target-type"
          v-model="targetType"
          class="input"
          type="text"
          :placeholder="t('settings.targetTypePlaceholder')"
        />
      </div>
      <div class="form-group">
        <label for="settings-audit-target-id" class="label">Target ID</label>
        <input
          id="settings-audit-target-id"
          v-model="targetId"
          class="input"
          type="text"
          :placeholder="t('settings.targetIdPlaceholder')"
        />
      </div>
      <div class="form-group">
        <label for="settings-audit-since" class="label">Since</label>
        <input
          id="settings-audit-since"
          v-model="since"
          class="input"
          type="datetime-local"
        />
      </div>
      <div class="form-group">
        <label for="settings-audit-until" class="label">Until</label>
        <input
          id="settings-audit-until"
          v-model="until"
          class="input"
          type="datetime-local"
        />
      </div>
      <AppButton
        class="settings-audit-log__apply"
        type="submit"
        variant="secondary"
        size="sm"
        :loading="loading"
      >
        {{ t("common.apply") }}
      </AppButton>
    </form>

    <div v-if="error" class="alert alert-danger settings-panel__alert" role="alert">
      {{ error }}
    </div>

    <div class="settings-audit-log__body">
      <div class="table-wrap settings-audit-log__table">
        <table class="table">
          <thead>
            <tr>
              <th>{{ t("settings.auditTable.event") }}</th>
              <th>{{ t("settings.auditTable.target") }}</th>
              <th>{{ t("settings.auditTable.actor") }}</th>
              <th>Network</th>
              <th>{{ t("settings.auditTable.created") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="5" class="settings-table-state">
                {{ t("settings.loadingAudit") }}
              </td>
            </tr>
            <tr v-else-if="events.length === 0">
              <td colspan="5" class="settings-table-state">
                {{ t("settings.noAudit") }}
              </td>
            </tr>
            <template v-else>
              <tr v-for="event in events" :key="event.id">
                <td>
                  <div class="settings-audit-log__event">
                    <span class="badge" :class="severityClass(event.event_type)">
                      {{ severityLabel(event.event_type) }}
                    </span>
                    <div>
                      <div class="settings-record-primary">
                        {{ event.event_type }}
                      </div>
                      <div class="settings-record-secondary">{{ event.id }}</div>
                    </div>
                  </div>
                </td>
                <td>
                  <div class="settings-record-primary">
                    {{ event.target_type || t("common.notAvailable") }}
                  </div>
                  <div class="settings-record-secondary">
                    {{ event.target_id || t("common.notAvailable") }}
                  </div>
                </td>
                <td>
                  {{ event.actor_id || t("settings.systemActor") }}
                </td>
                <td>
                  <div class="settings-record-primary">
                    {{ event.ip_address || t("common.notAvailable") }}
                  </div>
                  <div class="settings-record-secondary">
                    {{ event.user_agent || t("common.notAvailable") }}
                  </div>
                </td>
                <td>{{ formatDateTime(event.created_at) }}</td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <div class="settings-audit-log__mobile-list">
        <div v-if="loading" class="settings-card-state">
          {{ t("settings.loadingAudit") }}
        </div>
        <div v-else-if="events.length === 0" class="settings-card-state">
          {{ t("settings.noAudit") }}
        </div>
        <template v-else>
          <article
            v-for="event in events"
            :key="event.id"
            class="settings-audit-card"
          >
            <div class="settings-audit-card__head">
              <span class="badge" :class="severityClass(event.event_type)">
                {{ severityLabel(event.event_type) }}
              </span>
              <span>{{ formatDateTime(event.created_at) }}</span>
            </div>
            <div class="settings-record-primary">{{ event.event_type }}</div>
            <div class="settings-record-secondary">
              {{ event.target_type }} / {{ event.target_id }}
            </div>
            <div class="settings-record-secondary">
              {{ event.actor_id || t("settings.systemActor") }}
            </div>
            <div class="settings-record-secondary">
              {{ event.ip_address || t("common.notAvailable") }}
            </div>
          </article>
        </template>
      </div>
    </div>

    <footer class="settings-panel__footer">
      <span>{{ rangeLabel }}</span>
      <div class="settings-pagination">
        <AppButton
          variant="secondary"
          size="sm"
          :disabled="!canPageBack"
          @click="emit('page', Math.max(0, offset - limit))"
        >
          {{ t("common.previous") }}
        </AppButton>
        <AppButton
          variant="secondary"
          size="sm"
          :disabled="!canPageForward"
          @click="emit('page', offset + limit)"
        >
          {{ t("common.next") }}
        </AppButton>
      </div>
    </footer>
  </section>
</template>

<style scoped>
.settings-audit-log {
  min-width: 0;
}

.settings-audit-log__header-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.settings-audit-log__filters {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr)) auto;
  gap: var(--space-4);
  align-items: end;
  padding: var(--space-5);
  border-bottom: 1px solid var(--border-subtle);
}

.settings-audit-log__filters .form-group {
  margin-bottom: 0;
}

.settings-audit-log__apply {
  align-self: end;
}

.settings-audit-log__body {
  padding: var(--space-5);
}

.settings-audit-log__table {
  border-radius: var(--radius-md);
}

.settings-audit-log__mobile-list {
  display: none;
}

.settings-audit-log__event {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-3);
  align-items: start;
}

.settings-audit-card {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.settings-audit-card__head {
  display: flex;
  justify-content: space-between;
  gap: var(--space-3);
  align-items: center;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.settings-record-primary {
  display: block;
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
}

.settings-record-secondary {
  display: block;
  margin-top: var(--space-1);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-relaxed);
  overflow-wrap: anywhere;
}

.settings-table-state,
.settings-card-state {
  color: var(--text-secondary);
  text-align: center;
}

.settings-card-state {
  padding: var(--space-6);
}

@media (max-width: 1180px) {
  .settings-audit-log__filters {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .settings-audit-log__filters {
    grid-template-columns: 1fr;
  }

  .settings-audit-log__table {
    display: none;
  }

  .settings-audit-log__mobile-list {
    display: grid;
    gap: var(--space-3);
  }

  .settings-audit-log__header-actions {
    justify-content: stretch;
  }
}
</style>
