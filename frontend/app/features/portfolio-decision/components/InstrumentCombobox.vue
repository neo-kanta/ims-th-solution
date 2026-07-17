<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";

import type { ApiInstrument } from "~/features/investment-ledger/services/investmentLedgerApi";
import { useI18n } from "~/composables/useI18n";
import { useInstrumentSearch } from "../composables/useInstrumentSearch";
import type { EnrichedHolding } from "../composables/usePortfolioHoldingsDirectory";

const props = withDefaults(
  defineProps<{
    selected: ApiInstrument | null;
    ownedHoldings?: EnrichedHolding[];
    disabled?: boolean;
    errorMessage?: string | null;
    inputId?: string;
    /**
     * SELL-only hardening: when true, only owned holdings can ever be
     * selected — the free-text/backend search box is not rendered at all,
     * so a non-owned instrument can never appear as a selectable SELL
     * option (see decisionValidation's isOwnedInstrument gate for the
     * matching data-level guard).
     */
    restrictToOwned?: boolean;
  }>(),
  {
    ownedHoldings: () => [],
    disabled: false,
    errorMessage: null,
    inputId: "decision-instrument",
    restrictToOwned: false,
  },
);

const { t } = useI18n();

const emit = defineEmits<{
  select: [instrument: ApiInstrument];
  clear: [];
}>();

const search = useInstrumentSearch();
const inputEl = ref<HTMLInputElement | null>(null);

const ownedInstrumentIds = computed(
  () => new Set(props.ownedHoldings.map((h) => h.instrumentId)),
);

function ownedQuantityLabel(id: string | undefined | null): string | null {
  if (!id) return null;
  const holding = props.ownedHoldings.find((h) => h.instrumentId === id);
  return holding ? holding.quantity : null;
}

function pick(instrument: ApiInstrument) {
  emit("select", instrument);
  search.reset();
}

function pickOwned(holding: EnrichedHolding) {
  if (holding.instrument) pick(holding.instrument);
}

function onInput(event: Event) {
  if (event.target instanceof HTMLInputElement) {
    search.setQuery(event.target.value);
  }
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === "ArrowDown") {
    event.preventDefault();
    search.open();
    search.moveNext();
  } else if (event.key === "ArrowUp") {
    event.preventDefault();
    search.open();
    search.movePrev();
  } else if (event.key === "Enter") {
    if (search.activeItem.value) {
      event.preventDefault();
      pick(search.activeItem.value);
    }
  } else if (event.key === "Escape") {
    if (search.isOpen.value) {
      event.preventDefault();
      search.close();
    }
  }
}

function clearSelection() {
  emit("clear");
  search.reset();
  void nextTick(() => inputEl.value?.focus());
}

watch(
  () => props.disabled,
  (disabled) => {
    if (disabled) search.close();
  },
);

const activeOptionId = computed(() =>
  search.activeIndex.value >= 0 ? `${props.inputId}-option-${search.activeIndex.value}` : undefined,
);
</script>

<template>
  <div class="instrument-combobox">
    <div v-if="selected" class="instrument-combobox__picked">
      <div>
        <div class="instrument-combobox__picked-primary">
          {{ selected.primary_ticker ?? selected.name ?? "—" }}
        </div>
        <div class="instrument-combobox__picked-meta">
          <span v-if="selected.name">{{ selected.name }}</span>
          <span v-if="selected.primary_exchange"> · {{ selected.primary_exchange }}</span>
          <span v-if="selected.currency"> · {{ selected.currency }}</span>
        </div>
      </div>
      <button
        type="button"
        class="instrument-combobox__change-btn"
        :disabled="disabled"
        @click="clearSelection"
      >
        {{ t("portfolio.decisionNew.instrumentChange") }}
      </button>
    </div>

    <template v-else>
      <div
        v-if="ownedHoldings.length > 0"
        class="instrument-combobox__owned"
        :aria-label="t('portfolio.decisionNew.ownedPositions')"
      >
        <span class="instrument-combobox__owned-label">{{ t("portfolio.decisionNew.ownedPositions") }}</span>
        <div class="instrument-combobox__owned-list">
          <button
            v-for="holding in ownedHoldings"
            :key="holding.instrumentId"
            type="button"
            class="instrument-combobox__owned-chip"
            :disabled="disabled || !holding.instrument"
            @click="pickOwned(holding)"
          >
            <span class="instrument-combobox__owned-ticker">
              {{ holding.instrument?.primary_ticker ?? "—" }}
            </span>
            <span class="instrument-combobox__owned-qty">{{ t("portfolio.decisionNew.quantityAbbrev", { quantity: holding.quantity }) }}</span>
          </button>
        </div>
      </div>

      <p v-if="restrictToOwned && ownedHoldings.length === 0" class="instrument-combobox__hint">
        {{ t("portfolio.decisionNew.noOwnedHoldings") }}
      </p>

      <div v-if="!restrictToOwned" class="instrument-combobox__field">
        <input
          :id="inputId"
          ref="inputEl"
          type="text"
          role="combobox"
          autocomplete="off"
          aria-autocomplete="list"
          :aria-expanded="search.isOpen.value"
          :aria-controls="`${inputId}-listbox`"
          :aria-activedescendant="activeOptionId"
          :disabled="disabled"
          class="instrument-combobox__input"
          :placeholder="t('portfolio.decisionNew.instrumentSearchPlaceholder')"
          :value="search.query.value"
          @input="onInput"
          @focus="search.open()"
          @keydown="onKeydown"
        />

        <ul
          v-if="search.isOpen.value"
          :id="`${inputId}-listbox`"
          role="listbox"
          class="instrument-combobox__listbox"
        >
          <li v-if="search.loading.value" class="instrument-combobox__hint" role="presentation">
            {{ t("portfolio.decisionNew.instrumentSearching") }}
          </li>
          <li
            v-else-if="search.error.value"
            class="instrument-combobox__hint instrument-combobox__hint--error"
            role="presentation"
          >
            {{ search.error.value }}
          </li>
          <li
            v-else-if="search.items.value.length === 0"
            class="instrument-combobox__hint"
            role="presentation"
          >
            {{ t("portfolio.decisionNew.instrumentNoMatches") }}
          </li>
          <li
            v-for="(inst, index) in search.items.value"
            :id="`${inputId}-option-${index}`"
            :key="inst.id"
            role="option"
            :aria-selected="index === search.activeIndex.value"
            class="instrument-combobox__option"
            :class="{ 'is-active': index === search.activeIndex.value }"
            @mousedown.prevent="pick(inst)"
            @mouseenter="search.activeIndex.value = index"
          >
            <span class="instrument-combobox__option-primary">
              {{ inst.primary_ticker ?? inst.name ?? "—" }}
              <span v-if="ownedInstrumentIds.has(inst.id ?? '')" class="instrument-combobox__owned-badge">
                {{ t("portfolio.decisionNew.ownedBadge") }} · {{ ownedQuantityLabel(inst.id) }}
              </span>
            </span>
            <span class="instrument-combobox__option-secondary">
              {{ inst.name ?? "" }}
              <span v-if="inst.primary_exchange"> · {{ inst.primary_exchange }}</span>
              <span v-if="inst.currency"> · {{ inst.currency }}</span>
            </span>
          </li>
        </ul>
      </div>
      <span v-if="errorMessage" class="instrument-combobox__error">{{ errorMessage }}</span>
    </template>
  </div>
