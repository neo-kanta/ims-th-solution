import type {
  PersonalActivityEvent,
  PersonalAppearancePreferences,
  PersonalLeaveRequest,
  PersonalNotificationPreference,
  PersonalWorkPreferences,
} from "../personal.types";

const STORAGE_KEYS = {
  notifications: "ims.personal.notifications",
  appearance: "ims.personal.appearance",
  workPreferences: "ims.personal.workPreferences",
} as const;

const DEFAULT_NOTIFICATIONS: PersonalNotificationPreference[] = [
  {
    id: "approval-assigned",
    label: "Approval assigned to me",
    description: "Notify me when an approval item requires my action.",
    severity: "High",
    inApp: true,
    email: true,
    source: "local",
  },
  {
    id: "report-decision",
    label: "My report approved/rejected",
    description: "Notify me when my research report or workflow item is decided.",
    severity: "Medium",
    inApp: true,
    email: true,
    source: "local",
  },
  {
    id: "workflow-updates",
    label: "Workflow day-start/day-end/closing updates",
    description: "Operational updates for workflow milestones that affect my work queue.",
    severity: "Medium",
    inApp: true,
    email: false,
    source: "local",
  },
  {
    id: "irg-compliance",
    label: "IRG/compliance alert",
    description: "Compliance alerts that may block or require review before execution.",
    severity: "Critical",
    inApp: true,
    email: true,
    mandatory: true,
    mandatoryReason: "Compliance alerts are mandatory under the current operating policy.",
    source: "local",
  },
  {
    id: "stop-loss-threshold",
    label: "Stop-loss / threshold alert",
    description: "Risk threshold events connected to contracts or portfolios I can access.",
    severity: "High",
    inApp: true,
    email: true,
    source: "local",
  },
  {
    id: "security-login",
    label: "Security login alert",
    description: "Security notifications for login and session activity on my account.",
    severity: "High",
    inApp: true,
    email: true,
    mandatory: true,
    mandatoryReason: "Security alerts cannot be disabled for personal accounts.",
    source: "local",
  },
];

const DEFAULT_APPEARANCE: PersonalAppearancePreferences = {
  theme: "system",
  density: "comfortable",
  increaseContrast: false,
  reduceMotion: false,
  showKeyboardHints: false,
  dateFormat: "YYYY-MM-DD",
  timezone: "Asia/Bangkok",
  numberFormat: "1,234.56",
  currencyDisplay: "THB 1,234.56",
  sidebarBehavior: "collapsed",
  source: "local",
};

const DEFAULT_WORK_PREFERENCES: PersonalWorkPreferences = {
  defaultDashboard: "overview",
  defaultFund: "",
  language: "en",
  tablePageSize: 25,
  source: "local",
};

function canUseStorage() {
  return import.meta.client && typeof window !== "undefined";
}

function readStored<T>(key: string, fallback: T): T {
  if (!canUseStorage()) return fallback;

  try {
    const stored = window.localStorage.getItem(key);
    if (!stored) return fallback;

    const parsed = JSON.parse(stored);
    if (Array.isArray(fallback)) {
      return Array.isArray(parsed) ? (parsed as T) : fallback;
    }

    return { ...fallback, ...parsed };
  } catch {
    return fallback;
  }
}

function writeStored<T>(key: string, value: T): T {
  if (canUseStorage()) {
    window.localStorage.setItem(key, JSON.stringify(value));
  }

  return value;
}

export const personalSettingsApi = {
  async listNotifications(): Promise<PersonalNotificationPreference[]> {
    const stored = readStored(STORAGE_KEYS.notifications, DEFAULT_NOTIFICATIONS);
    return stored.map((item) => ({ ...item, source: "local" }));
  },

  async updateNotifications(
    preferences: PersonalNotificationPreference[],
  ): Promise<PersonalNotificationPreference[]> {
    const normalized = preferences.map((preference) => ({
      ...preference,
      inApp: preference.mandatory ? true : preference.inApp,
      email: preference.mandatory ? true : preference.email,
      source: "local" as const,
    }));

    return writeStored(STORAGE_KEYS.notifications, normalized);
  },

  async getAppearance(): Promise<PersonalAppearancePreferences> {
    return readStored(STORAGE_KEYS.appearance, DEFAULT_APPEARANCE);
  },

  async updateAppearance(
    preferences: PersonalAppearancePreferences,
  ): Promise<PersonalAppearancePreferences> {
    return writeStored(STORAGE_KEYS.appearance, {
      ...preferences,
      source: "local",
    });
  },

  async getWorkPreferences(): Promise<PersonalWorkPreferences> {
    return readStored(STORAGE_KEYS.workPreferences, DEFAULT_WORK_PREFERENCES);
  },

  async updateWorkPreferences(
    preferences: PersonalWorkPreferences,
  ): Promise<PersonalWorkPreferences> {
    return writeStored(STORAGE_KEYS.workPreferences, {
      ...preferences,
      source: "local",
    });
  },

  async listLeaveRequests(): Promise<PersonalLeaveRequest[]> {
    return [];
  },

  async listActivity(): Promise<PersonalActivityEvent[]> {
    return [];
  },
};
