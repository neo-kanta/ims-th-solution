<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import { useRouter } from "#imports";

interface Props {
  modelValue: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  "update:modelValue": [value: boolean];
}>();

const router = useRouter();
const query = ref("");
const selectedIndex = ref(0);
const searchInput = ref<HTMLInputElement | null>(null);

// Close modal handler
function handleClose() {
  emit("update:modelValue", false);
  query.value = "";
}

// Extensible Item interface
interface SuggestionItem {
  id: string;
  label: string;
  value: string;
  type: "feature" | "function";
  to?: string;
  actionText: string;
  iconName: "repo" | "workflow" | "compliance" | "settings" | "organization";
}

// Extensible Search Engine Architecture
interface SearchQueryContext {
  query: string;
  limit?: number;
}

interface SearchProvider {
  name: string;
  search: (ctx: SearchQueryContext) => Promise<SuggestionItem[]>;
}

// --- PROVIDER: Features (App Navigation pages) ---
const featureProvider: SearchProvider = {
  name: "Features",
  search: async (ctx) => {
    const q = ctx.query.toLowerCase().trim();
    const features: SuggestionItem[] = [
      { id: "nav-dashboard", label: "Dashboard Overview", value: "Dashboard", type: "feature", to: "/", actionText: "Jump to", iconName: "workflow" },
      { id: "nav-funds", label: "My investment funds", value: "My funds investment portfolio", type: "feature", to: "/investment/funds", actionText: "Jump to", iconName: "repo" },
      { id: "nav-research", label: "Investment Research console", value: "Investment Research analysis", type: "feature", to: "/investment/analysis", actionText: "Jump to", iconName: "repo" },
      { id: "nav-market-data", label: "Market Data feeds", value: "Market data stats global", type: "feature", to: "/market-data", actionText: "Jump to", iconName: "organization" },
      { id: "nav-compliance-dashboard", label: "Compliance dashboard", value: "Compliance overview dashboard", type: "feature", to: "/compliance", actionText: "Jump to", iconName: "compliance" },
      { id: "nav-rules", label: "Compliance rule library", value: "Rule library settings compliance", type: "feature", to: "/compliance/rules", actionText: "Jump to", iconName: "compliance" },
      { id: "nav-pretrade", label: "Pre-trade simulator", value: "Pre-trade simulator test compliance", type: "feature", to: "/compliance/pre-trade", actionText: "Jump to", iconName: "compliance" },
      { id: "nav-posttrade", label: "Post-trade breaches tracker", value: "Post-trade breaches warnings logs", type: "feature", to: "/compliance/post-trade", actionText: "Jump to", iconName: "compliance" },
      { id: "nav-personal-settings", label: "Personal profile & security settings", value: "Personal Settings profile security username", type: "feature", to: "/settings", actionText: "Jump to", iconName: "settings" },
    ];

    if (!q) {
      return features; // Return all features by default when query is empty
    }

    return features.filter(
      (item) =>
        item.label.toLowerCase().includes(q) ||
        item.value.toLowerCase().includes(q)
    );
  },
};

// --- REGISTER EXTENSIBLE PROVIDERS HERE ---
const activeProviders = [featureProvider];

const filteredSuggestions = ref<SuggestionItem[]>([]);
const isSearching = ref(false);

// Run all search providers asynchronously
async function executeSearch() {
  const currentQuery = query.value;
  isSearching.value = true;
  try {
    const providerResults = await Promise.all(
      activeProviders.map((provider) => provider.search({ query: currentQuery }))
    );
    filteredSuggestions.value = providerResults.flat();
  } catch (error) {
    console.error("Search providers error", error);
  } finally {
    isSearching.value = false;
  }
}

// Watch query to trigger search
watch(query, () => {
  executeSearch();
}, { immediate: true });

// Group suggestions for categorized rendering
interface SuggestionGroup {
  title: string;
  items: SuggestionItem[];
}

const groupedSuggestions = computed<SuggestionGroup[]>(() => {
  const list = filteredSuggestions.value;
  const groups: Record<string, SuggestionItem[]> = {
    Features: [],
  };

  list.forEach((item) => {
    if (item.type === "feature") groups.Features.push(item);
  });

  return [
    { title: "Suggested Features", items: groups.Features },
  ].filter((g) => g.items.length > 0);
});

