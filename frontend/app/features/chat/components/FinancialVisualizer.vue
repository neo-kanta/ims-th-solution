<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "~/composables/useI18n";

const props = defineProps<{
  headers: string[];
  rows: string[][];
}>();

const { locale } = useI18n();

// Simple translator helper for local i18n
function tLocal(en: string, th: string): string {
  return locale.value === "th" ? th : en;
}

// Visualizer mode: 'table' or 'chart'
const activeTab = ref<"table" | "chart">("table");

// ── Search & Sort Logic ─────────────────────────────────────────────────────
const searchText = ref("");
const sortColumn = ref<number | null>(null);
const sortDirection = ref<"asc" | "desc">("asc");

function toggleSort(colIdx: number) {
  if (sortColumn.value === colIdx) {
    sortDirection.value = sortDirection.value === "asc" ? "desc" : "asc";
  } else {
    sortColumn.value = colIdx;
    sortDirection.value = "asc";
  }
}

function parseNumber(val: string): number {
  if (!val) return 0;
  // Strip currency symbols, percentages, commas
  const clean = val.replace(/[$,฿€£%\s]/g, "").replace(/,/g, "");
  const parsed = parseFloat(clean);
  return isNaN(parsed) ? 0 : parsed;
}

// Filter and sort rows
const processedRows = computed(() => {
  let list = props.rows.map((row, idx) => ({ row, originalIndex: idx }));

  // Search filter
  if (searchText.value.trim()) {
    const query = searchText.value.toLowerCase().trim();
    list = list.filter(({ row }) =>
      row.some((cell) => cell.toLowerCase().includes(query))
    );
  }

  // Sorting
  if (sortColumn.value !== null) {
    const colIdx = sortColumn.value;
    const dir = sortDirection.value === "asc" ? 1 : -1;

    list.sort((a, b) => {
      const valA = a.row[colIdx] || "";
      const valB = b.row[colIdx] || "";

      // Try numeric sort first
      const numA = parseNumber(valA);
      const numB = parseNumber(valB);
      
      const isNumA = !isNaN(numA) && valA.trim() !== "" && isNumericString(valA);
      const isNumB = !isNaN(numB) && valB.trim() !== "" && isNumericString(valB);

      if (isNumA && isNumB) {
        return (numA - numB) * dir;
      }

      // Fallback to string sort
      return valA.localeCompare(valB, undefined, { numeric: true }) * dir;
    });
  }

  return list;
});

function isNumericString(str: string): boolean {
  const clean = str.replace(/[$,฿€£%\s]/g, "").replace(/,/g, "").trim();
  return clean !== "" && !isNaN(Number(clean));
}

// ── Column Analyzer for Charting ────────────────────────────────────────────
const numericCols = computed(() => {
  const cols: { index: number; name: string }[] = [];
  if (props.headers.length === 0 || props.rows.length === 0) return cols;

  for (let colIdx = 0; colIdx < props.headers.length; colIdx++) {
    let numCount = 0;
    let nonCheckCount = 0;
    for (let rowIdx = 0; rowIdx < Math.min(props.rows.length, 10); rowIdx++) {
      const cell = props.rows[rowIdx][colIdx];
      if (cell === undefined || cell.trim() === "") continue;
      nonCheckCount++;
      if (isNumericString(cell)) {
        numCount++;
      }
    }
    // If > 70% of rows are numeric, treat it as a values column
    if (nonCheckCount > 0 && numCount / nonCheckCount >= 0.7) {
      cols.push({ index: colIdx, name: props.headers[colIdx] });
    }
  }
  return cols;
});

const labelCols = computed(() => {
  const cols: { index: number; name: string }[] = [];
  props.headers.forEach((h, idx) => {
    // If not numeric, it's a label candidate
    if (!numericCols.value.some((c) => c.index === idx)) {
      cols.push({ index: idx, name: h });
    }
  });
  // Fallback: allow all as label
  if (cols.length === 0) {
    props.headers.forEach((h, idx) => cols.push({ index: idx, name: h }));
  }
  return cols;
});

// Selected columns for charting
const selectedLabelCol = ref<number>(0);
const selectedValueCol = ref<number>(0);

// Set default selected columns once analyzed
const watchTrigger = computed(() => `${numericCols.value.length}:${labelCols.value.length}`);
const initializedCols = ref(false);

const dataAvailableForChart = computed(() => {
  return numericCols.value.length > 0 && props.rows.length > 0;
});

