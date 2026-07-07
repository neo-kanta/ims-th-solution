<script setup lang="ts">
import { ref, computed, onBeforeUnmount } from "vue";

import { useAuthStore } from "~/stores/useAuthStore";
import { useI18n } from "~/composables/useI18n";
import type { MyPortfolioCard } from "../types";

interface Props {
  card: MyPortfolioCard;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "open", code: string): void;
  (e: "viewBreaches", fundId: string): void;
  (e: "writeResearch", fundId: string): void;
}>();

const auth = useAuthStore();
const { t } = useI18n();

const canCreateResearch = computed(() => auth.hasPermission("INVESTMENT_RESEARCH_CREATE"));
const canViewBreaches = computed(() => auth.hasPermission("IRG_VIEW_BREACHES"));

const showResearch = computed(() => canCreateResearch.value && props.card.role !== "MEMBER" && props.card.fund_id);
const showBreaches = computed(() => canViewBreaches.value && props.card.compliance.open_count > 0 && props.card.fund_id);

const showDropdown = ref(false);

function toggleDropdown(e: Event) {
  showDropdown.value = !showDropdown.value;
  if (showDropdown.value) {
    document.addEventListener("click", closeDropdown);
  } else {
    document.removeEventListener("click", closeDropdown);
  }
}

function closeDropdown() {
  showDropdown.value = false;
  document.removeEventListener("click", closeDropdown);
}

onBeforeUnmount(() => {
  document.removeEventListener("click", closeDropdown);
});

function handleOpen() {
  emit("open", props.card.code);
}

function handleResearch() {
  if (props.card.fund_id) {
    emit("writeResearch", props.card.fund_id);
  }
}

function handleBreaches() {
  if (props.card.fund_id) {
    emit("viewBreaches", props.card.fund_id);
  }
}
</script>

<template>
  <div class="role-actions">
    <div class="role-actions__split-btn" @click.stop>
      <button
        type="button"
        class="role-actions__btn-main"
        @click="handleOpen"
      >
        {{ t("portfolio.actions.open", "Open") }}
      </button>
      <button
        type="button"
        class="role-actions__btn-chevron"
        aria-label="More actions"
        @click="toggleDropdown"
      >
        <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </button>

      <div v-if="showDropdown" class="role-actions__dropdown">
        <button type="button" class="role-actions__dropdown-item" @click="handleOpen(); closeDropdown()">
          {{ t("portfolio.actions.openDetails", "Open details") }}
        </button>
        <button
          v-if="showResearch"
          type="button"
          class="role-actions__dropdown-item"
          @click="handleResearch(); closeDropdown()"
        >
          {{ t("portfolio.actions.writeResearch", "New research") }}
        </button>
        <button
          v-if="showBreaches"
          type="button"
          class="role-actions__dropdown-item"
          @click="handleBreaches(); closeDropdown()"
        >
          {{ t("portfolio.actions.viewBreaches", "View breaches") }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.role-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.role-actions__split-btn {
  position: relative;
  display: inline-flex;
  align-items: center;
  background: var(--state-info, #1f6feb);
  border-radius: 6px;
  height: 28px;
  box-shadow: 0 1px 2px rgba(31, 111, 235, 0.15);
  transition: background-color 0.12s ease;
}

.role-actions__split-btn:hover {
  background: #185ec2;
}

.role-actions__btn-main {
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  color: #ffffff;
  background: transparent;
  border: none;
  padding: 0 10px 0 12px;
  cursor: pointer;
  height: 100%;
  border-radius: 6px 0 0 6px;
}

.role-actions__btn-chevron {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.85);
  padding: 0 8px;
  cursor: pointer;
  height: 100%;
  border-radius: 0 6px 6px 0;
  border-left: 1px solid rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
}

.role-actions__btn-chevron:hover {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.1);
}

.role-actions__dropdown {
  position: absolute;
  right: 0;
  bottom: calc(100% + 6px);
  background: var(--bg-card, #ffffff);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  min-width: 140px;
  padding: 4px 0;
  z-index: 100;
  display: flex;
  flex-direction: column;
}

.role-actions__dropdown-item {
  font-family: inherit;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-primary, #1f2328);
  background: transparent;
  border: none;
  padding: 8px 12px;
  text-align: left;
  cursor: pointer;
  width: 100%;
  transition: all 0.1s ease;
}

.role-actions__dropdown-item:hover {
  background: var(--bg-card-muted, #f6f8fa);
  color: var(--state-info, #1f6feb);
}
</style>
