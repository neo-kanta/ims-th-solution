<script setup lang="ts">
/**
 * Small generic searchable combobox for the Breaches toolbar.
 *
 * Filters a client-side option list (portfolio directory) by label/sublabel
 * so the operator never has to type or paste a raw UUID — only the resolved
 * `value` (a business identifier) is ever emitted.
 */
import { computed, nextTick, ref, watch } from "vue";

export interface ComplianceComboboxOption {
  value: string;
  label: string;
  sublabel?: string;
}

const props = withDefaults(
  defineProps<{
    modelValue: string;
    options: ComplianceComboboxOption[];
    placeholder: string;
    /**
     * Accessible name for the input. Named without an `aria-` prefix on
     * purpose: Vue never coerces `aria-*`/`data-*` template attributes into
     * matching camelCase props (they always fall through as literal DOM
     * attributes), so a prop named `ariaLabel` could never actually receive
     * a `:aria-label="..."` binding from a parent.
     */
    fieldLabel: string;
    inputId: string;
    disabled?: boolean;
    loading?: boolean;
    loadingText?: string;
    emptyText?: string;
    changeLabel?: string;
  }>(),
  {
    disabled: false,
    loading: false,
    loadingText: "Loading…",
    emptyText: "No matches.",
    changeLabel: "Change",
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const query = ref("");
const isOpen = ref(false);
const activeIndex = ref(-1);
const inputEl = ref<HTMLInputElement | null>(null);

const selected = computed(
  () => props.options.find((o) => o.value === props.modelValue) ?? null,
);

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase();
  const list = q
    ? props.options.filter(
        (o) =>
          o.label.toLowerCase().includes(q) ||
          (o.sublabel ?? "").toLowerCase().includes(q),
      )
    : props.options;
  return list.slice(0, 50);
});

watch(filtered, () => {
  activeIndex.value = filtered.value.length > 0 ? 0 : -1;
});

function open() {
  if (props.disabled) return;
  isOpen.value = true;
}

function close() {
  isOpen.value = false;
  activeIndex.value = -1;
}

function pick(option: ComplianceComboboxOption) {
  emit("update:modelValue", option.value);
  query.value = "";
  close();
}

function clearSelection() {
  emit("update:modelValue", "");
  query.value = "";
  void nextTick(() => inputEl.value?.focus());
}

function onInput(event: Event) {
  query.value = (event.target as HTMLInputElement).value;
  open();
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === "ArrowDown") {
    event.preventDefault();
    open();
    activeIndex.value = Math.min(activeIndex.value + 1, filtered.value.length - 1);
  } else if (event.key === "ArrowUp") {
    event.preventDefault();
    open();
    activeIndex.value = Math.max(activeIndex.value - 1, 0);
  } else if (event.key === "Enter") {
    const active = filtered.value[activeIndex.value];
    if (active) {
      event.preventDefault();
      pick(active);
    }
  } else if (event.key === "Escape" && isOpen.value) {
    event.preventDefault();
    close();
  }
}

const activeOptionId = computed(() =>
  activeIndex.value >= 0 ? `${props.inputId}-option-${activeIndex.value}` : undefined,
);
</script>

<template>
  <div class="filter-combobox">
    <div v-if="selected" class="filter-combobox__picked">
      <span class="filter-combobox__picked-label">
        {{ selected.label }}
        <span v-if="selected.sublabel" class="filter-combobox__picked-sublabel">
          — {{ selected.sublabel }}
        </span>
      </span>
      <button
        type="button"
        class="filter-combobox__change"
        :disabled="disabled"
        @click="clearSelection"
      >
        {{ changeLabel }}
      </button>
    </div>

    <div v-else class="filter-combobox__field">
      <input
        :id="inputId"
        ref="inputEl"
        type="text"
        role="combobox"
        autocomplete="off"
        aria-autocomplete="list"
        :aria-expanded="isOpen"
        :aria-controls="`${inputId}-listbox`"
        :aria-activedescendant="activeOptionId"
        :aria-label="fieldLabel"
        :disabled="disabled"
        class="filter-combobox__input"
        :placeholder="loading ? loadingText : placeholder"
        :value="query"
        @input="onInput"
        @focus="open"
        @blur="close"
        @keydown="onKeydown"
      />

      <ul
        v-if="isOpen"
        :id="`${inputId}-listbox`"
        role="listbox"
        class="filter-combobox__listbox"
      >
        <li v-if="loading" class="filter-combobox__hint" role="presentation">
          {{ loadingText }}
        </li>
        <li
          v-else-if="filtered.length === 0"
          class="filter-combobox__hint"
          role="presentation"
        >
          {{ emptyText }}
        </li>
        <li
          v-for="(option, index) in filtered"
          :id="`${inputId}-option-${index}`"
          :key="option.value"
          role="option"
          :aria-selected="index === activeIndex"
          class="filter-combobox__option"
          :class="{ 'is-active': index === activeIndex }"
          @mousedown.prevent="pick(option)"
          @mouseenter="activeIndex = index"
        >
          <span class="filter-combobox__option-label">{{ option.label }}</span>
          <span v-if="option.sublabel" class="filter-combobox__option-sublabel">
            {{ option.sublabel }}
          </span>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
.filter-combobox {
  position: relative;
  min-width: 0;
}

.filter-combobox__picked {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  height: var(--size-control-md, 36px);
  padding: 0 var(--space-3);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
  font-size: var(--font-size-sm);
}

.filter-combobox__picked-label {
  overflow: hidden;
  min-width: 0;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-weight: var(--font-weight-medium);
  color: var(--text-primary);
}

.filter-combobox__picked-sublabel {
  font-weight: var(--font-weight-normal);
  color: var(--text-secondary);
}

.filter-combobox__change {
  flex-shrink: 0;
  border: none;
  background: transparent;
  color: var(--action-primary, #0969da);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  cursor: pointer;
  padding: var(--space-1) 0;
}

.filter-combobox__change:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.filter-combobox__field {
  position: relative;
}

.filter-combobox__input {
  width: 100%;
  height: var(--size-control-md, 36px);
  padding: 0 var(--space-3);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--bg-input);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}

.filter-combobox__input:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: var(--shadow-focus);
}

.filter-combobox__listbox {
  position: absolute;
  z-index: 30;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  max-height: 15rem;
  overflow-y: auto;
  margin: 0;
  padding: var(--space-1);
  list-style: none;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  box-shadow: var(--shadow-md, 0 4px 12px rgba(0, 0, 0, 0.12));
}

.filter-combobox__hint {
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.filter-combobox__option {
  display: grid;
  gap: 1px;
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  cursor: pointer;
}

.filter-combobox__option.is-active {
  background: var(--bg-card-hover, #f0f2f4);
}

.filter-combobox__option-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--text-primary);
}

.filter-combobox__option-sublabel {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}
</style>
