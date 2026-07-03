<script setup lang="ts">
import { computed } from "vue";
import AppIcon from "./AppIcon.vue";

interface AuditEvent {
  actor: string;
  action: string;
  timestamp: string;
  target?: string;
  module?: string;
  metadata?: Record<string, any> | null;
}

interface Props {
  events: AuditEvent[];
  compact?: boolean;
}

withDefaults(defineProps<Props>(), {
  compact: false,
});

function formatTimestamp(value: string): string {
  try {
    return new Intl.DateTimeFormat("en-CA", {
      timeZone: "Asia/Bangkok",
      year: "numeric",
      month: "short",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hourCycle: "h23",
    }).format(new Date(value));
  } catch {
    return value;
  }
}
</script>

<template>
  <div class="ims-audit-timeline" :class="{ 'ims-audit-timeline--compact': compact }">
    <div v-if="events.length === 0" class="ims-audit-timeline__empty">
      No activity logged.
    </div>
    
    <div v-else class="ims-audit-timeline__list">
      <div
        v-for="(event, idx) in events"
        :key="idx"
        class="ims-audit-timeline__item"
      >
        <!-- Timeline track line -->
        <div class="ims-audit-timeline__track">
          <div class="ims-audit-timeline__dot">
            <AppIcon name="audit" size="xs" />
          </div>
        </div>

        <!-- Event Body -->
        <div class="ims-audit-timeline__body">
          <div class="ims-audit-timeline__header">
            <span class="ims-audit-timeline__actor">{{ event.actor }}</span>
            <span class="ims-audit-timeline__action">{{ event.action }}</span>
            <span v-if="event.target" class="ims-audit-timeline__target">
              on <strong>{{ event.target }}</strong>
            </span>
            <span v-if="event.module" class="ims-audit-timeline__module-badge">
              {{ event.module }}
            </span>
          </div>

          <div class="ims-audit-timeline__time">
            {{ formatTimestamp(event.timestamp) }}
          </div>

          <!-- Metadata table when full mode and metadata is available -->
          <div
            v-if="!compact && event.metadata && Object.keys(event.metadata).length > 0"
            class="ims-audit-timeline__meta-box"
          >
            <details class="ims-audit-timeline__details">
              <summary class="ims-audit-timeline__summary">View Details</summary>
              <table class="ims-audit-timeline__meta-table">
                <tbody>
                  <tr v-for="(val, key) in event.metadata" :key="key">
                    <td class="ims-audit-timeline__meta-key">{{ key }}</td>
                    <td class="ims-audit-timeline__meta-val">
                      <pre><code>{{ val }}</code></pre>
                    </td>
                  </tr>
                </tbody>
              </table>
            </details>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ims-audit-timeline {
  width: 100%;
}

.ims-audit-timeline__empty {
  color: var(--text-tertiary, #6e7781);
  text-align: center;
  padding: var(--space-5, 20px);
  font-size: var(--font-size-sm, 14px);
}

.ims-audit-timeline__list {
  display: flex;
  flex-direction: column;
}

.ims-audit-timeline__item {
  display: grid;
  grid-template-columns: 2rem 1fr;
  gap: var(--space-3, 12px);
  position: relative;
}

.ims-audit-timeline__item:not(:last-child) .ims-audit-timeline__track::after {
  content: "";
  position: absolute;
  top: 1.5rem;
  bottom: -0.5rem;
  left: 50%;
  transform: translateX(-50%);
  width: 2px;
  background: var(--border-subtle, #d0d7de);
}

.ims-audit-timeline__track {
  display: flex;
  justify-content: center;
  position: relative;
}

.ims-audit-timeline__dot {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 50%;
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  color: var(--text-secondary, #57606a);
  z-index: 2;
}

.ims-audit-timeline__body {
  padding-bottom: var(--space-4, 16px);
  display: grid;
  gap: var(--space-1, 4px);
}

.ims-audit-timeline__header {
  font-size: var(--font-size-sm, 14px);
  color: var(--text-primary, #1f2328);
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-1, 4px);
}

.ims-audit-timeline__actor {
  font-weight: var(--font-weight-semibold, 600);
}

.ims-audit-timeline__module-badge {
  font-size: 10px;
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-sm, 4px);
  padding: 1px 4px;
  margin-left: auto;
  color: var(--text-secondary, #57606a);
  text-transform: uppercase;
  font-weight: bold;
}

.ims-audit-timeline__time {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-tertiary, #6e7781);
}

.ims-audit-timeline__meta-box {
  margin-top: var(--space-2, 8px);
}

.ims-audit-timeline__details {
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-card, #ffffff);
}

.ims-audit-timeline__summary {
  padding: var(--space-2, 8px);
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-semibold, 600);
  cursor: pointer;
  outline: none;
}

.ims-audit-timeline__meta-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-xs, 12px);
  border-top: 1px solid var(--border-subtle, #d0d7de);
}

.ims-audit-timeline__meta-table td {
  padding: var(--space-2, 8px);
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
  vertical-align: top;
}

.ims-audit-timeline__meta-table tr:last-child td {
  border-bottom: none;
}

.ims-audit-timeline__meta-key {
  width: 25%;
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-secondary, #57606a);
  background: var(--bg-card-muted, #f6f8fa);
}

.ims-audit-timeline__meta-val pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
