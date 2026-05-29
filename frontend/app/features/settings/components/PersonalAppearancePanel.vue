<script setup lang="ts">
import { ref, watch, computed } from "vue";
import type { PersonalAppearancePreferences } from "../personal.types";
import AppSettingsActions from "~/shared/ui/AppSettingsActions.vue";

const props = defineProps<{
  preferences: PersonalAppearancePreferences | null;
  loading: boolean;
}>();

const emit = defineEmits<{
  update: [preferences: PersonalAppearancePreferences];
}>();

const localPrefs = ref<PersonalAppearancePreferences | null>(null);

watch(
  () => props.preferences,
  (newPrefs) => {
    if (newPrefs) {
      localPrefs.value = { ...newPrefs };
    }
  },
  { immediate: true }
);

const isDirty = computed(() => {
  if (!props.preferences || !localPrefs.value) return false;
  return (
    localPrefs.value.theme !== props.preferences.theme ||
    localPrefs.value.density !== props.preferences.density ||
    localPrefs.value.increaseContrast !== props.preferences.increaseContrast ||
    localPrefs.value.reduceMotion !== props.preferences.reduceMotion ||
    localPrefs.value.showKeyboardHints !== props.preferences.showKeyboardHints ||
    localPrefs.value.dateFormat !== props.preferences.dateFormat ||
    localPrefs.value.timezone !== props.preferences.timezone ||
    localPrefs.value.numberFormat !== props.preferences.numberFormat ||
    localPrefs.value.currencyDisplay !== props.preferences.currencyDisplay
  );
});

function updateLocalPreference(patch: Partial<PersonalAppearancePreferences>) {
  if (localPrefs.value) {
    localPrefs.value = { ...localPrefs.value, ...patch };
  }
}

function handleSave() {
  if (localPrefs.value) {
    emit("update", { ...localPrefs.value });
  }
}

function handleCancel() {
  if (props.preferences) {
    localPrefs.value = { ...props.preferences };
  }
}

const themeOptions = [
  { id: 'system', title: 'Sync with system', desc: 'Light during day, dark at night', previewClass: 'theme-preview--system' },
  { id: 'light', title: 'Light default', desc: 'Calm, primary blue, slate ink', previewClass: 'theme-preview--light' },
  { id: 'light-high-contrast', title: 'Light high-contrast', desc: 'Maximum legibility for trading floors', previewClass: 'theme-preview--light-hc' },
  { id: 'dark', title: 'Dark default', desc: 'Bloomberg-style, low-glare nights', previewClass: 'theme-preview--dark' },
  { id: 'dark-dimmed', title: 'Dark dimmed', desc: 'Softer dark — easier on long sessions', previewClass: 'theme-preview--dark-dimmed' },
  { id: 'dark-high-contrast', title: 'Dark high-contrast', desc: 'Maximum legibility, dark surfaces', previewClass: 'theme-preview--dark-hc' }
] as const;

const densityOptions = [
  { id: 'comfortable', title: 'Comfortable', desc: 'Default — generous spacing in tables and lists' },
  { id: 'compact', title: 'Compact', desc: 'Bloomberg-density — 15% more rows visible per screen' },
  { id: 'system', title: 'System default', desc: 'Match the desktop OS' }
] as const;
</script>

