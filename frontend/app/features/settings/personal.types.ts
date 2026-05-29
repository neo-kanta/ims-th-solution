export type PersonalSettingsSectionId =
  | "public-profile"
  | "account"
  | "appearance"
  | "notifications"
  | "password-auth"
  | "sessions-mfa"
  | "delegation-leave"
  | "locale-timezone"
  | "workstation"
  | "activity-log"
  | "close-account";

export type UserPreferenceSource = "local" | "api";

export interface PersonalNotificationPreference {
  id: string;
  label: string;
  description: string;
  severity: "Critical" | "High" | "Medium" | "Low";
  inApp: boolean;
  email: boolean;
  mandatory?: boolean;
  mandatoryReason?: string;
  source: UserPreferenceSource;
}

export interface PersonalAppearancePreferences {
  theme: "system" | "light" | "light-high-contrast" | "dark" | "dark-dimmed" | "dark-high-contrast";
  density: "comfortable" | "compact" | "system";
  increaseContrast: boolean;
  reduceMotion: boolean;
  showKeyboardHints: boolean;
  dateFormat: "YYYY-MM-DD" | "DD/MM/YYYY" | "MM/DD/YYYY";
  timezone: string;
  numberFormat: "1,234.56" | "1.234,56";
  currencyDisplay: "THB 1,234.56" | "฿1,234.56";
  sidebarBehavior: "expanded" | "collapsed";
  source: UserPreferenceSource;
}

export interface PersonalWorkPreferences {
  defaultDashboard: "overview" | "workflow" | "investment" | "notifications";
  defaultFund: string;
  language: "en" | "th" | "zh";
  tablePageSize: number;
  source: UserPreferenceSource;
}

export type PersonalLeaveRequestType = "leave" | "temporary-leave" | "cancellation";
export type PersonalLeaveRequestStatus = "draft" | "pending" | "approved" | "cancelled" | "rejected";

export interface PersonalLeaveRequest {
  id: string;
  type: PersonalLeaveRequestType;
  status: PersonalLeaveRequestStatus;
  startDate: string;
  endDate: string;
  reason: string;
  delegationUser?: string;
  createdAt: string;
  source: UserPreferenceSource;
}

export interface PersonalActivityEvent {
  id: string;
  category: "security" | "settings" | "approval" | "workflow";
  title: string;
  description: string;
  createdAt: string;
  ipAddress?: string;
  source: UserPreferenceSource;
}
