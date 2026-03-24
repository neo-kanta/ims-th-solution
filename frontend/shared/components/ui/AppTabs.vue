<script setup lang="ts">
interface TabItem {
  key: string
  label: string
  count?: number
}

interface Props {
  tabs: TabItem[]
  modelValue: string
}

defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [key: string]
}>()

function select(key: string) {
  emit('update:modelValue', key)
}
</script>

<template>
  <div class="tab-bar">
    <div
      v-for="tab in tabs"
      :key="tab.key"
      class="tab-item"
      :class="{ active: modelValue === tab.key }"
      @click="select(tab.key)"
    >
      {{ tab.label }}
      <span v-if="tab.count !== undefined" class="tab-count">{{ tab.count }}</span>
    </div>
  </div>
</template>
