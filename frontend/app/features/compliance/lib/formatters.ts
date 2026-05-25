/**
 * Compliance formatters — Phase 1.
 *
 * Pure helpers used by badges, the result panel, and the rule table. They
 * never call out to the network and never invent backend data.
 */
import type {
  ComplianceBackendSeverity,
  ComplianceEffectiveWindow,
  ComplianceRule,
  ComplianceRuleDerivedStatus,
  ComplianceVerdict,
} from "../types";

export type StatusTone = "success" | "warning" | "error" | "info" | "neutral";

export function verdictTone(verdict: ComplianceVerdict): StatusTone {
  switch (verdict) {
    case "PASS":
      return "success";
    case "WARN":
      return "warning";
    case "BLOCK":
      return "error";
    default:
      return "neutral";
  }
}

export function verdictLabel(verdict: ComplianceVerdict): string {
  switch (verdict) {
    case "PASS":
      return "PASS";
    case "WARN":
      return "WARN";
    case "BLOCK":
      return "BLOCK";
    default:
      return verdict;
  }
}

export function severityTone(severity: ComplianceBackendSeverity): StatusTone {
  switch (severity) {
    case "BLOCK":
      return "error";
    case "WARN":
      return "warning";
    case "REQUIRE_APPROVAL":
      return "info";
    case "MONITOR":
      return "neutral";
    default:
      return "neutral";
  }
}

export function severityLabel(severity: ComplianceBackendSeverity): string {
  switch (severity) {
    case "BLOCK":
      return "Blocker";
    case "WARN":
      return "Warning";
    case "REQUIRE_APPROVAL":
      return "Requires approval";
    case "MONITOR":
      return "Monitor only";
    default:
      return severity;
  }
}

function asIsoDate(value: unknown): string | null {
  if (typeof value !== "string" || value.trim() === "") return null;
  return value;
}

function extractWindow(
  raw: ComplianceEffectiveWindow | undefined,
): { from: string | null; to: string | null } {
  if (!raw) return { from: null, to: null };
  const from = asIsoDate((raw as Record<string, unknown>).from);
  const to = asIsoDate((raw as Record<string, unknown>).to);
  return { from, to };
}

/**
 * Map a backend rule instance to one of the four UI-visible derived states.
 * No HTTP — pure derivation. Effective window uses today's date in UTC.
 */
export function deriveRuleStatus(
  rule: ComplianceRule,
  now: Date = new Date(),
): ComplianceRuleDerivedStatus {
  if (!rule.isActive) return "DISABLED";

  const { from, to } = extractWindow(rule.effectiveWindow);
  const today = now.toISOString().slice(0, 10);

  if (from && from > today) return "SCHEDULED";
  if (to && to < today) return "EXPIRED";
  return "ACTIVE";
}

export function ruleStatusTone(
  status: ComplianceRuleDerivedStatus,
): StatusTone {
  switch (status) {
    case "ACTIVE":
      return "success";
    case "SCHEDULED":
      return "info";
    case "EXPIRED":
      return "warning";
    case "DISABLED":
      return "neutral";
    default:
      return "neutral";
  }
}

export function ruleStatusLabel(status: ComplianceRuleDerivedStatus): string {
  switch (status) {
    case "ACTIVE":
      return "Active";
    case "SCHEDULED":
      return "Scheduled";
    case "EXPIRED":
      return "Expired";
    case "DISABLED":
      return "Disabled";
    default:
      return status;
  }
}

export function formatEffectiveWindow(
  raw: ComplianceEffectiveWindow | undefined,
): string {
  const { from, to } = extractWindow(raw);
  if (!from && !to) return "—";
  if (from && !to) return `From ${from}`;
  if (!from && to) return `Until ${to}`;
  return `${from} → ${to}`;
}

export function formatIsoDate(value: string | null | undefined): string {
  if (!value) return "—";
  if (value.length >= 10) return value.slice(0, 10);
  return value;
}

export function formatIsoDateTime(value: string | null | undefined): string {
  if (!value) return "—";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toISOString().replace("T", " ").slice(0, 16) + " UTC";
}

const UUID_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export function isUuid(value: string): boolean {
  return UUID_RE.test(value.trim());
}

export function isPositiveDecimal(value: string): boolean {
  if (!value || value.trim() === "") return false;
  if (!/^\d+(\.\d+)?$/.test(value.trim())) return false;
  return Number.parseFloat(value) > 0;
}