<template>
  <section class="personal-appearance-panel settings-panel">
    <header class="settings-panel__header">
      <h2 class="settings-panel__title">Appearance</h2>
    </header>

    <div v-if="localPrefs" class="settings-panel__body appearance-body" :aria-busy="loading" style="padding: var(--space-5);">
      
      <!-- A. Theme preference -->
      <div class="appearance-section">
        <h3 class="appearance-section-title">Theme preference</h3>
        <p class="appearance-section-subtitle">Choose how IMS looks to you. Select a single theme, or sync with your system.</p>
        <div class="theme-cards-grid" role="radiogroup" aria-label="Theme preference">
          <button
            v-for="theme in themeOptions"
            :key="theme.id"
            class="theme-card"
            :class="{ 'is-active': localPrefs.theme === theme.id }"
            role="radio"
            :aria-checked="localPrefs.theme === theme.id"
            @click="updateLocalPreference({ theme: theme.id as any })"
          >
            <div class="theme-preview" :class="theme.previewClass">
              <div class="mock-header">
                <div class="mock-logo"></div>
                <div class="mock-avatar"></div>
              </div>
              <div class="mock-body">
                <div class="mock-sidebar">
                  <div class="mock-nav-item"></div>
                  <div class="mock-nav-item"></div>
                  <div class="mock-nav-item"></div>
                </div>
                <div class="mock-content">
                  <div class="mock-card"></div>
                  <div class="mock-card"></div>
                </div>
              </div>
            </div>
            <div class="theme-card-info">
              <div class="theme-card-text">
                <div class="theme-card-title">{{ theme.title }}</div>
                <div class="theme-card-desc">{{ theme.desc }}</div>
              </div>
            </div>
          </button>
        </div>
      </div>

      <!-- B. Density -->
      <div class="appearance-section">
        <h3 class="appearance-section-title">Density</h3>
        <p class="appearance-section-subtitle">How much information should be packed into each view.</p>
        <div class="density-list" role="radiogroup" aria-label="Table density">
          <button
            v-for="density in densityOptions"
            :key="density.id"
            class="density-card"
            :class="{ 'is-active': localPrefs.density === density.id }"
            role="radio"
            :aria-checked="localPrefs.density === density.id"
            @click="updateLocalPreference({ density: density.id as any })"
          >
            <div class="radio-indicator">
              <div class="radio-dot" v-if="localPrefs.density === density.id"></div>
            </div>
            <div class="density-card-text">
              <span class="density-card-title">{{ density.title }}</span>
              <span class="density-card-desc">{{ density.desc }}</span>
            </div>
          </button>
        </div>
      </div>

      <!-- C. Accessibility -->
      <div class="appearance-section">
        <h3 class="appearance-section-title">Accessibility</h3>
        <p class="appearance-section-subtitle">Adjustments that change how content is presented.</p>
        <div class="accessibility-list">
          <label class="toggle-row">
            <div class="toggle-control">
              <input type="checkbox" :checked="localPrefs.increaseContrast" @change="updateLocalPreference({ increaseContrast: ($event.target as HTMLInputElement).checked })" />
              <div class="toggle-track"></div>
            </div>
            <div class="toggle-text">
              <div class="toggle-title">Increase contrast</div>
              <div class="toggle-desc">Boost text and border contrast above WCAG AA — required for low-vision auditors.</div>
            </div>
          </label>
          <label class="toggle-row">
            <div class="toggle-control">
              <input type="checkbox" :checked="localPrefs.reduceMotion" @change="updateLocalPreference({ reduceMotion: ($event.target as HTMLInputElement).checked })" />
              <div class="toggle-track"></div>
            </div>
            <div class="toggle-text">
              <div class="toggle-title">Reduce motion</div>
              <div class="toggle-desc">Disable transitions, stage-advance animations, and activity feed shimmer.</div>
            </div>
          </label>
          <label class="toggle-row">
            <div class="toggle-control">
              <input type="checkbox" :checked="localPrefs.showKeyboardHints" @change="updateLocalPreference({ showKeyboardHints: ($event.target as HTMLInputElement).checked })" />
              <div class="toggle-track"></div>
            </div>
            <div class="toggle-text">
              <div class="toggle-title">Show keyboard hints</div>
              <div class="toggle-desc">Show ⌘K, ↑↓, Enter overlays on tables and dialogs.</div>
            </div>
          </label>
        </div>
      </div>

      <!-- D. Numerics & formatting -->
      <div class="appearance-section">
        <h3 class="appearance-section-title">Numerics & formatting</h3>
        <p class="appearance-section-subtitle">These settings affect how IMS displays money, percentages, and dates. They do not change the underlying values stored on the server.</p>
        <div class="formatting-grid">
          <div class="form-group">
            <label>Date format</label>
            <select :value="localPrefs.dateFormat" @change="updateLocalPreference({ dateFormat: ($event.target as HTMLSelectElement).value as any })">
              <option value="YYYY-MM-DD">YYYY-MM-DD</option>
              <option value="DD/MM/YYYY">DD/MM/YYYY</option>
              <option value="MM/DD/YYYY">MM/DD/YYYY</option>
            </select>
          </div>
          <div class="form-group">
            <label>Timezone</label>
            <select :value="localPrefs.timezone" @change="updateLocalPreference({ timezone: ($event.target as HTMLSelectElement).value })">
              <option value="Asia/Bangkok">Asia/Bangkok</option>
              <option value="UTC">UTC</option>
            </select>
          </div>
          <div class="form-group">
            <label>Number format</label>
            <select :value="localPrefs.numberFormat" @change="updateLocalPreference({ numberFormat: ($event.target as HTMLSelectElement).value as any })">
              <option value="1,234.56">1,234.56</option>
              <option value="1.234,56">1.234,56</option>
            </select>
          </div>
          <div class="form-group">
            <label>Currency display</label>
            <select :value="localPrefs.currencyDisplay" @change="updateLocalPreference({ currencyDisplay: ($event.target as HTMLSelectElement).value as any })">
              <option value="THB 1,234.56">THB 1,234.56</option>
              <option value="฿1,234.56">฿1,234.56</option>
            </select>
          </div>
        </div>
      </div>

      <!-- Configurable Save & Cancel Actions -->
      <AppSettingsActions
        :show-save="true"
        :show-cancel="true"
        :save-loading="loading"
        :save-disabled="!isDirty"
        :cancel-disabled="!isDirty"
        @save="handleSave"
        @cancel="handleCancel"
      />
    </div>
  </section>
