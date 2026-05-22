<script setup lang="ts">
import { computed, ref, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import type { NavHistoryPayload, NavHistoryPoint } from "../types";

interface Props {
  payload: NavHistoryPayload | null;
  viewW?: number;
  viewH?: number;
  padTop?: number;
  padBottom?: number;
  padLeft?: number;
  padRight?: number;
}

const props = withDefaults(defineProps<Props>(), {
  viewW: 500,
  viewH: 180,
  padTop: 15,
  padBottom: 25,
  padLeft: 15,
  padRight: 50,
});

const emit = defineEmits<{
  hover: [point: NavHistoryPoint];
  leave: [];
}>();

const { t } = useI18n();

const chartWidth = computed(() => props.viewW - props.padLeft - props.padRight);
const chartHeight = computed(() => props.viewH - props.padTop - props.padBottom);

const minMax = computed(() => {
  const series = props.payload?.series ?? [];
  if (series.length < 2) return { min: 9.9, max: 10.1, span: 0.2 };
  const ys = series.map((p) => p.nav_per_unit);
  const min = Math.min(...ys);
  const max = Math.max(...ys);
  const span = max - min || 0.1;
  const padding = span * 0.08;
  return {
    min: min - padding,
    max: max + padding,
    span: span + padding * 2,
  };
});

const points = computed(() => {
  const series = props.payload?.series ?? [];
  if (series.length < 2) return [];
  const { min, span } = minMax.value;
  const w = chartWidth.value;
  const h = chartHeight.value;
  const stepX = w / (series.length - 1);
  return series.map((p, i) => {
    const x = props.padLeft + i * stepX;
    const y = props.padTop + h * (1 - (p.nav_per_unit - min) / span);
    return { x, y, val: p.nav_per_unit, date: p.business_date };
  });
});

const path = computed(() => {
  const pts = points.value;
  if (pts.length < 2) return "";
  return pts.map((p, i) => `${i === 0 ? "M" : "L"} ${p.x.toFixed(2)} ${p.y.toFixed(2)}`).join(" ");
});

const fillPath = computed(() => {
  const pts = points.value;
  if (pts.length < 2) return "";
  const first = pts[0];
  const last = pts[pts.length - 1];
  const bottomY = props.viewH - props.padBottom;
  return `M ${first.x.toFixed(2)} ${bottomY.toFixed(2)} ${path.value} L ${last.x.toFixed(2)} ${bottomY.toFixed(2)} Z`;
});

// Gridlines
const yGridLines = computed(() => {
  const { min, span } = minMax.value;
  const lines = [];
  const count = 4;
  const h = chartHeight.value;
  for (let i = 0; i <= count; i++) {
    const val = min + (span * i) / count;
    const y = props.padTop + h * (1 - i / count);
    lines.push({ y, val });
  }
  return lines;
});

const xGridLines = computed(() => {
  const pts = points.value;
  if (pts.length < 2) return [];
  const lines = [];
  const count = 5;
  const stepIdx = Math.floor((pts.length - 1) / (count - 1));
  for (let i = 0; i < count; i++) {
    const idx = Math.min(pts.length - 1, i * stepIdx);
    const p = pts[idx];
    if (p) {
      lines.push({ x: p.x, label: formatDateLabel(p.date) });
    }
  }
  return lines;
});

function formatDateLabel(dateStr: string): string {
  try {
    const parts = dateStr.split("-");
    if (parts.length === 3) {
      const months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
      const mIdx = parseInt(parts[1], 10) - 1;
      return `${months[mIdx]} ${parts[2]}`;
    }
  } catch (e) {}
  return dateStr;
}

// Interaction States
const chartRef = ref<SVGElement | null>(null);
const hoverIndex = ref<number | null>(null);

const activePoint = computed(() => {
  const series = props.payload?.series ?? [];
  if (hoverIndex.value !== null && series[hoverIndex.value]) {
    return series[hoverIndex.value];
  }
  return null;
});

const activePointCoords = computed(() => {
  const pts = points.value;
  if (hoverIndex.value !== null && pts[hoverIndex.value]) {
    return pts[hoverIndex.value];
  }
  return null;
});

const isPositive = computed(() => (props.payload?.delta_pct ?? 0) >= 0);

watch(activePoint, (p) => {
  if (p) {
    emit("hover", p);
  } else {
    emit("leave");
  }
});

function handleMouseMove(e: MouseEvent) {
  const pts = points.value;
  if (!pts.length || !chartRef.value) return;

  const rect = chartRef.value.getBoundingClientRect();
  const mouseX = e.clientX - rect.left;
  const pctX = mouseX / rect.width;
  
  const idx = Math.min(
    pts.length - 1,
    Math.max(0, Math.round(pctX * (pts.length - 1)))
  );
  hoverIndex.value = idx;
}

function handleMouseLeave() {
  hoverIndex.value = null;
}
</script>

<template>
  <div 
    class="ht-nav__chart-container"
    @mousemove="handleMouseMove"
    @mouseleave="handleMouseLeave"
  >
    <svg
      ref="chartRef"
      class="ht-nav__chart"
      :viewBox="`0 0 ${viewW} ${viewH}`"
      preserveAspectRatio="none"
      role="img"
      :aria-label="t('holdings.navHistory.title', 'Unit NAV history')"
    >
      <defs>
        <linearGradient id="gradient-success" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#10b981" stop-opacity="0.25" />
          <stop offset="100%" stop-color="#10b981" stop-opacity="0.0" />
        </linearGradient>
        <linearGradient id="gradient-danger" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#ef4444" stop-opacity="0.25" />
          <stop offset="100%" stop-color="#ef4444" stop-opacity="0.0" />
        </linearGradient>
      </defs>

      <!-- Horizontal Grid Lines -->
      <g class="ht-nav__grid-y">
        <line 
          v-for="(line, idx) in yGridLines" 
          :key="idx" 
          :x1="padLeft" 
          :y1="line.y" 
          :x2="viewW - padRight" 
          :y2="line.y" 
          stroke="var(--border-subtle)"
          stroke-dasharray="2,2"
          stroke-width="1"
        />
      </g>

      <!-- Y-Axis Labels -->
      <g class="ht-nav__axis-y-labels">
        <text 
          v-for="(line, idx) in yGridLines" 
          :key="idx" 
          :x="viewW - padRight + 6" 
          :y="line.y + 3"
        >
          {{ line.val.toFixed(2) }}
        </text>
      </g>

      <!-- Vertical Grid Lines -->
      <g class="ht-nav__grid-x">
        <line 
          v-for="(line, idx) in xGridLines" 
          :key="idx" 
          :x1="line.x" 
          :y1="padTop" 
          :x2="line.x" 
          :y2="viewH - padBottom" 
          stroke="var(--border-subtle)"
          stroke-dasharray="2,2"
          stroke-width="1"
        />
      </g>

      <!-- X-Axis Labels -->
      <g class="ht-nav__axis-x-labels">
        <text 
          v-for="(line, idx) in xGridLines" 
          :key="idx" 
          :x="line.x" 
          :y="viewH - 6"
          text-anchor="middle"
        >
          {{ line.label }}
        </text>
      </g>

      <!-- Chart Paths -->
      <path v-if="fillPath" :d="fillPath" class="ht-nav__chart-fill" :class="{ 'is-negative': !isPositive }" />
      <path v-if="path" :d="path" class="ht-nav__chart-line" :class="{ 'is-negative': !isPositive }" />

      <!-- Crosshair & Tracker -->
      <g v-if="activePointCoords">
        <line 
          :x1="activePointCoords.x" 
          :y1="padTop" 
          :x2="activePointCoords.x" 
          :y2="viewH - padBottom" 
          class="ht-nav__crosshair" 
          stroke="var(--border-strong)"
          stroke-dasharray="2,2"
          stroke-width="1"
        />
        <line 
          :x1="padLeft" 
          :y1="activePointCoords.y" 
          :x2="viewW - padRight" 
          :y2="activePointCoords.y" 
          class="ht-nav__crosshair" 
          stroke="var(--border-strong)"
          stroke-dasharray="2,2"
          stroke-width="1"
        />
        <circle 
          :cx="activePointCoords.x" 
          :cy="activePointCoords.y" 
          r="4" 
          class="ht-nav__tracker-dot" 
          :class="{ 'is-negative': !isPositive }"
        />
        <circle 
          :cx="activePointCoords.x" 
          :cy="activePointCoords.y" 
          r="8" 
          class="ht-nav__tracker-pulse" 
          :class="{ 'is-negative': !isPositive }"
        />
      </g>
    </svg>

    <!-- Tooltip -->
    <div 
      v-if="activePointCoords" 
      class="ht-nav__tooltip"
      :style="{
        left: `calc(${(activePointCoords.x / viewW) * 100}% - 40px)`,
        top: `calc(${(activePointCoords.y / viewH) * 100}% - 45px)`
      }"
    >
      <div class="ht-nav__tooltip-date">{{ formatDateLabel(activePointCoords.date) }}</div>
      <div class="ht-nav__tooltip-value">{{ activePointCoords.val.toFixed(4) }}</div>
    </div>
  </div>