</template>

<style scoped>
.instrument-combobox {
  display: grid;
  gap: 6px;
}

.instrument-combobox__picked {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border: 1px solid var(--border-default, #d0d7de);
  border-radius: var(--radius-md, 6px);
  background: var(--bg-card-muted, #f6f8fa);
}

.instrument-combobox__picked-primary {
  font-weight: 700;
  color: var(--text-primary);
}

.instrument-combobox__picked-meta {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 2px;
}

.instrument-combobox__change-btn {
  border: 1px solid var(--border-default, #d0d7de);
  background: var(--bg-card, #fff);
  border-radius: var(--radius-sm, 4px);
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

.instrument-combobox__owned {
  display: grid;
  gap: 4px;
}

.instrument-combobox__owned-label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary);
}

.instrument-combobox__owned-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.instrument-combobox__owned-chip {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 1px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-card, #fff);
  padding: 5px 10px;
  cursor: pointer;
  font-size: 12px;
}

.instrument-combobox__owned-chip:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.instrument-combobox__owned-ticker {
  font-weight: 700;
  color: var(--text-primary);
}

.instrument-combobox__owned-qty {
  color: var(--text-tertiary);
  font-size: 11px;
}

.instrument-combobox__field {
  position: relative;
}

.instrument-combobox__input {
  width: 100%;
  height: 36px;
  padding: 0 12px;
  border: 1px solid var(--border-default, #d0d7de);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-input, #fff);
  font-size: 13px;
  color: var(--text-primary);
}

.instrument-combobox__input:focus {
  outline: none;
  border-color: var(--border-focus, #0969da);
  box-shadow: var(--shadow-focus, 0 0 0 3px rgba(9, 105, 218, 0.15));
}

.instrument-combobox__listbox {
  position: absolute;
  z-index: 20;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  max-height: 240px;
  overflow-y: auto;
  margin: 0;
  padding: 4px;
  list-style: none;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  background: var(--bg-card, #fff);
  box-shadow: var(--shadow-md, 0 4px 12px rgba(0, 0, 0, 0.1));
}

.instrument-combobox__hint {
  padding: 8px 10px;
  font-size: 12px;
  color: var(--text-tertiary);
}

.instrument-combobox__hint--error {
  color: var(--state-danger, #cf222e);
}

.instrument-combobox__option {
  display: grid;
  gap: 2px;
  padding: 7px 10px;
  border-radius: var(--radius-sm, 4px);
  cursor: pointer;
}

.instrument-combobox__option.is-active {
  background: var(--bg-card-hover, #f0f2f4);
}

.instrument-combobox__option-primary {
  font-weight: 600;
  font-size: 13px;
  color: var(--text-primary);
}

.instrument-combobox__owned-badge {
  margin-left: 6px;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--state-success, #1a7f37);
}

.instrument-combobox__option-secondary {
  font-size: 11px;
  color: var(--text-tertiary);
}

.instrument-combobox__error {
  font-size: 12px;
  color: var(--state-danger, #cf222e);
}
</style>
