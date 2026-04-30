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