// Set default values when columns are computed
function initChartCols() {
  if (numericCols.value.length > 0) {
    selectedValueCol.value = numericCols.value[0].index;
  }
  if (labelCols.value.length > 0) {
    // Try to find a non-numeric label column, preferring the first column if it's text
    const textLabel = labelCols.value.find((c) => c.index === 0) || labelCols.value[0];
    selectedLabelCol.value = textLabel.index;
  }
}

// Chart type: 'donut' | 'bar' | 'line'
const chartType = ref<"donut" | "bar" | "line">("donut");

// Prepare chart datasets
const chartData = computed(() => {
  if (!dataAvailableForChart.value) return [];
  initChartCols();

  const labelIdx = selectedLabelCol.value;
  const valueIdx = selectedValueCol.value;

  return props.rows
    .map((row) => {
      const rawVal = row[valueIdx] || "";
      const val = parseNumber(rawVal);
      return {
        label: row[labelIdx] || "N/A",
        value: val,
        rawLabel: row[labelIdx],
        rawValue: rawVal,
      };
    })
    .filter((d) => !isNaN(d.value));
});

// Calculate total for percentages
const totalChartValue = computed(() => {
  return chartData.value.reduce((sum, d) => sum + d.value, 0);
});

// Color generator (HSL for balanced colors)
function getChartColor(index: number, totalCount: number): string {
  const hue = (index * (360 / Math.max(totalCount, 1))) % 360;
  return `hsl(${hue}, 70%, 55%)`;
}

// ── Donut Chart Generator ───────────────────────────────────────────────────
interface DonutSlice {
  path: string;
  color: string;
  label: string;
  rawValue: string;
  percentage: number;
  startAngle: number;
  endAngle: number;
}

const donutSlices = computed<DonutSlice[]>(() => {
  const data = chartData.value;
  const total = totalChartValue.value;
  const slices: DonutSlice[] = [];
  if (total === 0 || data.length === 0) return slices;

  const cx = 120;
  const cy = 120;
  const r = 90;
  const innerR = 55;

  let currentAngle = 0;

  data.forEach((d, idx) => {
    const pct = d.value / total;
    const sliceAngle = pct * 360;
    const startAngle = currentAngle;
    const endAngle = currentAngle + sliceAngle;
    currentAngle = endAngle;

    const rad1 = ((startAngle - 90) * Math.PI) / 180;
    const rad2 = ((endAngle - 90) * Math.PI) / 180;

    const x1 = cx + r * Math.cos(rad1);
    const y1 = cy + r * Math.sin(rad1);
    const x2 = cx + r * Math.cos(rad2);
    const y2 = cy + r * Math.sin(rad2);

    const ix1 = cx + innerR * Math.cos(rad1);
    const iy1 = cy + innerR * Math.sin(rad1);
    const ix2 = cx + innerR * Math.cos(rad2);
    const iy2 = cy + innerR * Math.sin(rad2);

    const largeArc = sliceAngle > 180 ? 1 : 0;
    
    let path = "";
    if (sliceAngle >= 359.99) {
      // Draw full circle path
      path = `
        M ${cx} ${cy - r}
        A ${r} ${r} 0 1 1 ${cx} ${cy + r}
        A ${r} ${r} 0 1 1 ${cx} ${cy - r}
        M ${cx} ${cy - innerR}
        A ${innerR} ${innerR} 0 1 0 ${cx} ${cy + innerR}
        A ${innerR} ${innerR} 0 1 0 ${cx} ${cy - innerR}
        Z
      `;
    } else {
      path = `
        M ${x1} ${y1}
        A ${r} ${r} 0 ${largeArc} 1 ${x2} ${y2}
        L ${ix2} ${iy2}
        A ${innerR} ${innerR} 0 ${largeArc} 0 ${ix1} ${iy1}
        Z
      `;
    }

    slices.push({
      path: path.trim(),
      color: getChartColor(idx, data.length),
      label: d.label,
      rawValue: d.rawValue,
      percentage: pct * 100,
      startAngle,
      endAngle,
    });
  });

  return slices;
});

// Hover state
const hoveredIndex = ref<number | null>(null);