// Flat list of visible suggestions for arrow key indexing
const flatVisibleSuggestions = computed<SuggestionItem[]>(() => {
  return groupedSuggestions.value.reduce<SuggestionItem[]>((acc, group) => {
    return [...acc, ...group.items];
  }, []);
});

// Watch visibility to focus input
watch(
  () => props.modelValue,
  async (newVal) => {
    if (newVal) {
      selectedIndex.value = 0;
      await nextTick();
      if (searchInput.value) {
        searchInput.value.focus();
      }
    }
  }
);

// Reset index when search query changes
watch(query, () => {
  selectedIndex.value = 0;
});

// Emit feedback toast or notification
const showFeedbackNotice = ref(false);

function selectItem(item: SuggestionItem) {
  if (item.to) {
    router.push(item.to);
    handleClose();
  }
}

// Keyboard navigation handlers
function handleKeyDown(event: KeyboardEvent) {
  const len = flatVisibleSuggestions.value.length;
  if (len === 0) return;

  if (event.key === "ArrowDown") {
    event.preventDefault();
    selectedIndex.value = (selectedIndex.value + 1) % len;
  } else if (event.key === "ArrowUp") {
    event.preventDefault();
    selectedIndex.value = (selectedIndex.value - 1 + len) % len;
  } else if (event.key === "Enter") {
    event.preventDefault();
    const item = flatVisibleSuggestions.value[selectedIndex.value];
    if (item) {
      selectItem(item);
    }
  } else if (event.key === "Escape") {
    handleClose();
  }
}

// Global hotkey: "/" to search, "Esc" to close
function handleGlobalKeyDown(event: KeyboardEvent) {
  const activeEl = document.activeElement;
  const isInputField =
    activeEl &&
    (activeEl.tagName === "INPUT" ||
      activeEl.tagName === "TEXTAREA" ||
      activeEl.tagName === "SELECT" ||
      activeEl.hasAttribute("contenteditable"));

  if (event.key === "/" && !isInputField && !props.modelValue) {
    event.preventDefault();
    emit("update:modelValue", true);
  }
}

onMounted(() => {
  window.addEventListener("keydown", handleGlobalKeyDown);
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleGlobalKeyDown);
});

function handleFeedbackClick() {
  showFeedbackNotice.value = true;
  setTimeout(() => {
    showFeedbackNotice.value = false;
  }, 3000);
}
</script>

