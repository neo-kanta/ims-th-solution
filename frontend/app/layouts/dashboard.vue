<script setup lang="ts">
// Dashboard layout — used for all authenticated pages
// Includes sidebar navigation and top bar

const config = useRuntimeConfig()
const appName = config.public.appName as string

const isSidebarOpen = ref(true)

const navigationItems = [
  { label: 'Dashboard', to: '/', icon: '📊' },
  { label: 'Workflow', to: '/workflow', icon: '🔄' },
  {
    label: 'Investment',
    icon: '📈',
    children: [
      { label: 'Analysis Reports', to: '/investment/analysis' },
      { label: 'Decisions', to: '/investment/decision' },
      { label: 'Execution', to: '/investment/execution' },
      { label: 'Review', to: '/investment/review' },
    ],
  },
  { label: 'Leave & Delegation', to: '/leave', icon: '📋' },
  { label: 'Approval', to: '/approval', icon: '✅' },
  {
    label: 'Permissions',
    icon: '🔐',
    children: [
      { label: 'Accounts', to: '/permissions/accounts' },
      { label: 'Groups', to: '/permissions/groups' },
    ],
  },
  {
    label: 'Settings',
    icon: '⚙️',
    children: [
      { label: 'Notifications', to: '/settings/notifications' },
      { label: 'Audit Log', to: '/settings/audit' },
    ],
  },
]
</script>

<template>
  <div class="flex min-h-screen bg-gray-50">
    <!-- Sidebar -->
    <aside
      :class="[
        'fixed inset-y-0 left-0 z-30 flex w-64 flex-col bg-white border-r border-gray-200 transition-transform duration-300',
        isSidebarOpen ? 'translate-x-0' : '-translate-x-full',
      ]"
    >
      <div class="flex h-16 items-center justify-between border-b border-gray-200 px-6">
        <span class="text-lg font-bold text-blue-600">{{ appName }}</span>
        <button class="text-gray-400 hover:text-gray-600 lg:hidden" @click="isSidebarOpen = false">
          ✕
        </button>
      </div>

      <nav class="flex-1 overflow-y-auto px-4 py-4">
        <template v-for="item in navigationItems" :key="item.label">
          <NuxtLink
            v-if="!item.children"
            :to="item.to"
            class="mb-1 flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 hover:text-gray-900"
            active-class="bg-blue-50 text-blue-700"
          >
            <span>{{ item.icon }}</span>
            <span>{{ item.label }}</span>
          </NuxtLink>

          <div v-else class="mb-2">
            <div class="flex items-center gap-3 px-3 py-2 text-xs font-semibold uppercase tracking-wider text-gray-400">
              <span>{{ item.icon }}</span>
              <span>{{ item.label }}</span>
            </div>
            <NuxtLink
              v-for="child in item.children"
              :key="child.to"
              :to="child.to"
              class="mb-0.5 ml-8 flex items-center rounded-lg px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 hover:text-gray-900"
              active-class="bg-blue-50 text-blue-700"
            >
              {{ child.label }}
            </NuxtLink>
          </div>
        </template>
      </nav>
    </aside>

    <!-- Main content -->
    <div :class="['flex-1 transition-all duration-300', isSidebarOpen ? 'ml-64' : 'ml-0']">
      <!-- Top bar -->
      <header class="sticky top-0 z-20 flex h-16 items-center justify-between border-b border-gray-200 bg-white px-6">
        <button
          class="text-gray-400 hover:text-gray-600"
          @click="isSidebarOpen = !isSidebarOpen"
        >
          ☰
        </button>
        <div class="flex items-center gap-4">
          <span class="text-sm text-gray-500">Welcome</span>
        </div>
      </header>

      <!-- Page content -->
      <main class="p-6">
        <slot />
      </main>
    </div>
  </div>
</template>