// ── Bar Chart Generator ─────────────────────────────────────────────────────
const barChartConfig = computed(() => {
  const data = chartData.value;
  const width = 500;
  const height = 240;
  const leftPadding = 50;
  const rightPadding = 20;
  const topPadding = 20;
  const bottomPadding = 45;

  const plotWidth = width - leftPadding - rightPadding;
  const plotHeight = height - topPadding - bottomPadding;

  const maxVal = Math.max(...data.map((d) => d.value), 1);
  const minVal = Math.min(...data.map((d) => d.value), 0);
  
  // Calculate grid lines (y axis ticks)
  const gridCount = 4;
  const gridLines = Array.from({ length: gridCount + 1 }, (_, i) => {
    const val = (maxVal / gridCount) * i;
    const y = topPadding + plotHeight - (val / maxVal) * plotHeight;
    return { y, label: val.toLocaleString(undefined, { maximumFractionDigits: 1 }) };
  });

  const barWidth = data.length > 0 ? (plotWidth / data.length) * 0.6 : 20;
  const barGap = data.length > 0 ? (plotWidth / data.length) * 0.4 : 10;

  const bars = data.map((d, idx) => {
    const x = leftPadding + idx * (barWidth + barGap) + barGap / 2;
    const barHeight = (d.value / maxVal) * plotHeight;
    const y = topPadding + plotHeight - barHeight;

    return {
      x,
      y,
      width: barWidth,
      height: barHeight,
      color: getChartColor(idx, data.length),
      label: d.label,
      value: d.value,
      rawValue: d.rawValue,
    };
  });

  return { width, height, leftPadding, rightPadding, topPadding, bottomPadding, plotWidth, plotHeight, bars, gridLines };
});

// ── Line Chart Generator ────────────────────────────────────────────────────
const lineChartConfig = computed(() => {
  const data = chartData.value;
  const width = 500;
  const height = 240;
  const leftPadding = 50;
  const rightPadding = 20;
  const topPadding = 20;
  const bottomPadding = 45;

  const plotWidth = width - leftPadding - rightPadding;
  const plotHeight = height - topPadding - bottomPadding;

  const maxVal = Math.max(...data.map((d) => d.value), 1);

  // Generate coordinates
  const points = data.map((d, idx) => {
    const x = leftPadding + (data.length > 1 ? (idx * plotWidth) / (data.length - 1) : plotWidth / 2);
    const y = topPadding + plotHeight - (d.value / maxVal) * plotHeight;
    return { x, y, label: d.label, value: d.value, rawValue: d.rawValue, index: idx };
  });

  // Build SVG Path string
  let linePath = "";
  let areaPath = "";
  if (points.length > 0) {
    linePath = `M ${points[0].x} ${points[0].y}`;
    for (let i = 1; i < points.length; i++) {
      linePath += ` L ${points[i].x} ${points[i].y}`;
    }

    // Area path closes at the bottom axis
    areaPath = `${linePath} L ${points[points.length - 1].x} ${topPadding + plotHeight} L ${points[0].x} ${topPadding + plotHeight} Z`;
  }

  const gridCount = 4;
  const gridLines = Array.from({ length: gridCount + 1 }, (_, i) => {
    const val = (maxVal / gridCount) * i;
    const y = topPadding + plotHeight - (val / maxVal) * plotHeight;
    return { y, label: val.toLocaleString(undefined, { maximumFractionDigits: 1 }) };
  });

  return { width, height, points, linePath, areaPath, gridLines, leftPadding, topPadding, plotHeight, plotWidth };
});

// ── Clipboard & CSV Actions ─────────────────────────────────────────────────
const copySuccess = ref(false);

function copyTable() {
  const headersStr = props.headers.join("\t");
  const rowsStr = props.rows.map((row) => row.join("\t")).join("\n");
  const fullText = `${headersStr}\n${rowsStr}`;

  navigator.clipboard.writeText(fullText).then(() => {
    copySuccess.value = true;
    setTimeout(() => {
      copySuccess.value = false;
    }, 2000);
  });
}

function exportCSV() {
  const headersStr = props.headers.map((h) => `"${h.replace(/"/g, '""')}"`).join(",");
  const rowsStr = props.rows
    .map((row) => row.map((cell) => `"${cell.replace(/"/g, '""')}"`).join(","))
    .join("\n");
  const csvContent = "\uFEFF" + `${headersStr}\n${rowsStr}`; // include UTF-8 BOM

  const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.setAttribute("href", url);
  link.setAttribute("download", `ims_data_export_${new Date().getTime()}.csv`);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}
</script>

