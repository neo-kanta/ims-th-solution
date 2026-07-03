<script setup lang="ts">
import type { BondUploadMetadata } from "../market-data.types";

defineProps<{ lastUpload: BondUploadMetadata | null }>();
const emit = defineEmits<{
  (e: "browse"): void;
  (e: "upload", files: FileList): void;
  (e: "viewDiff"): void;
}>();

const dragOver = ref(false);

function onDrop(e: DragEvent) {
  e.preventDefault();
  dragOver.value = false;
  if (e.dataTransfer?.files?.length) emit("upload", e.dataTransfer.files);
}

function onDrag(e: DragEvent, over: boolean) {
  e.preventDefault();
  dragOver.value = over;
}
</script>

<template>
  <div class="card md-upload-card">
    <div class="card-header">
      <div>
        <span class="card-title">Bond reference upload</span>
        <div class="card-subtitle">Manual ThaiBMA import</div>
      </div>
    </div>
    <div class="card-body md-upload-card__body">
      <div
        class="md-dropzone"
        :class="{ 'md-dropzone--active': dragOver }"
        role="button"
        tabindex="0"
        aria-label="Upload ThaiBMA CSV or XLSX file"
        @dragover="onDrag($event, true)"
        @dragleave="onDrag($event, false)"
        @drop="onDrop"
        @click="$emit('browse')"
        @keydown.enter.prevent="$emit('browse')"
        @keydown.space.prevent="$emit('browse')"
      >
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true">
          <path d="M12 16V4M6 10l6-6 6 6M4 20h16" />
        </svg>
        <div class="md-dropzone__title">Drop ThaiBMA CSV / XLSX here</div>
        <div class="md-dropzone__desc">
          Supported columns: ISIN, Name, Coupon, Maturity, Price, YTM, Rating
        </div>
        <button class="btn btn-secondary btn-sm md-dropzone__btn" @click.stop="$emit('browse')">
          Browse files
        </button>
      </div>

      <div v-if="lastUpload" class="md-upload-meta">
        <div class="md-upload-meta__row">
          <span class="md-meta-label">Last upload</span>
          <span class="md-meta-value">{{ lastUpload.fileName }}</span>
        </div>
        <div class="md-upload-meta__row">
          <span class="md-meta-label">Uploaded</span>
          <span class="md-meta-value">{{ lastUpload.uploadedAt }} · {{ lastUpload.uploadedBy }}</span>
        </div>
        <div class="md-upload-meta__row">
          <span class="md-meta-label">Rows</span>
          <span class="md-meta-value">{{ lastUpload.rows }}</span>
        </div>
        <button class="md-link" @click="$emit('viewDiff')">View diff</button>
      </div>
    </div>
  </div>
</template>
