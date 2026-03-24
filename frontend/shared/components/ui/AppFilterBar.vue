<script setup lang="ts">
interface FilterItem {
  key: string
  label: string
}

interface Props {
  filters: FilterItem[]
  modelValue: string
  searchPlaceholder?: string
  showSearch?: boolean
}

withDefaults(defineProps<Props>(), {
  showSearch: true,
  searchPlaceholder: 'Search...',
})

const emit = defineEmits<{
  'update:modelValue': [key: string]
  search: [query: string]
}>()

const searchQuery = ref('')

function selectFilter(key: string) {
  emit('update:modelValue', key)
}

function onSearch(event: Event) {
  const target = event.target as HTMLInputElement
  searchQuery.value = target.value
  emit('search', target.value)
}
</script>

<template>
  <div class="filter-bar">
    <div v-if="showSearch" class="search-wrap">
      <span class="search-icon">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <circle cx="11" cy="11" r="8" />
          <line x1="21" y1="21" x2="16.65" y2="16.65" />
        </svg>
      </span>
      <input
        type="text"
        class="form-input search-input"
        :placeholder="searchPlaceholder"
        :value="searchQuery"
        @input="onSearch"
      />
    </div>

    <div
      v-for="filter in filters"
      :key="filter.key"
      class="filter-chip"
      :class="{ active: modelValue === filter.key }"
      @click="selectFilter(filter.key)"
    >
      {{ filter.label }}
    </div>
  </div>
</template>
