export type AuditSeverity = "critical" | "high" | "medium" | "low";

const CRITICAL_EVENTS = new Set([
  "REFRESH_TOKEN_BREACH",
  "RATE_LIMIT_BLOCKED",
  "ACCOUNT_LOCKED",
]);

const HIGH_EVENTS = new Set([
  "USER_DEACTIVATED",
  "PASSWORD_CHANGE",
  "SESSION_REVOKED",
  "MFA_DISABLED",
  "LOGIN_FAILURE",
]);

const MEDIUM_EVENTS = new Set([
  "USER_CREATED",
  "USER_ACTIVATED",
  "ACCOUNT_UNLOCKED",
  "MFA_ENROLLED",
  "MFA_ENABLED",
  "MFA_CHALLENGE_FAILURE",
  "AUDIT_LOG_EXPORTED",
]);

export function getAuditSeverity(eventType: string): AuditSeverity {
  if (CRITICAL_EVENTS.has(eventType)) {
    return "critical";
  }

  if (HIGH_EVENTS.has(eventType)) {
    return "high";
  }

  if (MEDIUM_EVENTS.has(eventType)) {
    return "medium";
  }

  return "low";
}

export function getAuditSeverityKey(severity: AuditSeverity): string {
  switch (severity) {
    case "critical":
      return "settings.console.audit.severity.critical";
    case "high":
      return "settings.console.audit.severity.high";
    case "medium":
      return "settings.console.audit.severity.medium";
    case "low":
      return "settings.console.audit.severity.low";
  }
}

export function getAuditSeverityClass(severity: AuditSeverity): string {
  switch (severity) {
    case "critical":
      return "badge-error";
    case "high":
      return "badge-warning";
    case "medium":
      return "badge-info";
    case "low":
      return "badge-neutral";
  }
}

export function isRiskAuditEvent(eventType: string): boolean {
  return getAuditSeverity(eventType) !== "low";
}