</template>

<style scoped>
.personal-appearance-panel {
  padding-bottom: var(--space-6);
}

.appearance-body {
  display: grid !important;
  gap: var(--space-8) !important;
}

.appearance-section-title {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-medium);
  margin: 0 0 var(--space-1);
  color: var(--text-primary);
}

.appearance-section-subtitle {
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
  margin: 0 0 var(--space-4);
  line-height: var(--line-height-relaxed);
}

/* Theme Cards */
.theme-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: var(--space-4);
}

.theme-card {
  display: flex;
  flex-direction: column;
  text-align: left;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  overflow: hidden;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
  cursor: pointer;
  padding: 0;
  outline: none;
}

.theme-card:hover {
  border-color: var(--border-hover);
}

.theme-card:focus-visible {
  box-shadow: 0 0 0 2px var(--action-primary);
}

.theme-card.is-active {
  border-color: var(--action-primary);
  box-shadow: 0 0 0 1px var(--action-primary);
}

.theme-preview {
  height: 120px;
  background: #f6f8fa;
  border-bottom: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  padding: var(--space-3);
  gap: var(--space-2);
}

.theme-preview--light { background: #f6f8fa; }
.theme-preview--light-hc { background: #ffffff; border-color: #000; }
.theme-preview--dark { background: #0d1117; }
.theme-preview--dark-dimmed { background: #22272e; }
.theme-preview--dark-hc { background: #010409; }
.theme-preview--system { background: linear-gradient(135deg, #f6f8fa 50%, #0d1117 50%); }

.mock-header {
  height: 16px;
  background: #ebf0f4;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-2);
}
.theme-preview--dark .mock-header, .theme-preview--dark-hc .mock-header { background: #161b22; }
.theme-preview--dark-dimmed .mock-header { background: #2d333b; }
.theme-preview--light-hc .mock-header { background: #e1e4e8; border: 1px solid #1b1f24; }

.mock-logo { width: 24px; height: 8px; background: #d0d7de; border-radius: var(--radius-sm); }
.theme-preview--dark .mock-logo, .theme-preview--dark-hc .mock-logo, .theme-preview--dark-dimmed .mock-logo { background: #30363d; }
.theme-preview--light-hc .mock-logo { background: #24292e; }

.mock-avatar { width: 12px; height: 12px; background: #d0d7de; border-radius: 50%; }
.theme-preview--dark .mock-avatar, .theme-preview--dark-hc .mock-avatar, .theme-preview--dark-dimmed .mock-avatar { background: #30363d; }
.theme-preview--light-hc .mock-avatar { background: #24292e; }

.mock-body { display: flex; gap: var(--space-2); flex: 1; }

.mock-sidebar {
  width: 40px;
  background: #ffffff;
  border: 1px solid #d0d7de;
  border-radius: var(--radius-sm);
  padding: var(--space-1);
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}
.theme-preview--dark .mock-sidebar, .theme-preview--dark-hc .mock-sidebar { background: #010409; border-color: #30363d; }
.theme-preview--dark-dimmed .mock-sidebar { background: #22272e; border-color: #444c56; }
.theme-preview--light-hc .mock-sidebar { border-color: #1b1f24; }

.mock-nav-item { height: 6px; background: #ebf0f4; border-radius: var(--radius-sm); }
.theme-preview--dark .mock-nav-item, .theme-preview--dark-hc .mock-nav-item { background: #21262d; }
.theme-preview--dark-dimmed .mock-nav-item { background: #373e47; }
.theme-preview--light-hc .mock-nav-item { background: #24292e; }

.mock-content { flex: 1; display: flex; flex-direction: column; gap: var(--space-2); }
.mock-card { flex: 1; background: #ffffff; border: 1px solid #d0d7de; border-radius: var(--radius-sm); }
.theme-preview--dark .mock-card, .theme-preview--dark-hc .mock-card { background: #0d1117; border-color: #30363d; }
.theme-preview--dark-dimmed .mock-card { background: #22272e; border-color: #444c56; }
.theme-preview--light-hc .mock-card { border-color: #1b1f24; }

.theme-card-info {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-3);
}

.theme-card-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.theme-card-title {
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-sm);
}

.theme-card-desc {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-relaxed);
}

/* Density Cards */
.density-list {
  display: grid;
  gap: var(--space-3);
}

.density-card {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  cursor: pointer;
  text-align: left;
  outline: none;
}

.density-card:hover {
  border-color: var(--border-hover);
}

.density-card.is-active {
  border-color: var(--action-primary);
  background: var(--bg-card-selected);
}

.radio-indicator {
  margin-top: 2px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 1px solid var(--border-subtle);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: var(--bg-canvas);
}

.density-card.is-active .radio-indicator {
  border-color: var(--action-primary);
  background: var(--action-primary);
}

.radio-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #fff;
}

.density-card-text {
  display: flex;
  flex-direction: column;
}

.density-card-title {
  color: var(--text-primary);
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-sm);
}

.density-card-desc {
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

/* Accessibility Toggles */
.accessibility-list {
  display: flex;
  flex-direction: column;
}

.toggle-row {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-4) 0;
  border-bottom: 1px solid var(--border-subtle);
  cursor: pointer;
}
.toggle-row:last-child {
  border-bottom: none;
}

.toggle-control {
  position: relative;
  width: 40px;
  height: 24px;
  flex-shrink: 0;
}

.toggle-control input {
  opacity: 0;
  width: 0;
  height: 0;
  position: absolute;
}

.toggle-track {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background-color: var(--bg-canvas-muted);
  border-radius: 24px;
  transition: .2s;
  border: 1px solid var(--border-subtle);
}

.toggle-track:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 2px;
  bottom: 2px;
  background-color: white;
  border-radius: 50%;
  transition: .2s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}

.toggle-control input:checked + .toggle-track {
  background-color: var(--action-primary);
  border-color: var(--action-primary);
}

.toggle-control input:checked + .toggle-track:before {
  transform: translateX(16px);
}

.toggle-control input:focus-visible + .toggle-track {
  box-shadow: 0 0 0 2px var(--bg-canvas), 0 0 0 4px var(--action-primary);
}

.toggle-text {
  display: flex;
  flex-direction: column;
}

.toggle-title {
  color: var(--text-primary);
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-sm);
}

.toggle-desc {
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

/* Form Grid */
.formatting-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-5);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.form-group label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--text-primary);
}

.form-group select {
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}

.form-group select:focus {
  border-color: var(--action-primary);
  outline: none;
  box-shadow: 0 0 0 1px var(--action-primary);
}
</style>
