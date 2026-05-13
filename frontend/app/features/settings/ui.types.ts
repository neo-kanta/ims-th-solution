export type BooleanFilterValue = "all" | "true" | "false";

export interface UserDirectoryFilters {
  search: string;
  active: BooleanFilterValue;
  locked: BooleanFilterValue;
}

export interface AuditLogFilters {
  actorId: string;
  eventType: string;
  targetType: string;
  targetId: string;
  since: string;
  until: string;
}

export interface SettingsKpiMetric {
  id: string;
  label: string;
  value: string | number;
  helper: string;
  icon: string;
  tone: "primary" | "success" | "warning" | "danger" | "neutral";
}

export type SettingsSectionId =
  | "overview"
  | "personal-account"
  | "users"
  | "groups"
  | "function-permissions"
  | "data-permissions"
  | "security-policy"
  | "notifications"
  | "audit";

export interface SettingsNavigationItem {
  id: SettingsSectionId;
  label: string;
  description: string;
  icon: string;
  status: "live" | "read-only" | "pending";
  count?: string | number;
  disabled?: boolean;
}

export interface SettingsOverviewSignal {
  id: string;
  label: string;
  value: string;
  helper: string;
  tone: "success" | "warning" | "danger" | "neutral" | "info";
}

export interface SettingsApiCapability {
  id: string;
  label: string;
  capability: string;
  status: "live" | "not-exposed" | "read-only";
}

export interface SettingsGroupRole {
  id: string;
  name: string;
  description: string;
  source: "directory" | "demo";
  membersCount: number;
  responsibilities: string[];
  permissionFamilies: string[];
  dataScopes: string[];
  riskLevel: "standard" | "elevated" | "restricted";
}

export type PermissionActionKey =
  | "view"
  | "search"
  | "add"
  | "edit"
  | "delete"
  | "approve"
  | "revokeApproval"
  | "export"
  | "settings";

export interface PermissionActionColumn {
  key: PermissionActionKey;
  label: string;
}

export interface FunctionPermissionRow {
  id: string;
  menu: string;
  functionName: string;
  owner: string;
  status: "live-session" | "demo";
  permissions: Record<PermissionActionKey, boolean>;
}

export interface DataPermissionGrant {
  id: string;
  user: string;
  userLabel: string;
  contractId: string;
  contractName: string;
  scope: "All funds" | "Selected funds" | "Read only";
  source: "session" | "demo";
  updatedAt: string;
}

export interface SecurityPolicySetting {
  id: string;
  label: string;
  value: string | number | boolean;
  helper: string;
  unit?: string;
  disabled?: boolean;
}

export interface SecurityPolicyModel {
  authMode: "Internal" | "AD / SSO";
  source: "demo";
  settings: SecurityPolicySetting[];
}

export interface NotificationPreference {
  id: string;
  event: string;
  audience: string;
  channels: string[];
  severity: "Critical" | "High" | "Medium" | "Low";
  enabled: boolean;
  source: "demo";
}
