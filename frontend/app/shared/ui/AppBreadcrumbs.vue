<script setup lang="ts">
interface BreadcrumbItem {
  label: string;
  to?: string;
}

interface Props {
  items: BreadcrumbItem[];
}

defineProps<Props>();
</script>

<template>
  <nav class="app-breadcrumbs" aria-label="Breadcrumb">
    <ol class="app-breadcrumbs__list">
      <li
        v-for="(item, idx) in items"
        :key="idx"
        class="app-breadcrumbs__item"
      >
        <!-- Router link when to is specified and it is not the last item -->
        <NuxtLink
          v-if="item.to && idx < items.length - 1"
          :to="item.to"
          class="app-breadcrumbs__link"
        >
          {{ item.label }}
        </NuxtLink>
        <span v-else class="app-breadcrumbs__current" aria-current="page">
          {{ item.label }}
        </span>

        <!-- Separator -->
        <span
          v-if="idx < items.length - 1"
          class="app-breadcrumbs__separator"
          aria-hidden="true"
        >
          /
        </span>
      </li>
    </ol>
  </nav>
</template>

<style scoped>
.app-breadcrumbs {
  margin: 0;
  padding: 0;
}

.app-breadcrumbs__list {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2, 8px);
  list-style: none;
  margin: 0;
  padding: 0;
}

.app-breadcrumbs__item {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2, 8px);
  font-size: var(--font-size-xs, 12px);
  line-height: 1;
}

.app-breadcrumbs__link {
  color: var(--text-link, #0969da);
  text-decoration: none;
  font-weight: var(--font-weight-medium, 500);
}

.app-breadcrumbs__link:hover {
  text-decoration: underline;
  color: var(--text-link-hover, #0550ae);
}

.app-breadcrumbs__current {
  color: var(--text-secondary, #57606a);
  font-weight: var(--font-weight-normal, 400);
}

.app-breadcrumbs__separator {
  color: var(--text-tertiary, #6e7781);
  font-size: 10px;
  user-select: none;
}
</style>
