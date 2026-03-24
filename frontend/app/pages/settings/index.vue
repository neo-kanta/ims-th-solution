<script setup lang="ts">
import { ref } from 'vue'

definePageMeta({
  layout: 'dashboard'
})

const sections = ref([
  { title: 'Currencies', count: 156, desc: 'ISO 4217 standard currency codes and rates' },
  { title: 'Exchanges', count: 85, desc: 'Global stock and commodities exchanges' },
  { title: 'Asset Classes', count: 12, desc: 'High-level categorization of tradable assets' },
  { title: 'Counterparties', count: 342, desc: 'Approved brokers and trading partners' }
])
</script>

<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-bold text-slate-900">Reference Data</h1>
        <p class="text-sm text-slate-500 mt-1">Manage global system reference parameters and lookup tables.</p>
      </div>
      <button class="btn btn-primary btn-sm">
        Add Reference
      </button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="section in sections" :key="section.title" class="card p-4 hover:border-blue-500 cursor-pointer transition-colors">
        <h3 class="font-semibold text-slate-800">{{ section.title }}</h3>
        <p class="text-xs text-slate-500 mt-1 mb-3 line-clamp-2 h-8">{{ section.desc }}</p>
        <div class="flex items-center justify-between">
          <span class="text-xs font-medium text-slate-600">{{ section.count }} records</span>
          <span class="text-blue-600">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="5" y1="12" x2="19" y2="12"></line><polyline points="12 5 19 12 12 19"></polyline></svg>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