<template>
  <Teleport to="body">
    <Transition name="search-fade">
      <div v-if="modelValue" class="search-backdrop" @click.self="handleClose">
        <div class="search-dialog" role="dialog" aria-modal="true" @keydown="handleKeyDown">
          <!-- Input Header -->
          <div class="search-input-wrapper">
            <svg aria-hidden="true" class="search-input-leading-icon" viewBox="0 0 16 16" width="16" height="16" fill="currentColor">
              <path d="M10.68 11.74a6 6 0 0 1-7.922-8.982 6 6 0 0 1 8.982 7.922l3.04 3.04a.749.749 0 0 1-.326 1.275.749.749 0 0 1-.734-.215ZM11.5 7a4.499 4.499 0 1 0-8.997 0A4.499 4.499 0 0 0 11.5 7Z"></path>
            </svg>
            <input
              ref="searchInput"
              v-model="query"
              class="search-text-input"
              type="text"
              placeholder="Search or jump to..."
              autocomplete="off"
              spellcheck="false"
            />
            <button
              v-if="query"
              class="search-clear-btn"
              type="button"
              aria-label="Clear search text"
              @click="query = ''"
            >
              <svg aria-hidden="true" viewBox="0 0 16 16" width="16" height="16" fill="currentColor">
                <path d="M2.343 13.657A8 8 0 1 1 13.658 2.343 8 8 0 0 1 2.343 13.657ZM6.03 4.97a.751.751 0 0 0-1.042.018.751.751 0 0 0-.018 1.042L6.94 8 4.97 9.97a.749.749 0 0 0 .326 1.275.749.749 0 0 0 .734-.215L8 9.06l1.97 1.97a.749.749 0 0 0 1.275-.326.749.749 0 0 0-.215-.734L9.06 8l1.97-1.97a.749.749 0 0 0-.326-1.275.749.749 0 0 0-.734.215L8 6.94Z"></path>
              </svg>
            </button>
          </div>

          <!-- Body Suggestions List -->
          <div class="search-body">
            <div v-if="flatVisibleSuggestions.length === 0" class="search-no-results">
              No results found for "{{ query }}"
            </div>

            <div v-else class="search-groups">
              <div v-for="group in groupedSuggestions" :key="group.title" class="search-group">
                <h3 class="search-group-title">{{ group.title }}</h3>
                <ul class="search-group-list">
                  <li
                    v-for="item in group.items"
                    :key="item.id"
                    class="search-item"
                    :class="{ 'is-active': flatVisibleSuggestions[selectedIndex]?.id === item.id }"
                    @click="selectItem(item)"
                    @mouseenter="selectedIndex = flatVisibleSuggestions.findIndex((x) => x.id === item.id)"
                  >
                    <span class="search-item-left">
                      <!-- Custom inline SVG icons based on iconName -->
                      <span class="search-item-icon">
                        <svg v-if="item.iconName === 'organization'" viewBox="0 0 16 16" width="16" height="16" fill="currentColor">
                          <path d="M1.75 16A1.75 1.75 0 0 1 0 14.25V1.75C0 .784.784 0 1.75 0h8.5C11.216 0 12 .784 12 1.75v12.5c0 .085-.006.168-.018.25h2.268a.25.25 0 0 0 .25-.25V8.285a.25.25 0 0 0-.111-.208l-1.055-.703a.749.749 0 1 1 .832-1.248l1.055.703c.487.325.779.871.779 1.456v5.965A1.75 1.75 0 0 1 14.25 16h-3.5a.766.766 0 0 1-.197-.026c-.099.017-.2.026-.303.026h-3a.75.75 0 0 1-.75-.75V14h-1v1.25a.75.75 0 0 1-.75.75Zm-.25-1.75c0 .138.112.25.25.25H4v-1.25a.75.75 0 0 1 .75-.75h2.5a.75.75 0 0 1 .75.75v1.25h2.25a.25.25 0 0 0 .25-.25V1.75a.25.25 0 0 0-.25-.25h-8.5a.25.25 0 0 0-.25.25ZM3.75 6h.5a.75.75 0 0 1 0 1.5h-.5a.75.75 0 0 1 0-1.5ZM3 3.75A.75.75 0 0 1 3.75 3h.5a.75.75 0 0 1 0 1.5h-.5A.75.75 0 0 1 3 3.75Zm4 3A.75.75 0 0 1 7.75 6h.5a.75.75 0 0 1 0 1.5h-.5A.75.75 0 0 1 7 6.75ZM7.75 3h.5a.75.75 0 0 1 0 1.5h-.5a.75.75 0 0 1 0-1.5ZM3 9.75A.75.75 0 0 1 3.75 9h.5a.75.75 0 0 1 0 1.5h-.5A.75.75 0 0 1 3 9.75ZM7.75 9h.5a.75.75 0 0 1 0 1.5h-.5a.75.75 0 0 1 0-1.5Z"></path>
                        </svg>
                        <svg v-else-if="item.iconName === 'repo'" viewBox="0 0 16 16" width="16" height="16" fill="currentColor">
                          <path d="M2 2.5A2.5 2.5 0 0 1 4.5 0h8.75a.75.75 0 0 1 .75.75v12.5a.75.75 0 0 1-.75.75h-2.5a.75.75 0 0 1 0-1.5h1.75v-2h-8a1 1 0 0 0-.714 1.7.75.75 0 1 1-1.072 1.05A2.495 2.495 0 0 1 2 11.5Zm10.5-1h-8a1 1 0 0 0-1 1v6.708A2.486 2.486 0 0 1 4.5 9h8ZM5 12.25a.25.25 0 0 1 .25-.25h3.5a.25.25 0 0 1 .25.25v3.25a.25.25 0 0 1-.4.2l-1.45-1.087a.249.249 0 0 0-.3 0L5.4 15.7a.25.25 0 0 1-.4-.2Z"></path>
                        </svg>
                        <svg v-else-if="item.iconName === 'workflow'" viewBox="0 0 16 16" width="16" height="16" fill="currentColor">
                          <path d="M0 1.75C0 .784.784 0 1.75 0h3.5C6.216 0 7 .784 7 1.75v3.5A1.75 1.75 0 0 1 5.25 7H4v4a1 1 0 0 0 1 1h4v-1.25C9 9.784 9.784 9 10.75 9h3.5c.966 0 1.75.784 1.75 1.75v3.5A1.75 1.75 0 0 1 14.25 16h-3.5A1.75 1.75 0 0 1 9 14.25v-.75H5A2.5 2.5 0 0 1 2.5 11V7h-.75A1.75 1.75 0 0 1 0 5.25Zm1.75-.25a.25.25 0 0 0-.25.25v3.5c0 .138.112.25.25.25h3.5a.25.25 0 0 0 .25-.25v-3.5a.25.25 0 0 0-.25-.25Zm9 9a.25.25 0 0 0-.25.25v3.5c0 .138.112.25.25.25h3.5a.25.25 0 0 0 .25-.25v-3.5a.25.25 0 0 0-.25-.25Z"></path>
                        </svg>
                        <svg v-else-if="item.iconName === 'compliance'" viewBox="0 0 16 16" width="16" height="16" fill="currentColor">
                          <path d="m8.533.133 5.25 1.68A1.75 1.75 0 0 1 15 3.48V7c0 1.566-.32 3.182-1.303 4.682-.983 1.498-2.585 2.813-5.032 3.855a1.697 1.697 0 0 1-1.33 0c-2.447-1.042-4.049-2.357-5.032-3.855C1.32 10.182 1 8.566 1 7V3.48a1.75 1.75 0 0 1 1.217-1.667l5.25-1.68a1.748 1.748 0 0 1 1.066 0Zm-.61 1.429.001.001-5.25 1.68a.251.251 0 0 0-.174.237V7c0 1.36.275 2.666 1.057 3.859.784 1.194 2.121 2.342 4.366 3.298a.196.196 0 0 0 .154 0c2.245-.957 3.582-2.103 4.366-3.297C13.225 9.666 13.5 8.358 13.5 7V3.48a.25.25 0 0 0-.174-.238l-5.25-1.68a.25.25 0 0 0-.153 0Zm11.28 6.28l-3.5 3.5a.75.75 0 0 1-1.06 0l-1.5-1.5a.749.749 0 0 1 .326-1.275.749.749 0 0 1 .734.215l.97.97 2.97-2.97a.751.751 0 0 1 1.042.018.751.751 0 0 1 .018 1.042Z"></path>
                        </svg>
                        <svg v-else viewBox="0 0 16 16" width="16" height="16" fill="currentColor">
                          <path d="M8 4a4 4 0 1 1 0 8 4 4 0 0 1 0-8Z"></path>
                        </svg>
                      </span>
                      <span class="search-item-label">{{ item.label }}</span>
                    </span>
                    <span class="search-item-action-text">{{ item.actionText }}</span>
                  </li>
                </ul>
              </div>
            </div>
          </div>

          <!-- Footer Information -->
          <div class="search-footer">
            <a
              href="https://docs.github.com/search-github/github-code-search/understanding-github-code-search-syntax"
              target="_blank"
              class="search-footer-link"
            >
              Search syntax tips
            </a>
            <button class="search-footer-btn" type="button" @click="handleFeedbackClick">
              Give feedback
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>

  <!-- Notification Toast feedback -->
  <Teleport to="body">
    <Transition name="toast-slide">
      <div v-if="showFeedbackNotice" class="search-toast" role="status">
        <span class="search-toast-message">Thank you for your feedback!</span>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
