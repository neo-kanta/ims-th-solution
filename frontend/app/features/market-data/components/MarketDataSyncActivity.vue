<script setup lang="ts">
import type { SyncEvent } from "../market-data.types";

defineProps<{ events: SyncEvent[] }>();
defineEmits<{ (e: "viewAudit"): void }>();
</script>

<template>
  <div class="card md-sync-card">
    <div class="card-header">
      <div>
        <span class="card-title">Sync activity</span>
        <div class="card-subtitle">Last 2 hours</div>
      </div>
      <button class="btn btn-ghost btn-sm" @click="$emit('viewAudit')">View audit log</button>
    </div>
    <div class="card-body md-sync-card__body">
      <ul class="md-sync-list" role="list">
        <li v-for="evt in events" :key="evt.id" class="md-sync-item">
          <span class="md-sync-time">{{ evt.timestamp }}</span>
          <span :class="`md-sync-dot md-sync-dot--${evt.severity}`" aria-hidden="true" />
          <div class="md-sync-content">
            <div class="md-sync-message">
              <span class="md-sync-provider">{{ evt.provider }}</span>
              <span class="md-text-muted"> · </span>
              <span>{{ evt.message }}</span>
            </div>
            <div v-if="evt.duration || evt.code" class="md-sync-meta">
              <span v-if="evt.duration">{{ evt.duration }}</span>
              <span v-if="evt.duration && evt.code"> · </span>
              <span v-if="evt.code" :class="`md-sync-code md-sync-code--${evt.severity}`">{{ evt.code }}</span>
            </div>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>
