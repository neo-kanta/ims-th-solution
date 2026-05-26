<script setup lang="ts">
const props = defineProps<{ data: number[]; positive?: boolean }>();

const path = computed(() => {
  const d = props.data;
  if (!d || d.length === 0) return "";
  const min = Math.min(...d);
  const max = Math.max(...d);
  const range = max - min || 1;
  const w = 80;
  const h = 24;
  const step = w / (d.length - 1);
  return d
    .map((v, i) => {
      const x = (i * step).toFixed(2);
      const y = (h - ((v - min) / range) * h).toFixed(2);
      return `${i === 0 ? "M" : "L"}${x} ${y}`;
    })
    .join(" ");
});

const isPositive = computed(() => {
  if (props.positive != null) return props.positive;
  const d = props.data;
  if (d.length < 2) return true;
  const first = d[0] ?? 0;
  const last = d[d.length - 1] ?? 0;
  return last >= first;
});
</script>

<template>
  <svg
    class="md-spark"
    :class="isPositive ? 'md-spark--up' : 'md-spark--down'"
    viewBox="0 0 80 24"
    width="80"
    height="24"
    aria-hidden="true"
  >
    <path :d="path" fill="none" stroke="currentColor" stroke-width="1.5" />
  </svg>
</template>