/* Fade transition for backdrop */
.search-fade-enter-active,
.search-fade-leave-active {
  transition: opacity 0.15s ease;
}
.search-fade-enter-from,
.search-fade-leave-to {
  opacity: 0;
}

/* Slide dialog transition */
.search-fade-enter-active .search-dialog {
  animation: slideIn 0.2s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
.search-fade-leave-active .search-dialog {
  animation: slideOut 0.15s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

@keyframes slideIn {
  from {
    transform: translate(-50%, -40px) scale(0.96);
  }
  to {
    transform: translate(-50%, 0) scale(1);
  }
}
@keyframes slideOut {
  from {
    transform: translate(-50%, 0) scale(1);
  }
  to {
    transform: translate(-50%, -15px) scale(0.98);
  }
}

/* Backdrop */
.search-backdrop {
  position: fixed;
  inset: 0;
  background-color: var(--bg-overlay);
  backdrop-filter: blur(4px);
  z-index: 1050; /* Above regular content & headers */
}

/* Centered Dialog */
.search-dialog {
  position: fixed;
  top: 16px; /* Spacing from the top of the viewport */
  left: 50%;
  transform: translateX(-50%);
  width: 75vw; /* 3/4 of the screen on X axis */
  max-width: 1200px;
  background-color: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: 12px;
  box-shadow: var(--shadow-lg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  z-index: 1051;
}

@media (max-width: 768px) {
  .search-dialog {
    width: calc(100% - 32px);
  }
}

/* Input container style */
.search-input-wrapper {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-subtle);
  background-color: var(--bg-card);
  gap: 12px;
}

.search-input-leading-icon {
  color: var(--text-secondary);
  flex-shrink: 0;
}

.search-text-input {
  flex: 1;
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-size: 15px;
  padding-left: var(--space-3);
  outline: none;
  width: 100%;
}

.search-text-input::placeholder {
  color: var(--text-placeholder);
}

.search-clear-btn {
  color: var(--text-tertiary);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
  border-radius: 50%;
  transition: background 0.15s ease, color 0.15s ease;
}

.search-clear-btn:hover {
  background-color: var(--bg-card-hover);
  color: var(--text-primary);
}

/* Body / Suggestion style */
.search-body {
  max-height: 380px;
  overflow-y: auto;
  padding: 8px 0;
  background-color: var(--bg-card);
}

.search-no-results {
  padding: 32px 16px;
  text-align: center;
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.search-group {
  margin-bottom: 8px;
}

.search-group:last-child {
  margin-bottom: 0;
}

.search-group-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  padding: 8px 16px 4px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0;
}

.search-group-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

/* Individual list item style */
.search-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease;
  font-size: 13.5px;
  color: var(--text-primary);
}

.search-item-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.search-item-icon {
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.search-item-label {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.search-item-action-text {
  font-size: 11.5px;
  color: var(--text-tertiary);
  white-space: nowrap;
  opacity: 0.85;
}

/* Hover/active item states in both light & dark themes */
.search-item:hover,
.search-item.is-active {
  background-color: var(--action-primary);
  color: var(--text-inverse);
}

.search-item:hover .search-item-icon,
.search-item.is-active .search-item-icon,
.search-item:hover .search-item-action-text,
.search-item.is-active .search-item-action-text {
  color: var(--text-inverse);
  opacity: 1;
}

/* Footer layout */
.search-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  border-top: 1px solid var(--border-subtle);
  background-color: var(--bg-card);
  font-size: 12px;
}

.search-footer-link {
  color: var(--text-link);
  transition: color 0.15s ease;
}

.search-footer-link:hover {
  color: var(--text-link-hover);
  text-decoration: underline;
}

.search-footer-btn {
  color: var(--text-secondary);
  background: none;
  border: none;
  font-size: 12px;
  cursor: pointer;
  transition: color 0.15s ease;
}

.search-footer-btn:hover {
  color: var(--text-primary);
}

/* Simple toast notifications */
.search-toast {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  background-color: #1f2328;
  color: #ffffff;
  padding: 10px 20px;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  font-size: 13px;
  z-index: 2000;
  pointer-events: none;
}

/* Toast animations */
.toast-slide-enter-active,
.toast-slide-leave-active {
  transition: transform 0.25s ease, opacity 0.2s ease;
}

.toast-slide-enter-from {
  transform: translate(-50%, 15px);
  opacity: 0;
}

.toast-slide-leave-to {
  transform: translate(-50%, 10px);
  opacity: 0;
}
</style>
