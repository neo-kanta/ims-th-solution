<script setup lang="ts">
import type { SyncActivityItem } from "../market-data.types";

defineProps<{
  events: SyncActivityItem[];
}>();

function formatTime(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toISOString().slice(11, 19);
}
function statusClass(status: SyncActivityItem["status"]): string {
  switch (status) {
    case "success": return "md-sec-activity__pill--success";
    case "warning": return "md-sec-activity__pill--warning";
    case "error":   return "md-sec-activity__pill--error";
    default:        return "md-sec-activity__pill--info";
  }
}
function statusLabel(status: SyncActivityItem["status"]): string {
  switch (status) {
    case "success": return "OK";
    case "warning": return "Warning";
    case "error":   return "Error";
    default:        return "Info";
  }
}
</script>

<template>
  <article class="card md-sec-card">
    <header class="card-header">
      <span class="card-title">Sync activity</span>
    </header>
    <div class="card-body">
      <ul v-if="events.length > 0" class="md-sec-activity">
        <li
          v-for="evt in events"
          :key="evt.id"
          class="md-sec-activity__item"
        >
          <span class="md-sec-activity__time">{{ formatTime(evt.timestamp) }}</span>
          <span class="md-sec-activity__provider">{{ evt.provider }}</span>
          <span class="md-sec-activity__op">{{ evt.operation }}</span>
          <span
            class="md-sec-activity__pill"
            :class="statusClass(evt.status)"
            :aria-label="`Status ${statusLabel(evt.status)}`"
          >{{ statusLabel(evt.status) }}</span>
          <span class="md-sec-activity__msg">{{ evt.message }}</span>
        </li>
      </ul>
      <p v-else class="md-sec-activity__empty">
        No sync activity yet.
        <!-- TODO: backend GET /market-data/sync-activity?security_id=…
             is not implemented; populate from server when available. -->
      </p>
    </div>
  </article>
</template>

<style scoped>
.md-sec-activity {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 0.4rem;
}
.md-sec-activity__item {
  display: grid;
  grid-template-columns: 70px 100px 140px auto 1fr;
  gap: 0.5rem;
  align-items: center;
  font-size: 0.82rem;
  padding: 0.3rem 0;
  border-bottom: 1px solid var(--md-border);
}
.md-sec-activity__item:last-child { border-bottom: none; }
.md-sec-activity__time     { color: var(--md-text-muted); font-variant-numeric: tabular-nums; }
.md-sec-activity__provider { color: var(--md-text); }
.md-sec-activity__op       { color: var(--md-text-muted); }
.md-sec-activity__msg      { color: var(--md-text); }
.md-sec-activity__pill {
  display: inline-block;
  padding: 0.1rem 0.5rem;
  border-radius: 999px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.md-sec-activity__pill--success { background: var(--md-success-bg); color: var(--md-success-text); }
.md-sec-activity__pill--warning { background: var(--md-warning-bg); color: var(--md-warning-text); }
.md-sec-activity__pill--error   { background: var(--md-danger-bg);  color: var(--md-danger-text); }
.md-sec-activity__pill--info    { background: var(--md-info-bg);    color: var(--md-info-text); }
.md-sec-activity__empty {
  margin: 0;
  color: var(--md-text-muted);
  font-size: 0.85rem;
}
@media (max-width: 768px) {
  .md-sec-activity__item {
    grid-template-columns: 1fr;
    gap: 0.1rem;
  }
}
</style>