<template>
  <div class="visualizer">
    <!-- Visualizer Tabs Header -->
    <div class="visualizer__tabs">
      <div class="visualizer__tab-buttons">
        <button
          type="button"
          :class="['visualizer__tab', { 'is-active': activeTab === 'table' }]"
          @click="activeTab = 'table'"
        >
          📄 {{ tLocal("Table View", "มุมมองตาราง") }}
        </button>
        <button
          v-if="dataAvailableForChart"
          type="button"
          :class="['visualizer__tab', { 'is-active': activeTab === 'chart' }]"
          @click="activeTab = 'chart'"
        >
          📊 {{ tLocal("Chart View", "มุมมองแผนภูมิ") }}
        </button>
      </div>

      <!-- Action buttons -->
      <div class="visualizer__actions">
        <button
          type="button"
          class="visualizer__btn"
          @click="copyTable"
          :title="tLocal('Copy to Clipboard', 'คัดลอกลงคลิปบอร์ด')"
        >
          {{ copySuccess ? tLocal("Copied!", "คัดลอกแล้ว!") : tLocal("Copy", "คัดลอก") }}
        </button>
        <button
          type="button"
          class="visualizer__btn"
          @click="exportCSV"
          :title="tLocal('Export as CSV', 'ส่งออกเป็น CSV')"
        >
          CSV
        </button>
      </div>
    </div>

    <!-- TAB 1: TABLE VIEW -->
    <div v-show="activeTab === 'table'" class="visualizer__table-container">
      <div class="visualizer__table-toolbar">
        <input
          v-model="searchText"
          type="text"
          :placeholder="tLocal('Search table...', 'ค้นหาในตาราง...')"
          class="visualizer__search-input"
        />
      </div>

      <div class="visualizer__scrollable">
        <table class="visualizer__table">
          <thead>
            <tr>
              <th
                v-for="(h, colIdx) in headers"
                :key="`th-${colIdx}`"
                @click="toggleSort(colIdx)"
                :class="['visualizer__th', { 'is-sorted': sortColumn === colIdx }]"
              >
                <div class="visualizer__th-content">
                  <span>{{ h }}</span>
                  <span v-if="sortColumn === colIdx" class="visualizer__sort-arrow">
                    {{ sortDirection === "asc" ? "▲" : "▼" }}
                  </span>
                  <span v-else class="visualizer__sort-arrow-placeholder">⇅</span>
                </div>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="{ row, originalIndex } in processedRows"
              :key="`tr-${originalIndex}`"
              class="visualizer__tr"
            >
              <td
                v-for="(cell, cellIdx) in row"
                :key="`td-${originalIndex}-${cellIdx}`"
                :class="['visualizer__td', { 'is-numeric': isNumericString(cell) }]"
              >
                {{ cell }}
              </td>
            </tr>
            <tr v-if="processedRows.length === 0">
              <td :colspan="headers.length" class="visualizer__empty-cell">
                {{ tLocal("No matching rows found.", "ไม่พบข้อมูลที่ตรงกัน") }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- TAB 2: CHART VIEW -->
    <div v-show="activeTab === 'chart' && dataAvailableForChart" class="visualizer__chart-container">
      <div class="visualizer__chart-selectors">
        <!-- Chart type buttons -->
        <div class="visualizer__chart-types">
          <button
            type="button"
            :class="['visualizer__type-btn', { 'is-active': chartType === 'donut' }]"
            @click="chartType = 'donut'"
          >
            🍩 Donut
          </button>
          <button
            type="button"
            :class="['visualizer__type-btn', { 'is-active': chartType === 'bar' }]"
            @click="chartType = 'bar'"
          >
            📊 Bar
          </button>
          <button
            type="button"
            :class="['visualizer__type-btn', { 'is-active': chartType === 'line' }]"
            @click="chartType = 'line'"
          >
            📈 Line
          </button>
        </div>

        <!-- Axis selectors -->
        <div class="visualizer__axis-controls">
          <div class="visualizer__axis-select">
            <label class="visualizer__axis-label">{{ tLocal("Labels:", "ป้ายกำกับ:") }}</label>
            <select v-model="selectedLabelCol" class="visualizer__select">
              <option v-for="c in labelCols" :key="`lbl-${c.index}`" :value="c.index">
                {{ c.name }}
              </option>
            </select>
          </div>
          <div class="visualizer__axis-select">
            <label class="visualizer__axis-label">{{ tLocal("Values:", "ค่าข้อมูล:") }}</label>
            <select v-model="selectedValueCol" class="visualizer__select">
              <option v-for="c in numericCols" :key="`val-${c.index}`" :value="c.index">
                {{ c.name }}
              </option>
            </select>
          </div>
        </div>
      </div>

      <!-- Actual Charts Drawings -->
      <div class="visualizer__chart-wrapper">
        <!-- 1. Donut / Pie Chart -->
        <div v-if="chartType === 'donut'" class="donut-chart-layout">
          <div class="donut-chart-svg-wrap">
            <svg viewBox="0 0 240 240" class="donut-chart-svg">
              <g class="donut-slices">
                <path
                  v-for="(slice, sIdx) in donutSlices"
                  :key="`slice-${sIdx}`"
                  :d="slice.path"
                  :fill="slice.color"
                  class="donut-slice"
                  :class="{ 'is-hovered': hoveredIndex === sIdx }"
                  @mouseenter="hoveredIndex = sIdx"
                  @mouseleave="hoveredIndex = null"
                />
              </g>
            </svg>
            <div class="donut-center-label">
              <template v-if="hoveredIndex !== null && donutSlices[hoveredIndex]">
                <div class="donut-center-title">{{ donutSlices[hoveredIndex].label }}</div>
                <div class="donut-center-value">{{ donutSlices[hoveredIndex].rawValue }}</div>
                <div class="donut-center-pct">
                  {{ donutSlices[hoveredIndex].percentage.toFixed(1) }}%
                </div>
              </template>
              <template v-else>
                <div class="donut-center-title">{{ tLocal("Total", "ยอดรวม") }}</div>
                <div class="donut-center-value">
                  {{ totalChartValue.toLocaleString(undefined, { maximumFractionDigits: 2 }) }}
                </div>
              </template>
            </div>
          </div>

          <!-- Legend list -->
          <div class="donut-legend">
            <div
              v-for="(slice, sIdx) in donutSlices"
              :key="`leg-${sIdx}`"
              class="legend-item"
              :class="{ 'is-hovered': hoveredIndex === sIdx }"
              @mouseenter="hoveredIndex = sIdx"
              @mouseleave="hoveredIndex = null"
            >
              <span class="legend-dot" :style="{ backgroundColor: slice.color }" />
              <span class="legend-text" :title="slice.label">{{ slice.label }}</span>
              <span class="legend-value">{{ slice.rawValue }}</span>
              <span class="legend-pct">{{ slice.percentage.toFixed(1) }}%</span>
            </div>
          </div>
        </div>

        <!-- 2. Bar Chart -->
        <div v-else-if="chartType === 'bar'" class="bar-chart-layout">
          <svg
            :viewBox="`0 0 ${barChartConfig.width} ${barChartConfig.height}`"
            class="bar-chart-svg"
          >
            <!-- Grid lines -->
            <g class="chart-grids">
              <line
                v-for="(grid, gIdx) in barChartConfig.gridLines"
                :key="`grid-${gIdx}`"
                :x1="barChartConfig.leftPadding"
                :y1="grid.y"
                :x2="barChartConfig.width - barChartConfig.rightPadding"
                :y2="grid.y"
                stroke="var(--border-subtle, rgba(226, 232, 240, 0.1))"
                stroke-width="1"
                stroke-dasharray="3,3"
              />
              <text
                v-for="(grid, gIdx) in barChartConfig.gridLines"
                :key="`grid-lbl-${gIdx}`"
                :x="barChartConfig.leftPadding - 8"
                :y="grid.y + 4"
                text-anchor="end"
                class="chart-axis-text"
              >
                {{ grid.label }}
              </text>
            </g>

            <!-- Bars -->
            <g class="chart-bars">
              <rect
                v-for="(bar, bIdx) in barChartConfig.bars"
                :key="`bar-${bIdx}`"
                :x="bar.x"
                :y="bar.y"
                :width="bar.width"
                :height="bar.height"
                :fill="bar.color"
                class="bar-rect"
                :class="{ 'is-hovered': hoveredIndex === bIdx }"
                @mouseenter="hoveredIndex = bIdx"
                @mouseleave="hoveredIndex = null"
              />
            </g>

            <!-- Base axis line -->
            <line
              :x1="barChartConfig.leftPadding"
              :y1="barChartConfig.height - barChartConfig.bottomPadding"
              :x2="barChartConfig.width - barChartConfig.rightPadding"
              :y2="barChartConfig.height - barChartConfig.bottomPadding"
              stroke="var(--border-default, #94a3b8)"
              stroke-width="1"
            />

            <!-- X Axis Labels (if small list) -->
            <g v-if="barChartConfig.bars.length <= 10" class="chart-x-labels">
              <text
                v-for="(bar, bIdx) in barChartConfig.bars"
                :key="`x-lbl-${bIdx}`"
                :x="bar.x + bar.width / 2"
                :y="barChartConfig.height - barChartConfig.bottomPadding + 14"
                text-anchor="middle"
                class="chart-axis-text chart-axis-text--x"
              >
                {{ bar.label.length > 8 ? bar.label.slice(0, 7) + '..' : bar.label }}
              </text>
            </g>
          </svg>

          <!-- Interactive Tooltip details underneath -->
          <div class="chart-tooltip-detail">
            <template v-if="hoveredIndex !== null && barChartConfig.bars[hoveredIndex]">
              <span
                class="legend-dot"
                :style="{ backgroundColor: barChartConfig.bars[hoveredIndex].color }"
              />
              <strong class="tooltip-detail-label">{{ barChartConfig.bars[hoveredIndex].label }}</strong>
              <span class="tooltip-detail-value">{{ barChartConfig.bars[hoveredIndex].rawValue }}</span>
            </template>
            <template v-else>
              <span class="tooltip-detail-hint">
                {{ tLocal("Hover over bars to inspect values", "วางเมาส์เหนือแท่งเพื่อดูรายละเอียด") }}
              </span>
            </template>
          </div>
        </div>

        <!-- 3. Line Chart -->
        <div v-else-if="chartType === 'line'" class="line-chart-layout">
          <svg
            :viewBox="`0 0 ${lineChartConfig.width} ${lineChartConfig.height}`"
            class="line-chart-svg"
          >
            <!-- Grid lines -->
            <g class="chart-grids">
              <line
                v-for="(grid, gIdx) in lineChartConfig.gridLines"
                :key="`grid-${gIdx}`"
                :x1="lineChartConfig.leftPadding"
                :y1="grid.y"
                :x2="lineChartConfig.width - lineChartConfig.rightPadding"
                :y2="grid.y"
                stroke="var(--border-subtle, rgba(226, 232, 240, 0.1))"
                stroke-width="1"
                stroke-dasharray="3,3"
              />
              <text
                v-for="(grid, gIdx) in lineChartConfig.gridLines"
                :key="`grid-lbl-${gIdx}`"
                :x="lineChartConfig.leftPadding - 8"
                :y="grid.y + 4"
                text-anchor="end"
                class="chart-axis-text"
              >
                {{ grid.label }}
              </text>
            </g>

            <!-- Line gradients -->
            <defs>
              <linearGradient id="lineAreaGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="var(--action-primary, #3b82f6)" stop-opacity="0.3" />
                <stop offset="100%" stop-color="var(--action-primary, #3b82f6)" stop-opacity="0.0" />
              </linearGradient>
            </defs>

            <!-- Fill Area under the line -->
            <path
              v-if="lineChartConfig.areaPath"
              :d="lineChartConfig.areaPath"
              fill="url(#lineAreaGrad)"
            />

            <!-- Line path -->
            <path
              v-if="lineChartConfig.linePath"
              :d="lineChartConfig.linePath"
              fill="none"
              stroke="var(--action-primary, #3b82f6)"
              stroke-width="2.5"
              stroke-linecap="round"
              stroke-linejoin="round"
            />

            <!-- Dots -->
            <g class="chart-line-points">
              <circle
                v-for="(pt, pIdx) in lineChartConfig.points"
                :key="`pt-${pIdx}`"
                :cx="pt.x"
                :cy="pt.y"
                :r="hoveredIndex === pIdx ? 6 : 4"
                :fill="hoveredIndex === pIdx ? '#ffffff' : 'var(--action-primary, #3b82f6)'"
                :stroke="hoveredIndex === pIdx ? 'var(--action-primary, #3b82f6)' : '#ffffff'"
                :stroke-width="hoveredIndex === pIdx ? 3 : 1.5"
                class="line-point-dot"
                @mouseenter="hoveredIndex = pIdx"
                @mouseleave="hoveredIndex = null"
              />
            </g>

            <!-- Base axis line -->
            <line
              :x1="lineChartConfig.leftPadding"
              :y1="lineChartConfig.height - lineChartConfig.bottomPadding"
              :x2="lineChartConfig.width - lineChartConfig.rightPadding"
              :y2="lineChartConfig.height - lineChartConfig.bottomPadding"
              stroke="var(--border-default, #94a3b8)"
              stroke-width="1"
            />

            <!-- X Axis Labels -->
            <g v-if="lineChartConfig.points.length <= 8" class="chart-x-labels">
              <text
                v-for="(pt, pIdx) in lineChartConfig.points"
                :key="`x-lbl-${pIdx}`"
                :x="pt.x"
                :y="lineChartConfig.height - lineChartConfig.bottomPadding + 14"
                text-anchor="middle"
                class="chart-axis-text chart-axis-text--x"
              >
                {{ pt.label.length > 8 ? pt.label.slice(0, 7) + '..' : pt.label }}
              </text>
            </g>
          </svg>

          <!-- Interactive Tooltip details underneath -->
          <div class="chart-tooltip-detail">
            <template v-if="hoveredIndex !== null && lineChartConfig.points[hoveredIndex]">
              <strong class="tooltip-detail-label">
                {{ lineChartConfig.points[hoveredIndex].label }}
              </strong>
              <span class="tooltip-detail-value">
                {{ lineChartConfig.points[hoveredIndex].rawValue }}
              </span>
            </template>
            <template v-else>
              <span class="tooltip-detail-hint">
                {{ tLocal("Hover over nodes to inspect values", "วางเมาส์เหนือจุดข้อมูลเพื่อดูรายละเอียด") }}
              </span>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.visualizer {
  border: 1px solid var(--border-subtle, #e2e8f0);
  border-radius: 10px;
  background: var(--bg-elevated, #ffffff);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.04);
  margin: 12px 0;
  width: 100%;
  overflow: hidden;
}

/* Tabs Header */
.visualizer__tabs {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--bg-canvas, #f8fafc);
  border-bottom: 1px solid var(--border-subtle, #e2e8f0);
  padding: 4px 12px;
}

.visualizer__tab-buttons {
  display: flex;
  gap: 6px;
}

.visualizer__tab {
  border: none;
  background: transparent;
  color: var(--text-secondary, #64748b);
  padding: 8px 12px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.15s ease;
}

.visualizer__tab:hover {
  background: var(--bg-hover, rgba(0, 0, 0, 0.04));
  color: var(--text-primary, #0f172a);
}

.visualizer__tab.is-active {
  background: var(--bg-elevated, #ffffff);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.05);
  color: var(--text-accent, #3b82f6);
}

.visualizer__actions {
  display: flex;
  gap: 6px;
}

.visualizer__btn {
  background: transparent;
  border: 1px solid var(--border-default, #cbd5e1);
  color: var(--text-primary, #0f172a);
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.visualizer__btn:hover {
  border-color: var(--text-accent, #3b82f6);
  color: var(--text-accent, #3b82f6);
  background: rgba(59, 130, 246, 0.05);
}

/* Toolbar & search */
.visualizer__table-toolbar {
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-subtle, #e2e8f0);
}

.visualizer__search-input {
  width: 100%;
  max-width: 260px;
  padding: 6px 10px;
  border-radius: 6px;
  border: 1px solid var(--border-default, #cbd5e1);
  font-size: 12px;
  background: var(--bg-canvas, #ffffff);
  color: var(--text-primary, #0f172a);
  outline: none;
  transition: border-color 0.15s ease;
}

.visualizer__search-input:focus {
  border-color: var(--text-accent, #3b82f6);
}

/* Table elements */
.visualizer__scrollable {
  max-height: 280px;
  overflow-y: auto;
  overflow-x: auto;
}

.visualizer__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
  text-align: left;
}

.visualizer__th {
  position: sticky;
  top: 0;
  background: var(--bg-canvas, #f8fafc);
  z-index: 10;
  padding: 10px 14px;
  font-weight: 600;
  color: var(--text-secondary, #475569);
  cursor: pointer;
  user-select: none;
  border-bottom: 1px solid var(--border-subtle, #e2e8f0);
}

.visualizer__th:hover {
  background: var(--bg-hover, rgba(0, 0, 0, 0.06));
  color: var(--text-primary, #0f172a);
}

.visualizer__th-content {
  display: flex;
  align-items: center;
  gap: 6px;
}

.visualizer__sort-arrow {
  color: var(--text-accent, #3b82f6);
  font-size: 10px;
}

.visualizer__sort-arrow-placeholder {
  opacity: 0.3;
  font-size: 10px;
}

.visualizer__tr {
  border-bottom: 1px solid var(--border-subtle, #f1f5f9);
  transition: background-color 0.1s ease;
}

.visualizer__tr:hover {
  background: var(--bg-hover, #f8fafc);
}

.visualizer__td {
  padding: 10px 14px;
  color: var(--text-primary, #1e293b);
  white-space: nowrap;
}

.visualizer__td.is-numeric {
  text-align: right;
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 11px;
}

.visualizer__empty-cell {
  text-align: center;
  padding: 24px;
  color: var(--text-secondary, #64748b);
  font-style: italic;
}

/* Charts View styling */
.visualizer__chart-container {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.visualizer__chart-selectors {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.visualizer__chart-types {
  display: flex;
  background: var(--bg-canvas, #f1f5f9);
  padding: 3px;
  border-radius: 6px;
  gap: 2px;
}

.visualizer__type-btn {
  border: none;
  background: transparent;
  padding: 6px 12px;
  font-size: 11px;
  font-weight: 600;
  border-radius: 4px;
  color: var(--text-secondary, #475569);
  cursor: pointer;
  transition: all 0.15s ease;
}

.visualizer__type-btn.is-active {
  background: var(--bg-elevated, #ffffff);
  color: var(--text-accent, #3b82f6);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}

.visualizer__axis-controls {
  display: flex;
  gap: 12px;
}

.visualizer__axis-select {
  display: flex;
  align-items: center;
  gap: 6px;
}

.visualizer__axis-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary, #64748b);
}

.visualizer__select {
  padding: 4px 8px;
  font-size: 11px;
  border-radius: 4px;
  border: 1px solid var(--border-default, #cbd5e1);
  background: var(--bg-canvas, #ffffff);
  color: var(--text-primary, #0f172a);
  outline: none;
}

.visualizer__chart-wrapper {
  background: var(--bg-canvas, #f8fafc);
  border-radius: 8px;
  border: 1px solid var(--border-subtle, #f1f5f9);
  padding: 16px;
  min-height: 250px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Donut Chart styles */
.donut-chart-layout {
  display: flex;
  align-items: center;
  justify-content: space-around;
  width: 100%;
  gap: 24px;
  flex-wrap: wrap;
}

.donut-chart-svg-wrap {
  position: relative;
  width: 180px;
  height: 180px;
  flex-shrink: 0;
}

.donut-chart-svg {
  width: 100%;
  height: 100%;
}

.donut-slice {
  cursor: pointer;
  opacity: 0.85;
  transition: transform 0.15s ease, opacity 0.15s ease;
  transform-origin: 120px 120px;
}

.donut-slice:hover,
.donut-slice.is-hovered {
  opacity: 1;
  transform: scale(1.04);
}

.donut-center-label {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  pointer-events: none;
  width: 90px;
}

.donut-center-title {
  font-size: 10px;
  text-transform: uppercase;
  color: var(--text-secondary, #64748b);
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.donut-center-value {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-primary, #0f172a);
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.donut-center-pct {
  font-size: 10px;
  font-weight: 600;
  color: var(--text-accent, #3b82f6);
  margin-top: 1px;
}

.donut-legend {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 200px;
  overflow-y: auto;
  flex-grow: 1;
  max-width: 260px;
  min-width: 180px;
}

.legend-item {
  display: flex;
  align-items: center;
  font-size: 11px;
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.1s ease;
}

.legend-item:hover,
.legend-item.is-hovered {
  background: var(--bg-hover, rgba(0, 0, 0, 0.04));
}

.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 8px;
  flex-shrink: 0;
}

.legend-text {
  flex-grow: 1;
  color: var(--text-secondary, #475569);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-right: 12px;
}

.legend-value {
  font-weight: 600;
  color: var(--text-primary, #1e293b);
  margin-right: 8px;
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 10.5px;
}

.legend-pct {
  color: var(--text-secondary, #64748b);
  font-size: 10px;
  min-width: 36px;
  text-align: right;
}

/* Bar Chart styles */
.bar-chart-layout,
.line-chart-layout {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
}

.bar-chart-svg,
.line-chart-svg {
  width: 100%;
  max-width: 500px;
  height: auto;
}

.bar-rect {
  cursor: pointer;
  opacity: 0.85;
  transition: opacity 0.15s ease, y 0.15s ease, height 0.15s ease;
  rx: 3px;
}

.bar-rect:hover,
.bar-rect.is-hovered {
  opacity: 1;
}

.chart-axis-text {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 9px;
  fill: var(--text-secondary, #64748b);
}

.chart-axis-text--x {
  font-family: var(--font-family-sans, sans-serif);
  font-size: 8.5px;
}

.chart-tooltip-detail {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  background: var(--bg-elevated, #ffffff);
  border: 1px solid var(--border-subtle, #e2e8f0);
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 11px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.tooltip-detail-label {
  color: var(--text-primary, #0f172a);
}

.tooltip-detail-value {
  color: var(--text-accent, #3b82f6);
  font-weight: 700;
  font-family: var(--font-mono, ui-monospace, monospace);
}

.tooltip-detail-hint {
  color: var(--text-secondary, #64748b);
  font-style: italic;
}

/* Line Chart Point Dot */
.line-point-dot {
  cursor: pointer;
  transition: r 0.15s ease, stroke-width 0.15s ease;
}

/* Dark mode adjustments (handled globally by Nuxt but explicit for SVG elements) */
.dark-theme .visualizer__table-toolbar,
.dark-theme .visualizer__tabs {
  background: #1e1e1e;
}
.dark-theme .visualizer {
  border-color: #2e2e2e;
}
</style>