</template>

<style scoped>
.ht-nav__chart-container {
  position: relative;
  width: 100%;
}

.ht-nav__chart {
  width: 100%;
  height: 100%;
  min-height: inherit;
  background: var(--bg-input);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  display: block;
}

.ht-nav__axis-y-labels text,
.ht-nav__axis-x-labels text {
  font-size: 8px;
  fill: var(--text-placeholder);
  font-family: var(--font-sans, Inter, sans-serif);
  font-weight: 500;
}

.ht-nav__chart-line {
  fill: none;
  stroke: #10b981;
  stroke-width: 1.5;
  stroke-linejoin: round;
  stroke-linecap: round;
  vector-effect: non-scaling-stroke;
}

.ht-nav__chart-line.is-negative {
  stroke: #ef4444;
}

.ht-nav__chart-fill {
  fill: url(#gradient-success);
  stroke: none;
}

.ht-nav__chart-fill.is-negative {
  fill: url(#gradient-danger);
}

.ht-nav__tracker-dot {
  fill: #10b981;
  stroke: #fff;
  stroke-width: 1.5px;
}

.ht-nav__tracker-dot.is-negative {
  fill: #ef4444;
}

.ht-nav__tracker-pulse {
  fill: #10b981;
  fill-opacity: 0.2;
}

.ht-nav__tracker-pulse.is-negative {
  fill: #ef4444;
}

.ht-nav__tooltip {
  position: absolute;
  pointer-events: none;
  background: var(--bg-overlay);
  color: #fff;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 10px;
  box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
  z-index: 10;
  width: 80px;
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.ht-nav__tooltip-date {
  color: var(--text-secondary);
  font-size: 9px;
}

.ht-nav__tooltip-value {
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}
</style>
